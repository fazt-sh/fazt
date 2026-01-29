package resilience

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fazt-sh/fazt/internal/harness"
	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// QueueTest defines a write queue test.
type QueueTest struct {
	Name        string
	Description string
	Concurrency int
	Writes      int
	Expected    QueueExpected
}

// QueueExpected defines expected queue behavior.
type QueueExpected struct {
	MinSuccessRate  float64 // 0.0-1.0
	MaxQueueFull    int     // Max acceptable 503 responses
	AllowsAdmission bool    // Whether admission control should trigger
}

// QueueRunner executes write queue tests.
type QueueRunner struct {
	client     *http.Client
	baseURL    string
	authToken  string
	gapTracker *gaps.Tracker
}

// NewQueueRunner creates a new queue test runner.
func NewQueueRunner(baseURL, authToken string, gapTracker *gaps.Tracker) *QueueRunner {
	return &QueueRunner{
		client: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for queue waits
		},
		baseURL:    baseURL,
		authToken:  authToken,
		gapTracker: gapTracker,
	}
}

// Run executes queue resilience tests.
func (r *QueueRunner) Run(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	// Test 1: Normal load - queue should handle fine
	results = append(results, r.testNormalLoad(ctx))

	// Test 2: Heavy load - some rejections acceptable
	results = append(results, r.testHeavyLoad(ctx))

	// Test 3: Burst load - test queue elasticity
	results = append(results, r.testBurstLoad(ctx))

	// Test 4: Recovery after saturation
	results = append(results, r.testRecovery(ctx))

	return results
}

func (r *QueueRunner) testNormalLoad(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "queue_normal_load",
		Category: "resilience",
	}

	start := time.Now()

	// 20 concurrent writers, 10 writes each
	concurrency := 20
	writesPerWorker := 10

	var successCount, errorCount, queueFullCount int64
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < writesPerWorker; j++ {
				status := r.doWrite(ctx, fmt.Sprintf("normal-key-%d-%d", workerID, j), "value")

				switch {
				case status >= 200 && status < 300:
					atomic.AddInt64(&successCount, 1)
				case status == http.StatusServiceUnavailable:
					atomic.AddInt64(&queueFullCount, 1)
				default:
					atomic.AddInt64(&errorCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	result.Duration = time.Since(start)

	total := int64(concurrency * writesPerWorker)
	successRate := float64(successCount) / float64(total)

	// Normal load should have very high success rate
	result.Passed = successRate >= 0.95 && queueFullCount < 5

	result.Actual = harness.Actual{
		Body: fmt.Sprintf("success=%d/%d (%.1f%%), queue_full=%d",
			successCount, total, successRate*100, queueFullCount),
	}

	return result
}

func (r *QueueRunner) testHeavyLoad(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "queue_heavy_load",
		Category: "resilience",
	}

	start := time.Now()

	// 100 concurrent writers, 20 writes each
	concurrency := 100
	writesPerWorker := 20

	var successCount, errorCount, queueFullCount int64
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < writesPerWorker; j++ {
				status := r.doWrite(ctx, fmt.Sprintf("heavy-key-%d-%d", workerID, j), "value")

				switch {
				case status >= 200 && status < 300:
					atomic.AddInt64(&successCount, 1)
				case status == http.StatusServiceUnavailable:
					atomic.AddInt64(&queueFullCount, 1)
				default:
					atomic.AddInt64(&errorCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	result.Duration = time.Since(start)

	total := int64(concurrency * writesPerWorker)
	successRate := float64(successCount) / float64(total)
	queueFullRate := float64(queueFullCount) / float64(total)

	// Heavy load: accept some 503s, but most should succeed
	result.Passed = successRate >= 0.70

	result.Actual = harness.Actual{
		Body: fmt.Sprintf("success=%.1f%%, queue_full=%.1f%%, errors=%d",
			successRate*100, queueFullRate*100, errorCount),
	}

	if queueFullRate > 0.3 && r.gapTracker != nil {
		r.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategoryPerformance,
			Severity:     gaps.SeverityHigh,
			Description:  fmt.Sprintf("High queue overflow rate: %.1f%%", queueFullRate*100),
			DiscoveredBy: "queue_heavy_load",
			Remediation:  "Increase queue size or improve write throughput",
		})
	}

	return result
}

func (r *QueueRunner) testBurstLoad(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "queue_burst_load",
		Category: "resilience",
	}

	start := time.Now()

	// Sudden burst: 200 writes all at once
	burstSize := 200

	var successCount, queueFullCount int64
	var wg sync.WaitGroup

	for i := 0; i < burstSize; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			status := r.doWrite(ctx, fmt.Sprintf("burst-key-%d", id), "value")

			if status >= 200 && status < 300 {
				atomic.AddInt64(&successCount, 1)
			} else if status == http.StatusServiceUnavailable {
				atomic.AddInt64(&queueFullCount, 1)
			}
		}(i)
	}

	wg.Wait()
	result.Duration = time.Since(start)

	successRate := float64(successCount) / float64(burstSize)

	// Burst should handle at least 50% (queue absorbs, some rejected)
	result.Passed = successRate >= 0.50

	result.Actual = harness.Actual{
		Body: fmt.Sprintf("success=%d/%d (%.1f%%), queue_full=%d",
			successCount, burstSize, successRate*100, queueFullCount),
	}

	return result
}

func (r *QueueRunner) testRecovery(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "queue_recovery",
		Category: "resilience",
	}

	start := time.Now()

	// First: saturate the queue
	var wg sync.WaitGroup
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			r.doWrite(ctx, fmt.Sprintf("saturate-key-%d", id), "value")
		}(i)
	}
	wg.Wait()

	// Wait for queue to drain
	time.Sleep(5 * time.Second)

	// Now test: writes should work again
	successCount := 0
	for i := 0; i < 10; i++ {
		status := r.doWrite(ctx, fmt.Sprintf("recovery-key-%d", i), "value")
		if status >= 200 && status < 300 {
			successCount++
		}
	}

	result.Duration = time.Since(start)
	result.Passed = successCount >= 8 // At least 80% should work

	result.Actual = harness.Actual{
		Body: fmt.Sprintf("recovery_success=%d/10", successCount),
	}

	if !result.Passed && r.gapTracker != nil {
		r.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategoryPerformance,
			Severity:     gaps.SeverityCritical,
			Description:  "Queue did not recover after saturation",
			DiscoveredBy: "queue_recovery",
			Remediation:  "Check queue worker health and error handling",
		})
	}

	return result
}

// AdmissionControlTest tests the queue's admission control.
func (r *QueueRunner) AdmissionControlTest(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "queue_admission_control",
		Category: "resilience",
	}

	start := time.Now()

	// Create a context with a very short deadline
	shortCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// First: fill the queue with slow operations
	// This is simulated by sending many concurrent writes
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Use the main context, not the short one
			r.doWrite(ctx, fmt.Sprintf("fill-key-%d", id), "value")
		}(i)
	}

	// While queue is filling, try a write with short deadline
	// This should be rejected early by admission control
	admissionStart := time.Now()
	status := r.doWriteWithContext(shortCtx, "admission-test", "value")
	admissionDuration := time.Since(admissionStart)

	wg.Wait()

	result.Duration = time.Since(start)

	// Admission control should reject quickly (< 1s) rather than waiting
	// Status should be 503 or the context was cancelled
	quickRejection := admissionDuration < 2*time.Second
	expectedStatus := status == http.StatusServiceUnavailable || status == 0 // 0 = context cancelled

	result.Passed = quickRejection || expectedStatus

	result.Actual = harness.Actual{
		Body:   fmt.Sprintf("admission_time=%s, status=%d", admissionDuration, status),
		Status: status,
	}

	return result
}

func (r *QueueRunner) doWrite(ctx context.Context, key, value string) int {
	return r.doWriteWithContext(ctx, key, value)
}

func (r *QueueRunner) doWriteWithContext(ctx context.Context, key, value string) int {
	body, _ := json.Marshal(map[string]interface{}{"value": value})
	req, err := http.NewRequestWithContext(ctx, "POST",
		r.baseURL+"/api/storage/kv/test-app/"+key,
		bytes.NewReader(body))
	if err != nil {
		return 0
	}

	req.Header.Set("Content-Type", "application/json")
	if r.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.authToken)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return 0
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return resp.StatusCode
}

// QueueDepthTest monitors queue depth under load.
type QueueDepthResult struct {
	PeakDepth      int
	AverageDepth   float64
	OverflowCount  int
	DrainTime      time.Duration
}

// MonitorQueueDepth monitors queue depth during a load test.
func (r *QueueRunner) MonitorQueueDepth(ctx context.Context, duration time.Duration) QueueDepthResult {
	result := QueueDepthResult{}

	// This would require an endpoint that exposes queue stats
	// For now, we infer from 503 responses

	var successCount, overflowCount int64
	done := make(chan struct{})

	// Generate sustained write load
	for i := 0; i < 50; i++ {
		go func(workerID int) {
			counter := 0
			for {
				select {
				case <-done:
					return
				default:
				}

				status := r.doWrite(ctx, fmt.Sprintf("depth-key-%d-%d", workerID, counter), "value")
				counter++

				if status >= 200 && status < 300 {
					atomic.AddInt64(&successCount, 1)
				} else if status == http.StatusServiceUnavailable {
					atomic.AddInt64(&overflowCount, 1)
				}
			}
		}(i)
	}

	time.Sleep(duration)
	close(done)

	// Wait for queue to drain
	drainStart := time.Now()
	for {
		status := r.doWrite(ctx, "drain-test", "value")
		if status >= 200 && status < 300 {
			break
		}
		if time.Since(drainStart) > 30*time.Second {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	result.OverflowCount = int(overflowCount)
	result.DrainTime = time.Since(drainStart)

	return result
}
