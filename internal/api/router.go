package api

import (
	"net/http"

	"github.com/Upellift99/GateCHA/internal/dashboard"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

const keysIDRoute = "/keys/{id}"

// RouterConfig holds the request-handling and hardening options for NewRouter.
type RouterConfig struct {
	CORSAllowAll     bool
	EnableHSTS       bool
	MaxBodyBytes     int64
	RateLimitEnabled bool
	RateLimitLogin   int // requests per minute on /api/admin/login
	RateLimitAPI     int // requests per minute on /api/v1/*
}

func NewRouter(db *gorm.DB, secretKey string, cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(SecurityHeadersMiddleware(cfg.EnableHSTS))
	r.Use(MaxBodyBytesMiddleware(cfg.MaxBodyBytes))
	r.Use(CORSMiddleware(cfg.CORSAllowAll))

	publicHandler := &PublicHandler{DB: db}
	challengeHandler := &ChallengeHandler{DB: db}
	verifyHandler := &VerifyHandler{DB: db}
	adminHandler := &AdminHandler{DB: db, SecretKey: secretKey}

	// Public endpoints (no auth, used by login page)
	r.Route("/api/public", func(r chi.Router) {
		r.Get("/login-config", publicHandler.LoginConfig)
	})

	// Public API (API key auth)
	r.Route("/api/v1", func(r chi.Router) {
		if cfg.RateLimitEnabled {
			r.Use(RateLimitMiddleware(cfg.RateLimitAPI))
		}
		r.Use(APIKeyMiddleware(db))
		r.Get("/challenge", challengeHandler.ServeHTTP)
		r.Post("/verify", verifyHandler.ServeHTTP)
	})

	// Admin API
	r.Route("/api/admin", func(r chi.Router) {
		// Login is rate-limited more strictly to slow down credential brute-force.
		r.Group(func(r chi.Router) {
			if cfg.RateLimitEnabled {
				r.Use(RateLimitMiddleware(cfg.RateLimitLogin))
			}
			r.Post("/login", adminHandler.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(AdminAuthMiddleware(secretKey))
			r.Get("/me", adminHandler.Me)
			r.Post("/change-password", adminHandler.ChangePassword)
			r.Get("/settings", adminHandler.GetSettings)
			r.Put("/settings", adminHandler.UpdateSettings)

			// API Keys CRUD
			r.Get("/keys", adminHandler.ListKeys)
			r.Post("/keys", adminHandler.CreateKey)
			r.Get(keysIDRoute, adminHandler.GetKey)
			r.Put(keysIDRoute, adminHandler.UpdateKey)
			r.Delete(keysIDRoute, adminHandler.DeleteKey)
			r.Post(keysIDRoute+"/rotate-secret", adminHandler.RotateSecret)

			// Statistics
			r.Get("/stats/overview", adminHandler.StatsOverview)
			r.Get("/stats/keys-summary", adminHandler.KeysStatsSummary)
			r.Get("/stats/keys/{id}", adminHandler.KeyStats)
		})
	})

	// Health check
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	})

	// SPA Dashboard (catch-all)
	r.Handle("/*", dashboard.SPAHandler())

	return r
}
