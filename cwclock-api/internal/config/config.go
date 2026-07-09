package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"cwclock-api/internal/utils"
)

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	MaxWorkers         int
	PostgresPoolSize   int
	AllowedCurrencies  []string
	CorsEnabled        bool
	CorsAllowedOrigins []string
	Version            string
	ManifestPath       string
	OtelEndpoint       string
	OtelProto          string
	MaxImageSize       int64
}

// defaultMaxImageSize is applied when CWCLOCK_MAX_IMAGE_SIZE is unset or
// isn't a parsable number of bytes.
const defaultMaxImageSize int64 = 2 * 1024 * 1024

func Load() Config {
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := getEnv("POSTGRES_PORT", "5432")
	db := os.Getenv("POSTGRES_DB")
	sslmode := getEnv("POSTGRES_SSLMODE", "disable")
	allowedOrigins := getEnv("CWCLOCK_CORS_ALLOWED_ORIGINS", "*")

	maxWorkers, err := strconv.Atoi(getEnv("MAX_WORKERS", "5"))
	if err != nil {
		maxWorkers = 5
	}

	postgresPoolSize, err := strconv.Atoi(getEnv("POSTGRES_POOL_SIZE", "5"))
	if err != nil {
		postgresPoolSize = 5
	}

	maxImageSize, err := strconv.ParseInt(os.Getenv("CWCLOCK_MAX_IMAGE_SIZE"), 10, 64)
	if err != nil || maxImageSize <= 0 {
		maxImageSize = defaultMaxImageSize
	}

	return Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, db, sslmode),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		MaxWorkers:         maxWorkers,
		PostgresPoolSize:   postgresPoolSize,
		AllowedCurrencies:  parseAllowedCurrencies(os.Getenv("CWCLOCK_ALLOWED_CURRENCIES")),
		CorsEnabled:        utils.IsTrue(os.Getenv("CWCLOCK_CORS_ENABLED")),
		CorsAllowedOrigins: strings.Split(allowedOrigins, ","),
		Version:            versionFromManifest(getEnv("CWCLOCK_MANIFEST_PATH", "manifest.json"), getEnv("APP_VERSION", "1.0.0")),
		ManifestPath:       getEnv("CWCLOCK_MANIFEST_PATH", "manifest.json"),
		OtelEndpoint:       os.Getenv("CWCLOCK_OTEL_ENDPOINT"),
		OtelProto:          getEnv("CWCLOCK_OTEL_PROTO", "otlp/grpc"),
		MaxImageSize:       maxImageSize,
	}
}

// versionFromManifest reads the version field from the manifest JSON file at
// path. Falls back to fallback if the file cannot be read or parsed.
func versionFromManifest(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	var m struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &m); err != nil || utils.IsBlank(m.Version) {
		return fallback
	}
	return m.Version
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
		slog.Warn("invalid CWCLOCK_ALLOWED_CURRENCIES, ignoring", "error", err)
		return nil
	}
	return codes
}
