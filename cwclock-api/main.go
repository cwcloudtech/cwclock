package main

import (
	"context"
	"net/http"
	"runtime"

	"cwclock-api/internal/config"
	"cwclock-api/internal/db"
	"cwclock-api/internal/handlers"
	"cwclock-api/internal/metrics"
	"cwclock-api/internal/models"
	"cwclock-api/internal/router"
	"cwclock-api/internal/store"
	"cwclock-api/internal/telemetry"
)

func main() {
	cfg := config.Load()
	runtime.GOMAXPROCS(cfg.MaxWorkers)
	models.SetAllowedCurrencies(cfg.AllowedCurrencies)

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
	orgStore := store.NewOrgStore(pool)
	clientStore := store.NewClientStore(pool)
	projectStore := store.NewProjectStore(pool)
	timeEntryStore := store.NewTimeEntryStore(pool)

	userHandler := handlers.NewUserHandler(userStore, cfg.JWTSecret)
	orgHandler := handlers.NewOrganizationHandler(orgStore, userStore)
	clientHandler := handlers.NewClientHandler(clientStore)
	projectHandler := handlers.NewProjectHandler(projectStore)
	timeEntryHandler := handlers.NewTimeEntryHandler(timeEntryStore)
	adminHandler := handlers.NewAdminHandler(userStore)
	reportHandler := handlers.NewReportHandler(orgStore, clientStore, projectStore, timeEntryStore)

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
		userHandler, orgHandler, clientHandler, projectHandler, timeEntryHandler, reportHandler, adminHandler,
		orgStore, userStore, cfg.JWTSecret, cfg.CorsEnabled, cfg.CorsAllowedOrigins, cfg.Version, cfg.ManifestPath,
		tel, met.Observe, met.Handler,
	)

	tel.Logger.Info("server started", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		tel.Logger.Error("server stopped", "error", err)
		panic(err)
	}
}
