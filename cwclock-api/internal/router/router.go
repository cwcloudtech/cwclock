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
	mfaHandler *handlers.MFAHandler,
	importHandler *handlers.ImportHandler,
	apiKeyHandler *handlers.ApiKeyHandler,
	invoiceHandler *handlers.InvoiceHandler,
	currencyHandler *handlers.CurrencyHandler,
	countryHandler *handlers.CountryHandler,
	fieldHandler *handlers.FieldHandler,
	oidcHandler *handlers.OIDCHandler,
	contactHandler *handlers.ContactHandler,
	orgs *store.OrgStore,
	users *store.UserStore,
	apiKeys middleware.ApiKeyVerifier,
	jwtSecret string,
	activationMode string,
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
			AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Authorization", "Content-Type", "X-Api-Key"},
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
		r.Get("/currencies", currencyHandler.List)
		r.Get("/countries", countryHandler.List)
		r.Get("/fields", fieldHandler.List)

		// Public contact form (no auth) - forwards to CWCloud's /v1/contactreq.
		r.Post("/contact", contactHandler.Create)

		// Public logo endpoints (no auth) - emails reference these by URL
		// instead of embedding an image as a data: URI, which mail clients and
		// email-sending APIs commonly strip from <img src> outright.
		r.Get("/assets/logo.png", handlers.AssetsLogo)
		r.Get("/organizations/{orgId}/logo", orgHandler.PublicLogo)

		r.Route("/oidc", func(r chi.Router) {
			r.Get("/", oidcHandler.ListProviders)
			r.Get("/callback", oidcHandler.FrontendCallback)
			r.Get("/{provider}/login", oidcHandler.Login)
			r.Get("/{provider}/callback", oidcHandler.Callback)
		})

		r.Route("/user", func(r chi.Router) {
			r.Get("/confirmation", userHandler.Confirm)
		})

		r.Route("/users", func(r chi.Router) {
			r.Post("/", userHandler.Register)
			r.Post("/login", userHandler.Login)
			r.Post("/forgot-password", userHandler.ForgotPassword)
			r.Post("/reset-password", userHandler.ResetPassword)

			// Finishing a password login gated by MFA (see
			// UserHandler.Login/models.MFAChallengeResponse): these carry
			// their own short-lived challenge/ceremony token instead of a
			// session, so they sit outside the Auth() middleware group below.
			r.Route("/login/mfa", func(r chi.Router) {
				r.Post("/totp", mfaHandler.LoginTOTP)
				r.Post("/webauthn/begin", mfaHandler.LoginWebAuthnBegin)
				r.Post("/webauthn/finish", mfaHandler.LoginWebAuthnFinish)
			})

			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(jwtSecret, apiKeys))
				r.Get("/me", userHandler.Me)

				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireActiveUser(users, activationMode))
					r.Put("/me", userHandler.UpdateProfile)
					r.Put("/me/picture", userHandler.UpdatePicture)
					r.Get("/search", userHandler.Search)

					r.Route("/me/api-keys", func(r chi.Router) {
						r.Get("/", apiKeyHandler.List)
						r.Post("/", apiKeyHandler.Create)
						r.Delete("/{id}", apiKeyHandler.Delete)
					})

					r.Route("/me/mfa", func(r chi.Router) {
						r.Get("/", mfaHandler.Status)
						r.Post("/totp/setup", mfaHandler.TOTPSetup)
						r.Post("/totp/confirm", mfaHandler.TOTPConfirm)
						r.Delete("/totp", mfaHandler.TOTPDisable)
						r.Post("/webauthn/register/begin", mfaHandler.WebAuthnRegisterBegin)
						r.Post("/webauthn/register/finish", mfaHandler.WebAuthnRegisterFinish)
						r.Delete("/webauthn/{credentialId}", mfaHandler.WebAuthnDelete)
					})
				})
			})
		})

		r.Route("/admin/users", func(r chi.Router) {
			r.Use(middleware.Auth(jwtSecret, apiKeys))
			r.Use(middleware.RequireActiveUser(users, activationMode))
			r.Use(middleware.RequireSuperuser(users))
			r.Get("/", adminHandler.ListUsers)
			r.Put("/{id}", adminHandler.UpdateUser)
			r.Delete("/{id}", adminHandler.DeleteUser)
			r.Post("/{id}/disable-mfa", adminHandler.DisableMFA)
		})

		r.Route("/admin/organizations", func(r chi.Router) {
			r.Use(middleware.Auth(jwtSecret, apiKeys))
			r.Use(middleware.RequireActiveUser(users, activationMode))
			r.Use(middleware.RequireSuperuser(users))
			r.Get("/", orgHandler.AdminList)
		})

		r.Route("/organizations", func(r chi.Router) {
			r.Use(middleware.Auth(jwtSecret, apiKeys))
			r.Use(middleware.RequireActiveUser(users, activationMode))
			r.Post("/", orgHandler.Create)
			r.Get("/", orgHandler.List)

			r.Route("/{orgId}", func(r chi.Router) {
				r.Use(middleware.OrgMembership(orgs, users))

				r.Get("/", orgHandler.Get)
				r.With(middleware.RequireRole(models.RoleOwner)).Patch("/", orgHandler.Update)
				r.With(middleware.RequireRole(models.RoleOwner)).Delete("/", orgHandler.Delete)
				r.With(middleware.RequireRole(models.RoleOwner)).Put("/owner", orgHandler.TransferOwnership)
				r.With(middleware.RequireRole(models.RoleOwner)).Patch("/external-connections", orgHandler.AddExternalConnection)
				r.With(middleware.RequireRole(models.RoleOwner)).Patch("/external-connections/{connectionId}", orgHandler.RemoveExternalConnection)

				r.Route("/members", func(r chi.Router) {
					r.Get("/", orgHandler.ListMembers)
					r.With(middleware.RequireRole(models.RoleOwner)).Post("/", orgHandler.AddMember)
					r.With(middleware.RequireRole(models.RoleOwner)).Put("/{userId}", orgHandler.UpdateMember)
					r.With(middleware.RequireRole(models.RoleAdmin)).Delete("/{userId}", orgHandler.RemoveMember)
					r.With(middleware.RequireRole(models.RoleAdmin)).Put("/{userId}/rate", orgHandler.SetMemberRate)
				})

				r.Route("/clients", func(r chi.Router) {
					r.Get("/", clientHandler.List)
					r.With(middleware.RequireRole(models.RoleAdmin)).Post("/", clientHandler.Create)
					r.With(middleware.RequireRole(models.RoleAdmin)).Put("/{clientId}", clientHandler.Update)
					r.With(middleware.RequireRole(models.RoleAdmin)).Delete("/{clientId}", clientHandler.Delete)
					r.With(middleware.RequireRole(models.RoleOwner)).Put("/{clientId}/transfer", clientHandler.Transfer)

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
					r.Post("/detailed", reportHandler.Detailed)
					r.Post("/summary", reportHandler.Summary)
				})

				r.Route("/import", func(r chi.Router) {
					r.With(middleware.RequireRole(models.RoleAdmin)).Post("/csv", importHandler.ImportCSV)
				})

				// Invoices are owner/admin-only end to end: generating one
				// exposes billing amounts the same way the report PDF/CSV
				// exports already do for that role pair.
				r.Route("/invoices", func(r chi.Router) {
					r.Use(middleware.RequireRole(models.RoleAdmin))
					r.Get("/", invoiceHandler.List)
					r.Post("/preview", invoiceHandler.Preview)
					r.Post("/", invoiceHandler.Generate)
					r.Get("/{invoiceId}/pdf", invoiceHandler.DownloadPDF)
					r.Post("/{invoiceId}/reupload", invoiceHandler.Reupload)
					r.Post("/{invoiceId}/send", invoiceHandler.SendEmail)
					r.Put("/{invoiceId}", invoiceHandler.UpdateStatus)
					r.Delete("/{invoiceId}", invoiceHandler.Delete)
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
