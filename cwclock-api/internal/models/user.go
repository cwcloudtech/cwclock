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
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
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

type UserMeResponse struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Surname   string     `json:"surname"`
	Role      GlobalRole `json:"role"`
	Picture   string     `json:"picture,omitempty"`
	PictureX  float64    `json:"pictureX"`
	PictureY  float64    `json:"pictureY"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	// I18nCode is set when Role is disabled or ban, so the frontend can
	// display the right explanation without hardcoding the server's
	// activation mode (see I18nCodeForRole).
	I18nCode string `json:"i18nCode,omitempty"`
}
