package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Upellift99/GateCHA/internal/auth"
	"github.com/Upellift99/GateCHA/internal/testutil"
)

func TestVersionEndpoint_DefaultsToDev(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/admin/version", nil)
	req.Header.Set("Authorization", "Bearer "+getAdminToken(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// setupTestRouter leaves Version empty, so the handler falls back to "dev".
	if resp["version"] != "dev" {
		t.Errorf("version = %q, want \"dev\"", resp["version"])
	}
}

func TestVersionEndpoint_RequiresAuth(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/admin/version", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without a token, got %d", w.Code)
	}
}

func TestVersionEndpoint_ReportsConfiguredVersion(t *testing.T) {
	db := setupTestRouterDBWithVersion(t)
	req := httptest.NewRequest("GET", "/api/admin/version", nil)
	req.Header.Set("Authorization", "Bearer "+getAdminToken(t))
	w := httptest.NewRecorder()
	db.ServeHTTP(w, req)

	var resp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["version"] != "v1.2.3" {
		t.Errorf("version = %q, want \"v1.2.3\"", resp["version"])
	}
}

// setupTestRouterDBWithVersion builds a router with an explicit build version.
func setupTestRouterDBWithVersion(t *testing.T) http.Handler {
	t.Helper()
	db := testutil.SetupTestDB(t)
	auth.EnsureAdminUser(db, "admin", "password123")
	return NewRouter(db, testSecretKey, RouterConfig{CORSAllowAll: true, Version: "v1.2.3"})
}
