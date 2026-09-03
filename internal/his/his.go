// Package his implements a lightweight, self-hosted "Human Interaction
// Signature": it scores aggregate client interaction signals to estimate the
// probability that a verification was automated.
//
// This is GateCHA's open-source take on the concept behind ALTCHA Sentinel's
// proprietary HIS. It deliberately consumes only privacy-preserving aggregates
// (counts, distances, durations, timing variance) — never raw coordinates,
// timestamps, or key contents — so it leaks nothing about what the user typed
// or where they pointed.
//
// The score is advisory. In "Monitor" mode callers log it and move on; they
// never change a verification outcome based on it.
package his

import "math"

// Signals is the privacy-preserving aggregate of a single interaction window,
// produced by the client collector. All fields are optional; a zero value means
// "not observed". Negative sentinels (-1) distinguish "no interaction happened"
// from "happened at t=0" for the timing fields.
type Signals struct {
	// DurationMs is how long the interaction window was observed.
	DurationMs int `json:"duration_ms"`
	// TimeToFirstMs is the delay until the first interaction event; -1 if none.
	TimeToFirstMs int `json:"time_to_first_ms"`
	// PointerEvents counts sampled pointer/mouse move events.
	PointerEvents int `json:"pointer_events"`
	// PointerDistance is the total pointer path length in CSS pixels.
	PointerDistance float64 `json:"pointer_distance"`
	// Scrolls counts scroll events.
	Scrolls int `json:"scrolls"`
	// Touches counts touch events.
	Touches int `json:"touches"`
	// Keydowns counts key-down events (host forms only; 0 for a bare widget).
	Keydowns int `json:"keydowns"`
	// KeyIntervalStdevMs is the standard deviation of inter-keydown intervals;
	// genuine typing has natural jitter, scripted input tends toward 0.
	KeyIntervalStdevMs float64 `json:"key_interval_stdev_ms"`
}

// Heuristic weights and thresholds. These are intentionally simple and
// transparent; Monitor mode exists precisely to calibrate them against real
// traffic before any enforcement is considered. Tuning here is expected.
const (
	// BotSuspectThreshold is the probability at/above which a sample is counted
	// as automation-suspected for Monitor statistics.
	BotSuspectThreshold = 0.8

	wNoMotion      = 0.50 // no pointer/scroll/touch motion at all
	wNoPointerPath = 0.20 // solved without any pointer travel or touch
	wInstantSolve  = 0.20 // window closed almost immediately
	wInstantFirst  = 0.10 // first interaction landed implausibly fast
	mPointerTravel = 0.30 // credit for meaningful pointer travel
	mKeyJitter     = 0.20 // credit for natural typing jitter

	// noEvidenceScore is what a sample carrying no data at all is worth: bot-ward,
	// but short of BotSuspectThreshold, since a missing collector is not proof of
	// automation. Equals wNoMotion + wNoPointerPath by construction.
	noEvidenceScore = 0.70

	instantSolveMs  = 200 // below this, a solve is suspiciously fast
	instantFirstMs  = 50  // below this, the first event is suspiciously early
	pointerTravelPx = 50.0
	keyJitterMs     = 5.0
)

// Score returns the estimated probability in [0,1] that the interaction
// described by s was automated. Higher means more bot-like. Missing signals
// (a zero value) lean bot-ward but never reach certainty on their own, since a
// missing collector is not proof of automation.
func Score(s Signals) float64 {
	// A wholly empty sample states one fact, not four: nothing was observed.
	// Charging it every penalty in turn made an absent collector the single most
	// damning signal the heuristic can produce, which contradicts the paragraph
	// above. It only ever stayed under 1.0 by float rounding error. It stops at
	// the no-evidence tier instead, the same score as a sample that reported
	// real timings but no motion at all, and like that one it stays below
	// BotSuspectThreshold.
	if s == (Signals{}) {
		return noEvidenceScore
	}

	score := 0.0

	motion := s.PointerEvents + s.Scrolls + s.Touches
	if motion == 0 {
		score += wNoMotion
	}
	if s.PointerDistance == 0 && s.Touches == 0 {
		score += wNoPointerPath
	}
	// DurationMs <= 0 means the window was never really open; treat as instant.
	if s.DurationMs >= 0 && s.DurationMs < instantSolveMs {
		score += wInstantSolve
	}
	if s.TimeToFirstMs >= 0 && s.TimeToFirstMs < instantFirstMs {
		score += wInstantFirst
	}

	// Human-leaning mitigations.
	if s.PointerDistance > pointerTravelPx {
		score -= mPointerTravel
	}
	if s.KeyIntervalStdevMs > keyJitterMs {
		score -= mKeyJitter
	}

	return round2(clamp01(score))
}

// round2 snaps the score to two decimals. Every weight above is a multiple of
// 0.1, so the heuristic is a ladder in tenths by construction, but summing them
// as float64 lands on values like 0.8999999999999999. Left alone that leaks out
// of the API and quietly breaks the obvious integration ("score >= 0.9"), and it
// misfiles samples one bucket low in the calibration histogram.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// IsBotSuspected reports whether the score meets the default suspect threshold.
func IsBotSuspected(score float64) bool {
	return IsBotSuspectedAt(score, BotSuspectThreshold)
}

// IsBotSuspectedAt reports whether the score meets a caller-chosen threshold.
//
// Comparison is on scores already snapped to two decimals by Score, and the
// threshold arrives from configuration, so a value like 0.7 that has no exact
// binary representation must not decide the comparison by its last bit: a
// sample scoring exactly the threshold has to count as meeting it. The epsilon
// is far below the 0.01 granularity of the score ladder, so it can never pull
// in a neighbouring rung.
func IsBotSuspectedAt(score, threshold float64) bool {
	return score+1e-9 >= threshold
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
