package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	qrcode "github.com/skip2/go-qrcode"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

// configPayload carries the plaintext API key token to embed in the
// downloaded CLI config/QR code, in the POST body rather than a custom
// header (see ai-instruct-103 - a header here forces a CORS exception on
// reverse proxies that a plain JSON POST doesn't). ApiKeyStore never
// persists or exposes an API key's plaintext once it's been created (see
// models.ApiKey), so the caller - the frontend, right after
// ApiKeyHandler.Create - is the only possible source for it; these
// endpoints just format it, never look it up.
type configPayload struct {
	Key    string `json:"key"`
	OrgaID string `json:"orga_id"`
}

func decodeConfigPayload(r *http.Request) (configPayload, error) {
	var p configPayload
	err := json.NewDecoder(r.Body).Decode(&p)
	return p, err
}

// ConfigHandler builds a CLI config file (and a QR code of the same
// content) that a user can import with `cwclock configure import`, from an
// API key token the caller already has in hand (see ai-instruct-102).
type ConfigHandler struct {
	orgs       *store.OrgStore
	apiBaseURL string
}

func NewConfigHandler(orgs *store.OrgStore, apiBaseURL string) *ConfigHandler {
	return &ConfigHandler{orgs: orgs, apiBaseURL: apiBaseURL}
}

// resolveOrgID picks org_id for the config: the requested organization if
// given (and the user actually belongs to it), otherwise the user's first
// (oldest) organization. ListForUser orders newest-first, so "first" is the
// last element. Returns "" if the user has no organizations at all.
func (h *ConfigHandler) resolveOrgID(ctx context.Context, userID, requestedOrgID string) (string, error) {
	orgs, err := h.orgs.ListForUser(ctx, userID)
	if err != nil {
		return utils.EMPTY, err
	}

	if utils.IsBlank(requestedOrgID) {
		if len(orgs) == 0 {
			return utils.EMPTY, nil
		}
		return orgs[len(orgs)-1].ID, nil
	}

	for _, o := range orgs {
		if o.ID == requestedOrgID {
			return o.ID, nil
		}
	}
	return utils.EMPTY, store.ErrNotFound
}

func (h *ConfigHandler) buildConfigText(r *http.Request, token, requestedOrgID string) (string, error) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	orgID, err := h.resolveOrgID(r.Context(), userID, requestedOrgID)
	if err != nil {
		return utils.EMPTY, err
	}

	text := fmt.Sprintf("api_url = %s\napi_key = %s\n", h.apiBaseURL, token)
	if utils.IsNotBlank(orgID) {
		text += fmt.Sprintf("org_id = %s\n", orgID)
	}
	return text, nil
}

// File returns a `~/.cwclock/config`-formatted file, importable as-is with
// `cwclock configure import <path>`.
func (h *ConfigHandler) File(w http.ResponseWriter, r *http.Request) {
	p, err := decodeConfigPayload(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", CodeInvalidRequestBody)
		return
	}
	if utils.IsBlank(p.Key) {
		writeError(w, http.StatusBadRequest, "Missing API key token", CodeConfigTokenRequired)
		return
	}

	text, err := h.buildConfigText(r, p.Key, p.OrgaID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="config"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(text))
}

// QR renders the same content as File into a QR code, base64-embedded the
// same way MFAHandler.TOTPSetup ships its enrollment QR code.
func (h *ConfigHandler) QR(w http.ResponseWriter, r *http.Request) {
	p, err := decodeConfigPayload(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", CodeInvalidRequestBody)
		return
	}
	if utils.IsBlank(p.Key) {
		writeError(w, http.StatusBadRequest, "Missing API key token", CodeConfigTokenRequired)
		return
	}

	text, err := h.buildConfigText(r, p.Key, p.OrgaID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	png, err := qrcode.Encode(text, qrcode.Medium, 256)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"qrCodePng": "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
	})
}
