package models

import "time"

// Audit actions. They are namespaced by what they are about so the superadmin
// screen can group and translate them (its dictionary mirrors this shape
// under audit.actions.*), and so a Prometheus query can match a whole family
// with one regex.
//
// Two families are in this list. The first is the acts that change who can
// do what, or what the deployment is configured to do. The second is the
// time records themselves, and the projects they are logged against: those
// are not security acts, but they are the billable facts every report and
// invoice is built from, so who added, moved or removed one - and whose
// time it was - has to stay answerable after the row itself is gone.
//
// Records are by far the highest-volume family here, which is the reason
// every action is namespaced: a superadmin filtering on anything else, or a
// Prometheus query matching one family, is not drowned by them. Generating
// an invoice is still left out - the invoice is itself the durable record of
// its own generation, numbered and stored. Reads are never audited.
const (
	// Sessions. Only successful sign-ins are recorded: a failed attempt has
	// no authenticated actor to attribute it to, and auditing it would let
	// anyone fill the trail (and the metric's label space) by hammering the
	// login endpoint.
	AuditSignedIn  = "auth.signedIn"
	AuditSignedOut = "auth.signedOut"
	// Account lifecycle.
	AuditAccountRegistered    = "account.registered"
	AuditAccountConfirmed     = "account.confirmed"
	AuditAccountPasswordReset = "account.passwordReset"
	// Administrative acts on an account, by a superuser.
	AuditUserRoleChanged    = "user.roleChanged"
	AuditUserUpdatedByAdmin = "user.updatedByAdmin"
	AuditUserDeleted        = "user.deleted"
	AuditUserMFACleared     = "user.mfaCleared"
	// Second factors and other credentials a user manages themselves.
	AuditMFATOTPEnabled          = "mfa.totpEnabled"
	AuditMFATOTPDisabled         = "mfa.totpDisabled"
	AuditMFAWebAuthnRegistered   = "mfa.webauthnRegistered"
	AuditMFAWebAuthnDeleted      = "mfa.webauthnDeleted"
	AuditApiKeyCreated           = "apiKey.created"
	AuditApiKeyDeleted           = "apiKey.deleted"
	AuditCalendarFeedEnabled     = "calendarFeed.enabled"
	AuditCalendarFeedDisabled    = "calendarFeed.disabled"
	AuditCalendarFeedRegenerated = "calendarFeed.regenerated"
	// Organizations: their existence, who is in them with what role, and the
	// external credentials they hold.
	AuditOrgCreated              = "organization.created"
	AuditOrgUpdated              = "organization.updated"
	AuditOrgDeleted              = "organization.deleted"
	AuditOrgOwnershipTransferred = "organization.ownershipTransferred"
	AuditOrgMemberAdded          = "organization.memberAdded"
	AuditOrgMemberRoleChanged    = "organization.memberRoleChanged"
	AuditOrgMemberRemoved        = "organization.memberRemoved"
	AuditOrgMemberRateChanged    = "organization.memberRateChanged"
	AuditOrgConnectionAdded      = "organization.connectionAdded"
	AuditOrgConnectionRemoved    = "organization.connectionRemoved"
	// Clients, because transferring one moves billing data between
	// organizations and deleting one takes its projects and time with it.
	AuditClientCreated     = "client.created"
	AuditClientUpdated     = "client.updated"
	AuditClientDeleted     = "client.deleted"
	AuditClientTransferred = "client.transferred"
	// Projects, because a project carries the daily rate its client is
	// billed at, and deleting one takes every hour logged against it.
	AuditProjectCreated = "project.created"
	AuditProjectUpdated = "project.updated"
	AuditProjectDeleted = "project.deleted"
	// Time records - the hours themselves. Every one of these entries names
	// the member whose time it is as well as the account that acted: the two
	// differ whenever an admin edits or reassigns someone else's day, which
	// is exactly the case the trail is read for.
	AuditRecordCreated = "record.created"
	AuditRecordUpdated = "record.updated"
	AuditRecordDeleted = "record.deleted"
	// A CSV import creates clients, projects, accounts and hundreds of
	// records in one request. It is recorded once with its counts rather than
	// once per row: a per-row trail of a bulk load says nothing the single
	// entry doesn't, and would bury every other act on the screen.
	AuditRecordImported = "record.imported"
	// Export jobs are configuration: a schedule, a set of recipients, and
	// (in their targets) credentials for an S3 bucket, a Drive or a git
	// remote. Running one on demand pushes data out of the deployment, so
	// that is recorded too.
	AuditExportJobCreated = "exportJob.created"
	AuditExportJobUpdated = "exportJob.updated"
	AuditExportJobDeleted = "exportJob.deleted"
	AuditExportJobRun     = "exportJob.run"
)

// Audit subject types - what the action's subject_id points at.
const (
	AuditSubjectUser         = "user"
	AuditSubjectOrganization = "organization"
	AuditSubjectClient       = "client"
	AuditSubjectProject      = "project"
	AuditSubjectRecord       = "record"
	AuditSubjectApiKey       = "apiKey"
	AuditSubjectCredential   = "credential"
	AuditSubjectConnection   = "connection"
	AuditSubjectExportJob    = "exportJob"
)

// AuditActions and AuditSubjectTypes are the constants above as lists, for
// the superadmin screen's two filters. They are the code's own list rather
// than a SELECT DISTINCT over audit_log, which is how the screen built its
// filter while the trail held only access and configuration acts. Now that
// every time record lands in it as well, that table grows with the app's
// busiest write, and scanning it on each page load would be the one
// expensive thing on a screen that is otherwise two indexed queries.
//
// They have to be kept in step with the constants above, which is why they
// sit right under them: adding an action and forgetting its entry here shows
// up in a single diff. The cost of being out of step is small in one
// direction (a missing filter option) and nil in the other (an option that
// matches nothing yet simply comes back empty).
var AuditActions = []string{
	AuditSignedIn,
	AuditSignedOut,
	AuditAccountRegistered,
	AuditAccountConfirmed,
	AuditAccountPasswordReset,
	AuditUserRoleChanged,
	AuditUserUpdatedByAdmin,
	AuditUserDeleted,
	AuditUserMFACleared,
	AuditMFATOTPEnabled,
	AuditMFATOTPDisabled,
	AuditMFAWebAuthnRegistered,
	AuditMFAWebAuthnDeleted,
	AuditApiKeyCreated,
	AuditApiKeyDeleted,
	AuditCalendarFeedEnabled,
	AuditCalendarFeedDisabled,
	AuditCalendarFeedRegenerated,
	AuditOrgCreated,
	AuditOrgUpdated,
	AuditOrgDeleted,
	AuditOrgOwnershipTransferred,
	AuditOrgMemberAdded,
	AuditOrgMemberRoleChanged,
	AuditOrgMemberRemoved,
	AuditOrgMemberRateChanged,
	AuditOrgConnectionAdded,
	AuditOrgConnectionRemoved,
	AuditClientCreated,
	AuditClientUpdated,
	AuditClientDeleted,
	AuditClientTransferred,
	AuditProjectCreated,
	AuditProjectUpdated,
	AuditProjectDeleted,
	AuditRecordCreated,
	AuditRecordUpdated,
	AuditRecordDeleted,
	AuditRecordImported,
	AuditExportJobCreated,
	AuditExportJobUpdated,
	AuditExportJobDeleted,
	AuditExportJobRun,
}

var AuditSubjectTypes = []string{
	AuditSubjectUser,
	AuditSubjectOrganization,
	AuditSubjectClient,
	AuditSubjectProject,
	AuditSubjectRecord,
	AuditSubjectApiKey,
	AuditSubjectCredential,
	AuditSubjectConnection,
	AuditSubjectExportJob,
}

// AuditEntry is one recorded act, as the superadmin screen reads it.
type AuditEntry struct {
	ID string `json:"id"`
	// UserID/OrganizationID are empty when the row's reference is null -
	// either the act had no actor/organization, or the row it pointed at has
	// since been deleted (see V10__add_audit_log.sql).
	UserID         string `json:"userId,omitempty"`
	OrganizationID string `json:"organizationId,omitempty"`
	// UserEmail is the acting account's email as it was when the act
	// happened, snapshotted into the entry rather than joined at read time:
	// renaming or deleting an account must not rewrite what the trail says
	// about what it did.
	UserEmail   string         `json:"userEmail,omitempty"`
	Action      string         `json:"action"`
	SubjectType string         `json:"subjectType"`
	SubjectID   string         `json:"subjectId,omitempty"`
	Detail      map[string]any `json:"detail,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
}

// AuditPage is one page of the trail plus what the screen needs to render its
// controls: the values its two filters offer, and whether a next page exists.
// The filter values are sent with the page rather than hardcoded in the UI so
// that an older screen against a newer API still offers everything the API
// can filter on.
type AuditPage struct {
	Entries      []AuditEntry `json:"entries"`
	Actions      []string     `json:"actions"`
	SubjectTypes []string     `json:"subjectTypes"`
	HasMore      bool         `json:"hasMore"`
}
