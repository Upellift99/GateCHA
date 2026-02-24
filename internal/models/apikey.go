package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

type APIKey struct {
	ID            int64  `json:"id"`
	KeyID         string `json:"key_id"`
	HMACSecret    string `json:"hmac_secret,omitempty"`
	Name          string `json:"name"`
	Domain        string `json:"domain"`
	MaxNumber     int64  `json:"max_number"`
	ExpireSeconds int    `json:"expire_seconds"`
	Algorithm     string `json:"algorithm"`
	Enabled       bool   `json:"enabled"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
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

func CreateAPIKey(db *sql.DB, name, domain string, maxNumber int64, expireSeconds int, algorithm string) (*APIKey, error) {
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

	now := time.Now().UTC().Format(time.RFC3339)
	result, err := db.Exec(`
		INSERT INTO api_keys (key_id, hmac_secret, name, domain, max_number, expire_seconds, algorithm, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, keyID, hmacSecret, name, domain, maxNumber, expireSeconds, algorithm, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to insert API key: %w", err)
	}

	id, _ := result.LastInsertId()
	return &APIKey{
		ID:            id,
		KeyID:         keyID,
		HMACSecret:    hmacSecret,
		Name:          name,
		Domain:        domain,
		MaxNumber:     maxNumber,
		ExpireSeconds: expireSeconds,
		Algorithm:     algorithm,
		Enabled:       true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func GetAPIKeyByKeyID(db *sql.DB, keyID string) (*APIKey, error) {
	var k APIKey
	var enabled int
	err := db.QueryRow(`
		SELECT id, key_id, hmac_secret, name, domain, max_number, expire_seconds, algorithm, enabled, created_at, updated_at
		FROM api_keys WHERE key_id = ?
	`, keyID).Scan(&k.ID, &k.KeyID, &k.HMACSecret, &k.Name, &k.Domain, &k.MaxNumber, &k.ExpireSeconds, &k.Algorithm, &enabled, &k.CreatedAt, &k.UpdatedAt)
	if err != nil {
		return nil, err
	}
	k.Enabled = enabled == 1
	return &k, nil
}

func GetAPIKeyByID(db *sql.DB, id int64) (*APIKey, error) {
	var k APIKey
	var enabled int
	err := db.QueryRow(`
		SELECT id, key_id, hmac_secret, name, domain, max_number, expire_seconds, algorithm, enabled, created_at, updated_at
		FROM api_keys WHERE id = ?
	`, id).Scan(&k.ID, &k.KeyID, &k.HMACSecret, &k.Name, &k.Domain, &k.MaxNumber, &k.ExpireSeconds, &k.Algorithm, &enabled, &k.CreatedAt, &k.UpdatedAt)
	if err != nil {
		return nil, err
	}
	k.Enabled = enabled == 1
	return &k, nil
}

func ListAPIKeys(db *sql.DB) ([]APIKey, error) {
	rows, err := db.Query(`
		SELECT id, key_id, hmac_secret, name, domain, max_number, expire_seconds, algorithm, enabled, created_at, updated_at
		FROM api_keys ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		var enabled int
		if err := rows.Scan(&k.ID, &k.KeyID, &k.HMACSecret, &k.Name, &k.Domain, &k.MaxNumber, &k.ExpireSeconds, &k.Algorithm, &enabled, &k.CreatedAt, &k.UpdatedAt); err != nil {
			return nil, err
		}
		k.Enabled = enabled == 1
		keys = append(keys, k)
	}
	return keys, nil
}

func UpdateAPIKey(db *sql.DB, id int64, params UpdateAPIKeyParams) error {
	enabledInt := 0
	if params.Enabled {
		enabledInt = 1
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(`
		UPDATE api_keys SET name = ?, domain = ?, max_number = ?, expire_seconds = ?, algorithm = ?, enabled = ?, updated_at = ?
		WHERE id = ?
	`, params.Name, params.Domain, params.MaxNumber, params.ExpireSeconds, params.Algorithm, enabledInt, now, id)
	return err
}

func DeleteAPIKey(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM api_keys WHERE id = ?`, id)
	return err
}

func RotateHMACSecret(db *sql.DB, id int64) (string, error) {
	secret, err := GenerateHMACSecret()
	if err != nil {
		return "", err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.Exec(`UPDATE api_keys SET hmac_secret = ?, updated_at = ? WHERE id = ?`, secret, now, id)
	if err != nil {
		return "", err
	}
	return secret, nil
}
