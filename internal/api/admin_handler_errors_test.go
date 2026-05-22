package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Upellift99/GateCHA/internal/models"
)

// do is a small helper: performs an authenticated admin request and returns the
// recorder.
func do(t *testing.T, router http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, path, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	r.Header.Set("Authorization", "Bearer "+getAdminToken(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}

// ListKeys on an empty DB returns an empty (non-nil) list -> covers the nil
// normalization branch.
func TestListKeys_Empty(t *testing.T) {
	router, _ := setupTestRouter(t)
	w := do(t, router, "GET", "/api/admin/keys", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Keys []models.APIKey `json:"keys"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Keys == nil {
		t.Error("expected non-nil keys array")
	}
}

func TestListKeys_DBError(t *testing.T) {
	router, db := setupTestRouter(t)
	db.Migrator().DropTable(&models.APIKey{})
	if w := do(t, router, "GET", "/api/admin/keys", nil); w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestCreateKey_DBError(t *testing.T) {
	router, db := setupTestRouter(t)
	db.Migrator().DropTable(&models.APIKey{})
	body, _ := json.Marshal(map[string]interface{}{"name": "x"})
	if w := do(t, router, "POST", "/api/admin/keys", body); w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestDeleteKey_DBError(t *testing.T) {
	router, db := setupTestRouter(t)
	db.Migrator().DropTable(&models.APIKey{})
	if w := do(t, router, "DELETE", "/api/admin/keys/1", nil); w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestRotateSecret_DBError(t *testing.T) {
	router, db := setupTestRouter(t)
	db.Migrator().DropTable(&models.APIKey{})
	if w := do(t, router, "POST", "/api/admin/keys/1/rotate-secret", nil); w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestKeysStatsSummary_DBError(t *testing.T) {
	router, db := setupTestRouter(t)
	db.Migrator().DropTable(&models.DailyStat{})
	if w := do(t, router, "GET", "/api/admin/stats/keys-summary", nil); w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestStatsOverview_DBError(t *testing.T) {
	router, db := setupTestRouter(t)
	db.Migrator().DropTable(&models.DailyStat{})
	if w := do(t, router, "GET", "/api/admin/stats/overview", nil); w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestKeyStats_DBError(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")
	db.Migrator().DropTable(&models.DailyStat{})
	path := "/api/admin/stats/keys/" + strconv.FormatInt(key.ID, 10)
	if w := do(t, router, "GET", path, nil); w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGetSettings_DBError(t *testing.T) {
	router, db := setupTestRouter(t)
	db.Migrator().DropTable(&models.Setting{})
	if w := do(t, router, "GET", "/api/admin/settings", nil); w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestUpdateSettings_EnableDBError(t *testing.T) {
	router, db := setupTestRouter(t)
	db.Migrator().DropTable(&models.APIKey{}) // EnsureLoginCaptchaAPIKey will fail
	body, _ := json.Marshal(map[string]interface{}{"login_captcha_enabled": true})
	if w := do(t, router, "PUT", "/api/admin/settings", body); w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestUpdateSettings_SetDBError(t *testing.T) {
	router, db := setupTestRouter(t)
	db.Migrator().DropTable(&models.Setting{}) // SetSetting will fail
	body, _ := json.Marshal(map[string]interface{}{"login_captcha_enabled": false})
	if w := do(t, router, "PUT", "/api/admin/settings", body); w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// Login: GetLoginCaptchaEnabled fails (settings table missing) -> 500.
func TestLogin_GetCaptchaEnabledDBError(t *testing.T) {
	router, db := setupTestRouter(t)
	db.Migrator().DropTable(&models.Setting{})
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "password123"})
	req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// Login with captcha enabled where the captcha key can't be resolved
// (api_keys missing) -> 500 from verifyLoginCaptcha.
func TestLogin_VerifyCaptchaEnsureKeyError(t *testing.T) {
	router, db := setupTestRouter(t)
	models.SetSetting(db, models.SettingLoginCaptchaEnabled, "true")
	db.Migrator().DropTable(&models.APIKey{})
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "password123"})
	req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// Login with captcha enabled + invalid payload, daily_stats missing: the
// verifications_fail increment errors (logged) and the response is 401.
func TestLogin_CaptchaFailIncrementError(t *testing.T) {
	router, db := setupTestRouter(t)
	models.SetSetting(db, models.SettingLoginCaptchaEnabled, "true")
	models.EnsureLoginCaptchaAPIKey(db)
	db.Migrator().DropTable(&models.DailyStat{})
	body, _ := json.Marshal(map[string]interface{}{
		"username":       "admin",
		"password":       "password123",
		"altcha_payload": "invalid",
	})
	req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
