package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testSecret = "test-secret"

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
}

func TestRequireAdmin(t *testing.T) {
	valid := GenerateToken(testSecret, time.Now().Add(time.Hour))
	expired := GenerateToken(testSecret, time.Now().Add(-time.Hour))
	wrongSecret := GenerateToken("other", time.Now().Add(time.Hour))

	cases := []struct {
		name   string
		header string
		want   int
	}{
		{"missing", "", http.StatusUnauthorized},
		{"wrong scheme", "Basic " + valid, http.StatusUnauthorized},
		{"no scheme", valid, http.StatusUnauthorized},
		{"malformed", "Bearer abc", http.StatusUnauthorized},
		{"expired", "Bearer " + expired, http.StatusUnauthorized},
		{"wrong secret", "Bearer " + wrongSecret, http.StatusUnauthorized},
		{"tampered", "Bearer x" + valid, http.StatusUnauthorized},
		{"valid", "Bearer " + valid, http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rr := httptest.NewRecorder()
			RequireAdmin(testSecret)(okHandler()).ServeHTTP(rr, req)
			if rr.Code != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, rr.Code)
			}
		})
	}
}
