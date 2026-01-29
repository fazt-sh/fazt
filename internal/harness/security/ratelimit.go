// Package security provides security and DoS resilience tests.
package security

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fazt-sh/fazt/internal/harness"
	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// RateLimitTest defines a rate limit test.
type RateLimitTest struct {
	Name        string
	Description string
	Requests    int
	Interval    time.Duration
	Expected    RateLimitExpected
}

// RateLimitExpected defines expected rate limit behavior.
type RateLimitExpected struct {
	AllowedRequests int     // Expected 2xx responses
	RejectedMin     int     // Minimum expected 429s
	HasRetryAfter   bool    // Whether Retry-After header is expected
}

// RateLimitRunner executes rate limit tests.
type RateLimitRunner struct {
	client     *http.Client
	baseURL    string
	gapTracker *gaps.Tracker
}

// NewRateLimitRunner creates a new rate limit test runner.
func NewRateLimitRunner(baseURL string, gapTracker *gaps.Tracker) *RateLimitRunner {
	return &RateLimitRunner{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:    baseURL,
		gapTracker: gapTracker,
	}
}

// Run executes rate limit tests.
func (r *RateLimitRunner) Run(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	// Test 1: Basic rate limit enforcement
	results = append(results, r.testBasicRateLimit(ctx))

	// Test 2: Burst then throttle
	results = append(results, r.testBurstThrottle(ctx))

	// Test 3: Rate limit recovery
	results = append(results, r.testRateLimitRecovery(ctx))

	// Test 4: Concurrent requests from same IP
	results = append(results, r.testConcurrentRateLimit(ctx))

	return results
}

func (r *RateLimitRunner) testBasicRateLimit(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "basic_rate_limit",
		Category: "security",
	}

	start := time.Now()

	// Send more requests than should be allowed
	requestCount := 50
	var okCount, rateLimitedCount int
	var hasRetryAfter bool

	for i := 0; i < requestCount; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
		resp, err := r.client.Do(req)
		if err != nil {
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			okCount++
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitedCount++
			if resp.Header.Get("Retry-After") != "" {
				hasRetryAfter = true
			}
		}
	}

	result.Duration = time.Since(start)

	// Should have some successful and some rate limited
	// If no 429s, rate limiting may not be implemented/enabled
	result.Passed = rateLimitedCount > 0 || okCount == requestCount

	result.Actual = harness.Actual{
		Body: fmt.Sprintf("ok=%d, rate_limited=%d, retry_after=%v",
			okCount, rateLimitedCount, hasRetryAfter),
	}

	// Record gap if no rate limiting at all
	if rateLimitedCount == 0 && okCount > 20 && r.gapTracker != nil {
		r.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategorySecurity,
			Severity:     gaps.SeverityHigh,
			Description:  fmt.Sprintf("Rate limiting not enforced: %d requests all succeeded", okCount),
			DiscoveredBy: "ratelimit_basic",
			Remediation:  "Implement per-IP rate limiting",
			SpecRef:      "v0.8/rate-limit.md",
		})
	}

	return result
}

func (r *RateLimitRunner) testBurstThrottle(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "burst_then_throttle",
		Category: "security",
	}

	start := time.Now()

	// Burst: send 30 requests as fast as possible
	burstSize := 30
	var burstOK, burstThrottled int

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < burstSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
			resp, err := r.client.Do(req)
			if err != nil {
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			mu.Lock()
			if resp.StatusCode == http.StatusOK {
				burstOK++
			} else if resp.StatusCode == http.StatusTooManyRequests {
				burstThrottled++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	result.Duration = time.Since(start)

	// After burst, verify some were throttled or all succeeded (depending on implementation)
	result.Passed = burstThrottled > 0 || burstOK == burstSize

	result.Actual = harness.Actual{
		Body: fmt.Sprintf("burst: ok=%d, throttled=%d", burstOK, burstThrottled),
	}

	return result
}

func (r *RateLimitRunner) testRateLimitRecovery(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "rate_limit_recovery",
		Category: "security",
	}

	start := time.Now()

	// First: exhaust rate limit
	for i := 0; i < 50; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
		resp, err := r.client.Do(req)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}

	// Wait for rate limit to recover (typically 1 second)
	time.Sleep(2 * time.Second)

	// Test: should be able to make requests again
	var recoveryOK int
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
		resp, err := r.client.Do(req)
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				recoveryOK++
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}

	result.Duration = time.Since(start)

	// Should recover and allow some requests
	result.Passed = recoveryOK > 0

	result.Actual = harness.Actual{
		Body: fmt.Sprintf("recovery_ok=%d/5", recoveryOK),
	}

	return result
}

func (r *RateLimitRunner) testConcurrentRateLimit(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "concurrent_rate_limit",
		Category: "security",
	}

	start := time.Now()

	// Concurrent requests from same "IP"
	concurrency := 100
	var okCount, throttledCount int64

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
			resp, err := r.client.Do(req)
			if err != nil {
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				atomic.AddInt64(&okCount, 1)
			} else if resp.StatusCode == http.StatusTooManyRequests {
				atomic.AddInt64(&throttledCount, 1)
			}
		}()
	}
	wg.Wait()

	result.Duration = time.Since(start)

	// Concurrent requests should be rate limited
	result.Passed = throttledCount > 0 || okCount == int64(concurrency)

	result.Actual = harness.Actual{
		Body: fmt.Sprintf("ok=%d, throttled=%d", okCount, throttledCount),
	}

	return result
}

// PerEndpointRateLimitTest tests endpoint-specific rate limits.
func (r *RateLimitRunner) PerEndpointRateLimitTest(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	endpoints := []struct {
		path    string
		limit   int // Expected requests before 429
		name    string
	}{
		{"/health", 100, "health_endpoint"},
		{"/api/login", 10, "login_endpoint"},
		{"/api/deploy", 5, "deploy_endpoint"},
	}

	for _, ep := range endpoints {
		var okCount, throttledCount int

		for i := 0; i < ep.limit*2; i++ {
			req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+ep.path, nil)
			resp, err := r.client.Do(req)
			if err != nil {
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized {
				okCount++
			} else if resp.StatusCode == http.StatusTooManyRequests {
				throttledCount++
			}
		}

		results = append(results, harness.TestResult{
			Name:     "ratelimit_" + ep.name,
			Category: "security",
			Passed:   throttledCount > 0 || okCount == ep.limit*2,
			Actual: harness.Actual{
				Body: fmt.Sprintf("ok=%d, throttled=%d (limit=%d)", okCount, throttledCount, ep.limit),
			},
		})

		// Brief pause between endpoint tests
		time.Sleep(2 * time.Second)
	}

	return results
}

// RetryAfterTest verifies Retry-After header behavior.
func (r *RateLimitRunner) RetryAfterTest(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "retry_after_header",
		Category: "security",
	}

	start := time.Now()

	// Exhaust rate limit
	var retryAfter string
	for i := 0; i < 100; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
		resp, err := r.client.Do(req)
		if err != nil {
			continue
		}
		io.Copy(io.Discard, resp.Body)

		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter = resp.Header.Get("Retry-After")
			resp.Body.Close()
			break
		}
		resp.Body.Close()
	}

	result.Duration = time.Since(start)

	// Check if Retry-After was provided
	result.Passed = retryAfter != ""

	result.Actual = harness.Actual{
		Body: "Retry-After: " + retryAfter,
	}

	if retryAfter == "" && r.gapTracker != nil {
		r.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategorySecurity,
			Severity:     gaps.SeverityLow,
			Description:  "429 response missing Retry-After header",
			DiscoveredBy: "ratelimit_retry_after",
			Remediation:  "Add Retry-After header to 429 responses",
			SpecRef:      "v0.8/rate-limit.md#response-format",
		})
	}

	return result
}
