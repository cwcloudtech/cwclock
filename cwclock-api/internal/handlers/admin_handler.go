package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type AdminHandler struct {
	auditor
	users         *store.UserStore
	webauthnCreds *store.WebAuthnCredentialStore
	audits        *store.AuditStore
	maxImageSize  int64
}

func NewAdminHandler(users *store.UserStore, webauthnCreds *store.WebAuthnCredentialStore, audit *store.AuditStore, maxImageSize int64) *AdminHandler {
	return &AdminHandler{
		auditor: newAuditor(audit),
		users:   users, webauthnCreds: webauthnCreds, audits: audit, maxImageSize: maxImageSize,
	}
}

func toUserMeResponse(u models.User) models.UserMeResponse {
	return models.UserMeResponse{
		ID:         u.ID,
		Email:      u.Email,
		Name:       u.Name,
		Surname:    u.Surname,
		Role:       u.Role,
		Picture:    u.Picture,
		PictureX:   u.PictureX,
		PictureY:   u.PictureY,
		MFAEnabled: u.MFAEnabled,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

// ListUsers returns every account, for the superuser's user-management
// screen - or, when one or more ?email= query params are given (repeatable,
// OR logic), only the accounts matching one of those emails
// case-insensitively, for the CLI's id-or-email fallback on
// `cwclock admin user` (see ai-instruct-99).
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	var err error
	if emails := r.URL.Query()["email"]; len(emails) > 0 {
		users, err = h.users.ListByEmails(r.Context(), emails)
	} else {
		users, err = h.users.List(r.Context())
	}
	if err != nil {
		writeStoreError(w, err)
		return
	}

	results := make([]models.UserMeResponse, len(users))
	for i, u := range users {
		results[i] = toUserMeResponse(u)
	}
	writeJSON(w, http.StatusOK, results)
}

func validGlobalRole(role string) bool {
	switch models.GlobalRole(role) {
	case models.GlobalRoleSuperuser, models.GlobalRoleConfirmed, models.GlobalRoleDisabled, models.GlobalRoleBan:
		return true
	default:
		return false
	}
}

type adminUpdateUserPayload struct {
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Surname  string   `json:"surname"`
	Role     string   `json:"role"`
	Password *string  `json:"password"`
	Picture  *string  `json:"picture"`
	PictureX *float64 `json:"pictureX"`
	PictureY *float64 `json:"pictureY"`
}

// UpdateUser lets the superuser edit any account: email, profile, role,
// avatar and optionally a new password. An already-set password is never
// returned to the client; sending a non-empty "password" simply overrides it.
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var p adminUpdateUserPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil ||
		utils.IsBlank(p.Email) || utils.IsBlank(p.Name) || utils.IsBlank(p.Surname) || !validGlobalRole(p.Role) {
		writeError(w, http.StatusBadRequest, "Please add valid email, name, surname and role fields", CodeInvalidAdminUserEdit)
		return
	}

	if p.Picture != nil && utils.ImageSizeExceeds(*p.Picture, h.maxImageSize) {
		writeError(w, http.StatusBadRequest, "Image is too large", CodeImageTooLarge)
		return
	}

	fields := store.AdminUserFields{Email: p.Email, Name: p.Name, Surname: p.Surname, Role: p.Role}

	if p.Password != nil && utils.IsNotBlank(*p.Password) {
		if ok, code := utils.IsPasswordValid(*p.Password); !ok {
			writeInvalidPassword(w, code)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*p.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error(), CodeInternal)
			return
		}
		hashed := string(hash)
		fields.PasswordHash = &hashed
	}
	if p.Picture != nil {
		fields.Picture = p.Picture
		fields.PictureX = p.PictureX
		fields.PictureY = p.PictureY
	}

	// The previous role has to be read before the update to say what changed.
	// A failed read only costs the trail the "from" value, so it is not worth
	// failing the edit over.
	var previousRole models.GlobalRole
	if before, err := h.users.FindByID(r.Context(), id); err == nil {
		previousRole = before.Role
	}

	user, err := h.users.AdminUpdate(r.Context(), id, fields)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	// One entry per request, but under the action that describes what
	// actually happened: a role change is the one a superadmin reviewing the
	// trail will filter for, so it isn't buried under a generic "edited".
	detail := map[string]any{"email": user.Email, "passwordChanged": fields.PasswordHash != nil}
	action := models.AuditUserUpdatedByAdmin
	if previousRole != user.Role {
		action = models.AuditUserRoleChanged
		detail["from"] = string(previousRole)
		detail["to"] = string(user.Role)
	}
	h.record(r, action, models.AuditSubjectUser, id, detail)

	writeJSON(w, http.StatusOK, toUserMeResponse(user))
}

// DeleteUser removes an account. The superuser can't delete their own
// account this way, to avoid locking everyone out of user management.
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	callerID, _ := middleware.UserIDFromContext(r.Context())

	if id == callerID {
		writeError(w, http.StatusBadRequest, "You can't delete your own account", CodeCantDeleteOwnAccount)
		return
	}

	// Named before the delete: afterwards audit_log.user_id for this account's
	// own past entries is nulled (ON DELETE SET NULL), so the email in this
	// entry's detail is what keeps the trail readable.
	var email string
	if target, err := h.users.FindByID(r.Context(), id); err == nil {
		email = target.Email
	}

	if err := h.users.Delete(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	h.record(r, models.AuditUserDeleted, models.AuditSubjectUser, id,
		map[string]any{"email": email})
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}

// DisableMFA turns off every MFA factor (TOTP and any registered WebAuthn
// security keys) on an account, for when a user is locked out of their
// authenticator/security key (ai-instruct-68). Unlike UpdateUser this
// doesn't require the caller to resubmit the account's other fields.
func (h *AdminHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := h.users.DisableTOTP(r.Context(), id, false); err != nil {
		writeStoreError(w, err)
		return
	}
	if err := h.webauthnCreds.DeleteAllForUser(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	user, err := h.users.SetMFAEnabled(r.Context(), id, false)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	// Audited by name: one account weakening another account's security is
	// exactly what an audit trail is for.
	h.record(r, models.AuditUserMFACleared, models.AuditSubjectUser, id,
		map[string]any{"email": user.Email})
	writeJSON(w, http.StatusOK, toUserMeResponse(user))
}

// Audit serves the trail, newest first, paginated and filterable by action
// and subject. Superuser-only, enforced by the route's RequireSuperuser
// middleware and not by the screen: it names who decided what across every
// organization, which no organization's own owner is entitled to see.
func (h *AdminHandler) Audit(w http.ResponseWriter, r *http.Request) {
	page, err := h.audits.List(r.Context(), auditFilterFromQuery(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}
