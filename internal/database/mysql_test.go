package database

import (
	"os"
	"testing"

	"gorm.io/gorm"
)

func skipIfNoMySQL(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("GATECHA_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("GATECHA_TEST_MYSQL_DSN not set — skipping MySQL integration test")
	}
	return dsn
}

// dropTablesMySQL drops all app tables in FK-safe order (referencing tables
// first). Each statement is independent so the connection pool is not an issue.
func dropTablesMySQL(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, table := range []string{
		"consumed_challenges", // references api_keys
		"daily_stats",         // references api_keys
		"settings",
		"api_keys",
		"admin_users",
	} {
		if err := db.Exec("DROP TABLE IF EXISTS `" + table + "`").Error; err != nil {
			t.Fatalf("failed to drop table %s: %v", table, err)
		}
	}
}

func TestOpen_MySQL(t *testing.T) {
	dsn := skipIfNoMySQL(t)

	db, err := Open("mysql", dsn)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	dropTablesMySQL(t, db)

	if err := RunMigrations(db, testModels...); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Verify tables exist using information_schema (MySQL-specific)
	tables := []string{"admin_users", "api_keys", "consumed_challenges", "daily_stats", "settings"}
	for _, table := range tables {
		var count int64
		err := db.Raw(
			"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?",
			table,
		).Scan(&count).Error
		if err != nil || count == 0 {
			t.Errorf("table %s should exist (err: %v)", table, err)
		}
	}
}

// TestRunMigrations_MySQL_Idempotent verifies that calling RunMigrations on a
// database where tables already exist does not fail. TestOpen_MySQL runs first
// and creates the schema, so this test exercises the "tables already exist"
// path of AutoMigrate.
func TestRunMigrations_MySQL_Idempotent(t *testing.T) {
	dsn := skipIfNoMySQL(t)

	db, err := Open("mysql", dsn)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	if err := RunMigrations(db, testModels...); err != nil {
		t.Fatalf("RunMigrations on existing schema failed: %v", err)
	}
	if err := RunMigrations(db, testModels...); err != nil {
		t.Fatalf("RunMigrations (2nd call) failed: %v", err)
	}
}
