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
	DBDriver        string
	DBDSN           string
	SecretKey       string
	AdminUsername   string
	AdminPassword   string
	LogLevel        string
	CleanupInterval time.Duration
	CORSAllowAll    bool
	EnableHSTS      bool
	MaxBodyBytes    int64
	RateLimit       bool
	RateLimitLogin  int
	RateLimitAPI    int
}

func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr:    envOrDefault("GATECHA_LISTEN_ADDR", ":8080"),
		DBDriver:      envOrDefault("GATECHA_DB_DRIVER", "sqlite"),
		DBDSN:         envOrDefault("GATECHA_DB_DSN", "./data/gatecha.db"),
		SecretKey:     os.Getenv("GATECHA_SECRET_KEY"),
		AdminUsername: envOrDefault("GATECHA_ADMIN_USERNAME", "admin"),
		AdminPassword: os.Getenv("GATECHA_ADMIN_PASSWORD"),
		LogLevel:      envOrDefault("GATECHA_LOG_LEVEL", "info"),
		CORSAllowAll:  envBool("GATECHA_CORS_ALLOW_ALL", false),
		EnableHSTS:    envBool("GATECHA_ENABLE_HSTS", false),
		RateLimit:     envBool("GATECHA_RATE_LIMIT_ENABLED", true),
	}

	intervalStr := envOrDefault("GATECHA_CLEANUP_INTERVAL", "10")
	intervalMin, err := strconv.Atoi(intervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid GATECHA_CLEANUP_INTERVAL: %w", err)
	}
	cfg.CleanupInterval = time.Duration(intervalMin) * time.Minute

	if cfg.MaxBodyBytes, err = envInt64("GATECHA_MAX_BODY_BYTES", 1<<20); err != nil {
		return nil, err
	}
	if cfg.RateLimitLogin, err = envInt("GATECHA_RATE_LIMIT_LOGIN", 5); err != nil {
		return nil, err
	}
	if cfg.RateLimitAPI, err = envInt("GATECHA_RATE_LIMIT_API", 60); err != nil {
		return nil, err
	}

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

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v == "true" || v == "1"
}

func envInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return n, nil
}

func envInt64(key string, fallback int64) (int64, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return n, nil
}

func generateRandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
