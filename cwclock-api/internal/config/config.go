package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

type Config struct {
	Port                     string
	DatabaseURL              string
	JWTSecret                string
	MaxWorkers               int
	PostgresPoolSize         int
	CorsEnabled              bool
	CorsAllowedOrigins       []string
	Version                  string
	ManifestPath             string
	OtelEndpoint             string
	OtelProto                string
	MaxImageSize             int64
	MaxReportSize            int
	APIBaseURL               string
	UIBaseURL                string
	OIDCGoogleClientID       string
	OIDCGoogleClientSecret   string
	OIDCGithubClientID       string
	OIDCGithubClientSecret   string
	OIDCKeycloakBaseURL      string
	OIDCKeycloakClientID     string
	OIDCKeycloakClientSecret string
	OIDCKeycloakGroups       []string
	CWCloudAPIURL            string
	CWCloudAPIKey            string
	CWCloudContactFormID     string
	EmailFrom                string
	ConfirmationEmailTTL     time.Duration
	ActivationMode           string
}

// defaultMaxImageSize is applied when CWCLOCK_MAX_IMAGE_SIZE is unset or
// isn't a parsable number of bytes.
const defaultMaxImageSize int64 = 2 * 1024 * 1024

// defaultMaxReportSize is applied when CWCLOCK_MAX_REPORT_SIZE is unset or
// isn't a parsable number of entries; it caps how many time entries a single
// report/export or invoice generation may cover.
const defaultMaxReportSize int = 5000

// defaultConfirmationEmailExpirationHours is applied when
// CWCLOCK_CONFIRMATION_EMAIL_EXPIRATION is unset or isn't a parsable number
// of hours; it bounds how long an account-confirmation or password-reset
// link emailed to a user stays usable.
const defaultConfirmationEmailExpirationHours int = 24

func Load() Config {
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := utils.GetEnv("POSTGRES_PORT", "5432")
	db := os.Getenv("POSTGRES_DB")
	sslmode := utils.GetEnv("POSTGRES_SSLMODE", "disable")
	allowedOrigins := utils.GetEnv("CWCLOCK_CORS_ALLOWED_ORIGINS", "*")

	maxWorkers, err := strconv.Atoi(utils.GetEnv("MAX_WORKERS", "5"))
	if err != nil || maxWorkers <= 0 {
		maxWorkers = 5
	}

	postgresPoolSize, err := strconv.Atoi(utils.GetEnv("POSTGRES_POOL_SIZE", "5"))
	if err != nil || postgresPoolSize <= 0 {
		postgresPoolSize = 5
	}

	maxImageSize, err := strconv.ParseInt(os.Getenv("CWCLOCK_MAX_IMAGE_SIZE"), 10, 64)
	if err != nil || maxImageSize <= 0 {
		maxImageSize = defaultMaxImageSize
	}

	maxReportSize, err := strconv.Atoi(os.Getenv("CWCLOCK_MAX_REPORT_SIZE"))
	if err != nil || maxReportSize <= 0 {
		maxReportSize = defaultMaxReportSize
	}

	activationMode := utils.GetEnv("CWCLOCK_ACTIVATION_MODE", models.ActivationModeAdmin)
	if !models.IsValidActivationMode(activationMode) {
		activationMode = models.ActivationModeAdmin
	}

	confirmationEmailExpirationHours, err := strconv.Atoi(os.Getenv("CWCLOCK_CONFIRMATION_EMAIL_EXPIRATION"))
	if err != nil || confirmationEmailExpirationHours <= 0 {
		confirmationEmailExpirationHours = defaultConfirmationEmailExpirationHours
	}

	return Config{
		Port:                     utils.GetEnv("PORT", "8080"),
		DatabaseURL:              fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, db, sslmode),
		JWTSecret:                os.Getenv("JWT_SECRET"),
		MaxWorkers:               maxWorkers,
		PostgresPoolSize:         postgresPoolSize,
		CorsEnabled:              utils.IsTrue(os.Getenv("CWCLOCK_CORS_ENABLED")),
		CorsAllowedOrigins:       strings.Split(allowedOrigins, ","),
		Version:                  versionFromManifest(utils.GetEnv("CWCLOCK_MANIFEST_PATH", "manifest.json"), utils.GetEnv("APP_VERSION", "1.0.0")),
		ManifestPath:             utils.GetEnv("CWCLOCK_MANIFEST_PATH", "manifest.json"),
		OtelEndpoint:             os.Getenv("CWCLOCK_OTEL_ENDPOINT"),
		OtelProto:                utils.GetEnv("CWCLOCK_OTEL_PROTO", "otlp/grpc"),
		MaxImageSize:             maxImageSize,
		MaxReportSize:            maxReportSize,
		APIBaseURL:               utils.GetBaseUrlFromEnvWithFallback("CWCLOCK_API_URL", "http://localhost:8080"),
		UIBaseURL:                utils.GetBaseUrlFromEnvWithFallback("CWCLOCK_UI_URL", "http://localhost:3000"),
		OIDCGoogleClientID:       os.Getenv("CWCLOCK_OIDC_GOOGLE_CLIENT_ID"),
		OIDCGoogleClientSecret:   os.Getenv("CWCLOCK_OIDC_GOOGLE_CLIENT_SECRET"),
		OIDCGithubClientID:       os.Getenv("CWCLOCK_OIDC_GITHUB_CLIENT_ID"),
		OIDCGithubClientSecret:   os.Getenv("CWCLOCK_OIDC_GITHUB_CLIENT_SECRET"),
		OIDCKeycloakBaseURL:      utils.GetBaseUrlFromEnv("CWCLOCK_OIDC_KEYCLOAK_BASE_URL"),
		OIDCKeycloakClientID:     os.Getenv("CWCLOCK_OIDC_KEYCLOAK_CLIENT_ID"),
		OIDCKeycloakClientSecret: os.Getenv("CWCLOCK_OIDC_KEYCLOAK_CLIENT_SECRET"),
		OIDCKeycloakGroups:       utils.SplitList(os.Getenv("CWCLOCK_OIDC_KEYCLOAK_GROUPS")),
		CWCloudAPIURL:            utils.GetBaseUrlFromEnvWithFallback("CWCLOUD_API_URL", "https://api.cwcloud.tech"),
		CWCloudAPIKey:            os.Getenv("CWCLOUD_API_KEY"),
		CWCloudContactFormID:     os.Getenv("CWCLOUD_CONTACT_FORM_ID"),
		EmailFrom:                utils.GetEnv("CWCLOCK_MAIL_FROM", "noreply@cwcloud.tech"),
		ConfirmationEmailTTL:     time.Duration(confirmationEmailExpirationHours) * time.Hour,
		ActivationMode:           activationMode,
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
