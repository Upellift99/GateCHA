package models

import (
	"database/sql"
	"fmt"
	"time"
)

type DailyStat struct {
	Date              string `json:"date"`
	ChallengesIssued  int    `json:"challenges_issued"`
	VerificationsOK   int    `json:"verifications_ok"`
	VerificationsFail int    `json:"verifications_fail"`
}

type StatsOverview struct {
	TotalChallenges       int         `json:"total_challenges"`
	TotalVerificationsOK  int         `json:"total_verifications_ok"`
	TotalVerificationsFail int        `json:"total_verifications_fail"`
	ActiveKeys            int         `json:"active_keys"`
	Daily                 []DailyStat `json:"daily"`
}

func IncrementChallengesIssued(db *sql.DB, apiKeyID int64) error {
	date := time.Now().UTC().Format("2006-01-02")
	_, err := db.Exec(`
		INSERT INTO daily_stats (api_key_id, date, challenges_issued)
		VALUES (?, ?, 1)
		ON CONFLICT(api_key_id, date)
		DO UPDATE SET challenges_issued = challenges_issued + 1
	`, apiKeyID, date)
	return err
}

func IncrementVerificationsOK(db *sql.DB, apiKeyID int64) error {
	date := time.Now().UTC().Format("2006-01-02")
	_, err := db.Exec(`
		INSERT INTO daily_stats (api_key_id, date, verifications_ok)
		VALUES (?, ?, 1)
		ON CONFLICT(api_key_id, date)
		DO UPDATE SET verifications_ok = verifications_ok + 1
	`, apiKeyID, date)
	return err
}

func IncrementVerificationsFail(db *sql.DB, apiKeyID int64) error {
	date := time.Now().UTC().Format("2006-01-02")
	_, err := db.Exec(`
		INSERT INTO daily_stats (api_key_id, date, verifications_fail)
		VALUES (?, ?, 1)
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
		SELECT date, SUM(challenges_issued), SUM(verifications_ok), SUM(verifications_fail)
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

func GetKeyStats(db *sql.DB, apiKeyID int64, days int) ([]DailyStat, error) {
	rows, err := db.Query(`
		SELECT date, challenges_issued, verifications_ok, verifications_fail
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
