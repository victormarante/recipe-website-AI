package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestJWTAuthFailures(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{name: "missing"},
		{name: "malformed", header: "Token abc"},
		{name: "invalid", header: "Bearer invalid.token.value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rr := httptest.NewRecorder()
			JWTAuth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("handler should not be called")
			})).ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", rr.Code)
			}
			if rr.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("expected json content type, got %q", rr.Header().Get("Content-Type"))
			}
		})
	}
}

func TestJWTAuthExpiredAndValidToken(t *testing.T) {
	expired, err := GenerateTokenWithExpiry("admin", "secret", time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("GenerateTokenWithExpiry expired: %v", err)
	}
	if err := ValidateToken(expired, "secret"); err == nil {
		t.Fatal("expected expired token validation to fail")
	}

	valid, err := GenerateToken("admin", "secret")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if err := ValidateToken(valid, "secret"); err != nil {
		t.Fatalf("ValidateToken valid: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+valid)
	rr := httptest.NewRecorder()
	called := false
	JWTAuth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler was not called")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestGenerateTokenUsesNinetyDayExpiry(t *testing.T) {
	originalNow := now
	defer func() { now = originalNow }()

	base := time.Date(2026, time.July, 21, 12, 0, 0, 0, time.UTC)
	now = func() time.Time { return base }

	token, err := GenerateToken("admin", "secret")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	now = func() time.Time { return base.Add(89 * 24 * time.Hour) }
	if err := ValidateToken(token, "secret"); err != nil {
		t.Fatalf("expected token to still be valid before 90 days: %v", err)
	}

	now = func() time.Time { return base.Add(91 * 24 * time.Hour) }
	if err := ValidateToken(token, "secret"); err == nil {
		t.Fatal("expected token to expire after 90 days")
	}
}
