package models

import (
	"strings"
	"testing"

	"github.com/Upellift99/GateCHA/internal/testutil"
)

func TestGenerateKeyID(t *testing.T) {
	id, err := GenerateKeyID()
	if err != nil {
		t.Fatalf("GenerateKeyID failed: %v", err)
	}
	if !strings.HasPrefix(id, "gk_") {
		t.Errorf("expected gk_ prefix, got %s", id)
	}
	if len(id) != 27 { // "gk_" + 24 hex chars
		t.Errorf("expected length 27, got %d", len(id))
	}

	id2, _ := GenerateKeyID()
	if id == id2 {
		t.Error("expected unique key IDs")
	}
}

func TestGenerateHMACSecret(t *testing.T) {
	secret, err := GenerateHMACSecret()
	if err != nil {
		t.Fatalf("GenerateHMACSecret failed: %v", err)
	}
	if len(secret) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("expected length 64, got %d", len(secret))
	}
}

func TestCreateAPIKey(t *testing.T) {
	db := testutil.SetupTestDB(t)

	key, err := CreateAPIKey(db, "Test Key", "example.com", 50000, 600, "SHA-256")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}
	if key.Name != "Test Key" {
		t.Errorf("expected name 'Test Key', got %q", key.Name)
	}
	if key.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", key.Domain)
	}
	if key.MaxNumber != 50000 {
		t.Errorf("expected max_number 50000, got %d", key.MaxNumber)
	}
	if key.ExpireSeconds != 600 {
		t.Errorf("expected expire_seconds 600, got %d", key.ExpireSeconds)
	}
	if key.Algorithm != "SHA-256" {
		t.Errorf("expected algorithm SHA-256, got %s", key.Algorithm)
	}
	if !key.Enabled {
		t.Error("expected key to be enabled")
	}
	if !strings.HasPrefix(key.KeyID, "gk_") {
		t.Errorf("expected gk_ prefix, got %s", key.KeyID)
	}
	if key.HMACSecret == "" {
		t.Error("expected non-empty HMAC secret")
	}
}

func TestCreateAPIKey_Defaults(t *testing.T) {
	db := testutil.SetupTestDB(t)

	key, err := CreateAPIKey(db, "Default Key", "", 0, 0, "")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}
	if key.MaxNumber != 100000 {
		t.Errorf("expected default max_number 100000, got %d", key.MaxNumber)
	}
	if key.ExpireSeconds != 300 {
		t.Errorf("expected default expire_seconds 300, got %d", key.ExpireSeconds)
	}
	if key.Algorithm != "SHA-256" {
		t.Errorf("expected default algorithm SHA-256, got %s", key.Algorithm)
	}
}

func TestGetAPIKeyByKeyID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	created, _ := CreateAPIKey(db, "Test", "", 0, 0, "")

	found, err := GetAPIKeyByKeyID(db, created.KeyID)
	if err != nil {
		t.Fatalf("GetAPIKeyByKeyID failed: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, found.ID)
	}
}

func TestGetAPIKeyByKeyID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)

	_, err := GetAPIKeyByKeyID(db, "gk_nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestGetAPIKeyByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	created, _ := CreateAPIKey(db, "Test", "", 0, 0, "")

	found, err := GetAPIKeyByID(db, created.ID)
	if err != nil {
		t.Fatalf("GetAPIKeyByID failed: %v", err)
	}
	if found.KeyID != created.KeyID {
		t.Errorf("expected KeyID %s, got %s", created.KeyID, found.KeyID)
	}
}

func TestGetAPIKeyByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)

	_, err := GetAPIKeyByID(db, 99999)
	if err == nil {
		t.Error("expected error for nonexistent ID")
	}
}

func TestListAPIKeys(t *testing.T) {
	db := testutil.SetupTestDB(t)

	keys, err := ListAPIKeys(db)
	if err != nil {
		t.Fatalf("ListAPIKeys failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}

	CreateAPIKey(db, "Key 1", "", 0, 0, "")
	CreateAPIKey(db, "Key 2", "", 0, 0, "")

	keys, err = ListAPIKeys(db)
	if err != nil {
		t.Fatalf("ListAPIKeys failed: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestUpdateAPIKey(t *testing.T) {
	db := testutil.SetupTestDB(t)
	created, _ := CreateAPIKey(db, "Original", "old.com", 10000, 100, "SHA-256")

	err := UpdateAPIKey(db, created.ID, UpdateAPIKeyParams{
		Name:          "Updated",
		Domain:        "new.com",
		MaxNumber:     200000,
		ExpireSeconds: 600,
		Algorithm:     "SHA-512",
		Enabled:       false,
	})
	if err != nil {
		t.Fatalf("UpdateAPIKey failed: %v", err)
	}

	updated, _ := GetAPIKeyByID(db, created.ID)
	if updated.Name != "Updated" {
		t.Errorf("expected name 'Updated', got %q", updated.Name)
	}
	if updated.Domain != "new.com" {
		t.Errorf("expected domain 'new.com', got %q", updated.Domain)
	}
	if updated.MaxNumber != 200000 {
		t.Errorf("expected max_number 200000, got %d", updated.MaxNumber)
	}
	if updated.ExpireSeconds != 600 {
		t.Errorf("expected expire_seconds 600, got %d", updated.ExpireSeconds)
	}
	if updated.Algorithm != "SHA-512" {
		t.Errorf("expected algorithm SHA-512, got %s", updated.Algorithm)
	}
	if updated.Enabled {
		t.Error("expected key to be disabled")
	}
}

func TestDeleteAPIKey(t *testing.T) {
	db := testutil.SetupTestDB(t)
	created, _ := CreateAPIKey(db, "ToDelete", "", 0, 0, "")

	err := DeleteAPIKey(db, created.ID)
	if err != nil {
		t.Fatalf("DeleteAPIKey failed: %v", err)
	}

	_, err = GetAPIKeyByID(db, created.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestRotateHMACSecret(t *testing.T) {
	db := testutil.SetupTestDB(t)
	created, _ := CreateAPIKey(db, "Test", "", 0, 0, "")
	oldSecret := created.HMACSecret

	newSecret, err := RotateHMACSecret(db, created.ID)
	if err != nil {
		t.Fatalf("RotateHMACSecret failed: %v", err)
	}
	if newSecret == oldSecret {
		t.Error("expected new secret to differ from old")
	}

	updated, _ := GetAPIKeyByID(db, created.ID)
	if updated.HMACSecret != newSecret {
		t.Error("expected stored secret to match returned secret")
	}
}
