package models

import (
	"database/sql"
	"time"
)

func IsConsumed(db *sql.DB, challenge string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM consumed_challenges WHERE challenge = ?`, challenge).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func MarkConsumed(db *sql.DB, challenge string, apiKeyID int64, expiresAt time.Time) error {
	_, err := db.Exec(`
		INSERT OR IGNORE INTO consumed_challenges (challenge, api_key_id, expires_at)
		VALUES (?, ?, ?)
	`, challenge, apiKeyID, expiresAt.UTC().Format(time.RFC3339))
	return err
}

func CleanupExpired(db *sql.DB) (int64, error) {
	result, err := db.Exec(`DELETE FROM consumed_challenges WHERE datetime(expires_at) < datetime('now')`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
