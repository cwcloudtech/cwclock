package handlers

import (
	"net/http"
	"os"
)

// NewManifestHandler serves the raw content of the build-time manifest.json
// (version/sha/details baked in by the CI/CD pipeline), read per request so
// a swapped file takes effect without restarting the process.
func NewManifestHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(path)
		if err != nil {
			writeError(w, http.StatusNotFound, "manifest not found", CodeNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}
