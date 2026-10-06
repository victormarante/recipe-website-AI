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
