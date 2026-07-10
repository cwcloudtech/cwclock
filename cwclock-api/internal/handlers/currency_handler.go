package handlers

import (
	"net/http"

	"cwclock-api/internal/store"
)

type CurrencyHandler struct {
	currencies *store.CurrencyStore
}

func NewCurrencyHandler(currencies *store.CurrencyStore) *CurrencyHandler {
	return &CurrencyHandler{currencies: currencies}
}

// List returns every ISO 4217 currency organizations may be billed in, from
// the currencies table (see ai-instruct-35 - this replaced the
// CWCLOCK_ALLOWED_CURRENCIES env var, so a deployment now edits the table
// instead of redeploying to change the list). Public and read-only: the
// frontend uses it instead of hardcoding the list.
func (h *CurrencyHandler) List(w http.ResponseWriter, r *http.Request) {
	currencies, err := h.currencies.List(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"currencies": currencies})
}
