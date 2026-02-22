package models

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
)

const (
	SettingLoginCaptchaEnabled  = "login_captcha_enabled"
	SettingLoginCaptchaAPIKeyID = "login_captcha_api_key_id"
)

// GetSetting retrieves a single setting value by key.
// Returns ("", nil) if the key does not exist.
func GetSetting(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return value, err
}

// SetSetting upserts a setting value.
func SetSetting(db *sql.DB, key, value string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(`
		INSERT INTO settings (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, value, now)
	return err
}

// GetLoginCaptchaEnabled returns whether the login CAPTCHA is enabled.
func GetLoginCaptchaEnabled(db *sql.DB) (bool, error) {
	v, err := GetSetting(db, SettingLoginCaptchaEnabled)
	return v == "true", err
}

// EnsureLoginCaptchaAPIKey returns the existing login CAPTCHA API key,
// or creates a dedicated one if none exists yet.
func EnsureLoginCaptchaAPIKey(db *sql.DB) (*APIKey, error) {
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
