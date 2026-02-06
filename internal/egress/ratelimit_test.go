package egress

import (
	"testing"
)

func TestRateLimiterDisabledByDefault(t *testing.T) {
	rl := NewRateLimiter(0, 0) // 0 = no limit

	// Should always allow when rate is 0
	for i := 0; i < 100; i++ {
		if !rl.Allow("example.com", 0, 0) {
			t.Fatalf("Allow() returned false on call %d with rate=0", i+1)
		}
	}
}

func TestRateLimiterEnforces(t *testing.T) {
	rl := NewRateLimiter(60, 3) // 60 req/min (1/sec), burst 3

	// First 3 should pass (burst)
	for i := 0; i < 3; i++ {
		if !rl.Allow("example.com", 0, 0) {
			t.Fatalf("Allow() should pass for burst call %d", i+1)
		}
	}

	// Next should be rate limited
	if rl.Allow("example.com", 0, 0) {
		t.Error("Allow() should be rate limited after burst exhausted")
	}
}

func TestRateLimiterPerDomainOverride(t *testing.T) {
	rl := NewRateLimiter(0, 0) // System default: no limit

	// Per-domain rate: 60/min, burst 2
	for i := 0; i < 2; i++ {
		if !rl.Allow("limited.com", 60, 2) {
			t.Fatalf("Allow() should pass for burst call %d", i+1)
		}
	}

	// Should be limited now
	if rl.Allow("limited.com", 60, 2) {
		t.Error("Allow() should be rate limited after per-domain burst exhausted")
	}

	// Different domain should not be affected
	if !rl.Allow("unlimited.com", 0, 0) {
		t.Error("different domain should not be rate limited")
	}
}
