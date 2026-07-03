package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port             string
	DatabaseURL      string
	JWTSecret        string
	MaxWorkers       int
	PostgresPoolSize int
}

func Load() Config {
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := getEnv("POSTGRES_PORT", "5432")
	db := os.Getenv("POSTGRES_DB")
	sslmode := getEnv("POSTGRES_SSLMODE", "disable")

	maxWorkers, err := strconv.Atoi(getEnv("MAX_WORKERS", "5"))
	if err != nil {
		maxWorkers = 5
	}

	postgresPoolSize, err := strconv.Atoi(getEnv("POSTGRES_POOL_SIZE", "5"))
	if err != nil {
		postgresPoolSize = 5
	}

	return Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, db, sslmode),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		MaxWorkers:       maxWorkers,
		PostgresPoolSize: postgresPoolSize,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
