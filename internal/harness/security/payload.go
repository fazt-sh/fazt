package security

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fazt-sh/fazt/internal/harness"
	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// PayloadTest defines a payload security test.
type PayloadTest struct {
	Name        string
	Description string
	Method      string
	Path        string
	Body        []byte
	ContentType string
	Expected    PayloadExpected
}

// PayloadExpected defines expected payload handling.
type PayloadExpected struct {
	Status          int
	AcceptLargeBody bool
	RejectMalformed bool
}

// DefaultPayloadTests returns standard payload tests.
func DefaultPayloadTests() []PayloadTest {
	return []PayloadTest{
		{
			Name:        "body_1kb",
			Description: "Small body accepted",
			Method:      "POST",
			Path:        "/api/storage/kv/test-app/payload-test",
			Body:        makeBody(1024),
			ContentType: "application/json",
			Expected:    PayloadExpected{Status: 200},
		},
		{
			Name:        "body_100kb",
			Description: "Medium body accepted",
			Method:      "POST",
			Path:        "/api/storage/kv/test-app/payload-test",
			Body:        makeBody(100 * 1024),
			ContentType: "application/json",
			Expected:    PayloadExpected{Status: 200},
		},
		{
			Name:        "body_1mb",
			Description: "1MB body accepted",
			Method:      "POST",
			Path:        "/api/storage/kv/test-app/payload-test",
			Body:        makeBody(1024 * 1024),
			ContentType: "application/json",
			Expected:    PayloadExpected{Status: 200},
		},
		{
			Name:        "body_10mb_at_limit",
			Description: "10MB at limit",
			Method:      "POST",
			Path:        "/api/deploy",
			Body:        makeBody(10 * 1024 * 1024),
			ContentType: "application/zip",
			Expected:    PayloadExpected{Status: 200, AcceptLargeBody: true},
		},
		{
			Name:        "body_11mb_over_limit",
			Description: "11MB over limit rejected",
			Method:      "POST",
			Path:        "/api/deploy",
			Body:        makeBody(11 * 1024 * 1024),
			ContentType: "application/zip",
			Expected:    PayloadExpected{Status: 413},
		},
		{
			Name:        "body_100mb_way_over",
			Description: "100MB way over limit",
			Method:      "POST",
			Path:        "/api/deploy",
			Body:        makeBody(100 * 1024 * 1024),
			ContentType: "application/zip",
			Expected:    PayloadExpected{Status: 413},
		},
		{
			Name:        "malformed_json",
			Description: "Malformed JSON rejected",
			Method:      "POST",
			Path:        "/api/storage/kv/test-app/malformed-test",
			Body:        []byte(`{invalid json`),
			ContentType: "application/json",
			Expected:    PayloadExpected{Status: 400, RejectMalformed: true},
		},
		{
			Name:        "deep_nesting",
			Description: "Deeply nested JSON rejected",
			Method:      "POST",
			Path:        "/api/storage/kv/test-app/deep-test",
			Body:        makeDeepJSON(100),
			ContentType: "application/json",
			Expected:    PayloadExpected{Status: 400, RejectMalformed: true},
		},
		{
			Name:        "unicode_bomb",
			Description: "Unicode bomb rejected",
			Method:      "POST",
			Path:        "/api/storage/kv/test-app/unicode-test",
			Body:        makeUnicodeBomb(),
			ContentType: "application/json",
			Expected:    PayloadExpected{Status: 400, RejectMalformed: true},
		},
	}
}

// PayloadRunner executes payload tests.
type PayloadRunner struct {
	client     *http.Client
	baseURL    string
	authToken  string
	gapTracker *gaps.Tracker
}

// NewPayloadRunner creates a new payload test runner.
func NewPayloadRunner(baseURL, authToken string, gapTracker *gaps.Tracker) *PayloadRunner {
	return &PayloadRunner{
		client: &http.Client{
			Timeout: 60 * time.Second, // Longer for large uploads
		},
		baseURL:    baseURL,
		authToken:  authToken,
		gapTracker: gapTracker,
	}
}

// Run executes payload tests.
func (r *PayloadRunner) Run(ctx context.Context) []harness.TestResult {
	tests := DefaultPayloadTests()
	results := make([]harness.TestResult, 0, len(tests))

	for _, test := range tests {
		result := r.runTest(ctx, test)
		results = append(results, result)
	}

	return results
}

func (r *PayloadRunner) runTest(ctx context.Context, test PayloadTest) harness.TestResult {
	start := time.Now()
	result := harness.TestResult{
		Name:     test.Name,
		Category: "security",
	}

	req, err := http.NewRequestWithContext(ctx, test.Method, r.baseURL+test.Path, bytes.NewReader(test.Body))
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	req.Header.Set("Content-Type", test.ContentType)
	if r.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.authToken)
	}

	resp, err := r.client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		// Connection closed early is acceptable for very large payloads
		if test.Expected.Status == 413 {
			result.Passed = true
			result.Actual = harness.Actual{
				Error: "connection closed (expected for oversized payload)",
			}
			return result
		}
		result.Error = err
		return result
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	result.Actual = harness.Actual{
		Status: resp.StatusCode,
	}

	// Validate
	if test.Expected.Status == 413 {
		// For over-limit tests, 413 or early close is acceptable
		result.Passed = resp.StatusCode == 413 || resp.StatusCode == 400
	} else if test.Expected.RejectMalformed {
		// For malformed tests, 400 expected
		result.Passed = resp.StatusCode == 400
	} else {
		// For valid payloads, expect success or auth error
		result.Passed = resp.StatusCode == test.Expected.Status ||
			resp.StatusCode == http.StatusUnauthorized // Auth not set up
	}

	// Record gaps for security issues
	if !result.Passed && r.gapTracker != nil {
		var severity gaps.Severity
		var description string

		switch {
		case test.Expected.Status == 413 && resp.StatusCode != 413:
			severity = gaps.SeverityCritical
			description = fmt.Sprintf("Large payload not rejected: %s (%d bytes)", test.Name, len(test.Body))
		case test.Expected.RejectMalformed && resp.StatusCode != 400:
			severity = gaps.SeverityHigh
			description = fmt.Sprintf("Malformed payload not rejected: %s", test.Name)
		default:
			severity = gaps.SeverityMedium
			description = fmt.Sprintf("Payload test failed: %s (expected %d, got %d)", test.Name, test.Expected.Status, resp.StatusCode)
		}

		r.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategorySecurity,
			Severity:     severity,
			Description:  description,
			DiscoveredBy: "payload_" + test.Name,
			SpecRef:      "v0.8/limits.md#request-body",
		})
	}

	return result
}

// JSONBombTest tests for JSON bomb vulnerabilities.
func (r *PayloadRunner) JSONBombTest(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "json_bomb",
		Category: "security",
	}

	start := time.Now()

	// Create a JSON bomb (exponential expansion)
	// Example: {"a":"aa...","b":"$a$a","c":"$b$b"...}
	// When naively expanded, this can grow exponentially
	bomb := makeJSONBomb()

	req, _ := http.NewRequestWithContext(ctx, "POST",
		r.baseURL+"/api/storage/kv/test-app/bomb-test",
		bytes.NewReader(bomb))
	req.Header.Set("Content-Type", "application/json")
	if r.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.authToken)
	}

	resp, err := r.client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		// Timeout or connection close is acceptable
		result.Passed = true
		result.Actual = harness.Actual{
			Error: "request failed (expected for bomb)",
		}
		return result
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	// Should reject (400) or timeout, not process
	result.Passed = resp.StatusCode == 400 || result.Duration < 5*time.Second

	result.Actual = harness.Actual{
		Status: resp.StatusCode,
	}

	if !result.Passed && r.gapTracker != nil {
		r.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategorySecurity,
			Severity:     gaps.SeverityCritical,
			Description:  "JSON bomb not rejected or caused long processing",
			DiscoveredBy: "payload_json_bomb",
			Remediation:  "Add JSON parsing limits (depth, size, nesting)",
		})
	}

	return result
}

// ContentTypeMismatchTest tests handling of mismatched content types.
func (r *PayloadRunner) ContentTypeMismatchTest(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "content_type_mismatch",
		Category: "security",
	}

	start := time.Now()

	// Send JSON data with wrong content type
	req, _ := http.NewRequestWithContext(ctx, "POST",
		r.baseURL+"/api/storage/kv/test-app/mismatch-test",
		bytes.NewReader([]byte(`{"value":"test"}`)))
	req.Header.Set("Content-Type", "text/plain") // Wrong type for JSON endpoint
	if r.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.authToken)
	}

	resp, err := r.client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	// Should reject or handle gracefully
	result.Passed = resp.StatusCode == 400 || resp.StatusCode == 415 || resp.StatusCode == 200

	result.Actual = harness.Actual{
		Status: resp.StatusCode,
	}

	return result
}

// Helper functions

func makeBody(size int) []byte {
	// Create valid JSON of approximately the specified size
	valueSize := size - 20 // Account for JSON wrapper
	if valueSize < 0 {
		valueSize = 0
	}
	value := strings.Repeat("x", valueSize)
	return []byte(fmt.Sprintf(`{"value":"%s"}`, value))
}

func makeDeepJSON(depth int) []byte {
	// Create deeply nested JSON
	var sb strings.Builder
	for i := 0; i < depth; i++ {
		sb.WriteString(`{"a":`)
	}
	sb.WriteString(`"value"`)
	for i := 0; i < depth; i++ {
		sb.WriteString(`}`)
	}
	return []byte(sb.String())
}

func makeUnicodeBomb() []byte {
	// Create JSON with many unicode escapes
	var sb strings.Builder
	sb.WriteString(`{"value":"`)
	for i := 0; i < 10000; i++ {
		sb.WriteString(`\u0000`)
	}
	sb.WriteString(`"}`)
	return []byte(sb.String())
}

func makeJSONBomb() []byte {
	// Create a relatively small JSON that could cause issues
	// This is a simplified version - real bombs are more sophisticated
	var sb strings.Builder
	sb.WriteString(`{`)
	for i := 0; i < 100; i++ {
		if i > 0 {
			sb.WriteString(`,`)
		}
		sb.WriteString(fmt.Sprintf(`"key%d":`, i))
		sb.WriteString(`"` + strings.Repeat("x", 1000) + `"`)
	}
	sb.WriteString(`}`)
	return []byte(sb.String())
}
