package models

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// StatsTokenPrefix marks a read-only per-key stats credential, distinct from
// the public gk_ key ID (which ships in every page's HTML and so cannot guard
// anything) and from the gm_ MCP admin tokens.
const StatsTokenPrefix = "gs_"

// statsTokenDisplayLen mirrors mcpTokenDisplayLen: the prefix plus 8 hex
// characters, enough to recognise a token without narrowing a search.
const statsTokenDisplayLen = len(StatsTokenPrefix) + 8

// StatsTokenMaxDays bounds the window a stats token can read.
const StatsTokenMaxDays = 90

// statsTokenTopCountries caps the per-country breakdown of a stats report.
const statsTokenTopCountries = 10

// ErrStatsTokenNotFound is returned when no key holds the presented token.
var ErrStatsTokenNotFound = errors.New("stats token not found")

// IssueStatsToken generates a new stats token for the key, replacing any
// previous one, and returns the secret. It is shown once and unrecoverable
// afterwards. The digest is the same SHA-256 the MCP tokens use, for the same
// reason: 128 bits of CSPRNG output need no slow hash.
func IssueStatsToken(db *gorm.DB, id int64) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	secret := StatsTokenPrefix + hex.EncodeToString(b)
	hash := HashMCPToken(secret)
	res := db.Model(&APIKey{}).Where(byID, id).Updates(map[string]any{
		"stats_token_hash":    hash,
		"stats_token_display": secret[:statsTokenDisplayLen],
	})
	if res.Error != nil {
		return "", res.Error
	}
	if res.RowsAffected == 0 {
		return "", gorm.ErrRecordNotFound
	}
	return secret, nil
}

// RevokeStatsToken removes the key's stats token, if any. Existence is checked
// apart from the update: MySQL reports a write that changes nothing (a key that
// had no token) as zero rows affected, which is not a missing key.
func RevokeStatsToken(db *gorm.DB, id int64) error {
	var n int64
	if err := db.Model(&APIKey{}).Where(byID, id).Count(&n).Error; err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return db.Model(&APIKey{}).Where(byID, id).Updates(map[string]any{
		"stats_token_hash":    nil,
		"stats_token_display": "",
	}).Error
}

// GetAPIKeyByStatsToken returns the key a stats token belongs to. A disabled
// key still resolves: its owner may want to read why traffic stopped.
func GetAPIKeyByStatsToken(db *gorm.DB, secret string) (*APIKey, error) {
	if !strings.HasPrefix(secret, StatsTokenPrefix) {
		return nil, ErrStatsTokenNotFound
	}
	var key APIKey
	err := db.Omit("hmac_secret").Where("stats_token_hash = ?", HashMCPToken(secret)).First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrStatsTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// StatsDay is one day of a stats report.
type StatsDay struct {
	Date              string `json:"date"`
	ChallengesIssued  int    `json:"challenges_issued"`
	VerificationsOK   int    `json:"verifications_ok"`
	VerificationsFail int    `json:"verifications_fail"`
	HISBotSuspected   int    `json:"his_bot_suspected"`
}

// StatsTotals sums a stats report's window.
type StatsTotals struct {
	ChallengesIssued  int `json:"challenges_issued"`
	VerificationsOK   int `json:"verifications_ok"`
	VerificationsFail int `json:"verifications_fail"`
	HISObservations   int `json:"his_observations"`
	HISBotSuspected   int `json:"his_bot_suspected"`
}

// KeyStatsReport is what a stats token can read: one key's counters over a
// window, with every day present (zero-filled) so a client can draw a series
// without knowing which days had traffic.
type KeyStatsReport struct {
	Days      int           `json:"days"`
	Totals    StatsTotals   `json:"totals"`
	Daily     []StatsDay    `json:"daily"`
	Countries []CountryStat `json:"countries"`
}

// GetKeyStatsReport builds the report for one key over the last days days,
// today included, oldest first.
func GetKeyStatsReport(db *gorm.DB, apiKeyID int64, days int) (*KeyStatsReport, error) {
	if days < 1 {
		days = 1
	}
	if days > StatsTokenMaxDays {
		days = StatsTokenMaxDays
	}
	today := time.Now().UTC()
	since := today.AddDate(0, 0, -(days - 1)).Format(dateFormatYMD)

	var rows []DailyStat
	if err := db.Where("api_key_id = ? AND date >= ?", apiKeyID, since).Find(&rows).Error; err != nil {
		return nil, err
	}
	byDate := make(map[string]DailyStat, len(rows))
	for _, r := range rows {
		byDate[r.Date] = r
	}

	report := &KeyStatsReport{Days: days, Daily: make([]StatsDay, 0, days)}
	for i := days - 1; i >= 0; i-- {
		date := today.AddDate(0, 0, -i).Format(dateFormatYMD)
		r := byDate[date]
		report.Daily = append(report.Daily, StatsDay{
			Date:              date,
			ChallengesIssued:  r.ChallengesIssued,
			VerificationsOK:   r.VerificationsOK,
			VerificationsFail: r.VerificationsFail,
			HISBotSuspected:   r.HISBotSuspected,
		})
		report.Totals.ChallengesIssued += r.ChallengesIssued
		report.Totals.VerificationsOK += r.VerificationsOK
		report.Totals.VerificationsFail += r.VerificationsFail
		report.Totals.HISObservations += r.HISObservations
		report.Totals.HISBotSuspected += r.HISBotSuspected
	}

	// GetCountryStats counts its window as "date >= now - days", one day wider
	// than the series above; days-1 keeps the two windows identical.
	countries, err := GetCountryStats(db, &apiKeyID, days-1, statsTokenTopCountries)
	if err != nil {
		return nil, err
	}
	report.Countries = countries
	return report, nil
}
