package models

// ActivationMode controls how a newly registered (non-first) account moves
// from disabled to confirmed: an administrator flips its role by hand
// (ActivationModeAdmin, the default) or the user follows a confirmation
// link emailed to them (ActivationModeEmail).
const (
	ActivationModeAdmin = "admin"
	ActivationModeEmail = "email"
)

// IsValidActivationMode reports whether mode is a known activation mode.
func IsValidActivationMode(mode string) bool {
	return mode == ActivationModeAdmin || mode == ActivationModeEmail
}

// I18n codes describing why an account is blocked, shared between the
// middleware layer (which enforces the block) and the handlers layer (which
// surfaces the same reason on /me and login/register responses) so both
// agree on the same wording without importing one another.
const (
	I18nAccountDisabledAdmin = "errors.accountDisabledAdmin"
	I18nAccountDisabledEmail = "errors.accountDisabledEmail"
	I18nAccountBanned        = "errors.accountBanned"
)

// I18nCodeForRole returns the i18n code describing why an account with role
// is blocked, given the server's current activation mode - or "" for a role
// that isn't blocked at all.
func I18nCodeForRole(role GlobalRole, activationMode string) string {
	switch role {
	case GlobalRoleBan:
		return I18nAccountBanned
	case GlobalRoleDisabled:
		if activationMode == ActivationModeEmail {
			return I18nAccountDisabledEmail
		}
		return I18nAccountDisabledAdmin
	default:
		return ""
	}
}
