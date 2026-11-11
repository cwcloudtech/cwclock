# AI instruction 130

## cwclock-api / cwclock-ui

I want an audit trail, visible to the superadmins, like the one in `~/shadai`.

An append-only `audit_log` table (migration in `helm/migrations`): id, user_id
(nullable, `ON DELETE SET NULL`), organization_id (nullable, for the acts that
belong to no organization), action, subject_type, subject_id, data JSONB,
created_at, indexed on created_at desc and on (subject_type, subject_id). No
update, no delete. A store beside the others with a `Record(...)` that never
fails the caller's action - a failed audit write is logged and the action
stands, an audit trail must not be a single point of failure for the app.

Call it from the acts worth recording: sign-in and sign-out, account role
changes and deletions, MFA resets, organization/client create, update, delete
and ownership transfer, credential and configuration changes, and anything
else that changes who can do what or what the deployment is configured to do.
Not the reads.

Every entry also goes through `log/slog` at info with structured fields, so it
reaches stdout and the OTEL collector through the existing `internal/telemetry`
handler. And a counter of audit events labelled by action, tenant and user, on
the Prometheus endpoint and the OTLP meter.

In `cwclock-ui`, a superadmin screen: the trail newest first, paginated,
filterable by action and by subject, with the acting user's email resolved
server-side. Superadmin only, enforced on the route, not just in the UI.
