package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Upellift99/GateCHA/internal/models"
)

var testModels = []any{
	&models.AdminUser{},
	&models.APIKey{},
	&models.ConsumedChallenge{},
	&models.DailyStat{},
	&models.Setting{},
}

func TestOpen(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "test.db")

	db, err := Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	if err := RunMigrations(db, testModels...); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Verify tables exist
	tables := []string{"admin_users", "api_keys", "consumed_challenges", "daily_stats", "settings"}
	for _, table := range tables {
		var name string
		if err := db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name).Error; err != nil || name == "" {
			t.Errorf("table %s should exist (err: %v)", table, err)
		}
	}

	// Verify the file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected database file to exist")
	}
}

func TestOpen_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "deep", "nested", "dir", "test.db")

	db, err := Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}

	dir := filepath.Dir(dbPath)
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// Running migrations twice should not fail
	if err := RunMigrations(db, testModels...); err != nil {
		t.Fatalf("RunMigrations (1st call) failed: %v", err)
	}
	if err := RunMigrations(db, testModels...); err != nil {
		t.Fatalf("RunMigrations (2nd call) failed: %v", err)
	}
}

func TestOpen_InvalidDriver(t *testing.T) {
	_, err := Open("postgres", "something")
	if err == nil {
		t.Error("expected error for unsupported driver")
	}
}
