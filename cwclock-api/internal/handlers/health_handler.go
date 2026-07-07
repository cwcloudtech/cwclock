package handlers

import "net/http"

// Health is a liveness probe: no dependency checks, just confirms the
// process is up and serving requests.
func Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "alive": true})
}
