package baseline

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/fazt-sh/fazt/internal/harness"
	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// LatencyBaseline defines expected latency for a scenario.
type LatencyBaseline struct {
	Scenario    string
	Description string
	Method      string
	Path        string
	P50         time.Duration
	P95         time.Duration
	P99         time.Duration
}

// DefaultLatencyBaselines returns expected latency baselines.
func DefaultLatencyBaselines() []LatencyBaseline {
	return []LatencyBaseline{
		{
			Scenario:    "static_read",
			Description: "Static file read",
			Method:      "GET",
			Path:        "/health",
			P50:         1 * time.Millisecond,
			P95:         5 * time.Millisecond,
			P99:         20 * time.Millisecond,
		},
		{
			Scenario:    "api_read",
			Description: "API read endpoint",
			Method:      "GET",
			Path:        "/health",
			P50:         2 * time.Millisecond,
			P95:         10 * time.Millisecond,
			P99:         50 * time.Millisecond,
		},
	}
}

// LatencyRunner measures latency.
type LatencyRunner struct {
	client     *http.Client
	baseURL    string
	gapTracker *gaps.Tracker
}

// NewLatencyRunner creates a new latency runner.
func NewLatencyRunner(baseURL string, gapTracker *gaps.Tracker) *LatencyRunner {
	return &LatencyRunner{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		baseURL:    baseURL,
		gapTracker: gapTracker,
	}
}

// Run executes latency measurements.
func (r *LatencyRunner) Run(ctx context.Context, samples int) []harness.LatencyResult {
	baselines := DefaultLatencyBaselines()
	var results []harness.LatencyResult

	for _, baseline := range baselines {
		result := r.measure(ctx, baseline, samples)
		results = append(results, result)

		// Record gap if P99 significantly exceeds baseline
		if baseline.P99 > 0 && result.P99 > baseline.P99*2 && r.gapTracker != nil {
			r.gapTracker.Record(gaps.Gap{
				Category: gaps.CategoryPerformance,
				Severity: gaps.SeverityMedium,
				Description: fmt.Sprintf("%s P99 latency: %s (expected <%s)",
					baseline.Scenario, result.P99, baseline.P99),
				DiscoveredBy: "latency_" + baseline.Scenario,
				Remediation:  "Investigate tail latency sources",
			})
		}
	}

	return results
}

// MeasureCustom measures latency for a custom scenario.
func (r *LatencyRunner) MeasureCustom(ctx context.Context, name, method, path string, samples int) harness.LatencyResult {
	baseline := LatencyBaseline{
		Scenario: name,
		Method:   method,
		Path:     path,
	}
	return r.measure(ctx, baseline, samples)
}

func (r *LatencyRunner) measure(ctx context.Context, baseline LatencyBaseline, samples int) harness.LatencyResult {
	result := harness.LatencyResult{
		Scenario: baseline.Scenario,
	}

	latencies := make([]time.Duration, 0, samples)

	for i := 0; i < samples; i++ {
		select {
		case <-ctx.Done():
			break
		default:
		}

		start := time.Now()
		req, _ := http.NewRequestWithContext(ctx, baseline.Method, r.baseURL+baseline.Path, nil)
		resp, err := r.client.Do(req)
		latency := time.Since(start)

		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				latencies = append(latencies, latency)
			}
		}
	}

	if len(latencies) == 0 {
		return result
	}

	result.SampleSize = len(latencies)

	// Sort for percentile calculation
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	result.Min = latencies[0]
	result.Max = latencies[len(latencies)-1]
	result.P50 = percentile(latencies, 50)
	result.P95 = percentile(latencies, 95)
	result.P99 = percentile(latencies, 99)

	var sum time.Duration
	for _, d := range latencies {
		sum += d
	}
	result.Mean = sum / time.Duration(len(latencies))

	return result
}

// MeasureUnderLoad measures latency while under concurrent load.
func (r *LatencyRunner) MeasureUnderLoad(ctx context.Context, scenario string, path string, concurrency int, duration time.Duration) harness.LatencyResult {
	result := harness.LatencyResult{
		Scenario: scenario + "_under_load",
	}

	var (
		latencies []time.Duration
		mu        sync.Mutex
		wg        sync.WaitGroup
	)

	done := make(chan struct{})

	// Start background load
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

				start := time.Now()
				req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+path, nil)
				resp, err := r.client.Do(req)
				latency := time.Since(start)

				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					if resp.StatusCode >= 200 && resp.StatusCode < 400 {
						mu.Lock()
						latencies = append(latencies, latency)
						mu.Unlock()
					}
				}
			}
		}()
	}

	// Wait for duration
	time.Sleep(duration)
	close(done)
	wg.Wait()

	if len(latencies) == 0 {
		return result
	}

	result.SampleSize = len(latencies)

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	result.Min = latencies[0]
	result.Max = latencies[len(latencies)-1]
	result.P50 = percentile(latencies, 50)
	result.P95 = percentile(latencies, 95)
	result.P99 = percentile(latencies, 99)

	var sum time.Duration
	for _, d := range latencies {
		sum += d
	}
	result.Mean = sum / time.Duration(len(latencies))

	return result
}

// LatencyDistribution represents a latency histogram.
type LatencyDistribution struct {
	Buckets    []LatencyBucket
	TotalCount int
}

// LatencyBucket represents a histogram bucket.
type LatencyBucket struct {
	LowerBound time.Duration
	UpperBound time.Duration
	Count      int
	Percentage float64
}

// MeasureDistribution measures latency distribution.
func (r *LatencyRunner) MeasureDistribution(ctx context.Context, path string, samples int) LatencyDistribution {
	bucketBounds := []time.Duration{
		1 * time.Millisecond,
		5 * time.Millisecond,
		10 * time.Millisecond,
		25 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		250 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}

	buckets := make([]int, len(bucketBounds)+1)
	var latencies []time.Duration

	for i := 0; i < samples; i++ {
		start := time.Now()
		req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+path, nil)
		resp, err := r.client.Do(req)
		latency := time.Since(start)

		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			latencies = append(latencies, latency)

			// Find bucket
			bucketIdx := len(bucketBounds)
			for i, bound := range bucketBounds {
				if latency < bound {
					bucketIdx = i
					break
				}
			}
			buckets[bucketIdx]++
		}
	}

	dist := LatencyDistribution{
		TotalCount: len(latencies),
		Buckets:    make([]LatencyBucket, 0, len(bucketBounds)+1),
	}

	// Build bucket results
	prevBound := time.Duration(0)
	for i, bound := range bucketBounds {
		pct := 0.0
		if dist.TotalCount > 0 {
			pct = float64(buckets[i]) / float64(dist.TotalCount) * 100
		}
		dist.Buckets = append(dist.Buckets, LatencyBucket{
			LowerBound: prevBound,
			UpperBound: bound,
			Count:      buckets[i],
			Percentage: pct,
		})
		prevBound = bound
	}

	// Final bucket (>= last bound)
	lastPct := 0.0
	if dist.TotalCount > 0 {
		lastPct = float64(buckets[len(bucketBounds)]) / float64(dist.TotalCount) * 100
	}
	dist.Buckets = append(dist.Buckets, LatencyBucket{
		LowerBound: prevBound,
		UpperBound: time.Duration(0), // Unbounded
		Count:      buckets[len(bucketBounds)],
		Percentage: lastPct,
	})

	return dist
}

func percentile(sorted []time.Duration, pct int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := len(sorted) * pct / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
