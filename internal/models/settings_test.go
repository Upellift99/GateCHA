package models_test

import (
	"testing"

	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
)

func TestGetSetting_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)

	val, err := models.GetSetting(db, "nonexistent")
	if err != nil {
		t.Fatalf("GetSetting failed: %v", err)
	}
	if val != "" {
		t.Errorf("expected empty string, got %q", val)
	}
}

func TestSetAndGetSetting(t *testing.T) {
	db := testutil.SetupTestDB(t)

	if err := models.SetSetting(db, "test_key", "test_value"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}

	val, err := models.GetSetting(db, "test_key")
	if err != nil {
		t.Fatalf("GetSetting failed: %v", err)
	}
	if val != "test_value" {
		t.Errorf("expected 'test_value', got %q", val)
	}
}

func TestSetSetting_Upsert(t *testing.T) {
	db := testutil.SetupTestDB(t)

	models.SetSetting(db, "key", "original")
	models.SetSetting(db, "key", "updated")

	val, _ := models.GetSetting(db, "key")
	if val != "updated" {
		t.Errorf("expected 'updated', got %q", val)
	}
}

func TestGetLoginCaptchaEnabled_Default(t *testing.T) {
	db := testutil.SetupTestDB(t)

	enabled, err := models.GetLoginCaptchaEnabled(db)
	if err != nil {
		t.Fatalf("GetLoginCaptchaEnabled failed: %v", err)
	}
	if enabled {
		t.Error("expected false by default")
	}
}

func TestGetLoginCaptchaEnabled_True(t *testing.T) {
	db := testutil.SetupTestDB(t)

	models.SetSetting(db, models.SettingLoginCaptchaEnabled, "true")

	enabled, _ := models.GetLoginCaptchaEnabled(db)
	if !enabled {
		t.Error("expected true after setting")
	}
}

func TestEnsureLoginCaptchaAPIKey(t *testing.T) {
	db := testutil.SetupTestDB(t)

	key, err := models.EnsureLoginCaptchaAPIKey(db)
	if err != nil {
		t.Fatalf("EnsureLoginCaptchaAPIKey failed: %v", err)
	}
	if key.Name != "Login CAPTCHA" {
		t.Errorf("expected name 'Login CAPTCHA', got %q", key.Name)
	}

	// Second call should return same key
	key2, err := models.EnsureLoginCaptchaAPIKey(db)
	if err != nil {
		t.Fatalf("EnsureLoginCaptchaAPIKey (2nd) failed: %v", err)
	}
	if key2.ID != key.ID {
		t.Errorf("expected same key ID %d, got %d", key.ID, key2.ID)
	}
}

func TestEnsureLoginCaptchaAPIKey_RecreatesAfterDeletion(t *testing.T) {
	db := testutil.SetupTestDB(t)

	key, _ := models.EnsureLoginCaptchaAPIKey(db)
	models.DeleteAPIKey(db, key.ID)

	// Should create a new key
	key2, err := models.EnsureLoginCaptchaAPIKey(db)
	if err != nil {
		t.Fatalf("EnsureLoginCaptchaAPIKey after deletion failed: %v", err)
	}
	if key2.ID == key.ID {
		t.Error("expected a new key after deletion")
	}
}
