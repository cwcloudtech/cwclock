package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"cwclock-api/internal/utils"
)

type Config struct {
	Port              string
	DatabaseURL       string
	JWTSecret         string
	MaxWorkers        int
	PostgresPoolSize  int
	AllowedCurrencies []string
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
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, db, sslmode),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		MaxWorkers:        maxWorkers,
		PostgresPoolSize:  postgresPoolSize,
		AllowedCurrencies: parseAllowedCurrencies(os.Getenv("CWCLOCK_ALLOWED_CURRENCIES")),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	return utils.If(utils.IsNotBlank(v), v, fallback)
}

// parseAllowedCurrencies decodes CWCLOCK_ALLOWED_CURRENCIES, a JSON array of
// ISO 4217 codes like ["EUR","USD","GBP"]. An empty or invalid value yields
// nil, which callers treat as "keep the built-in default list".
func parseAllowedCurrencies(raw string) []string {
	if utils.IsBlank(raw) {
		return nil
	}
	var codes []string
	if err := json.Unmarshal([]byte(raw), &codes); err != nil {
		log.Printf("invalid CWCLOCK_ALLOWED_CURRENCIES, ignoring: %v", err)
		return nil
	}
	return codes
}
