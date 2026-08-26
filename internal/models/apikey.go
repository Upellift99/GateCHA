package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// APIKey represents a site-specific API key used to generate and verify ALTCHA challenges.
type APIKey struct {
	ID            int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	KeyID         string `gorm:"not null;uniqueIndex;size:32" json:"key_id"`
	HMACSecret    string `gorm:"not null" json:"hmac_secret,omitempty"`
	Name          string `gorm:"not null;default:''" json:"name"`
	Domain        string `gorm:"not null;default:''" json:"domain"`
	MaxNumber     int64  `gorm:"not null;default:100000" json:"max_number"`
	ExpireSeconds int    `gorm:"not null;default:300" json:"expire_seconds"`
	Algorithm     string `gorm:"not null;default:'SHA-256'" json:"algorithm"`
	// RateLimitPerMin caps how many /api/v1 requests this key accepts per minute,
	// aggregated across all clients. 0 means unlimited (only the global per-IP
	// limiter applies).
	RateLimitPerMin int `gorm:"not null;default:0" json:"rate_limit_per_min"`
	// AdaptiveDifficulty, when set, raises the proof-of-work MaxNumber above this
	// key's configured base for clients (by IP) that request challenges at an
	// abusive rate, capped server-side. MaxNumber stays the floor.
	AdaptiveDifficulty bool `gorm:"not null;default:false" json:"adaptive_difficulty"`
	// HISSampling, when set, persists each scored HIS observation for this key
	// (raw aggregates + score) so enforcement thresholds can be calibrated on
	// real traffic. Samples are pruned after the configured retention window.
	HISSampling bool      `gorm:"not null;default:false" json:"his_sampling"`
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UpdateAPIKeyParams holds the fields for updating an API key.
type UpdateAPIKeyParams struct {
	Name               string
	Domain             string
	MaxNumber          int64
	ExpireSeconds      int
	Algorithm          string
	RateLimitPerMin    int
	AdaptiveDifficulty bool
	HISSampling        bool
	Enabled            bool
}

func GenerateKeyID() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "gk_" + hex.EncodeToString(b), nil
}

func GenerateHMACSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func CreateAPIKey(db *gorm.DB, name, domain string, maxNumber int64, expireSeconds int, algorithm string) (*APIKey, error) {
	keyID, err := GenerateKeyID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key ID: %w", err)
	}

	hmacSecret, err := GenerateHMACSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate HMAC secret: %w", err)
	}

	if maxNumber <= 0 {
		maxNumber = 100000
	}
	if expireSeconds <= 0 {
		expireSeconds = 300
	}
	if algorithm == "" {
		algorithm = "SHA-256"
	}

	key := &APIKey{
		KeyID:         keyID,
		HMACSecret:    hmacSecret,
		Name:          name,
		Domain:        domain,
		MaxNumber:     maxNumber,
		ExpireSeconds: expireSeconds,
		Algorithm:     algorithm,
		Enabled:       true,
	}
	if err := db.Create(key).Error; err != nil {
		return nil, fmt.Errorf("failed to insert API key: %w", err)
	}
	return key, nil
}

func GetAPIKeyByKeyID(db *gorm.DB, keyID string) (*APIKey, error) {
	var key APIKey
	err := db.Where("key_id = ?", keyID).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func GetAPIKeyByID(db *gorm.DB, id int64) (*APIKey, error) {
	var key APIKey
	err := db.First(&key, id).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// ListAPIKeys returns every key without its HMAC secret. The listing feeds the
// dashboard table, which never needs the secret; callers that do (key detail,
// creation, rotation) go through GetAPIKeyByID or RotateHMACSecret instead.
// Omitting the column keeps the whole secret set out of the response rather
// than relying on the caller to strip it.
func ListAPIKeys(db *gorm.DB) ([]APIKey, error) {
	var keys []APIKey
	return keys, db.Omit("hmac_secret").Order("created_at desc").Find(&keys).Error
}

// UpdateAPIKeyFields writes only the named columns, leaving every other one
// alone.
//
// UpdateAPIKey writes the whole row, which forces callers into a
// read-modify-write and makes concurrent edits lose updates: a rename that read
// Enabled=true before a disable committed would write the key back to enabled.
// Restricting the write to the fields that actually changed removes the race
// rather than serialising around it. An empty map is a no-op, not an error.
func UpdateAPIKeyFields(db *gorm.DB, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return db.Model(&APIKey{}).Where("id = ?", id).Updates(fields).Error
}

// SetAPIKeyEnabled flips only the enabled flag, so enabling or disabling a key
// can never clobber a concurrent edit to its settings.
func SetAPIKeyEnabled(db *gorm.DB, id int64, enabled bool) error {
	return UpdateAPIKeyFields(db, id, map[string]any{"enabled": enabled})
}

func UpdateAPIKey(db *gorm.DB, id int64, params UpdateAPIKeyParams) error {
	return db.Model(&APIKey{}).Where("id = ?", id).Updates(map[string]any{
		"name":                params.Name,
		"domain":              params.Domain,
		"max_number":          params.MaxNumber,
		"expire_seconds":      params.ExpireSeconds,
		"algorithm":           params.Algorithm,
		"rate_limit_per_min":  params.RateLimitPerMin,
		"adaptive_difficulty": params.AdaptiveDifficulty,
		"his_sampling":        params.HISSampling,
		"enabled":             params.Enabled,
	}).Error
}

func DeleteAPIKey(db *gorm.DB, id int64) error {
	return db.Delete(&APIKey{}, id).Error
}

func RotateHMACSecret(db *gorm.DB, id int64) (string, error) {
	secret, err := GenerateHMACSecret()
	if err != nil {
		return "", err
	}
	if err := db.Model(&APIKey{}).Where("id = ?", id).Update("hmac_secret", secret).Error; err != nil {
		return "", err
	}
	return secret, nil
}
