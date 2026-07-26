package utils

import (
	"net/http"
	"slices"
)

// CorsMiddleware applies CORS headers matching the config-driven allowed origins,
// methods and headers described in the alertmanager CORS spec.
func CorsMiddleware(next http.Handler, allowedOrigins []string) http.Handler {
	const allowedMethods = "GET, POST, PUT, DELETE, OPTIONS"
	const allowedHeaders = "Authorization, Content-Type"
	// Content-Disposition must be explicitly exposed, otherwise the
	// browser hides it from JS on cross-origin responses (like report
	// exports), so the frontend can't read the backend-provided
	// filename and falls back to a generic one.
	const exposedHeaders = "Content-Disposition"

	allowAllOrigins := slices.Contains(allowedOrigins, "*")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if IsNotBlank(origin) {
			switch {
			case allowAllOrigins:
				w.Header().Set("Access-Control-Allow-Origin", "*")
			case slices.Contains(allowedOrigins, origin):
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Expose-Headers", exposedHeaders)
		}

		if r.Method == http.MethodOptions && IsNotBlank(r.Header.Get("Access-Control-Request-Method")) {
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
