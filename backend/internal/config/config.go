package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Port         string
	Environment  string
	DatabasePath string
	CORSOrigins  []string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error in production)
	_ = godotenv.Load()

	return &Config{
		Port:         getEnv("PORT", "8080"),
		Environment:  getEnv("APP_ENV", "development"),
		DatabasePath: getEnv("DATABASE_PATH", "./recipes.db"),
		CORSOrigins:  strings.Split(getEnv("CORS_ORIGIN", "http://localhost:8080"), ","),
	}, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
