package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Upellift99/GateCHA/internal/altcha"
	"github.com/Upellift99/GateCHA/internal/auth"
	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	DB        *sql.DB
	SecretKey string
}

// POST /api/admin/login
func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username      string `json:"username"`
		Password      string `json:"password"`
		AltchaPayload string `json:"altcha_payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	ok, err := auth.ValidateCredentials(h.DB, req.Username, req.Password)
	if err != nil || !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	// Verify ALTCHA payload if login CAPTCHA is enabled
	captchaEnabled, err := models.GetLoginCaptchaEnabled(h.DB)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if captchaEnabled {
		if req.AltchaPayload == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "captcha required"})
			return
		}
		key, err := models.EnsureLoginCaptchaAPIKey(h.DB)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		valid, err := altcha.VerifyPayload(key.HMACSecret, req.AltchaPayload)
		if err != nil || !valid {
			if err := models.IncrementVerificationsFail(h.DB, key.ID); err != nil {
				slog.Error("failed to increment verifications_fail", "error", err, "api_key_id", key.ID)
			}
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid captcha"})
			return
		}
		if err := models.IncrementVerificationsOK(h.DB, key.ID); err != nil {
			slog.Error("failed to increment verifications_ok", "error", err, "api_key_id", key.ID)
		}
	}

	token, expiresAt, err := auth.GenerateJWT(req.Username, h.SecretKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":      token,
		"expires_at": expiresAt,
	})
}

// GET /api/admin/me
func (h *AdminHandler) Me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"username": "admin"})
}

// GET /api/admin/keys
func (h *AdminHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := models.ListAPIKeys(h.DB)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list keys"})
		return
	}
	if keys == nil {
		keys = []models.APIKey{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"keys": keys})
}

// POST /api/admin/keys
func (h *AdminHandler) CreateKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string `json:"name"`
		Domain        string `json:"domain"`
		MaxNumber     int64  `json:"max_number"`
		ExpireSeconds int    `json:"expire_seconds"`
		Algorithm     string `json:"algorithm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	key, err := models.CreateAPIKey(h.DB, req.Name, req.Domain, req.MaxNumber, req.ExpireSeconds, req.Algorithm)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create key"})
		return
	}

	writeJSON(w, http.StatusCreated, key)
}

// GET /api/admin/keys/{id}
func (h *AdminHandler) GetKey(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid key ID"})
		return
	}

	key, err := models.GetAPIKeyByID(h.DB, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "key not found"})
		return
	}

	writeJSON(w, http.StatusOK, key)
}

// PUT /api/admin/keys/{id}
func (h *AdminHandler) UpdateKey(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid key ID"})
		return
	}

	var req struct {
		Name          string `json:"name"`
		Domain        string `json:"domain"`
		MaxNumber     int64  `json:"max_number"`
		ExpireSeconds int    `json:"expire_seconds"`
		Algorithm     string `json:"algorithm"`
		Enabled       *bool  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	existing, err := models.GetAPIKeyByID(h.DB, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "key not found"})
		return
	}

	name := existing.Name
	if req.Name != "" {
		name = req.Name
	}
	domain := existing.Domain
	if req.Domain != "" {
		domain = req.Domain
	}
	maxNumber := existing.MaxNumber
	if req.MaxNumber > 0 {
		maxNumber = req.MaxNumber
	}
	expireSeconds := existing.ExpireSeconds
	if req.ExpireSeconds > 0 {
		expireSeconds = req.ExpireSeconds
	}
	algorithm := existing.Algorithm
	if req.Algorithm != "" {
		algorithm = req.Algorithm
	}
	enabled := existing.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	if err := models.UpdateAPIKey(h.DB, id, name, domain, maxNumber, expireSeconds, algorithm, enabled); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update key"})
		return
	}

	updated, _ := models.GetAPIKeyByID(h.DB, id)
	writeJSON(w, http.StatusOK, updated)
}

// DELETE /api/admin/keys/{id}
func (h *AdminHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid key ID"})
		return
	}

	if err := models.DeleteAPIKey(h.DB, id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete key"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// POST /api/admin/keys/{id}/rotate-secret
func (h *AdminHandler) RotateSecret(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid key ID"})
		return
	}

	newSecret, err := models.RotateHMACSecret(h.DB, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to rotate secret"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"hmac_secret": newSecret})
}

// GET /api/admin/stats/keys-summary
func (h *AdminHandler) KeysStatsSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := models.GetAllKeysStatsSummary(h.DB)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch stats summary"})
		return
	}

	// Convert map to a JSON-friendly structure keyed by string IDs
	result := make(map[string]models.KeyStatsSummary)
	for id, s := range summary {
		result[strconv.FormatInt(id, 10)] = s
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"keys": result})
}

// GET /api/admin/stats/overview
func (h *AdminHandler) StatsOverview(w http.ResponseWriter, r *http.Request) {
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	overview, err := models.GetStatsOverview(h.DB, days)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch stats"})
		return
	}

	writeJSON(w, http.StatusOK, overview)
}

// GET /api/admin/stats/keys/{id}
func (h *AdminHandler) KeyStats(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid key ID"})
		return
	}

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	key, err := models.GetAPIKeyByID(h.DB, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "key not found"})
		return
	}

	stats, err := models.GetKeyStats(h.DB, id, days)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch stats"})
		return
	}
	if stats == nil {
		stats = []models.DailyStat{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"key_id": key.KeyID,
		"name":   key.Name,
		"days":   stats,
	})
}

// POST /api/admin/change-password
func (h *AdminHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	ok, err := auth.ValidateCredentials(h.DB, "admin", req.CurrentPassword)
	if err != nil || !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid current password"})
		return
	}

	if err := auth.ChangePassword(h.DB, "admin", req.NewPassword); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to change password"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "password changed"})
}

// GET /api/admin/settings
func (h *AdminHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	enabled, err := models.GetLoginCaptchaEnabled(h.DB)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch settings"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"login_captcha_enabled": enabled,
	})
}

// PUT /api/admin/settings
func (h *AdminHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LoginCaptchaEnabled *bool `json:"login_captcha_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.LoginCaptchaEnabled != nil {
		val := "false"
		if *req.LoginCaptchaEnabled {
			val = "true"
			if _, err := models.EnsureLoginCaptchaAPIKey(h.DB); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to init captcha key"})
				return
			}
		}
		if err := models.SetSetting(h.DB, models.SettingLoginCaptchaEnabled, val); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update settings"})
			return
		}
	}

	enabled, _ := models.GetLoginCaptchaEnabled(h.DB)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"login_captcha_enabled": enabled,
	})
}
