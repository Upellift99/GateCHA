package api

import (
	"log/slog"
	"net/http"

	"github.com/Upellift99/GateCHA/internal/altcha"
	"github.com/Upellift99/GateCHA/internal/models"
	"gorm.io/gorm"
)

type ChallengeHandler struct {
	DB *gorm.DB
	// Adaptive scales proof-of-work difficulty by source request rate for keys
	// that opt in via AdaptiveDifficulty.
	Adaptive *adaptiveLimiter
}

func (h *ChallengeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := GetAPIKeyFromContext(r)
	if key == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing API key"})
		return
	}

	maxNumber := key.MaxNumber
	if key.AdaptiveDifficulty && h.Adaptive != nil {
		// Rate is tracked per (key, IP) so one abusive source is escalated
		// without penalizing other clients of the same key.
		rate := h.Adaptive.observe(key.KeyID + "|" + clientIP(r))
		maxNumber = adaptiveMaxNumber(key.MaxNumber, rate)
	}

	challenge, err := altcha.GenerateChallenge(key.HMACSecret, maxNumber, key.Algorithm, key.ExpireSeconds)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate challenge"})
		return
	}

	if err := models.IncrementChallengesIssued(h.DB, key.ID); err != nil {
		slog.Error("failed to increment challenges_issued", "error", err, "api_key_id", key.ID)
	}

	writeJSON(w, http.StatusOK, challenge)
}
