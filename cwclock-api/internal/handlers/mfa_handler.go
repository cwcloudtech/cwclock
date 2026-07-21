package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"

	"cwclock-api/internal/authtoken"
	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

// webauthnCeremonyTTL bounds how long a client has to complete a WebAuthn
// registration or login ceremony once "begin" hands back its options - the
// same order of magnitude as mfaLoginTokenTTL.
const webauthnCeremonyTTL = 5 * time.Minute

// MFAHandler implements MFA login verification and self-service enrollment
// (TOTP authenticator apps and WebAuthn security keys - see ai-instruct-68).
// Password login itself (and the initial MFA challenge) stays in
// UserHandler.Login; this handler only covers what happens once that
// challenge has been issued, plus enrollment management.
type MFAHandler struct {
	users          *store.UserStore
	webauthnCreds  *store.WebAuthnCredentialStore
	jwtSecret      string
	activationMode string
	wa             *webauthn.WebAuthn
	issuer         string
}

func NewMFAHandler(users *store.UserStore, webauthnCreds *store.WebAuthnCredentialStore, jwtSecret, activationMode string, wa *webauthn.WebAuthn, issuer string) *MFAHandler {
	return &MFAHandler{
		users: users, webauthnCreds: webauthnCreds, jwtSecret: jwtSecret,
		activationMode: activationMode, wa: wa, issuer: issuer,
	}
}

// webauthnUser adapts a models.User (plus its already-loaded credentials)
// to the webauthn.User interface the go-webauthn library operates on.
type webauthnUser struct {
	id          string
	email       string
	credentials []webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte                         { return []byte(u.id) }
func (u *webauthnUser) WebAuthnName() string                       { return u.email }
func (u *webauthnUser) WebAuthnDisplayName() string                { return u.email }
func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

func (h *MFAHandler) loadWebAuthnUser(ctx context.Context, user models.User) (*webauthnUser, error) {
	rows, err := h.webauthnCreds.ListByUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	creds := make([]webauthn.Credential, len(rows))
	for i, row := range rows {
		transports := make([]protocol.AuthenticatorTransport, len(row.Transports))
		for j, t := range row.Transports {
			transports[j] = protocol.AuthenticatorTransport(t)
		}
		creds[i] = webauthn.Credential{
			ID:            row.CredentialID,
			PublicKey:     row.PublicKey,
			Transport:     transports,
			Authenticator: webauthn.Authenticator{SignCount: row.SignCount},
		}
	}
	return &webauthnUser{id: user.ID, email: user.Email, credentials: creds}, nil
}

// respondSession finishes a login (password + verified second factor) by
// minting the real session token, the same response shape UserHandler.Login
// returns when MFA isn't enabled.
func (h *MFAHandler) respondSession(w http.ResponseWriter, user models.User) {
	token, err := authtoken.Generate(h.jwtSecret, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
		return
	}
	writeJSON(w, http.StatusOK, models.UserResponse{
		ID: user.ID, Email: user.Email, Name: user.Name, Surname: user.Surname,
		Role: user.Role, Token: token, Picture: user.Picture,
		PictureX: user.PictureX, PictureY: user.PictureY,
		I18nCode: models.I18nCodeForRole(user.Role, h.activationMode),
	})
}

// userFromMFAChallenge resolves the user a challengeToken (minted by
// UserHandler.Login) was issued for, rejecting anything expired, tampered
// with, or not actually a login challenge.
func (h *MFAHandler) userFromMFAChallenge(r *http.Request, challengeToken string) (models.User, bool) {
	userID, err := authtoken.ParsePurpose(h.jwtSecret, challengeToken, authtoken.PurposeMFALogin)
	if err != nil {
		return models.User{}, false
	}
	user, err := h.users.FindByID(r.Context(), userID)
	return user, err == nil
}

type mfaTOTPLoginPayload struct {
	ChallengeToken string `json:"challengeToken"`
	Code           string `json:"code"`
}

// LoginTOTP finishes a password login gated by MFA using an authenticator
// app code.
func (h *MFAHandler) LoginTOTP(w http.ResponseWriter, r *http.Request) {
	var p mfaTOTPLoginPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.ChallengeToken) || utils.IsBlank(p.Code) {
		writeError(w, http.StatusBadRequest, "Please add a challengeToken and code", CodeInvalidRequestBody)
		return
	}

	user, ok := h.userFromMFAChallenge(r, p.ChallengeToken)
	if !ok || utils.IsBlank(user.MFATOTPSecret) {
		writeError(w, http.StatusUnauthorized, "This login challenge is invalid or has expired", CodeInvalidToken)
		return
	}

	if !totp.Validate(p.Code, user.MFATOTPSecret) {
		writeError(w, http.StatusUnauthorized, "Invalid code", CodeInvalidMFACode)
		return
	}

	h.respondSession(w, user)
}

type mfaChallengeTokenPayload struct {
	ChallengeToken string `json:"challengeToken"`
}

// LoginWebAuthnBegin starts the assertion ceremony for finishing a password
// login gated by MFA using a registered security key.
func (h *MFAHandler) LoginWebAuthnBegin(w http.ResponseWriter, r *http.Request) {
	var p mfaChallengeTokenPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.ChallengeToken) {
		writeError(w, http.StatusBadRequest, "Please add a challengeToken", CodeInvalidRequestBody)
		return
	}

	user, ok := h.userFromMFAChallenge(r, p.ChallengeToken)
	if !ok {
		writeError(w, http.StatusUnauthorized, "This login challenge is invalid or has expired", CodeInvalidToken)
		return
	}

	wu, err := h.loadWebAuthnUser(r.Context(), user)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	options, session, err := h.wa.BeginLogin(wu)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), CodeInvalidMFACode)
		return
	}

	ceremonyToken, err := authtoken.GeneratePurposeWithData(h.jwtSecret, user.ID, authtoken.PurposeWebAuthnCeremony, webauthnCeremonyTTL, session)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ceremonyToken": ceremonyToken, "options": options})
}

type mfaWebAuthnFinishPayload struct {
	CeremonyToken string          `json:"ceremonyToken"`
	Credential    json.RawMessage `json:"credential"`
}

// LoginWebAuthnFinish verifies the security key's assertion and, on success,
// completes the login.
func (h *MFAHandler) LoginWebAuthnFinish(w http.ResponseWriter, r *http.Request) {
	var p mfaWebAuthnFinishPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.CeremonyToken) || len(p.Credential) == 0 {
		writeError(w, http.StatusBadRequest, "Please add a ceremonyToken and credential", CodeInvalidRequestBody)
		return
	}

	var session webauthn.SessionData
	userID, err := authtoken.ParsePurposeWithData(h.jwtSecret, p.CeremonyToken, authtoken.PurposeWebAuthnCeremony, &session)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "This security key challenge is invalid or has expired", CodeInvalidToken)
		return
	}

	user, err := h.users.FindByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid credentials", CodeInvalidCredentials)
		return
	}

	wu, err := h.loadWebAuthnUser(r.Context(), user)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	parsed, err := protocol.ParseCredentialRequestResponseBytes(p.Credential)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid security key response", CodeInvalidMFACode)
		return
	}

	cred, err := h.wa.ValidateLogin(wu, session, parsed)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Security key verification failed", CodeInvalidMFACode)
		return
	}

	if err := h.webauthnCreds.UpdateSignCount(r.Context(), cred.ID, cred.Authenticator.SignCount); err != nil {
		writeStoreError(w, err)
		return
	}

	h.respondSession(w, user)
}

// Status reports the connected user's current MFA enrollment, for the
// self-service security settings screen.
func (h *MFAHandler) Status(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	user, err := h.users.FindByID(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	creds, err := h.webauthnCreds.ListByUser(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"totpEnabled":         user.MFAEnabled && utils.IsNotBlank(user.MFATOTPSecret),
		"webauthnCredentials": creds,
	})
}

// TOTPSetup generates a fresh TOTP secret for the connected user (stored as
// "pending" - see UserStore.SetPendingTOTPSecret) and returns everything
// needed to enroll it: the raw secret, its otpauth:// URI, and a QR code PNG
// rendering that URI for scanning with an authenticator app. MFA isn't
// turned on until TOTPConfirm verifies a code generated from this secret.
func (h *MFAHandler) TOTPSetup(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	user, err := h.users.FindByID(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{Issuer: h.issuer, AccountName: user.Email})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
		return
	}

	if _, err := h.users.SetPendingTOTPSecret(r.Context(), userID, key.Secret()); err != nil {
		writeStoreError(w, err)
		return
	}

	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"secret":     key.Secret(),
		"otpauthUrl": key.URL(),
		"qrCodePng":  "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
	})
}

type mfaCodePayload struct {
	Code string `json:"code"`
}

// TOTPConfirm turns on MFA once the user proves they scanned/enrolled the
// pending secret from TOTPSetup by submitting a code it currently generates.
func (h *MFAHandler) TOTPConfirm(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var p mfaCodePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.Code) {
		writeError(w, http.StatusBadRequest, "Please add a code", CodeInvalidRequestBody)
		return
	}

	user, err := h.users.FindByID(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if utils.IsBlank(user.MFATOTPSecret) {
		writeError(w, http.StatusBadRequest, "Start a TOTP setup before confirming it", CodeInvalidMFACode)
		return
	}
	if !totp.Validate(p.Code, user.MFATOTPSecret) {
		writeError(w, http.StatusBadRequest, "Invalid code", CodeInvalidMFACode)
		return
	}

	if _, err := h.users.ConfirmTOTP(r.Context(), userID); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"totpEnabled": true})
}

// TOTPDisable turns off the authenticator-app factor, leaving MFA enabled
// overall if the user still has at least one registered security key.
func (h *MFAHandler) TOTPDisable(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	webauthnCount, err := h.webauthnCreds.CountByUser(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	if _, err := h.users.DisableTOTP(r.Context(), userID, webauthnCount > 0); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"totpEnabled": false})
}

// WebAuthnRegisterBegin starts registering a new security key for the
// connected user.
func (h *MFAHandler) WebAuthnRegisterBegin(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	user, err := h.users.FindByID(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	wu, err := h.loadWebAuthnUser(r.Context(), user)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	options, session, err := h.wa.BeginRegistration(wu)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
		return
	}

	ceremonyToken, err := authtoken.GeneratePurposeWithData(h.jwtSecret, user.ID, authtoken.PurposeWebAuthnCeremony, webauthnCeremonyTTL, session)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ceremonyToken": ceremonyToken, "options": options})
}

type mfaWebAuthnRegisterFinishPayload struct {
	CeremonyToken string          `json:"ceremonyToken"`
	Credential    json.RawMessage `json:"credential"`
	Name          string          `json:"name"`
}

// WebAuthnRegisterFinish verifies the new security key's attestation, stores
// it and turns MFA on.
func (h *MFAHandler) WebAuthnRegisterFinish(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var p mfaWebAuthnRegisterFinishPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.CeremonyToken) || len(p.Credential) == 0 {
		writeError(w, http.StatusBadRequest, "Please add a ceremonyToken and credential", CodeInvalidRequestBody)
		return
	}

	var session webauthn.SessionData
	tokenUserID, err := authtoken.ParsePurposeWithData(h.jwtSecret, p.CeremonyToken, authtoken.PurposeWebAuthnCeremony, &session)
	if err != nil || tokenUserID != userID {
		writeError(w, http.StatusUnauthorized, "This security key challenge is invalid or has expired", CodeInvalidToken)
		return
	}

	user, err := h.users.FindByID(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	wu, err := h.loadWebAuthnUser(r.Context(), user)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	parsed, err := protocol.ParseCredentialCreationResponseBytes(p.Credential)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid security key response", CodeInvalidMFACode)
		return
	}

	cred, err := h.wa.CreateCredential(wu, session, parsed)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Security key registration failed", CodeInvalidMFACode)
		return
	}

	transports := make([]string, len(cred.Transport))
	for i, t := range cred.Transport {
		transports[i] = string(t)
	}

	name := p.Name
	if utils.IsBlank(name) {
		name = "Security key"
	}

	stored, err := h.webauthnCreds.Create(r.Context(), userID, cred.ID, cred.PublicKey, transports, name)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	if _, err := h.users.SetMFAEnabled(r.Context(), userID, true); err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, stored)
}

// WebAuthnDelete revokes one of the connected user's registered security
// keys, turning MFA back off entirely if it was the last factor left.
func (h *MFAHandler) WebAuthnDelete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "credentialId")

	if err := h.webauthnCreds.Delete(r.Context(), userID, id); err != nil {
		writeStoreError(w, err)
		return
	}

	user, err := h.users.FindByID(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	remaining, err := h.webauthnCreds.CountByUser(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if remaining == 0 && utils.IsBlank(user.MFATOTPSecret) {
		if _, err := h.users.SetMFAEnabled(r.Context(), userID, false); err != nil {
			writeStoreError(w, err)
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}
