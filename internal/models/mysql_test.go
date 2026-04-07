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
