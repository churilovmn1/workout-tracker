package config

import (
	"fmt"
	"os"
)

// Config holds all application configuration.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	BotToken    string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	cfg := &Config{
		Port:        getEnvOrDefault("PORT", "8080"),
		DatabaseURL: dbURL,
		JWTSecret:   getEnvOrDefault("JWT_SECRET", "default-secret-change-me"),
		BotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
	}

	return cfg, nil
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
