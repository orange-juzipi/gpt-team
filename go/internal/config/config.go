package config

import (
	"os"
	"strings"
	"time"

	"gpt-team-api/internal/apperr"
)

const defaultMeiguodizhiBaseURL = "https://www.meiguodizhi.com/api/v1/dz"

type Config struct {
	Port                 string
	PostgresDSN          string
	DatabaseName         string
	EfuncardBaseURL      string
	EfuncardAPIKey       string
	MeiguodizhiBaseURL   string
	CloudmailBaseURL     string
	CloudmailAPIToken    string
	DuckmailBaseURL      string
	DuckmailHTTPTimeout  time.Duration
	AccountEncryptionKey string
	HTTPTimeout          time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Port:                 valueOrDefault("PORT", "8080"),
		PostgresDSN:          os.Getenv("POSTGRES_DSN"),
		DatabaseName:         valueOrDefault("POSTGRES_DB_NAME", "gpt_team"),
		EfuncardBaseURL:      valueOrDefault("EFUNCARD_BASE_URL", "https://card.efuncard.com"),
		EfuncardAPIKey:       os.Getenv("EFUNCARD_API_KEY"),
		MeiguodizhiBaseURL:   defaultMeiguodizhiBaseURL,
		CloudmailBaseURL:     valueOrDefault("CLOUDMAIL_BASE_URL", "https://puax.cloud"),
		CloudmailAPIToken:    firstNonEmptyEnv("CLOUDMAIL_API_TOKEN", "CLOUDMAIL_AUTHORIZATION"),
		DuckmailBaseURL:      valueOrDefault("DUCKMAIL_BASE_URL", "https://api.duckmail.sbs"),
		DuckmailHTTPTimeout:  durationFromEnv("DUCKMAIL_HTTP_TIMEOUT", 90*time.Second),
		AccountEncryptionKey: os.Getenv("ACCOUNT_ENCRYPTION_KEY"),
		HTTPTimeout:          15 * time.Second,
	}

	if cfg.PostgresDSN == "" {
		return Config{}, apperr.BadRequest("missing_postgres_dsn", "POSTGRES_DSN is required")
	}

	if cfg.AccountEncryptionKey == "" {
		return Config{}, apperr.BadRequest("missing_encryption_key", "ACCOUNT_ENCRYPTION_KEY is required")
	}

	return cfg, nil
}

func valueOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}

	return ""
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return duration
}
