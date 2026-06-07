package database

import "gorm.io/gorm"

// RunMigrations creates or updates the schema for the given model types.
// It is safe to call multiple times (AutoMigrate is idempotent for additions).
//
// Pass all application model types as modelList, e.g.:
//
//	database.RunMigrations(db, &models.AdminUser{}, &models.APIKey{}, ...)
func RunMigrations(db *gorm.DB, modelList ...any) error {
	// SQLite's AutoMigrate rebuilds a changed table by DROP-ing the original.
	// With foreign keys enabled, DROP TABLE performs an implicit DELETE of the
	// table's rows, which fires ON DELETE CASCADE on referencing tables — so a
	// rebuild of api_keys would wipe every daily_stats row. Disable foreign-key
	// enforcement for the duration of the migration to make rebuilds safe.
	// (PRAGMA is SQLite-only; the single-connection pool ensures it sticks.)
	if db.Dialector.Name() == "sqlite" {
		if err := db.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
			return err
		}
		defer db.Exec("PRAGMA foreign_keys = ON")
	}

	if err := db.AutoMigrate(modelList...); err != nil {
		return err
	}

	// Fix NULL counter values left by an older bug.
	return db.Exec(`
		UPDATE daily_stats SET
			challenges_issued  = COALESCE(challenges_issued, 0),
			verifications_ok   = COALESCE(verifications_ok, 0),
			verifications_fail = COALESCE(verifications_fail, 0)
		WHERE challenges_issued IS NULL
		   OR verifications_ok  IS NULL
		   OR verifications_fail IS NULL
	`).Error
}
