package api

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Upellift99/GateCHA/internal/models"
	"gorm.io/gorm"
)

// KeyStatsHandler serves GET /api/v1/stats: one key's statistics, read with
// that key's stats token (gs_) rather than its key ID. The key ID is public, it
// sits in the HTML of every protected page; the token is a server-side secret
// a site's backend (the WordPress plugin, for one) keeps to show the owner
// their numbers without admin access to the instance.
type KeyStatsHandler struct {
	DB *gorm.DB
}

type keyStatsResponse struct {
	KeyID   string `json:"key_id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	*models.KeyStatsReport
}

func (h *KeyStatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), bearerPrefix)
	if !strings.HasPrefix(token, models.StatsTokenPrefix) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid stats token"})
		return
	}

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		n, err := strconv.Atoi(d)
		if err != nil || n < 1 || n > models.StatsTokenMaxDays {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "days must be between 1 and " + strconv.Itoa(models.StatsTokenMaxDays),
			})
			return
		}
		days = n
	}

	key, err := models.GetAPIKeyByStatsToken(h.DB, token)
	if errors.Is(err, models.ErrStatsTokenNotFound) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid stats token"})
		return
	}
	if err != nil {
		slog.Error("stats token lookup failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch stats"})
		return
	}

	report, err := models.GetKeyStatsReport(h.DB, key.ID, days)
	if err != nil {
		slog.Error("stats report failed", "error", err, "api_key_id", key.ID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch stats"})
		return
	}

	// Counters move with every verification; a cache in between would show
	// the owner stale numbers right after they test their form.
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, keyStatsResponse{
		KeyID:          key.KeyID,
		Name:           key.Name,
		Enabled:        key.Enabled,
		KeyStatsReport: report,
	})
}
