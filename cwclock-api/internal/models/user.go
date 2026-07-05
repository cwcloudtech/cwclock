package models

import "time"

// GlobalRole is the account-wide role, separate from a user's per-organization
// Role (owner/admin/member/reader).
type GlobalRole string

const (
	GlobalRoleSuperuser GlobalRole = "superuser"
	GlobalRoleConfirmed GlobalRole = "confirmed"
	GlobalRoleDisabled  GlobalRole = "disabled"
)

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	Surname      string     `json:"surname"`
	Role         GlobalRole `json:"role"`
	PasswordHash string     `json:"-"`
	Picture      string     `json:"picture,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type UserResponse struct {
	ID      string     `json:"id"`
	Email   string     `json:"email"`
	Name    string     `json:"name"`
	Surname string     `json:"surname"`
	Role    GlobalRole `json:"role"`
	Token   string     `json:"token"`
	Picture string     `json:"picture,omitempty"`
}

type UserMeResponse struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Surname   string     `json:"surname"`
	Role      GlobalRole `json:"role"`
	Picture   string     `json:"picture,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}
