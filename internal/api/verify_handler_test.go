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

// humanSignals scores 0: real motion and pointer travel, unhurried window.
func humanSignals() map[string]interface{} {
	return map[string]interface{}{
		"duration_ms":      5000,
		"time_to_first_ms": 800,
		"pointer_events":   40,
		"pointer_distance": 400,
	}
}

// botSignals scores 0.9: no motion at all, no pointer path, instant solve.
func botSignals() map[string]interface{} {
	return map[string]interface{}{
		"duration_ms":      10,
		"time_to_first_ms": -1,
		"pointer_events":   0,
		"pointer_distance": 0,
	}
}

// postVerify sends body to /api/v1/verify for key and returns the decoded
// response as a raw map, so tests can assert on field *presence* and not only
// on the zero values a typed decode would hide.
func postVerify(t *testing.T, router http.Handler, key *models.APIKey, body []byte) map[string]interface{} {
	t.Helper()
	req := httptest.NewRequest("POST", "/api/v1/verify?apiKey="+key.KeyID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var raw map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&raw); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return raw
}

// A caller enforcing its own threshold needs the score on rejected submissions
// too, which are exactly the ones worth inspecting.
func TestVerify_HISScoreReturnedOnFailedVerification(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "HIS score", "", 100, 60, "SHA-256")

	body, _ := json.Marshal(map[string]interface{}{
		"payload":     "not-valid-base64!!",
		"his_signals": botSignals(),
	})
	raw := postVerify(t, router, key, body)

	if raw["ok"] != false {
		t.Errorf("ok = %v, want false: HIS must not rescue an invalid payload", raw["ok"])
	}
	if got := raw["his_bot_score"]; got != 0.9 {
		t.Errorf("his_bot_score = %v, want 0.9", got)
	}
	if got := raw["his_bot_suspected"]; got != true {
		t.Errorf("his_bot_suspected = %v, want true", got)
	}
}

func TestVerify_HISScoreReturnedOnSuccess(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "HIS ok", "", 100, 60, "SHA-256")

	var body map[string]interface{}
	if err := json.Unmarshal(solvedPayloadBody(t, key), &body); err != nil {
		t.Fatalf("unmarshal solved body: %v", err)
	}
	body["his_signals"] = botSignals()
	encoded, _ := json.Marshal(body)
	raw := postVerify(t, router, key, encoded)

	// The score rides along with a successful verification rather than
	// overriding it: acting on it is the integrator's call, not ours.
	if raw["ok"] != true {
		t.Fatalf("ok = %v, want true: bot-like signals must not block", raw["ok"])
	}
	if got := raw["his_bot_score"]; got != 0.9 {
		t.Errorf("his_bot_score = %v, want 0.9", got)
	}
	if got := raw["his_bot_suspected"]; got != true {
		t.Errorf("his_bot_suspected = %v, want true", got)
	}
}

// Absence must stay absence. A client shipping no collector is not thereby a
// human, so the fields are omitted rather than reported as a clean 0.
func TestVerify_HISFieldsOmittedWithoutSignals(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "HIS none", "", 100, 60, "SHA-256")

	raw := postVerify(t, router, key, []byte(`{"payload":"not-valid-base64!!"}`))

	if _, present := raw["his_bot_score"]; present {
		t.Errorf("his_bot_score present without his_signals: %v", raw)
	}
	if _, present := raw["his_bot_suspected"]; present {
		t.Errorf("his_bot_suspected present without his_signals: %v", raw)
	}
}

// The mirror image: a genuine 0 is a real score and must be reported. This is
// what a plain float64 with omitempty would silently swallow.
func TestVerify_HISZeroScoreIsReported(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "HIS zero", "", 100, 60, "SHA-256")

	body, _ := json.Marshal(map[string]interface{}{
		"payload":     "not-valid-base64!!",
		"his_signals": humanSignals(),
	})
	raw := postVerify(t, router, key, body)

	score, present := raw["his_bot_score"]
	if !present {
		t.Fatalf("his_bot_score omitted for a clean human sample: %v", raw)
	}
	if score != 0.0 {
		t.Errorf("his_bot_score = %v, want 0", score)
	}
	if got := raw["his_bot_suspected"]; got != false {
		t.Errorf("his_bot_suspected = %v, want false", got)
	}
}
