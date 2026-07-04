package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
)

type TimeEntryHandler struct {
	entries *store.TimeEntryStore
}

func NewTimeEntryHandler(entries *store.TimeEntryStore) *TimeEntryHandler {
	return &TimeEntryHandler{entries: entries}
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
}

func (p timeEntryPayload) toFields() store.TimeEntryFields {
	return store.TimeEntryFields{
		Text:   p.Text,
		Day:    p.Day,
		Start:  p.Start,
		End:    p.End,
		AllDay: p.AllDay,
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

func (h *TimeEntryHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID, _ := middleware.UserIDFromContext(r.Context())
	role, _ := middleware.OrgRoleFromContext(r.Context())

	filterUserID := userID
	if role == models.RoleAdmin || role == models.RoleOwner {
		filterUserID = ""
	}

	entries, err := h.entries.List(r.Context(), orgID, filterUserID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *TimeEntryHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID, _ := middleware.UserIDFromContext(r.Context())

	var p timeEntryPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if p.Text == "" || p.ClientID == "" || p.ProjectID == "" || p.Day == "" {
		writeError(w, http.StatusBadRequest, "Please add text, clientId, projectId and day fields")
		return
	}
	if !p.AllDay && (p.Start == nil || p.End == nil || *p.Start == "" || *p.End == "") {
		writeError(w, http.StatusBadRequest, "Please add start and end fields, or check allDay")
		return
	}

	entry, err := h.entries.Create(r.Context(), orgID, p.ClientID, p.ProjectID, userID, p.toFields())
	if err != nil {
		writeStoreError(w, err)
		return
	}
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
		writeError(w, http.StatusForbidden, "You can only update your own time entries")
		return
	}

	p := timeEntryPayload{
		ClientID: entry.ClientID, ProjectID: entry.ProjectID,
		Text: entry.Text, Day: entry.Day,
		Start: entry.Start, End: entry.End, AllDay: entry.AllDay,
	}
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	reassign := store.TimeEntryReassignment{}
	if p.ClientID != entry.ClientID {
		reassign.ClientID = p.ClientID
	}
	if p.ProjectID != entry.ProjectID {
		reassign.ProjectID = p.ProjectID
	}
	if p.UserID != "" && p.UserID != entry.UserID {
		if !isAdminOrOwner(r) {
			writeError(w, http.StatusForbidden, "Only an admin or the owner can reassign a time entry to another member")
			return
		}
		reassign.UserID = p.UserID
	}

	updated, err := h.entries.Update(r.Context(), id, reassign, p.toFields())
	if err != nil {
		writeStoreError(w, err)
		return
	}
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
		writeError(w, http.StatusForbidden, "You can only delete your own time entries")
		return
	}

	if err := h.entries.Delete(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}
