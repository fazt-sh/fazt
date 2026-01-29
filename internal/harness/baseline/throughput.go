// Package baseline provides performance baseline measurements.
package baseline

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

// ThroughputBaseline defines expected throughput for a scenario.
type ThroughputBaseline struct {
	Scenario    string
	Description string
	Method      string
	Path        string
	Body        []byte
	ExpectedRPS map[int]float64 // concurrency -> expected RPS
	Tolerance   float64         // ±10% by default
}

// DefaultThroughputBaselines returns expected throughput baselines.
func DefaultThroughputBaselines() []ThroughputBaseline {
	return []ThroughputBaseline{
		{
			Scenario:    "static_read",
			Description: "Static file read (health endpoint)",
			Method:      "GET",
			Path:        "/health",
			ExpectedRPS: map[int]float64{
				1:   500,
				10:  3000,
				50:  10000,
				100: 15000,
			},
			Tolerance: 0.15,
		},
		{
			Scenario:    "api_read",
			Description: "API read endpoint",
			Method:      "GET",
			Path:        "/health",
			ExpectedRPS: map[int]float64{
				1:   400,
				10:  2500,
				50:  8000,
				100: 12000,
			},
			Tolerance: 0.15,
		},
	}
}

// ThroughputRunner measures throughput.
type ThroughputRunner struct {
	client     *http.Client
	baseURL    string
	gapTracker *gaps.Tracker
}

// NewThroughputRunner creates a new throughput runner.
func NewThroughputRunner(baseURL string, gapTracker *gaps.Tracker) *ThroughputRunner {
	return &ThroughputRunner{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  true,
			},
		},
		baseURL:    baseURL,
		gapTracker: gapTracker,
	}
}

// Run executes throughput measurements.
func (r *ThroughputRunner) Run(ctx context.Context, duration time.Duration, warmupDuration time.Duration) []harness.ThroughputResult {
	baselines := DefaultThroughputBaselines()
	var results []harness.ThroughputResult

	// Warmup
	r.warmup(ctx, warmupDuration)

	for _, baseline := range baselines {
		for concurrency, expectedRPS := range baseline.ExpectedRPS {
			result := r.measure(ctx, baseline, concurrency, duration, expectedRPS)
			results = append(results, result)

			// Record gap if significantly below baseline
			if result.RPS < expectedRPS*(1-baseline.Tolerance*2) && r.gapTracker != nil {
				r.gapTracker.Record(gaps.Gap{
					Category: gaps.CategoryPerformance,
					Severity: gaps.SeverityHigh,
					Description: fmt.Sprintf("%s@%d: %.0f RPS (expected %.0f, %.1f%% below)",
						baseline.Scenario, concurrency, result.RPS, expectedRPS,
						(1-result.RPS/expectedRPS)*100),
					DiscoveredBy: fmt.Sprintf("throughput_%s_%d", baseline.Scenario, concurrency),
					Remediation:  "Profile and optimize hot paths",
				})
			}
		}
	}

	return results
}

// MeasureCustom measures throughput for a custom scenario.
func (r *ThroughputRunner) MeasureCustom(ctx context.Context, name, method, path string, body []byte, concurrency int, duration time.Duration) harness.ThroughputResult {
	baseline := ThroughputBaseline{
		Scenario: name,
		Method:   method,
		Path:     path,
		Body:     body,
	}
	return r.measure(ctx, baseline, concurrency, duration, 0)
}

func (r *ThroughputRunner) warmup(ctx context.Context, duration time.Duration) {
	end := time.Now().Add(duration)
	for time.Now().Before(end) {
		select {
		case <-ctx.Done():
			return
		default:
		}
		req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
		resp, err := r.client.Do(req)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
}

func (r *ThroughputRunner) measure(ctx context.Context, baseline ThroughputBaseline, concurrency int, duration time.Duration, expectedRPS float64) harness.ThroughputResult {
	result := harness.ThroughputResult{
		Scenario:    baseline.Scenario,
		Duration:    duration,
		Concurrency: concurrency,
		ExpectedRPS: expectedRPS,
		Errors:      make(map[string]int),
	}

	var (
		totalRequests int64
		successCount  int64
		errorCount    int64
		wg            sync.WaitGroup
		errorMu       sync.Mutex
	)

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
				}

				req, _ := http.NewRequestWithContext(ctx, baseline.Method, r.baseURL+baseline.Path, nil)
				resp, err := r.client.Do(req)

				atomic.AddInt64(&totalRequests, 1)

				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					errorMu.Lock()
					result.Errors[truncateError(err.Error())]++
					errorMu.Unlock()
					continue
				}

				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 400 {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&errorCount, 1)
					errorMu.Lock()
					result.Errors[fmt.Sprintf("status_%d", resp.StatusCode)]++
					errorMu.Unlock()
				}
			}
		}()
	}

	// Wait for duration
	time.Sleep(duration)
	close(done)
	wg.Wait()

	result.TotalRequests = totalRequests
	result.SuccessCount = successCount
	result.ErrorCount = errorCount
	result.RPS = float64(successCount) / duration.Seconds()
	result.WithinBaseline = expectedRPS == 0 || result.RPS >= expectedRPS*(1-baseline.Tolerance)

	return result
}

// MixedWorkloadResult holds results for mixed read/write workload.
type MixedWorkloadResult struct {
	ReadRPS      float64
	WriteRPS     float64
	CombinedRPS  float64
	ReadLatency  time.Duration
	WriteLatency time.Duration
	Errors       int64
}

// MeasureMixedWorkload measures throughput with a read/write mix.
func (r *ThroughputRunner) MeasureMixedWorkload(ctx context.Context, readPercent int, concurrency int, duration time.Duration) MixedWorkloadResult {
	result := MixedWorkloadResult{}

	var (
		readCount    int64
		writeCount   int64
		readLatSum   int64
		writeLatSum  int64
		errorCount   int64
		wg           sync.WaitGroup
	)

	done := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			counter := 0
			for {
				select {
				case <-done:
					return
				default:
				}

				// Decide read or write based on percentage
				isRead := (counter % 100) < readPercent
				counter++

				start := time.Now()
				var err error
				var resp *http.Response

				if isRead {
					req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
					resp, err = r.client.Do(req)
				} else {
					// Simulated write (to health endpoint for now)
					req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
					resp, err = r.client.Do(req)
				}

				latency := time.Since(start)

				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					continue
				}

				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 400 {
					if isRead {
						atomic.AddInt64(&readCount, 1)
						atomic.AddInt64(&readLatSum, int64(latency))
					} else {
						atomic.AddInt64(&writeCount, 1)
						atomic.AddInt64(&writeLatSum, int64(latency))
					}
				} else {
					atomic.AddInt64(&errorCount, 1)
				}
			}
		}(i)
	}

	time.Sleep(duration)
	close(done)
	wg.Wait()

	result.ReadRPS = float64(readCount) / duration.Seconds()
	result.WriteRPS = float64(writeCount) / duration.Seconds()
	result.CombinedRPS = float64(readCount+writeCount) / duration.Seconds()
	result.Errors = errorCount

	if readCount > 0 {
		result.ReadLatency = time.Duration(readLatSum / readCount)
	}
	if writeCount > 0 {
		result.WriteLatency = time.Duration(writeLatSum / writeCount)
	}

	return result
}

// ScalabilityResult holds results for scalability testing.
type ScalabilityResult struct {
	Concurrency int
	RPS         float64
	Efficiency  float64 // RPS per concurrent connection
}

// MeasureScalability measures how throughput scales with concurrency.
func (r *ThroughputRunner) MeasureScalability(ctx context.Context, concurrencyLevels []int, duration time.Duration) []ScalabilityResult {
	var results []ScalabilityResult

	for _, concurrency := range concurrencyLevels {
		baseline := ThroughputBaseline{
			Scenario: "scalability",
			Method:   "GET",
			Path:     "/health",
		}

		throughput := r.measure(ctx, baseline, concurrency, duration, 0)

		results = append(results, ScalabilityResult{
			Concurrency: concurrency,
			RPS:         throughput.RPS,
			Efficiency:  throughput.RPS / float64(concurrency),
		})

		// Brief pause between measurements
		time.Sleep(500 * time.Millisecond)
	}

	return results
}

func truncateError(s string) string {
	if len(s) > 50 {
		return s[:50] + "..."
	}
	return s
}
