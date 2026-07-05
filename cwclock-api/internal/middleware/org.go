package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
)

type orgContextKey string

const (
	orgIDKey   orgContextKey = "orgID"
	orgRoleKey orgContextKey = "orgRole"
)

var roleRank = map[models.Role]int{
	models.RoleReader: 0,
	models.RoleMember: 1,
	models.RoleAdmin:  2,
	models.RoleOwner:  3,
}

// OrgMembership resolves the caller's role in the organization identified by
// the {orgId} URL param and rejects the request if they aren't a member.
// The global superuser is granted an implicit owner role in every
// organization, even ones they never joined, so they can manage or delete
// any organization and transfer its ownership.
func OrgMembership(orgs *store.OrgStore, users *store.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, _ := UserIDFromContext(r.Context())
			orgID := chi.URLParam(r, "orgId")

			if user, err := users.FindByID(r.Context(), userID); err == nil && user.Role == models.GlobalRoleSuperuser {
				ctx := context.WithValue(r.Context(), orgIDKey, orgID)
				ctx = context.WithValue(ctx, orgRoleKey, models.RoleOwner)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			role, err := orgs.GetRole(r.Context(), orgID, userID)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					jsonError(w, http.StatusForbidden, "Not a member of this organization")
					return
				}
				jsonError(w, http.StatusInternalServerError, err.Error())
				return
			}

			ctx := context.WithValue(r.Context(), orgIDKey, orgID)
			ctx = context.WithValue(ctx, orgRoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole rejects the request unless the caller's resolved role in the
// organization is at least as privileged as min (reader < member < admin < owner).
func RequireRole(min models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := OrgRoleFromContext(r.Context())
			if roleRank[role] < roleRank[min] {
				jsonError(w, http.StatusForbidden, "Insufficient role for this action")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func OrgIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(orgIDKey).(string)
	return id, ok
}

func OrgRoleFromContext(ctx context.Context) (models.Role, bool) {
	role, ok := ctx.Value(orgRoleKey).(models.Role)
	return role, ok
}

func jsonError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
