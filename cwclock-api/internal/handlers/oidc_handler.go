package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"slices"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/authtoken"
	"cwclock-api/internal/models"
	"cwclock-api/internal/oidc"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

// originParam is set by the frontend to ask Login for a frontend-bound
// redirect_uri instead of the API-bound default - see Login and
// FrontendCallback.
const originParam = "origin"
const originFrontend = "frontend"

type OIDCHandler struct {
	providers      []oidc.Provider
	users          *store.UserStore
	webauthnCreds  *store.WebAuthnCredentialStore
	jwtSecret      string
	apiBaseURL     string
	uiBaseURL      string
	keycloakGroups []string
	activationMode string
}

func NewOIDCHandler(providers []oidc.Provider, users *store.UserStore, webauthnCreds *store.WebAuthnCredentialStore, jwtSecret, apiBaseURL, uiBaseURL string, keycloakGroups []string, activationMode string) *OIDCHandler {
	return &OIDCHandler{
		providers:      providers,
		users:          users,
		webauthnCreds:  webauthnCreds,
		jwtSecret:      jwtSecret,
		apiBaseURL:     apiBaseURL,
		uiBaseURL:      uiBaseURL,
		keycloakGroups: keycloakGroups,
		activationMode: activationMode,
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

// redirectURI is the value sent to the provider as redirect_uri and later
// echoed back to it during the code exchange (the two must match exactly).
// It defaults to the API's own callback route, which is what lets this
// handler do the exchange itself before handing off to the frontend; when
// frontend is true it points at the frontend's /oidc/callback route instead,
// for callers that asked for that via originParam (see Login).
func (h *OIDCHandler) redirectURI(provider string, frontend bool) string {
	if frontend {
		return h.uiBaseURL + "/oidc/callback"
	}
	return h.apiBaseURL + "/v1/oidc/" + provider + "/callback"
}

// Login starts the flow for a provider by redirecting the browser to its
// authorization endpoint. By default redirect_uri points at the API's own
// callback; if the caller passes ?origin=frontend it points at the
// frontend's /oidc/callback instead, so the provider redirects the browser
// there directly once the user consents.
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

	frontend := r.URL.Query().Get(originParam) == originFrontend
	authURL := provider.AuthorizationURL(h.redirectURI(provider.Name, frontend), state)

	http.Redirect(w, r, authURL, http.StatusFound)
}

// completeLogin exchanges code for an access token, resolves the identity it
// belongs to, and finds or creates the matching local account, returning it
// for the caller to turn into a session via finishSession. Shared by
// Callback and FrontendCallback, which differ only in which redirect_uri
// they exchange against and how they report the result back to the browser.
func (h *OIDCHandler) completeLogin(ctx context.Context, provider oidc.Provider, code, redirectURI string) (models.User, string) {
	accessToken, err := oidc.ExchangeCode(ctx, provider, code, redirectURI)
	if err != nil {
		slog.Error("oidc token exchange failed", "provider", provider.Name, "error", err)
		return models.User{}, "oidc_exchange_failed"
	}

	identity, err := oidc.FetchIdentity(ctx, provider, accessToken)
	if err != nil {
		slog.Error("oidc identity fetch failed", "provider", provider.Name, "error", err)
		return models.User{}, "oidc_identity_failed"
	}

	if provider.Name == oidc.Keycloak && len(h.keycloakGroups) > 0 && !hasAllowedGroup(identity.Groups, h.keycloakGroups) {
		return models.User{}, "oidc_forbidden_group"
	}

	user, err := h.users.FindOrCreateOIDC(ctx, identity.Email, identity.Name, identity.Surname, h.activationMode)
	if err != nil {
		slog.Error("oidc user lookup/creation failed", "provider", provider.Name, "error", err)
		return models.User{}, "oidc_account_failed"
	}

	return user, utils.EMPTY
}

// finishSession turns a resolved OIDC user into a real session token, unless
// the account has MFA enabled, in which case OIDC has only proven the first
// factor: it mints the same short-lived challenge UserHandler.Login would,
// leaving the caller to report it back so the client can clear it via
// LoginTOTP/LoginWebAuthnFinish exactly as it would after a password login.
func (h *OIDCHandler) finishSession(ctx context.Context, user models.User) (string, *models.MFAChallengeResponse, error) {
	if !user.MFAEnabled {
		token, err := authtoken.Generate(h.jwtSecret, user.ID)
		return token, nil, err
	}

	challengeToken, err := authtoken.GeneratePurpose(h.jwtSecret, user.ID, authtoken.PurposeMFALogin, mfaLoginTokenTTL)
	if err != nil {
		return utils.EMPTY, nil, err
	}
	webauthnCount, err := h.webauthnCreds.CountByUser(ctx, user.ID)
	if err != nil {
		return utils.EMPTY, nil, err
	}

	return utils.EMPTY, &models.MFAChallengeResponse{
		MFARequired:    true,
		ChallengeToken: challengeToken,
		HasTOTP:        utils.IsNotBlank(user.MFATOTPSecret),
		HasWebAuthn:    webauthnCount > 0,
	}, nil
}

// Callback finishes the flow when the provider redirects the browser
// straight back to the API (the default, API-bound redirect_uri from
// Login): it completes the login and hands control back to the frontend
// with a session token, an MFA challenge (see finishSession) or an error
// flag in the redirect's query string, since the browser only ever talks to
// the frontend origin.
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

	user, errCode := h.completeLogin(r.Context(), provider, code, h.redirectURI(provider.Name, false))
	if utils.IsNotBlank(errCode) {
		h.redirectWithError(w, r, errCode)
		return
	}

	token, challenge, err := h.finishSession(r.Context(), user)
	if err != nil {
		h.redirectWithError(w, r, "oidc_account_failed")
		return
	}

	if challenge != nil {
		target := h.uiBaseURL + "/oidc/callback?mfaChallenge=" + url.QueryEscape(challenge.ChallengeToken) +
			"&hasTotp=" + queryBool(challenge.HasTOTP) + "&hasWebAuthn=" + queryBool(challenge.HasWebAuthn)
		http.Redirect(w, r, target, http.StatusFound)
		return
	}

	target := h.uiBaseURL + "/oidc/callback?token=" + url.QueryEscape(token)
	http.Redirect(w, r, target, http.StatusFound)
}

// FrontendCallback finishes the flow when Login handed the frontend a
// frontend-bound redirect_uri (via originParam): the provider redirects the
// browser to the frontend's own /oidc/callback route with code/state, and
// the frontend calls this endpoint to complete the exchange and get a
// session token (or, per finishSession, an MFA challenge) back as JSON,
// since the browser already left the API's origin and there's nothing left
// to redirect. The provider isn't in the path here (redirect_uri is
// provider-agnostic on the frontend side), so it comes from the state
// parameter instead.
func (h *OIDCHandler) FrontendCallback(w http.ResponseWriter, r *http.Request) {
	if errParam := r.URL.Query().Get("error"); utils.IsNotBlank(errParam) {
		writeError(w, http.StatusUnauthorized, "OIDC login was denied", CodeInternal)
		return
	}

	state := r.URL.Query().Get("state")
	name, err := oidc.ProviderFromState(state)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid OIDC state", CodeInternal)
		return
	}

	provider, ok := oidc.Find(h.providers, name)
	if !ok {
		writeError(w, http.StatusNotFound, "Unknown or disabled OIDC provider", CodeNotFound)
		return
	}

	if err := oidc.VerifyState(h.jwtSecret, provider.Name, state); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid OIDC state", CodeInternal)
		return
	}

	code := r.URL.Query().Get("code")
	if utils.IsBlank(code) {
		writeError(w, http.StatusBadRequest, "Missing OIDC code", CodeInternal)
		return
	}

	user, errCode := h.completeLogin(r.Context(), provider, code, h.redirectURI(provider.Name, true))
	if utils.IsNotBlank(errCode) {
		writeError(w, http.StatusBadGateway, errCode, CodeInternal)
		return
	}

	token, challenge, err := h.finishSession(r.Context(), user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
		return
	}

	if challenge != nil {
		writeJSON(w, http.StatusOK, challenge)
		return
	}

	writeJSON(w, http.StatusOK, oidcTokenResponse{Token: token})
}

type oidcTokenResponse struct {
	Token string `json:"token"`
}

// queryBool renders a bool as the "1"/"0" the redirect-based Callback flow
// carries in its query string (there's no JSON body to send a real bool in).
func queryBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
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
