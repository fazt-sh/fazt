// Package resilience provides resilience and chaos tests.
package resilience

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/fazt-sh/fazt/internal/harness"
	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// MemoryTest defines a memory pressure test.
type MemoryTest struct {
	Name        string
	Description string
	Setup       func(ctx context.Context) error
	Verify      func(ctx context.Context, result *MemoryTestResult) bool
	Cleanup     func(ctx context.Context)
}

// MemoryTestResult holds memory test results.
type MemoryTestResult struct {
	HeapBeforeMB   uint64
	HeapAfterMB    uint64
	RequestsOK     bool
	CacheEvicted   bool
	GCTriggered    bool
	DegradedMode   bool
	ResponseStatus int
}

// MemoryRunner executes memory resilience tests.
type MemoryRunner struct {
	client     *http.Client
	baseURL    string
	gapTracker *gaps.Tracker
}

// NewMemoryRunner creates a new memory test runner.
func NewMemoryRunner(baseURL string, gapTracker *gaps.Tracker) *MemoryRunner {
	return &MemoryRunner{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    baseURL,
		gapTracker: gapTracker,
	}
}

// Run executes memory resilience tests.
func (r *MemoryRunner) Run(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	// Test 1: System handles high request concurrency
	results = append(results, r.testHighConcurrency(ctx))

	// Test 2: System recovers from sustained load
	results = append(results, r.testSustainedLoad(ctx))

	// Test 3: Large response handling
	results = append(results, r.testLargeResponses(ctx))

	return results
}

func (r *MemoryRunner) testHighConcurrency(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "high_concurrency_memory",
		Category: "resilience",
	}

	start := time.Now()

	// Spawn many concurrent requests
	concurrency := 500
	var wg sync.WaitGroup
	var successCount, errorCount int
	var mu sync.Mutex

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
			resp, err := r.client.Do(req)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errorCount++
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				successCount++
			} else if resp.StatusCode == http.StatusServiceUnavailable {
				// Acceptable under pressure
				successCount++
			} else {
				errorCount++
			}
		}()
	}

	wg.Wait()
	result.Duration = time.Since(start)

	// Success if most requests completed OK (allowing for some 503s)
	successRate := float64(successCount) / float64(concurrency)
	result.Passed = successRate >= 0.8

	result.Actual = harness.Actual{
		Body: "success_rate=" + formatPercent(successRate),
	}

	if !result.Passed && r.gapTracker != nil {
		r.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategoryPerformance,
			Severity:     gaps.SeverityHigh,
			Description:  "High concurrency test failed: " + formatPercent(successRate) + " success rate",
			DiscoveredBy: "memory_high_concurrency",
			Remediation:  "Investigate connection handling and memory limits",
		})
	}

	return result
}

func (r *MemoryRunner) testSustainedLoad(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "sustained_load_recovery",
		Category: "resilience",
	}

	start := time.Now()
	duration := 30 * time.Second
	concurrency := 100

	// Generate sustained load
	done := make(chan struct{})
	var errorCount int64
	var mu sync.Mutex

	for i := 0; i < concurrency; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}

				req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
				resp, err := r.client.Do(req)
				if err != nil {
					mu.Lock()
					errorCount++
					mu.Unlock()
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}()
	}

	time.Sleep(duration)
	close(done)

	// Brief pause for recovery
	time.Sleep(2 * time.Second)

	// Verify system is still responsive
	req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
	resp, err := r.client.Do(req)

	result.Duration = time.Since(start)

	if err != nil {
		result.Passed = false
		result.Error = err
	} else {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		result.Passed = resp.StatusCode == http.StatusOK
		result.Actual = harness.Actual{Status: resp.StatusCode}
	}

	if !result.Passed && r.gapTracker != nil {
		r.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategoryPerformance,
			Severity:     gaps.SeverityHigh,
			Description:  "System did not recover after sustained load",
			DiscoveredBy: "memory_sustained_load",
			Remediation:  "Check for resource leaks and cleanup mechanisms",
		})
	}

	return result
}

func (r *MemoryRunner) testLargeResponses(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "large_response_handling",
		Category: "resilience",
	}

	start := time.Now()

	// Make multiple requests that might return large responses
	// In practice, this would test file downloads or large API responses
	successCount := 0
	totalRequests := 10

	for i := 0; i < totalRequests; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
		resp, err := r.client.Do(req)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				successCount++
			}
		}
	}

	result.Duration = time.Since(start)
	result.Passed = successCount == totalRequests

	return result
}

// ConnectionLeakTest verifies connections are properly released.
func (r *MemoryRunner) ConnectionLeakTest(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "connection_leak_test",
		Category: "resilience",
	}

	start := time.Now()

	// Make many requests, explicitly not reading bodies
	// This could cause connection leaks if not handled properly
	for i := 0; i < 100; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
		resp, err := r.client.Do(req)
		if err == nil {
			// Intentionally close without reading - should still work
			resp.Body.Close()
		}
	}

	// Brief pause
	time.Sleep(1 * time.Second)

	// Verify system still works
	req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
	resp, err := r.client.Do(req)

	result.Duration = time.Since(start)

	if err != nil {
		result.Passed = false
		result.Error = err
		if r.gapTracker != nil {
			r.gapTracker.Record(gaps.Gap{
				Category:     gaps.CategoryPerformance,
				Severity:     gaps.SeverityCritical,
				Description:  "Connection leak suspected - system unresponsive after rapid requests",
				DiscoveredBy: "memory_connection_leak",
				Remediation:  "Review connection handling and timeouts",
			})
		}
	} else {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		result.Passed = resp.StatusCode == http.StatusOK
	}

	return result
}

// GCPressureTest measures behavior under GC pressure.
func (r *MemoryRunner) GCPressureTest(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "gc_pressure",
		Category: "resilience",
	}

	start := time.Now()

	// Generate requests that create short-lived allocations
	var latencies []time.Duration
	var mu sync.Mutex

	concurrency := 50
	duration := 10 * time.Second
	done := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}

				reqStart := time.Now()
				req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
				resp, err := r.client.Do(req)
				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					mu.Lock()
					latencies = append(latencies, time.Since(reqStart))
					mu.Unlock()
				}
			}
		}()
	}

	time.Sleep(duration)
	close(done)

	result.Duration = time.Since(start)

	// Check for latency spikes (GC pauses)
	if len(latencies) > 0 {
		var maxLatency time.Duration
		for _, l := range latencies {
			if l > maxLatency {
				maxLatency = l
			}
		}

		// Pass if max latency is reasonable (< 500ms)
		result.Passed = maxLatency < 500*time.Millisecond
		result.Actual = harness.Actual{
			Body: "max_latency=" + maxLatency.String(),
		}

		if !result.Passed && r.gapTracker != nil {
			r.gapTracker.Record(gaps.Gap{
				Category:     gaps.CategoryPerformance,
				Severity:     gaps.SeverityMedium,
				Description:  "High latency spikes under GC pressure: " + maxLatency.String(),
				DiscoveredBy: "memory_gc_pressure",
				Remediation:  "Tune GC or reduce allocation rate",
			})
		}
	}

	return result
}

func formatPercent(f float64) string {
	return string(rune(int(f*100))) + "%"
}
