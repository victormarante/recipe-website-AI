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

const testPIN = "654321"

// login returns an admin Bearer header value obtained via the login endpoint.
func login(t *testing.T, h http.Handler) string {
	t.Helper()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"pin":"`+testPIN+`"}`))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("login expected 200, got %d body %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil || resp.Token == "" {
		t.Fatalf("decode login: %v token %q", err, resp.Token)
	}
	return "Bearer " + resp.Token
}

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
		AdminPIN:     testPIN,
		AuthSecret:   "router-test-secret",
	}

	return router.New(
		handlers.NewRecipeHandler(repo, cfg),
		handlers.NewCategoryHandler(repo),
		cfg,
		db,
	)
}

func TestRecipeListingInvalidJSONValidationAndMissingRecipe(t *testing.T) {
	h := setupRouter(t)
	auth := login(t, h)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/recipes", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d body %s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipes", strings.NewReader(`{`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("invalid json expected 400, got %d", rr.Code)
	}
	assertJSONError(t, rr, "Invalid request body")

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/recipes", strings.NewReader(`{"title":""}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("validation expected 400, got %d", rr.Code)
	}
	assertJSONError(t, rr, "Validation failed")

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/recipes/999", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("missing expected 404, got %d", rr.Code)
	}
	assertJSONError(t, rr, "Recipe not found")
}

func TestCategoryJSONEscapingAndReadiness(t *testing.T) {
	h := setupRouter(t)
	auth := login(t, h)

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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create expected 201, got %d body %s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil))
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
	auth := login(t, h)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipes/1/image", strings.NewReader(""))
	req.Header.Set("Authorization", auth)
	h.ServeHTTP(rr, req)
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

func TestWriteRoutesRequireAdminToken(t *testing.T) {
	h := setupRouter(t)

	writes := []struct{ method, path string }{
		{http.MethodPost, "/api/v1/recipes"},
		{http.MethodPut, "/api/v1/recipes/1"},
		{http.MethodDelete, "/api/v1/recipes/1"},
		{http.MethodPost, "/api/v1/recipes/1/image"},
		{http.MethodDelete, "/api/v1/recipes/1/image"},
	}
	for _, w := range writes {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(w.method, w.path, strings.NewReader("{}")))
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s expected 401, got %d", w.method, w.path, rr.Code)
		}
	}

	for _, path := range []string{"/api/v1/recipes", "/api/v1/categories"} {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("GET %s expected 200 without token, got %d", path, rr.Code)
		}
	}
}

func TestLoginRejectsWrongPINAndRateLimits(t *testing.T) {
	h := setupRouter(t)

	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"pin":"000000"}`)))
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d expected 401, got %d", i, rr.Code)
		}
	}

	// Locked out, even for the correct PIN.
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"pin":"`+testPIN+`"}`)))
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
}
