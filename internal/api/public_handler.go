package api

import (
	"net/http"

	"github.com/Upellift99/GateCHA/internal/dashboard"
	"github.com/Upellift99/GateCHA/internal/models"
	"gorm.io/gorm"
)

type PublicHandler struct {
	DB *gorm.DB
}

// GET /api/public/login-config
func (h *PublicHandler) LoginConfig(w http.ResponseWriter, r *http.Request) {
	enabled, err := models.GetLoginCaptchaEnabled(h.DB)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch config"})
		return
	}

	resp := map[string]interface{}{
		"captcha_required": enabled,
	}

	if enabled {
		key, err := models.EnsureLoginCaptchaAPIKey(h.DB)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get captcha key"})
			return
		}
		resp["challenge_url"] = "/api/v1/challenge?apiKey=" + key.KeyID
	}

	writeJSON(w, http.StatusOK, resp)
}

// GET /api/public/his.js
//
// Serves the standalone interaction-signal collector so a site that does not
// build this dashboard can still emit `his_signals` on POST /api/v1/verify.
// Without it the only client producing signals is GateCHA's own login page,
// which is why calibration data has been unrepresentative.
//
// Unauthenticated on purpose: it is public client code with no secrets in it,
// and a <script src> cannot carry an API key. It is scoped per request by the
// key the site already uses on /verify, not by who fetched the script.
func (h *PublicHandler) HISCollector(w http.ResponseWriter, r *http.Request) {
	script, err := dashboard.CollectorJS()
	if err != nil {
		// The binary was built without the frontend assets.
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write(script)
}
