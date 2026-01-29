//go:build integration

package harness

import (
	"context"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Latency Baseline Tests
// =============================================================================

func TestBaseline_HealthLatency(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	// Warmup
	for i := 0; i < 100; i++ {
		doRequest(t, client, "GET", target+"/api/health")
	}

	// Measure
	samples := 1000
	latencies := make([]time.Duration, 0, samples)

	for i := 0; i < samples; i++ {
		start := time.Now()
		resp, err := client.Get(target + "/health")
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			latencies = append(latencies, time.Since(start))
		}
	}

	if len(latencies) < samples/2 {
		t.Fatalf("too many failed requests: %d/%d succeeded", len(latencies), samples)
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p50 := latencies[len(latencies)*50/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]

	t.Logf("Health latency: P50=%v, P95=%v, P99=%v (n=%d)", p50, p95, p99, len(latencies))

	// Baseline expectations
	if p50 > 5*time.Millisecond {
		t.Errorf("P50 %v > 5ms threshold", p50)
	}
	if p95 > 20*time.Millisecond {
		t.Errorf("P95 %v > 20ms threshold", p95)
	}
	if p99 > 100*time.Millisecond {
		t.Errorf("P99 %v > 100ms threshold", p99)
	}
}

func TestBaseline_LatencyUnderLoad(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	concurrency := 50
	duration := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), duration+5*time.Second)
	defer cancel()

	var (
		latencies []time.Duration
		mu        sync.Mutex
		wg        sync.WaitGroup
	)

	done := make(chan struct{})

	// Start concurrent load
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
				req, _ := http.NewRequestWithContext(ctx, "GET", target+"/api/health", nil)
				resp, err := client.Do(req)
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

	time.Sleep(duration)
	close(done)
	wg.Wait()

	if len(latencies) == 0 {
		t.Fatal("no successful requests")
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p50 := latencies[len(latencies)*50/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]

	t.Logf("Latency under %d concurrent: P50=%v, P95=%v, P99=%v (n=%d)",
		concurrency, p50, p95, p99, len(latencies))

	// More relaxed thresholds under load
	if p99 > 500*time.Millisecond {
		t.Errorf("P99 %v > 500ms threshold under load", p99)
	}
}

// =============================================================================
// Throughput Baseline Tests
// =============================================================================

func TestBaseline_Throughput(t *testing.T) {
	target := getTarget(t)
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        500,
			MaxIdleConnsPerHost: 500,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Warmup
	warmup(t, target, 2*time.Second)

	// Thresholds are conservative for serverless overhead
	concurrencyLevels := []struct {
		concurrency int
		expectedRPS float64
	}{
		{1, 300},
		{10, 1500},
		{50, 2500},
		{100, 2500},
	}

	for _, level := range concurrencyLevels {
		t.Run(string(rune('0'+level.concurrency/10))+"0_concurrent", func(t *testing.T) {
			rps := measureThroughput(t, client, target, level.concurrency, 10*time.Second)
			t.Logf("Throughput @%d concurrent: %.0f RPS", level.concurrency, rps)

			// Allow 30% tolerance below expected
			minRPS := level.expectedRPS * 0.7
			if rps < minRPS {
				t.Errorf("RPS %.0f < %.0f threshold (expected %.0f)", rps, minRPS, level.expectedRPS)
			}
		})
	}
}

func TestBaseline_ThroughputScalability(t *testing.T) {
	target := getTarget(t)
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        500,
			MaxIdleConnsPerHost: 500,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	warmup(t, target, 2*time.Second)

	concurrencyLevels := []int{1, 10, 25, 50, 100}
	var results []struct {
		concurrency int
		rps         float64
		efficiency  float64
	}

	for _, concurrency := range concurrencyLevels {
		rps := measureThroughput(t, client, target, concurrency, 5*time.Second)
		efficiency := rps / float64(concurrency)

		results = append(results, struct {
			concurrency int
			rps         float64
			efficiency  float64
		}{concurrency, rps, efficiency})

		time.Sleep(500 * time.Millisecond) // Brief pause between measurements
	}

	t.Log("Scalability results:")
	for _, r := range results {
		t.Logf("  @%d concurrent: %.0f RPS (%.1f RPS/conn)", r.concurrency, r.rps, r.efficiency)
	}

	// Check that throughput increases with concurrency (at least initially)
	if len(results) >= 2 && results[1].rps <= results[0].rps {
		t.Error("throughput did not increase from 1 to 10 concurrent")
	}
}

func measureThroughput(t *testing.T, client *http.Client, target string, concurrency int, duration time.Duration) float64 {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), duration+5*time.Second)
	defer cancel()

	var (
		successCount int64
		wg           sync.WaitGroup
	)

	done := make(chan struct{})

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

				req, _ := http.NewRequestWithContext(ctx, "GET", target+"/api/health", nil)
				resp, err := client.Do(req)

				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					if resp.StatusCode >= 200 && resp.StatusCode < 400 {
						atomic.AddInt64(&successCount, 1)
					}
				}
			}
		}()
	}

	time.Sleep(duration)
	close(done)
	wg.Wait()

	return float64(successCount) / duration.Seconds()
}

// =============================================================================
// Resource Baseline Tests
// =============================================================================

func TestBaseline_IdleResources(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	// Recovery time after throughput tests - with retry
	var status int
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		status, _, _ = doRequest(t, client, "GET", target+"/api/health")
		if status == 200 {
			break
		}
	}
	if status != 200 {
		t.Fatalf("health check failed: %d", status)
	}

	// Note: Full resource measurement requires /api/system/stats endpoint
	// which may not be available. Just log what we can measure from client side.
	t.Log("Resource baseline: client-side only (server stats endpoint not tested)")
}

func TestBaseline_ResourcesUnderLoad(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	// Recovery time after previous tests
	time.Sleep(2 * time.Second)

	concurrency := 100
	duration := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), duration+5*time.Second)
	defer cancel()

	done := make(chan struct{})

	// Start load
	for i := 0; i < concurrency; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}
				req, _ := http.NewRequestWithContext(ctx, "GET", target+"/api/health", nil)
				resp, err := client.Do(req)
				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
			}
		}()
	}

	time.Sleep(duration)
	close(done)

	// Verify system eventually becomes responsive - allow more recovery time
	var lastStatus int
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		lastStatus, _, _ = doRequest(t, client, "GET", target+"/api/health")
		if lastStatus == 200 {
			break
		}
	}
	if lastStatus != 200 {
		t.Errorf("system unresponsive after load test: status=%d", lastStatus)
	}
}

// =============================================================================
// Mixed Workload Test
// =============================================================================

func TestBaseline_MixedWorkload(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	readPercent := 80 // 80% reads, 20% writes (simulated)
	concurrency := 50
	duration := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), duration+5*time.Second)
	defer cancel()

	var (
		readCount  int64
		writeCount int64
		wg         sync.WaitGroup
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

				isRead := (counter % 100) < readPercent
				counter++

				req, _ := http.NewRequestWithContext(ctx, "GET", target+"/api/health", nil)
				resp, err := client.Do(req)

				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					if resp.StatusCode >= 200 && resp.StatusCode < 400 {
						if isRead {
							atomic.AddInt64(&readCount, 1)
						} else {
							atomic.AddInt64(&writeCount, 1)
						}
					}
				}
			}
		}(i)
	}

	time.Sleep(duration)
	close(done)
	wg.Wait()

	readRPS := float64(readCount) / duration.Seconds()
	writeRPS := float64(writeCount) / duration.Seconds()
	totalRPS := float64(readCount+writeCount) / duration.Seconds()

	t.Logf("Mixed workload (%d%% read): read=%.0f RPS, write=%.0f RPS, total=%.0f RPS",
		readPercent, readRPS, writeRPS, totalRPS)

	if totalRPS < 1000 {
		t.Errorf("total RPS %.0f < 1000 threshold", totalRPS)
	}
}
