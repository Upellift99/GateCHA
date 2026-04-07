package database

import "gorm.io/gorm"

// RunMigrations creates or updates the schema for the given model types.
// It is safe to call multiple times (AutoMigrate is idempotent for additions).
//
// Pass all application model types as modelList, e.g.:
//
//	database.RunMigrations(db, &models.AdminUser{}, &models.APIKey{}, ...)
func RunMigrations(db *gorm.DB, modelList ...any) error {
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
