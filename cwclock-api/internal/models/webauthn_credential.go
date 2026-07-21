package models

import "time"

// WebAuthnCredential is a hardware or platform security key (e.g. a YubiKey)
// registered as an MFA factor (see ai-instruct-68). CredentialID, PublicKey
// and SignCount are never serialized to the client - they're only used
// server-side to verify a login assertion.
type WebAuthnCredential struct {
	ID           string    `json:"id"`
	UserID       string    `json:"-"`
	CredentialID []byte    `json:"-"`
	PublicKey    []byte    `json:"-"`
	SignCount    uint32    `json:"-"`
	Transports   []string  `json:"transports,omitempty"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
