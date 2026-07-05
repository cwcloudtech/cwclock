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

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
