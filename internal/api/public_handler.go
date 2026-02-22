package api

import (
	"database/sql"
	"net/http"

	"github.com/Upellift99/GateCHA/internal/models"
)

type PublicHandler struct {
	DB *sql.DB
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
