package config

import "os"

type Config struct {
	Port          string
	DatabaseURL   string
	SessionSecret string
}

func Load() Config {
	return Config{
		Port:        envOrDefault("PORT", "8080"),
		DatabaseURL: envOrDefault("DATABASE_URL", "postgres://dev:dev@localhost:5432/dormatory_manager?sslmode=disable"),
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
