package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fazt-sh/fazt/internal/harness"
	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// APIWriteTest defines an API write operation test.
type APIWriteTest struct {
	Name        string
	Method      string
	Path        string
	Body        interface{}
	ContentType string
	Headers     map[string]string
	Expected    harness.Expected
	Description string
}

// DefaultAPIWriteTests returns standard API write tests.
func DefaultAPIWriteTests() []APIWriteTest {
	return []APIWriteTest{
		{
			Name:        "kv_set_string",
			Method:      "POST",
			Path:        "/api/storage/kv/test-app/test-key",
			Body:        map[string]interface{}{"value": "test-value"},
			ContentType: "application/json",
			Expected:    harness.Expect(200),
			Description: "KV set string value",
		},
		{
			Name:        "kv_set_json",
			Method:      "POST",
			Path:        "/api/storage/kv/test-app/json-key",
			Body:        map[string]interface{}{"value": map[string]interface{}{"nested": true}},
			ContentType: "application/json",
			Expected:    harness.Expect(200),
			Description: "KV set JSON value",
		},
		{
			Name:        "kv_set_with_ttl",
			Method:      "POST",
			Path:        "/api/storage/kv/test-app/ttl-key",
			Body:        map[string]interface{}{"value": "expires", "ttl": 60},
			ContentType: "application/json",
			Expected:    harness.Expect(200),
			Description: "KV set with TTL",
		},
		{
			Name:        "ds_insert",
			Method:      "POST",
			Path:        "/api/storage/ds/test-app/test-collection",
			Body:        map[string]interface{}{"name": "test", "count": 1},
			ContentType: "application/json",
			Expected:    harness.Expect(200).WithBody("id"),
			Description: "DS insert returns ID",
		},
		{
			Name:        "ds_update",
			Method:      "PUT",
			Path:        "/api/storage/ds/test-app/test-collection",
			Body:        map[string]interface{}{"query": map[string]interface{}{"name": "test"}, "changes": map[string]interface{}{"count": 2}},
			ContentType: "application/json",
			Expected:    harness.Expect(200),
			Description: "DS update by query",
		},
		{
			Name:        "ds_delete",
			Method:      "DELETE",
			Path:        "/api/storage/ds/test-app/test-collection",
			Body:        map[string]interface{}{"query": map[string]interface{}{"name": "test"}},
			ContentType: "application/json",
			Expected:    harness.Expect(200),
			Description: "DS delete by query",
		},
	}
}

// APIWriteRunner executes API write tests.
type APIWriteRunner struct {
	client     *http.Client
	baseURL    string
	authToken  string
	gapTracker *gaps.Tracker
}

// NewAPIWriteRunner creates a new API write test runner.
func NewAPIWriteRunner(baseURL, authToken string, gapTracker *gaps.Tracker) *APIWriteRunner {
	return &APIWriteRunner{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    baseURL,
		authToken:  authToken,
		gapTracker: gapTracker,
	}
}

// Run executes all API write tests.
func (r *APIWriteRunner) Run(ctx context.Context) []harness.TestResult {
	tests := DefaultAPIWriteTests()
	results := make([]harness.TestResult, 0, len(tests))

	for _, test := range tests {
		// Add auth header
		if test.Headers == nil {
			test.Headers = make(map[string]string)
		}
		if r.authToken != "" {
			test.Headers["Authorization"] = "Bearer " + r.authToken
		}

		result := r.runTest(ctx, test)
		results = append(results, result)
	}

	return results
}

func (r *APIWriteRunner) runTest(ctx context.Context, test APIWriteTest) harness.TestResult {
	start := time.Now()
	result := harness.TestResult{
		Name:     test.Name,
		Category: "api_write",
		Expected: test.Expected,
	}

	var bodyReader io.Reader
	if test.Body != nil {
		bodyBytes, err := json.Marshal(test.Body)
		if err != nil {
			result.Error = fmt.Errorf("failed to marshal body: %w", err)
			result.Duration = time.Since(start)
			return result
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, test.Method, r.baseURL+test.Path, bodyReader)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	if test.ContentType != "" {
		req.Header.Set("Content-Type", test.ContentType)
	}
	for k, v := range test.Headers {
		req.Header.Set(k, v)
	}

	resp, err := r.client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	result.Actual = harness.Actual{
		Status:  resp.StatusCode,
		Latency: result.Duration,
		Body:    string(body),
	}

	// Validate
	result.Passed = r.validate(test, resp, body)

	// Check for write queue issues
	if resp.StatusCode == http.StatusServiceUnavailable {
		if r.gapTracker != nil {
			r.gapTracker.Record(gaps.Gap{
				Category:     gaps.CategoryPerformance,
				Severity:     gaps.SeverityHigh,
				Description:  fmt.Sprintf("%s returned 503 - write queue may be full", test.Name),
				DiscoveredBy: "api_write_" + test.Name,
				Remediation:  "Check write queue capacity and throughput",
			})
		}
	}

	return result
}

func (r *APIWriteRunner) validate(test APIWriteTest, resp *http.Response, body []byte) bool {
	if resp.StatusCode != test.Expected.Status {
		return false
	}

	if test.Expected.BodyContains != "" {
		if !strings.Contains(string(body), test.Expected.BodyContains) {
			return false
		}
	}

	return true
}

// ConcurrentWriteTest tests write behavior under concurrent load.
type ConcurrentWriteTest struct {
	Name        string
	Concurrency int
	Writes      int // Per goroutine
	Expected    ConcurrentWriteExpected
}

// ConcurrentWriteExpected defines expected concurrent write behavior.
type ConcurrentWriteExpected struct {
	SuccessRate   float64 // Minimum success rate (0.0-1.0)
	MaxErrors     int
	QueueOverflow bool // Whether queue overflow is acceptable
}

// RunConcurrentWriteTests tests concurrent write handling.
func (r *APIWriteRunner) RunConcurrentWriteTests(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	test := ConcurrentWriteTest{
		Name:        "concurrent_kv_writes",
		Concurrency: 50,
		Writes:      20,
		Expected: ConcurrentWriteExpected{
			SuccessRate: 0.95,
		},
	}

	type writeResult struct {
		success bool
		err     error
		status  int
	}

	resultChan := make(chan writeResult, test.Concurrency*test.Writes)
	done := make(chan struct{})

	// Start workers
	for i := 0; i < test.Concurrency; i++ {
		go func(workerID int) {
			for j := 0; j < test.Writes; j++ {
				select {
				case <-done:
					return
				default:
				}

				key := fmt.Sprintf("test-key-%d-%d", workerID, j)
				body := bytes.NewReader([]byte(fmt.Sprintf(`{"value":"value-%d-%d"}`, workerID, j)))
				req, _ := http.NewRequestWithContext(ctx, "POST",
					r.baseURL+"/api/storage/kv/test-app/"+key, body)
				req.Header.Set("Content-Type", "application/json")
				if r.authToken != "" {
					req.Header.Set("Authorization", "Bearer "+r.authToken)
				}

				resp, err := r.client.Do(req)
				wr := writeResult{}
				if err != nil {
					wr.err = err
				} else {
					wr.status = resp.StatusCode
					wr.success = resp.StatusCode >= 200 && resp.StatusCode < 300
					resp.Body.Close()
				}
				resultChan <- wr
			}
		}(i)
	}

	// Collect results
	totalExpected := test.Concurrency * test.Writes
	var successCount, errorCount, queueFullCount int

	timeout := time.After(60 * time.Second)
	for i := 0; i < totalExpected; i++ {
		select {
		case wr := <-resultChan:
			if wr.success {
				successCount++
			} else {
				errorCount++
				if wr.status == http.StatusServiceUnavailable {
					queueFullCount++
				}
			}
		case <-timeout:
			close(done)
			break
		}
	}

	successRate := float64(successCount) / float64(totalExpected)
	passed := successRate >= test.Expected.SuccessRate

	results = append(results, harness.TestResult{
		Name:     test.Name,
		Category: "api_write",
		Passed:   passed,
		Actual: harness.Actual{
			Body: fmt.Sprintf("success=%d/%d (%.1f%%), queue_full=%d",
				successCount, totalExpected, successRate*100, queueFullCount),
		},
	})

	// Record gap if queue overflows too much
	if queueFullCount > totalExpected/10 && r.gapTracker != nil {
		r.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategoryPerformance,
			Severity:     gaps.SeverityHigh,
			Description:  fmt.Sprintf("Write queue overflow: %d/%d requests got 503", queueFullCount, totalExpected),
			DiscoveredBy: "concurrent_write_test",
			Remediation:  "Increase queue size or improve write throughput",
		})
	}

	return results
}

// TransactionIsolationTest verifies write isolation.
func (r *APIWriteRunner) RunTransactionTests(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	// Test: Write and immediate read should see new value
	testKey := fmt.Sprintf("isolation-test-%d", time.Now().UnixNano())
	testValue := "initial-value"

	// Write
	body := bytes.NewReader([]byte(fmt.Sprintf(`{"value":"%s"}`, testValue)))
	req, _ := http.NewRequestWithContext(ctx, "POST",
		r.baseURL+"/api/storage/kv/test-app/"+testKey, body)
	req.Header.Set("Content-Type", "application/json")
	if r.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.authToken)
	}

	resp, err := r.client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		results = append(results, harness.TestResult{
			Name:     "write_read_consistency",
			Category: "api_write",
			Passed:   false,
			Error:    err,
		})
		return results
	}
	resp.Body.Close()

	// Immediate read
	req2, _ := http.NewRequestWithContext(ctx, "GET",
		r.baseURL+"/api/storage/kv/test-app/"+testKey, nil)
	if r.authToken != "" {
		req2.Header.Set("Authorization", "Bearer "+r.authToken)
	}

	resp2, err := r.client.Do(req2)
	passed := false
	if err == nil && resp2.StatusCode == 200 {
		readBody, _ := io.ReadAll(resp2.Body)
		passed = strings.Contains(string(readBody), testValue)
		resp2.Body.Close()
	}

	results = append(results, harness.TestResult{
		Name:     "write_read_consistency",
		Category: "api_write",
		Passed:   passed,
	})

	return results
}
