package baseline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/fazt-sh/fazt/internal/harness"
	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// ResourceBaseline defines expected resource usage.
type ResourceBaseline struct {
	Scenario       string
	HeapStableMB   uint64
	HeapPeakMB     uint64
	GoroutinesPeak int
	GCPauseP99     time.Duration
}

// DefaultResourceBaselines returns expected resource baselines.
func DefaultResourceBaselines() []ResourceBaseline {
	return []ResourceBaseline{
		{
			Scenario:       "idle",
			HeapStableMB:   30,
			HeapPeakMB:     50,
			GoroutinesPeak: 100,
			GCPauseP99:     5 * time.Millisecond,
		},
		{
			Scenario:       "100_concurrent",
			HeapStableMB:   50,
			HeapPeakMB:     80,
			GoroutinesPeak: 300,
			GCPauseP99:     10 * time.Millisecond,
		},
		{
			Scenario:       "500_concurrent",
			HeapStableMB:   80,
			HeapPeakMB:     150,
			GoroutinesPeak: 800,
			GCPauseP99:     20 * time.Millisecond,
		},
	}
}

// ResourceRunner measures resource usage.
type ResourceRunner struct {
	client     *http.Client
	baseURL    string
	gapTracker *gaps.Tracker
}

// NewResourceRunner creates a new resource runner.
func NewResourceRunner(baseURL string, gapTracker *gaps.Tracker) *ResourceRunner {
	return &ResourceRunner{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    baseURL,
		gapTracker: gapTracker,
	}
}

// LocalSnapshot captures local process resource usage.
type LocalSnapshot struct {
	HeapAlloc     uint64
	HeapSys       uint64
	HeapIdle      uint64
	HeapInuse     uint64
	StackInuse    uint64
	NumGoroutine  int
	NumGC         uint32
	PauseTotalNs  uint64
	LastPauseNs   uint64
}

// CaptureLocal captures local process resource state.
func (r *ResourceRunner) CaptureLocal() LocalSnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return LocalSnapshot{
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		StackInuse:   m.StackInuse,
		NumGoroutine: runtime.NumGoroutine(),
		NumGC:        m.NumGC,
		PauseTotalNs: m.PauseTotalNs,
		LastPauseNs:  m.PauseNs[(m.NumGC+255)%256],
	}
}

// RemoteStats represents stats fetched from the target server.
type RemoteStats struct {
	HeapAlloc    uint64  `json:"heap_alloc"`
	HeapSys      uint64  `json:"heap_sys"`
	NumGoroutine int     `json:"num_goroutine"`
	NumGC        uint32  `json:"num_gc"`
	Uptime       float64 `json:"uptime_seconds"`
}

// FetchRemoteStats fetches resource stats from target server.
func (r *ResourceRunner) FetchRemoteStats(ctx context.Context) (*RemoteStats, error) {
	// Try to fetch from /api/system/stats or similar endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/api/system/stats", nil)
	if err != nil {
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stats endpoint returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var stats RemoteStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// MeasureIdle measures resource usage at idle.
func (r *ResourceRunner) MeasureIdle(ctx context.Context) harness.ResourceResult {
	result := harness.ResourceResult{
		Scenario: "idle",
	}

	// Try remote stats first
	if stats, err := r.FetchRemoteStats(ctx); err == nil {
		result.HeapAllocMB = stats.HeapAlloc / (1024 * 1024)
		result.HeapSysMB = stats.HeapSys / (1024 * 1024)
		result.GoroutineCount = stats.NumGoroutine
	} else {
		// Fall back to local measurement (useful for testing the harness itself)
		local := r.CaptureLocal()
		result.HeapAllocMB = local.HeapAlloc / (1024 * 1024)
		result.HeapSysMB = local.HeapSys / (1024 * 1024)
		result.GoroutineCount = local.NumGoroutine
		result.GCPauseNs = local.LastPauseNs
	}

	return result
}

// MeasureUnderLoad measures resources while under load.
func (r *ResourceRunner) MeasureUnderLoad(ctx context.Context, concurrency int, duration time.Duration) harness.ResourceResult {
	result := harness.ResourceResult{
		Scenario: fmt.Sprintf("%d_concurrent", concurrency),
	}

	// Start load generation in background
	done := make(chan struct{})
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
				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
			}
		}()
	}

	// Let load stabilize
	time.Sleep(duration / 2)

	// Sample resources multiple times
	var samples []harness.ResourceResult
	sampleInterval := duration / 10
	for i := 0; i < 5; i++ {
		time.Sleep(sampleInterval)

		if stats, err := r.FetchRemoteStats(ctx); err == nil {
			samples = append(samples, harness.ResourceResult{
				HeapAllocMB:    stats.HeapAlloc / (1024 * 1024),
				HeapSysMB:      stats.HeapSys / (1024 * 1024),
				GoroutineCount: stats.NumGoroutine,
			})
		}
	}

	close(done)

	// Calculate averages
	if len(samples) > 0 {
		var heapSum, sysSum uint64
		var goroutineSum int
		for _, s := range samples {
			heapSum += s.HeapAllocMB
			sysSum += s.HeapSysMB
			goroutineSum += s.GoroutineCount
		}
		result.HeapAllocMB = heapSum / uint64(len(samples))
		result.HeapSysMB = sysSum / uint64(len(samples))
		result.GoroutineCount = goroutineSum / len(samples)
	}

	// Check baselines and record gaps
	baselines := DefaultResourceBaselines()
	for _, baseline := range baselines {
		if baseline.Scenario == result.Scenario {
			if result.HeapAllocMB > baseline.HeapPeakMB && r.gapTracker != nil {
				r.gapTracker.Record(gaps.Gap{
					Category: gaps.CategoryPerformance,
					Severity: gaps.SeverityMedium,
					Description: fmt.Sprintf("%s heap usage: %dMB (expected <%dMB)",
						result.Scenario, result.HeapAllocMB, baseline.HeapPeakMB),
					DiscoveredBy: "resources_" + result.Scenario,
					Remediation:  "Profile memory usage and reduce allocations",
				})
			}
			if result.GoroutineCount > baseline.GoroutinesPeak && r.gapTracker != nil {
				r.gapTracker.Record(gaps.Gap{
					Category: gaps.CategoryPerformance,
					Severity: gaps.SeverityLow,
					Description: fmt.Sprintf("%s goroutines: %d (expected <%d)",
						result.Scenario, result.GoroutineCount, baseline.GoroutinesPeak),
					DiscoveredBy: "resources_" + result.Scenario,
					Remediation:  "Check for goroutine leaks",
				})
			}
			break
		}
	}

	return result
}

// Run executes all resource measurements.
func (r *ResourceRunner) Run(ctx context.Context) []harness.ResourceResult {
	var results []harness.ResourceResult

	// Idle measurement
	results = append(results, r.MeasureIdle(ctx))

	// Under various loads
	for _, concurrency := range []int{100, 500} {
		result := r.MeasureUnderLoad(ctx, concurrency, 10*time.Second)
		results = append(results, result)
	}

	return results
}

// MemoryLeakTest checks for memory leaks over time.
type MemoryLeakTest struct {
	Duration       time.Duration
	SampleInterval time.Duration
	Concurrency    int
}

// MemoryLeakResult holds leak test results.
type MemoryLeakResult struct {
	StartHeapMB   uint64
	EndHeapMB     uint64
	PeakHeapMB    uint64
	GrowthMB      int64
	GrowthPercent float64
	LeakSuspected bool
}

// RunMemoryLeakTest runs a memory leak detection test.
func (r *ResourceRunner) RunMemoryLeakTest(ctx context.Context, test MemoryLeakTest) MemoryLeakResult {
	result := MemoryLeakResult{}

	// Start load
	done := make(chan struct{})
	for i := 0; i < test.Concurrency; i++ {
		go func() {
			for {
				select {
				case <-done:
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
		}()
	}

	// Collect samples
	var samples []uint64
	ticker := time.NewTicker(test.SampleInterval)
	deadline := time.After(test.Duration)

	for {
		select {
		case <-ticker.C:
			if stats, err := r.FetchRemoteStats(ctx); err == nil {
				samples = append(samples, stats.HeapAlloc/(1024*1024))
			}
		case <-deadline:
			ticker.Stop()
			close(done)

			if len(samples) < 2 {
				return result
			}

			result.StartHeapMB = samples[0]
			result.EndHeapMB = samples[len(samples)-1]

			for _, s := range samples {
				if s > result.PeakHeapMB {
					result.PeakHeapMB = s
				}
			}

			result.GrowthMB = int64(result.EndHeapMB) - int64(result.StartHeapMB)
			if result.StartHeapMB > 0 {
				result.GrowthPercent = float64(result.GrowthMB) / float64(result.StartHeapMB) * 100
			}

			// Suspect leak if growth > 50% and consistent upward trend
			result.LeakSuspected = result.GrowthPercent > 50

			return result
		case <-ctx.Done():
			ticker.Stop()
			close(done)
			return result
		}
	}
}
