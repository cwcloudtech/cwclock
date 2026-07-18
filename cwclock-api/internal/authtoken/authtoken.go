// Package authtoken mints the session JWT shared by password login,
// registration and OIDC login, so all three issue tokens middleware.Auth can
// verify the same way.
package authtoken

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Generate signs a 20-day session token for userID.
func Generate(secret, userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(20 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Purpose-scoped token kinds minted by GeneratePurpose - each carries a
// "purpose" claim so a confirmation or password-reset link can never be
// replayed as a session token (or vice versa) even though they share the
// same signing secret.
const (
	PurposeConfirmAccount = "confirm_account"
	PurposeResetPassword  = "reset_password"
)

// ErrInvalidPurposeToken is returned by ParsePurpose for a token that is
// missing, expired, tampered with, or minted for a different purpose.
var ErrInvalidPurposeToken = errors.New("invalid or expired token")

// GeneratePurpose signs a purpose-scoped token for userID, valid for ttl.
// Used for confirmation and password-reset links sent by email.
func GeneratePurpose(secret, userID, purpose string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":     userID,
		"purpose": purpose,
		"exp":     time.Now().Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParsePurpose verifies a purpose-scoped token minted by GeneratePurpose and
// returns the user id it was issued for.
func ParsePurpose(secret, tokenString, purpose string) (userID string, err error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", ErrInvalidPurposeToken
	}
	if p, _ := claims["purpose"].(string); p != purpose {
		return "", ErrInvalidPurposeToken
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", ErrInvalidPurposeToken
	}
	return sub, nil
}
