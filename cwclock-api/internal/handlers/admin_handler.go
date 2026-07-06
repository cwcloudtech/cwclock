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
	users *store.UserStore
}

func NewAdminHandler(users *store.UserStore) *AdminHandler {
	return &AdminHandler{users: users}
}

func toUserMeResponse(u models.User) models.UserMeResponse {
	return models.UserMeResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Surname:   u.Surname,
		Role:      u.Role,
		Picture:   u.Picture,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ListUsers returns every account, for the superuser's user-management screen.
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
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
	case models.GlobalRoleSuperuser, models.GlobalRoleConfirmed, models.GlobalRoleDisabled:
		return true
	default:
		return false
	}
}

type adminUpdateUserPayload struct {
	Email    string  `json:"email"`
	Name     string  `json:"name"`
	Surname  string  `json:"surname"`
	Role     string  `json:"role"`
	Password *string `json:"password"`
	Picture  *string `json:"picture"`
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

	fields := store.AdminUserFields{Email: p.Email, Name: p.Name, Surname: p.Surname, Role: p.Role}

	if p.Password != nil && utils.IsNotBlank(*p.Password) {
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
	}

	user, err := h.users.AdminUpdate(r.Context(), id, fields)
	if err != nil {
		writeStoreError(w, err)
		return
	}
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

	if err := h.users.Delete(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}
