package testutil

import (
	"os"
	"testing"

	"github.com/Upellift99/GateCHA/internal/database"
	"github.com/Upellift99/GateCHA/internal/models"
	"gorm.io/gorm"
)

// SetupTestMySQL connects to a real MySQL instance for integration testing.
// The test is skipped when GATECHA_TEST_MYSQL_DSN is not set.
//
// Example DSN:
//
//	root:root@tcp(localhost:3306)/gatecha_test?parseTime=true&charset=utf8mb4&loc=UTC
func SetupTestMySQL(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("GATECHA_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("GATECHA_TEST_MYSQL_DSN not set — skipping MySQL integration test")
	}

	db, err := database.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to connect to MySQL: %v", err)
	}

	// Reset the schema between runs. The table list is derived from models.All()
	// rather than written out again: this helper used to carry its own hardcoded
	// list, which had silently drifted and left rows from earlier runs behind on
	// a shared database. models.All() is in dependency order (parents first), so
	// dropping in reverse removes referencing tables before the ones they point
	// at and no foreign key constraint is violated.
	all := models.All()
	for i := len(all) - 1; i >= 0; i-- {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(all[i]); err != nil {
			t.Fatalf("failed to resolve table name for %T: %v", all[i], err)
		}
		// Safe to run even when the table does not exist yet.
		if err := db.Exec("DROP TABLE IF EXISTS `" + stmt.Schema.Table + "`").Error; err != nil {
			t.Fatalf("failed to drop table %s: %v", stmt.Schema.Table, err)
		}
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
