package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"cwclock-api/internal/store"
)

type errorBody struct {
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorBody{Message: message})
}

// writeStoreError maps a store error to its HTTP status: 404 when the
// resource doesn't exist, 500 for anything else.
func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Resource not found")
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}
