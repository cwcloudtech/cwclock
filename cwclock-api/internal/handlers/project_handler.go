package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type ProjectHandler struct {
	auditor
	projects *store.ProjectStore
	clients  *store.ClientStore
}

func NewProjectHandler(projects *store.ProjectStore, clients *store.ClientStore, audit *store.AuditStore) *ProjectHandler {
	return &ProjectHandler{auditor: newAuditor(audit), projects: projects, clients: clients}
}

// auditProjectDetail names a project and the client it bills to. The client
// is named rather than referenced: an id means nothing on the audit screen,
// and deleting a client takes its projects with it, so the id would be
// unresolvable exactly when the entry matters. The daily rate is in there
// because it is what the client is invoiced at - changing it changes money.
func (h *ProjectHandler) auditProjectDetail(ctx context.Context, p models.Project) map[string]any {
	detail := map[string]any{"name": p.Name}
	if client, err := h.clients.FindByID(ctx, p.ClientID); err == nil {
		detail["client"] = client.Name
	}
	if p.DailyRate != nil {
		detail["dailyRate"] = *p.DailyRate
	}
	return detail
}

type projectPayload struct {
	ClientID     string   `json:"clientId"`
	Name         string   `json:"name"`
	Color        string   `json:"color"`
	DailyRate    *float64 `json:"dailyRate"`
	Subdivisions []string `json:"subdivisions"`
}

// redactProjectRates hides each project's daily rate from the response
// unless the caller is an admin or the owner, matching redactRates for
// member rates: readers/members can list projects to log time against, but
// the billing rate stays admin/owner-only.
func redactProjectRates(r *http.Request, projects []models.Project) []models.Project {
	role, _ := middleware.OrgRoleFromContext(r.Context())
	if role == models.RoleAdmin || role == models.RoleOwner {
		return projects
	}
	redacted := make([]models.Project, len(projects))
	for i, p := range projects {
		p.DailyRate = nil
		redacted[i] = p
	}
	return redacted
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	clientID := chi.URLParam(r, "clientId")
	clientID = utils.If(utils.IsBlank(clientID), r.URL.Query().Get("clientId"), clientID)

	projects, err := h.projects.List(r.Context(), orgID, clientID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, redactProjectRates(r, projects))
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	clientID := chi.URLParam(r, "clientId")

	var p projectPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.Name) {
		writeError(w, http.StatusBadRequest, "Please add a name field", CodeNameRequired)
		return
	}

	project, err := h.projects.Create(r.Context(), orgID, clientID, p.Name, p.Color, p.DailyRate, p.Subdivisions)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	h.record(r, models.AuditProjectCreated, models.AuditSubjectProject, project.ID,
		h.auditProjectDetail(r.Context(), project))
	writeJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	id := chi.URLParam(r, "projectId")

	var p projectPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.Name) {
		writeError(w, http.StatusBadRequest, "Please add a name field", CodeNameRequired)
		return
	}

	if utils.IsNotBlank(p.ClientID) {
		client, err := h.clients.FindByID(r.Context(), p.ClientID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		if client.OrganizationID != orgID {
			writeError(w, http.StatusBadRequest, "That client does not belong to this organization", CodeInvalidClientForOrg)
			return
		}
	}

	project, err := h.projects.Update(r.Context(), id, p.ClientID, p.Name, p.Color, p.DailyRate, p.Subdivisions)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	h.record(r, models.AuditProjectUpdated, models.AuditSubjectProject, id,
		h.auditProjectDetail(r.Context(), project))
	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectId")

	// Described before the delete, like a client's: afterwards nothing is
	// left to say what was removed, and deleting a project takes every hour
	// logged against it with it (ON DELETE CASCADE in V1__init.sql).
	detail := map[string]any{}
	if project, err := h.projects.FindByID(r.Context(), id); err == nil {
		detail = h.auditProjectDetail(r.Context(), project)
	}

	if err := h.projects.Delete(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	h.record(r, models.AuditProjectDeleted, models.AuditSubjectProject, id, detail)
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}
