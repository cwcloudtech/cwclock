package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/aidownload"
)

// AIDownloadHandler serves the files the AI assistant generated (report / invoice
// PDFs). It is deliberately public — no Auth middleware — because a markdown link
// clicked from the chat carries no Authorization header. The guard is the
// capability in the path: `{userID}-ttl-{random}`, where the 128-bit random is
// unguessable and the whole thing expires on a TTL. This mirrors the calendar-feed
// token already in this API, where a secret in the path stands in for a session.
type AIDownloadHandler struct {
	store *aidownload.Store
}

func NewAIDownloadHandler(store *aidownload.Store) *AIDownloadHandler {
	return &AIDownloadHandler{store: store}
}

// Download streams one generated file by its capability id, as an attachment. A
// missing, expired or malformed id is a flat 404 — an attacker learns nothing
// about which ids exist.
func (h *AIDownloadHandler) Download(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		http.NotFound(w, r)
		return
	}

	path, filename, contentType, err := h.store.Open(chi.URLParam(r, "id"))
	if err != nil {
		if errors.Is(err, aidownload.ErrExpired) {
			writeError(w, http.StatusGone, "This download link has expired", CodeNotFound)
			return
		}
		http.NotFound(w, r)
		return
	}

	file, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, filename))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Private and uncacheable: the link is a capability, not a public asset.
	w.Header().Set("Cache-Control", "private, no-store")
	if info, statErr := file.Stat(); statErr == nil {
		http.ServeContent(w, r, filename, info.ModTime(), file)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, file)
}
