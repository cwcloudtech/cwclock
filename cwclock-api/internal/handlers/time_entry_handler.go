package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

const (
	defaultTimeEntryPageSize = 50
	maxTimeEntryPageSize     = 200
)

// paginationParams reads page/pageSize query params, defaulting to page 1
// and defaultTimeEntryPageSize, clamped to [1, maxTimeEntryPageSize].
func paginationParams(r *http.Request) (page, pageSize int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ = strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = defaultTimeEntryPageSize
	}
	if pageSize > maxTimeEntryPageSize {
		pageSize = maxTimeEntryPageSize
	}
	return page, pageSize
}

type TimeEntryHandler struct {
	auditor
	entries *store.TimeEntryStore
}

func NewTimeEntryHandler(entries *store.TimeEntryStore, audit *store.AuditStore) *TimeEntryHandler {
	return &TimeEntryHandler{auditor: newAuditor(audit), entries: entries}
}

// auditRecordDetail describes a time record for the trail: whose time it is,
// what it was logged against, and which day. The member is named as well as
// the acting account, because the two differ whenever an admin edits or
// reassigns someone else's day - the case the trail is actually read for.
// The text is in there too, since on an invoice it is the only thing telling
// two otherwise identical entries apart.
func auditRecordDetail(entry models.TimeEntry, l store.TimeEntryLabels) map[string]any {
	return map[string]any{
		"user":    l.UserEmail,
		"client":  l.ClientName,
		"project": l.ProjectName,
		"day":     entry.Day,
		"text":    entry.Text,
	}
}

type timeEntryPayload struct {
	ClientID  string  `json:"clientId"`
	ProjectID string  `json:"projectId"`
	UserID    string  `json:"userId"`
	Text      string  `json:"text"`
	Day       string  `json:"day"`
	Start     *string `json:"start"`
	End       *string `json:"end"`
	AllDay    bool    `json:"allDay"`
	Half      bool    `json:"half"`
}

func (p timeEntryPayload) toFields() store.TimeEntryFields {
	return store.TimeEntryFields{
		Text:   p.Text,
		Day:    p.Day,
		Start:  p.Start,
		End:    p.End,
		AllDay: p.AllDay,
		Half:   p.Half,
	}
}

// canManage reports whether the caller may modify entry, i.e. it is their
// own record, or they hold admin/owner privileges in the organization.
func canManage(r *http.Request, entryUserID string) bool {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == entryUserID {
		return true
	}
	role, _ := middleware.OrgRoleFromContext(r.Context())
	return role == models.RoleAdmin || role == models.RoleOwner
}

// isAdminOrOwner reports whether the caller holds admin/owner privileges,
// required to reassign a time entry to another member.
func isAdminOrOwner(r *http.Request) bool {
	role, _ := middleware.OrgRoleFromContext(r.Context())
	return role == models.RoleAdmin || role == models.RoleOwner
}

// validateAllDayHalf rejects a request marking a time entry as both all-day
// and half-day - the two are mutually exclusive, same model on both flags.
// Returns false (having already written the error response) when invalid.
func validateAllDayHalf(w http.ResponseWriter, allDay, half bool) bool {
	if allDay && half {
		writeError(w, http.StatusBadRequest, "Please check either all day or half day, not both", CodeTimeEntryAllDayHalf)
		return false
	}
	return true
}

// List returns one page of the connected user's own time entries for the
// org - this is the personal time tracker screen, not a management view, so
// it's scoped to the caller regardless of role. Admins/owners see every
// member's entries through the reports screen instead (see
// ReportHandler.Detailed/Summary).
func (h *TimeEntryHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID, _ := middleware.UserIDFromContext(r.Context())
	page, pageSize := paginationParams(r)

	entries, hasMore, err := h.entries.List(r.Context(), orgID, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":    entries,
		"page":     page,
		"pageSize": pageSize,
		"hasMore":  hasMore,
	})
}

// ListRange returns the caller's own time entries within a [start, end] day
// range, unpaginated - used by the Calendar view to load a whole visible
// month/week grid in one call instead of paging through List like the
// classic time tracker screen does.
func (h *TimeEntryHandler) ListRange(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID, _ := middleware.UserIDFromContext(r.Context())
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	if utils.IsBlank(start) || utils.IsBlank(end) {
		writeError(w, http.StatusBadRequest, "Please provide start and end query parameters", CodeInvalidRequestBody)
		return
	}

	entries, err := h.entries.ListByRange(r.Context(), orgID, userID, start, end)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": entries})
}

func (h *TimeEntryHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID, _ := middleware.UserIDFromContext(r.Context())

	var p timeEntryPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", CodeInvalidRequestBody)
		return
	}
	if utils.IsBlank(p.Text) || utils.IsBlank(p.ClientID) || utils.IsBlank(p.ProjectID) || utils.IsBlank(p.Day) {
		writeError(w, http.StatusBadRequest, "Please add text, clientId, projectId and day fields", CodeTimeEntryFields)
		return
	}
	if !validateAllDayHalf(w, p.AllDay, p.Half) {
		return
	}
	if !p.AllDay && !p.Half && (p.Start == nil || p.End == nil || utils.IsBlank(*p.Start) || utils.IsBlank(*p.End)) {
		writeError(w, http.StatusBadRequest, "Please add start and end fields, or check allDay/half", CodeTimeEntryStartEnd)
		return
	}

	entry, err := h.entries.Create(r.Context(), orgID, p.ClientID, p.ProjectID, userID, p.toFields())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	h.record(r, models.AuditRecordCreated, models.AuditSubjectRecord, entry.ID,
		auditRecordDetail(entry, h.entries.Labels(r.Context(), entry)))
	writeJSON(w, http.StatusCreated, entry)
}

func (h *TimeEntryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	entry, err := h.entries.FindByID(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !canManage(r, entry.UserID) {
		writeError(w, http.StatusForbidden, "You can only update your own time entries", CodeOwnTimeEntriesOnly)
		return
	}

	p := timeEntryPayload{
		ClientID: entry.ClientID, ProjectID: entry.ProjectID,
		Text: entry.Text, Day: entry.Day,
		Start: entry.Start, End: entry.End, AllDay: entry.AllDay, Half: entry.Half,
	}
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", CodeInvalidRequestBody)
		return
	}
	if !validateAllDayHalf(w, p.AllDay, p.Half) {
		return
	}

	// Changing which client/project an entry belongs to is allowed for
	// whoever can already manage the entry (its own owner, or an
	// admin/owner) - same bar as the rest of its fields. Only reassigning it
	// to a *different member* needs the stricter admin/owner-only check
	// below, since that moves the entry out from under its current owner.
	reassign := store.TimeEntryReassignment{}
	if p.ClientID != entry.ClientID {
		reassign.ClientID = p.ClientID
	}
	if p.ProjectID != entry.ProjectID {
		reassign.ProjectID = p.ProjectID
	}
	if utils.IsNotBlank(p.UserID) && p.UserID != entry.UserID {
		if !isAdminOrOwner(r) {
			writeError(w, http.StatusForbidden, "Only an admin or the owner can reassign a time entry to another member", CodeReassignForbidden)
			return
		}
		reassign.UserID = p.UserID
	}

	updated, err := h.entries.Update(r.Context(), id, reassign, p.toFields())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	detail := auditRecordDetail(updated, h.entries.Labels(r.Context(), updated))
	// Moving an entry to another member moves billable time off one person's
	// sheet and onto another's, so the trail says where it came from. Looked
	// up only when that actually happened: the ordinary edit is a time or
	// text change and shouldn't pay for a second query.
	if utils.IsNotBlank(reassign.UserID) {
		detail["previousUser"] = h.entries.Labels(r.Context(), entry).UserEmail
	}
	h.record(r, models.AuditRecordUpdated, models.AuditSubjectRecord, id, detail)
	writeJSON(w, http.StatusOK, updated)
}

func (h *TimeEntryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	entry, err := h.entries.FindByID(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !canManage(r, entry.UserID) {
		writeError(w, http.StatusForbidden, "You can only delete your own time entries", CodeOwnTimeEntriesOnly)
		return
	}

	if err := h.entries.Delete(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	// entry was loaded above for the ownership check, so the deleted day and
	// text are still in hand to describe what went.
	h.record(r, models.AuditRecordDeleted, models.AuditSubjectRecord, id,
		auditRecordDetail(entry, h.entries.Labels(r.Context(), entry)))
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}
