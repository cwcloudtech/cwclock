package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type UserHandler struct {
	users     *store.UserStore
	jwtSecret string
}

func NewUserHandler(users *store.UserStore, jwtSecret string) *UserHandler {
	return &UserHandler{users: users, jwtSecret: jwtSecret}
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
}

func (h *UserHandler) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(20 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var p registerPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Please add all fields")
		return
	}
	if utils.IsBlank(p.Email) || utils.IsBlank(p.Password) || utils.IsBlank(p.Name) || utils.IsBlank(p.Surname) {
		writeError(w, http.StatusBadRequest, "Please add all fields")
		return
	}

	if _, err := h.users.FindByEmail(r.Context(), p.Email); err == nil {
		writeError(w, http.StatusBadRequest, "A user with this email already exists")
		return
	} else if !errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := h.users.Create(r.Context(), p.Email, string(hash), p.Name, p.Surname)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user data")
		return
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.UserResponse{
		ID: user.ID, Email: user.Email, Name: user.Name, Surname: user.Surname,
		Role: user.Role, Token: token, Picture: user.Picture,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid credentials")
		return
	}

	user, err := h.users.FindByEmail(r.Context(), creds.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.UserResponse{
		ID: user.ID, Email: user.Email, Name: user.Name, Surname: user.Surname,
		Role: user.Role, Token: token, Picture: user.Picture,
	})
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	user, err := h.users.FindByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	writeJSON(w, http.StatusOK, models.UserMeResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Surname:   user.Surname,
		Role:      user.Role,
		Picture:   user.Picture,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

type updatePicturePayload struct {
	Picture string `json:"picture"`
}

// UpdatePicture lets the connected user set their own avatar picture
// (base64), shown in the profile dropdown.
func (h *UserHandler) UpdatePicture(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var p updatePicturePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.users.UpdatePicture(r.Context(), userID, p.Picture)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, models.UserMeResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Surname:   user.Surname,
		Role:      user.Role,
		Picture:   user.Picture,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

type updateProfilePayload struct {
	Name            string `json:"name"`
	Surname         string `json:"surname"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

// UpdateProfile lets the connected user set their own name, surname and,
// optionally, a new password (left untouched when the field is empty).
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var p updateProfilePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if utils.IsBlank(p.Name) || utils.IsBlank(p.Surname) {
		writeError(w, http.StatusBadRequest, "Please add name and surname fields")
		return
	}

	var passwordHash *string
	if utils.IsNotBlank(p.Password) {
		if p.Password != p.ConfirmPassword {
			writeError(w, http.StatusBadRequest, "Passwords do not match")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		hashed := string(hash)
		passwordHash = &hashed
	}

	user, err := h.users.UpdateProfile(r.Context(), userID, p.Name, p.Surname, passwordHash)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, models.UserMeResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Surname:   user.Surname,
		Role:      user.Role,
		Picture:   user.Picture,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

// Search powers email autocomplete when inviting members: it returns users
// whose email contains the "q" query param.
func (h *UserHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		writeJSON(w, http.StatusOK, []models.UserMeResponse{})
		return
	}

	users, err := h.users.SearchByEmail(r.Context(), query, 10)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	results := make([]models.UserMeResponse, len(users))
	for i, u := range users {
		results[i] = models.UserMeResponse{
			ID: u.ID, Email: u.Email, Name: u.Name, Surname: u.Surname,
			Role: u.Role, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
		}
	}
	writeJSON(w, http.StatusOK, results)
}
