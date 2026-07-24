package models

import "time"

// GlobalRole is the account-wide role, separate from a user's per-organization
// Role (owner/admin/member/reader).
type GlobalRole string

const (
	GlobalRoleSuperuser GlobalRole = "superuser"
	GlobalRoleConfirmed GlobalRole = "confirmed"
	GlobalRoleDisabled  GlobalRole = "disabled"
	// GlobalRoleBan is like GlobalRoleDisabled (blocks every action beyond
	// login/reading own status) except it's a deliberate administrator
	// action rather than a pending-approval state, so it carries its own
	// i18n code and can never be lifted by a confirmation link or password
	// renewal (see I18nAccountBanned).
	GlobalRoleBan GlobalRole = "ban"
)

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	Surname      string     `json:"surname"`
	Role         GlobalRole `json:"role"`
	PasswordHash string     `json:"-"`
	Picture      string     `json:"picture,omitempty"`
	PictureX     float64    `json:"pictureX"`
	PictureY     float64    `json:"pictureY"`
	// MFAEnabled is true once at least one MFA factor (TOTP or a WebAuthn
	// credential) has been confirmed - see ai-instruct-68. It gates password
	// login (handlers.UserHandler.Login) behind a second factor.
	MFAEnabled bool `json:"-"`
	// MFATOTPSecret is the base32-encoded TOTP shared secret. It's written as
	// soon as enrollment starts (handlers.MFAHandler.SetupTOTP) but MFAEnabled
	// stays false until the user confirms a code generated from it.
	MFATOTPSecret string `json:"-"`
	// CalendarFeedToken authenticates the public, unauthenticated calendar
	// feed URL (see ai-instruct-85's "share the calendar with Outlook or
	// Google Calendar" - a subscribable ICS feed, since neither provider can
	// send an Authorization header when polling a subscribed calendar).
	// CalendarFeedEnabled gates the feed even once a token exists, so
	// disabling sharing doesn't require throwing away (and needing to
	// re-share) the URL.
	CalendarFeedToken   string    `json:"-"`
	CalendarFeedEnabled bool      `json:"-"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type UserResponse struct {
	ID       string     `json:"id"`
	Email    string     `json:"email"`
	Name     string     `json:"name"`
	Surname  string     `json:"surname"`
	Role     GlobalRole `json:"role"`
	Token    string     `json:"token"`
	Picture  string     `json:"picture,omitempty"`
	PictureX float64    `json:"pictureX"`
	PictureY float64    `json:"pictureY"`
	// I18nCode is set when Role is disabled or ban, so the frontend can
	// display the right explanation without hardcoding the server's
	// activation mode (see I18nCodeForRole).
	I18nCode string `json:"i18nCode,omitempty"`
}

// MFAChallengeResponse replaces UserResponse when password login succeeds
// but the account has MFA enabled (see ai-instruct-68): no session Token is
// issued yet, only a short-lived ChallengeToken the client exchanges for one
// via one of the /v1/users/login/mfa/* endpoints once the second factor is
// verified.
type MFAChallengeResponse struct {
	MFARequired    bool   `json:"mfaRequired"`
	ChallengeToken string `json:"challengeToken"`
	HasTOTP        bool   `json:"hasTotp"`
	HasWebAuthn    bool   `json:"hasWebAuthn"`
}

type UserMeResponse struct {
	ID         string     `json:"id"`
	Email      string     `json:"email"`
	Name       string     `json:"name"`
	Surname    string     `json:"surname"`
	Role       GlobalRole `json:"role"`
	Picture    string     `json:"picture,omitempty"`
	PictureX   float64    `json:"pictureX"`
	PictureY   float64    `json:"pictureY"`
	MFAEnabled bool       `json:"mfaEnabled"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	// I18nCode is set when Role is disabled or ban, so the frontend can
	// display the right explanation without hardcoding the server's
	// activation mode (see I18nCodeForRole).
	I18nCode string `json:"i18nCode,omitempty"`
}
