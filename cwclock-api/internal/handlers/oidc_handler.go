package handlers

import (
	"log/slog"
	"net/http"
	"net/url"
	"slices"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/authtoken"
	"cwclock-api/internal/oidc"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type OIDCHandler struct {
	providers      []oidc.Provider
	users          *store.UserStore
	jwtSecret      string
	apiBaseURL     string
	uiBaseURL      string
	keycloakGroups []string
}

func NewOIDCHandler(providers []oidc.Provider, users *store.UserStore, jwtSecret, apiBaseURL, uiBaseURL string, keycloakGroups []string) *OIDCHandler {
	return &OIDCHandler{
		providers:      providers,
		users:          users,
		jwtSecret:      jwtSecret,
		apiBaseURL:     apiBaseURL,
		uiBaseURL:      uiBaseURL,
		keycloakGroups: keycloakGroups,
	}
}

type oidcProvidersResponse struct {
	Providers []string `json:"providers"`
}

// ListProviders is unauthenticated: the frontend calls it to know which
// login/signup buttons to display.
func (h *OIDCHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, oidcProvidersResponse{Providers: oidc.Names(h.providers)})
}

func (h *OIDCHandler) redirectURI(provider string) string {
	return h.apiBaseURL + "/v1/oidc/" + provider + "/callback"
}

// Login starts the flow for a provider by redirecting the browser to its
// authorization endpoint.
func (h *OIDCHandler) Login(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "provider")
	provider, ok := oidc.Find(h.providers, name)
	if !ok {
		writeError(w, http.StatusNotFound, "Unknown or disabled OIDC provider", CodeNotFound)
		return
	}

	state, err := oidc.SignState(h.jwtSecret, provider.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
		return
	}

	http.Redirect(w, r, provider.AuthorizationURL(h.redirectURI(provider.Name), state), http.StatusFound)
}

// Callback finishes the flow: it exchanges the code, resolves the identity,
// finds or creates the matching local account, and hands control back to the
// frontend with a session token (or an error flag) in the redirect's query
// string, since the browser only ever talks to the frontend origin.
func (h *OIDCHandler) Callback(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "provider")
	provider, ok := oidc.Find(h.providers, name)
	if !ok {
		writeError(w, http.StatusNotFound, "Unknown or disabled OIDC provider", CodeNotFound)
		return
	}

	if errParam := r.URL.Query().Get("error"); utils.IsNotBlank(errParam) {
		h.redirectWithError(w, r, "oidc_denied")
		return
	}

	state := r.URL.Query().Get("state")
	if err := oidc.VerifyState(h.jwtSecret, provider.Name, state); err != nil {
		h.redirectWithError(w, r, "oidc_invalid_state")
		return
	}

	code := r.URL.Query().Get("code")
	if utils.IsBlank(code) {
		h.redirectWithError(w, r, "oidc_missing_code")
		return
	}

	accessToken, err := oidc.ExchangeCode(r.Context(), provider, code, h.redirectURI(provider.Name))
	if err != nil {
		slog.Error("oidc token exchange failed", "provider", provider.Name, "error", err)
		h.redirectWithError(w, r, "oidc_exchange_failed")
		return
	}

	identity, err := oidc.FetchIdentity(r.Context(), provider, accessToken)
	if err != nil {
		slog.Error("oidc identity fetch failed", "provider", provider.Name, "error", err)
		h.redirectWithError(w, r, "oidc_identity_failed")
		return
	}

	if provider.Name == oidc.Keycloak && len(h.keycloakGroups) > 0 && !hasAllowedGroup(identity.Groups, h.keycloakGroups) {
		h.redirectWithError(w, r, "oidc_forbidden_group")
		return
	}

	user, err := h.users.FindOrCreateOIDC(r.Context(), identity.Email, identity.Name, identity.Surname)
	if err != nil {
		slog.Error("oidc user lookup/creation failed", "provider", provider.Name, "error", err)
		h.redirectWithError(w, r, "oidc_account_failed")
		return
	}

	token, err := authtoken.Generate(h.jwtSecret, user.ID)
	if err != nil {
		slog.Error("oidc token generation failed", "provider", provider.Name, "error", err)
		h.redirectWithError(w, r, "oidc_account_failed")
		return
	}

	target := h.uiBaseURL + "/oidc/callback?token=" + url.QueryEscape(token)
	http.Redirect(w, r, target, http.StatusFound)
}

func (h *OIDCHandler) redirectWithError(w http.ResponseWriter, r *http.Request, code string) {
	target := h.uiBaseURL + "/oidc/callback?error=" + url.QueryEscape(code)
	http.Redirect(w, r, target, http.StatusFound)
}

func hasAllowedGroup(userGroups, allowed []string) bool {
	for _, g := range userGroups {
		if slices.Contains(allowed, g) {
			return true
		}
	}
	return false
}
