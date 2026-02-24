package api

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Upellift99/GateCHA/internal/auth"
	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
)

const testSecretKey = "test-secret-key-for-jwt"

func setupTestRouter(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	auth.EnsureAdminUser(db, "admin", "password123")
	router := NewRouter(db, testSecretKey, true)
	return router, db
}

func getAdminToken(t *testing.T) string {
	t.Helper()
	token, _, err := auth.GenerateJWT("admin", testSecretKey)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}
	return token
}

func TestLogin_Success(t *testing.T) {
	router, _ := setupTestRouter(t)

	body, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected token in response")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	router, _ := setupTestRouter(t)

	body, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "wrong",
	})
	req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_InvalidBody(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestMe(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	req := httptest.NewRequest("GET", "/api/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["username"] != "admin" {
		t.Errorf("expected admin, got %s", resp["username"])
	}
}

func TestMe_Unauthorized(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/admin/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestKeyCRUD(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	// Create
	body, _ := json.Marshal(map[string]interface{}{
		"name":           "Test Key",
		"domain":         "test.com",
		"max_number":     50000,
		"expire_seconds": 120,
	})
	req := httptest.NewRequest("POST", "/api/admin/keys", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var created map[string]interface{}
	json.NewDecoder(w.Body).Decode(&created)
	keyID := int64(created["id"].(float64))
	idStr := strconv.FormatInt(keyID, 10)

	// List
	req = httptest.NewRequest("GET", "/api/admin/keys", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", w.Code)
	}

	// Get
	req = httptest.NewRequest("GET", "/api/admin/keys/"+idStr, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", w.Code)
	}

	// Update
	updateBody, _ := json.Marshal(map[string]interface{}{"name": "Updated Key"})
	req = httptest.NewRequest("PUT", "/api/admin/keys/"+idStr, bytes.NewReader(updateBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Rotate secret
	req = httptest.NewRequest("POST", "/api/admin/keys/"+idStr+"/rotate-secret", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("rotate: expected 200, got %d", w.Code)
	}

	// Delete
	req = httptest.NewRequest("DELETE", "/api/admin/keys/"+idStr, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d", w.Code)
	}
}

func TestCreateKey_InvalidBody(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	req := httptest.NewRequest("POST", "/api/admin/keys", bytes.NewReader([]byte("bad")))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetKey_NotFound(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	req := httptest.NewRequest("GET", "/api/admin/keys/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetKey_InvalidID(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	req := httptest.NewRequest("GET", "/api/admin/keys/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateKey_InvalidID(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	body, _ := json.Marshal(map[string]string{"name": "test"})
	req := httptest.NewRequest("PUT", "/api/admin/keys/abc", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateKey_NotFound(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	body, _ := json.Marshal(map[string]string{"name": "test"})
	req := httptest.NewRequest("PUT", "/api/admin/keys/99999", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDeleteKey_InvalidID(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	req := httptest.NewRequest("DELETE", "/api/admin/keys/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRotateSecret_InvalidID(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	req := httptest.NewRequest("POST", "/api/admin/keys/abc/rotate-secret", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestStatsEndpoints(t *testing.T) {
	router, db := setupTestRouter(t)
	token := getAdminToken(t)

	key, _ := models.CreateAPIKey(db, "Stats Key", "", 0, 0, "")
	models.IncrementChallengesIssued(db, key.ID)

	// Overview
	req := httptest.NewRequest("GET", "/api/admin/stats/overview", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("overview: expected 200, got %d", w.Code)
	}

	// Overview with days param
	req = httptest.NewRequest("GET", "/api/admin/stats/overview?days=7", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("overview with days: expected 200, got %d", w.Code)
	}

	// Keys summary
	req = httptest.NewRequest("GET", "/api/admin/stats/keys-summary", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("keys-summary: expected 200, got %d", w.Code)
	}

	// Key stats
	req = httptest.NewRequest("GET", "/api/admin/stats/keys/"+strconv.FormatInt(key.ID, 10), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("key stats: expected 200, got %d", w.Code)
	}
}

func TestKeyStats_InvalidID(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	req := httptest.NewRequest("GET", "/api/admin/stats/keys/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestKeyStats_NotFound(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	req := httptest.NewRequest("GET", "/api/admin/stats/keys/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestChangePassword_Success(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	body, _ := json.Marshal(map[string]string{
		"current_password": "password123",
		"new_password":     "newpassword456",
	})
	req := httptest.NewRequest("POST", "/api/admin/change-password", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	body, _ := json.Marshal(map[string]string{
		"current_password": "wrong",
		"new_password":     "newpassword456",
	})
	req := httptest.NewRequest("POST", "/api/admin/change-password", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestChangePassword_InvalidBody(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	req := httptest.NewRequest("POST", "/api/admin/change-password", bytes.NewReader([]byte("bad")))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSettingsEndpoints(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	// Get settings
	req := httptest.NewRequest("GET", "/api/admin/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get settings: expected 200, got %d", w.Code)
	}

	// Update settings - enable captcha
	body, _ := json.Marshal(map[string]interface{}{
		"login_captcha_enabled": true,
	})
	req = httptest.NewRequest("PUT", "/api/admin/settings", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("update settings: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Update settings - disable captcha
	body, _ = json.Marshal(map[string]interface{}{
		"login_captcha_enabled": false,
	})
	req = httptest.NewRequest("PUT", "/api/admin/settings", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("disable settings: expected 200, got %d", w.Code)
	}
}

func TestUpdateSettings_InvalidBody(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	req := httptest.NewRequest("PUT", "/api/admin/settings", bytes.NewReader([]byte("bad")))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHealthz(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestChallengeEndpoint(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+key.KeyID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChallengeEndpoint_NoKey(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/v1/challenge", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestVerifyEndpoint_InvalidBody(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	req := httptest.NewRequest("POST", "/api/v1/verify?apiKey="+key.KeyID, bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVerifyEndpoint_EmptyPayload(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	body, _ := json.Marshal(map[string]string{"payload": ""})
	req := httptest.NewRequest("POST", "/api/v1/verify?apiKey="+key.KeyID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestVerifyEndpoint_InvalidPayload(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	body, _ := json.Marshal(map[string]string{"payload": "not-base64!!!"})
	req := httptest.NewRequest("POST", "/api/v1/verify?apiKey="+key.KeyID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp verifyResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.OK {
		t.Error("expected OK=false for invalid payload")
	}
}

func TestVerifyEndpoint_InvalidBase64Content(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	// Valid base64 but not valid JSON inside
	body, _ := json.Marshal(map[string]string{"payload": "bm90anNvbg=="}) // "notjson" in base64
	req := httptest.NewRequest("POST", "/api/v1/verify?apiKey="+key.KeyID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp verifyResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.OK {
		t.Error("expected OK=false for invalid JSON inside base64")
	}
}

func TestVerifyEndpoint_ValidBase64InvalidSolution(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	// Construct a payload that decodes to valid JSON but has wrong solution
	payload := map[string]interface{}{
		"algorithm": "SHA-256",
		"challenge": "abcdef1234567890",
		"number":    42,
		"salt":      "test-salt",
		"signature": "invalid-signature",
	}
	payloadJSON, _ := json.Marshal(payload)
	payloadB64 := base64.StdEncoding.EncodeToString(payloadJSON)

	body, _ := json.Marshal(map[string]string{"payload": payloadB64})
	req := httptest.NewRequest("POST", "/api/v1/verify?apiKey="+key.KeyID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp verifyResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.OK {
		t.Error("expected OK=false for invalid solution")
	}
}

func TestVerifyEndpoint_NoKey(t *testing.T) {
	router, _ := setupTestRouter(t)

	body, _ := json.Marshal(map[string]string{"payload": "test"})
	req := httptest.NewRequest("POST", "/api/v1/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_WithCaptchaEnabled_NoCaptcha(t *testing.T) {
	router, db := setupTestRouter(t)

	// Enable login captcha
	models.SetSetting(db, models.SettingLoginCaptchaEnabled, "true")
	models.EnsureLoginCaptchaAPIKey(db)

	body, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 when captcha required but not provided, got %d", w.Code)
	}
}

func TestLogin_WithCaptchaEnabled_InvalidCaptcha(t *testing.T) {
	router, db := setupTestRouter(t)

	models.SetSetting(db, models.SettingLoginCaptchaEnabled, "true")
	models.EnsureLoginCaptchaAPIKey(db)

	body, _ := json.Marshal(map[string]interface{}{
		"username":       "admin",
		"password":       "password123",
		"altcha_payload": "invalid-captcha-payload",
	})
	req := httptest.NewRequest("POST", "/api/admin/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid captcha, got %d", w.Code)
	}
}

func TestPublicLoginConfig(t *testing.T) {
	router, _ := setupTestRouter(t)

	req := httptest.NewRequest("GET", "/api/public/login-config", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["captcha_required"] != false {
		t.Error("expected captcha_required=false by default")
	}
}

func solveChallenge(t *testing.T, challenge, salt string, maxNumber int64) int {
	t.Helper()
	for i := 0; i <= int(maxNumber); i++ {
		input := salt + strconv.Itoa(i)
		h := sha256.Sum256([]byte(input))
		if hex.EncodeToString(h[:]) == challenge {
			return i
		}
	}
	t.Fatal("failed to solve challenge")
	return -1
}

func TestVerifyEndpoint_FullFlow(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 100, 300, "SHA-256")

	// Get challenge
	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+key.KeyID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("challenge: expected 200, got %d", w.Code)
	}

	var challenge struct {
		Algorithm string `json:"algorithm"`
		Challenge string `json:"challenge"`
		MaxNumber int64  `json:"maxnumber"`
		Salt      string `json:"salt"`
		Signature string `json:"signature"`
	}
	json.NewDecoder(w.Body).Decode(&challenge)

	// Solve the challenge
	number := solveChallenge(t, challenge.Challenge, challenge.Salt, challenge.MaxNumber)

	// Create payload
	payload := map[string]interface{}{
		"algorithm": challenge.Algorithm,
		"challenge": challenge.Challenge,
		"number":    number,
		"salt":      challenge.Salt,
		"signature": challenge.Signature,
	}
	payloadJSON, _ := json.Marshal(payload)
	payloadB64 := base64.StdEncoding.EncodeToString(payloadJSON)

	// Verify - should succeed
	body, _ := json.Marshal(map[string]string{"payload": payloadB64})
	req = httptest.NewRequest("POST", "/api/v1/verify?apiKey="+key.KeyID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("verify: expected 200, got %d", w.Code)
	}

	var resp verifyResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.OK {
		t.Errorf("expected OK=true, got error: %s", resp.Error)
	}

	// Replay should fail
	body, _ = json.Marshal(map[string]string{"payload": payloadB64})
	req = httptest.NewRequest("POST", "/api/v1/verify?apiKey="+key.KeyID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var respReplay verifyResponse
	json.NewDecoder(w.Body).Decode(&respReplay)
	if respReplay.OK {
		t.Error("expected replay to fail")
	}
	if respReplay.Error != "already_used" {
		t.Errorf("expected error 'already_used', got %q", respReplay.Error)
	}
}

func TestChallengeEndpoint_CustomSettings(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Custom", "", 500, 60, "SHA-256")

	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+key.KeyID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["algorithm"] != "SHA-256" {
		t.Errorf("expected SHA-256 algorithm, got %v", resp["algorithm"])
	}
	if resp["maxNumber"] == nil {
		t.Error("expected maxNumber in response")
	}
}

func TestUpdateKey_AllFields(t *testing.T) {
	router, db := setupTestRouter(t)
	token := getAdminToken(t)

	key, _ := models.CreateAPIKey(db, "Original", "", 0, 0, "")
	idStr := strconv.FormatInt(key.ID, 10)

	enabled := false
	body, _ := json.Marshal(map[string]interface{}{
		"name":           "Updated Name",
		"domain":         "newdomain.com",
		"max_number":     999,
		"expire_seconds": 300,
		"algorithm":      "SHA-512",
		"enabled":        enabled,
	})
	req := httptest.NewRequest("PUT", "/api/admin/keys/"+idStr, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["name"] != "Updated Name" {
		t.Errorf("expected 'Updated Name', got %v", resp["name"])
	}
	if resp["domain"] != "newdomain.com" {
		t.Errorf("expected 'newdomain.com', got %v", resp["domain"])
	}
}

func TestUpdateKey_InvalidBody(t *testing.T) {
	router, db := setupTestRouter(t)
	token := getAdminToken(t)

	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")
	idStr := strconv.FormatInt(key.ID, 10)

	req := httptest.NewRequest("PUT", "/api/admin/keys/"+idStr, bytes.NewReader([]byte("bad")))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeleteKey_NotFound(t *testing.T) {
	router, _ := setupTestRouter(t)
	token := getAdminToken(t)

	req := httptest.NewRequest("DELETE", "/api/admin/keys/99999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Delete on nonexistent key still returns 200 (no error from SQLite)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRotateSecret_Success(t *testing.T) {
	router, db := setupTestRouter(t)
	token := getAdminToken(t)

	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")
	idStr := strconv.FormatInt(key.ID, 10)

	req := httptest.NewRequest("POST", "/api/admin/keys/"+idStr+"/rotate-secret", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["hmac_secret"] == "" {
		t.Error("expected new hmac_secret in response")
	}
	if resp["hmac_secret"] == key.HMACSecret {
		t.Error("expected hmac_secret to change")
	}
}

func TestKeyStats_WithDaysParam(t *testing.T) {
	router, db := setupTestRouter(t)
	token := getAdminToken(t)

	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")
	idStr := strconv.FormatInt(key.ID, 10)

	req := httptest.NewRequest("GET", "/api/admin/stats/keys/"+idStr+"?days=7", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLogin_WithCaptchaEnabled_ValidCaptcha(t *testing.T) {
	router, db := setupTestRouter(t)

	// Enable login captcha
	models.SetSetting(db, models.SettingLoginCaptchaEnabled, "true")
	captchaKey, _ := models.EnsureLoginCaptchaAPIKey(db)

	// Get a challenge for the captcha key
	req := httptest.NewRequest("GET", "/api/v1/challenge?apiKey="+captchaKey.KeyID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("challenge: expected 200, got %d", w.Code)
	}

	var challenge struct {
		Algorithm string `json:"algorithm"`
		Challenge string `json:"challenge"`
		MaxNumber int64  `json:"maxnumber"`
		Salt      string `json:"salt"`
		Signature string `json:"signature"`
	}
	json.NewDecoder(w.Body).Decode(&challenge)

	// Solve the challenge
	number := solveChallenge(t, challenge.Challenge, challenge.Salt, challenge.MaxNumber)

	// Create ALTCHA payload
	payload := map[string]interface{}{
		"algorithm": challenge.Algorithm,
		"challenge": challenge.Challenge,
		"number":    number,
		"salt":      challenge.Salt,
		"signature": challenge.Signature,
	}
	payloadJSON, _ := json.Marshal(payload)
	payloadB64 := base64.StdEncoding.EncodeToString(payloadJSON)

	// Login with valid captcha
	body, _ := json.Marshal(map[string]interface{}{
		"username":       "admin",
		"password":       "password123",
		"altcha_payload": payloadB64,
	})
	req = httptest.NewRequest("POST", "/api/admin/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid captcha login, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected token in login response")
	}
}

func TestPublicLoginConfig_CaptchaEnabled(t *testing.T) {
	router, db := setupTestRouter(t)

	models.SetSetting(db, models.SettingLoginCaptchaEnabled, "true")
	models.EnsureLoginCaptchaAPIKey(db)

	req := httptest.NewRequest("GET", "/api/public/login-config", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["captcha_required"] != true {
		t.Error("expected captcha_required=true when enabled")
	}
	if resp["challenge_url"] == nil || resp["challenge_url"] == "" {
		t.Error("expected challenge_url when captcha enabled")
	}
}
