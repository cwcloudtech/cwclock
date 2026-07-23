package main

import (
	"context"
	"net/http"
	"net/url"
	"runtime"

	"github.com/go-webauthn/webauthn/webauthn"

	"cwclock-api/internal/config"
	"cwclock-api/internal/contact"
	"cwclock-api/internal/db"
	"cwclock-api/internal/email"
	"cwclock-api/internal/handlers"
	"cwclock-api/internal/metrics"
	"cwclock-api/internal/oidc"
	"cwclock-api/internal/router"
	"cwclock-api/internal/scheduler"
	"cwclock-api/internal/store"
	"cwclock-api/internal/telemetry"
	"cwclock-api/internal/utils"
)

// mfaIssuer is the "issuer" shown by authenticator apps (Google
// Authenticator, etc.) next to an enrolled TOTP entry.
const mfaIssuer = "CWClock"

func main() {
	cfg := config.Load()
	runtime.GOMAXPROCS(cfg.MaxWorkers)

	ctx := context.Background()

	tel, err := telemetry.Setup(ctx, telemetry.Config{
		Endpoint: cfg.OtelEndpoint,
		Proto:    cfg.OtelProto,
		Version:  cfg.Version,
	})
	if err != nil {
		panic(err)
	}
	defer func() { _ = tel.Shutdown(context.Background()) }()

	pool, err := db.Connect(ctx, cfg.DatabaseURL, cfg.PostgresPoolSize)
	if err != nil {
		tel.Logger.Error("failed to connect to database", "error", err)
		panic(err)
	}
	defer pool.Close()

	userStore := store.NewUserStore(pool)
	countryStore := store.NewCountryStore(pool)
	currencyStore := store.NewCurrencyStore(pool)
	fieldStore := store.NewFieldStore(pool)
	orgStore := store.NewOrgStore(pool, countryStore)
	clientStore := store.NewClientStore(pool)
	projectStore := store.NewProjectStore(pool)
	timeEntryStore := store.NewTimeEntryStore(pool)
	apiKeyStore := store.NewApiKeyStore(pool)
	invoiceStore := store.NewInvoiceStore(pool)
	webauthnCredStore := store.NewWebAuthnCredentialStore(pool)
	exportJobStore := store.NewExportJobStore(pool)
	mailCounterStore := store.NewMailCounterStore(pool)

	mailer := email.NewSender(cfg.CWCloudAPIURL, cfg.CWCloudAPIKey, cfg.EmailFrom, cfg.APIBaseURL)

	rpID := cfg.UIBaseURL
	if u, err := url.Parse(cfg.UIBaseURL); err == nil && utils.IsNotBlank(u.Hostname()) {
		rpID = u.Hostname()
	}

	waInstance, err := webauthn.New(&webauthn.Config{
		RPID:          rpID,
		RPDisplayName: mfaIssuer,
		RPOrigins:     []string{cfg.UIBaseURL},
	})

	if err != nil {
		tel.Logger.Error("failed to configure WebAuthn", "error", err)
		panic(err)
	}

	userHandler := handlers.NewUserHandler(userStore, webauthnCredStore, cfg.JWTSecret, cfg.MaxImageSize, cfg.ActivationMode, mailer, cfg.APIBaseURL, cfg.UIBaseURL, cfg.ConfirmationEmailTTL)
	orgHandler := handlers.NewOrganizationHandler(orgStore, userStore, countryStore, currencyStore, cfg.MaxImageSize)
	clientHandler := handlers.NewClientHandler(clientStore, orgStore, countryStore)
	projectHandler := handlers.NewProjectHandler(projectStore, clientStore)
	timeEntryHandler := handlers.NewTimeEntryHandler(timeEntryStore)
	adminHandler := handlers.NewAdminHandler(userStore, webauthnCredStore, cfg.MaxImageSize)
	mfaHandler := handlers.NewMFAHandler(userStore, webauthnCredStore, cfg.JWTSecret, cfg.ActivationMode, waInstance, mfaIssuer)
	importHandler := handlers.NewImportHandler(userStore, clientStore, projectStore, timeEntryStore)
	reportHandler := handlers.NewReportHandler(orgStore, clientStore, projectStore, timeEntryStore, userStore, cfg.MaxReportSize)
	apiKeyHandler := handlers.NewApiKeyHandler(apiKeyStore)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceStore, orgStore, clientStore, projectStore, timeEntryStore, userStore, cfg.MaxReportSize, mailer, reportHandler, mailCounterStore, cfg.MailLimit)
	currencyHandler := handlers.NewCurrencyHandler(currencyStore)
	countryHandler := handlers.NewCountryHandler(countryStore)
	fieldHandler := handlers.NewFieldHandler(fieldStore)
	oidcProviders := oidc.BuildProviders(cfg)
	oidcHandler := handlers.NewOIDCHandler(oidcProviders, userStore, webauthnCredStore, cfg.JWTSecret, cfg.APIBaseURL, cfg.UIBaseURL, cfg.OIDCKeycloakGroups, cfg.ActivationMode)
	contactHandler := handlers.NewContactHandler(contact.New(cfg.CWCloudAPIURL, cfg.CWCloudContactFormID))

	exportReportGenerator := handlers.NewExportReportGenerator(reportHandler, invoiceStore)
	exportDelivery := handlers.NewExportDeliveryService(mailer, orgStore, mailCounterStore, cfg.MailLimit)
	exportScheduler := scheduler.NewExportJobScheduler(exportJobStore, exportReportGenerator, exportDelivery)
	if err := exportScheduler.Start(ctx); err != nil {
		tel.Logger.Error("failed to start export job scheduler", "error", err)
		panic(err)
	}
	defer exportScheduler.Stop()
	exportJobHandler := handlers.NewExportJobHandler(exportJobStore, exportScheduler)

	met, err := metrics.Setup(ctx, metrics.Config{
		Endpoint: cfg.OtelEndpoint,
		Proto:    cfg.OtelProto,
		Version:  cfg.Version,
	}, orgStore, clientStore, projectStore, timeEntryStore)
	if err != nil {
		tel.Logger.Error("failed to set up metrics", "error", err)
		panic(err)
	}
	defer func() { _ = met.Shutdown(context.Background()) }()

	r := router.New(
		userHandler, orgHandler, clientHandler, projectHandler, timeEntryHandler, reportHandler, adminHandler, mfaHandler, importHandler, apiKeyHandler, invoiceHandler,
		currencyHandler, countryHandler, fieldHandler, oidcHandler, contactHandler, exportJobHandler,
		orgStore, userStore, apiKeyStore, cfg.JWTSecret, cfg.ActivationMode, cfg.CorsEnabled, cfg.CorsAllowedOrigins, cfg.Version, cfg.ManifestPath,
		tel, met.Observe, met.Handler,
	)

	tel.Logger.Info("server started", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		tel.Logger.Error("server stopped", "error", err)
		panic(err)
	}
}
