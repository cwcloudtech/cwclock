package handlers

import (
	"context"
	"net/http"
	"strconv"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
)

// auditor is embedded by every handler that records audit entries. It exists
// so the two facts every entry needs - who acted and in which organization -
// are read from the request context in one place rather than at each of the
// forty-odd call sites, and so the store stays a plain store: the tenant is
// resolved by internal/middleware, which imports internal/store, so a store
// can never import it back to look the tenant up itself.
type auditor struct {
	audit *store.AuditStore
}

func newAuditor(audit *store.AuditStore) auditor {
	return auditor{audit: audit}
}

// record writes an entry for the act this request just performed, attributing
// it to the authenticated caller and the organization the route resolved.
// Call it after the act succeeded: the trail records what happened, not what
// was attempted.
func (a auditor) record(r *http.Request, action, subjectType, subjectID string, detail map[string]any) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	a.recordIn(r, orgID, action, subjectType, subjectID, detail)
}

// recordIn is record for the acts whose organization isn't the one the route
// resolved. Two cases: creating an organization (there was no {orgId} in the
// path yet) and deleting one, where orgID must be left empty - the entry is
// written after the row is gone, and audit_log.organization_id references
// organizations, so naming it would make the insert fail outright. The
// organization is still named by subject_id in that case.
func (a auditor) recordIn(r *http.Request, orgID, action, subjectType, subjectID string, detail map[string]any) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	a.audit.Record(r.Context(), store.AuditEvent{
		UserID:         userID,
		OrganizationID: orgID,
		Action:         action,
		SubjectType:    subjectType,
		SubjectID:      subjectID,
		Detail:         detail,
	})
}

// recordAs is record for the acts that happen outside the Auth middleware, so
// there is no caller in the context to attribute them to: signing in,
// registering, confirming an account, finishing a password reset from an
// emailed link. None of those belong to an organization either.
func (a auditor) recordAs(ctx context.Context, userID, action, subjectType, subjectID string, detail map[string]any) {
	a.audit.Record(ctx, store.AuditEvent{
		UserID:      userID,
		Action:      action,
		SubjectType: subjectType,
		SubjectID:   subjectID,
		Detail:      detail,
	})
}

// auditFilterFromQuery reads the superadmin screen's filters off the query
// string. Anything missing or unparseable means "no filter" / "first page",
// so a hand-written URL can't 400 the trail.
func auditFilterFromQuery(r *http.Request) store.AuditFilter {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	return store.AuditFilter{
		Action:      r.URL.Query().Get("action"),
		SubjectType: r.URL.Query().Get("subjectType"),
		SubjectID:   r.URL.Query().Get("subject"),
		Page:        page,
		Size:        size,
	}
}

// auditConnectionDetail describes an external connection for the trail by
// where it points, never by how it authenticates: an audit entry is readable
// by every superadmin and ends up in the logs and the collector, so an S3
// secret key or an SSH private key must not be anywhere near it.
func auditConnectionDetail(conn models.ExternalConnection) map[string]any {
	detail := map[string]any{"type": string(conn.Type)}
	switch conn.Type {
	case models.ExternalConnectionS3:
		detail["endpoint"] = conn.Endpoint
		detail["bucket"] = conn.BucketName
	case models.ExternalConnectionGoogleDrive:
		detail["folderId"] = conn.FolderID
	case models.ExternalConnectionGit:
		detail["repoUrl"] = conn.RepoURL
	}
	return detail
}
