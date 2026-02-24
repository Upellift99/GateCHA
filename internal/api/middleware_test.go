package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Upellift99/GateCHA/internal/auth"
	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
)

func TestAuthenticateAPIKey_QueryParam(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+key.KeyID, nil)
	w := httptest.NewRecorder()

	result, ok := authenticateAPIKey(db, w, req)
	if !ok {
		t.Fatal("expected authentication to succeed")
	}

	ctxKey := GetAPIKeyFromContext(result)
	if ctxKey == nil {
		t.Fatal("expected API key in context")
	}
	if ctxKey.ID != key.ID {
		t.Errorf("expected key ID %d, got %d", key.ID, ctxKey.ID)
	}
}

func TestAuthenticateAPIKey_BearerHeader(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	req := httptest.NewRequest("GET", "/api/v1/challenge", nil)
	req.Header.Set("Authorization", "Bearer "+key.KeyID)
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if !ok {
		t.Fatal("expected authentication to succeed via Bearer header")
	}
}

func TestAuthenticateAPIKey_Missing(t *testing.T) {
	db := testutil.SetupTestDB(t)

	req := httptest.NewRequest("GET", "/api/v1/challenge", nil)
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if ok {
		t.Fatal("expected authentication to fail for missing key")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticateAPIKey_InvalidPrefix(t *testing.T) {
	db := testutil.SetupTestDB(t)

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey=invalid_prefix", nil)
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if ok {
		t.Fatal("expected authentication to fail for invalid prefix")
	}
}

func TestAuthenticateAPIKey_NonexistentKey(t *testing.T) {
	db := testutil.SetupTestDB(t)

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey=gk_nonexistent000000000000", nil)
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if ok {
		t.Fatal("expected authentication to fail for nonexistent key")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticateAPIKey_Disabled(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")
	models.UpdateAPIKey(db, key.ID, models.UpdateAPIKeyParams{
		Name: key.Name, Domain: key.Domain, MaxNumber: key.MaxNumber,
		ExpireSeconds: key.ExpireSeconds, Algorithm: key.Algorithm, Enabled: false,
	})

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+key.KeyID, nil)
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if ok {
		t.Fatal("expected authentication to fail for disabled key")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestAuthenticateAPIKey_DomainAllowed(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "allowed.com", 0, 0, "")

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+key.KeyID, nil)
	req.Header.Set("Origin", "https://allowed.com")
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if !ok {
		t.Fatal("expected authentication to succeed for matching domain")
	}
}

func TestAuthenticateAPIKey_DomainBlocked(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "allowed.com", 0, 0, "")

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+key.KeyID, nil)
	req.Header.Set("Origin", "https://evil.com")
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if ok {
		t.Fatal("expected authentication to fail for non-matching domain")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestAuthenticateAPIKey_NoDomainNoOrigin(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "restricted.com", 0, 0, "")

	// No Origin header = no domain check
	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+key.KeyID, nil)
	w := httptest.NewRecorder()

	_, ok := authenticateAPIKey(db, w, req)
	if !ok {
		t.Fatal("expected authentication to succeed when no Origin header")
	}
}

func TestAuthenticateAdmin_Valid(t *testing.T) {
	secret := "test-secret"
	token, _, _ := auth.GenerateJWT("admin", secret)

	req := httptest.NewRequest("GET", "/api/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	ok := authenticateAdmin(secret, w, req)
	if !ok {
		t.Fatal("expected admin authentication to succeed")
	}
}

func TestAuthenticateAdmin_NoHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/admin/me", nil)
	w := httptest.NewRecorder()

	ok := authenticateAdmin("secret", w, req)
	if ok {
		t.Fatal("expected admin authentication to fail without header")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticateAdmin_InvalidToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/admin/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	ok := authenticateAdmin("secret", w, req)
	if ok {
		t.Fatal("expected admin authentication to fail with invalid token")
	}
}

func TestCORSMiddleware_AllowAll(t *testing.T) {
	handler := CORSMiddleware(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS origin *")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected CORS methods header")
	}
}

func TestCORSMiddleware_SpecificOrigin(t *testing.T) {
	handler := CORSMiddleware(false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected origin echo, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	handler := CORSMiddleware(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", w.Code)
	}
}

func TestMatchDomain(t *testing.T) {
	tests := []struct {
		url, domain string
		expected    bool
	}{
		{"https://example.com/path", "example.com", true},
		{"http://example.com:8080/path", "example.com", true},
		{"https://other.com", "example.com", false},
		{"https://EXAMPLE.COM", "example.com", true},
		{"example.com/path", "example.com", true},
	}

	for _, tt := range tests {
		result := matchDomain(tt.url, tt.domain)
		if result != tt.expected {
			t.Errorf("matchDomain(%q, %q) = %v, want %v", tt.url, tt.domain, result, tt.expected)
		}
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"hello": "world"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", w.Header().Get("Content-Type"))
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["hello"] != "world" {
		t.Errorf("expected hello=world, got %v", body)
	}
}
