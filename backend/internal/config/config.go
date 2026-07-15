package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Port             string
	Environment      string
	DatabasePath     string
	CORSOrigins      []string
	AuthUsername     string
	AuthPassword     string
	AuthPasswordHash string
	JWTSecret        string
	R2AccountID      string
	R2AccessKey      string
	R2SecretKey      string
	R2BucketName     string
	R2PublicURL      string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error in production)
	_ = godotenv.Load()

	cfg := &Config{
		Port:             getEnv("PORT", "8080"),
		Environment:      getEnv("APP_ENV", "development"),
		DatabasePath:     getEnv("DATABASE_PATH", "./recipes.db"),
		CORSOrigins:      strings.Split(getEnv("CORS_ORIGIN", "http://localhost:8080"), ","),
		AuthUsername:     getEnv("AUTH_USERNAME", ""),
		AuthPassword:     getEnv("AUTH_PASSWORD", ""),
		AuthPasswordHash: getEnv("AUTH_PASSWORD_HASH", ""),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		R2AccountID:      getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKey:      getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretKey:      getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:     getEnv("R2_BUCKET_NAME", ""),
		R2PublicURL:      getEnv("R2_PUBLIC_URL", ""),
	}

	if cfg.AuthUsername == "" || (cfg.AuthPassword == "" && cfg.AuthPasswordHash == "") || cfg.JWTSecret == "" {
		return nil, fmt.Errorf("AUTH_USERNAME, AUTH_PASSWORD or AUTH_PASSWORD_HASH, and JWT_SECRET must be set")
	}

	r2Values := map[string]string{
		"R2_ACCOUNT_ID":        cfg.R2AccountID,
		"R2_ACCESS_KEY_ID":     cfg.R2AccessKey,
		"R2_SECRET_ACCESS_KEY": cfg.R2SecretKey,
		"R2_BUCKET_NAME":       cfg.R2BucketName,
		"R2_PUBLIC_URL":        cfg.R2PublicURL,
	}
	hasR2Value := false
	for _, value := range r2Values {
		if value != "" {
			hasR2Value = true
			break
		}
	}
	if hasR2Value {
		var missing []string
		for key, value := range r2Values {
			if value == "" {
				missing = append(missing, key)
			}
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf("incomplete R2 configuration, missing: %s", strings.Join(missing, ", "))
		}
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
