package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/notify"
	"cwclock-api/internal/store"
)

// ipBanPageSize bounds one page of the IP roster; the screen filters and pages
// server-side, so this is the slice sent per request, not a cap on how many
// addresses may exist.
const ipBanPageSize = 50

// IPBanHandler is the superuser's IP administration: the roster of addresses seen
// (with their members and geolocation), the ban counts, banning/unbanning an
// address, and the per-user list of addresses an account has connected from.
// Ported from the sf-api admin IP endpoints.
type IPBanHandler struct {
	sessions *store.SessionStore
	disc     *notify.Discord // [BAN]/[UNBAN] admin-action alerts (optional)
}

func NewIPBanHandler(sessions *store.SessionStore) *IPBanHandler {
	return &IPBanHandler{sessions: sessions}
}

// WithNotifier wires the Discord transport so an admin ban/unban posts a
// [BAN]/[UNBAN] alert. Optional and best-effort (a nil/disabled transport is a
// no-op and never affects the moderation action).
func (h *IPBanHandler) WithNotifier(disc *notify.Discord) *IPBanHandler {
	h.disc = disc
	return h
}

// ListIPs serves GET /v1/admin/ip-bans. Returns a page of the address roster,
// filtered by ?term= (ip / email / name) and paged by ?page= (0-based).
func (h *IPBanHandler) ListIPs(w http.ResponseWriter, r *http.Request) {
	term := strings.TrimSpace(r.URL.Query().Get("term"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 0 {
		page = 0
	}
	// Fetch one extra to tell the client whether a next page exists.
	rows, err := h.sessions.ListIps(r.Context(), term, page*ipBanPageSize, ipBanPageSize+1)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	hasNext := len(rows) > ipBanPageSize
	if hasNext {
		rows = rows[:ipBanPageSize]
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"results":  rows,
		"page":     page,
		"pageSize": ipBanPageSize,
		"hasNext":  hasNext,
	})
}

// Stats serves GET /v1/admin/ip-bans/stats: the distinct-address and banned counts.
func (h *IPBanHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.sessions.Stats(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// BanIP serves PUT /v1/admin/ip-bans. The body's action ("ban"/"unban") selects
// the operation; motif carries the ban reason.
func (h *IPBanHandler) BanIP(w http.ResponseWriter, r *http.Request) {
	callerID, _ := middleware.UserIDFromContext(r.Context())

	var p models.IpBanAction
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", CodeInvalidRequestBody)
		return
	}
	ip := strings.TrimSpace(p.IP)
	if net.ParseIP(ip) == nil {
		writeError(w, http.StatusBadRequest, "Please provide a valid IP address", CodeInvalidRequestBody)
		return
	}

	switch p.Action {
	case models.IPActionBan:
		if err := h.sessions.BanIp(r.Context(), ip, p.Motif, callerID); err != nil {
			writeStoreError(w, err)
			return
		}
		notifyDiscord(h.disc, fmt.Sprintf("[BAN] :no_entry: ip = %s, motif = %s, by = %s", ip, p.Motif, callerID))
	case models.IPActionUnban:
		if err := h.sessions.UnbanIp(r.Context(), ip); err != nil {
			writeStoreError(w, err)
			return
		}
		notifyDiscord(h.disc, fmt.Sprintf("[UNBAN] :white_check_mark: ip = %s, by = %s", ip, callerID))
	default:
		writeError(w, http.StatusBadRequest, "Unknown action (expected 'ban' or 'unban')", CodeInvalidRequestBody)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ip": ip, "action": p.Action})
}

// UserIPs serves GET /v1/admin/users/{id}/ips: every address the account has
// connected from, newest-first.
func (h *IPBanHandler) UserIPs(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	ips, err := h.sessions.ListByUser(r.Context(), userID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ips)
}

// notifyDiscord posts a "[TAG] :emoji: …" notice to Discord off the request path:
// best-effort, async and nil-safe (a nil/disabled transport is a no-op).
func notifyDiscord(disc *notify.Discord, text string) {
	go func() { _ = disc.SendText(context.Background(), text) }()
}
