package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Upellift99/GateCHA/internal/his"
	"github.com/Upellift99/GateCHA/internal/models"
)

// noEvidenceSignals is the shape that dominated the histogram in #149: the
// collector ran for a while and observed nothing at all. No motion (0.50) and
// no pointer path (0.20) and nothing else, so it scores exactly 0.70, the
// no-evidence tier that sits below the default threshold on purpose.
func noEvidenceSignals() map[string]interface{} {
	return map[string]interface{}{
		"duration_ms":      4000,
		"time_to_first_ms": -1,
		"pointer_events":   0,
		"pointer_distance": 0,
	}
}

// solvedBodyWith returns a genuinely solved verification body carrying the
// given signals, so enforcement is exercised on a request that would otherwise
// succeed rather than on one already doomed by the proof of work.
func solvedBodyWith(t *testing.T, key *models.APIKey, signals map[string]interface{}) []byte {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(solvedPayloadBody(t, key), &body); err != nil {
		t.Fatalf("unmarshal solved body: %v", err)
	}
	if signals != nil {
		body["his_signals"] = signals
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return encoded
}

// Monitor stays the default: a key nobody configured must not start blocking
// because its instance was upgraded.
func TestHISEnforce_OffByDefaultLetsSuspectedThrough(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Monitor", "", 100, 60, "SHA-256")

	raw := postVerify(t, router, key, solvedBodyWith(t, key, botSignals()))
	if raw["ok"] != true {
		t.Fatalf("ok = %v, want true with enforcement off (error %v)", raw["ok"], raw["error"])
	}
	if raw["his_bot_suspected"] != true {
		t.Error("the sample must still be reported as suspected")
	}
}

func TestHISEnforce_RejectsSuspectedAndCountsAFailure(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Blocking", "", 100, 60, "SHA-256")
	if err := models.UpdateAPIKeyFields(db, key.ID, map[string]any{"his_enforce": true}); err != nil {
		t.Fatalf("enable enforcement: %v", err)
	}

	raw := postVerify(t, router, key, solvedBodyWith(t, key, botSignals()))
	if raw["ok"] != false {
		t.Fatal("expected the verification to be rejected")
	}
	if raw["error"] != "bot_suspected" {
		t.Errorf("error = %v, want bot_suspected", raw["error"])
	}
	// The score still rides along with the rejection: an integrator has to be
	// able to tell why the request was refused.
	if raw["his_bot_score"] != 0.9 {
		t.Errorf("his_bot_score = %v, want 0.9", raw["his_bot_score"])
	}

	stats, _ := models.GetKeyStats(db, key.ID, 1)
	if len(stats) == 0 {
		t.Fatal("expected a stat row")
	}
	if stats[0].VerificationsFail != 1 {
		t.Errorf("verifications_fail = %d, want 1", stats[0].VerificationsFail)
	}
	if stats[0].VerificationsOK != 0 {
		t.Errorf("verifications_ok = %d, want 0", stats[0].VerificationsOK)
	}
}

// A rejected attempt must not be retryable. Enforcement runs after the
// challenge is consumed, so replaying the same solved payload cannot buy a
// second roll of the dice with friendlier signals attached.
func TestHISEnforce_RejectedAttemptConsumesTheChallenge(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Blocking", "", 100, 60, "SHA-256")
	if err := models.UpdateAPIKeyFields(db, key.ID, map[string]any{"his_enforce": true}); err != nil {
		t.Fatalf("enable enforcement: %v", err)
	}

	body := solvedBodyWith(t, key, botSignals())
	if got := postVerify(t, router, key, body)["error"]; got != "bot_suspected" {
		t.Fatalf("first attempt: error = %v, want bot_suspected", got)
	}

	// Same solved payload, now presented with impeccably human signals.
	var parsed map[string]interface{}
	json.Unmarshal(body, &parsed)
	parsed["his_signals"] = humanSignals()
	retry, _ := json.Marshal(parsed)

	raw := postVerify(t, router, key, retry)
	if raw["ok"] == true {
		t.Fatal("a rejected payload must not be replayable")
	}
	if raw["error"] != "already_used" {
		t.Errorf("error = %v, want already_used", raw["error"])
	}
}

// A site that ships no collector sends no signals, is scored not at all, and
// passes whatever the switch says. This is what bounds the blast radius of
// turning enforcement on.
func TestHISEnforce_IgnoresRequestsCarryingNoSignals(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Blocking", "", 100, 60, "SHA-256")
	if err := models.UpdateAPIKeyFields(db, key.ID, map[string]any{"his_enforce": true}); err != nil {
		t.Fatalf("enable enforcement: %v", err)
	}

	raw := postVerify(t, router, key, solvedBodyWith(t, key, nil))
	if raw["ok"] != true {
		t.Fatalf("ok = %v, want true with no signals (error %v)", raw["ok"], raw["error"])
	}
	if _, present := raw["his_bot_score"]; present {
		t.Errorf("no score should be reported, got %v", raw["his_bot_score"])
	}
}

// The no-evidence tier passes on the default threshold and is caught once the
// operator lowers it. This is precisely the choice #149 turned on: 135 of 227
// samples had this exact shape, and only the site owner knows what they were.
func TestHISEnforce_ThresholdDecidesTheNoEvidenceTier(t *testing.T) {
	for _, tc := range []struct {
		name      string
		threshold float64
		wantOK    bool
	}{
		{"default lets the no-evidence tier through", models.DefaultHISThreshold, true},
		{"lowered to 0.5 catches it", 0.5, false},
		{"exactly at the score still counts as meeting it", 0.7, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router, db := setupTestRouter(t)
			key, _ := models.CreateAPIKey(db, "Threshold", "", 100, 60, "SHA-256")
			if err := models.UpdateAPIKeyFields(db, key.ID, map[string]any{
				"his_enforce": true, "his_threshold": tc.threshold,
			}); err != nil {
				t.Fatalf("configure key: %v", err)
			}

			raw := postVerify(t, router, key, solvedBodyWith(t, key, noEvidenceSignals()))
			if raw["his_bot_score"] != 0.7 {
				t.Fatalf("his_bot_score = %v, want 0.7", raw["his_bot_score"])
			}
			if raw["ok"] != tc.wantOK {
				t.Fatalf("ok = %v, want %v (error %v)", raw["ok"], tc.wantOK, raw["error"])
			}
			// The reported flag has to agree with the threshold the key applies,
			// otherwise an integrator branching on it disagrees with the server.
			if raw["his_bot_suspected"] == tc.wantOK {
				t.Errorf("his_bot_suspected = %v, should be the inverse of ok here", raw["his_bot_suspected"])
			}
		})
	}
}

// A submission that failed the proof of work is not reported as a bot: the
// maths failure keeps precedence, so a broken widget is never misread as an
// automation wave.
func TestHISEnforce_PoWFailureKeepsPrecedence(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Blocking", "", 100, 60, "SHA-256")
	if err := models.UpdateAPIKeyFields(db, key.ID, map[string]any{"his_enforce": true}); err != nil {
		t.Fatalf("enable enforcement: %v", err)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"payload":     "not-valid-base64!!",
		"his_signals": botSignals(),
	})
	raw := postVerify(t, router, key, body)
	if raw["ok"] == true {
		t.Fatal("expected rejection")
	}
	if raw["error"] == "bot_suspected" {
		t.Error("a proof-of-work failure must not be reported as bot_suspected")
	}
}

// The two defaults live in different packages on purpose, so that models keeps
// no dependency on the scoring package. They are not allowed to drift.
func TestDefaultHISThresholdMatchesScoringPackage(t *testing.T) {
	if models.DefaultHISThreshold != his.BotSuspectThreshold {
		t.Errorf("models.DefaultHISThreshold = %v, his.BotSuspectThreshold = %v",
			models.DefaultHISThreshold, his.BotSuspectThreshold)
	}
}

func TestSuspectThreshold_FallsBackOnOutOfRangeValues(t *testing.T) {
	for _, tc := range []struct {
		stored float64
		want   float64
	}{
		{0, models.DefaultHISThreshold},   // column added before the field existed
		{-1, models.DefaultHISThreshold},  // nonsense write
		{1.5, models.DefaultHISThreshold}, // above any reachable score
		{0.5, 0.5},                        // the value asked for in #149
		{1, 1},                            // only a perfect score blocks
	} {
		key := &models.APIKey{HISThreshold: tc.stored}
		if got := key.SuspectThreshold(); got != tc.want {
			t.Errorf("stored %v: SuspectThreshold() = %v, want %v", tc.stored, got, tc.want)
		}
	}
}

// An out-of-range threshold is refused rather than clamped: a threshold quietly
// corrected to something else is a quietly different blocking policy, and the
// caller would never learn which one they got. 0 matters most, since every
// score is >= 0 and a stored 0 on an enforcing key rejects all its traffic.
func TestUpdateKey_RejectsOutOfRangeThreshold(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Threshold", "", 100, 60, "SHA-256")
	token := getAdminToken(t)

	for _, bad := range []float64{0, -0.5, 1.01, 42} {
		body, _ := json.Marshal(map[string]interface{}{"his_threshold": bad})
		req := httptest.NewRequest("PUT", "/api/admin/keys/"+strconv.FormatInt(key.ID, 10), bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("threshold %v: status = %d, want 400", bad, w.Code)
		}
	}

	// The stored value is untouched by the refused writes.
	after, _ := models.GetAPIKeyByID(db, key.ID)
	if after.SuspectThreshold() != models.DefaultHISThreshold {
		t.Errorf("threshold = %v, want the default left in place", after.SuspectThreshold())
	}
}

func TestUpdateKey_AcceptsAValidThreshold(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Threshold", "", 100, 60, "SHA-256")

	body, _ := json.Marshal(map[string]interface{}{"his_threshold": 0.5, "his_enforce": true})
	req := httptest.NewRequest("PUT", "/api/admin/keys/"+strconv.FormatInt(key.ID, 10), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+getAdminToken(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	after, _ := models.GetAPIKeyByID(db, key.ID)
	if after.HISThreshold != 0.5 || !after.HISEnforce {
		t.Errorf("stored key = threshold %v enforce %v, want 0.5 / true", after.HISThreshold, after.HISEnforce)
	}
}

// Editing an unrelated setting must not silently reset the threshold: the
// update path reads the existing value through the same fallback the scorer
// uses, so a key left at the default keeps it.
func TestUpdateKey_PreservesThresholdOnUnrelatedEdit(t *testing.T) {
	router, db := setupTestRouter(t)
	key, _ := models.CreateAPIKey(db, "Threshold", "", 100, 60, "SHA-256")
	if err := models.UpdateAPIKeyFields(db, key.ID, map[string]any{"his_threshold": 0.55}); err != nil {
		t.Fatalf("seed threshold: %v", err)
	}

	body, _ := json.Marshal(map[string]interface{}{"name": "Renamed"})
	req := httptest.NewRequest("PUT", "/api/admin/keys/"+strconv.FormatInt(key.ID, 10), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+getAdminToken(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	after, _ := models.GetAPIKeyByID(db, key.ID)
	if after.HISThreshold != 0.55 {
		t.Errorf("threshold = %v, want 0.55 preserved across a rename", after.HISThreshold)
	}
}
