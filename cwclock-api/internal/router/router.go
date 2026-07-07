package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"cwclock-api/internal/handlers"
	"cwclock-api/internal/middleware"
	"cwclock-api/internal/models"
	"cwclock-api/internal/openapi"
	"cwclock-api/internal/store"
	"cwclock-api/internal/telemetry"
)

func New(
	userHandler *handlers.UserHandler,
	orgHandler *handlers.OrganizationHandler,
	clientHandler *handlers.ClientHandler,
	projectHandler *handlers.ProjectHandler,
	timeEntryHandler *handlers.TimeEntryHandler,
	reportHandler *handlers.ReportHandler,
	adminHandler *handlers.AdminHandler,
	importHandler *handlers.ImportHandler,
	orgs *store.OrgStore,
	users *store.UserStore,
	jwtSecret string,
	corsEnabled bool,
	corsAllowedOrigins []string,
	apiVersion string,
	manifestPath string,
	tel *telemetry.Providers,
	observe middleware.EndpointObserver,
	metricsHandler http.Handler,
) http.Handler {
	r := chi.NewRouter()

	// Spans + request logs (+ metrics, via observe) for every endpoint call.
	r.Use(middleware.Instrument(tel.Tracer, tel.Logger, observe))

	if corsEnabled {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins: corsAllowedOrigins,
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Authorization", "Content-Type"},
			// Content-Disposition must be explicitly exposed, otherwise the
			// browser hides it from JS on cross-origin responses (like report
			// exports), so the frontend can't read the backend-provided
			// filename and falls back to a generic one.
			ExposedHeaders: []string{"Content-Disposition"},
		}))
	}

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", handlers.Health)
		r.Get("/manifest", handlers.NewManifestHandler(manifestPath))
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

				r.Route("/import", func(r chi.Router) {
					r.With(middleware.RequireRole(models.RoleAdmin)).Post("/clockify", importHandler.ImportClockify)
				})
			})
		})
	})

	if metricsHandler != nil {
		r.Get("/metrics", metricsHandler.ServeHTTP)
	}

	// Generated after every /v1 route above is registered, so it's always in
	// sync with the router — never a hand-maintained spec file to go stale.
	spec := openapi.Generate(r, "CWClock API", apiVersion)
	r.Get("/openapi.json", handlers.NewOpenAPIHandler(spec))
	r.Get("/", handlers.ServeSwaggerUI)

	return r
}
