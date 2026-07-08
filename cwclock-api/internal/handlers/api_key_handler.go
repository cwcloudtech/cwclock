package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type ApiKeyHandler struct {
	keys *store.ApiKeyStore
}

func NewApiKeyHandler(keys *store.ApiKeyStore) *ApiKeyHandler {
	return &ApiKeyHandler{keys: keys}
}

// List returns the connected user's own API keys - this is self-service,
// there is no admin view over other users' keys.
func (h *ApiKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	keys, err := h.keys.ListByUser(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, keys)
}

type apiKeyPayload struct {
	Description string  `json:"description"`
	ExpiresAt   *string `json:"expiresAt"`
}

func (h *ApiKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var p apiKeyPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", CodeInvalidRequestBody)
		return
	}
	if utils.IsBlank(p.Description) {
		writeError(w, http.StatusBadRequest, "Please add a description", CodeApiKeyDescription)
		return
	}

	var expiresAt *time.Time
	if p.ExpiresAt != nil && utils.IsNotBlank(*p.ExpiresAt) {
		t, err := time.Parse(time.RFC3339, *p.ExpiresAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid expiration date", CodeInvalidExpiration)
			return
		}
		expiresAt = &t
	}

	key, token, err := h.keys.Create(r.Context(), userID, p.Description, expiresAt)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, models.ApiKeyCreated{
		ID:          key.ID,
		Description: key.Description,
		ExpiresAt:   key.ExpiresAt,
		CreatedAt:   key.CreatedAt,
		Token:       token,
	})
}

func (h *ApiKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.keys.Delete(r.Context(), userID, id); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}
