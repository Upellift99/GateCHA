package his

import "testing"

func TestScore_HumanLikeIsLow(t *testing.T) {
	// Real pointer travel over a sensible window, with typing jitter.
	s := Signals{
		DurationMs:         3500,
		TimeToFirstMs:      400,
		PointerEvents:      42,
		PointerDistance:    640,
		Scrolls:            2,
		Keydowns:           12,
		KeyIntervalStdevMs: 35,
	}
	got := Score(s)
	if got >= 0.3 {
		t.Errorf("human-like interaction should score low, got %.2f", got)
	}
	if IsBotSuspected(got) {
		t.Errorf("human-like interaction must not be bot-suspected, got %.2f", got)
	}
}

func TestScore_NoMotionInstantIsHigh(t *testing.T) {
	// Solved instantly with zero motion of any kind — classic headless bot.
	s := Signals{
		DurationMs:    10,
		TimeToFirstMs: -1, // never interacted
	}
	got := Score(s)
	if !IsBotSuspected(got) {
		t.Errorf("no-motion instant solve should be bot-suspected, got %.2f", got)
	}
}

func TestScore_EmptySignalsLeanBotButNotCertain(t *testing.T) {
	// A completely zero value (collector missing) should lean bot-ward without
	// reaching certainty — absence of data is not proof of automation.
	got := Score(Signals{})
	if got <= 0 {
		t.Errorf("empty signals should lean bot-ward, got %.2f", got)
	}
	if got >= 1 {
		t.Errorf("empty signals must not be certain, got %.2f", got)
	}
}

func TestScore_IsClampedTo01(t *testing.T) {
	// Maximal bot evidence must not exceed 1.
	high := Score(Signals{DurationMs: 0, TimeToFirstMs: 0})
	if high < 0 || high > 1 {
		t.Errorf("score out of range: %.2f", high)
	}
	// Strong human evidence must not go below 0.
	low := Score(Signals{
		DurationMs:         5000,
		TimeToFirstMs:      800,
		PointerEvents:      60,
		PointerDistance:    2000,
		KeyIntervalStdevMs: 50,
	})
	if low < 0 || low > 1 {
		t.Errorf("score out of range: %.2f", low)
	}
}

func TestScore_PointerTravelMitigatesNoMotionAbsence(t *testing.T) {
	// Pointer travel present but a fast window: travel credit should keep it
	// below the suspect threshold.
	s := Signals{DurationMs: 150, TimeToFirstMs: 60, PointerEvents: 20, PointerDistance: 300}
	if got := Score(s); IsBotSuspected(got) {
		t.Errorf("meaningful pointer travel should avoid bot suspicion, got %.2f", got)
	}
}
