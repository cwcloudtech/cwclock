// Package authtoken mints the session JWT shared by password login,
// registration and OIDC login, so all three issue tokens middleware.Auth can
// verify the same way.
package authtoken

import (
	"cwclock-api/internal/utils"
	"encoding/json"
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
	// PurposeMFALogin scopes the short-lived token handed back by Login when
	// the password was correct but the account has MFA enabled (see
	// ai-instruct-68) - it authorizes finishing the login via one of the
	// /v1/users/login/mfa/* endpoints, nothing else.
	PurposeMFALogin = "mfa_login"
	// PurposeWebAuthnCeremony scopes a token carrying a JSON-encoded
	// webauthn.SessionData between a WebAuthn "begin" and "finish" call
	// (registration or login). Using a signed token instead of server-side
	// session storage keeps every MFA endpoint stateless, the same way
	// PurposeConfirmAccount/PurposeResetPassword already work.
	PurposeWebAuthnCeremony = "webauthn_ceremony"
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
		return utils.EMPTY, ErrInvalidPurposeToken
	}
	if p, _ := claims["purpose"].(string); p != purpose {
		return utils.EMPTY, ErrInvalidPurposeToken
	}
	sub, ok := claims["sub"].(string)
	if !ok || utils.IsBlank(sub) {
		return utils.EMPTY, ErrInvalidPurposeToken
	}
	return sub, nil
}

// GeneratePurposeWithData is GeneratePurpose plus an arbitrary JSON-able
// payload carried in a "data" claim - used for PurposeWebAuthnCeremony,
// which needs to round-trip a whole webauthn.SessionData value between a
// "begin" and "finish" call.
func GeneratePurposeWithData(secret, userID, purpose string, ttl time.Duration, data any) (string, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return utils.EMPTY, err
	}
	claims := jwt.MapClaims{
		"sub":     userID,
		"purpose": purpose,
		"data":    string(raw),
		"exp":     time.Now().Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParsePurposeWithData is ParsePurpose plus decoding the "data" claim set by
// GeneratePurposeWithData into dataOut.
func ParsePurposeWithData(secret, tokenString, purpose string, dataOut any) (userID string, err error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return utils.EMPTY, ErrInvalidPurposeToken
	}
	if p, _ := claims["purpose"].(string); p != purpose {
		return utils.EMPTY, ErrInvalidPurposeToken
	}
	sub, ok := claims["sub"].(string)
	if !ok || utils.IsBlank(sub) {
		return utils.EMPTY, ErrInvalidPurposeToken
	}
	if raw, _ := claims["data"].(string); dataOut != nil && utils.IsNotBlank(raw) {
		if err := json.Unmarshal([]byte(raw), dataOut); err != nil {
			return utils.EMPTY, ErrInvalidPurposeToken
		}
	}
	return sub, nil
}
