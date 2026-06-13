package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Upellift99/GateCHA/internal/his"
	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
)

func TestRecordHISMonitor_SamplingOptIn(t *testing.T) {
	db := testutil.SetupTestDB(t)

	sampled := &models.APIKey{KeyID: "gk_sampled", HMACSecret: "x", HISSampling: true, Enabled: true}
	plain := &models.APIKey{KeyID: "gk_plain", HMACSecret: "x", HISSampling: false, Enabled: true}
	db.Create(sampled)
	db.Create(plain)

	signals := &his.Signals{DurationMs: 50, TimeToFirstMs: -1, PointerEvents: 0}

	recordHISMonitor(db, sampled, signals)
	recordHISMonitor(db, plain, signals)
	recordHISMonitor(db, nil, signals) // must be a safe no-op
	recordHISMonitor(db, sampled, nil) // nil signals: no-op

	var sampledRows, plainRows int64
	db.Model(&models.HISSample{}).Where("api_key_id = ?", sampled.ID).Count(&sampledRows)
	db.Model(&models.HISSample{}).Where("api_key_id = ?", plain.ID).Count(&plainRows)

	if sampledRows != 1 {
		t.Errorf("sampled key rows = %d, want 1", sampledRows)
	}
	if plainRows != 0 {
		t.Errorf("opted-out key rows = %d, want 0", plainRows)
	}

	// Monitor counters bump for every non-nil sample regardless of sampling.
	var sample models.HISSample
	if err := db.Where("api_key_id = ?", sampled.ID).First(&sample).Error; err != nil {
		t.Fatalf("load stored sample: %v", err)
	}
	if sample.DurationMs != 50 || !sample.BotSuspected {
		t.Errorf("stored sample = %+v, want DurationMs 50 and bot-suspected", sample)
	}
}

func TestHISCalibrationEndpoint(t *testing.T) {
	router, db := setupTestRouter(t)

	key := &models.APIKey{KeyID: "gk_cal", HMACSecret: "x", HISSampling: true, Enabled: true}
	db.Create(key)
	db.Create(&models.HISSample{APIKeyID: key.ID, Score: 0.9, BotSuspected: true})

	req := httptest.NewRequest("GET", "/api/admin/his/calibration", nil)
	req.Header.Set("Authorization", "Bearer "+getAdminToken(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var cal models.HISCalibration
	if err := json.NewDecoder(w.Body).Decode(&cal); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if cal.Samples != 1 || cal.Suspected != 1 {
		t.Errorf("cal = %+v, want 1 sample / 1 suspected", cal)
	}
	if len(cal.ScoreHistogram) != 10 || cal.ScoreHistogram[9].Count != 1 {
		t.Errorf("histogram = %+v, want bucket9 == 1", cal.ScoreHistogram)
	}
}
