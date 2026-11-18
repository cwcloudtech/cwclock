package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
)

// SessionStore is the connection log and the IP ban list. Every connection - a
// member's or an anonymous visitor's - is a row in `sessions`; a banned address
// is a row in `ip_bans`. It backs the guard middleware's per-request ban check,
// the admin IP screen, and the per-member view of where an account has connected
// from. Ported from the sf-api SessionStore.
type SessionStore struct {
	pool *pgxpool.Pool
}

func NewSessionStore(pool *pgxpool.Pool) *SessionStore {
	return &SessionStore{pool: pool}
}

// RecordVisit notes a connection from ip, by the given member or - when userID is
// nil - anonymously, AT MOST ONCE PER CALENDAR DAY (UTC). It returns whether this
// call actually recorded (a first-ever sighting, or the first from that (ip,
// member) on a new day), which is the caller's cue to (re)check geo. "Once per
// day" is enforced in the DB (the DO UPDATE fires only when the row was last
// touched before today), so two replicas racing still bump the counter once; a
// same-day repeat comes back as pgx.ErrNoRows, mapped to (false, nil).
func (s *SessionStore) RecordVisit(ctx context.Context, userID *string, ip string, at time.Time) (bool, error) {
	day := at.UTC().Truncate(24 * time.Hour)
	var one int
	var err error
	if userID != nil && *userID != "" {
		err = s.pool.QueryRow(ctx, `
			INSERT INTO sessions (user_id, ip, data)
			VALUES ($1::uuid, $2, jsonb_build_object('count', 1))
			ON CONFLICT (ip, user_id) WHERE user_id IS NOT NULL
			DO UPDATE SET
				data = jsonb_set(sessions.data, '{count}',
					to_jsonb(COALESCE((sessions.data->>'count')::int, 0) + 1)),
				updated_at = $3::timestamptz
			WHERE sessions.updated_at < $4::timestamptz
			RETURNING 1
		`, *userID, ip, at, day).Scan(&one)
	} else {
		err = s.pool.QueryRow(ctx, `
			INSERT INTO sessions (user_id, ip, data)
			VALUES (NULL, $1, jsonb_build_object('count', 1))
			ON CONFLICT (ip) WHERE user_id IS NULL
			DO UPDATE SET
				data = jsonb_set(sessions.data, '{count}',
					to_jsonb(COALESCE((sessions.data->>'count')::int, 0) + 1)),
				updated_at = $2::timestamptz
			WHERE sessions.updated_at < $3::timestamptz
			RETURNING 1
		`, ip, at, day).Scan(&one)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil // already recorded today
		}
		return false, err
	}
	return true, nil
}

// NeedsGeo reports whether no row for ip carries geo yet, so enrichment runs once
// per address rather than once per (address, member).
func (s *SessionStore) NeedsGeo(ctx context.Context, ip string) (bool, error) {
	var missing bool
	err := s.pool.QueryRow(ctx,
		`SELECT NOT EXISTS(SELECT 1 FROM sessions WHERE ip = $1 AND data ? 'geo')`, ip).Scan(&missing)
	return missing, err
}

// SetGeo writes the enrichment onto every row for the address: the geolocation is
// a property of the address, not of any one connection from it.
func (s *SessionStore) SetGeo(ctx context.Context, ip string, info map[string]string) error {
	raw, err := json.Marshal(info)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx,
		`UPDATE sessions SET data = jsonb_set(data, '{geo}', $2::jsonb) WHERE ip = $1`, ip, raw)
	return err
}

// IsIpBanned reports whether ip has a ban row - the one query the guard runs on
// every request, so it is a single indexed existence check.
func (s *SessionStore) IsIpBanned(ctx context.Context, ip string) (bool, error) {
	var banned bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ip_bans WHERE ip = $1)`, ip).Scan(&banned)
	return banned, err
}

// BanIp adds or refreshes a ban (upsert, so re-banning updates the reason).
func (s *SessionStore) BanIp(ctx context.Context, ip, reason, bannedBy string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO ip_bans (ip, data)
		VALUES ($1, jsonb_build_object('reason', $2::text, 'bannedBy', $3::text))
		ON CONFLICT (ip) DO UPDATE
			SET data = jsonb_build_object('reason', $2::text, 'bannedBy', $3::text)
	`, ip, reason, bannedBy)
	return err
}

// UnbanIp lifts a ban by deleting the row, keeping the table to exactly the
// addresses banned right now.
func (s *SessionStore) UnbanIp(ctx context.Context, ip string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM ip_bans WHERE ip = $1`, ip)
	return err
}

// ListByUser returns every address a member has connected from, newest-first,
// each with its hit count, geolocation, and whether it is currently banned.
func (s *SessionStore) ListByUser(ctx context.Context, userID string) ([]models.ConnectionIP, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT s.ip,
		       COALESCE((s.data->>'count')::int, 0),
		       s.created_at,
		       s.updated_at,
		       (b.ip IS NOT NULL),
		       COALESCE(s.data->'geo', '{}'::jsonb)
		FROM sessions s
		LEFT JOIN ip_bans b ON b.ip = s.ip
		WHERE s.user_id = $1::uuid
		ORDER BY s.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ips := []models.ConnectionIP{}
	for rows.Next() {
		var ip models.ConnectionIP
		var geo []byte
		if err := rows.Scan(&ip.IP, &ip.Count, &ip.FirstSeen, &ip.LastSeen, &ip.Banned, &geo); err != nil {
			return nil, err
		}
		ip.Geo = decodeSessionGeo(geo)
		ips = append(ips, ip)
	}
	return ips, rows.Err()
}

// ListIps is the administration roster: one row per distinct address (every
// address seen in `sessions`, plus any banned before they were ever seen),
// carrying its members, hit count, geolocation and ban state. term filters by
// address or by a member's email or name; a blank term lists everything.
func (s *SessionStore) ListIps(ctx context.Context, term string, offset, limit int) ([]models.IpSummary, error) {
	rows, err := s.pool.Query(ctx, `
		WITH universe AS (
			SELECT ip FROM sessions
			UNION
			SELECT ip FROM ip_bans
		)
		SELECT u.ip,
		       (b.ip IS NOT NULL) AS banned,
		       COALESCE(b.data->>'reason', '') AS reason,
		       COALESCE(SUM(COALESCE((s.data->>'count')::int, 0)), 0)::int AS cnt,
		       MIN(s.created_at) AS first_seen,
		       MAX(s.updated_at) AS last_seen,
		       COALESCE((array_agg(s.data->'geo') FILTER (WHERE s.data ? 'geo'))[1], '{}'::jsonb) AS geo,
		       COALESCE(
		         jsonb_agg(DISTINCT jsonb_build_object(
		           'id', usr.id, 'email', usr.email,
		           'name', usr.data->>'name', 'surname', usr.data->>'surname', 'handle', usr.data->>'handle'
		         )) FILTER (WHERE usr.id IS NOT NULL), '[]'::jsonb) AS members
		FROM universe u
		LEFT JOIN sessions s ON s.ip = u.ip
		LEFT JOIN ip_bans b ON b.ip = u.ip
		LEFT JOIN users usr ON usr.id = s.user_id
		WHERE ($1 = ''
		       OR u.ip ILIKE '%' || $1 || '%'
		       OR usr.email ILIKE '%' || $1 || '%'
		       OR usr.data->>'name' ILIKE '%' || $1 || '%'
		       OR usr.data->>'surname' ILIKE '%' || $1 || '%')
		GROUP BY u.ip, b.ip, b.data
		ORDER BY last_seen DESC NULLS LAST, u.ip
		LIMIT $2 OFFSET $3
	`, term, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := []models.IpSummary{}
	for rows.Next() {
		var sm models.IpSummary
		var geo, members []byte
		if err := rows.Scan(&sm.IP, &sm.Banned, &sm.Reason, &sm.Count,
			&sm.FirstSeen, &sm.LastSeen, &geo, &members); err != nil {
			return nil, err
		}
		sm.Geo = decodeSessionGeo(geo)
		if len(members) > 0 {
			_ = json.Unmarshal(members, &sm.Members)
		}
		if sm.Members == nil {
			sm.Members = []models.IpMember{}
		}
		summaries = append(summaries, sm)
	}
	return summaries, rows.Err()
}

// Stats counts the distinct addresses seen and how many are banned.
func (s *SessionStore) Stats(ctx context.Context) (models.IpStats, error) {
	var out models.IpStats
	err := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM (SELECT ip FROM sessions UNION SELECT ip FROM ip_bans) t),
			(SELECT COUNT(*) FROM ip_bans)
	`).Scan(&out.Total, &out.Banned)
	return out, err
}

// decodeSessionGeo turns a jsonb geo payload into a map, dropping an empty object
// to nil so it serializes away rather than as "geo": {}.
func decodeSessionGeo(raw []byte) models.GeoInfo {
	if len(raw) == 0 {
		return nil
	}
	var geo models.GeoInfo
	if err := json.Unmarshal(raw, &geo); err != nil || len(geo) == 0 {
		return nil
	}
	return geo
}
