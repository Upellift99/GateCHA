package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Upellift99/GateCHA/internal/api"
	"github.com/Upellift99/GateCHA/internal/auth"
	"github.com/Upellift99/GateCHA/internal/config"
	"github.com/Upellift99/GateCHA/internal/database"
	"github.com/Upellift99/GateCHA/internal/models"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	setupLogger(cfg.LogLevel)

	slog.Info("starting GateCHA", "listen", cfg.ListenAddr, "db_driver", cfg.DBDriver)

	// Per-IP rate limiting keys off the TCP peer address. Behind a reverse proxy
	// with GATECHA_TRUST_PROXY=false, that peer is the proxy, so every visitor
	// shares one bucket — which exhausts almost immediately and makes the public
	// ALTCHA challenge endpoint return 429s, breaking the login captcha. Warn so
	// the operator enables TRUST_PROXY (only safe behind a trusted proxy).
	if cfg.RateLimit && !cfg.TrustProxy {
		slog.Warn("rate limiting is enabled but GATECHA_TRUST_PROXY=false; " +
			"if GateCHA runs behind a reverse proxy, all clients share one rate-limit bucket " +
			"(keyed on the proxy IP), which can break the login captcha. " +
			"Set GATECHA_TRUST_PROXY=true when behind a trusted proxy that sets X-Forwarded-For/X-Real-IP.")
	}

	db, err := database.Open(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	if err := database.RunMigrations(db,
		&models.AdminUser{},
		&models.APIKey{},
		&models.ConsumedChallenge{},
		&models.DailyStat{},
		&models.DailyCountryStat{},
		&models.HISSample{},
		&models.Setting{},
	); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	if err := auth.EnsureAdminUser(db, cfg.AdminUsername, cfg.AdminPassword); err != nil {
		slog.Error("failed to ensure admin user", "error", err)
		os.Exit(1)
	}

	// Start cleanup worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cleanupWorker(ctx, db, cfg.CleanupInterval, cfg.HISSampleRetention)

	router := api.NewRouter(db, cfg.SecretKey, api.RouterConfig{
		CORSAllowAll:     cfg.CORSAllowAll,
		TrustProxy:       cfg.TrustProxy,
		EnableHSTS:       cfg.EnableHSTS,
		MaxBodyBytes:     cfg.MaxBodyBytes,
		RateLimitEnabled: cfg.RateLimit,
		RateLimitLogin:   cfg.RateLimitLogin,
		RateLimitAPI:     cfg.RateLimitAPI,
	})

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Printf("\n  GateCHA is running at http://localhost%s\n", cfg.ListenAddr)
		// Don't log the admin password on every startup; an auto-generated one
		// is shown once by config.Load when GATECHA_ADMIN_PASSWORD is unset.
		fmt.Printf("  Admin user: %s\n\n", cfg.AdminUsername)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}

func cleanupWorker(ctx context.Context, db *gorm.DB, interval, hisRetention time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			deleted, err := models.CleanupExpired(db)
			if err != nil {
				slog.Error("cleanup error", "error", err)
			} else if deleted > 0 {
				slog.Info("cleaned up expired challenges", "count", deleted)
			}

			pruned, err := models.PruneHISSamples(db, time.Now().Add(-hisRetention))
			if err != nil {
				slog.Error("his sample prune error", "error", err)
			} else if pruned > 0 {
				slog.Info("pruned old HIS samples", "count", pruned)
			}
		}
	}
}

func setupLogger(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))
}
