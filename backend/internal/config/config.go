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
	R2AccountID  string
	R2AccessKey  string
	R2SecretKey  string
	R2BucketName string
	R2PublicURL  string
	AdminPIN     string
	AuthSecret   string
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
		R2AccountID:  getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKey:  getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretKey:  getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName: getEnv("R2_BUCKET_NAME", ""),
		R2PublicURL:  getEnv("R2_PUBLIC_URL", ""),
		AdminPIN:     os.Getenv("ADMIN_PIN"),
		AuthSecret:   os.Getenv("AUTH_SECRET"),
	}

	if err := cfg.loadAuth(); err != nil {
		return nil, err
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

const (
	minPINLength = 6

	// Development-only fallbacks so a local setup works without extra env vars.
	devAdminPIN   = "123456"
	devAuthSecret = "dev-only-insecure-secret"
)

// loadAuth validates the admin PIN and token secret. Both are required in
// production; development falls back to insecure defaults.
func (c *Config) loadAuth() error {
	if c.Environment != "production" {
		if c.AdminPIN == "" {
			c.AdminPIN = devAdminPIN
		}
		if c.AuthSecret == "" {
			c.AuthSecret = devAuthSecret
		}
	}
	if c.AdminPIN == "" || c.AuthSecret == "" {
		return fmt.Errorf("ADMIN_PIN and AUTH_SECRET are required when APP_ENV=production")
	}
	if len(c.AdminPIN) < minPINLength {
		return fmt.Errorf("ADMIN_PIN must be at least %d digits", minPINLength)
	}
	for _, r := range c.AdminPIN {
		if r < '0' || r > '9' {
			return fmt.Errorf("ADMIN_PIN must contain digits only")
		}
	}
	return nil
}
