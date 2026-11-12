package store

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

// defaultAuditPageSize/maxAuditPageSize bound what the screen can ask for.
// The cap is there because the trail only ever grows: without it a client
// could ask for the whole table in one request.
const (
	defaultAuditPageSize = 50
	maxAuditPageSize     = 200
)

// AuditObserver is notified once per entry actually written, so audit events
// can be counted as a metric without this package importing
// internal/metrics - which imports this one. Same shape, and same reason, as
// middleware.EndpointObserver.
type AuditObserver func(action, organizationID, userID string)

// AuditStore is the append-only trail of acts that change who can do what, or
// what this deployment is configured to do. It has no Update and no Delete,
// by construction: a trail that can be rewritten is not one. The only way a
// row ever changes is a reference being nulled when the account or
// organization it points at is deleted (see V10__add_audit_log.sql).
type AuditStore struct {
	pool    *pgxpool.Pool
	observe AuditObserver
}

func NewAuditStore(pool *pgxpool.Pool) *AuditStore {
	return &AuditStore{pool: pool}
}

// SetObserver attaches the metrics counter. It is a setter rather than a
// constructor argument because metrics.Setup needs the stores to build its
// gauges, so the store necessarily exists first; main calls this once, before
// the server starts serving.
func (s *AuditStore) SetObserver(observe AuditObserver) {
	s.observe = observe
}

// AuditEvent is one act to record. UserID is empty for the acts with no
// authenticated actor, OrganizationID for those that belong to no
// organization; Detail is free-form context for the screen - what changed,
// from what to what - and is never relied on by code.
type AuditEvent struct {
	UserID         string
	OrganizationID string
	Action         string
	SubjectType    string
	SubjectID      string
	Detail         map[string]any
}

// auditData is the shape of the row's data column, following the same
// pattern as the other stores here. It is only ever unmarshalled: the column
// is written by the INSERT in Record, which fills ActorEmail from users
// itself.
type auditData struct {
	ActorEmail string         `json:"actorEmail,omitempty"`
	Detail     map[string]any `json:"detail,omitempty"`
}

// Record writes one entry. It deliberately returns nothing: an audit write
// must never be the reason a user's action fails, so a failure is logged at
// error level and the action stands. Refusing the action instead would make
// the audit trail a single point of failure for the whole app - every login,
// every organization edit would depend on one more successful write.
//
// Every entry also goes out through the process logger at info level. That
// is one slog call, not two sinks to keep in step: internal/telemetry
// installs a multi-handler as the default logger, so the same record reaches
// stdout and - when CWCLOCK_OTEL_ENDPOINT is set - the OTLP collector.
func (s *AuditStore) Record(ctx context.Context, e AuditEvent) {
	// An entry with no detail stores an empty object rather than JSON null,
	// so data->>'detail' is always an object for anything reading the table
	// directly with SQL.
	detail := []byte("{}")
	if len(e.Detail) > 0 {
		encoded, err := json.Marshal(e.Detail)
		if err != nil {
			slog.Error("audit detail could not be encoded", "audit_action", e.Action, "error", err)
			return
		}
		detail = encoded
	}

	// WithoutCancel, not ctx itself: Record is always called after the act
	// has already succeeded, so a client that hangs up at that moment would
	// otherwise cancel the write and leave the act unrecorded. The context's
	// values (the request's trace span) are kept, only its cancellation is
	// dropped.
	ctx = context.WithoutCancel(ctx)

	// INSERT ... SELECT so the acting account's email is snapshotted from
	// users in the same statement, without a second round trip: the trail
	// must keep naming who acted after that account is renamed or deleted.
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO audit_log (user_id, organization_id, action, subject_type, subject_id, data)
		SELECT $1::uuid, $2::uuid, $3, $4, $5, jsonb_build_object(
			'actorEmail', coalesce((SELECT email FROM users WHERE id = $1::uuid), ''),
			'detail', $6::jsonb
		)
	`, nullUUID(e.UserID), nullUUID(e.OrganizationID), e.Action, e.SubjectType, e.SubjectID, detail); err != nil {
		slog.Error("audit entry could not be written", "audit_action", e.Action,
			"user", e.UserID, "organization", e.OrganizationID, "error", err)
		return
	}

	slog.Info("audit",
		"audit_action", e.Action,
		"user", e.UserID,
		"organization", e.OrganizationID,
		"subject_type", e.SubjectType,
		"subject_id", e.SubjectID,
		"detail", string(detail),
	)

	if s.observe != nil {
		s.observe(e.Action, e.OrganizationID, e.UserID)
	}
}

// AuditFilter narrows the trail for the superadmin screen. An empty field
// means "no filter on this field", so a single query shape serves every
// combination of filters the screen can ask for. Page is 1-based.
type AuditFilter struct {
	Action      string
	SubjectType string
	SubjectID   string
	Page        int
	Size        int
}

// limitOffset clamps the requested page and asks for one row more than the
// page needs, which is how List knows whether a next page exists without a
// second COUNT query over a table that only grows.
func (f AuditFilter) limitOffset() (int, int) {
	size := f.Size
	if size <= 0 {
		size = defaultAuditPageSize
	}
	if size > maxAuditPageSize {
		size = maxAuditPageSize
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	return size + 1, (page - 1) * size
}

// List returns one page of the trail, newest first, plus the distinct actions
// present in the whole table so the screen can offer them as a filter without
// a hardcoded list that drifts from models.Audit* .
func (s *AuditStore) List(ctx context.Context, f AuditFilter) (models.AuditPage, error) {
	limit, offset := f.limitOffset()

	rows, err := s.pool.Query(ctx, `
		SELECT id, coalesce(user_id::text, ''), coalesce(organization_id::text, ''),
		       action, subject_type, subject_id, data, created_at
		FROM audit_log
		WHERE ($3 = '' OR action = $3)
		  AND ($4 = '' OR subject_type = $4)
		  AND ($5 = '' OR subject_id = $5)
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`, limit, offset, f.Action, f.SubjectType, f.SubjectID)
	if err != nil {
		return models.AuditPage{}, err
	}
	defer rows.Close()

	page := models.AuditPage{Entries: []models.AuditEntry{}, Actions: []string{}}
	for rows.Next() {
		var e models.AuditEntry
		var raw []byte
		if err := rows.Scan(&e.ID, &e.UserID, &e.OrganizationID,
			&e.Action, &e.SubjectType, &e.SubjectID, &raw, &e.CreatedAt); err != nil {
			return models.AuditPage{}, err
		}
		var d auditData
		if err := json.Unmarshal(raw, &d); err != nil {
			return models.AuditPage{}, err
		}
		e.UserEmail = d.ActorEmail
		e.Detail = d.Detail
		page.Entries = append(page.Entries, e)
	}
	if err := rows.Err(); err != nil {
		return models.AuditPage{}, err
	}

	// The extra row fetched by limitOffset is the answer to "is there a next
	// page?", and is not part of this one.
	if len(page.Entries) == limit {
		page.HasMore = true
		page.Entries = page.Entries[:limit-1]
	}

	// The filter values come from the code's own lists, not from a DISTINCT
	// over audit_log: the table now grows with every time record, and the
	// screen should not pay for a scan of it on each page load. See
	// models.AuditActions.
	page.Actions = models.AuditActions
	page.SubjectTypes = models.AuditSubjectTypes
	return page, nil
}

// nullUUID turns an empty id into a SQL NULL, so the nullable references in
// audit_log can be left unset without the caller building a *string.
func nullUUID(id string) any {
	if utils.IsBlank(id) {
		return nil
	}
	return id
}
