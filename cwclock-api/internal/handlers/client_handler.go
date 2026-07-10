package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type ClientHandler struct {
	clients   *store.ClientStore
	orgs      *store.OrgStore
	countries *store.CountryStore
}

func NewClientHandler(clients *store.ClientStore, orgs *store.OrgStore, countries *store.CountryStore) *ClientHandler {
	return &ClientHandler{clients: clients, orgs: orgs, countries: countries}
}

type clientPayload struct {
	Name                 string   `json:"name"`
	Email                string   `json:"email"`
	ContactName          string   `json:"contactName"`
	Address              string   `json:"address"`
	PostalCode           string   `json:"postalCode"`
	City                 string   `json:"city"`
	Country              string   `json:"country"`
	VATNumber            string   `json:"vatNumber"`
	VATRate              *float64 `json:"vatRate"`
	VATDischargeMotive   string   `json:"vatDischargeMotive"`
	SIREN                string   `json:"siren"`
	SIRET                string   `json:"siret"`
	NAF                  string   `json:"naf"`
	MF                   string   `json:"mf"`
	IdentificationNumber string   `json:"identificationNumber"`
	PurchaseOrder        string   `json:"purchaseOrder"`
	HoursPerDay          float64  `json:"hoursPerDay"`
	DailyRate            *float64 `json:"dailyRate"`
}

// nameValid and Country's own blank check (see Create/Update) are kept
// separate rather than one combined "valid" bool, so a blank name and a
// blank country produce their own specific error message instead of both
// being reported as "Please fill in the Name field" (ai-instruct-37).
func (p clientPayload) nameValid() bool {
	return utils.IsNotBlank(p.Name)
}

// validEmail lets the field stay blank (it's optional) but requires a
// plausible email shape when it's set.
func (p clientPayload) validEmail() bool {
	return utils.IsBlank(p.Email) || utils.IsValidEmail(p.Email)
}

func (h *ClientHandler) validCountry(ctx context.Context, p clientPayload) (bool, error) {
	return h.countries.Exists(ctx, p.Country)
}

func (p clientPayload) toFields() store.ClientFields {
	return store.ClientFields{
		Name:                 p.Name,
		Email:                p.Email,
		ContactName:          p.ContactName,
		Address:              p.Address,
		PostalCode:           p.PostalCode,
		City:                 p.City,
		Country:              p.Country,
		VATNumber:            p.VATNumber,
		VATRate:              p.VATRate,
		VATDischargeMotive:   p.VATDischargeMotive,
		SIREN:                p.SIREN,
		SIRET:                p.SIRET,
		NAF:                  p.NAF,
		MF:                   p.MF,
		IdentificationNumber: p.IdentificationNumber,
		PurchaseOrder:        p.PurchaseOrder,
		HoursPerDay:          p.HoursPerDay,
		DailyRate:            p.DailyRate,
	}
}

// redactClientRates hides each client's daily rate from the response unless
// the caller is an admin or the owner, matching redactRates for member
// rates: readers/members can list clients to log time against, but the
// billing rate stays admin/owner-only.
func redactClientRates(r *http.Request, clients []models.Client) []models.Client {
	role, _ := middleware.OrgRoleFromContext(r.Context())
	if role == models.RoleAdmin || role == models.RoleOwner {
		return clients
	}
	redacted := make([]models.Client, len(clients))
	for i, c := range clients {
		c.DailyRate = nil
		redacted[i] = c
	}
	return redacted
}

func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	clients, err := h.clients.List(r.Context(), orgID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, redactClientRates(r, clients))
}

func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	var p clientPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || !p.nameValid() {
		writeError(w, http.StatusBadRequest, "Please fill in the Name field", CodeNameRequired)
		return
	}
	if !p.validEmail() {
		writeError(w, http.StatusBadRequest, "Please add a valid email", CodeInvalidEmail)
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

	client, err := h.clients.Create(r.Context(), orgID, p.toFields())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, client)
}

func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "clientId")

	var p clientPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || !p.nameValid() {
		writeError(w, http.StatusBadRequest, "Please fill in the Name field", CodeNameRequired)
		return
	}
	if !p.validEmail() {
		writeError(w, http.StatusBadRequest, "Please add a valid email", CodeInvalidEmail)
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

	client, err := h.clients.Update(r.Context(), id, p.toFields())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, client)
}

type transferClientPayload struct {
	TargetOrgID string `json:"targetOrgId"`
}

// Transfer moves a client (and its projects/time entries) to a different
// organization the acting user owns. Restricted to owners (see the
// RequireRole(RoleOwner) route gate) since it's a cross-organization data
// move - ai-instruct-34's "an owner should be able to transfer... to
// another organization he's owner as well".
func (h *ClientHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())
	userID, _ := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "clientId")

	var p transferClientPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || utils.IsBlank(p.TargetOrgID) {
		writeError(w, http.StatusBadRequest, "Please add a targetOrgId", CodeInvalidRequestBody)
		return
	}
	if p.TargetOrgID == orgID {
		writeError(w, http.StatusBadRequest, "Please pick a different organization", CodeInvalidRequestBody)
		return
	}

	client, err := h.clients.FindByID(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if client.OrganizationID != orgID {
		writeError(w, http.StatusNotFound, "Resource not found", CodeNotFound)
		return
	}

	targetOrg, err := h.orgs.FindByID(r.Context(), p.TargetOrgID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if targetOrg.OwnerID != userID {
		writeError(w, http.StatusForbidden, "You must own the target organization", CodeMustOwnTargetOrg)
		return
	}

	updated, err := h.clients.Transfer(r.Context(), id, p.TargetOrgID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "clientId")

	if err := h.clients.Delete(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}
