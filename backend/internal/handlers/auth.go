package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"recipe-backend/internal/config"
	"recipe-backend/internal/middleware"
	"recipe-backend/internal/respond"
)

const (
	maxLoginAttempts = 5
	loginWindow      = 15 * time.Minute
)

// AuthHandler exchanges the admin PIN for a signed token.
type AuthHandler struct {
	pin    string
	secret string

	mu       sync.Mutex
	failures map[string][]time.Time
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{pin: cfg.AdminPIN, secret: cfg.AuthSecret, failures: map[string][]time.Time{}}
}

type loginRequest struct {
	PIN string `json:"pin"`
}

type loginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	if h.tooManyFailures(ip) {
		respond.Error(w, http.StatusTooManyRequests, "Too many attempts, try again later")
		return
	}

	var req loginRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if subtle.ConstantTimeCompare([]byte(req.PIN), []byte(h.pin)) != 1 {
		h.recordFailure(ip)
		respond.Error(w, http.StatusUnauthorized, "Invalid PIN")
		return
	}

	expiresAt := time.Now().Add(middleware.TokenLifetime)
	respond.JSON(w, http.StatusOK, loginResponse{
		Token:     middleware.GenerateToken(h.secret, expiresAt),
		ExpiresAt: expiresAt,
	})
}

// recent returns failures inside the window and prunes the rest. Caller holds mu.
func (h *AuthHandler) recent(ip string) []time.Time {
	cutoff := time.Now().Add(-loginWindow)
	kept := h.failures[ip][:0]
	for _, t := range h.failures[ip] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		delete(h.failures, ip)
		return nil
	}
	h.failures[ip] = kept
	return kept
}

func (h *AuthHandler) tooManyFailures(ip string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.recent(ip)) >= maxLoginAttempts
}

func (h *AuthHandler) recordFailure(ip string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.failures[ip] = append(h.recent(ip), time.Now())
}
