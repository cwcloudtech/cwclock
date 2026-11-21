package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"runtime"

	"github.com/go-webauthn/webauthn/webauthn"

	"cwclock-api/internal/aidownload"
	"cwclock-api/internal/config"
	"cwclock-api/internal/contact"
	"cwclock-api/internal/db"
	"cwclock-api/internal/email"
	"cwclock-api/internal/enrich"
	"cwclock-api/internal/handlers"
	"cwclock-api/internal/metrics"
	"cwclock-api/internal/middleware"
	"cwclock-api/internal/notify"
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
	auditStore := store.NewAuditStore(pool)
	sessionStore := store.NewSessionStore(pool)
	aiStore := store.NewAIStore(pool)

	mailer := email.NewSender(cfg.CWCloudAPIURL, cfg.CWCloudAPIKey, cfg.EmailFrom, cfg.APIBaseURL)

	// Discord notifications (sign-in/out, bans, creations, cron, errors) + visitor
	// geo enrichment. Both are best-effort and disabled (no-ops) without their config.
	discord := notify.NewDiscord(cfg.DiscordToken)
	enricher := enrich.NewGeo(cfg.IPInfoToken).Enrich

	// A deduped "[BAN] :rotating_light:" alert when a banned IP/user keeps hitting
	// the API, and an "[ERROR] :fire:" alert on every 5xx. Nil (disabled) without a webhook.
	var banAlerter *middleware.BanAlerter
	var onServerError middleware.ErrorFunc
	if discord.Enabled() {
		banAlerter = middleware.NewBanAlerter(func(ctx context.Context, text string) {
			_ = discord.SendText(ctx, text)
		})
		onServerError = func(method, route string, status int) {
			go func() {
				_ = discord.SendText(context.Background(),
					fmt.Sprintf("[ERROR] :fire: %d on %s %s", status, method, route))
			}()
		}
	}

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

	userHandler := handlers.NewUserHandler(userStore, webauthnCredStore, auditStore, cfg.JWTSecret, cfg.MaxImageSize, cfg.ActivationMode, mailer, cfg.APIBaseURL, cfg.UIBaseURL, cfg.ConfirmationEmailTTL).
		WithConn(sessionStore, enricher, discord)
	orgHandler := handlers.NewOrganizationHandler(orgStore, userStore, countryStore, currencyStore, auditStore, cfg.MaxImageSize).WithNotifier(discord)
	clientHandler := handlers.NewClientHandler(clientStore, orgStore, countryStore, auditStore).WithNotifier(discord)
	projectHandler := handlers.NewProjectHandler(projectStore, clientStore, auditStore).WithNotifier(discord)
	timeEntryHandler := handlers.NewTimeEntryHandler(timeEntryStore, auditStore).WithNotifier(discord)
	adminHandler := handlers.NewAdminHandler(userStore, webauthnCredStore, auditStore, cfg.MaxImageSize)
	mfaHandler := handlers.NewMFAHandler(userStore, webauthnCredStore, auditStore, cfg.JWTSecret, cfg.ActivationMode, waInstance, mfaIssuer).
		WithConn(sessionStore, enricher, discord)
	importHandler := handlers.NewImportHandler(userStore, clientStore, projectStore, timeEntryStore, auditStore)
	reportHandler := handlers.NewReportHandler(orgStore, clientStore, projectStore, timeEntryStore, userStore, cfg.MaxReportSize)
	apiKeyHandler := handlers.NewApiKeyHandler(apiKeyStore, auditStore)
	configHandler := handlers.NewConfigHandler(orgStore, cfg.APIBaseURL)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceStore, orgStore, clientStore, projectStore, timeEntryStore, userStore, cfg.MaxReportSize, mailer, reportHandler, mailCounterStore, cfg.MailLimit).WithNotifier(discord)
	currencyHandler := handlers.NewCurrencyHandler(currencyStore)
	countryHandler := handlers.NewCountryHandler(countryStore)
	fieldHandler := handlers.NewFieldHandler(fieldStore)
	oidcProviders := oidc.BuildProviders(cfg)
	oidcHandler := handlers.NewOIDCHandler(oidcProviders, userStore, webauthnCredStore, auditStore, cfg.JWTSecret, cfg.APIBaseURL, cfg.UIBaseURL, cfg.OIDCKeycloakGroups, cfg.ActivationMode).
		WithConn(sessionStore, enricher, discord)
	contactHandler := handlers.NewContactHandler(contact.New(cfg.CWCloudAPIURL, cfg.CWCloudContactFormID))

	exportReportGenerator := handlers.NewExportReportGenerator(reportHandler, invoiceStore)
	exportDelivery := handlers.NewExportDeliveryService(mailer, orgStore, mailCounterStore, cfg.MailLimit)
	exportScheduler := scheduler.NewExportJobScheduler(exportJobStore, exportReportGenerator, exportDelivery).WithNotifier(discord)
	if err := exportScheduler.Start(ctx); err != nil {
		tel.Logger.Error("failed to start export job scheduler", "error", err)
		panic(err)
	}
	defer exportScheduler.Stop()
	exportJobHandler := handlers.NewExportJobHandler(exportJobStore, exportScheduler, auditStore)
	calendarFeedHandler := handlers.NewCalendarFeedHandler(userStore, timeEntryStore, projectStore, clientStore, auditStore, cfg.APIBaseURL)
	ipBanHandler := handlers.NewIPBanHandler(sessionStore).WithNotifier(discord)

	// The AI assistant (ai-instruct-132). The cwclock_api tool calls this very API
	// over the loopback (never through the corporate proxy or the public URL), as the
	// signed-in user, so its own Auth/OrgMembership/RequireRole gates decide what the
	// model may read or generate. A report/invoice PDF it generates is parked in the
	// download store (served from the public API URL, swept on a TTL) so the agent
	// can hand the user a clickable link.
	aiDownloadStore := aidownload.New(cfg.AIDownloadDir, cfg.AIDownloadTTL, cfg.AIDownloadSweep, cfg.AIDownloadMaxBytes, cfg.APIBaseURL, tel.Logger)
	go aiDownloadStore.StartSweeper(ctx)
	aiHandler := handlers.NewAIHandler(aiStore, userStore, orgStore, fmt.Sprintf("http://127.0.0.1:%s", cfg.Port), aiDownloadStore, auditStore)
	aiDownloadHandler := handlers.NewAIDownloadHandler(aiDownloadStore)

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

	// Audit events are counted on the same meter as everything else, so they
	// land on both /metrics and the OTLP endpoint. Wired here rather than
	// through the constructor because metrics.Setup needs the stores to build
	// its gauges - the store necessarily exists first - and this runs before
	// the server starts serving, so nothing can record an entry in between.
	auditStore.SetObserver(met.RecordAudit)

	r := router.New(
		userHandler, orgHandler, clientHandler, projectHandler, timeEntryHandler, reportHandler, adminHandler, mfaHandler, importHandler, apiKeyHandler, configHandler, invoiceHandler,
		currencyHandler, countryHandler, fieldHandler, oidcHandler, contactHandler, exportJobHandler, calendarFeedHandler, aiHandler, aiDownloadHandler,
		orgStore, userStore, apiKeyStore, cfg.JWTSecret, cfg.ActivationMode, cfg.CorsEnabled, cfg.CorsAllowedOrigins, cfg.Version, cfg.ManifestPath,
		tel, met.Observe, met.Handler,
		sessionStore, ipBanHandler, enricher, banAlerter, onServerError,
	)

	tel.Logger.Info("server started", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		tel.Logger.Error("server stopped", "error", err)
		panic(err)
	}
}
