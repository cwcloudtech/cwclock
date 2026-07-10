package handlers

import (
	"net/http"

	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
	"cwclock-api/internal/utils"
)

type FieldHandler struct {
	fields *store.FieldStore
}

func NewFieldHandler(fields *store.FieldStore) *FieldHandler {
	return &FieldHandler{fields: fields}
}

// List returns the business identification fields to display for a given
// country (ai-instruct-35's decision table), e.g. ?country=FR ->
// ["SIRET","SIREN","NAF","VAT Code"]. Public and read-only.
func (h *FieldHandler) List(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	if utils.IsBlank(country) {
		writeError(w, http.StatusBadRequest, "Please provide a country query parameter", CodeInvalidRequestBody)
		return
	}

	rows, err := h.fields.ListForCountry(r.Context(), country)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"fields": models.ResolveFields(rows)})
}
