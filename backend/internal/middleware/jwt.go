package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"recipe-backend/internal/respond"
)

var now = time.Now

const tokenLifetime = 90 * 24 * time.Hour

// GenerateToken creates a signed HS256 JWT valid for 90 days.
func GenerateToken(subject, secret string) (string, error) {
	return GenerateTokenWithExpiry(subject, secret, now().Add(tokenLifetime))
}

// GenerateTokenWithExpiry creates a signed HS256 JWT with a caller-supplied expiry.
func GenerateTokenWithExpiry(subject, secret string, expiresAt time.Time) (string, error) {
	header := base64url([]byte(`{"alg":"HS256","typ":"JWT"}`))

	payload, err := json.Marshal(map[string]any{
		"sub": subject,
		"exp": expiresAt.Unix(),
	})
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	unsigned := header + "." + base64url(payload)
	sig := sign(unsigned, secret)
	return unsigned + "." + sig, nil
}

// JWTAuth is middleware that validates a Bearer JWT on every request.
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respond.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				respond.Error(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			if err := ValidateToken(parts[1], secret); err != nil {
				respond.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ValidateToken validates a signed HS256 JWT created by GenerateToken.
func ValidateToken(tokenStr, secret string) error {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return fmt.Errorf("malformed token")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("invalid header encoding")
	}
	var header map[string]any
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return fmt.Errorf("invalid header json")
	}
	if header["alg"] != "HS256" || header["typ"] != "JWT" {
		return fmt.Errorf("invalid token header")
	}

	unsigned := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(sign(unsigned, secret))) {
		return fmt.Errorf("invalid signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("invalid payload encoding")
	}

	var claims map[string]any
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return fmt.Errorf("invalid payload json")
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return fmt.Errorf("missing sub claim")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("missing exp claim")
	}
	if now().Unix() > int64(exp) {
		return fmt.Errorf("token expired")
	}

	return nil
}

func sign(data, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return base64url(mac.Sum(nil))
}

func base64url(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
