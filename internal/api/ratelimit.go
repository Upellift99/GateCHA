package api

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// visitor tracks the token-bucket state for a single client IP.
type visitor struct {
	tokens   float64
	lastSeen time.Time
}

// ipRateLimiter is a small, dependency-free, per-IP token-bucket limiter.
//
// Each IP gets a bucket that refills at ratePerSec tokens per second up to a
// maximum of burst tokens; a request is allowed when at least one token is
// available. Stale buckets are evicted lazily to keep memory bounded, so no
// background goroutine is required.
type ipRateLimiter struct {
	mu         sync.Mutex
	visitors   map[string]*visitor
	ratePerSec float64
	burst      float64
	ttl        time.Duration
	lastSweep  time.Time
	now        func() time.Time // injectable for tests
}

// newIPRateLimiter builds a limiter allowing requestsPerMinute requests per IP,
// bursting up to a full minute's worth.
func newIPRateLimiter(requestsPerMinute int) *ipRateLimiter {
	if requestsPerMinute < 1 {
		requestsPerMinute = 1
	}
	return &ipRateLimiter{
		visitors:   make(map[string]*visitor),
		ratePerSec: float64(requestsPerMinute) / 60.0,
		burst:      float64(requestsPerMinute),
		ttl:        10 * time.Minute,
		now:        time.Now,
	}
}

// allow reports whether a request from ip may proceed, consuming a token if so.
func (l *ipRateLimiter) allow(ip string) bool {
	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.sweep(now)

	v, ok := l.visitors[ip]
	if !ok {
		// First request from this IP: full bucket minus the one we consume now.
		l.visitors[ip] = &visitor{tokens: l.burst - 1, lastSeen: now}
		return true
	}

	// Refill based on elapsed time, capped at burst.
	elapsed := now.Sub(v.lastSeen).Seconds()
	v.tokens = min(l.burst, v.tokens+elapsed*l.ratePerSec)
	v.lastSeen = now

	if v.tokens < 1 {
		return false
	}
	v.tokens--
	return true
}

// sweep evicts buckets untouched for longer than ttl. Caller must hold the lock.
func (l *ipRateLimiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < l.ttl {
		return
	}
	l.lastSweep = now
	for ip, v := range l.visitors {
		if now.Sub(v.lastSeen) > l.ttl {
			delete(l.visitors, ip)
		}
	}
}

// clientIP extracts the client IP from RemoteAddr, which chi's RealIP middleware
// populates from trusted proxy headers. It is therefore only as trustworthy as
// the upstream reverse proxy.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimitMiddleware limits requests per client IP over a one-minute window and
// answers with a JSON 429 when the limit is exceeded.
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	limiter := newIPRateLimiter(requestsPerMinute)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.allow(clientIP(r)) {
				writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
