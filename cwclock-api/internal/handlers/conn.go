package handlers

import (
	"fmt"
	"net/http"

	"cwclock-api/internal/contact"
	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/notify"
	"cwclock-api/internal/store"
)

// connTracker is embedded by the auth handlers (password login, OIDC, MFA) to
// associate the caller's IP with the signed-in/registered user in the sessions
// table, and to post the Discord sign-in/out notices. Both are best-effort and
// off the request path; a nil store / disabled webhook makes them no-ops.
type connTracker struct {
	visits *store.SessionStore
	enrich middleware.Enricher
	disc   *notify.Discord
}

// setConn wires the session store, geo enricher and Discord transport.
func (c *connTracker) setConn(visits *store.SessionStore, enrich middleware.Enricher, disc *notify.Discord) {
	c.visits, c.enrich, c.disc = visits, enrich, disc
}

// associate links the caller's IP to the user (a sessions row, nb-connexion-style
// count bumped once a day). Best-effort and async.
func (c *connTracker) associate(r *http.Request, userID string) {
	if c.visits == nil || userID == "" {
		return
	}
	ip := contact.ResolveClientIP(r)
	if ip == "" {
		return
	}
	id := userID
	go middleware.TrackVisit(c.visits, c.enrich, &id, ip)
}

func (c *connTracker) notifyLogin(r *http.Request, u models.User) {
	c.authNotice(r, "[CONNECT] :heart_eyes:", u)
}
func (c *connTracker) notifyRegister(r *http.Request, u models.User) {
	c.authNotice(r, "[REGISTER] :heart_eyes:", u)
}
func (c *connTracker) notifyLogout(r *http.Request, u models.User) {
	c.authNotice(r, "[DISCONNECT] :wave:", u)
}

// authNotice posts one "[TAG] :emoji: email = …, id = …, ip = …" auth notice.
func (c *connTracker) authNotice(r *http.Request, tag string, u models.User) {
	notifyDiscord(c.disc, fmt.Sprintf("%s email = %s, id = %s, ip = %s",
		tag, u.Email, u.ID, contact.ResolveClientIP(r)))
}
