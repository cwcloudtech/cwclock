package middleware

import (
	"context"
	"sync"
	"time"
)

// banAlertTTL dedups repeated ban-attempt alerts for the same subject: a banned IP
// or user hammering endpoints raises at most one Discord alert per this window, so an
// attacker cannot flood the channel (or trip the webhook's own rate limit).
const banAlertTTL = time.Hour

// BanAlerter posts a "[BAN] :rotating_light: …" Discord notice the first time a
// banned IP or banned user is seen trying to reach the API, deduped per subject and
// detached from the request. It is optional (a nil *BanAlerter is a no-op) and never
// blocks or fails a request - the post runs in its own goroutine.
type BanAlerter struct {
	post func(ctx context.Context, text string)
	mu   sync.Mutex
	seen map[string]time.Time
}

// NewBanAlerter builds an alerter from a transport post function (typically the
// Discord SendText). A nil post yields a nil alerter, so callers can wire it
// unconditionally and have it disabled when no webhook is configured.
func NewBanAlerter(post func(ctx context.Context, text string)) *BanAlerter {
	if post == nil {
		return nil
	}
	return &BanAlerter{post: post, seen: map[string]time.Time{}}
}

// Alert raises one "[BAN] :rotating_light: <message>" notice for key, unless the same
// key already fired within banAlertTTL. Safe on a nil receiver.
func (b *BanAlerter) Alert(key, message string) {
	if b == nil {
		return
	}
	now := time.Now()

	b.mu.Lock()
	if last, ok := b.seen[key]; ok && now.Sub(last) < banAlertTTL {
		b.mu.Unlock()
		return
	}
	b.seen[key] = now
	// Opportunistic sweep so the dedup map can't grow without bound under a flood
	// of distinct banned subjects.
	if len(b.seen) > 1024 {
		for k, t := range b.seen {
			if now.Sub(t) >= banAlertTTL {
				delete(b.seen, k)
			}
		}
	}
	b.mu.Unlock()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.post(ctx, "[BAN] :rotating_light: "+message)
	}()
}
