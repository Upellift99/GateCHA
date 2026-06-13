package models

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newCountryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&DailyCountryStat{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestIncrementCountryVerificationUpserts(t *testing.T) {
	db := newCountryTestDB(t)

	for i := 0; i < 3; i++ {
		if err := IncrementCountryVerification(db, 1, "FR", true); err != nil {
			t.Fatalf("increment ok: %v", err)
		}
	}
	if err := IncrementCountryVerification(db, 1, "FR", false); err != nil {
		t.Fatalf("increment fail: %v", err)
	}
	if err := IncrementCountryVerification(db, 1, "US", true); err != nil {
		t.Fatalf("increment us: %v", err)
	}
	// Unknown source bucketed under empty country.
	if err := IncrementCountryVerification(db, 1, "", false); err != nil {
		t.Fatalf("increment unknown: %v", err)
	}

	var fr DailyCountryStat
	if err := db.Where("api_key_id = ? AND country = ?", 1, "FR").First(&fr).Error; err != nil {
		t.Fatalf("load FR row: %v", err)
	}
	if fr.VerificationsOK != 3 || fr.VerificationsFail != 1 {
		t.Errorf("FR = {ok:%d fail:%d}, want {ok:3 fail:1}", fr.VerificationsOK, fr.VerificationsFail)
	}

	var rows int64
	db.Model(&DailyCountryStat{}).Count(&rows)
	if rows != 3 { // FR, US, "" — one row per (key, date, country)
		t.Errorf("got %d rows, want 3", rows)
	}
}

func TestGetCountryStatsOrdersByTotal(t *testing.T) {
	db := newCountryTestDB(t)

	// FR: 5 total, US: 2 total, DE: 1 total.
	for i := 0; i < 4; i++ {
		_ = IncrementCountryVerification(db, 7, "FR", true)
	}
	_ = IncrementCountryVerification(db, 7, "FR", false)
	_ = IncrementCountryVerification(db, 7, "US", true)
	_ = IncrementCountryVerification(db, 7, "US", false)
	_ = IncrementCountryVerification(db, 7, "DE", true)
	// Another key's data must not leak into the scoped query.
	_ = IncrementCountryVerification(db, 99, "JP", true)

	id := int64(7)
	got, err := GetCountryStats(db, &id, 30, 10)
	if err != nil {
		t.Fatalf("GetCountryStats: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d countries, want 3 (%+v)", len(got), got)
	}
	if got[0].Country != "FR" || got[0].Total != 5 {
		t.Errorf("top = %+v, want FR total 5", got[0])
	}
	if got[0].VerificationsOK != 4 || got[0].VerificationsFail != 1 {
		t.Errorf("FR breakdown = {ok:%d fail:%d}, want {4 1}", got[0].VerificationsOK, got[0].VerificationsFail)
	}

	// Limit is honored.
	limited, err := GetCountryStats(db, &id, 30, 2)
	if err != nil {
		t.Fatalf("GetCountryStats limit: %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("limit=2 returned %d rows", len(limited))
	}

	// nil key spans all keys (FR, US, DE, JP = 4 countries).
	all, err := GetCountryStats(db, nil, 30, 10)
	if err != nil {
		t.Fatalf("GetCountryStats all: %v", err)
	}
	if len(all) != 4 {
		t.Errorf("all-keys returned %d countries, want 4", len(all))
	}
}
