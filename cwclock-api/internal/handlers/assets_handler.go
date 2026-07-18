package handlers

import (
	"net/http"

	"cwclock-api/internal/assets"
)

// AssetsLogo serves the bundled CWClock logo as a real image over HTTP, so
// transactional emails can reference it by URL instead of embedding it as a
// data: URI - mail clients and email-sending APIs commonly strip data:
// URIs from <img src> outright, a limitation no amount of correctly
// escaping the source HTML can work around. No auth: it needs to be
// fetchable by a mail client with no credentials, and it's just a logo.
func AssetsLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(assets.CWClockLogoPNG)
}
