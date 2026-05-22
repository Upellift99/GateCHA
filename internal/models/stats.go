package models

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const dateFormatYMD = "2006-01-02"

// DailyStat holds per-key per-day counters.
type DailyStat struct {
	ID                int64  `gorm:"primaryKey;autoIncrement" json:"-"`
	APIKeyID          int64  `gorm:"not null;uniqueIndex:idx_key_date" json:"api_key_id"`
	Date              string `gorm:"not null;uniqueIndex:idx_key_date;size:10" json:"date"`
	ChallengesIssued  int    `gorm:"not null;default:0" json:"challenges_issued"`
	VerificationsOK   int    `gorm:"not null;default:0" json:"verifications_ok"`
	VerificationsFail int    `gorm:"not null;default:0" json:"verifications_fail"`
}

// StatsOverview holds aggregated statistics for the dashboard.
type StatsOverview struct {
	TotalChallenges        int         `json:"total_challenges"`
	TotalVerificationsOK   int         `json:"total_verifications_ok"`
	TotalVerificationsFail int         `json:"total_verifications_fail"`
	ActiveKeys             int         `json:"active_keys"`
	Daily                  []DailyStat `json:"daily"`
}

// KeyStatsSummary holds all-time totals for a single API key.
type KeyStatsSummary struct {
	APIKeyID          int64  `json:"api_key_id"`
	ChallengesIssued  int    `json:"challenges_issued"`
	VerificationsOK   int    `json:"verifications_ok"`
	VerificationsFail int    `json:"verifications_fail"`
	LastUsedAt        string `json:"last_used_at"`
}

func IncrementChallengesIssued(db *gorm.DB, apiKeyID int64) error {
	date := time.Now().UTC().Format(dateFormatYMD)
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "api_key_id"}, {Name: "date"}},
		DoUpdates: clause.Assignments(map[string]any{
			"challenges_issued": gorm.Expr("challenges_issued + 1"),
		}),
	}).Create(&DailyStat{APIKeyID: apiKeyID, Date: date, ChallengesIssued: 1}).Error
}

func IncrementVerificationsOK(db *gorm.DB, apiKeyID int64) error {
	date := time.Now().UTC().Format(dateFormatYMD)
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "api_key_id"}, {Name: "date"}},
		DoUpdates: clause.Assignments(map[string]any{
			"verifications_ok": gorm.Expr("verifications_ok + 1"),
		}),
	}).Create(&DailyStat{APIKeyID: apiKeyID, Date: date, VerificationsOK: 1}).Error
}

func IncrementVerificationsFail(db *gorm.DB, apiKeyID int64) error {
	date := time.Now().UTC().Format(dateFormatYMD)
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "api_key_id"}, {Name: "date"}},
		DoUpdates: clause.Assignments(map[string]any{
			"verifications_fail": gorm.Expr("verifications_fail + 1"),
		}),
	}).Create(&DailyStat{APIKeyID: apiKeyID, Date: date, VerificationsFail: 1}).Error
}

func GetStatsOverview(db *gorm.DB, days int) (*StatsOverview, error) {
	overview := &StatsOverview{}

	// Total counters across all time
	row := db.Model(&DailyStat{}).
		Select("COALESCE(SUM(challenges_issued), 0), COALESCE(SUM(verifications_ok), 0), COALESCE(SUM(verifications_fail), 0)").
		Row()
	if err := row.Scan(&overview.TotalChallenges, &overview.TotalVerificationsOK, &overview.TotalVerificationsFail); err != nil {
		return nil, err
	}

	// Active key count
	var activeKeys int64
	if err := db.Model(&APIKey{}).Where("enabled = ?", true).Count(&activeKeys).Error; err != nil {
		return nil, err
	}
	overview.ActiveKeys = int(activeKeys)

	// Daily breakdown for the requested window
	since := time.Now().UTC().AddDate(0, 0, -days).Format(dateFormatYMD)
	if err := db.Model(&DailyStat{}).
		Select("date, "+
			"COALESCE(SUM(challenges_issued), 0) AS challenges_issued, "+
			"COALESCE(SUM(verifications_ok), 0) AS verifications_ok, "+
			"COALESCE(SUM(verifications_fail), 0) AS verifications_fail").
		Where("date >= ?", since).
		Group("date").
		Order("date DESC").
		Scan(&overview.Daily).Error; err != nil {
		return nil, err
	}

	return overview, nil
}

// GetAllKeysStatsSummary returns all-time totals grouped by API key ID.
func GetAllKeysStatsSummary(db *gorm.DB) (map[int64]KeyStatsSummary, error) {
	var rows []KeyStatsSummary
	err := db.Model(&DailyStat{}).
		Select("api_key_id, " +
			"COALESCE(SUM(challenges_issued), 0) AS challenges_issued, " +
			"COALESCE(SUM(verifications_ok), 0) AS verifications_ok, " +
			"COALESCE(SUM(verifications_fail), 0) AS verifications_fail, " +
			"COALESCE(MAX(date), '') AS last_used_at").
		Group("api_key_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]KeyStatsSummary, len(rows))
	for _, s := range rows {
		result[s.APIKeyID] = s
	}
	return result, nil
}

func GetKeyStats(db *gorm.DB, apiKeyID int64, days int) ([]DailyStat, error) {
	since := time.Now().UTC().AddDate(0, 0, -days).Format(dateFormatYMD)
	var stats []DailyStat
	err := db.Where("api_key_id = ? AND date >= ?", apiKeyID, since).
		Order("date DESC").
		Find(&stats).Error
	return stats, err
}
