package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
)

// GetLoginCaptchaEnabled fails (settings table missing) -> 500.
func TestPublicHandler_LoginConfig_GetEnabledError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if err := db.Migrator().DropTable(&models.Setting{}); err != nil {
		t.Fatalf("drop settings: %v", err)
	}

	h := &PublicHandler{DB: db}
	req := httptest.NewRequest("GET", "/api/public/login-config", nil)
	w := httptest.NewRecorder()
	h.LoginConfig(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// Captcha is enabled but resolving the captcha API key fails (api_keys table
// missing) -> 500.
func TestPublicHandler_LoginConfig_EnsureKeyError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	if err := models.SetSetting(db, models.SettingLoginCaptchaEnabled, "true"); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	if err := db.Migrator().DropTable(&models.APIKey{}); err != nil {
		t.Fatalf("drop api_keys: %v", err)
	}

	h := &PublicHandler{DB: db}
	req := httptest.NewRequest("GET", "/api/public/login-config", nil)
	w := httptest.NewRecorder()
	h.LoginConfig(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// The collector script is served unauthenticated, as JavaScript, and is the
// build output of web/src/lib/his-embed.ts. A 404 here means the binary was
// built without the frontend assets: run `make frontend` first.
func TestPublicHandler_HISCollector(t *testing.T) {
	h := &PublicHandler{DB: testutil.SetupTestDB(t)}
	req := httptest.NewRequest("GET", "/api/public/his.js", nil)
	w := httptest.NewRecorder()
	h.HISCollector(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (embedded dist/his.js missing? run make frontend)", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/javascript; charset=utf-8" {
		t.Errorf("unexpected Content-Type %q", ct)
	}
	if w.Header().Get("Cache-Control") == "" {
		t.Error("expected the script to be cacheable")
	}

	body := w.Body.String()
	if body == "" {
		t.Fatal("served an empty script")
	}
	// The hidden field name is the contract with integrators: the server reads
	// whatever the site forwards under this name into `his_signals`.
	if !strings.Contains(body, "gatecha_his_signals") {
		t.Error("script does not reference the gatecha_his_signals field")
	}
}
