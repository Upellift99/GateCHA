package testutil

import (
	"testing"

	"github.com/Upellift99/GateCHA/internal/database"
	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database with all migrations applied.
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(
		sqlite.Open(":memory:?_pragma=foreign_keys(1)"),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	// Match production: database.Open caps SQLite at a single connection. Without
	// this, a second connection to ":memory:" opens a fresh empty database, so any
	// test that touches the DB concurrently fails with "no such table".
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	if err := database.RunMigrations(db, models.All()...); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	return db
}
