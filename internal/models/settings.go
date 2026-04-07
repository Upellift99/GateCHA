package models

import (
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Setting holds a key-value configuration entry.
type Setting struct {
	Key       string    `gorm:"primaryKey;size:64" json:"key"`
	Value     string    `gorm:"not null;default:''" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	SettingLoginCaptchaEnabled  = "login_captcha_enabled"
	SettingLoginCaptchaAPIKeyID = "login_captcha_api_key_id"
)

// GetSetting retrieves a single setting value by key.
// Returns ("", nil) if the key does not exist.
func GetSetting(db *gorm.DB, key string) (string, error) {
	var s Setting
	// Map-form Where quotes the column name per dialect, avoiding MySQL's reserved word "key".
	err := db.Where(map[string]any{"key": key}).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return s.Value, err
}

// SetSetting upserts a setting value.
func SetSetting(db *gorm.DB, key, value string) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&Setting{Key: key, Value: value, UpdatedAt: time.Now()}).Error
}

// GetLoginCaptchaEnabled returns whether the login CAPTCHA is enabled.
func GetLoginCaptchaEnabled(db *gorm.DB) (bool, error) {
	v, err := GetSetting(db, SettingLoginCaptchaEnabled)
	return v == "true", err
}

// EnsureLoginCaptchaAPIKey returns the existing login CAPTCHA API key,
// or creates a dedicated one if none exists yet.
func EnsureLoginCaptchaAPIKey(db *gorm.DB) (*APIKey, error) {
	idStr, err := GetSetting(db, SettingLoginCaptchaAPIKeyID)
	if err != nil {
		return nil, err
	}

	if idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, err
		}
		key, err := GetAPIKeyByID(db, id)
		if err == nil {
			return key, nil
		}
		// Key was deleted — fall through to create a new one
	}

	key, err := CreateAPIKey(db, "Login CAPTCHA", "", 50000, 300, "SHA-256")
	if err != nil {
		return nil, err
	}

	if err := SetSetting(db, SettingLoginCaptchaAPIKeyID, strconv.FormatInt(key.ID, 10)); err != nil {
		return nil, err
	}

	return key, nil
}
