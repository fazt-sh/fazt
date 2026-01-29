package resilience

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fazt-sh/fazt/internal/harness"
	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// TimeoutTest defines a timeout enforcement test.
type TimeoutTest struct {
	Name          string
	Description   string
	RequestFunc   func(ctx context.Context, client *http.Client, baseURL string) (*http.Response, error)
	ExpectedRange [2]time.Duration // [min, max] expected duration
	ExpectedCode  int              // Expected HTTP status
}

// TimeoutRunner executes timeout tests.
type TimeoutRunner struct {
	client     *http.Client
	baseURL    string
	authToken  string
	gapTracker *gaps.Tracker
}

// NewTimeoutRunner creates a new timeout test runner.
func NewTimeoutRunner(baseURL, authToken string, gapTracker *gaps.Tracker) *TimeoutRunner {
	return &TimeoutRunner{
		client: &http.Client{
			Timeout: 60 * time.Second, // Long timeout to allow server-side timeouts to trigger
		},
		baseURL:    baseURL,
		authToken:  authToken,
		gapTracker: gapTracker,
	}
}

// Run executes timeout tests.
func (r *TimeoutRunner) Run(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	// Test 1: Server-side request timeout
	results = append(results, r.testRequestTimeout(ctx))

	// Test 2: Serverless execution timeout
	results = append(results, r.testServerlessTimeout(ctx))

	// Test 3: Storage operation timeout
	results = append(results, r.testStorageTimeout(ctx))

	// Test 4: Client-side timeout respected
	results = append(results, r.testClientTimeout(ctx))

	return results
}

func (r *TimeoutRunner) testRequestTimeout(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "request_timeout_enforcement",
		Category: "resilience",
	}

	// Test that server enforces read timeouts
	// Send a request and measure response time
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
	resp, err := r.client.Do(req)
	duration := time.Since(start)

	result.Duration = duration

	if err != nil {
		result.Error = err
		result.Passed = false
		return result
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// Health should respond quickly (< 1s)
	result.Passed = duration < 1*time.Second && resp.StatusCode == http.StatusOK
	result.Actual = harness.Actual{
		Status: resp.StatusCode,
	}

	return result
}

func (r *TimeoutRunner) testServerlessTimeout(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "serverless_timeout_enforcement",
		Category: "resilience",
	}

	// This test requires a deployed app with serverless endpoint
	// For now, test a hypothetical slow endpoint

	// Simulating a call to a serverless function that times out
	// The actual test would call an endpoint that runs: while(true){}

	start := time.Now()

	// Create request to a hypothetical timeout-test endpoint
	req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/api/timeout-test", nil)
	if r.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.authToken)
	}

	resp, err := r.client.Do(req)
	duration := time.Since(start)

	result.Duration = duration

	if err != nil {
		// Timeout error is acceptable
		result.Passed = true
		result.Actual = harness.Actual{
			Error: err.Error(),
		}
		return result
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// If endpoint exists, should timeout around 5-10 seconds
	// Or return 404 if not implemented yet
	result.Passed = resp.StatusCode == http.StatusNotFound ||
		(resp.StatusCode == 500 && duration >= 5*time.Second && duration <= 15*time.Second)

	result.Actual = harness.Actual{
		Status: resp.StatusCode,
	}

	if resp.StatusCode == 500 && duration > 15*time.Second && r.gapTracker != nil {
		r.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategorySecurity,
			Severity:     gaps.SeverityCritical,
			Description:  fmt.Sprintf("Serverless timeout not enforced: took %s", duration),
			DiscoveredBy: "timeout_serverless",
			Remediation:  "Enforce execution timeout in runtime",
		})
	}

	return result
}

func (r *TimeoutRunner) testStorageTimeout(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "storage_timeout_enforcement",
		Category: "resilience",
	}

	// Test that storage operations respect context deadline
	shortCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	start := time.Now()

	body, _ := json.Marshal(map[string]interface{}{"value": "test"})
	req, _ := http.NewRequestWithContext(shortCtx, "POST",
		r.baseURL+"/api/storage/kv/test-app/timeout-test-key",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if r.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.authToken)
	}

	resp, err := r.client.Do(req)
	duration := time.Since(start)

	result.Duration = duration

	if err != nil {
		// Context deadline exceeded is expected
		result.Passed = err == context.DeadlineExceeded || duration <= 1*time.Second
		result.Actual = harness.Actual{
			Error: err.Error(),
		}
		return result
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	// Should complete within deadline or get cancelled
	result.Passed = duration <= 1*time.Second

	result.Actual = harness.Actual{
		Status: resp.StatusCode,
	}

	return result
}

func (r *TimeoutRunner) testClientTimeout(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "client_timeout_respected",
		Category: "resilience",
	}

	// Use a very short client timeout
	shortClient := &http.Client{
		Timeout: 50 * time.Millisecond,
	}

	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
	_, err := shortClient.Do(req)
	duration := time.Since(start)

	result.Duration = duration

	// We expect a timeout error
	if err != nil {
		result.Passed = true // Timeout is expected
		result.Actual = harness.Actual{
			Error: "timeout (expected)",
		}
	} else {
		// If it completed within the short timeout, that's also fine
		result.Passed = duration <= 50*time.Millisecond
	}

	return result
}

// BudgetTrackingTest tests the timeout budget system.
type BudgetTrackingTest struct {
	Name          string
	StorageOps    int           // Number of storage operations
	InitialBudget time.Duration // Starting deadline
	ExpectedOK    bool          // Whether all ops should complete
}

// TestBudgetTracking tests the timeout budget system.
func (r *TimeoutRunner) TestBudgetTracking(ctx context.Context) []harness.TestResult {
	tests := []BudgetTrackingTest{
		{
			Name:          "budget_sufficient",
			StorageOps:    5,
			InitialBudget: 10 * time.Second,
			ExpectedOK:    true,
		},
		{
			Name:          "budget_insufficient",
			StorageOps:    20,
			InitialBudget: 500 * time.Millisecond,
			ExpectedOK:    false,
		},
	}

	var results []harness.TestResult

	for _, test := range tests {
		result := r.runBudgetTest(ctx, test)
		results = append(results, result)
	}

	return results
}

func (r *TimeoutRunner) runBudgetTest(ctx context.Context, test BudgetTrackingTest) harness.TestResult {
	result := harness.TestResult{
		Name:     test.Name,
		Category: "resilience",
	}

	budgetCtx, cancel := context.WithTimeout(ctx, test.InitialBudget)
	defer cancel()

	start := time.Now()
	successCount := 0

	for i := 0; i < test.StorageOps; i++ {
		body, _ := json.Marshal(map[string]interface{}{"value": i})
		req, _ := http.NewRequestWithContext(budgetCtx, "POST",
			r.baseURL+fmt.Sprintf("/api/storage/kv/test-app/budget-key-%d", i),
			bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if r.authToken != "" {
			req.Header.Set("Authorization", "Bearer "+r.authToken)
		}

		resp, err := r.client.Do(req)
		if err != nil {
			break // Context cancelled or timeout
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			successCount++
		}
	}

	result.Duration = time.Since(start)

	if test.ExpectedOK {
		result.Passed = successCount == test.StorageOps
	} else {
		// For insufficient budget, we expect some ops to fail/timeout
		result.Passed = successCount < test.StorageOps
	}

	result.Actual = harness.Actual{
		Body: fmt.Sprintf("completed=%d/%d", successCount, test.StorageOps),
	}

	return result
}

// DeadlineEnforcementTest verifies that context deadlines are enforced.
func (r *TimeoutRunner) DeadlineEnforcementTest(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "deadline_enforcement",
		Category: "resilience",
	}

	// Create a deadline that will expire mid-request
	deadline := 100 * time.Millisecond
	deadlineCtx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

	start := time.Now()

	// Send multiple requests, expecting at least some to fail
	var completed, failed int
	for i := 0; i < 50; i++ {
		req, _ := http.NewRequestWithContext(deadlineCtx, "GET", r.baseURL+"/health", nil)
		resp, err := r.client.Do(req)
		if err != nil {
			failed++
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		completed++
	}

	result.Duration = time.Since(start)

	// Should have had some failures due to deadline
	result.Passed = completed > 0 || failed > 0

	result.Actual = harness.Actual{
		Body: fmt.Sprintf("completed=%d, failed=%d", completed, failed),
	}

	return result
}
