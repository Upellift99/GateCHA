package api

import (
	"net/http"
	"net/http/httptest"
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
