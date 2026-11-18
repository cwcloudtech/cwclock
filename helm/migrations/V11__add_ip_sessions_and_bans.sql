-- IP connection log (sessions) + IP ban list, ported from the sf-api / uprodit
-- "IP ban + geo + IP↔user association" feature. Follows the cwclock convention:
-- a uuid id, typed FK + timestamp columns, and everything else in a JSONB `data`.
--
--   sessions.data : { "count": <distinct days connected>, "geo": { country, city, ... } }
--   ip_bans.data  : { "reason": <motif>, "bannedBy": <superuser id> }

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    ip TEXT NOT NULL,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One row per (ip, member) and a separate row per anonymous ip. The two partial
-- unique indexes let the connection upsert reconcile each bucket independently
-- (ON CONFLICT names the matching one).
CREATE UNIQUE INDEX idx_sessions_ip_user ON sessions (ip, user_id) WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX idx_sessions_ip_anon ON sessions (ip) WHERE user_id IS NULL;
CREATE INDEX idx_sessions_user ON sessions (user_id, updated_at DESC);
CREATE INDEX idx_sessions_ip ON sessions (ip);

CREATE TABLE ip_bans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip TEXT NOT NULL UNIQUE,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ip_bans_ip ON ip_bans (ip);
