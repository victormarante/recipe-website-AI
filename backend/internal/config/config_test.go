package config

import "testing"

func TestLoadRejectsPartialR2Config(t *testing.T) {
	t.Setenv("R2_ACCOUNT_ID", "account")
	t.Setenv("R2_ACCESS_KEY_ID", "")
	t.Setenv("R2_SECRET_ACCESS_KEY", "")
	t.Setenv("R2_BUCKET_NAME", "")
	t.Setenv("R2_PUBLIC_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected partial R2 config error")
	}
}

func clearAuthEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"R2_ACCOUNT_ID", "R2_ACCESS_KEY_ID", "R2_SECRET_ACCESS_KEY", "R2_BUCKET_NAME", "R2_PUBLIC_URL"} {
		t.Setenv(k, "")
	}
	t.Setenv("ADMIN_PIN", "")
	t.Setenv("AUTH_SECRET", "")
}

func TestLoadAuthDevDefaults(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv("APP_ENV", "development")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AdminPIN == "" || cfg.AuthSecret == "" {
		t.Fatal("expected dev defaults for PIN and secret")
	}
}

func TestLoadAuthRequiredInProduction(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv("APP_ENV", "production")

	if _, err := Load(); err == nil {
		t.Fatal("expected error when PIN and secret are missing in production")
	}
}

func TestLoadRejectsWeakPIN(t *testing.T) {
	for _, pin := range []string{"12345", "12345a", "abcdef"} {
		clearAuthEnv(t)
		t.Setenv("APP_ENV", "production")
		t.Setenv("AUTH_SECRET", "secret")
		t.Setenv("ADMIN_PIN", pin)
		if _, err := Load(); err == nil {
			t.Fatalf("expected error for PIN %q", pin)
		}
	}
}

func TestLoadAcceptsValidPIN(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_SECRET", "secret")
	t.Setenv("ADMIN_PIN", "246810")
	if _, err := Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
}
