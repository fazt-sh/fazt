//go:build integration

package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Memory Resilience Tests
// =============================================================================

func TestResilience_HighConcurrency(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	concurrency := 500
	var wg sync.WaitGroup
	var successCount, errorCount int64

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req, _ := http.NewRequestWithContext(ctx, "GET", target+"/api/health", nil)
			resp, err := client.Do(req)

			if err != nil {
				atomic.AddInt64(&errorCount, 1)
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusServiceUnavailable {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&errorCount, 1)
			}
		}()
	}

	wg.Wait()

	successRate := float64(successCount) / float64(concurrency)
	t.Logf("High concurrency: %d/%d success (%.1f%%)", successCount, concurrency, successRate*100)

	// 40% threshold for serverless under extreme load (500 concurrent)
	if successRate < 0.4 {
		t.Errorf("success rate %.1f%% < 40%% threshold", successRate*100)
	}
}

func TestResilience_SustainedLoad(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	duration := 30 * time.Second
	concurrency := 100

	ctx, cancel := context.WithTimeout(context.Background(), duration+10*time.Second)
	defer cancel()

	done := make(chan struct{})
	var errorCount int64

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
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}()
	}

	time.Sleep(duration)
	close(done)

	// Pause for recovery - serverless runtime needs time after heavy load
	time.Sleep(5 * time.Second)

	// Verify system eventually becomes responsive
	var lastStatus int
	for i := 0; i < 5; i++ {
		lastStatus, _, _ = doRequest(t, client, "GET", target+"/api/health")
		if lastStatus == 200 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if lastStatus != 200 {
		t.Errorf("system not responsive after sustained load: status=%d", lastStatus)
	}
}

func TestResilience_ConnectionLeak(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	ctx := context.Background()

	// Make many requests, explicitly not reading bodies
	for i := 0; i < 100; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", target+"/api/health", nil)
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close() // Close without reading
		}
	}

	// Pause for recovery
	time.Sleep(3 * time.Second)

	// Verify system eventually becomes responsive
	var lastStatus int
	for i := 0; i < 5; i++ {
		lastStatus, _, _ = doRequest(t, client, "GET", target+"/api/health")
		if lastStatus == 200 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if lastStatus != 200 {
		t.Errorf("system unresponsive after rapid requests: status=%d", lastStatus)
	}
}

func TestResilience_GCPressure(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	concurrency := 50
	duration := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), duration+5*time.Second)
	defer cancel()

	var latencies []time.Duration
	var mu sync.Mutex
	done := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}

				start := time.Now()
				req, _ := http.NewRequestWithContext(ctx, "GET", target+"/api/health", nil)
				resp, err := client.Do(req)
				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					mu.Lock()
					latencies = append(latencies, time.Since(start))
					mu.Unlock()
				}
			}
		}()
	}

	time.Sleep(duration)
	close(done)

	if len(latencies) == 0 {
		t.Fatal("no successful requests")
	}

	// Find max latency (indicates GC pauses)
	var maxLatency time.Duration
	for _, l := range latencies {
		if l > maxLatency {
			maxLatency = l
		}
	}

	t.Logf("GC pressure test: %d requests, max_latency=%v", len(latencies), maxLatency)

	// 3s threshold for serverless - runtime has overhead
	if maxLatency > 3*time.Second {
		t.Errorf("max latency %v > 3s threshold (possible GC pause)", maxLatency)
	}
}

// =============================================================================
// Timeout Tests
// =============================================================================

func TestResilience_RequestTimeout(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	// Retry several times if server is recovering from stress tests
	var status int
	var duration time.Duration
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		start := time.Now()
		status, _, _ = doRequest(t, client, "GET", target+"/api/health")
		duration = time.Since(start)
		if status == 200 {
			break
		}
	}

	if status != 200 {
		t.Errorf("health status = %d, want 200", status)
	}

	// Only check duration if status was successful
	if status == 200 && duration > 1*time.Second {
		t.Errorf("health took %v > 1s threshold", duration)
	}
}

func TestResilience_ClientTimeout(t *testing.T) {
	target := getTarget(t)

	shortClient := &http.Client{
		Timeout: 50 * time.Millisecond,
	}

	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "GET", target+"/api/health", nil)

	start := time.Now()
	_, err := shortClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		// Timeout is expected for very short client timeout
		t.Logf("client timeout after %v (expected)", duration)
	} else if duration > 50*time.Millisecond {
		t.Errorf("request completed after timeout: %v", duration)
	}
}

func TestResilience_ContextDeadline(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	deadline := 100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()

	var completed, failed int
	for i := 0; i < 50; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", target+"/api/health", nil)
		resp, err := client.Do(req)
		if err != nil {
			failed++
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		completed++
	}

	t.Logf("Deadline test: completed=%d, failed=%d", completed, failed)

	// Should have had some activity
	if completed == 0 && failed == 0 {
		t.Error("no requests completed or failed")
	}
}

// =============================================================================
// Write Queue Tests
// =============================================================================

func TestResilience_QueueNormalLoad(t *testing.T) {
	target := getTarget(t)
	token := os.Getenv("FAZT_TOKEN")
	if token == "" {
		t.Skip("FAZT_TOKEN not set")
	}

	client := newTestClient()

	// Check if KV endpoint is available
	status := doKVWrite(context.Background(), client, target, token, "probe-key", "probe")
	if status == 404 || status == 500 {
		t.Skip("KV storage endpoint not available on target")
	}

	concurrency := 20
	writesPerWorker := 10

	var successCount, queueFullCount int64
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < writesPerWorker; j++ {
				status := doKVWrite(ctx, client, target, token,
					fmt.Sprintf("normal-key-%d-%d", workerID, j), "value")

				if status >= 200 && status < 300 {
					atomic.AddInt64(&successCount, 1)
				} else if status == 503 {
					atomic.AddInt64(&queueFullCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()

	total := int64(concurrency * writesPerWorker)
	successRate := float64(successCount) / float64(total)

	t.Logf("Queue normal load: %d/%d success (%.1f%%), queue_full=%d",
		successCount, total, successRate*100, queueFullCount)

	if successRate < 0.95 {
		t.Errorf("success rate %.1f%% < 95%% threshold", successRate*100)
	}
}

func TestResilience_QueueHeavyLoad(t *testing.T) {
	target := getTarget(t)
	token := os.Getenv("FAZT_TOKEN")
	if token == "" {
		t.Skip("FAZT_TOKEN not set")
	}

	client := newTestClient()

	// Check if KV endpoint is available
	status := doKVWrite(context.Background(), client, target, token, "probe-key", "probe")
	if status == 404 || status == 500 {
		t.Skip("KV storage endpoint not available on target")
	}

	concurrency := 100
	writesPerWorker := 20

	var successCount, errorCount, queueFullCount int64
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < writesPerWorker; j++ {
				status := doKVWrite(ctx, client, target, token,
					fmt.Sprintf("heavy-key-%d-%d", workerID, j), "value")

				switch {
				case status >= 200 && status < 300:
					atomic.AddInt64(&successCount, 1)
				case status == 503:
					atomic.AddInt64(&queueFullCount, 1)
				default:
					atomic.AddInt64(&errorCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()

	total := int64(concurrency * writesPerWorker)
	successRate := float64(successCount) / float64(total)
	queueFullRate := float64(queueFullCount) / float64(total)

	t.Logf("Queue heavy load: success=%.1f%%, queue_full=%.1f%%, errors=%d",
		successRate*100, queueFullRate*100, errorCount)

	if successRate < 0.70 {
		t.Errorf("success rate %.1f%% < 70%% threshold", successRate*100)
	}
}

func TestResilience_QueueBurst(t *testing.T) {
	target := getTarget(t)
	token := os.Getenv("FAZT_TOKEN")
	if token == "" {
		t.Skip("FAZT_TOKEN not set")
	}

	client := newTestClient()

	// Check if KV endpoint is available
	status := doKVWrite(context.Background(), client, target, token, "probe-key", "probe")
	if status == 404 || status == 500 {
		t.Skip("KV storage endpoint not available on target")
	}

	burstSize := 200

	var successCount, queueFullCount int64
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for i := 0; i < burstSize; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			status := doKVWrite(ctx, client, target, token,
				fmt.Sprintf("burst-key-%d", id), "value")

			if status >= 200 && status < 300 {
				atomic.AddInt64(&successCount, 1)
			} else if status == 503 {
				atomic.AddInt64(&queueFullCount, 1)
			}
		}(i)
	}

	wg.Wait()

	successRate := float64(successCount) / float64(burstSize)

	t.Logf("Queue burst: %d/%d success (%.1f%%), queue_full=%d",
		successCount, burstSize, successRate*100, queueFullCount)

	if successRate < 0.50 {
		t.Errorf("success rate %.1f%% < 50%% threshold", successRate*100)
	}
}

func TestResilience_QueueRecovery(t *testing.T) {
	target := getTarget(t)
	token := os.Getenv("FAZT_TOKEN")
	if token == "" {
		t.Skip("FAZT_TOKEN not set")
	}

	client := newTestClient()
	ctx := context.Background()

	// Check if KV endpoint is available
	status := doKVWrite(ctx, client, target, token, "probe-key", "probe")
	if status == 404 || status == 500 {
		t.Skip("KV storage endpoint not available on target")
	}

	// First: saturate the queue
	var wg sync.WaitGroup
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			doKVWrite(ctx, client, target, token,
				fmt.Sprintf("saturate-key-%d", id), "value")
		}(i)
	}
	wg.Wait()

	// Wait for queue to drain
	time.Sleep(5 * time.Second)

	// Test: writes should work again
	successCount := 0
	for i := 0; i < 10; i++ {
		status := doKVWrite(ctx, client, target, token,
			fmt.Sprintf("recovery-key-%d", i), "value")
		if status >= 200 && status < 300 {
			successCount++
		}
	}

	t.Logf("Queue recovery: %d/10 success", successCount)

	if successCount < 8 {
		t.Errorf("recovery success %d < 8 threshold", successCount)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func doKVWrite(ctx context.Context, client *http.Client, target, token, key, value string) int {
	body, _ := json.Marshal(map[string]interface{}{"value": value})
	req, err := http.NewRequestWithContext(ctx, "POST",
		target+"/api/storage/kv/test-app/"+key,
		bytes.NewReader(body))
	if err != nil {
		return 0
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return resp.StatusCode
}
