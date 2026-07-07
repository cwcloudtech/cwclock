package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

type TimeEntryStore struct {
	pool *pgxpool.Pool
}

func NewTimeEntryStore(pool *pgxpool.Pool) *TimeEntryStore {
	return &TimeEntryStore{pool: pool}
}

type timeEntryData struct {
	Text   string  `json:"text"`
	Day    string  `json:"day"`
	Start  *string `json:"start"`
	End    *string `json:"end"`
	AllDay bool    `json:"allDay"`
}

// TimeEntryFields holds the editable, non-relational fields of a time entry.
// When AllDay is true, Start/End are always stored as null regardless of
// what's passed.
type TimeEntryFields struct {
	Text   string
	Day    string
	Start  *string
	End    *string
	AllDay bool
}

func toTimeEntryData(f TimeEntryFields) timeEntryData {
	d := timeEntryData{Text: f.Text, Day: f.Day, AllDay: f.AllDay}
	if !f.AllDay {
		d.Start = f.Start
		d.End = f.End
	}
	return d
}

func scanTimeEntry(row pgx.Row) (models.TimeEntry, error) {
	var t models.TimeEntry
	var raw []byte
	if err := row.Scan(&t.ID, &t.OrganizationID, &t.ClientID, &t.ProjectID, &t.UserID, &raw, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.TimeEntry{}, ErrNotFound
		}
		return models.TimeEntry{}, err
	}
	var d timeEntryData
	if err := json.Unmarshal(raw, &d); err != nil {
		return models.TimeEntry{}, err
	}
	t.Text = d.Text
	t.Day = d.Day
	t.Start = d.Start
	t.End = d.End
	t.AllDay = d.AllDay
	return t, nil
}

func (s *TimeEntryStore) Create(ctx context.Context, orgID, clientID, projectID, userID string, f TimeEntryFields) (models.TimeEntry, error) {
	data, err := json.Marshal(toTimeEntryData(f))
	if err != nil {
		return models.TimeEntry{}, err
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO time_entries (organization_id, client_id, project_id, user_id, data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, organization_id, client_id, project_id, user_id, data, created_at, updated_at
	`, orgID, clientID, projectID, userID, data)
	return scanTimeEntry(row)
}

func (s *TimeEntryStore) FindByID(ctx context.Context, id string) (models.TimeEntry, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, client_id, project_id, user_id, data, created_at, updated_at
		FROM time_entries WHERE id = $1
	`, id)
	return scanTimeEntry(row)
}

// List returns time entries for an organization. When userID is non-empty,
// results are restricted to that user's own entries (used to enforce that
// members only see their own time records).
func (s *TimeEntryStore) List(ctx context.Context, orgID, userID string) ([]models.TimeEntry, error) {
	var rows pgx.Rows
	var err error
	if utils.IsBlank(userID) {
		rows, err = s.pool.Query(ctx, `
			SELECT id, organization_id, client_id, project_id, user_id, data, created_at, updated_at
			FROM time_entries WHERE organization_id = $1
			ORDER BY created_at DESC
		`, orgID)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT id, organization_id, client_id, project_id, user_id, data, created_at, updated_at
			FROM time_entries WHERE organization_id = $1 AND user_id = $2
			ORDER BY created_at DESC
		`, orgID, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []models.TimeEntry{}
	for rows.Next() {
		t, err := scanTimeEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, t)
	}
	return entries, rows.Err()
}

// TimeEntryReassignment holds the optional relational reassignments applied
// on top of a field update. Empty strings mean "leave unchanged".
type TimeEntryReassignment struct {
	UserID    string
	ClientID  string
	ProjectID string
}

// Update updates a time entry's fields. Any non-empty field on reassign
// also moves the entry to that user/client/project (callers must have
// already authorized the reassignment).
func (s *TimeEntryStore) Update(ctx context.Context, id string, reassign TimeEntryReassignment, f TimeEntryFields) (models.TimeEntry, error) {
	data, err := json.Marshal(toTimeEntryData(f))
	if err != nil {
		return models.TimeEntry{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE time_entries
		SET data = $2,
		    user_id = COALESCE(NULLIF($3, '')::uuid, user_id),
		    client_id = COALESCE(NULLIF($4, '')::uuid, client_id),
		    project_id = COALESCE(NULLIF($5, '')::uuid, project_id),
		    updated_at = now()
		WHERE id = $1
		RETURNING id, organization_id, client_id, project_id, user_id, data, created_at, updated_at
	`, id, data, reassign.UserID, reassign.ClientID, reassign.ProjectID)
	return scanTimeEntry(row)
}

// ReportFilter narrows the entries a report is built from.
type ReportFilter struct {
	Start      string // inclusive, "YYYY-MM-DD"
	End        string // inclusive, "YYYY-MM-DD"
	ClientIDs  []string
	ProjectIDs []string
	UserIDs    []string
}

// ListForReport returns an organization's time entries within a date range,
// optionally narrowed to specific clients/projects/members, oldest first
// within a day so summary/detailed reports can present a stable order.
func (s *TimeEntryStore) ListForReport(ctx context.Context, orgID string, f ReportFilter) ([]models.TimeEntry, error) {
	query := `
		SELECT id, organization_id, client_id, project_id, user_id, data, created_at, updated_at
		FROM time_entries
		WHERE organization_id = $1
		  AND data->>'day' >= $2
		  AND data->>'day' <= $3
	`
	args := []any{orgID, f.Start, f.End}

	addFilter := func(column string, ids []string) {
		if len(ids) == 0 {
			return
		}
		args = append(args, ids)
		query += fmt.Sprintf(" AND %s::text = ANY($%d)", column, len(args))
	}
	addFilter("client_id", f.ClientIDs)
	addFilter("project_id", f.ProjectIDs)
	addFilter("user_id", f.UserIDs)

	query += " ORDER BY data->>'day', data->>'start'"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []models.TimeEntry{}
	for rows.Next() {
		t, err := scanTimeEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, t)
	}
	return entries, rows.Err()
}

// ListRecent returns time entries whose day falls within the last 24h,
// across every organization, for the "task duration in the last 24h" metric.
func (s *TimeEntryStore) ListRecent(ctx context.Context) ([]models.TimeEntry, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, client_id, project_id, user_id, data, created_at, updated_at
		FROM time_entries
		WHERE (data->>'day')::date >= (now() - interval '24 hours')::date
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []models.TimeEntry{}
	for rows.Next() {
		t, err := scanTimeEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, t)
	}
	return entries, rows.Err()
}

func (s *TimeEntryStore) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM time_entries WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
