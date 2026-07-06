package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type ProjectHandler struct {
	projects *store.ProjectStore
}

func NewProjectHandler(projects *store.ProjectStore) *ProjectHandler {
	return &ProjectHandler{projects: projects}
}

type projectPayload struct {
	Name  string `json:"name"`
	Color string `json:"color"`
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
	writeJSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	clientID := chi.URLParam(r, "clientId")

	var p projectPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.Name) {
		writeError(w, http.StatusBadRequest, "Please add a name field", CodeNameRequired)
		return
	}

	project, err := h.projects.Create(r.Context(), orgID, clientID, p.Name, p.Color)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectId")

	var p projectPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.Name) {
		writeError(w, http.StatusBadRequest, "Please add a name field", CodeNameRequired)
		return
	}

	project, err := h.projects.Update(r.Context(), id, p.Name, p.Color)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectId")

	if err := h.projects.Delete(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}
