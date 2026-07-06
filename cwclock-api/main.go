package main

import (
	"context"
	"log"
	"net/http"
	"runtime"

	"cwclock-api/internal/config"
	"cwclock-api/internal/db"
	"cwclock-api/internal/handlers"
	"cwclock-api/internal/models"
	"cwclock-api/internal/router"
	"cwclock-api/internal/store"
)

func main() {
	cfg := config.Load()
	runtime.GOMAXPROCS(cfg.MaxWorkers)
	models.SetAllowedCurrencies(cfg.AllowedCurrencies)

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL, cfg.PostgresPoolSize)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
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

	r := router.New(
		userHandler, orgHandler, clientHandler, projectHandler, timeEntryHandler, reportHandler, adminHandler,
		orgStore, userStore, cfg.JWTSecret, cfg.CorsEnabled, cfg.CorsAllowedOrigins, cfg.Version,
	)

	log.Printf("Server started on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
