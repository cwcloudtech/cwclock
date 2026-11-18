package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

// visitTimeout bounds the background work a visit kicks off - an upsert and, on a
// first-seen address, a geo lookup. It runs detached from the request.
const visitTimeout = 10 * time.Second

// VisitorStore is what the IP guard needs from the session store: the per-request
// ban check, the once-a-day visit upsert, and the geo backfill.
type VisitorStore interface {
	IsIpBanned(ctx context.Context, ip string) (bool, error)
	RecordVisit(ctx context.Context, userID *string, ip string, at time.Time) (bool, error)
	NeedsGeo(ctx context.Context, ip string) (bool, error)
	SetGeo(ctx context.Context, ip string, info map[string]string) error
}

// Enricher resolves an address's geolocation via an external provider. Nil records
// visits without geo. Structurally matches enrich.Geo.Enrich.
type Enricher func(ctx context.Context, ip string) (map[string]string, error)

// BanIP sits in front of every route: a banned address is refused on the login
// page and the public feed alike (403), and the visit is logged here - anonymously,
// at most once per calendar day. The IP↔user association is recorded separately at
// login/register/OIDC. The ban check fails open: a lookup error allows the request
// rather than taking the whole site down over a moderation feature. An address that
// will not parse is left alone (behind a trusted proxy the header is a real address).
func BanIP(store VisitorStore, enrich Enricher, alerter *BanAlerter) func(http.Handler) http.Handler {
	seen := newDayCache()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isInfraPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			ip := clientIP(r)
			if utils.IsBlank(ip) || store == nil || net.ParseIP(ip) == nil {
				next.ServeHTTP(w, r)
				return
			}

			banned, err := store.IsIpBanned(r.Context(), ip)
			if err != nil {
				slog.WarnContext(r.Context(), "ip ban check failed, allowing", "ip", ip, "error", err)
			} else if banned {
				// Alert (deduped, async) that a banned IP is still knocking, then refuse.
				alerter.Alert("ip:"+ip, fmt.Sprintf("banned ip %s tried %s %s", ip, r.Method, r.URL.Path))
				jsonErrorCode(w, http.StatusForbidden, "This address has been banned.", models.I18nIPBanned)
				return
			}

			// Log the visit anonymously, at most once a day per address (the DB is the
			// real once-a-day authority; this cache keeps the hot path write-free).
			if seen.firstToday(ip) {
				go TrackVisit(store, enrich, nil, ip)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// TrackVisit upserts a visit and, when the address is newly recorded, enriches it
// with a geo lookup. Safe in its own goroutine: fresh background context, recovers
// from a panic in the enricher, returns nothing - a lost counter or missing city
// must never disturb the request that spawned it.
func TrackVisit(store VisitorStore, enrich Enricher, userID *string, ip string) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic recording a visit", "ip", ip, "panic", rec)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), visitTimeout)
	defer cancel()

	recorded, err := store.RecordVisit(ctx, userID, ip, time.Now().UTC())
	if err != nil {
		slog.WarnContext(ctx, "failed to record a visit", "ip", ip, "error", err)
		return
	}
	if !recorded || enrich == nil {
		return
	}

	needs, err := store.NeedsGeo(ctx, ip)
	if err != nil || !needs {
		return
	}
	info, err := enrich(ctx, ip)
	if err != nil || len(info) == 0 {
		return
	}
	if err := store.SetGeo(ctx, ip, info); err != nil {
		slog.WarnContext(ctx, "failed to store geo enrichment", "ip", ip, "error", err)
	}
}

// clientIP reads the caller's address from the reverse-proxy headers (X-Real-IP,
// then X-Forwarded-By), the same order internal/contact.ResolveClientIP uses - the
// proxy sets them, so there is no RemoteAddr port to strip.
func clientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	return strings.TrimSpace(r.Header.Get("X-Forwarded-By"))
}

// dayCache remembers which addresses have already been recorded today in this
// process, so the guard records each only once a day without a database round trip
// on every request. The database is the real authority (across replicas/restarts);
// this is only the fast path in front of it.
type dayCache struct {
	mu   sync.Mutex
	day  string
	seen map[string]struct{}
}

func newDayCache() *dayCache {
	return &dayCache{day: today(), seen: make(map[string]struct{})}
}

// firstToday reports whether key has not been seen yet today, recording it when so.
func (c *dayCache) firstToday(key string) bool {
	now := today()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.day != now {
		c.day = now
		c.seen = make(map[string]struct{})
	}
	if _, ok := c.seen[key]; ok {
		return false
	}
	c.seen[key] = struct{}{}
	return true
}

func today() string { return time.Now().UTC().Format("2006-01-02") }

// isInfraPath reports whether a path is an operational endpoint the guard passes
// straight through - the health check and the metrics scrape.
func isInfraPath(path string) bool {
	return strings.HasSuffix(path, "/health") || strings.HasSuffix(path, "/metrics")
}
