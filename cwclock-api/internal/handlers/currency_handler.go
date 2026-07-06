package handlers

import (
	"net/http"

	"cwclock-api/internal/models"
)

// ListCurrencies returns the effective, ordered list of ISO 4217 currency
// codes organizations may be billed in (the built-in default, or the
// CWCLOCK_ALLOWED_CURRENCIES override if the deployment set one). It's
// public and has no store dependency: the frontend uses it instead of
// hardcoding the list, so an override takes effect without a UI deploy.
func ListCurrencies(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, models.AllowedCurrencies)
}
