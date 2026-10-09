package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"recipe-backend/internal/respond"
)

var now = time.Now

// TokenLifetime is how long an issued admin token stays valid.
const TokenLifetime = 90 * 24 * time.Hour

// GenerateToken creates a signed admin token: base64url(expiry unix) "." base64url(HMAC-SHA256).
func GenerateToken(secret string, expiresAt time.Time) string {
	payload := base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatInt(expiresAt.Unix(), 10)))
	return payload + "." + sign(payload, secret)
}

// ValidateToken verifies the signature and expiry of a token from GenerateToken.
func ValidateToken(token, secret string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return fmt.Errorf("malformed token")
	}
	if !hmac.Equal([]byte(parts[1]), []byte(sign(parts[0], secret))) {
		return fmt.Errorf("invalid signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("invalid payload encoding")
	}
	exp, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid payload")
	}
	if now().Unix() >= exp {
		return fmt.Errorf("token expired")
	}
	return nil
}

// RequireAdmin rejects requests without a valid Bearer admin token.
func RequireAdmin(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				respond.Error(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}
			scheme, token, ok := strings.Cut(header, " ")
			if !ok || scheme != "Bearer" {
				respond.Error(w, http.StatusUnauthorized, "Invalid authorization format")
				return
			}
			if err := ValidateToken(token, secret); err != nil {
				respond.Error(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func sign(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
