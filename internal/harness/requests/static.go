// Package requests contains request lifecycle tests.
package requests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fazt-sh/fazt/internal/harness"
	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// StaticTest defines a static file serving test.
type StaticTest struct {
	Name        string
	Path        string
	Host        string // Optional subdomain.host header
	Expected    harness.Expected
	Description string
}

// DefaultStaticTests returns the standard static file tests.
func DefaultStaticTests() []StaticTest {
	return []StaticTest{
		{
			Name:        "api_health",
			Path:        "/api/system/health",
			Expected:    harness.Expect(200),
			Description: "System health API endpoint",
		},
		{
			Name:        "root_redirect",
			Path:        "/",
			Expected:    harness.Expect(303), // Redirect to login or dashboard
			Description: "Root redirects appropriately",
		},
		{
			Name:        "missing_file",
			Path:        "/nonexistent-file-12345.html",
			Expected:    harness.Expect(404),
			Description: "404 for missing files",
		},
		{
			Name:        "deep_path",
			Path:        "/a/b/c/d/e/file.html",
			Expected:    harness.Expect(404),
			Description: "Deep path handling",
		},
		{
			Name:        "dotfile_blocked",
			Path:        "/.env",
			Expected:    harness.Expect(403),
			Description: "Dotfiles should be blocked",
		},
		{
			Name:        "dotfile_gitignore",
			Path:        "/.gitignore",
			Expected:    harness.Expect(403),
			Description: "Git files blocked",
		},
		{
			Name:        "path_traversal_basic",
			Path:        "/../../../etc/passwd",
			Expected:    harness.Expect(400),
			Description: "Path traversal attempt",
		},
		{
			Name:        "path_traversal_encoded",
			Path:        "/%2e%2e/%2e%2e/etc/passwd",
			Expected:    harness.Expect(400),
			Description: "Encoded path traversal",
		},
		{
			Name:        "api_dir_blocked",
			Path:        "/api/",
			Expected:    harness.Expect(404),
			Description: "API directory listing blocked",
		},
	}
}

// StaticRunner executes static file tests.
type StaticRunner struct {
	client     *http.Client
	baseURL    string
	gapTracker *gaps.Tracker
}

// NewStaticRunner creates a new static test runner.
func NewStaticRunner(baseURL string, gapTracker *gaps.Tracker) *StaticRunner {
	return &StaticRunner{
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects
			},
		},
		baseURL:    baseURL,
		gapTracker: gapTracker,
	}
}

// Run executes all static tests and returns results.
func (r *StaticRunner) Run(ctx context.Context) []harness.TestResult {
	tests := DefaultStaticTests()
	results := make([]harness.TestResult, 0, len(tests))

	for _, test := range tests {
		result := r.runTest(ctx, test)
		results = append(results, result)
	}

	return results
}

// RunTest executes a single static test.
func (r *StaticRunner) runTest(ctx context.Context, test StaticTest) harness.TestResult {
	start := time.Now()
	result := harness.TestResult{
		Name:     test.Name,
		Category: "static",
		Expected: test.Expected,
	}

	url := r.baseURL + test.Path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	if test.Host != "" {
		req.Host = test.Host
	}

	resp, err := r.client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		// Check if this is an expected error
		if test.Expected.Error != "" && strings.Contains(err.Error(), test.Expected.Error) {
			result.Passed = true
		}
		return result
	}
	defer resp.Body.Close()

	// Read body for validation
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

	result.Actual = harness.Actual{
		Status:  resp.StatusCode,
		Latency: result.Duration,
		Body:    string(body),
		Headers: make(map[string]string),
	}

	// Copy relevant headers
	for _, key := range []string{"Content-Type", "ETag", "Cache-Control", "X-Content-Type-Options"} {
		if v := resp.Header.Get(key); v != "" {
			result.Actual.Headers[key] = v
		}
	}

	// Validate status
	result.Passed = resp.StatusCode == test.Expected.Status

	// Validate latency
	if test.Expected.MaxLatency > 0 && result.Duration > test.Expected.MaxLatency {
		result.Passed = false
	}

	// Validate body content
	if test.Expected.BodyContains != "" && !strings.Contains(string(body), test.Expected.BodyContains) {
		result.Passed = false
	}

	// Validate header
	if test.Expected.HeaderKey != "" {
		headerVal := resp.Header.Get(test.Expected.HeaderKey)
		if test.Expected.HeaderValue != "" && headerVal != test.Expected.HeaderValue {
			result.Passed = false
		} else if headerVal == "" {
			result.Passed = false
		}
	}

	// Record gap if test failed unexpectedly
	if !result.Passed && r.gapTracker != nil {
		gap := gaps.Gap{
			Category:     gaps.CategoryBehavior,
			Severity:     r.severityForTest(test.Name),
			Description:  fmt.Sprintf("%s: expected %d, got %d", test.Name, test.Expected.Status, resp.StatusCode),
			DiscoveredBy: "static_" + test.Name,
		}

		// Security-related failures are more severe
		if strings.Contains(test.Name, "traversal") || strings.Contains(test.Name, "dotfile") {
			gap.Category = gaps.CategorySecurity
			gap.Severity = gaps.SeverityCritical
		}

		gapID := r.gapTracker.Record(gap)
		result.Gap = &gaps.Gap{ID: gapID}
	}

	return result
}

func (r *StaticRunner) severityForTest(testName string) gaps.Severity {
	// Path traversal and dotfile tests are security-critical
	if strings.Contains(testName, "traversal") || strings.Contains(testName, "dotfile") {
		return gaps.SeverityCritical
	}
	// Missing file handling is medium
	if strings.Contains(testName, "missing") || strings.Contains(testName, "404") {
		return gaps.SeverityMedium
	}
	return gaps.SeverityLow
}

// VFSCacheTest tests VFS caching behavior.
type VFSCacheTest struct {
	Name        string
	Path        string
	Iterations  int
	Expected    VFSCacheExpected
	Description string
}

// VFSCacheExpected defines expected cache behavior.
type VFSCacheExpected struct {
	CacheHitAfter int           // After N requests, expect cache hit
	MaxLatency    time.Duration // Cached responses should be fast
}

// RunVFSCacheTests validates VFS caching.
func (r *StaticRunner) RunVFSCacheTests(ctx context.Context) []harness.TestResult {
	tests := []VFSCacheTest{
		{
			Name:       "repeated_file_access",
			Path:       "/health",
			Iterations: 10,
			Expected: VFSCacheExpected{
				CacheHitAfter: 2,
				MaxLatency:    5 * time.Millisecond,
			},
			Description: "Repeated access should hit cache",
		},
	}

	var results []harness.TestResult

	for _, test := range tests {
		var latencies []time.Duration

		for i := 0; i < test.Iterations; i++ {
			start := time.Now()
			req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+test.Path, nil)
			resp, err := r.client.Do(req)
			latency := time.Since(start)

			if err == nil {
				resp.Body.Close()
				latencies = append(latencies, latency)
			}
		}

		// Check that later requests are faster (indicating cache hit)
		passed := true
		if len(latencies) >= test.Iterations {
			avgLater := averageDuration(latencies[test.Expected.CacheHitAfter:])
			if avgLater > test.Expected.MaxLatency {
				passed = false
			}
		}

		results = append(results, harness.TestResult{
			Name:     test.Name,
			Category: "static",
			Passed:   passed,
		})
	}

	return results
}

// ETagTest tests ETag/conditional request handling.
func (r *StaticRunner) RunETagTests(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	// First request to get ETag
	req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
	resp, err := r.client.Do(req)
	if err != nil {
		return results
	}
	etag := resp.Header.Get("ETag")
	resp.Body.Close()

	if etag == "" {
		// No ETag support - record as a gap
		if r.gapTracker != nil {
			r.gapTracker.Record(gaps.Gap{
				Category:     gaps.CategoryPerformance,
				Severity:     gaps.SeverityMedium,
				Description:  "No ETag header for static files",
				DiscoveredBy: "etag_test",
				Remediation:  "Add ETag headers based on file hash",
			})
		}
		return results
	}

	// Conditional request with If-None-Match
	req2, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
	req2.Header.Set("If-None-Match", etag)
	resp2, err := r.client.Do(req2)
	if err != nil {
		return results
	}
	defer resp2.Body.Close()

	passed := resp2.StatusCode == http.StatusNotModified
	results = append(results, harness.TestResult{
		Name:     "etag_304_response",
		Category: "static",
		Passed:   passed,
		Actual: harness.Actual{
			Status: resp2.StatusCode,
		},
		Expected: harness.Expect(304),
	})

	return results
}

func averageDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	return sum / time.Duration(len(durations))
}
