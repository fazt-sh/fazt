package egress

import (
	"sync"
	"time"
)

// tokenBucket implements a simple token bucket rate limiter.
type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
	rate       float64 // tokens per second
	burst      int
}

func newTokenBucket(ratePerMin, burst int) *tokenBucket {
	return &tokenBucket{
		tokens:     float64(burst),
		lastRefill: time.Now(),
		rate:       float64(ratePerMin) / 60.0, // convert to per-second
		burst:      burst,
	}
}

func (b *tokenBucket) allow() bool {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.lastRefill = now

	// Refill tokens
	b.tokens += elapsed * b.rate
	if b.tokens > float64(b.burst) {
		b.tokens = float64(b.burst)
	}

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// RateLimiter enforces per-domain rate limits using token buckets.
type RateLimiter struct {
	buckets     map[string]*tokenBucket
	mu          sync.Mutex
	defaultRate int // From Limits.Net.RateLimit (req/min, 0 = no limit)
	defaultBurst int // From Limits.Net.RateBurst
}

// NewRateLimiter creates a RateLimiter with the given system defaults.
func NewRateLimiter(defaultRate, defaultBurst int) *RateLimiter {
	return &RateLimiter{
		buckets:      make(map[string]*tokenBucket),
		defaultRate:  defaultRate,
		defaultBurst: defaultBurst,
	}
}

// Allow returns true if a request to the given domain should be allowed.
// domainRate/domainBurst are per-domain overrides from the allowlist (0 = use default).
func (rl *RateLimiter) Allow(domain string, domainRate, domainBurst int) bool {
	rate := rl.defaultRate
	burst := rl.defaultBurst
	if domainRate > 0 {
		rate = domainRate
	}
	if domainBurst > 0 {
		burst = domainBurst
	}

	// Rate 0 means no limit
	if rate == 0 {
		return true
	}

	// Ensure burst is at least 1
	if burst == 0 {
		burst = 1
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, ok := rl.buckets[domain]
	if !ok {
		bucket = newTokenBucket(rate, burst)
		rl.buckets[domain] = bucket
	}

	return bucket.allow()
}
