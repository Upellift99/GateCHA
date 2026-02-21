package api

import (
	"database/sql"
	"net/http"

	"github.com/Upellift99/GateCHA/internal/altcha"
	"github.com/Upellift99/GateCHA/internal/models"
)

type ChallengeHandler struct {
	DB *sql.DB
}

func (h *ChallengeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := GetAPIKeyFromContext(r)
	if key == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing API key"})
		return
	}

	challenge, err := altcha.GenerateChallenge(key.HMACSecret, key.MaxNumber, key.Algorithm, key.ExpireSeconds)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate challenge"})
		return
	}

	_ = models.IncrementChallengesIssued(h.DB, key.ID)

	writeJSON(w, http.StatusOK, challenge)
}
