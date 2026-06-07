package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Upellift99/GateCHA/internal/auth"
	"github.com/Upellift99/GateCHA/internal/models"
	"gorm.io/gorm"
)

const bearerPrefix = "Bearer "

type contextKey string

const apiKeyContextKey contextKey = "apiKey"

func authenticateAPIKey(db *gorm.DB, w http.ResponseWriter, r *http.Request) (*http.Request, bool) {
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

func APIKeyMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
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

// securityCSP is a baseline Content-Security-Policy tuned for the embedded Vue
// SPA: all scripts/styles/assets ship from the same origin (the ALTCHA widget is
// bundled, not loaded from a CDN), so 'self' is sufficient. 'unsafe-inline' is
// kept for styles only, to tolerate runtime style injection by Vue/Tailwind.
// worker-src allows blob: because the ALTCHA widget runs its proof-of-work in a
// Worker created from a blob URL; without this the login captcha cannot start.
const securityCSP = "default-src 'self'; base-uri 'self'; frame-ancestors 'none'; " +
	"form-action 'self'; object-src 'none'; img-src 'self' data:; " +
	"style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'; " +
	"worker-src 'self' blob:"

// SecurityHeadersMiddleware sets a baseline of security-related response headers.
// HSTS is opt-in because GateCHA is commonly run over plain HTTP locally or
// behind a TLS-terminating reverse proxy.
func SecurityHeadersMiddleware(enableHSTS bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Frame-Options", "DENY")
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Content-Security-Policy", securityCSP)
			if enableHSTS {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}

// MaxBodyBytesMiddleware caps the size of request bodies to guard against
// memory-exhaustion. An oversized body surfaces as a read error, which the
// JSON-decoding handlers already translate into a 4xx response.
func MaxBodyBytesMiddleware(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limit > 0 && r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, limit)
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
