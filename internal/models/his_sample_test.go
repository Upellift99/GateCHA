package models

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newHISSampleTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&HISSample{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestPruneHISSamples(t *testing.T) {
	db := newHISSampleTestDB(t)
	now := time.Now().UTC()

	// One old sample (40 days), one recent (1 day).
	if err := db.Create(&HISSample{APIKeyID: 1, CreatedAt: now.AddDate(0, 0, -40), Score: 0.5}).Error; err != nil {
		t.Fatalf("insert old: %v", err)
	}
	if err := db.Create(&HISSample{APIKeyID: 1, CreatedAt: now.AddDate(0, 0, -1), Score: 0.5}).Error; err != nil {
		t.Fatalf("insert recent: %v", err)
	}

	pruned, err := PruneHISSamples(db, now.AddDate(0, 0, -30))
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if pruned != 1 {
		t.Errorf("pruned = %d, want 1", pruned)
	}

	var remaining int64
	db.Model(&HISSample{}).Count(&remaining)
	if remaining != 1 {
		t.Errorf("remaining = %d, want 1", remaining)
	}
}

func TestGetHISCalibration(t *testing.T) {
	db := newHISSampleTestDB(t)

	samples := []HISSample{
		// key 1 — varied scores and motion.
		{APIKeyID: 1, Score: 0.0, BotSuspected: false, DurationMs: 5000, PointerEvents: 40}, // bucket 0, motion
		{APIKeyID: 1, Score: 0.85, BotSuspected: true, DurationMs: 100, PointerEvents: 0},   // bucket 8, no motion
		{APIKeyID: 1, Score: 0.9, BotSuspected: true, DurationMs: 100, PointerEvents: 0},    // bucket 9, no motion
		{APIKeyID: 1, Score: 1.0, BotSuspected: true, DurationMs: 50, PointerEvents: 0},     // bucket 9 (clamped), no motion
		// key 2 — must not leak into a key-1 scoped query.
		{APIKeyID: 2, Score: 0.2, BotSuspected: false, DurationMs: 8000, PointerEvents: 100},
	}
	for i := range samples {
		if err := CreateHISSample(db, &samples[i]); err != nil {
			t.Fatalf("create sample %d: %v", i, err)
		}
	}

	id := int64(1)
	cal, err := GetHISCalibration(db, &id, 30, 0.8)
	if err != nil {
		t.Fatalf("GetHISCalibration: %v", err)
	}

	if cal.Samples != 4 {
		t.Errorf("samples = %d, want 4", cal.Samples)
	}
	if cal.Suspected != 3 {
		t.Errorf("suspected = %d, want 3", cal.Suspected)
	}
	if cal.Threshold != 0.8 {
		t.Errorf("threshold = %v, want 0.8", cal.Threshold)
	}
	if len(cal.ScoreHistogram) != 10 {
		t.Fatalf("histogram len = %d, want 10", len(cal.ScoreHistogram))
	}
	if cal.ScoreHistogram[0].Count != 1 {
		t.Errorf("bucket0 = %d, want 1", cal.ScoreHistogram[0].Count)
	}
	if cal.ScoreHistogram[8].Count != 1 {
		t.Errorf("bucket8 = %d, want 1", cal.ScoreHistogram[8].Count)
	}
	if cal.ScoreHistogram[9].Count != 2 { // 0.9 and clamped 1.0
		t.Errorf("bucket9 = %d, want 2", cal.ScoreHistogram[9].Count)
	}
	// 3 of 4 key-1 samples have zero motion.
	if cal.NoMotionPct < 74 || cal.NoMotionPct > 76 {
		t.Errorf("no_motion_pct = %v, want ~75", cal.NoMotionPct)
	}

	// nil key spans all keys (5 samples).
	all, err := GetHISCalibration(db, nil, 30, 0.8)
	if err != nil {
		t.Fatalf("GetHISCalibration all: %v", err)
	}
	if all.Samples != 5 {
		t.Errorf("all samples = %d, want 5", all.Samples)
	}
}
