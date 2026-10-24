package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Upellift99/GateCHA/internal/models"
	"github.com/Upellift99/GateCHA/internal/testutil"
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

func TestKeyRateLimiter_PerCallLimit(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now, _ := fakeClock(base)
	l := newKeyRateLimiter()
	l.now = now

	for i := 0; i < 3; i++ {
		if !l.allow("gk_a", 3) {
			t.Fatalf("request %d should be allowed within the per-key budget", i+1)
		}
	}
	if l.allow("gk_a", 3) {
		t.Fatal("4th request should be blocked once the budget is exhausted")
	}
	// A different key has its own independent bucket.
	if !l.allow("gk_b", 3) {
		t.Fatal("a different key must not share gk_a's bucket")
	}
}

func TestKeyRateLimiter_ZeroMeansUnlimited(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now, _ := fakeClock(base)
	l := newKeyRateLimiter()
	l.now = now

	for i := 0; i < 1000; i++ {
		if !l.allow("gk_unlimited", 0) {
			t.Fatalf("limit 0 should never throttle (request %d blocked)", i+1)
		}
	}
	if len(l.buckets) != 0 {
		t.Errorf("unlimited keys should not allocate buckets, got %d", len(l.buckets))
	}
}

func TestKeyRateLimiter_RefillsOverTime(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now, advance := fakeClock(base)
	l := newKeyRateLimiter()
	l.now = now

	for i := 0; i < 60; i++ {
		l.allow("gk_x", 60) // 60/min => 1 token/sec
	}
	if l.allow("gk_x", 60) {
		t.Fatal("should be blocked after consuming the full burst")
	}
	advance(2 * time.Second)
	if !l.allow("gk_x", 60) {
		t.Fatal("should be allowed again after refill")
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

// limitedRouter is the full router with the per-IP limit on and set to n.
func limitedRouter(t *testing.T, n int) (http.Handler, string, string) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	a, _ := models.CreateAPIKey(db, "site A", "", 0, 0, "")
	b, _ := models.CreateAPIKey(db, "site B", "", 0, 0, "")
	router := NewRouter(db, testSecretKey, RouterConfig{RateLimitEnabled: true, RateLimitAPI: n, RateLimitLogin: 5})
	return router, a.KeyID, b.KeyID
}

func hit(router http.Handler, method, path, ip string) int {
	req := httptest.NewRequest(method, path, strings.NewReader(`{"payload":"x"}`))
	req.RemoteAddr = ip + ":1234"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Code
}

// Sites on one shared-hosting IP must not throttle each other on /verify, yet
// each still gets no more than the per-IP budget from that IP.
func TestVerifyRateLimit_PerSourceAndKey(t *testing.T) {
	router, keyA, keyB := limitedRouter(t, 3)
	const host = "203.0.113.7"

	for i := range 3 {
		if code := hit(router, "POST", "/api/v1/verify?apiKey="+keyA, host); code == http.StatusTooManyRequests {
			t.Fatalf("site A request %d throttled early", i+1)
		}
	}
	if code := hit(router, "POST", "/api/v1/verify?apiKey="+keyA, host); code != http.StatusTooManyRequests {
		t.Errorf("site A over budget: expected 429, got %d", code)
	}
	if code := hit(router, "POST", "/api/v1/verify?apiKey="+keyB, host); code == http.StatusTooManyRequests {
		t.Error("site B on the same IP was throttled by site A's traffic")
	}
	if code := hit(router, "POST", "/api/v1/verify?apiKey="+keyA, "198.51.100.9"); code == http.StatusTooManyRequests {
		t.Error("site A from another IP was throttled")
	}
}

// Challenges come from browsers: the budget stays per IP, whatever the key.
func TestChallengeRateLimit_StaysPerIP(t *testing.T) {
	router, keyA, keyB := limitedRouter(t, 3)
	const visitor = "192.0.2.44"

	for range 3 {
		hit(router, "GET", "/api/v1/challenge?apiKey="+keyA, visitor)
	}
	if code := hit(router, "GET", "/api/v1/challenge?apiKey="+keyB, visitor); code != http.StatusTooManyRequests {
		t.Errorf("challenge over the per-IP budget with another key: expected 429, got %d", code)
	}
	// Verifications do not draw on the challenge budget.
	if code := hit(router, "POST", "/api/v1/verify?apiKey="+keyA, visitor); code == http.StatusTooManyRequests {
		t.Error("verify throttled by the challenge budget")
	}
}

// Made-up keys are refused (401) and, past the per-IP budget, stopped (429)
// before any key lookup; a valid key from the same IP is not affected.
func TestVerifyRateLimit_FailedAuthCappedPerIP(t *testing.T) {
	router, keyA, _ := limitedRouter(t, 3)
	const attacker = "203.0.113.8"

	for i := range 3 {
		if code := hit(router, "POST", "/api/v1/verify?apiKey=gk_000000000000000000000000", attacker); code != http.StatusUnauthorized {
			t.Fatalf("request %d: expected 401, got %d", i+1, code)
		}
	}
	if code := hit(router, "POST", "/api/v1/verify?apiKey=gk_111111111111111111111111", attacker); code != http.StatusTooManyRequests {
		t.Errorf("failed-auth budget spent: expected 429, got %d", code)
	}
	if code := hit(router, "POST", "/api/v1/verify?apiKey=gk_000000000000000000000000", "198.51.100.20"); code != http.StatusUnauthorized {
		t.Errorf("another IP: expected 401, got %d", code)
	}

	// Valid keys never draw on the failed-auth budget.
	router, keyA, _ = limitedRouter(t, 3)
	for range 3 {
		hit(router, "POST", "/api/v1/verify?apiKey="+keyA, attacker)
	}
	if code := hit(router, "POST", "/api/v1/verify?apiKey=gk_000000000000000000000000", attacker); code != http.StatusUnauthorized {
		t.Errorf("valid traffic consumed the failed-auth budget: expected 401, got %d", code)
	}
}
