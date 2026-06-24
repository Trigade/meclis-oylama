package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	BridgeSecret  string
	SessionSecret string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/meclis?sslmode=disable"),
		BridgeSecret:  getEnv("BRIDGE_SECRET", "dev-secret-change-me"),
		SessionSecret: getEnv("SESSION_SECRET", "dev-session-secret-change-me"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
