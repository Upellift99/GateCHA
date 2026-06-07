package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// fakeClock returns a controllable time source for deterministic limiter tests.
func fakeClock(start time.Time) (func() time.Time, func(d time.Duration)) {
	cur := start
	now := func() time.Time { return cur }
	advance := func(d time.Duration) { cur = cur.Add(d) }
	return now, advance
}

func TestIPRateLimiter_AllowsBurstThenBlocks(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now, _ := fakeClock(base)
	l := newIPRateLimiter(5) // 5 per minute
	l.now = now

	for i := 0; i < 5; i++ {
		if !l.allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed within burst", i+1)
		}
	}
	if l.allow("1.2.3.4") {
		t.Fatal("6th request should be blocked once the burst is exhausted")
	}
}

func TestIPRateLimiter_RefillsOverTime(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now, advance := fakeClock(base)
	l := newIPRateLimiter(60) // 60/min => 1 token per second
	l.now = now

	for i := 0; i < 60; i++ {
		l.allow("ip")
	}
	if l.allow("ip") {
		t.Fatal("should be blocked after consuming the full burst")
	}

	advance(2 * time.Second) // refills ~2 tokens
	if !l.allow("ip") {
		t.Fatal("should be allowed again after refill")
	}
}

func TestIPRateLimiter_PerIPIsolation(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now, _ := fakeClock(base)
	l := newIPRateLimiter(1)
	l.now = now

	if !l.allow("a") {
		t.Fatal("first request from a should pass")
	}
	if l.allow("a") {
		t.Fatal("second request from a should be blocked")
	}
	if !l.allow("b") {
		t.Fatal("a different IP must have its own bucket")
	}
}

func TestIPRateLimiter_SweepEvictsStale(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now, advance := fakeClock(base)
	l := newIPRateLimiter(5)
	l.now = now

	l.allow("stale")
	if len(l.visitors) != 1 {
		t.Fatalf("expected 1 tracked visitor, got %d", len(l.visitors))
	}

	advance(l.ttl + time.Minute) // past TTL so the next sweep evicts it
	l.allow("fresh")

	if _, ok := l.visitors["stale"]; ok {
		t.Error("stale visitor should have been evicted")
	}
	if _, ok := l.visitors["fresh"]; !ok {
		t.Error("fresh visitor should be tracked")
	}
}

func TestRateLimitMiddleware_Returns429(t *testing.T) {
	handler := RateLimitMiddleware(1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	call := func() int {
		req := httptest.NewRequest("POST", "/api/admin/login", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w.Code
	}

	if code := call(); code != http.StatusOK {
		t.Fatalf("first request should pass, got %d", code)
	}
	if code := call(); code != http.StatusTooManyRequests {
		t.Fatalf("second request should be rate limited, got %d", code)
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		remoteAddr, want string
	}{
		{"1.2.3.4:5678", "1.2.3.4"},
		{"[::1]:8080", "::1"},
		{"noport", "noport"},
	}
	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = tt.remoteAddr
		if got := clientIP(req); got != tt.want {
			t.Errorf("clientIP(%q) = %q, want %q", tt.remoteAddr, got, tt.want)
		}
	}
}
