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

const bearerPrefix = "Bearer "

type contextKey string

const apiKeyContextKey contextKey = "apiKey"

func authenticateAPIKey(db *sql.DB, w http.ResponseWriter, r *http.Request) (*http.Request, bool) {
	keyID := r.URL.Query().Get("apiKey")
	if keyID == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, bearerPrefix) {
			keyID = strings.TrimPrefix(authHeader, bearerPrefix)
		}
	}

	if keyID == "" || !strings.HasPrefix(keyID, "gk_") {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid API key"})
		return nil, false
	}

	key, err := models.GetAPIKeyByKeyID(db, keyID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid API key"})
		return nil, false
	}

	if !key.Enabled {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "API key is disabled"})
		return nil, false
	}

	if key.Domain != "" {
		origin := r.Header.Get("Origin")
		referer := r.Header.Get("Referer")
		if origin != "" && !matchDomain(origin, key.Domain) && !matchDomain(referer, key.Domain) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "domain not allowed"})
			return nil, false
		}
	}

	ctx := context.WithValue(r.Context(), apiKeyContextKey, key)
	return r.WithContext(ctx), true
}

func APIKeyMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req, ok := authenticateAPIKey(db, w, r)
			if !ok {
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

func GetAPIKeyFromContext(r *http.Request) *models.APIKey {
	key, _ := r.Context().Value(apiKeyContextKey).(*models.APIKey)
	return key
}

func authenticateAdmin(secretKey string, w http.ResponseWriter, r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization"})
		return false
	}
	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if _, err := auth.ValidateJWT(token, secretKey); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
		return false
	}
	return true
}

func AdminAuthMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !authenticateAdmin(secretKey, w, r) {
				return
			}
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
