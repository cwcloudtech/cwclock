package middleware

import (
	"net/http"

	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
)

// RequireActiveUser blocks disabled accounts from every authenticated action
// beyond logging in and reading their own status. The role is re-read from
// the database on every request since it can change after the token was
// issued (eg. the superuser confirms or disables the account later).
func RequireActiveUser(users *store.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, _ := UserIDFromContext(r.Context())
			user, err := users.FindByID(r.Context(), userID)
			if err != nil {
				jsonError(w, http.StatusUnauthorized, "Not Authorised")
				return
			}
			if user.Role == models.GlobalRoleDisabled {
				jsonError(w, http.StatusForbidden, "Your account is disabled. Please contact an administrator.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireSuperuser restricts a route to the account holding the global
// superuser role, used for the user-management screen and for touching
// organizations the caller isn't otherwise a member of.
func RequireSuperuser(users *store.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, _ := UserIDFromContext(r.Context())
			user, err := users.FindByID(r.Context(), userID)
			if err != nil || user.Role != models.GlobalRoleSuperuser {
				jsonError(w, http.StatusForbidden, "Superuser access required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
