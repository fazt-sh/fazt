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

// ServerlessTest defines a serverless function execution test.
type ServerlessTest struct {
	Name        string
	AppID       string
	Code        string // JavaScript code to execute
	Request     ServerlessRequest
	Expected    ServerlessExpected
	Description string
}

// ServerlessRequest represents the request to a serverless function.
type ServerlessRequest struct {
	Method  string
	Path    string
	Query   map[string]string
	Headers map[string]string
	Body    interface{}
}

// ServerlessExpected defines expected serverless execution results.
type ServerlessExpected struct {
	Status       int
	BodyContains string
	MaxDuration  time.Duration
	ErrorType    string // Expected error type (e.g., "TimeoutError")
}

// DefaultServerlessTests returns standard serverless execution tests.
func DefaultServerlessTests(appID string) []ServerlessTest {
	return []ServerlessTest{
		{
			Name:  "simple_return",
			AppID: appID,
			Code:  `return { ok: true, message: "hello" }`,
			Request: ServerlessRequest{
				Method: "GET",
				Path:   "/api/hello",
			},
			Expected: ServerlessExpected{
				Status:       200,
				BodyContains: "ok",
				MaxDuration:  100 * time.Millisecond,
			},
			Description: "Simple object return",
		},
		{
			Name:  "respond_helper",
			AppID: appID,
			Code:  `return respond({ message: "from respond" })`,
			Request: ServerlessRequest{
				Method: "GET",
				Path:   "/api/respond-test",
			},
			Expected: ServerlessExpected{
				Status:       200,
				BodyContains: "from respond",
			},
			Description: "Using respond() helper",
		},
		{
			Name:  "respond_status",
			AppID: appID,
			Code:  `return respond(201, { created: true })`,
			Request: ServerlessRequest{
				Method: "POST",
				Path:   "/api/create",
			},
			Expected: ServerlessExpected{
				Status: 201,
			},
			Description: "Custom status code",
		},
		{
			Name:  "request_body_access",
			AppID: appID,
			Code:  `return { received: request.body }`,
			Request: ServerlessRequest{
				Method: "POST",
				Path:   "/api/echo",
				Body:   map[string]interface{}{"test": "data"},
			},
			Expected: ServerlessExpected{
				Status:       200,
				BodyContains: "test",
			},
			Description: "Access request body",
		},
		{
			Name:  "request_query_access",
			AppID: appID,
			Code:  `return { name: request.query.name }`,
			Request: ServerlessRequest{
				Method: "GET",
				Path:   "/api/greet",
				Query:  map[string]string{"name": "world"},
			},
			Expected: ServerlessExpected{
				Status:       200,
				BodyContains: "world",
			},
			Description: "Access query parameters",
		},
		{
			Name:  "console_log",
			AppID: appID,
			Code: `
				console.log("test log message");
				return { logged: true }
			`,
			Request: ServerlessRequest{
				Method: "GET",
				Path:   "/api/logger",
			},
			Expected: ServerlessExpected{
				Status: 200,
			},
			Description: "Console.log works",
		},
		{
			Name:  "error_handling",
			AppID: appID,
			Code:  `throw new Error("intentional error")`,
			Request: ServerlessRequest{
				Method: "GET",
				Path:   "/api/error",
			},
			Expected: ServerlessExpected{
				Status:    500,
				ErrorType: "Error",
			},
			Description: "Error propagates correctly",
		},
		{
			Name:  "syntax_error",
			AppID: appID,
			Code:  `return { invalid syntax here`,
			Request: ServerlessRequest{
				Method: "GET",
				Path:   "/api/syntax",
			},
			Expected: ServerlessExpected{
				Status:    500,
				ErrorType: "SyntaxError",
			},
			Description: "Syntax error reported correctly",
		},
	}
}

// TimeoutTests returns tests for timeout behavior.
func TimeoutTests(appID string) []ServerlessTest {
	return []ServerlessTest{
		{
			Name:  "infinite_loop_timeout",
			AppID: appID,
			Code:  `while(true) {}`,
			Request: ServerlessRequest{
				Method: "GET",
				Path:   "/api/timeout",
			},
			Expected: ServerlessExpected{
				Status:      500,
				ErrorType:   "TimeoutError",
				MaxDuration: 15 * time.Second, // Should timeout around 10s
			},
			Description: "Infinite loop times out",
		},
		{
			Name:  "long_computation",
			AppID: appID,
			Code: `
				let sum = 0;
				for (let i = 0; i < 100000000; i++) {
					sum += i;
				}
				return { sum: sum };
			`,
			Request: ServerlessRequest{
				Method: "GET",
				Path:   "/api/compute",
			},
			Expected: ServerlessExpected{
				Status:      500,
				MaxDuration: 15 * time.Second,
			},
			Description: "Long computation times out",
		},
	}
}

// StorageTests returns tests for storage access from serverless.
func StorageTests(appID string) []ServerlessTest {
	return []ServerlessTest{
		{
			Name:  "kv_read",
			AppID: appID,
			Code:  `return fazt.storage.kv.get("test-key")`,
			Request: ServerlessRequest{
				Method: "GET",
				Path:   "/api/kv-read",
			},
			Expected: ServerlessExpected{
				Status:      200,
				MaxDuration: 200 * time.Millisecond,
			},
			Description: "KV read from serverless",
		},
		{
			Name:  "kv_write",
			AppID: appID,
			Code:  `fazt.storage.kv.set("from-serverless", "value"); return { ok: true }`,
			Request: ServerlessRequest{
				Method: "POST",
				Path:   "/api/kv-write",
			},
			Expected: ServerlessExpected{
				Status:      200,
				MaxDuration: 500 * time.Millisecond,
			},
			Description: "KV write from serverless",
		},
		{
			Name:  "ds_insert",
			AppID: appID,
			Code: `
				const id = fazt.storage.ds.insert("test-col", { name: "test" });
				return { id: id }
			`,
			Request: ServerlessRequest{
				Method: "POST",
				Path:   "/api/ds-insert",
			},
			Expected: ServerlessExpected{
				Status:       200,
				BodyContains: "id",
				MaxDuration:  500 * time.Millisecond,
			},
			Description: "DS insert from serverless",
		},
		{
			Name:  "multi_storage_ops",
			AppID: appID,
			Code: `
				fazt.storage.kv.set("key1", "value1");
				fazt.storage.kv.set("key2", "value2");
				const v1 = fazt.storage.kv.get("key1");
				const v2 = fazt.storage.kv.get("key2");
				return { v1: v1, v2: v2 }
			`,
			Request: ServerlessRequest{
				Method: "GET",
				Path:   "/api/multi-storage",
			},
			Expected: ServerlessExpected{
				Status:      200,
				MaxDuration: 1 * time.Second,
			},
			Description: "Multiple storage operations",
		},
	}
}

// ServerlessRunner executes serverless tests.
type ServerlessRunner struct {
	client     *http.Client
	baseURL    string
	authToken  string
	gapTracker *gaps.Tracker
}

// NewServerlessRunner creates a new serverless test runner.
func NewServerlessRunner(baseURL, authToken string, gapTracker *gaps.Tracker) *ServerlessRunner {
	return &ServerlessRunner{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    baseURL,
		authToken:  authToken,
		gapTracker: gapTracker,
	}
}

// Run executes serverless tests against an existing app.
func (r *ServerlessRunner) Run(ctx context.Context, appID string) []harness.TestResult {
	tests := DefaultServerlessTests(appID)
	results := make([]harness.TestResult, 0, len(tests))

	for _, test := range tests {
		result := r.runTest(ctx, test)
		results = append(results, result)
	}

	return results
}

// RunTimeoutTests runs tests specifically for timeout behavior.
func (r *ServerlessRunner) RunTimeoutTests(ctx context.Context, appID string) []harness.TestResult {
	// Use a longer timeout client for timeout tests
	longClient := &http.Client{
		Timeout: 60 * time.Second,
	}
	originalClient := r.client
	r.client = longClient
	defer func() { r.client = originalClient }()

	tests := TimeoutTests(appID)
	results := make([]harness.TestResult, 0, len(tests))

	for _, test := range tests {
		result := r.runTest(ctx, test)
		results = append(results, result)
	}

	return results
}

// RunStorageTests runs storage integration tests.
func (r *ServerlessRunner) RunStorageTests(ctx context.Context, appID string) []harness.TestResult {
	tests := StorageTests(appID)
	results := make([]harness.TestResult, 0, len(tests))

	for _, test := range tests {
		result := r.runTest(ctx, test)
		results = append(results, result)
	}

	return results
}

func (r *ServerlessRunner) runTest(ctx context.Context, test ServerlessTest) harness.TestResult {
	start := time.Now()
	result := harness.TestResult{
		Name:     test.Name,
		Category: "serverless",
	}

	// Build the request URL (app subdomain + path)
	// Format: http://app-name.domain.com/path
	// For testing, we use Host header
	url := r.baseURL + test.Request.Path
	if len(test.Request.Query) > 0 {
		url += "?"
		for k, v := range test.Request.Query {
			url += fmt.Sprintf("%s=%s&", k, v)
		}
		url = url[:len(url)-1] // Remove trailing &
	}

	var bodyReader io.Reader
	if test.Request.Body != nil {
		bodyBytes, _ := json.Marshal(test.Request.Body)
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, test.Request.Method, url, bodyReader)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	// Set app host header
	if test.AppID != "" {
		// Assumes baseURL like http://192.168.64.3:8080
		// Host should be app.192.168.64.3
		req.Host = test.AppID + ".local"
	}

	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range test.Request.Headers {
		req.Header.Set(k, v)
	}

	resp, err := r.client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		// Check if timeout was expected
		if test.Expected.ErrorType == "TimeoutError" && strings.Contains(err.Error(), "timeout") {
			result.Passed = true
		}
		return result
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))

	result.Actual = harness.Actual{
		Status:  resp.StatusCode,
		Latency: result.Duration,
		Body:    string(body),
	}

	// Validate
	result.Passed = r.validate(test, resp, body, result.Duration)

	// Record gaps
	if !result.Passed && r.gapTracker != nil {
		var severity gaps.Severity
		switch {
		case test.Expected.ErrorType == "TimeoutError" && resp.StatusCode != 500:
			severity = gaps.SeverityCritical
		case result.Duration > test.Expected.MaxDuration:
			severity = gaps.SeverityHigh
		default:
			severity = gaps.SeverityMedium
		}

		gap := gaps.Gap{
			Category:     gaps.CategoryBehavior,
			Severity:     severity,
			Description:  fmt.Sprintf("%s: expected %d got %d (%s)", test.Name, test.Expected.Status, resp.StatusCode, result.Duration),
			DiscoveredBy: "serverless_" + test.Name,
		}
		gapID := r.gapTracker.Record(gap)
		result.Gap = &gaps.Gap{ID: gapID}
	}

	return result
}

func (r *ServerlessRunner) validate(test ServerlessTest, resp *http.Response, body []byte, duration time.Duration) bool {
	// Status check
	if resp.StatusCode != test.Expected.Status {
		return false
	}

	// Duration check
	if test.Expected.MaxDuration > 0 && duration > test.Expected.MaxDuration {
		return false
	}

	// Body content check
	if test.Expected.BodyContains != "" {
		if !strings.Contains(string(body), test.Expected.BodyContains) {
			return false
		}
	}

	// Error type check
	if test.Expected.ErrorType != "" {
		if !strings.Contains(string(body), test.Expected.ErrorType) {
			return false
		}
	}

	return true
}

// MemoryTest tests memory limits in serverless.
type MemoryTest struct {
	Name        string
	AllocateMB  int
	ShouldFail  bool
	Description string
}

// RunMemoryTests tests serverless memory limits.
func (r *ServerlessRunner) RunMemoryTests(ctx context.Context, appID string) []harness.TestResult {
	tests := []MemoryTest{
		{Name: "small_alloc_ok", AllocateMB: 1, ShouldFail: false},
		{Name: "medium_alloc_ok", AllocateMB: 10, ShouldFail: false},
		{Name: "large_alloc_fail", AllocateMB: 100, ShouldFail: true},
	}

	var results []harness.TestResult

	for _, test := range tests {
		// Generate code that allocates memory
		code := fmt.Sprintf(`
			const arr = [];
			const mb = %d;
			for (let i = 0; i < mb * 1024; i++) {
				arr.push(new Array(1024).fill(0));
			}
			return { allocated: mb }
		`, test.AllocateMB)

		st := ServerlessTest{
			Name:  test.Name,
			AppID: appID,
			Code:  code,
			Request: ServerlessRequest{
				Method: "GET",
				Path:   "/api/memory-test",
			},
			Expected: ServerlessExpected{
				Status: 200,
			},
		}

		if test.ShouldFail {
			st.Expected.Status = 500
		}

		result := r.runTest(ctx, st)
		result.Name = test.Name
		results = append(results, result)
	}

	return results
}
