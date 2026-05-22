package models

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ConsumedChallenge tracks used challenge tokens to prevent replay attacks.
type ConsumedChallenge struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	Challenge  string    `gorm:"not null;uniqueIndex;size:256"`
	APIKeyID   int64     `gorm:"not null;index"`
	ExpiresAt  time.Time `gorm:"not null"`
	ConsumedAt time.Time `gorm:"not null;autoCreateTime"`
}

func IsConsumed(db *gorm.DB, challenge string) (bool, error) {
	var count int64
	err := db.Model(&ConsumedChallenge{}).Where("challenge = ?", challenge).Count(&count).Error
	return count > 0, err
}

func MarkConsumed(db *gorm.DB, challenge string, apiKeyID int64, expiresAt time.Time) error {
	// DoNothing: true translates to INSERT OR IGNORE (SQLite) / INSERT IGNORE (MySQL)
	return db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&ConsumedChallenge{
			Challenge: challenge,
			APIKeyID:  apiKeyID,
			ExpiresAt: expiresAt,
		}).Error
}

func CleanupExpired(db *gorm.DB) (int64, error) {
	result := db.Where("expires_at < ?", time.Now().UTC()).Delete(&ConsumedChallenge{})
	return result.RowsAffected, result.Error
}
