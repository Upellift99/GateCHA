package api

import (
	"log/slog"

	"github.com/Upellift99/GateCHA/internal/his"
	"github.com/Upellift99/GateCHA/internal/models"
	"gorm.io/gorm"
)

// hisOutcome is the Monitor-mode result of scoring one signal sample. Callers
// must keep a nil outcome distinct from a zero Score: nil means nothing was
// scored, and an absent collector is not evidence of a human.
type hisOutcome struct {
	Score     float64
	Suspected bool
}

// recordHISMonitor scores an optional HIS signal sample and records it for the
// given key in Monitor mode: it logs the result and bumps the Monitor counters
// but never affects the verification outcome. A nil sample is a no-op returning
// nil, so the HIS feature degrades cleanly when a client ships no collector.
// When the key has opted into sampling, the raw observation is also persisted
// for later calibration of enforcement thresholds. The outcome is returned so
// handlers can report it back to the caller; acting on it stays their choice.
func recordHISMonitor(db *gorm.DB, key *models.APIKey, signals *his.Signals) *hisOutcome {
	if key == nil || signals == nil {
		return nil
	}
	score := his.Score(*signals)
	suspected := his.IsBotSuspected(score)
	slog.Info("his monitor sample", "api_key_id", key.ID, "score", score, "bot_suspected", suspected)
	if err := models.IncrementHISObservation(db, key.ID, suspected); err != nil {
		slog.Error("failed to increment his observation", "error", err, "api_key_id", key.ID)
	}
	if key.HISSampling {
		if err := models.CreateHISSample(db, hisSampleFrom(key.ID, *signals, score, suspected)); err != nil {
			slog.Error("failed to store his sample", "error", err, "api_key_id", key.ID)
		}
	}
	return &hisOutcome{Score: score, Suspected: suspected}
}

// hisSampleFrom maps a scored signal observation to its persistable row.
func hisSampleFrom(apiKeyID int64, s his.Signals, score float64, suspected bool) *models.HISSample {
	return &models.HISSample{
		APIKeyID:           apiKeyID,
		Score:              score,
		BotSuspected:       suspected,
		DurationMs:         s.DurationMs,
		TimeToFirstMs:      s.TimeToFirstMs,
		PointerEvents:      s.PointerEvents,
		PointerDistance:    s.PointerDistance,
		Scrolls:            s.Scrolls,
		Touches:            s.Touches,
		Keydowns:           s.Keydowns,
		KeyIntervalStdevMs: s.KeyIntervalStdevMs,
	}
}
