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

	return clamp01(score)
}

// IsBotSuspected reports whether the score meets the Monitor suspect threshold.
func IsBotSuspected(score float64) bool {
	return score >= BotSuspectThreshold
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
