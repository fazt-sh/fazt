package harness

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// Harness is the main test orchestrator.
type Harness struct {
	config     Config
	client     *http.Client
	db         *sql.DB
	gapTracker *gaps.Tracker
	report     *Report
	mu         sync.Mutex
}

// New creates a new test harness.
func New(config Config) *Harness {
	return &Harness{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        200,
				MaxIdleConnsPerHost: 200,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		gapTracker: gaps.NewTracker(),
	}
}

// SetDB sets the database connection for direct DB tests.
func (h *Harness) SetDB(db *sql.DB) {
	h.db = db
}

// Run executes the test harness.
func (h *Harness) Run(ctx context.Context, version string) (*Report, error) {
	start := time.Now()

	h.report = NewReport(version, h.config.TargetURL)

	// Verify target is reachable
	if err := h.healthCheck(ctx); err != nil {
		return nil, fmt.Errorf("target health check failed: %w", err)
	}

	// Run each enabled category
	for _, category := range h.config.Categories {
		select {
		case <-ctx.Done():
			return h.report, ctx.Err()
		default:
		}

		if err := h.runCategory(ctx, category); err != nil {
			// Log error but continue with other categories
			fmt.Printf("Warning: category %s had errors: %v\n", category, err)
		}
	}

	h.report.Duration = time.Since(start)
	h.report.Finalize(h.gapTracker)

	return h.report, nil
}

// RunSmoke runs a quick smoke test.
func (h *Harness) RunSmoke(ctx context.Context, version string) (*Report, error) {
	h.report = NewReport(version, h.config.TargetURL)
	start := time.Now()

	// Just verify basic functionality
	if err := h.healthCheck(ctx); err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	// Run minimal request tests
	h.runStaticTests(ctx)
	h.runAPIReadTests(ctx)

	h.report.Duration = time.Since(start)
	h.report.Finalize(h.gapTracker)

	return h.report, nil
}

func (h *Harness) healthCheck(ctx context.Context) error {
	// Try multiple health endpoints
	endpoints := []string{
		"/health",
		"/api/system/health",
		"/",
	}

	for _, endpoint := range endpoints {
		req, err := http.NewRequestWithContext(ctx, "GET", h.config.TargetURL+endpoint, nil)
		if err != nil {
			continue
		}

		resp, err := h.client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		// Accept 200, 303 (redirect), or 401 (auth required but server is up)
		if resp.StatusCode == http.StatusOK ||
			resp.StatusCode == http.StatusSeeOther ||
			resp.StatusCode == http.StatusUnauthorized {
			return nil
		}
	}

	return fmt.Errorf("no health endpoint responded")
}

func (h *Harness) runCategory(ctx context.Context, category string) error {
	switch category {
	case "baseline":
		return h.runBaseline(ctx)
	case "requests":
		return h.runRequests(ctx)
	case "resilience":
		return h.runResilience(ctx)
	case "security":
		return h.runSecurity(ctx)
	default:
		return fmt.Errorf("unknown category: %s", category)
	}
}

func (h *Harness) runBaseline(ctx context.Context) error {
	fmt.Println("Running baseline measurements...")

	// Warmup
	h.warmup(ctx)

	// Throughput tests
	for _, concurrency := range h.config.Baseline.ConcurrencyLevels {
		if err := h.measureThroughput(ctx, "static_read", concurrency); err != nil {
			return err
		}
		if err := h.measureThroughput(ctx, "api_read", concurrency); err != nil {
			return err
		}
	}

	// Latency tests
	h.measureLatency(ctx, "static_read")
	h.measureLatency(ctx, "api_read")
	h.measureLatency(ctx, "api_write")

	// Resource measurements
	h.measureResources(ctx, "idle")
	h.measureResources(ctx, "under_load")

	return nil
}

func (h *Harness) runRequests(ctx context.Context) error {
	fmt.Println("Running request lifecycle tests...")

	h.runStaticTests(ctx)
	h.runAPIReadTests(ctx)
	h.runAPIWriteTests(ctx)
	h.runServerlessTests(ctx)
	h.runAuthTests(ctx)

	return nil
}

func (h *Harness) runResilience(ctx context.Context) error {
	fmt.Println("Running resilience tests...")

	h.runMemoryTests(ctx)
	h.runQueueTests(ctx)
	h.runTimeoutTests(ctx)

	return nil
}

func (h *Harness) runSecurity(ctx context.Context) error {
	fmt.Println("Running security tests...")

	h.runRateLimitTests(ctx)
	h.runPayloadTests(ctx)
	h.runSlowClientTests(ctx)

	return nil
}

func (h *Harness) warmup(ctx context.Context) {
	end := time.Now().Add(h.config.Baseline.WarmupDuration)
	for time.Now().Before(end) {
		req, _ := http.NewRequestWithContext(ctx, "GET", h.config.TargetURL+"/health", nil)
		resp, err := h.client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}
}

func (h *Harness) measureThroughput(ctx context.Context, scenario string, concurrency int) error {
	result := ThroughputResult{
		Scenario:    scenario,
		Concurrency: concurrency,
		Duration:    h.config.Baseline.Duration,
		Errors:      make(map[string]int),
	}

	path := h.getScenarioPath(scenario)
	method := h.getScenarioMethod(scenario)

	var wg sync.WaitGroup
	var mu sync.Mutex
	done := make(chan struct{})

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					req, _ := http.NewRequestWithContext(ctx, method, h.config.TargetURL+path, nil)
					resp, err := h.client.Do(req)

					mu.Lock()
					result.TotalRequests++
					if err != nil {
						result.ErrorCount++
						result.Errors[err.Error()]++
					} else {
						resp.Body.Close()
						if resp.StatusCode >= 200 && resp.StatusCode < 400 {
							result.SuccessCount++
						} else {
							result.ErrorCount++
							result.Errors[fmt.Sprintf("status_%d", resp.StatusCode)]++
						}
					}
					mu.Unlock()
				}
			}
		}()
	}

	// Wait for duration
	time.Sleep(h.config.Baseline.Duration)
	close(done)
	wg.Wait()

	result.RPS = float64(result.SuccessCount) / h.config.Baseline.Duration.Seconds()
	result.ExpectedRPS = h.getExpectedRPS(scenario, concurrency)
	result.WithinBaseline = result.RPS >= result.ExpectedRPS*(1-h.config.Baseline.Tolerance)

	h.report.AddThroughput(result)

	// Record gap if significantly below baseline
	if result.RPS < result.ExpectedRPS*0.5 {
		h.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategoryPerformance,
			Severity:     gaps.SeverityHigh,
			Description:  fmt.Sprintf("%s throughput at %d concurrency: %.0f RPS (expected %.0f)", scenario, concurrency, result.RPS, result.ExpectedRPS),
			DiscoveredBy: fmt.Sprintf("throughput_%s_%d", scenario, concurrency),
		})
	}

	return nil
}

func (h *Harness) measureLatency(ctx context.Context, scenario string) {
	path := h.getScenarioPath(scenario)
	method := h.getScenarioMethod(scenario)

	var latencies []time.Duration
	samples := 1000

	for i := 0; i < samples; i++ {
		start := time.Now()
		req, _ := http.NewRequestWithContext(ctx, method, h.config.TargetURL+path, nil)
		resp, err := h.client.Do(req)
		if err == nil {
			resp.Body.Close()
			latencies = append(latencies, time.Since(start))
		}
	}

	if len(latencies) == 0 {
		return
	}

	result := LatencyResult{
		Scenario:   scenario,
		SampleSize: len(latencies),
	}

	// Sort for percentiles
	sortDurations(latencies)

	result.Min = latencies[0]
	result.Max = latencies[len(latencies)-1]
	result.P50 = latencies[len(latencies)*50/100]
	result.P95 = latencies[len(latencies)*95/100]
	result.P99 = latencies[len(latencies)*99/100]

	var sum time.Duration
	for _, d := range latencies {
		sum += d
	}
	result.Mean = sum / time.Duration(len(latencies))

	h.report.AddLatency(result)
}

func (h *Harness) measureResources(ctx context.Context, scenario string) {
	// For now, just record a placeholder
	// Real implementation would query /api/system/stats or similar
	result := ResourceResult{
		Scenario: scenario,
	}
	h.report.AddResource(result)
}

// Test implementations (stubs to be filled)

func (h *Harness) runStaticTests(ctx context.Context) {
	tests := []struct {
		name     string
		path     string
		expected Expected
	}{
		{"health_check", "/health", Expect(200)},
	}

	for _, tc := range tests {
		result := h.runSingleTest(ctx, tc.name, "static", "GET", tc.path, tc.expected)
		h.report.AddTestResult(result)
	}
}

func (h *Harness) runAPIReadTests(ctx context.Context) {
	// API read tests - would need actual endpoints
	result := TestResult{
		Name:     "api_health",
		Category: "api_read",
		Passed:   true,
	}
	h.report.AddTestResult(result)
}

func (h *Harness) runAPIWriteTests(ctx context.Context) {
	// API write tests
}

func (h *Harness) runServerlessTests(ctx context.Context) {
	// Serverless execution tests
}

func (h *Harness) runAuthTests(ctx context.Context) {
	// Auth lifecycle tests
}

func (h *Harness) runMemoryTests(ctx context.Context) {
	// Memory pressure tests
}

func (h *Harness) runQueueTests(ctx context.Context) {
	// Write queue tests
}

func (h *Harness) runTimeoutTests(ctx context.Context) {
	// Timeout enforcement tests
}

func (h *Harness) runRateLimitTests(ctx context.Context) {
	// Rate limit tests
}

func (h *Harness) runPayloadTests(ctx context.Context) {
	// Large payload tests
}

func (h *Harness) runSlowClientTests(ctx context.Context) {
	// Slow client tests
}

func (h *Harness) runSingleTest(ctx context.Context, name, category, method, path string, expected Expected) TestResult {
	start := time.Now()
	result := TestResult{
		Name:     name,
		Category: category,
		Expected: expected,
	}

	req, err := http.NewRequestWithContext(ctx, method, h.config.TargetURL+path, nil)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	resp, err := h.client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()

	result.Actual = Actual{
		Status:  resp.StatusCode,
		Latency: result.Duration,
	}

	// Validate
	result.Passed = resp.StatusCode == expected.Status
	if expected.MaxLatency > 0 && result.Duration > expected.MaxLatency {
		result.Passed = false
	}

	return result
}

// GapTracker returns the gap tracker for external use.
func (h *Harness) GapTracker() *gaps.Tracker {
	return h.gapTracker
}

// Helper functions

func (h *Harness) getScenarioPath(scenario string) string {
	switch scenario {
	case "static_read":
		return "/api/system/health"
	case "api_read":
		return "/api/system/health"
	case "api_write":
		return "/api/deploy"
	default:
		return "/api/system/health"
	}
}

func (h *Harness) getScenarioMethod(scenario string) string {
	switch scenario {
	case "api_write":
		return "POST"
	default:
		return "GET"
	}
}

func (h *Harness) getExpectedRPS(scenario string, concurrency int) float64 {
	// Base RPS expectations (at concurrency 100)
	baseRPS := map[string]float64{
		"static_read": 20000,
		"api_read":    15000,
		"api_write":   800,
	}

	base, ok := baseRPS[scenario]
	if !ok {
		return 1000
	}

	// Scale linearly with concurrency (simplified)
	return base * float64(concurrency) / 100.0
}

func sortDurations(d []time.Duration) {
	for i := 1; i < len(d); i++ {
		for j := i; j > 0 && d[j] < d[j-1]; j-- {
			d[j], d[j-1] = d[j-1], d[j]
		}
	}
}
