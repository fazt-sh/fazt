package auth

import (
	"sync"
	"time"
)

// RateLimiter tracks failed login attempts by IP address
type RateLimiter struct {
	attempts map[string]*loginAttempts
	mu       sync.RWMutex
}

type loginAttempts struct {
	count      int
	firstAttempt time.Time
	lastAttempt  time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	limiter := &RateLimiter{
		attempts: make(map[string]*loginAttempts),
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// AllowLogin checks if a login attempt is allowed for an IP
func (rl *RateLimiter) AllowLogin(ip string) bool {
	rl.mu.RLock()
	attempts, exists := rl.attempts[ip]
	rl.mu.RUnlock()

	if !exists {
		return true
	}

	// Reset if more than 15 minutes have passed
	if time.Since(attempts.firstAttempt) > 15*time.Minute {
		rl.Reset(ip)
		return true
	}

	// Allow if less than 5 attempts
	return attempts.count < 5
}

// RecordAttempt records a failed login attempt
func (rl *RateLimiter) RecordAttempt(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	attempts, exists := rl.attempts[ip]
	if !exists {
		rl.attempts[ip] = &loginAttempts{
			count:        1,
			firstAttempt: time.Now(),
			lastAttempt:  time.Now(),
		}
		return
	}

	// Reset if more than 15 minutes have passed
	if time.Since(attempts.firstAttempt) > 15*time.Minute {
		attempts.count = 1
		attempts.firstAttempt = time.Now()
	} else {
		attempts.count++
	}

	attempts.lastAttempt = time.Now()
}

// Reset clears attempts for an IP (called on successful login)
func (rl *RateLimiter) Reset(ip string) {
	rl.mu.Lock()
	delete(rl.attempts, ip)
	rl.mu.Unlock()
}

// GetAttempts returns the number of attempts for an IP
func (rl *RateLimiter) GetAttempts(ip string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	attempts, exists := rl.attempts[ip]
	if !exists {
		return 0
	}

	return attempts.count
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, attempts := range rl.attempts {
			if time.Since(attempts.firstAttempt) > 15*time.Minute {
				delete(rl.attempts, ip)
			}
		}
		rl.mu.Unlock()
	}
}
