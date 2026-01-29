package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides per-IP rate limiting using token bucket algorithm.
type RateLimiter struct {
	limiters map[string]*clientLimiter
	mu       sync.RWMutex
	rate     rate.Limit // requests per second
	burst    int        // burst size
	cleanup  time.Duration
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// DefaultRateLimit is 500 requests per second sustained per IP.
// High limit for personal PaaS - allows legitimate high-traffic usage
// while still providing protection against malicious floods.
const DefaultRateLimit rate.Limit = 500

// DefaultBurst is 1000 requests burst per IP.
// Allows page loads with many assets without triggering limits.
const DefaultBurst = 1000

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(rps rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*clientLimiter),
		rate:     rps,
		burst:    burst,
		cleanup:  time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// cleanupLoop removes stale entries every minute.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	for range ticker.C {
		rl.mu.Lock()
		for ip, client := range rl.limiters {
			if time.Since(client.lastSeen) > 3*time.Minute {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request from the given IP should be allowed.
func (rl *RateLimiter) Allow(ip string) bool {
	// Fast path: check if limiter exists with read lock
	rl.mu.RLock()
	client, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if !exists {
		// Slow path: create new limiter with write lock
		rl.mu.Lock()
		// Double-check after acquiring write lock
		client, exists = rl.limiters[ip]
		if !exists {
			client = &clientLimiter{
				limiter:  rate.NewLimiter(rl.rate, rl.burst),
				lastSeen: time.Now(),
			}
			rl.limiters[ip] = client
		}
		rl.mu.Unlock()
	}

	// Update lastSeen (atomic operation on limiter is safe)
	client.lastSeen = time.Now()

	return client.limiter.Allow()
}

// Middleware returns an HTTP middleware that enforces rate limiting.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)

		if !rl.Allow(ip) {
			w.Header().Set("Retry-After", "1")
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.burst))
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ConnectionLimiter limits concurrent connections per IP.
type ConnectionLimiter struct {
	connections map[string]int
	mu          sync.Mutex
	maxPerIP    int
}

// DefaultMaxConnectionsPerIP is 200 concurrent connections per IP
const DefaultMaxConnectionsPerIP = 200

// NewConnectionLimiter creates a new connection limiter.
func NewConnectionLimiter(maxPerIP int) *ConnectionLimiter {
	return &ConnectionLimiter{
		connections: make(map[string]int),
		maxPerIP:    maxPerIP,
	}
}

// Acquire tries to acquire a connection slot for the given IP.
// Returns true if allowed, false if limit reached.
func (cl *ConnectionLimiter) Acquire(ip string) bool {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.connections[ip] >= cl.maxPerIP {
		return false
	}
	cl.connections[ip]++
	return true
}

// Release releases a connection slot for the given IP.
func (cl *ConnectionLimiter) Release(ip string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.connections[ip] > 0 {
		cl.connections[ip]--
	}
	if cl.connections[ip] == 0 {
		delete(cl.connections, ip)
	}
}

// Middleware returns an HTTP middleware that enforces connection limits.
func (cl *ConnectionLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)

		if !cl.Acquire(ip) {
			http.Error(w, "Too Many Connections", http.StatusServiceUnavailable)
			return
		}
		defer cl.Release(ip)

		next.ServeHTTP(w, r)
	})
}

// extractIP gets the client IP from the request.
// Checks X-Forwarded-For and X-Real-IP headers first (for proxies),
// then falls back to RemoteAddr.
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For (may contain multiple IPs)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (original client), before any comma
		for i, c := range xff {
			if c == ',' {
				return strings.TrimSpace(xff[:i])
			}
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
