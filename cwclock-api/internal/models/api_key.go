package models

import "time"

// ApiKey lets a user authenticate scripts via the X-Api-Key header instead
// of a JWT. KeyHash is the sha256 of the plaintext token; the plaintext is
// only ever returned once, at creation time (see ApiKeyCreated).
type ApiKey struct {
	ID          string     `json:"id"`
	UserID      string     `json:"userId"`
	KeyHash     string     `json:"-"`
	Description string     `json:"description"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// ApiKeyCreated is returned only from the create endpoint: it's the only
// time the plaintext Token is ever available.
type ApiKeyCreated struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	Token       string     `json:"token"`
}
