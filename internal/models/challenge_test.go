package models_test

import (
	"testing"
	"time"

	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
)

func TestIsConsumed_NotConsumed(t *testing.T) {
	db := testutil.SetupTestDB(t)

	consumed, err := models.IsConsumed(db, "test-challenge-hash")
	if err != nil {
		t.Fatalf("IsConsumed failed: %v", err)
	}
	if consumed {
		t.Error("expected not consumed")
	}
}

func TestMarkConsumed(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	expiresAt := time.Now().Add(5 * time.Minute)
	if err := models.MarkConsumed(db, "test-hash", key.ID, expiresAt); err != nil {
		t.Fatalf("MarkConsumed failed: %v", err)
	}

	consumed, err := models.IsConsumed(db, "test-hash")
	if err != nil {
		t.Fatalf("IsConsumed failed: %v", err)
	}
	if !consumed {
		t.Error("expected consumed after marking")
	}
}

func TestMarkConsumed_Duplicate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	expiresAt := time.Now().Add(5 * time.Minute)
	models.MarkConsumed(db, "dup-hash", key.ID, expiresAt)

	// Second insert should be ignored (INSERT OR IGNORE)
	err := models.MarkConsumed(db, "dup-hash", key.ID, expiresAt)
	if err != nil {
		t.Fatalf("duplicate MarkConsumed should not error: %v", err)
	}
}

func TestCleanupExpired(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	// Insert an expired challenge
	past := time.Now().Add(-1 * time.Hour)
	models.MarkConsumed(db, "expired-hash", key.ID, past)

	// Insert a valid challenge
	future := time.Now().Add(1 * time.Hour)
	models.MarkConsumed(db, "valid-hash", key.ID, future)

	deleted, err := models.CleanupExpired(db)
	if err != nil {
		t.Fatalf("CleanupExpired failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	consumed, _ := models.IsConsumed(db, "expired-hash")
	if consumed {
		t.Error("expected expired challenge to be cleaned up")
	}

	consumed, _ = models.IsConsumed(db, "valid-hash")
	if !consumed {
		t.Error("expected valid challenge to remain")
	}
}
