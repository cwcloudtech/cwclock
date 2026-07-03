package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/store"
)

type ClientHandler struct {
	clients *store.ClientStore
}

func NewClientHandler(clients *store.ClientStore) *ClientHandler {
	return &ClientHandler{clients: clients}
}

type clientPayload struct {
	Name               string  `json:"name"`
	Address            string  `json:"address"`
	PostalCode         string  `json:"postalCode"`
	City               string  `json:"city"`
	Country            string  `json:"country"`
	VATNumber          string  `json:"vatNumber"`
	VATRate            float64 `json:"vatRate"`
	VATDischargeMotive string  `json:"vatDischargeMotive"`
	PurchaseOrder      string  `json:"purchaseOrder"`
	HoursPerDay        float64 `json:"hoursPerDay"`
}

func (p clientPayload) valid() bool {
	return p.Name != ""
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
	}
}

func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	clients, err := h.clients.List(r.Context(), orgID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, clients)
}

func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, _ := middleware.OrgIDFromContext(r.Context())

	var p clientPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil || !p.valid() {
		writeError(w, http.StatusBadRequest, "Please add a name field")
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
		writeError(w, http.StatusBadRequest, "Please add a name field")
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
