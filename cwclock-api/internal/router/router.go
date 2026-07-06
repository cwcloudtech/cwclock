package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"cwclock-api/internal/handlers"
	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/store"
)

func New(
	userHandler *handlers.UserHandler,
	orgHandler *handlers.OrganizationHandler,
	clientHandler *handlers.ClientHandler,
	projectHandler *handlers.ProjectHandler,
	timeEntryHandler *handlers.TimeEntryHandler,
	adminHandler *handlers.AdminHandler,
	reportHandler *handlers.ReportHandler,
	orgs *store.OrgStore,
	users *store.UserStore,
	jwtSecret string,
) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello Welcome to cwclock Backend"))
	})

	r.Route("/v1", func(r chi.Router) {
		r.Get("/currencies", handlers.ListCurrencies)

		r.Route("/users", func(r chi.Router) {
			r.Post("/", userHandler.Register)
			r.Post("/login", userHandler.Login)

			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(jwtSecret))
				r.Get("/me", userHandler.Me)

				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireActiveUser(users))
					r.Put("/me", userHandler.UpdateProfile)
					r.Put("/me/picture", userHandler.UpdatePicture)
					r.Get("/search", userHandler.Search)
				})
			})
		})

		r.Route("/admin/users", func(r chi.Router) {
			r.Use(middleware.Auth(jwtSecret))
			r.Use(middleware.RequireActiveUser(users))
			r.Use(middleware.RequireSuperuser(users))
			r.Get("/", adminHandler.ListUsers)
			r.Put("/{id}", adminHandler.UpdateUser)
			r.Delete("/{id}", adminHandler.DeleteUser)
		})

		r.Route("/admin/organizations", func(r chi.Router) {
			r.Use(middleware.Auth(jwtSecret))
			r.Use(middleware.RequireActiveUser(users))
			r.Use(middleware.RequireSuperuser(users))
			r.Get("/", orgHandler.AdminList)
		})

		r.Route("/organizations", func(r chi.Router) {
			r.Use(middleware.Auth(jwtSecret))
			r.Use(middleware.RequireActiveUser(users))
			r.Post("/", orgHandler.Create)
			r.Get("/", orgHandler.List)

			r.Route("/{orgId}", func(r chi.Router) {
				r.Use(middleware.OrgMembership(orgs, users))

				r.Get("/", orgHandler.Get)
				r.With(middleware.RequireRole(models.RoleOwner)).Put("/", orgHandler.Update)
				r.With(middleware.RequireRole(models.RoleOwner)).Delete("/", orgHandler.Delete)
				r.With(middleware.RequireRole(models.RoleOwner)).Put("/owner", orgHandler.TransferOwnership)

				r.Route("/members", func(r chi.Router) {
					r.Get("/", orgHandler.ListMembers)
					r.With(middleware.RequireRole(models.RoleOwner)).Post("/", orgHandler.AddMember)
					r.With(middleware.RequireRole(models.RoleOwner)).Put("/{userId}", orgHandler.UpdateMember)
					r.With(middleware.RequireRole(models.RoleOwner)).Delete("/{userId}", orgHandler.RemoveMember)
					r.With(middleware.RequireRole(models.RoleAdmin)).Put("/{userId}/rate", orgHandler.SetMemberRate)
				})

				r.Route("/clients", func(r chi.Router) {
					r.Get("/", clientHandler.List)
					r.With(middleware.RequireRole(models.RoleAdmin)).Post("/", clientHandler.Create)
					r.With(middleware.RequireRole(models.RoleAdmin)).Put("/{clientId}", clientHandler.Update)
					r.With(middleware.RequireRole(models.RoleAdmin)).Delete("/{clientId}", clientHandler.Delete)

					r.Route("/{clientId}/projects", func(r chi.Router) {
						r.Get("/", projectHandler.List)
						r.With(middleware.RequireRole(models.RoleAdmin)).Post("/", projectHandler.Create)
					})
				})

				r.Route("/projects", func(r chi.Router) {
					r.Get("/", projectHandler.List)
					r.With(middleware.RequireRole(models.RoleAdmin)).Put("/{projectId}", projectHandler.Update)
					r.With(middleware.RequireRole(models.RoleAdmin)).Delete("/{projectId}", projectHandler.Delete)
				})

				r.Route("/time-entries", func(r chi.Router) {
					r.Get("/", timeEntryHandler.List)
					r.With(middleware.RequireRole(models.RoleMember)).Post("/", timeEntryHandler.Create)
					r.With(middleware.RequireRole(models.RoleMember)).Put("/{id}", timeEntryHandler.Update)
					r.With(middleware.RequireRole(models.RoleMember)).Delete("/{id}", timeEntryHandler.Delete)
				})

				r.Route("/reports", func(r chi.Router) {
					r.Use(middleware.RequireRole(models.RoleMember))
					r.Get("/", reportHandler.Get)
					r.Get("/export", reportHandler.Export)
				})
			})
		})
	})

	return r
}
