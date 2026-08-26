package models_test

import (
	"strings"
	"testing"

	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
)

func TestGenerateKeyID(t *testing.T) {
	id, err := models.GenerateKeyID()
	if err != nil {
		t.Fatalf("GenerateKeyID failed: %v", err)
	}
	if !strings.HasPrefix(id, "gk_") {
		t.Errorf("expected gk_ prefix, got %s", id)
	}
	if len(id) != 27 { // "gk_" + 24 hex chars
		t.Errorf("expected length 27, got %d", len(id))
	}

	id2, _ := models.GenerateKeyID()
	if id == id2 {
		t.Error("expected unique key IDs")
	}
}

func TestGenerateHMACSecret(t *testing.T) {
	secret, err := models.GenerateHMACSecret()
	if err != nil {
		t.Fatalf("GenerateHMACSecret failed: %v", err)
	}
	if len(secret) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("expected length 64, got %d", len(secret))
	}
}

func TestCreateAPIKey(t *testing.T) {
	db := testutil.SetupTestDB(t)

	key, err := models.CreateAPIKey(db, "Test Key", "example.com", 50000, 600, "SHA-256")
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

	key, err := models.CreateAPIKey(db, "Default Key", "", 0, 0, "")
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
	created, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	found, err := models.GetAPIKeyByKeyID(db, created.KeyID)
	if err != nil {
		t.Fatalf("GetAPIKeyByKeyID failed: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, found.ID)
	}
}

func TestGetAPIKeyByKeyID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)

	_, err := models.GetAPIKeyByKeyID(db, "gk_nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestGetAPIKeyByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	created, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")

	found, err := models.GetAPIKeyByID(db, created.ID)
	if err != nil {
		t.Fatalf("GetAPIKeyByID failed: %v", err)
	}
	if found.KeyID != created.KeyID {
		t.Errorf("expected KeyID %s, got %s", created.KeyID, found.KeyID)
	}
}

func TestGetAPIKeyByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)

	_, err := models.GetAPIKeyByID(db, 99999)
	if err == nil {
		t.Error("expected error for nonexistent ID")
	}
}

func TestListAPIKeys(t *testing.T) {
	db := testutil.SetupTestDB(t)

	keys, err := models.ListAPIKeys(db)
	if err != nil {
		t.Fatalf("ListAPIKeys failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}

	models.CreateAPIKey(db, "Key 1", "", 0, 0, "")
	models.CreateAPIKey(db, "Key 2", "", 0, 0, "")

	keys, err = models.ListAPIKeys(db)
	if err != nil {
		t.Fatalf("ListAPIKeys failed: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestUpdateAPIKey(t *testing.T) {
	db := testutil.SetupTestDB(t)
	created, _ := models.CreateAPIKey(db, "Original", "old.com", 10000, 100, "SHA-256")

	err := models.UpdateAPIKey(db, created.ID, models.UpdateAPIKeyParams{
		Name:               "Updated",
		Domain:             "new.com",
		MaxNumber:          200000,
		ExpireSeconds:      600,
		Algorithm:          "SHA-512",
		RateLimitPerMin:    120,
		AdaptiveDifficulty: true,
		Enabled:            false,
	})
	if err != nil {
		t.Fatalf("UpdateAPIKey failed: %v", err)
	}

	updated, _ := models.GetAPIKeyByID(db, created.ID)
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
	if updated.RateLimitPerMin != 120 {
		t.Errorf("expected rate_limit_per_min 120, got %d", updated.RateLimitPerMin)
	}
	if !updated.AdaptiveDifficulty {
		t.Error("expected adaptive_difficulty to be enabled")
	}
	if updated.Enabled {
		t.Error("expected key to be disabled")
	}
}

func TestDeleteAPIKey(t *testing.T) {
	db := testutil.SetupTestDB(t)
	created, _ := models.CreateAPIKey(db, "ToDelete", "", 0, 0, "")

	err := models.DeleteAPIKey(db, created.ID)
	if err != nil {
		t.Fatalf("DeleteAPIKey failed: %v", err)
	}

	_, err = models.GetAPIKeyByID(db, created.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestRotateHMACSecret(t *testing.T) {
	db := testutil.SetupTestDB(t)
	created, _ := models.CreateAPIKey(db, "Test", "", 0, 0, "")
	oldSecret := created.HMACSecret

	newSecret, err := models.RotateHMACSecret(db, created.ID)
	if err != nil {
		t.Fatalf("RotateHMACSecret failed: %v", err)
	}
	if newSecret == oldSecret {
		t.Error("expected new secret to differ from old")
	}

	updated, _ := models.GetAPIKeyByID(db, created.ID)
	if updated.HMACSecret != newSecret {
		t.Error("expected stored secret to match returned secret")
	}
}

func TestListAPIKeysOmitsHMACSecret(t *testing.T) {
	db := testutil.SetupTestDB(t)

	created, err := models.CreateAPIKey(db, "Listed", "listed.com", 10000, 100, "SHA-256")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}
	if created.HMACSecret == "" {
		t.Fatal("expected the created key to carry a secret")
	}

	keys, err := models.ListAPIKeys(db)
	if err != nil {
		t.Fatalf("ListAPIKeys failed: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}

	if keys[0].HMACSecret != "" {
		t.Error("ListAPIKeys must not return the HMAC secret")
	}

	// Omitting one column must not blank out the rest of the row.
	if keys[0].ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, keys[0].ID)
	}
	if keys[0].KeyID != created.KeyID {
		t.Errorf("expected key_id %q, got %q", created.KeyID, keys[0].KeyID)
	}
	if keys[0].Name != "Listed" || keys[0].Domain != "listed.com" {
		t.Errorf("unexpected name/domain: %q / %q", keys[0].Name, keys[0].Domain)
	}
	if keys[0].MaxNumber != 10000 || keys[0].ExpireSeconds != 100 {
		t.Errorf("unexpected max_number/expire_seconds: %d / %d", keys[0].MaxNumber, keys[0].ExpireSeconds)
	}
	if !keys[0].Enabled {
		t.Error("expected the key to stay enabled")
	}

	// The secret is still on disk and still reachable where it is needed.
	fetched, err := models.GetAPIKeyByID(db, created.ID)
	if err != nil {
		t.Fatalf("GetAPIKeyByID failed: %v", err)
	}
	if fetched.HMACSecret != created.HMACSecret {
		t.Error("GetAPIKeyByID must still return the HMAC secret")
	}
}

func TestUpdateAPIKeyFieldsTouchesOnlyNamedColumns(t *testing.T) {
	db := testutil.SetupTestDB(t)

	key, err := models.CreateAPIKey(db, "Original", "original.com", 20000, 120, "SHA-512")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	if err := models.UpdateAPIKeyFields(db, key.ID, map[string]any{"name": "Renamed"}); err != nil {
		t.Fatalf("UpdateAPIKeyFields failed: %v", err)
	}

	got, err := models.GetAPIKeyByID(db, key.ID)
	if err != nil {
		t.Fatalf("GetAPIKeyByID failed: %v", err)
	}
	if got.Name != "Renamed" {
		t.Errorf("expected the name to change, got %q", got.Name)
	}
	if got.Domain != "original.com" || got.MaxNumber != 20000 || got.ExpireSeconds != 120 || got.Algorithm != "SHA-512" {
		t.Errorf("an unnamed column was overwritten: %+v", got)
	}
	if !got.Enabled {
		t.Error("enabled was overwritten by a write that never named it")
	}
}

func TestUpdateAPIKeyFieldsEmptyIsANoOp(t *testing.T) {
	db := testutil.SetupTestDB(t)

	key, err := models.CreateAPIKey(db, "Untouched", "u.com", 0, 0, "")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	// GORM errors on an empty Updates map, so this has to be handled rather than
	// forwarded: a tool call that changes nothing is not a failure.
	if err := models.UpdateAPIKeyFields(db, key.ID, nil); err != nil {
		t.Errorf("an empty update should be a no-op, got %v", err)
	}
	if err := models.UpdateAPIKeyFields(db, key.ID, map[string]any{}); err != nil {
		t.Errorf("an empty update should be a no-op, got %v", err)
	}

	got, _ := models.GetAPIKeyByID(db, key.ID)
	if got.Name != "Untouched" {
		t.Errorf("an empty update changed the row: %+v", got)
	}
}

func TestSetAPIKeyEnabledLeavesSettingsAlone(t *testing.T) {
	db := testutil.SetupTestDB(t)

	key, err := models.CreateAPIKey(db, "Site", "site.com", 20000, 120, "SHA-512")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	if err := models.SetAPIKeyEnabled(db, key.ID, false); err != nil {
		t.Fatalf("SetAPIKeyEnabled failed: %v", err)
	}
	got, _ := models.GetAPIKeyByID(db, key.ID)
	if got.Enabled {
		t.Error("expected the key to be disabled")
	}
	if got.Name != "Site" || got.Domain != "site.com" || got.MaxNumber != 20000 {
		t.Errorf("flipping enabled overwrote other settings: %+v", got)
	}

	if err := models.SetAPIKeyEnabled(db, key.ID, true); err != nil {
		t.Fatalf("SetAPIKeyEnabled failed: %v", err)
	}
	got, _ = models.GetAPIKeyByID(db, key.ID)
	if !got.Enabled {
		t.Error("expected the key to be enabled again")
	}
}
