package models

import (
	"database/sql"
	"fmt"
	"time"
)

const dateFormatYMD = "2006-01-02"

type DailyStat struct {
	Date              string `json:"date"`
	ChallengesIssued  int    `json:"challenges_issued"`
	VerificationsOK   int    `json:"verifications_ok"`
	VerificationsFail int    `json:"verifications_fail"`
}

type StatsOverview struct {
	TotalChallenges        int         `json:"total_challenges"`
	TotalVerificationsOK   int         `json:"total_verifications_ok"`
	TotalVerificationsFail int         `json:"total_verifications_fail"`
	ActiveKeys             int         `json:"active_keys"`
	Daily                  []DailyStat `json:"daily"`
}

func IncrementChallengesIssued(db *sql.DB, apiKeyID int64) error {
	date := time.Now().UTC().Format(dateFormatYMD)
	_, err := db.Exec(`
		INSERT INTO daily_stats (api_key_id, date, challenges_issued, verifications_ok, verifications_fail)
		VALUES (?, ?, 1, 0, 0)
		ON CONFLICT(api_key_id, date)
		DO UPDATE SET challenges_issued = challenges_issued + 1
	`, apiKeyID, date)
	return err
}

func IncrementVerificationsOK(db *sql.DB, apiKeyID int64) error {
	date := time.Now().UTC().Format(dateFormatYMD)
	_, err := db.Exec(`
		INSERT INTO daily_stats (api_key_id, date, challenges_issued, verifications_ok, verifications_fail)
		VALUES (?, ?, 0, 1, 0)
		ON CONFLICT(api_key_id, date)
		DO UPDATE SET verifications_ok = verifications_ok + 1
	`, apiKeyID, date)
	return err
}

func IncrementVerificationsFail(db *sql.DB, apiKeyID int64) error {
	date := time.Now().UTC().Format(dateFormatYMD)
	_, err := db.Exec(`
		INSERT INTO daily_stats (api_key_id, date, challenges_issued, verifications_ok, verifications_fail)
		VALUES (?, ?, 0, 0, 1)
		ON CONFLICT(api_key_id, date)
		DO UPDATE SET verifications_fail = verifications_fail + 1
	`, apiKeyID, date)
	return err
}

func GetStatsOverview(db *sql.DB, days int) (*StatsOverview, error) {
	overview := &StatsOverview{}

	err := db.QueryRow(`
		SELECT COALESCE(SUM(challenges_issued), 0), COALESCE(SUM(verifications_ok), 0), COALESCE(SUM(verifications_fail), 0)
		FROM daily_stats
	`).Scan(&overview.TotalChallenges, &overview.TotalVerificationsOK, &overview.TotalVerificationsFail)
	if err != nil {
		return nil, err
	}

	err = db.QueryRow(`SELECT COUNT(*) FROM api_keys WHERE enabled = 1`).Scan(&overview.ActiveKeys)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`
		SELECT date, COALESCE(SUM(challenges_issued), 0), COALESCE(SUM(verifications_ok), 0), COALESCE(SUM(verifications_fail), 0)
		FROM daily_stats
		WHERE date >= date('now', ?)
		GROUP BY date
		ORDER BY date DESC
	`, fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s DailyStat
		if err := rows.Scan(&s.Date, &s.ChallengesIssued, &s.VerificationsOK, &s.VerificationsFail); err != nil {
			return nil, err
		}
		overview.Daily = append(overview.Daily, s)
	}

	return overview, nil
}

// KeyStatsSummary holds all-time totals for a single API key.
type KeyStatsSummary struct {
	APIKeyID          int64  `json:"api_key_id"`
	ChallengesIssued  int    `json:"challenges_issued"`
	VerificationsOK   int    `json:"verifications_ok"`
	VerificationsFail int    `json:"verifications_fail"`
	LastUsedAt        string `json:"last_used_at"`
}

// GetAllKeysStatsSummary returns all-time totals grouped by API key ID.
func GetAllKeysStatsSummary(db *sql.DB) (map[int64]KeyStatsSummary, error) {
	rows, err := db.Query(`
		SELECT api_key_id,
		       COALESCE(SUM(challenges_issued), 0),
		       COALESCE(SUM(verifications_ok), 0),
		       COALESCE(SUM(verifications_fail), 0),
		       COALESCE(MAX(date), '')
		FROM daily_stats
		GROUP BY api_key_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]KeyStatsSummary)
	for rows.Next() {
		var s KeyStatsSummary
		if err := rows.Scan(&s.APIKeyID, &s.ChallengesIssued, &s.VerificationsOK, &s.VerificationsFail, &s.LastUsedAt); err != nil {
			return nil, err
		}
		result[s.APIKeyID] = s
	}
	return result, nil
}

func GetKeyStats(db *sql.DB, apiKeyID int64, days int) ([]DailyStat, error) {
	rows, err := db.Query(`
		SELECT date, COALESCE(challenges_issued, 0), COALESCE(verifications_ok, 0), COALESCE(verifications_fail, 0)
		FROM daily_stats
		WHERE api_key_id = ? AND date >= date('now', ?)
		ORDER BY date DESC
	`, apiKeyID, fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []DailyStat
	for rows.Next() {
		var s DailyStat
		if err := rows.Scan(&s.Date, &s.ChallengesIssued, &s.VerificationsOK, &s.VerificationsFail); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}
