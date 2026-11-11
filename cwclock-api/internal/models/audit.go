package models

import "time"

// Audit actions. They are namespaced by what they are about so the superadmin
// screen can group and translate them (its dictionary mirrors this shape
// under audit.actions.*), and so a Prometheus query can match a whole family
// with one regex.
//
// Only acts that change who can do what, or what the deployment is
// configured to do, are in this list. Ordinary bookkeeping - adding a time
// entry, generating an invoice, editing a project - is not audited: it is
// already visible in the data itself, and burying the handful of
// security-relevant acts under thousands of routine ones would defeat the
// point of the trail. Reads are never audited.
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
	AuditSubjectApiKey       = "apiKey"
	AuditSubjectCredential   = "credential"
	AuditSubjectConnection   = "connection"
	AuditSubjectExportJob    = "exportJob"
)

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
// controls: the distinct actions present (for the filter) and whether a next
// page exists.
type AuditPage struct {
	Entries []AuditEntry `json:"entries"`
	Actions []string     `json:"actions"`
	HasMore bool         `json:"hasMore"`
}
