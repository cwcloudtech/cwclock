package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type OrganizationHandler struct {
	orgs         *store.OrgStore
	users        *store.UserStore
	countries    *store.CountryStore
	currencies   *store.CurrencyStore
	maxImageSize int64
}

func NewOrganizationHandler(orgs *store.OrgStore, users *store.UserStore, countries *store.CountryStore, currencies *store.CurrencyStore, maxImageSize int64) *OrganizationHandler {
	return &OrganizationHandler{orgs: orgs, users: users, countries: countries, currencies: currencies, maxImageSize: maxImageSize}
}

type organizationPayload struct {
	Name                 string  `json:"name"`
	Address              string  `json:"address"`
	PostalCode           string  `json:"postalCode"`
	City                 string  `json:"city"`
	Country              string  `json:"country"`
	VATNumber            string  `json:"vatNumber"`
	SIREN                string  `json:"siren"`
	SIRET                string  `json:"siret"`
	NAF                  string  `json:"naf"`
	MF                   string  `json:"mf"`
	IdentificationNumber string  `json:"identificationNumber"`
	Picture              string  `json:"picture"`
	PictureX             float64 `json:"pictureX"`
	PictureY             float64 `json:"pictureY"`
	Stamp                string  `json:"stamp"`
	StampX               float64 `json:"stampX"`
	StampY               float64 `json:"stampY"`
	Currency             string  `json:"currency"`
}

// nameValid and Country's own blank check (see Create/Update) are kept
// separate rather than one combined "valid" bool, so a blank name and a
// blank country produce their own specific error message instead of both
// being reported as "Please fill in the Name field" (ai-instruct-37).
func (p organizationPayload) nameValid() bool {
	return utils.IsNotBlank(p.Name)
}

func (h *OrganizationHandler) validCountry(ctx context.Context, p organizationPayload) (bool, error) {
	return h.countries.Exists(ctx, p.Country)
}

func (h *OrganizationHandler) validCurrency(ctx context.Context, p organizationPayload) (bool, error) {
	if utils.IsBlank(p.Currency) {
		return true, nil
	}
	return h.currencies.Exists(ctx, p.Currency)
}

func (p organizationPayload) imageTooLarge(maxImageSize int64) bool {
	return utils.ImageSizeExceeds(p.Picture, maxImageSize) || utils.ImageSizeExceeds(p.Stamp, maxImageSize)
}

func (p organizationPayload) toFields() store.OrganizationFields {
	return store.OrganizationFields{
		Name:                 p.Name,
		Address:              p.Address,
		PostalCode:           p.PostalCode,
		City:                 p.City,
		Country:              p.Country,
		VATNumber:            p.VATNumber,
		SIREN:                p.SIREN,
		SIRET:                p.SIRET,
		NAF:                  p.NAF,
		MF:                   p.MF,
		IdentificationNumber: p.IdentificationNumber,
		Picture:              p.Picture,
		PictureX:             p.PictureX,
		PictureY:             p.PictureY,
		Stamp:                p.Stamp,
		StampX:               p.StampX,
		StampY:               p.StampY,
		Currency:             p.Currency,
	}
}

func (h *OrganizationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var p organizationPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || !p.nameValid() {
		writeError(w, http.StatusBadRequest, "Please fill in the Name field", CodeNameRequired)
		return
	}
	if utils.IsBlank(p.Country) {
		writeError(w, http.StatusBadRequest, "Please select a country", CodeCountryRequired)
		return
	}
	if ok, err := h.validCountry(r.Context(), p); err != nil {
		writeStoreError(w, err)
		return
	} else if !ok {
		writeError(w, http.StatusBadRequest, "Please use a supported country code", CodeInvalidCountry)
		return
	}
	if ok, err := h.validCurrency(r.Context(), p); err != nil {
		writeStoreError(w, err)
		return
	} else if !ok {
		writeError(w, http.StatusBadRequest, "Please use a supported currency code", CodeInvalidCurrency)
		return
	}
	if p.imageTooLarge(h.maxImageSize) {
		writeError(w, http.StatusBadRequest, "Image is too large", CodeImageTooLarge)
		return
	}

	org, err := h.orgs.Create(r.Context(), userID, p.toFields())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, org)
}

func (h *OrganizationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	orgs, err := h.orgs.ListForUser(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, orgs)
}

// AdminList returns every organization, regardless of membership, for the
// superuser's organization-management screen.
func (h *OrganizationHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.orgs.ListAllWithOwner(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, orgs)
}

func (h *OrganizationHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	org, err := h.orgs.FindByID(r.Context(), orgID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, org)
}

func (h *OrganizationHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	var p organizationPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || !p.nameValid() {
		writeError(w, http.StatusBadRequest, "Please fill in the Name field", CodeNameRequired)
		return
	}
	if utils.IsBlank(p.Country) {
		writeError(w, http.StatusBadRequest, "Please select a country", CodeCountryRequired)
		return
	}
	if ok, err := h.validCountry(r.Context(), p); err != nil {
		writeStoreError(w, err)
		return
	} else if !ok {
		writeError(w, http.StatusBadRequest, "Please use a supported country code", CodeInvalidCountry)
		return
	}
	if ok, err := h.validCurrency(r.Context(), p); err != nil {
		writeStoreError(w, err)
		return
	} else if !ok {
		writeError(w, http.StatusBadRequest, "Please use a supported currency code", CodeInvalidCurrency)
		return
	}
	if p.imageTooLarge(h.maxImageSize) {
		writeError(w, http.StatusBadRequest, "Image is too large", CodeImageTooLarge)
		return
	}

	org, err := h.orgs.Update(r.Context(), orgID, p.toFields())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, org)
}

func (h *OrganizationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	if err := h.orgs.Delete(r.Context(), orgID); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": orgID})
}

type memberPayload struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func validRole(role string) bool {
	switch models.Role(role) {
	case models.RoleAdmin, models.RoleMember, models.RoleReader:
		return true
	default:
		return false
	}
}

// redactRates hides the daily rate from the response unless the caller is
// an admin or the owner, per "the daily rate must appear only for the
// admins and owner of the organization".
func redactRates(r *http.Request, members []models.Member) []models.Member {
	role, _ := middleware.OrgRoleFromContext(r.Context())
	if role == models.RoleAdmin || role == models.RoleOwner {
		return members
	}
	redacted := make([]models.Member, len(members))
	for i, m := range members {
		m.DailyRate = nil
		redacted[i] = m
	}
	return redacted
}

func (h *OrganizationHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	members, err := h.orgs.ListMembers(r.Context(), orgID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, redactRates(r, members))
}

func (h *OrganizationHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	var p memberPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.Email) || !validRole(p.Role) {
		writeError(w, http.StatusBadRequest, "Please add a valid email and role (admin, member or reader)", CodeInvalidMemberInvite)
		return
	}

	user, err := h.users.FindByEmail(r.Context(), p.Email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "No user with this email", CodeNoUserWithEmail)
			return
		}
		writeStoreError(w, err)
		return
	}

	member, err := h.orgs.AddMember(r.Context(), orgID, user.ID, models.Role(p.Role))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, member)
}

func (h *OrganizationHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID := chi.URLParam(r, "userId")

	var p memberPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || !validRole(p.Role) {
		writeError(w, http.StatusBadRequest, "Please add a valid role (admin, member or reader)", CodeInvalidRole)
		return
	}

	member, err := h.orgs.UpdateMemberRole(r.Context(), orgID, userID, models.Role(p.Role))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, member)
}

type memberRatePayload struct {
	DailyRate float64 `json:"dailyRate"`
}

func (h *OrganizationHandler) SetMemberRate(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID := chi.URLParam(r, "userId")

	var p memberRatePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || p.DailyRate <= 0 {
		writeError(w, http.StatusBadRequest, "Please add a valid dailyRate", CodeInvalidDailyRate)
		return
	}

	member, err := h.orgs.SetMemberRate(r.Context(), orgID, userID, p.DailyRate)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, member)
}

func (h *OrganizationHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID := chi.URLParam(r, "userId")

	if err := h.orgs.RemoveMember(r.Context(), orgID, userID); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"userId": userID})
}

type transferOwnershipPayload struct {
	Email string `json:"email"`
}

func (h *OrganizationHandler) TransferOwnership(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	previousOwnerID, _ := middleware.UserIDFromContext(r.Context())

	var p transferOwnershipPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.Email) {
		writeError(w, http.StatusBadRequest, "Please add a valid email", CodeInvalidEmail)
		return
	}

	newOwner, err := h.users.FindByEmail(r.Context(), p.Email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "No user with this email", CodeNoUserWithEmail)
			return
		}
		writeStoreError(w, err)
		return
	}

	org, err := h.orgs.TransferOwnership(r.Context(), orgID, previousOwnerID, newOwner.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, org)
}
