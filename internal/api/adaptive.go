package api

import (
	"sync"
	"time"
)

// Adaptive-difficulty tuning. These are server-side constants rather than
// per-key config to keep the admin surface to a single opt-in toggle.
const (
	// adaptiveThreshold is the challenge-request count per fixed one-minute
	// window, per (key, IP), below which difficulty stays at the key's base.
	adaptiveThreshold = 60
	// adaptiveMaxFactor caps how many times the base MaxNumber may be multiplied.
	adaptiveMaxFactor = 8
	// adaptiveHardCap is the absolute ceiling on the scaled MaxNumber, matching
	// the maximum the admin UI allows for the base difficulty.
	adaptiveHardCap = 10_000_000
)

// adaptiveCounter tracks how many challenge requests a (key, IP) made in the
// current fixed one-minute window.
type adaptiveCounter struct {
	count       int
	windowStart time.Time
}

// adaptiveLimiter estimates per-(key, IP) challenge-request rate using fixed
// one-minute windows. It is used to escalate proof-of-work difficulty for
// abusive sources; it never blocks. Stale counters are evicted lazily.
type adaptiveLimiter struct {
	mu        sync.Mutex
	counters  map[string]*adaptiveCounter
	ttl       time.Duration
	lastSweep time.Time
	now       func() time.Time // injectable for tests
}

func newAdaptiveLimiter() *adaptiveLimiter {
	return &adaptiveLimiter{
		counters: make(map[string]*adaptiveCounter),
		ttl:      10 * time.Minute,
		now:      time.Now,
	}
}

// observe records one challenge request for key and returns the request count in
// the active one-minute window (including this one).
func (a *adaptiveLimiter) observe(key string) int {
	now := a.now()

	a.mu.Lock()
	defer a.mu.Unlock()

	a.sweep(now)

	c, ok := a.counters[key]
	if !ok || now.Sub(c.windowStart) >= time.Minute {
		a.counters[key] = &adaptiveCounter{count: 1, windowStart: now}
		return 1
	}
	c.count++
	return c.count
}

// sweep evicts counters untouched for longer than ttl. Caller must hold the lock.
func (a *adaptiveLimiter) sweep(now time.Time) {
	if now.Sub(a.lastSweep) < a.ttl {
		return
	}
	a.lastSweep = now
	for key, c := range a.counters {
		if now.Sub(c.windowStart) > a.ttl {
			delete(a.counters, key)
		}
	}
}

// adaptiveMaxNumber scales base upward once ratePerMin exceeds adaptiveThreshold.
// The multiplier is ceil(ratePerMin/threshold), clamped to adaptiveMaxFactor, and
// the result is capped at adaptiveHardCap. base is always the floor.
func adaptiveMaxNumber(base int64, ratePerMin int) int64 {
	if ratePerMin <= adaptiveThreshold {
		return base
	}
	factor := int64((ratePerMin + adaptiveThreshold - 1) / adaptiveThreshold) // ceil
	if factor > adaptiveMaxFactor {
		factor = adaptiveMaxFactor
	}
	scaled := base * factor
	if scaled > adaptiveHardCap {
		scaled = adaptiveHardCap
	}
	return scaled
}
