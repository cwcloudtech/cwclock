package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"cwclock-api/internal/utils"
)

type contextKey string

const userIDKey contextKey = "userID"

// ApiKeyVerifier resolves the user an API key token's hash belongs to. It's
// a narrow interface (rather than importing *store.ApiKeyStore directly) so
// this package doesn't depend on store.
type ApiKeyVerifier interface {
	VerifyHash(ctx context.Context, hash string) (userID string, err error)
}

// Auth authenticates a request either via the X-Api-Key header or a JWT
// Bearer token, in that order: when X-Api-Key is present it is authoritative
// (a bad key is rejected outright, it never falls back to the JWT), matching
// "prior to the JWT token if both are present". Both paths only ever set the
// same userIDKey context value, so everything downstream (RequireActiveUser,
// OrgMembership, ...) is unaffected by which one authenticated the request.
func Auth(secret string, apiKeys ApiKeyVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if apiKey := r.Header.Get("X-Api-Key"); apiKey != "" {
				userID, err := apiKeys.VerifyHash(r.Context(), utils.HashToken(apiKey))
				if err != nil || utils.IsBlank(userID) {
					unauthorized(w, "Not Authorised")
					return
				}
				ctx := context.WithValue(r.Context(), userIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				unauthorized(w, "Not athorised,no token")
				return
			}
			tokenString := strings.TrimPrefix(header, "Bearer ")

			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				unauthorized(w, "Not Authorised")
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok || utils.IsBlank(sub) {
				unauthorized(w, "Not Authorised")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}
