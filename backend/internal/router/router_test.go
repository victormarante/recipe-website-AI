package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"recipe-backend/internal/config"
	"recipe-backend/internal/database"
	"recipe-backend/internal/handlers"
	"recipe-backend/internal/repository"
	"recipe-backend/internal/router"
)

func setupRouter(t *testing.T) http.Handler {
	t.Helper()

	db, err := database.New(filepath.Join(t.TempDir(), "router.db"))
	if err != nil {
		t.Fatalf("database.New: %v", err)
	}
	t.Cleanup(func() { database.Close(db) })
	if err := database.RunMigrations(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	repo := repository.NewRecipeRepository(db)
	cfg := &config.Config{
		Port:         "8080",
		Environment:  "test",
		DatabasePath: ":memory:",
		CORSOrigins:  []string{"*"},
		AuthUsername: "admin",
		AuthPassword: "password",
		JWTSecret:    "test-secret",
	}

	return router.New(
		handlers.NewRecipeHandler(repo, cfg),
		handlers.NewCategoryHandler(repo),
		handlers.NewAuthHandler(cfg),
		cfg,
		db,
	)
}

func authToken(t *testing.T, h http.Handler) string {
	t.Helper()

	body := strings.NewReader(`{"username":"admin","password":"password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("login status %d body %s", rr.Code, rr.Body.String())
	}
	var response struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	return response.Token
}

func authorizedRequest(method, path, token string, body *strings.Reader) *http.Request {
	if body == nil {
		body = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestProtectedRoutesRejectUnauthenticatedRequests(t *testing.T) {
	h := setupRouter(t)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/recipes", nil))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	assertJSONError(t, rr, "missing authorization header")
}

func TestLoginInvalidCredentials(t *testing.T) {
	h := setupRouter(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	assertJSONError(t, rr, "invalid credentials")
}

func TestAuthenticatedRecipeListingInvalidJSONValidationAndMissingRecipe(t *testing.T) {
	h := setupRouter(t)
	token := authToken(t, h)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, authorizedRequest(http.MethodGet, "/api/v1/recipes", token, nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d body %s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, authorizedRequest(http.MethodPost, "/api/v1/recipes", token, strings.NewReader(`{`)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("invalid json expected 400, got %d", rr.Code)
	}
	assertJSONError(t, rr, "Invalid request body")

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, authorizedRequest(http.MethodPost, "/api/v1/recipes", token, strings.NewReader(`{"title":""}`)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("validation expected 400, got %d", rr.Code)
	}
	assertJSONError(t, rr, "Validation failed")

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, authorizedRequest(http.MethodGet, "/api/v1/recipes/999", token, nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("missing expected 404, got %d", rr.Code)
	}
	assertJSONError(t, rr, "Recipe not found")
}

func TestCategoryJSONEscapingAndReadiness(t *testing.T) {
	h := setupRouter(t)
	token := authToken(t, h)

	payload := []byte(`{
		"title":"Quoted",
		"description":"category escaping",
		"categories":["Dinner \"Special\"","Sauce\\Glaze"],
		"ingredients":["salt"],
		"steps":["cook"],
		"links":[]
	}`)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipes", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create expected 201, got %d body %s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, authorizedRequest(http.MethodGet, "/api/v1/categories", token, nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("categories expected 200, got %d body %s", rr.Code, rr.Body.String())
	}
	var categories []string
	if err := json.NewDecoder(rr.Body).Decode(&categories); err != nil {
		t.Fatalf("decode categories: %v", err)
	}
	if len(categories) != 2 {
		t.Fatalf("expected 2 categories, got %#v", categories)
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("ready expected 200, got %d", rr.Code)
	}
}

func TestImageEndpointUnavailableWithoutStorage(t *testing.T) {
	h := setupRouter(t)
	token := authToken(t, h)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, authorizedRequest(http.MethodPost, "/api/v1/recipes/1/image", token, strings.NewReader("")))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
	assertJSONError(t, rr, "Image storage not configured")
}

func assertJSONError(t *testing.T, rr *httptest.ResponseRecorder, expected string) {
	t.Helper()
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected json content type, got %q", rr.Header().Get("Content-Type"))
	}
	var response struct {
		Error   string `json:"error"`
		Message string `json:"message,omitempty"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v body %s", err, rr.Body.String())
	}
	if response.Error != expected {
		t.Fatalf("expected error %q, got %#v", expected, response)
	}
	if response.Message != "" {
		t.Fatalf("internal message should not be exposed: %#v", response)
	}
}
