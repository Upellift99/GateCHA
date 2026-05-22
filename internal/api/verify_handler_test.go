package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Upellift99/GateCHA/internal/altcha"
	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
)

func verifyRequestWithKey(key *models.APIKey, body []byte) *http.Request {
	req := httptest.NewRequest("POST", "/api/v1/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if key == nil {
		return req
	}
	ctx := context.WithValue(req.Context(), apiKeyContextKey, key)
	return req.WithContext(ctx)
}

// solvedPayloadBody returns a verify request body carrying a correctly solved
// challenge for key, so VerifyPayload returns ok=true.
func solvedPayloadBody(t *testing.T, key *models.APIKey) []byte {
	t.Helper()
	ch, err := altcha.GenerateChallenge(key.HMACSecret, key.MaxNumber, key.Algorithm, key.ExpireSeconds)
	if err != nil {
		t.Fatalf("GenerateChallenge failed: %v", err)
	}
	number := solveChallenge(t, ch.Challenge, ch.Salt, ch.MaxNumber)
	payload := map[string]interface{}{
		"algorithm": ch.Algorithm,
		"challenge": ch.Challenge,
		"number":    number,
		"salt":      ch.Salt,
		"signature": ch.Signature,
	}
	pj, _ := json.Marshal(payload)
	body, _ := json.Marshal(map[string]string{"payload": base64.StdEncoding.EncodeToString(pj)})
	return body
}

// Defensive branch: no API key in context -> 401.
func TestVerifyHandler_NilKey(t *testing.T) {
	db := testutil.SetupTestDB(t)
	h := &VerifyHandler{DB: db}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, verifyRequestWithKey(nil, []byte(`{"payload":"x"}`)))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// VerifyPayload errors (unsupported algorithm) -> "verification failed". With
// daily_stats dropped, the recordFail counter increment also errors and is
// logged (covers recordFail's error branch).
func TestVerifyHandler_VerifyErrorAndFailIncrementError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, err := models.CreateAPIKey(db, "Test", "", 100, 60, "SHA-256")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}
	if err := db.Migrator().DropTable(&models.DailyStat{}); err != nil {
		t.Fatalf("drop daily_stats: %v", err)
	}

	payload := map[string]interface{}{
		"algorithm": "NOT-A-REAL-ALGO",
		"challenge": "abcdef",
		"number":    1,
		"salt":      "salt",
		"signature": "sig",
	}
	pj, _ := json.Marshal(payload)
	body, _ := json.Marshal(map[string]string{"payload": base64.StdEncoding.EncodeToString(pj)})

	h := &VerifyHandler{DB: db}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, verifyRequestWithKey(key, body))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp verifyResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.OK || resp.Error != "verification failed" {
		t.Errorf("expected ok=false error='verification failed', got ok=%v error=%q", resp.OK, resp.Error)
	}
}

// Replay check fails (consumed_challenges table missing) -> 500.
func TestVerifyHandler_IsConsumedError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, err := models.CreateAPIKey(db, "Test", "", 100, 60, "SHA-256")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}
	body := solvedPayloadBody(t, key)

	if err := db.Migrator().DropTable(&models.ConsumedChallenge{}); err != nil {
		t.Fatalf("drop consumed_challenges: %v", err)
	}

	h := &VerifyHandler{DB: db}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, verifyRequestWithKey(key, body))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// Success path where the verifications_ok counter increment fails: the error is
// logged and the request still succeeds (200, ok=true).
func TestVerifyHandler_IncrementOKError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, err := models.CreateAPIKey(db, "Test", "", 100, 60, "SHA-256")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}
	body := solvedPayloadBody(t, key)

	if err := db.Migrator().DropTable(&models.DailyStat{}); err != nil {
		t.Fatalf("drop daily_stats: %v", err)
	}

	h := &VerifyHandler{DB: db}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, verifyRequestWithKey(key, body))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp verifyResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.OK {
		t.Errorf("expected ok=true, got error=%q", resp.Error)
	}
}
