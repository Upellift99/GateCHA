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
	HISSampling bool `gorm:"not null;default:false" json:"his_sampling"`
	// HISEnforce, when set, rejects a verification whose HIS score reaches
	// HISThreshold, instead of merely reporting it. Off by default: HIS ships
	// in Monitor mode and blocking is a decision the operator makes after
	// reading their own calibration histogram, never one inherited from us.
	// It can only ever act on a request that actually carried `his_signals`;
	// a client shipping no collector sends nothing, is scored not at all, and
	// passes untouched whatever this says.
	HISEnforce bool `gorm:"not null;default:false" json:"his_enforce"`
	// HISThreshold is the score at or above which this key treats a sample as
	// automation: it drives the reported `his_bot_suspected`, the Monitor
	// counters, the calibration histogram's marker and, when HISEnforce is on,
	// the rejection. One number, one meaning.
	//
	// The default is deliberately his.BotSuspectThreshold rather than anything
	// lower. A sample whose collector ran and observed nothing at all scores
	// exactly 0.70, and that shape is produced both by automation and by a
	// keyboard-only or assistive-technology visitor. Operators who have read
	// their own histogram can and do lower this; operators who have not should
	// not inherit a number that rejects their accessible traffic.
	HISThreshold float64   `gorm:"not null;default:0.8" json:"his_threshold"`
	Enabled      bool      `gorm:"not null;default:true" json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DefaultHISThreshold mirrors his.BotSuspectThreshold. It is restated here
// rather than imported so that models keeps no dependency on the scoring
// package; his_test asserts the two stay equal.
const DefaultHISThreshold = 0.8

// AfterFind normalises the stored threshold on every read path, so no caller
// ever sees an out-of-range value: not the scorer, not the admin API, and not
// the edit form, which would otherwise display a 0 that its own validation then
// refuses to save back. GORM runs this for Find, First and Scan alike.
func (k *APIKey) AfterFind(*gorm.DB) error {
	k.HISThreshold = k.SuspectThreshold()
	return nil
}

// SuspectThreshold is the score at or above which this key counts a sample as
// automation. Out-of-range values (a column added before this field existed,
// or a bad write) fall back to the default rather than silently meaning
// "reject everything", which is what a stored 0 would otherwise do.
func (k *APIKey) SuspectThreshold() float64 {
	if k.HISThreshold <= 0 || k.HISThreshold > 1 {
		return DefaultHISThreshold
	}
	return k.HISThreshold
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
	HISEnforce         bool
	HISThreshold       float64
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
		HISThreshold:  DefaultHISThreshold,
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

// byID is the WHERE clause every single-key write shares, spelled once so the
// three writers cannot drift apart.
const byID = "id = ?"

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
	return db.Model(&APIKey{}).Where(byID, id).Updates(fields).Error
}

// SetAPIKeyEnabled flips only the enabled flag, so enabling or disabling a key
// can never clobber a concurrent edit to its settings.
func SetAPIKeyEnabled(db *gorm.DB, id int64, enabled bool) error {
	return UpdateAPIKeyFields(db, id, map[string]any{"enabled": enabled})
}

func UpdateAPIKey(db *gorm.DB, id int64, params UpdateAPIKeyParams) error {
	return db.Model(&APIKey{}).Where(byID, id).Updates(map[string]any{
		"name":                params.Name,
		"domain":              params.Domain,
		"max_number":          params.MaxNumber,
		"expire_seconds":      params.ExpireSeconds,
		"algorithm":           params.Algorithm,
		"rate_limit_per_min":  params.RateLimitPerMin,
		"adaptive_difficulty": params.AdaptiveDifficulty,
		"his_sampling":        params.HISSampling,
		"his_enforce":         params.HISEnforce,
		"his_threshold":       params.HISThreshold,
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
	if err := db.Model(&APIKey{}).Where(byID, id).Update("hmac_secret", secret).Error; err != nil {
		return "", err
	}
	return secret, nil
}
