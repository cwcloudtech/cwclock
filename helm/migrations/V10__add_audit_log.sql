-- Append-only trail of the acts that change who can do what, or what this
-- deployment is configured to do (see internal/store/audit_store.go).
--
-- There is deliberately no updated_at and no code path that updates or
-- deletes a row: an audit trail that can be rewritten is not one. The table
-- is only ever INSERTed into and SELECTed from.
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Both references are nullable with ON DELETE SET NULL rather than
    -- CASCADE: deleting an account or an organization must not erase the
    -- record of what it did - deleting it is itself one of the audited acts.
    -- Because the id then goes away, the acting account's email is
    -- snapshotted into data->>'actorEmail' when the entry is written, so an
    -- entry still names who acted afterwards. user_id is also null for the
    -- few acts with no authenticated actor (finishing a password reset from
    -- an emailed link), and organization_id for everything that belongs to
    -- no organization: signing in or out, and anything a superuser does
    -- deployment-wide.
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    action VARCHAR(64) NOT NULL,
    subject_type VARCHAR(32) NOT NULL,
    -- Not a UUID: a subject is sometimes identified by something else (an
    -- external connection's id, a WebAuthn credential id, an email).
    subject_id VARCHAR(255) NOT NULL DEFAULT '',
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- The screen reads the trail newest first and pages through it. The id is in
-- the index as a tiebreaker because two entries written in the same
-- transaction share created_at to the microsecond, and a paged query whose
-- order isn't total can show or skip a row between two pages.
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at DESC, id DESC);

-- "What happened to this organization / this account?" - the other way the
-- trail gets read.
CREATE INDEX idx_audit_log_subject ON audit_log(subject_type, subject_id);
