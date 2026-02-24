package models

import (
	"testing"

	"github.com/Upellift99/GateCHA/internal/testutil"
)

func TestIncrementChallengesIssued(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := CreateAPIKey(db, "Test", "", 0, 0, "")

	if err := IncrementChallengesIssued(db, key.ID); err != nil {
		t.Fatalf("IncrementChallengesIssued failed: %v", err)
	}
	if err := IncrementChallengesIssued(db, key.ID); err != nil {
		t.Fatalf("IncrementChallengesIssued (2nd) failed: %v", err)
	}

	stats, err := GetKeyStats(db, key.ID, 1)
	if err != nil {
		t.Fatalf("GetKeyStats failed: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat row, got %d", len(stats))
	}
	if stats[0].ChallengesIssued != 2 {
		t.Errorf("expected 2 challenges, got %d", stats[0].ChallengesIssued)
	}
}

func TestIncrementVerificationsOK(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := CreateAPIKey(db, "Test", "", 0, 0, "")

	if err := IncrementVerificationsOK(db, key.ID); err != nil {
		t.Fatalf("IncrementVerificationsOK failed: %v", err)
	}

	stats, _ := GetKeyStats(db, key.ID, 1)
	if len(stats) == 0 {
		t.Fatal("expected stats")
	}
	if stats[0].VerificationsOK != 1 {
		t.Errorf("expected 1, got %d", stats[0].VerificationsOK)
	}
}

func TestIncrementVerificationsFail(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := CreateAPIKey(db, "Test", "", 0, 0, "")

	if err := IncrementVerificationsFail(db, key.ID); err != nil {
		t.Fatalf("IncrementVerificationsFail failed: %v", err)
	}

	stats, _ := GetKeyStats(db, key.ID, 1)
	if len(stats) == 0 {
		t.Fatal("expected stats")
	}
	if stats[0].VerificationsFail != 1 {
		t.Errorf("expected 1, got %d", stats[0].VerificationsFail)
	}
}

func TestGetStatsOverview(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := CreateAPIKey(db, "Test", "", 0, 0, "")

	IncrementChallengesIssued(db, key.ID)
	IncrementChallengesIssued(db, key.ID)
	IncrementVerificationsOK(db, key.ID)
	IncrementVerificationsFail(db, key.ID)

	overview, err := GetStatsOverview(db, 30)
	if err != nil {
		t.Fatalf("GetStatsOverview failed: %v", err)
	}
	if overview.TotalChallenges != 2 {
		t.Errorf("expected 2 total challenges, got %d", overview.TotalChallenges)
	}
	if overview.TotalVerificationsOK != 1 {
		t.Errorf("expected 1 total OK, got %d", overview.TotalVerificationsOK)
	}
	if overview.TotalVerificationsFail != 1 {
		t.Errorf("expected 1 total fail, got %d", overview.TotalVerificationsFail)
	}
	if overview.ActiveKeys != 1 {
		t.Errorf("expected 1 active key, got %d", overview.ActiveKeys)
	}
	if len(overview.Daily) != 1 {
		t.Errorf("expected 1 daily entry, got %d", len(overview.Daily))
	}
}

func TestGetAllKeysStatsSummary(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key1, _ := CreateAPIKey(db, "Key1", "", 0, 0, "")
	key2, _ := CreateAPIKey(db, "Key2", "", 0, 0, "")

	IncrementChallengesIssued(db, key1.ID)
	IncrementChallengesIssued(db, key2.ID)
	IncrementChallengesIssued(db, key2.ID)

	summary, err := GetAllKeysStatsSummary(db)
	if err != nil {
		t.Fatalf("GetAllKeysStatsSummary failed: %v", err)
	}
	if len(summary) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(summary))
	}
	if summary[key1.ID].ChallengesIssued != 1 {
		t.Errorf("expected 1 for key1, got %d", summary[key1.ID].ChallengesIssued)
	}
	if summary[key2.ID].ChallengesIssued != 2 {
		t.Errorf("expected 2 for key2, got %d", summary[key2.ID].ChallengesIssued)
	}
}

func TestGetKeyStats_Empty(t *testing.T) {
	db := testutil.SetupTestDB(t)
	key, _ := CreateAPIKey(db, "Test", "", 0, 0, "")

	stats, err := GetKeyStats(db, key.ID, 30)
	if err != nil {
		t.Fatalf("GetKeyStats failed: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected empty stats, got %d", len(stats))
	}
}
