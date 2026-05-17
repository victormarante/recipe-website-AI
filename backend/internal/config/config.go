package config

import (
	"fmt"
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
	AuthUsername string
	AuthPassword string
	JWTSecret    string
	R2AccountID  string
	R2AccessKey  string
	R2SecretKey  string
	R2BucketName string
	R2PublicURL  string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error in production)
	_ = godotenv.Load()

	cfg := &Config{
		Port:         getEnv("PORT", "8080"),
		Environment:  getEnv("APP_ENV", "development"),
		DatabasePath: getEnv("DATABASE_PATH", "./recipes.db"),
		CORSOrigins:  strings.Split(getEnv("CORS_ORIGIN", "http://localhost:8080"), ","),
		AuthUsername: getEnv("AUTH_USERNAME", ""),
		AuthPassword: getEnv("AUTH_PASSWORD", ""),
		JWTSecret:    getEnv("JWT_SECRET", ""),
		R2AccountID:  getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKey:  getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretKey:  getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName: getEnv("R2_BUCKET_NAME", ""),
		R2PublicURL:  getEnv("R2_PUBLIC_URL", ""),
	}

	if cfg.AuthUsername == "" || cfg.AuthPassword == "" || cfg.JWTSecret == "" {
		return nil, fmt.Errorf("AUTH_USERNAME, AUTH_PASSWORD, and JWT_SECRET must be set")
	}

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
