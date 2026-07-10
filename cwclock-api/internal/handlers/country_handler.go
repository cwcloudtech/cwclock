package handlers

import (
	"net/http"

	"cwclock-api/internal/store"
)

type CountryHandler struct {
	countries *store.CountryStore
}

func NewCountryHandler(countries *store.CountryStore) *CountryHandler {
	return &CountryHandler{countries: countries}
}

// List returns every ISO 3166-1 alpha-2 country organizations/clients may
// pick as their country (see ai-instruct-35). Public and read-only.
func (h *CountryHandler) List(w http.ResponseWriter, r *http.Request) {
	countries, err := h.countries.List(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"countries": countries})
}
