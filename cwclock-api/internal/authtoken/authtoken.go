// Package authtoken mints the session JWT shared by password login,
// registration and OIDC login, so all three issue tokens middleware.Auth can
// verify the same way.
package authtoken

import (
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
