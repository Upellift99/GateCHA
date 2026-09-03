package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Upellift99/GateCHA/internal/altcha"
	"github.com/Upellift99/GateCHA/internal/auth"
	"github.com/Upellift99/GateCHA/internal/his"
	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

const (
	errInvalidRequest = "invalid request"
	errInvalidKeyID   = "invalid key ID"
	errKeyNotFound    = "key not found"
	errFetchStats     = "failed to fetch stats"
	errUpdateSettings = "failed to update settings"

	errInvalidHISThreshold = "his_threshold must be greater than 0 and at most 1"
)

// orNil, orZero and orNonPositive spell the three "omitted means unchanged"
// conventions this API uses, so a partial update reads as a list of fields
// rather than as a wall of branches. They differ in how the caller signals
// "leave it alone", and the difference is not cosmetic: a pointer is the only
// one of the three that can carry a meaningful zero.
func orNil[T any](supplied *T, current T) T {
	if supplied == nil {
		return current
	}
	return *supplied
}

// orZero treats the zero value as "omitted". Only safe for fields where the
// zero value is not a setting anyone would want, such as an empty name.
func orZero[T comparable](supplied, current T) T {
	var zero T
	if supplied == zero {
		return current
	}
	return supplied
}

// orNonPositive additionally treats negatives as "omitted", for the numeric
// fields where a negative is nonsense rather than a request.
func orNonPositive[T int | int64](supplied, current T) T {
	if supplied <= 0 {
		return current
	}
	return supplied
}

// validHISThreshold bounds the per-key suspect threshold. 0 is excluded on
// purpose: every score is >= 0, so a threshold of 0 means "treat every scored
// request as automation", which on a key with enforcement on takes the site
// down. Nobody types 0 meaning that.
func validHISThreshold(v float64) bool {
	return v > 0 && v <= 1
}

type AdminHandler struct {
	DB           *gorm.DB
	SecretKey    string
	BuildVersion string
}

// verifyLoginCaptcha validates the ALTCHA captcha payload during login.
// Returns true if the captcha is valid, false otherwise (response is already written).
func (h *AdminHandler) verifyLoginCaptcha(w http.ResponseWriter, payload string) bool {
	key, err := models.EnsureLoginCaptchaAPIKey(h.DB)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return false
	}
	if payload == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "captcha required"})
		return false
	}
	valid, err := altcha.VerifyPayload(key.HMACSecret, payload)
	if err != nil || !valid {
		if err := models.IncrementVerificationsFail(h.DB, key.ID); err != nil {
			slog.Error("failed to increment verifications_fail", "error", err, "api_key_id", key.ID)
		}
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid captcha"})
		return false
	}
	if err := models.IncrementVerificationsOK(h.DB, key.ID); err != nil {
		slog.Error("failed to increment verifications_ok", "error", err, "api_key_id", key.ID)
	}
	return true
}

// POST /api/admin/login
func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username      string       `json:"username"`
		Password      string       `json:"password"`
		AltchaPayload string       `json:"altcha_payload"`
		HISSignals    *his.Signals `json:"his_signals,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidRequest})
		return
	}

	ok, err := auth.ValidateCredentials(h.DB, req.Username, req.Password)
	if err != nil || !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	captchaEnabled, err := models.GetLoginCaptchaEnabled(h.DB)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if captchaEnabled {
		// Record HIS in Monitor mode against the login captcha key before the
		// captcha check, so failed attempts are observed too. Never blocks login.
		if req.HISSignals != nil {
			if key, err := models.EnsureLoginCaptchaAPIKey(h.DB); err == nil {
				recordHISMonitor(h.DB, key, req.HISSignals)
			}
		}
		if !h.verifyLoginCaptcha(w, req.AltchaPayload) {
			return
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

// GET /api/admin/version
// Returns the build version. Kept behind admin auth so the running version is
// not disclosed to unauthenticated visitors.
func (h *AdminHandler) Version(w http.ResponseWriter, r *http.Request) {
	v := h.BuildVersion
	if v == "" {
		v = "dev"
	}
	writeJSON(w, http.StatusOK, map[string]string{"version": v})
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
		Name               string `json:"name"`
		Domain             string `json:"domain"`
		MaxNumber          int64  `json:"max_number"`
		ExpireSeconds      int    `json:"expire_seconds"`
		Algorithm          string `json:"algorithm"`
		RateLimitPerMin    int    `json:"rate_limit_per_min"`
		AdaptiveDifficulty bool   `json:"adaptive_difficulty"`
		HISSampling        bool   `json:"his_sampling"`
		HISEnforce         bool   `json:"his_enforce"`
		// A pointer so that omitting the field selects the model default while
		// an explicit value is always honoured or refused. Spelled as a plain
		// float it made `{"his_threshold": 0}` indistinguishable from silence,
		// so the caller was answered 201 with a policy they had not asked for,
		// where the update path rejects the same body.
		HISThreshold *float64 `json:"his_threshold"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidRequest})
		return
	}

	if req.HISThreshold != nil && !validHISThreshold(*req.HISThreshold) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidHISThreshold})
		return
	}

	key, err := models.CreateAPIKey(h.DB, req.Name, req.Domain, req.MaxNumber, req.ExpireSeconds, req.Algorithm)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create key"})
		return
	}

	// CreateAPIKey keeps its narrow signature; apply the optional advanced levers
	// as a follow-up update so create stays a single source of defaults.
	hisThreshold := orNil(req.HISThreshold, key.SuspectThreshold())
	if req.RateLimitPerMin > 0 || req.AdaptiveDifficulty || req.HISSampling || req.HISEnforce || req.HISThreshold != nil {
		if err := models.UpdateAPIKey(h.DB, key.ID, models.UpdateAPIKeyParams{
			Name:               key.Name,
			Domain:             key.Domain,
			MaxNumber:          key.MaxNumber,
			ExpireSeconds:      key.ExpireSeconds,
			Algorithm:          key.Algorithm,
			RateLimitPerMin:    req.RateLimitPerMin,
			AdaptiveDifficulty: req.AdaptiveDifficulty,
			HISSampling:        req.HISSampling,
			HISEnforce:         req.HISEnforce,
			HISThreshold:       hisThreshold,
			Enabled:            key.Enabled,
		}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create key"})
			return
		}
		key.RateLimitPerMin = req.RateLimitPerMin
		key.AdaptiveDifficulty = req.AdaptiveDifficulty
		key.HISSampling = req.HISSampling
		key.HISEnforce = req.HISEnforce
		key.HISThreshold = hisThreshold
	}

	writeJSON(w, http.StatusCreated, key)
}

// GET /api/admin/keys/{id}
func (h *AdminHandler) GetKey(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidKeyID})
		return
	}

	key, err := models.GetAPIKeyByID(h.DB, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": errKeyNotFound})
		return
	}

	writeJSON(w, http.StatusOK, key)
}

// PUT /api/admin/keys/{id}
func (h *AdminHandler) UpdateKey(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidKeyID})
		return
	}

	var req struct {
		Name               string   `json:"name"`
		Domain             string   `json:"domain"`
		MaxNumber          int64    `json:"max_number"`
		ExpireSeconds      int      `json:"expire_seconds"`
		Algorithm          string   `json:"algorithm"`
		RateLimitPerMin    *int     `json:"rate_limit_per_min"`
		AdaptiveDifficulty *bool    `json:"adaptive_difficulty"`
		HISSampling        *bool    `json:"his_sampling"`
		HISEnforce         *bool    `json:"his_enforce"`
		HISThreshold       *float64 `json:"his_threshold"`
		Enabled            *bool    `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidRequest})
		return
	}

	// Rejected rather than clamped, and before anything is loaded: a threshold
	// quietly corrected to something else is a quietly different blocking
	// policy, and the caller would never learn which one they got.
	if req.HISThreshold != nil && !validHISThreshold(*req.HISThreshold) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidHISThreshold})
		return
	}

	existing, err := models.GetAPIKeyByID(h.DB, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": errKeyNotFound})
		return
	}

	if err := models.UpdateAPIKey(h.DB, id, models.UpdateAPIKeyParams{
		Name:          orZero(req.Name, existing.Name),
		Domain:        orZero(req.Domain, existing.Domain),
		MaxNumber:     orNonPositive(req.MaxNumber, existing.MaxNumber),
		ExpireSeconds: orNonPositive(req.ExpireSeconds, existing.ExpireSeconds),
		Algorithm:     orZero(req.Algorithm, existing.Algorithm),
		// These arrive as pointers because their zero value is meaningful:
		// rate_limit_per_min 0 disables the per-key limit, and false is a real
		// setting for the three switches, so "omitted" cannot be inferred from
		// the value itself.
		RateLimitPerMin:    orNil(req.RateLimitPerMin, existing.RateLimitPerMin),
		AdaptiveDifficulty: orNil(req.AdaptiveDifficulty, existing.AdaptiveDifficulty),
		HISSampling:        orNil(req.HISSampling, existing.HISSampling),
		HISEnforce:         orNil(req.HISEnforce, existing.HISEnforce),
		HISThreshold:       orNil(req.HISThreshold, existing.SuspectThreshold()),
		Enabled:            orNil(req.Enabled, existing.Enabled),
	}); err != nil {
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidKeyID})
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidKeyID})
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": errFetchStats})
		return
	}

	writeJSON(w, http.StatusOK, overview)
}

// GET /api/admin/stats/keys/{id}
func (h *AdminHandler) KeyStats(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidKeyID})
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
		writeJSON(w, http.StatusNotFound, map[string]string{"error": errKeyNotFound})
		return
	}

	stats, err := models.GetKeyStats(h.DB, id, days)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": errFetchStats})
		return
	}
	if stats == nil {
		stats = []models.DailyStat{}
	}

	countries, err := models.GetCountryStats(h.DB, &id, days, 20)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": errFetchStats})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"key_id":    key.KeyID,
		"name":      key.Name,
		"days":      stats,
		"countries": countries,
	})
}

// GET /api/admin/his/calibration?key_id=&days=
// Returns the score distribution and signal averages over stored HIS samples,
// to help calibrate an enforcement threshold before any blocking is enabled.
func (h *AdminHandler) HISCalibration(w http.ResponseWriter, r *http.Request) {
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	var keyID *int64
	if k := r.URL.Query().Get("key_id"); k != "" {
		parsed, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidKeyID})
			return
		}
		keyID = &parsed
	}

	// The histogram's marker has to show the number this key actually applies,
	// not the package default: on a key whose threshold was lowered, a marker
	// drawn at 0.8 would describe a policy the key no longer follows.
	threshold := his.BotSuspectThreshold
	if keyID != nil {
		if key, err := models.GetAPIKeyByID(h.DB, *keyID); err == nil {
			threshold = key.SuspectThreshold()
		}
	}

	cal, err := models.GetHISCalibration(h.DB, keyID, days, threshold)
	if err != nil {
		// Logged, not just returned: the response body is deliberately vague,
		// and without this an operator hitting the 500 had nothing to report
		// and no way to tell a query failure from an empty result.
		slog.Error("failed to fetch HIS calibration", "error", err, "key_id", keyID, "days", days)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch calibration"})
		return
	}

	writeJSON(w, http.StatusOK, cal)
}

// POST /api/admin/change-password
func (h *AdminHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidRequest})
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
	settings, err := h.readSettings()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch settings"})
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

// readSettings gathers every exposed setting. Settings are added here rather
// than inline in the handlers so the read and write paths cannot drift apart.
func (h *AdminHandler) readSettings() (map[string]interface{}, error) {
	captcha, err := models.GetLoginCaptchaEnabled(h.DB)
	if err != nil {
		return nil, err
	}
	mcp, err := models.GetMCPEnabled(h.DB)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"login_captcha_enabled": captcha,
		"mcp_enabled":           mcp,
	}, nil
}

// PUT /api/admin/settings
func (h *AdminHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LoginCaptchaEnabled *bool `json:"login_captcha_enabled"`
		MCPEnabled          *bool `json:"mcp_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidRequest})
		return
	}

	if req.LoginCaptchaEnabled != nil {
		if *req.LoginCaptchaEnabled {
			if _, err := models.EnsureLoginCaptchaAPIKey(h.DB); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to init captcha key"})
				return
			}
		}
		if err := models.SetSetting(h.DB, models.SettingLoginCaptchaEnabled, boolSetting(*req.LoginCaptchaEnabled)); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": errUpdateSettings})
			return
		}
	}

	if req.MCPEnabled != nil {
		if err := models.SetSetting(h.DB, models.SettingMCPEnabled, boolSetting(*req.MCPEnabled)); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": errUpdateSettings})
			return
		}
	}

	settings, err := h.readSettings()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch settings"})
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func boolSetting(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

// GET /api/admin/mcp-tokens
func (h *AdminHandler) ListMCPTokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := models.ListMCPTokens(h.DB)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list mcp tokens"})
		return
	}
	if tokens == nil {
		tokens = []models.MCPToken{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"tokens": tokens})
}

// POST /api/admin/mcp-tokens
func (h *AdminHandler) CreateMCPToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		ReadOnly bool   `json:"read_only"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": errInvalidRequest})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	token, secret, err := models.CreateMCPToken(h.DB, name, req.ReadOnly)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create mcp token"})
		return
	}

	// The secret is returned exactly once, like an API key's HMAC secret.
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"token":  token,
		"secret": secret,
	})
}

// DELETE /api/admin/mcp-tokens/{id}
func (h *AdminHandler) DeleteMCPToken(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid token ID"})
		return
	}

	deleted, err := models.DeleteMCPToken(h.DB, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete mcp token"})
		return
	}
	if !deleted {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "token not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
