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

	// Drop tables in dependency order (referencing tables first) so no FK
	// constraint violation occurs. Each statement is independent and safe to
	// run even if the table does not yet exist.
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

	if err := database.RunMigrations(db,
		&models.AdminUser{},
		&models.APIKey{},
		&models.ConsumedChallenge{},
		&models.DailyStat{},
		&models.Setting{},
	); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})

	return db
}
