package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ListenAddr      string
	DBPath          string
	SecretKey       string
	AdminUsername   string
	AdminPassword   string
	LogLevel        string
	CleanupInterval time.Duration
	CORSAllowAll   bool
}

func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr:      envOrDefault("GATECHA_LISTEN_ADDR", ":8080"),
		DBPath:          envOrDefault("GATECHA_DB_PATH", "./data/gatecha.db"),
		SecretKey:       os.Getenv("GATECHA_SECRET_KEY"),
		AdminUsername:   envOrDefault("GATECHA_ADMIN_USERNAME", "admin"),
		AdminPassword:   os.Getenv("GATECHA_ADMIN_PASSWORD"),
		LogLevel:        envOrDefault("GATECHA_LOG_LEVEL", "info"),
		CORSAllowAll:    envOrDefault("GATECHA_CORS_ALLOW_ALL", "false") == "true",
	}

	intervalStr := envOrDefault("GATECHA_CLEANUP_INTERVAL", "10")
	intervalMin, err := strconv.Atoi(intervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid GATECHA_CLEANUP_INTERVAL: %w", err)
	}
	cfg.CleanupInterval = time.Duration(intervalMin) * time.Minute

	if cfg.SecretKey == "" {
		key, err := generateRandomHex(32)
		if err != nil {
			return nil, fmt.Errorf("failed to generate secret key: %w", err)
		}
		cfg.SecretKey = key
		fmt.Printf("⚠ No GATECHA_SECRET_KEY set. Generated: %s\n", cfg.SecretKey)
		fmt.Println("  Set this as an environment variable to persist sessions across restarts.")
	}

	if cfg.AdminPassword == "" {
		pw, err := generateRandomHex(16)
		if err != nil {
			return nil, fmt.Errorf("failed to generate admin password: %w", err)
		}
		cfg.AdminPassword = pw
		fmt.Printf("⚠ No GATECHA_ADMIN_PASSWORD set. Generated: %s\n", cfg.AdminPassword)
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func generateRandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
