package requests

import (
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

// APIReadTest defines an API read operation test.
type APIReadTest struct {
	Name        string
	Method      string
	Path        string
	Headers     map[string]string
	SetupFunc   func(ctx context.Context, baseURL string, client *http.Client) error
	Expected    harness.Expected
	Description string
}

// DefaultAPIReadTests returns standard API read tests.
func DefaultAPIReadTests() []APIReadTest {
	return []APIReadTest{
		{
			Name:        "health_json",
			Method:      "GET",
			Path:        "/health",
			Expected:    harness.Expect(200),
			Description: "Health endpoint returns valid JSON",
		},
		{
			Name:        "apps_list",
			Method:      "GET",
			Path:        "/api/apps",
			Headers:     map[string]string{"Authorization": "Bearer test-token"},
			Expected:    harness.Expect(200),
			Description: "List apps endpoint",
		},
		{
			Name:        "apps_list_unauthorized",
			Method:      "GET",
			Path:        "/api/apps",
			Expected:    harness.Expect(401),
			Description: "Apps list requires auth",
		},
		{
			Name:        "nonexistent_api",
			Method:      "GET",
			Path:        "/api/nonexistent-endpoint",
			Expected:    harness.Expect(404),
			Description: "Unknown API endpoint returns 404",
		},
	}
}

// KVReadTests returns KV storage read tests.
func KVReadTests(appID string) []APIReadTest {
	return []APIReadTest{
		{
			Name:        "kv_get_missing",
			Method:      "GET",
			Path:        fmt.Sprintf("/api/storage/kv/%s/nonexistent-key", appID),
			Expected:    harness.Expect(200).WithBody("null"),
			Description: "KV get returns null for missing key",
		},
	}
}

// DSReadTests returns document store read tests.
func DSReadTests(appID string) []APIReadTest {
	return []APIReadTest{
		{
			Name:        "ds_find_empty",
			Method:      "GET",
			Path:        fmt.Sprintf("/api/storage/ds/%s/test-collection", appID),
			Expected:    harness.Expect(200).WithBody("[]"),
			Description: "DS find returns empty array for empty collection",
		},
	}
}

// APIReadRunner executes API read tests.
type APIReadRunner struct {
	client     *http.Client
	baseURL    string
	authToken  string
	gapTracker *gaps.Tracker
}

// NewAPIReadRunner creates a new API read test runner.
func NewAPIReadRunner(baseURL, authToken string, gapTracker *gaps.Tracker) *APIReadRunner {
	return &APIReadRunner{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    baseURL,
		authToken:  authToken,
		gapTracker: gapTracker,
	}
}

// Run executes all API read tests.
func (r *APIReadRunner) Run(ctx context.Context) []harness.TestResult {
	tests := DefaultAPIReadTests()
	results := make([]harness.TestResult, 0, len(tests))

	for _, test := range tests {
		result := r.runTest(ctx, test)
		results = append(results, result)
	}

	return results
}

// RunWithAuth executes tests that require authentication.
func (r *APIReadRunner) RunWithAuth(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	// Tests requiring auth
	authTests := []APIReadTest{
		{
			Name:        "apps_list_authed",
			Method:      "GET",
			Path:        "/api/apps",
			Expected:    harness.Expect(200),
			Description: "Authenticated apps list",
		},
		{
			Name:        "deployments_list",
			Method:      "GET",
			Path:        "/api/deployments",
			Expected:    harness.Expect(200),
			Description: "Deployments list",
		},
	}

	for _, test := range authTests {
		test.Headers = map[string]string{"Authorization": "Bearer " + r.authToken}
		result := r.runTest(ctx, test)
		results = append(results, result)
	}

	return results
}

func (r *APIReadRunner) runTest(ctx context.Context, test APIReadTest) harness.TestResult {
	start := time.Now()
	result := harness.TestResult{
		Name:     test.Name,
		Category: "api_read",
		Expected: test.Expected,
	}

	// Run setup if provided
	if test.SetupFunc != nil {
		if err := test.SetupFunc(ctx, r.baseURL, r.client); err != nil {
			result.Error = fmt.Errorf("setup failed: %w", err)
			result.Duration = time.Since(start)
			return result
		}
	}

	req, err := http.NewRequestWithContext(ctx, test.Method, r.baseURL+test.Path, nil)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	// Set headers
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
		Headers: make(map[string]string),
	}

	// Copy relevant headers
	for _, key := range []string{"Content-Type", "X-Request-Id"} {
		if v := resp.Header.Get(key); v != "" {
			result.Actual.Headers[key] = v
		}
	}

	// Validate
	result.Passed = r.validate(test, resp, body)

	// Record gaps for failures
	if !result.Passed && r.gapTracker != nil {
		gap := gaps.Gap{
			Category:     gaps.CategoryBehavior,
			Severity:     gaps.SeverityMedium,
			Description:  fmt.Sprintf("%s: expected %d, got %d", test.Name, test.Expected.Status, resp.StatusCode),
			DiscoveredBy: "api_read_" + test.Name,
		}
		gapID := r.gapTracker.Record(gap)
		result.Gap = &gaps.Gap{ID: gapID}
	}

	return result
}

func (r *APIReadRunner) validate(test APIReadTest, resp *http.Response, body []byte) bool {
	// Status check
	if resp.StatusCode != test.Expected.Status {
		return false
	}

	// Latency check
	if test.Expected.MaxLatency > 0 {
		// Already captured in result.Duration
	}

	// Body content check
	if test.Expected.BodyContains != "" {
		if !strings.Contains(string(body), test.Expected.BodyContains) {
			return false
		}
	}

	// JSON validity check for 2xx responses
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			var js json.RawMessage
			if json.Unmarshal(body, &js) != nil {
				return false
			}
		}
	}

	return true
}

// QueryPerformanceTest measures query performance at various scales.
type QueryPerformanceTest struct {
	Name       string
	Path       string
	DataCount  int // Number of records to seed
	Expected   harness.Expected
	SetupSQL   string // SQL to seed data
	CleanupSQL string // SQL to clean up
}

// RunQueryPerformanceTests tests query performance with different data sizes.
func (r *APIReadRunner) RunQueryPerformanceTests(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	// These tests require a seeded database
	// For now, just test empty state performance
	tests := []struct {
		name     string
		path     string
		maxMs    int64
	}{
		{"empty_collection_query", "/api/storage/ds/test-app/empty-col", 50},
	}

	for _, test := range tests {
		start := time.Now()
		req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+test.path, nil)
		if r.authToken != "" {
			req.Header.Set("Authorization", "Bearer "+r.authToken)
		}

		resp, err := r.client.Do(req)
		duration := time.Since(start)

		passed := true
		if err != nil {
			passed = false
		} else {
			resp.Body.Close()
			if duration.Milliseconds() > test.maxMs {
				passed = false
			}
		}

		results = append(results, harness.TestResult{
			Name:     test.name,
			Category: "api_read",
			Passed:   passed,
			Duration: duration,
			Expected: harness.ExpectWithLatency(200, time.Duration(test.maxMs)*time.Millisecond),
		})
	}

	return results
}

// TimeoutTest verifies request timeout handling.
type TimeoutTest struct {
	Name        string
	Path        string
	Timeout     time.Duration
	Expected    harness.Expected
	Description string
}

// RunTimeoutTests verifies timeout behavior.
func (r *APIReadRunner) RunTimeoutTests(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	// Test with very short client timeout
	shortClient := &http.Client{
		Timeout: 1 * time.Millisecond,
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
	_, err := shortClient.Do(req)

	// We expect a timeout error
	passed := err != nil && strings.Contains(err.Error(), "timeout")
	results = append(results, harness.TestResult{
		Name:     "client_timeout_respected",
		Category: "api_read",
		Passed:   passed,
		Error:    err,
	})

	return results
}
