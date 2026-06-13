package models

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DailyCountryStat holds per-key, per-day, per-country verification counters.
// Country is an ISO 3166-1 alpha-2 code resolved from the client IP at request
// time; the raw IP is never stored (privacy-first). An empty Country means the
// source could not be geolocated (private/loopback/unknown) and is bucketed as
// "unknown" so per-country totals reconcile with the overall verification ones.
type DailyCountryStat struct {
	ID                int64  `gorm:"primaryKey;autoIncrement" json:"-"`
	APIKeyID          int64  `gorm:"not null;uniqueIndex:idx_key_date_country" json:"api_key_id"`
	Date              string `gorm:"not null;uniqueIndex:idx_key_date_country;size:10" json:"date"`
	Country           string `gorm:"uniqueIndex:idx_key_date_country;size:2" json:"country"`
	VerificationsOK   int    `gorm:"not null;default:0" json:"verifications_ok"`
	VerificationsFail int    `gorm:"not null;default:0" json:"verifications_fail"`
}

// CountryStat is an aggregated per-country row for the dashboard.
type CountryStat struct {
	Country           string `json:"country"`
	VerificationsOK   int    `json:"verifications_ok"`
	VerificationsFail int    `json:"verifications_fail"`
	Total             int    `json:"total"`
}

// IncrementCountryVerification records one verification outcome for the
// key/day/country, atomically bumping the matching counter via UPSERT. country
// may be empty (unknown source); it is stored as-is so totals reconcile.
func IncrementCountryVerification(db *gorm.DB, apiKeyID int64, country string, ok bool) error {
	date := time.Now().UTC().Format(dateFormatYMD)
	col := "verifications_fail"
	row := &DailyCountryStat{APIKeyID: apiKeyID, Date: date, Country: country, VerificationsFail: 1}
	if ok {
		col = "verifications_ok"
		row.VerificationsFail = 0
		row.VerificationsOK = 1
	}
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "api_key_id"}, {Name: "date"}, {Name: "country"}},
		DoUpdates: clause.Assignments(map[string]any{
			col: gorm.Expr(col + " + 1"),
		}),
	}).Create(row).Error
}

// GetCountryStats returns per-country verification totals over the last `days`,
// ordered by total descending and capped at `limit`. When apiKeyID is nil the
// totals span every key; otherwise they are scoped to that one key.
func GetCountryStats(db *gorm.DB, apiKeyID *int64, days, limit int) ([]CountryStat, error) {
	since := time.Now().UTC().AddDate(0, 0, -days).Format(dateFormatYMD)
	q := db.Model(&DailyCountryStat{}).
		Select("country, "+
			"COALESCE(SUM(verifications_ok), 0) AS verifications_ok, "+
			"COALESCE(SUM(verifications_fail), 0) AS verifications_fail, "+
			"COALESCE(SUM(verifications_ok + verifications_fail), 0) AS total").
		Where("date >= ?", since)
	if apiKeyID != nil {
		q = q.Where("api_key_id = ?", *apiKeyID)
	}
	var rows []CountryStat
	err := q.Group("country").Order("total DESC").Limit(limit).Scan(&rows).Error
	if rows == nil {
		rows = []CountryStat{}
	}
	return rows, err
}
