package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Upellift99/GateCHA/internal/auth"
	"github.com/Upellift99/GateCHA/internal/models"
)

type contextKey string

const apiKeyContextKey contextKey = "apiKey"

func APIKeyMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			keyID := r.URL.Query().Get("apiKey")
			if keyID == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					keyID = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if keyID == "" || !strings.HasPrefix(keyID, "gk_") {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid API key"})
				return
			}

			key, err := models.GetAPIKeyByKeyID(db, keyID)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid API key"})
				return
			}

			if !key.Enabled {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "API key is disabled"})
				return
			}

			if key.Domain != "" {
				origin := r.Header.Get("Origin")
				referer := r.Header.Get("Referer")
				if origin != "" && !matchDomain(origin, key.Domain) && !matchDomain(referer, key.Domain) {
					writeJSON(w, http.StatusForbidden, map[string]string{"error": "domain not allowed"})
					return
				}
			}

			ctx := context.WithValue(r.Context(), apiKeyContextKey, key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAPIKeyFromContext(r *http.Request) *models.APIKey {
	key, _ := r.Context().Value(apiKeyContextKey).(*models.APIKey)
	return key
}

func AdminAuthMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization"})
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := auth.ValidateJWT(token, secretKey)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
				return
			}
			_ = claims
			next.ServeHTTP(w, r)
		})
	}
}

func CORSMiddleware(allowAll bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				origin := r.Header.Get("Origin")
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func matchDomain(urlStr, domain string) bool {
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")
	parts := strings.SplitN(urlStr, "/", 2)
	host := strings.SplitN(parts[0], ":", 2)[0]
	return strings.EqualFold(host, domain)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
