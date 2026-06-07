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
	TrustProxy      bool
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
	}

	if err := loadOptions(cfg); err != nil {
		return nil, err
	}
	if err := ensureSecrets(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// loadOptions parses the boolean and numeric tuning knobs from the environment,
// applying defaults and validating that sizes/rates are strictly positive.
func loadOptions(cfg *Config) error {
	var err error
	if cfg.CORSAllowAll, err = envBool("GATECHA_CORS_ALLOW_ALL", false); err != nil {
		return err
	}
	if cfg.TrustProxy, err = envBool("GATECHA_TRUST_PROXY", false); err != nil {
		return err
	}
	if cfg.EnableHSTS, err = envBool("GATECHA_ENABLE_HSTS", false); err != nil {
		return err
	}
	if cfg.RateLimit, err = envBool("GATECHA_RATE_LIMIT_ENABLED", true); err != nil {
		return err
	}

	intervalMin, err := strconv.Atoi(envOrDefault("GATECHA_CLEANUP_INTERVAL", "10"))
	if err != nil {
		return fmt.Errorf("invalid GATECHA_CLEANUP_INTERVAL: %w", err)
	}
	cfg.CleanupInterval = time.Duration(intervalMin) * time.Minute

	if cfg.MaxBodyBytes, err = envInt64("GATECHA_MAX_BODY_BYTES", 1<<20); err != nil {
		return err
	}
	if cfg.RateLimitLogin, err = envInt("GATECHA_RATE_LIMIT_LOGIN", 5); err != nil {
		return err
	}
	if cfg.RateLimitAPI, err = envInt("GATECHA_RATE_LIMIT_API", 60); err != nil {
		return err
	}

	// A zero/negative body cap or rate would silently disable the protection or
	// surprise the operator.
	for _, c := range []struct {
		key string
		val int64
	}{
		{"GATECHA_MAX_BODY_BYTES", cfg.MaxBodyBytes},
		{"GATECHA_RATE_LIMIT_LOGIN", int64(cfg.RateLimitLogin)},
		{"GATECHA_RATE_LIMIT_API", int64(cfg.RateLimitAPI)},
	} {
		if c.val <= 0 {
			return fmt.Errorf("%s must be > 0, got %d", c.key, c.val)
		}
	}
	return nil
}

// ensureSecrets fills in a random SecretKey and AdminPassword when unset,
// printing each generated value once so the operator can capture it.
func ensureSecrets(cfg *Config) error {
	if cfg.SecretKey == "" {
		key, err := generateRandomHex(32)
		if err != nil {
			return fmt.Errorf("failed to generate secret key: %w", err)
		}
		cfg.SecretKey = key
		fmt.Printf("⚠ No GATECHA_SECRET_KEY set. Generated: %s\n", cfg.SecretKey)
		fmt.Println("  Set this as an environment variable to persist sessions across restarts.")
	}

	if cfg.AdminPassword == "" {
		pw, err := generateRandomHex(16)
		if err != nil {
			return fmt.Errorf("failed to generate admin password: %w", err)
		}
		cfg.AdminPassword = pw
		fmt.Printf("⚠ No GATECHA_ADMIN_PASSWORD set. Generated: %s\n", cfg.AdminPassword)
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) (bool, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("invalid %s: %q is not a boolean (use true/false)", key, v)
	}
	return b, nil
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
