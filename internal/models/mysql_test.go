package models_test

import (
	"testing"
	"time"

	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
)

// TestMySQL_CreateAndGetAPIKey verifies basic CRUD works against a real MySQL instance.
func TestMySQL_CreateAndGetAPIKey(t *testing.T) {
	db := testutil.SetupTestMySQL(t)

	key, err := models.CreateAPIKey(db, "MySQL Test", "example.com", 50000, 600, "SHA-256")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}
	if key.Name != "MySQL Test" {
		t.Errorf("expected name 'MySQL Test', got %q", key.Name)
	}

	found, err := models.GetAPIKeyByKeyID(db, key.KeyID)
	if err != nil {
		t.Fatalf("GetAPIKeyByKeyID failed: %v", err)
	}
	if found.ID != key.ID {
		t.Errorf("expected ID %d, got %d", key.ID, found.ID)
	}
}

// TestMySQL_UpsertCounters verifies that ON DUPLICATE KEY UPDATE increments
// counters atomically rather than inserting duplicate rows.
func TestMySQL_UpsertCounters(t *testing.T) {
	db := testutil.SetupTestMySQL(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	for i := 0; i < 3; i++ {
		if err := models.IncrementChallengesIssued(db, key.ID); err != nil {
			t.Fatalf("IncrementChallengesIssued failed: %v", err)
		}
	}
	if err := models.IncrementVerificationsOK(db, key.ID); err != nil {
		t.Fatalf("IncrementVerificationsOK failed: %v", err)
	}
	if err := models.IncrementVerificationsFail(db, key.ID); err != nil {
		t.Fatalf("IncrementVerificationsFail failed: %v", err)
	}

	stats, err := models.GetKeyStats(db, key.ID, 1)
	if err != nil {
		t.Fatalf("GetKeyStats failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 row (upsert), got %d — duplicate rows indicate ON DUPLICATE KEY UPDATE is broken", len(stats))
	}
	if stats[0].ChallengesIssued != 3 {
		t.Errorf("expected 3, got %d", stats[0].ChallengesIssued)
	}
	if stats[0].VerificationsOK != 1 {
		t.Errorf("expected 1, got %d", stats[0].VerificationsOK)
	}
	if stats[0].VerificationsFail != 1 {
		t.Errorf("expected 1, got %d", stats[0].VerificationsFail)
	}
}

// TestMySQL_InsertIgnore_MarkConsumed verifies that INSERT IGNORE suppresses
// duplicate-key errors on the consumed_challenges unique index.
func TestMySQL_InsertIgnore_MarkConsumed(t *testing.T) {
	db := testutil.SetupTestMySQL(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	expiresAt := time.Now().Add(5 * time.Minute)
	if err := models.MarkConsumed(db, "mysql-hash", key.ID, expiresAt); err != nil {
		t.Fatalf("MarkConsumed failed: %v", err)
	}

	// Second insert on the same challenge hash must not return an error.
	if err := models.MarkConsumed(db, "mysql-hash", key.ID, expiresAt); err != nil {
		t.Fatalf("duplicate MarkConsumed should not error on MySQL: %v", err)
	}

	consumed, err := models.IsConsumed(db, "mysql-hash")
	if err != nil {
		t.Fatalf("IsConsumed failed: %v", err)
	}
	if !consumed {
		t.Error("expected challenge to be consumed")
	}
}

// TestMySQL_ReservedWord_Setting verifies that the "key" column (a reserved word
// in MySQL) is correctly quoted by GORM's map-form Where clause.
func TestMySQL_ReservedWord_Setting(t *testing.T) {
	db := testutil.SetupTestMySQL(t)

	if err := models.SetSetting(db, "key", "reserved-word-value"); err != nil {
		t.Fatalf("SetSetting with reserved column name 'key' failed: %v", err)
	}

	val, err := models.GetSetting(db, "key")
	if err != nil {
		t.Fatalf("GetSetting with reserved column name 'key' failed: %v", err)
	}
	if val != "reserved-word-value" {
		t.Errorf("expected 'reserved-word-value', got %q", val)
	}

	// Upsert — existing row should be updated, not duplicated.
	if err := models.SetSetting(db, "key", "updated-value"); err != nil {
		t.Fatalf("SetSetting upsert failed: %v", err)
	}
	val, _ = models.GetSetting(db, "key")
	if val != "updated-value" {
		t.Errorf("expected 'updated-value', got %q", val)
	}
}

// TestMySQL_CleanupExpired verifies that expired challenges are deleted and
// valid ones retained using native MySQL DATETIME comparison.
func TestMySQL_CleanupExpired(t *testing.T) {
	db := testutil.SetupTestMySQL(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)
	models.MarkConsumed(db, "expired", key.ID, past)
	models.MarkConsumed(db, "valid", key.ID, future)

	deleted, err := models.CleanupExpired(db)
	if err != nil {
		t.Fatalf("CleanupExpired failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	consumed, _ := models.IsConsumed(db, "expired")
	if consumed {
		t.Error("expected expired challenge to be cleaned up")
	}
	consumed, _ = models.IsConsumed(db, "valid")
	if !consumed {
		t.Error("expected valid challenge to remain")
	}
}

// TestMySQL_HISCalibration guards the score histogram against dialect-specific
// SQL. The bucketing expression used to be `CAST(score * 10 AS INTEGER)`, which
// SQLite accepts and MySQL rejects as a syntax error, so the calibration
// endpoint returned 500 on every MySQL-backed instance while the SQLite-only
// unit test stayed green.
func TestMySQL_HISCalibration(t *testing.T) {
	db := testutil.SetupTestMySQL(t)
	key, _ := models.CreateAPIKey(db, "Calibration", "", 0, 0, "")

	samples := []models.HISSample{
		{APIKeyID: key.ID, Score: 0.0, BotSuspected: false, DurationMs: 5000, PointerEvents: 40},
		{APIKeyID: key.ID, Score: 0.85, BotSuspected: true, DurationMs: 100},
		{APIKeyID: key.ID, Score: 0.9, BotSuspected: true, DurationMs: 100},
		{APIKeyID: key.ID, Score: 1.0, BotSuspected: true, DurationMs: 50},
	}
	for i := range samples {
		if err := models.CreateHISSample(db, &samples[i]); err != nil {
			t.Fatalf("CreateHISSample %d failed: %v", i, err)
		}
	}

	cal, err := models.GetHISCalibration(db, &key.ID, 30, 0.8)
	if err != nil {
		t.Fatalf("GetHISCalibration failed: %v", err)
	}
	if cal.Samples != 4 || cal.Suspected != 3 {
		t.Errorf("expected 4 samples / 3 suspected, got %d / %d", cal.Samples, cal.Suspected)
	}
	if len(cal.ScoreHistogram) != 10 {
		t.Fatalf("expected 10 buckets, got %d", len(cal.ScoreHistogram))
	}
	if cal.ScoreHistogram[0].Count != 1 {
		t.Errorf("expected bucket0 == 1, got %d", cal.ScoreHistogram[0].Count)
	}
	if cal.ScoreHistogram[8].Count != 1 {
		t.Errorf("expected bucket8 == 1, got %d", cal.ScoreHistogram[8].Count)
	}
	// 0.9 plus the clamped 1.0.
	if cal.ScoreHistogram[9].Count != 2 {
		t.Errorf("expected bucket9 == 2, got %d", cal.ScoreHistogram[9].Count)
	}
	// 3 of 4 samples carry no pointer, scroll or touch event.
	if cal.NoMotionPct < 74 || cal.NoMotionPct > 76 {
		t.Errorf("expected no_motion_pct ~75, got %v", cal.NoMotionPct)
	}

	// A key with no samples must still produce a full result, not an error:
	// that is the state every instance is in right after enabling sampling.
	other, _ := models.CreateAPIKey(db, "No samples", "", 0, 0, "")
	empty, err := models.GetHISCalibration(db, &other.ID, 30, 0.8)
	if err != nil {
		t.Fatalf("GetHISCalibration with no samples failed: %v", err)
	}
	if empty.Samples != 0 || len(empty.ScoreHistogram) != 10 {
		t.Errorf("expected 0 samples and 10 buckets, got %d / %d", empty.Samples, len(empty.ScoreHistogram))
	}
}
