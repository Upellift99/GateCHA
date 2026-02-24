package testutil

import (
	"database/sql"
	"testing"

	"github.com/Upellift99/GateCHA/internal/database"
	_ "modernc.org/sqlite"
)

// SetupTestDB creates an in-memory SQLite database with all migrations applied.
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
