package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	os.Unsetenv("GATECHA_LISTEN_ADDR")
	os.Unsetenv("GATECHA_DB_PATH")
	os.Unsetenv("GATECHA_SECRET_KEY")
	os.Unsetenv("GATECHA_ADMIN_USERNAME")
	os.Unsetenv("GATECHA_ADMIN_PASSWORD")
	os.Unsetenv("GATECHA_LOG_LEVEL")
	os.Unsetenv("GATECHA_CLEANUP_INTERVAL")
	os.Unsetenv("GATECHA_CORS_ALLOW_ALL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected :8080, got %s", cfg.ListenAddr)
	}
	if cfg.DBPath != "./data/gatecha.db" {
		t.Errorf("expected default DB path, got %s", cfg.DBPath)
	}
	if cfg.AdminUsername != "admin" {
		t.Errorf("expected admin, got %s", cfg.AdminUsername)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected info, got %s", cfg.LogLevel)
	}
	if cfg.CORSAllowAll {
		t.Error("expected CORSAllowAll to be false by default")
	}
	if cfg.CleanupInterval != 10*time.Minute {
		t.Errorf("expected 10m, got %v", cfg.CleanupInterval)
	}
	if cfg.SecretKey == "" {
		t.Error("expected auto-generated SecretKey")
	}
	if cfg.AdminPassword == "" {
		t.Error("expected auto-generated AdminPassword")
	}
}

func TestLoad_CustomEnv(t *testing.T) {
	t.Setenv("GATECHA_LISTEN_ADDR", ":9090")
	t.Setenv("GATECHA_DB_PATH", "/tmp/test.db")
	t.Setenv("GATECHA_SECRET_KEY", "my-secret")
	t.Setenv("GATECHA_ADMIN_USERNAME", "superadmin")
	t.Setenv("GATECHA_ADMIN_PASSWORD", "my-password")
	t.Setenv("GATECHA_LOG_LEVEL", "debug")
	t.Setenv("GATECHA_CLEANUP_INTERVAL", "5")
	t.Setenv("GATECHA_CORS_ALLOW_ALL", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.ListenAddr != ":9090" {
		t.Errorf("expected :9090, got %s", cfg.ListenAddr)
	}
	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("expected /tmp/test.db, got %s", cfg.DBPath)
	}
	if cfg.SecretKey != "my-secret" {
		t.Errorf("expected my-secret, got %s", cfg.SecretKey)
	}
	if cfg.AdminUsername != "superadmin" {
		t.Errorf("expected superadmin, got %s", cfg.AdminUsername)
	}
	if cfg.AdminPassword != "my-password" {
		t.Errorf("expected my-password, got %s", cfg.AdminPassword)
	}
	if !cfg.CORSAllowAll {
		t.Error("expected CORSAllowAll to be true")
	}
	if cfg.CleanupInterval != 5*time.Minute {
		t.Errorf("expected 5m, got %v", cfg.CleanupInterval)
	}
}

func TestLoad_InvalidCleanupInterval(t *testing.T) {
	t.Setenv("GATECHA_CLEANUP_INTERVAL", "notanumber")

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid cleanup interval")
	}
}

func TestEnvOrDefault(t *testing.T) {
	key := "TEST_GATECHA_ENV_OR_DEFAULT"
	os.Unsetenv(key)

	if v := envOrDefault(key, "fallback"); v != "fallback" {
		t.Errorf("expected fallback, got %s", v)
	}

	t.Setenv(key, "custom")
	if v := envOrDefault(key, "fallback"); v != "custom" {
		t.Errorf("expected custom, got %s", v)
	}
}
