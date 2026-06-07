package database

import (
	"path/filepath"
	"testing"

	"github.com/Upellift99/GateCHA/internal/models"
)

// TestRunMigrations_RestoresForeignKeys verifies RunMigrations re-enables
// foreign-key enforcement after it finishes (it disables it during AutoMigrate
// to keep SQLite table rebuilds from cascade-deleting rows).
func TestRunMigrations_RestoresForeignKeys(t *testing.T) {
	db, err := Open("sqlite", filepath.Join(t.TempDir(), "fk.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := RunMigrations(db,
		&models.AdminUser{}, &models.APIKey{}, &models.ConsumedChallenge{},
		&models.DailyStat{}, &models.Setting{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var fk int
	if err := db.Raw("PRAGMA foreign_keys").Scan(&fk).Error; err != nil {
		t.Fatalf("read pragma: %v", err)
	}
	if fk != 1 {
		t.Fatalf("foreign_keys left disabled after migration (got %d, want 1)", fk)
	}
}

// TestForeignKeysOff_PreventsCascadeOnRebuild locks in the invariant the fix
// relies on: SQLite's AutoMigrate rebuilds a changed table by DROP-ing the
// original, and DROP TABLE performs an implicit DELETE that fires ON DELETE
// CASCADE. With foreign keys disabled — as RunMigrations does around
// AutoMigrate — that rebuild must not delete child rows.
func TestForeignKeysOff_PreventsCascadeOnRebuild(t *testing.T) {
	db, err := Open("sqlite", filepath.Join(t.TempDir(), "cascade.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	seed := []string{
		`CREATE TABLE api_keys (id INTEGER PRIMARY KEY AUTOINCREMENT, key_id TEXT NOT NULL)`,
		`CREATE TABLE daily_stats (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			api_key_id INTEGER NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
			date       TEXT    NOT NULL)`,
		`INSERT INTO api_keys (id, key_id) VALUES (1, 'gk_x')`,
		`INSERT INTO daily_stats (api_key_id, date) VALUES (1, '2026-05-01')`,
	}
	for _, s := range seed {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	// Reproduce the exact rebuild primitive AutoMigrate performs on the parent,
	// bracketed by the foreign-key toggle RunMigrations applies.
	rebuild := []string{
		"PRAGMA foreign_keys = OFF",
		"ALTER TABLE api_keys RENAME TO api_keys__old",
		"CREATE TABLE api_keys (id INTEGER PRIMARY KEY AUTOINCREMENT, key_id TEXT NOT NULL)",
		"INSERT INTO api_keys SELECT * FROM api_keys__old",
		"DROP TABLE api_keys__old",
		"PRAGMA foreign_keys = ON",
	}
	for _, s := range rebuild {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("rebuild: %v", err)
		}
	}

	var stats int64
	db.Model(&models.DailyStat{}).Count(&stats)
	if stats != 1 {
		t.Fatalf("child rows cascade-deleted by parent rebuild: got %d, want 1", stats)
	}
}
