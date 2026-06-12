package api

import (
	"log/slog"

	"github.com/Upellift99/GateCHA/internal/his"
	"github.com/Upellift99/GateCHA/internal/models"
	"gorm.io/gorm"
)

// recordHISMonitor scores an optional HIS signal sample and records it for the
// given key in Monitor mode: it logs the result and bumps the Monitor counters
// but never affects the verification outcome. A nil sample is a no-op, so the
// HIS feature degrades cleanly when a client ships no collector.
func recordHISMonitor(db *gorm.DB, apiKeyID int64, signals *his.Signals) {
	if signals == nil {
		return
	}
	score := his.Score(*signals)
	suspected := his.IsBotSuspected(score)
	slog.Info("his monitor sample", "api_key_id", apiKeyID, "score", score, "bot_suspected", suspected)
	if err := models.IncrementHISObservation(db, apiKeyID, suspected); err != nil {
		slog.Error("failed to increment his observation", "error", err, "api_key_id", apiKeyID)
	}
}
