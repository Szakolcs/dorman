package config

import "os"

type Config struct {
	Port          string
	DatabaseURL   string
	SessionSecret string
}

func Load() Config {
	return Config{
		Port:          envOrDefault("PORT", "8080"),
		DatabaseURL:   envOrDefault("DATABASE_URL", "postgres://dev:dev@localhost:5432/dormatory_manager?sslmode=disable"),
		SessionSecret: envOrDefault("SESSION_SECRET", "dev-session-secret-change-in-production"),
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
