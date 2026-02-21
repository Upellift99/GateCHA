package api

import (
	"database/sql"
	"net/http"

	"github.com/Upellift99/GateCHA/internal/dashboard"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(db *sql.DB, secretKey string, corsAllowAll bool) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RealIP)
	r.Use(CORSMiddleware(corsAllowAll))

	challengeHandler := &ChallengeHandler{DB: db}
	verifyHandler := &VerifyHandler{DB: db}
	adminHandler := &AdminHandler{DB: db, SecretKey: secretKey}

	// Public API (API key auth)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(APIKeyMiddleware(db))
		r.Get("/challenge", challengeHandler.ServeHTTP)
		r.Post("/verify", verifyHandler.ServeHTTP)
	})

	// Admin API
	r.Route("/api/admin", func(r chi.Router) {
		r.Post("/login", adminHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(AdminAuthMiddleware(secretKey))
			r.Get("/me", adminHandler.Me)
			r.Post("/change-password", adminHandler.ChangePassword)

			// API Keys CRUD
			r.Get("/keys", adminHandler.ListKeys)
			r.Post("/keys", adminHandler.CreateKey)
			r.Get("/keys/{id}", adminHandler.GetKey)
			r.Put("/keys/{id}", adminHandler.UpdateKey)
			r.Delete("/keys/{id}", adminHandler.DeleteKey)
			r.Post("/keys/{id}/rotate-secret", adminHandler.RotateSecret)

			// Statistics
			r.Get("/stats/overview", adminHandler.StatsOverview)
			r.Get("/stats/keys/{id}", adminHandler.KeyStats)
		})
	})

	// Health check
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	})

	// SPA Dashboard (catch-all)
	r.Handle("/*", dashboard.SPAHandler())

	return r
}
