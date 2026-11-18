package models

import "time"

// A session is one address a connection came from, enriched with what a geo
// lookup made of it. These types back the IP administration screen - the roster
// of addresses, the view of one member's addresses - and the ban list the guard
// middleware reads. Ported from the sf-api / uprodit "IP ban + geo" feature. Geo
// is a free-form map rather than a struct: the providers return overlapping-but-
// not-identical fields, and a new one tomorrow is a key, not a migration.

// I18nIPBanned is the i18n_code the IP ban guard returns with its 403, so the
// frontend can show a "this address is banned" message distinct from the
// account-ban one (I18nAccountBanned).
const I18nIPBanned = "errors.ipBanned"

// GeoInfo is what is known about an address's location and network. Nil when no
// lookup has succeeded for it yet.
type GeoInfo map[string]string

// ConnectionIP is one address a member has connected from, with how often, and
// whether it is currently banned. It is the per-member view's row.
type ConnectionIP struct {
	IP        string    `json:"ip"`
	Count     int       `json:"count"`
	FirstSeen time.Time `json:"firstSeen"`
	LastSeen  time.Time `json:"lastSeen"`
	Banned    bool      `json:"banned"`
	Geo       GeoInfo   `json:"geo,omitempty"`
}

// IpMember is a member seen behind an address, as the roster shows them.
type IpMember struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name,omitempty"`
	Surname string `json:"surname,omitempty"`
	Handle  string `json:"handle,omitempty"`
}

// IpSummary is one address on the administration roster: who has been seen
// behind it, where it is, and whether it is banned. Members is empty for an
// address only anonymous visitors have used, or one an administrator banned
// before it was ever seen.
type IpSummary struct {
	IP        string     `json:"ip"`
	Banned    bool       `json:"banned"`
	Reason    string     `json:"reason,omitempty"`
	Count     int        `json:"count"`
	FirstSeen *time.Time `json:"firstSeen,omitempty"`
	LastSeen  *time.Time `json:"lastSeen,omitempty"`
	Geo       GeoInfo    `json:"geo,omitempty"`
	Members   []IpMember `json:"members"`
}

// IpStats backs the badges above the roster: how many distinct addresses have
// been seen, and how many are banned.
type IpStats struct {
	Total  int `json:"total"`
	Banned int `json:"banned"`
}

// IpBanAction is the body of the ban/unban endpoint. One endpoint, two actions,
// because banning and lifting a ban are the same decision from opposite sides.
type IpBanAction struct {
	IP     string `json:"ip"`
	Action string `json:"action"`
	Motif  string `json:"motif"`
}

const (
	// IPActionBan and IPActionUnban are the two values IpBanAction.Action takes.
	IPActionBan   = "ban"
	IPActionUnban = "unban"
)
