package config

import "testing"

func TestLoadRequiresAuthAndJWT(t *testing.T) {
	t.Setenv("AUTH_USERNAME", "")
	t.Setenv("AUTH_PASSWORD", "")
	t.Setenv("AUTH_PASSWORD_HASH", "")
	t.Setenv("JWT_SECRET", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing auth config error")
	}
}

func TestLoadAcceptsPasswordHashWithoutPlaintextPassword(t *testing.T) {
	t.Setenv("AUTH_USERNAME", "admin")
	t.Setenv("AUTH_PASSWORD", "")
	t.Setenv("AUTH_PASSWORD_HASH", "$2a$10$abcdefghijklmnopqrstuu")
	t.Setenv("JWT_SECRET", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AuthPassword != "" || cfg.AuthPasswordHash == "" {
		t.Fatalf("unexpected auth config: %#v", cfg)
	}
}

func TestLoadRejectsPartialR2Config(t *testing.T) {
	t.Setenv("AUTH_USERNAME", "admin")
	t.Setenv("AUTH_PASSWORD", "password")
	t.Setenv("AUTH_PASSWORD_HASH", "")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("R2_ACCOUNT_ID", "account")
	t.Setenv("R2_ACCESS_KEY_ID", "")
	t.Setenv("R2_SECRET_ACCESS_KEY", "")
	t.Setenv("R2_BUCKET_NAME", "")
	t.Setenv("R2_PUBLIC_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected partial R2 config error")
	}
}
