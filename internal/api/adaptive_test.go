package api

import (
	"testing"
	"time"
)

func TestAdaptiveMaxNumber_ScalesWithRate(t *testing.T) {
	base := int64(1000)
	tests := []struct {
		rate int
		want int64
	}{
		{0, 1000},                       // idle
		{adaptiveThreshold, 1000},       // at threshold => base
		{adaptiveThreshold + 1, 2000},   // just over => factor 2
		{adaptiveThreshold * 2, 2000},   // exactly 2x threshold => factor 2
		{adaptiveThreshold*2 + 1, 3000}, // => factor 3
		{adaptiveThreshold * 100, 8000}, // clamped at adaptiveMaxFactor (8)
	}
	for _, tt := range tests {
		if got := adaptiveMaxNumber(base, tt.rate); got != tt.want {
			t.Errorf("adaptiveMaxNumber(%d, %d) = %d, want %d", base, tt.rate, got, tt.want)
		}
	}
}

func TestAdaptiveMaxNumber_HonorsHardCap(t *testing.T) {
	// A high base with a high factor must not exceed the absolute ceiling.
	got := adaptiveMaxNumber(5_000_000, adaptiveThreshold*8)
	if got != adaptiveHardCap {
		t.Errorf("expected hard cap %d, got %d", adaptiveHardCap, got)
	}
}

func TestAdaptiveLimiter_CountsWithinWindow(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now, advance := fakeClock(base)
	a := newAdaptiveLimiter()
	a.now = now

	for i := 1; i <= 5; i++ {
		if got := a.observe("k|ip"); got != i {
			t.Fatalf("observe #%d returned %d, want %d", i, got, i)
		}
	}

	// A separate (key, IP) has its own counter.
	if got := a.observe("k|other"); got != 1 {
		t.Fatalf("distinct source should start at 1, got %d", got)
	}

	// Crossing the one-minute boundary resets the window.
	advance(time.Minute)
	if got := a.observe("k|ip"); got != 1 {
		t.Fatalf("window should reset after a minute, got %d", got)
	}
}

func TestAdaptiveLimiter_SweepEvictsStale(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now, advance := fakeClock(base)
	a := newAdaptiveLimiter()
	a.now = now

	a.observe("stale|ip")
	advance(a.ttl + time.Minute)
	a.observe("fresh|ip")

	if _, ok := a.counters["stale|ip"]; ok {
		t.Error("stale counter should have been evicted")
	}
	if _, ok := a.counters["fresh|ip"]; !ok {
		t.Error("fresh counter should be tracked")
	}
}
