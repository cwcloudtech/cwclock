package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type ClientHandler struct {
	clients *store.ClientStore
}

func NewClientHandler(clients *store.ClientStore) *ClientHandler {
	return &ClientHandler{clients: clients}
}

type clientPayload struct {
	Name               string   `json:"name"`
	Address            string   `json:"address"`
	PostalCode         string   `json:"postalCode"`
	City               string   `json:"city"`
	Country            string   `json:"country"`
	VATNumber          string   `json:"vatNumber"`
	VATRate            float64  `json:"vatRate"`
	VATDischargeMotive string   `json:"vatDischargeMotive"`
	PurchaseOrder      string   `json:"purchaseOrder"`
	HoursPerDay        float64  `json:"hoursPerDay"`
	DailyRate          *float64 `json:"dailyRate"`
}

func (p clientPayload) valid() bool {
	return utils.IsNotBlank(p.Name)
}

func (p clientPayload) toFields() store.ClientFields {
	return store.ClientFields{
		Name:               p.Name,
		Address:            p.Address,
		PostalCode:         p.PostalCode,
		City:               p.City,
		Country:            p.Country,
		VATNumber:          p.VATNumber,
		VATRate:            p.VATRate,
		VATDischargeMotive: p.VATDischargeMotive,
		PurchaseOrder:      p.PurchaseOrder,
		HoursPerDay:        p.HoursPerDay,
		DailyRate:          p.DailyRate,
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
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || !p.valid() {
		writeError(w, http.StatusBadRequest, "Please add a name field", CodeNameRequired)
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
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || !p.valid() {
		writeError(w, http.StatusBadRequest, "Please add a name field", CodeNameRequired)
		return
	}

	client, err := h.clients.Update(r.Context(), id, p.toFields())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, client)
}

func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "clientId")

	if err := h.clients.Delete(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id})
}
