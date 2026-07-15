package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"recipe-backend/internal/config"
	"recipe-backend/internal/middleware"
	"recipe-backend/internal/respond"
)

type AuthHandler struct {
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !h.validCredentials(req.Username, req.Password) {
		respond.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := middleware.GenerateToken(req.Username, h.cfg.JWTSecret)
	if err != nil {
		respond.InternalError(w, "failed to sign token", err)
		return
	}

	respond.JSON(w, http.StatusOK, loginResponse{Token: token})
}

func (h *AuthHandler) validCredentials(username, password string) bool {
	if subtle.ConstantTimeCompare([]byte(username), []byte(h.cfg.AuthUsername)) != 1 {
		return false
	}
	if h.cfg.AuthPasswordHash != "" {
		return bcrypt.CompareHashAndPassword([]byte(h.cfg.AuthPasswordHash), []byte(password)) == nil
	}
	return subtle.ConstantTimeCompare([]byte(password), []byte(h.cfg.AuthPassword)) == 1
}
