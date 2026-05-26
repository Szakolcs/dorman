package config

import (
	"os"
	"strings"
)

type Config struct {
	Port          string
	DatabaseURL   string
	SessionSecret string
	CookieSecure  bool
}

func Load() Config {
	return Config{
		Port:          envOrDefault("PORT", "8080"),
		DatabaseURL:   envOrDefault("DATABASE_URL", "postgres://dev:dev@localhost:5432/dormatory_manager?sslmode=disable"),
		SessionSecret: envOrDefault("SESSION_SECRET", "dev-session-secret"),
		CookieSecure:  envTruthy("COOKIE_SECURE"),
	}
}

func envTruthy(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
