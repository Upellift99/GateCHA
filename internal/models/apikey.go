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
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	KeyID         string    `gorm:"not null;uniqueIndex;size:32" json:"key_id"`
	HMACSecret    string    `gorm:"not null" json:"hmac_secret,omitempty"`
	Name          string    `gorm:"not null;default:''" json:"name"`
	Domain        string    `gorm:"not null;default:''" json:"domain"`
	MaxNumber     int64     `gorm:"not null;default:100000" json:"max_number"`
	ExpireSeconds int       `gorm:"not null;default:300" json:"expire_seconds"`
	Algorithm     string    `gorm:"not null;default:'SHA-256'" json:"algorithm"`
	Enabled       bool      `gorm:"not null;default:true" json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UpdateAPIKeyParams holds the fields for updating an API key.
type UpdateAPIKeyParams struct {
	Name          string
	Domain        string
	MaxNumber     int64
	ExpireSeconds int
	Algorithm     string
	Enabled       bool
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

func ListAPIKeys(db *gorm.DB) ([]APIKey, error) {
	var keys []APIKey
	return keys, db.Order("created_at desc").Find(&keys).Error
}

func UpdateAPIKey(db *gorm.DB, id int64, params UpdateAPIKeyParams) error {
	return db.Model(&APIKey{}).Where("id = ?", id).Updates(map[string]any{
		"name":           params.Name,
		"domain":         params.Domain,
		"max_number":     params.MaxNumber,
		"expire_seconds": params.ExpireSeconds,
		"algorithm":      params.Algorithm,
		"enabled":        params.Enabled,
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
