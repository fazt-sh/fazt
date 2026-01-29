//go:build integration

// Package harness provides integration tests for fazt performance,
// resilience, and security validation.
package harness

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// getTarget returns the target URL from FAZT_TARGET env var.
// This should be a full URL to a test app, e.g., http://test-harness.192.168.64.3.nip.io:8080
// Skips the test if not set.
func getTarget(t *testing.T) string {
	t.Helper()
	target := os.Getenv("FAZT_TARGET")
	if target == "" {
		t.Skip("FAZT_TARGET not set")
	}
	return strings.TrimSuffix(target, "/")
}

// newTestClient creates an HTTP client configured for integration tests.
func newTestClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 200,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

// newTestClientNoRedirect creates a client that doesn't follow redirects.
func newTestClientNoRedirect() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// doRequest performs an HTTP request and returns status, body, and headers.
func doRequest(t *testing.T, client *http.Client, method, url string) (int, string, http.Header) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return resp.StatusCode, string(body), resp.Header
}

// doRequestWithHost performs an HTTP request with a custom Host header.
func doRequestWithHost(t *testing.T, client *http.Client, method, url, host string) (int, string, http.Header) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if host != "" {
		req.Host = host
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return resp.StatusCode, string(body), resp.Header
}

// assertStatus checks that the response status matches expected.
func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
}

// assertBodyContains checks that the response body contains the expected string.
func assertBodyContains(t *testing.T, body, want string) {
	t.Helper()
	if !strings.Contains(body, want) {
		t.Errorf("body does not contain %q, got: %s", want, truncate(body, 200))
	}
}

// assertHeaderPresent checks that a header is present.
func assertHeaderPresent(t *testing.T, headers http.Header, key string) {
	t.Helper()
	if headers.Get(key) == "" {
		t.Errorf("header %q not present", key)
	}
}

// assertHeaderValue checks a header has the expected value.
func assertHeaderValue(t *testing.T, headers http.Header, key, want string) {
	t.Helper()
	got := headers.Get(key)
	if got != want {
		t.Errorf("header %q = %q, want %q", key, got, want)
	}
}

// truncate shortens a string for error messages.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// measureLatency runs a function n times and returns P50, P95, P99.
func measureLatency(t *testing.T, n int, fn func()) (p50, p95, p99 time.Duration) {
	t.Helper()

	latencies := make([]time.Duration, 0, n)
	for i := 0; i < n; i++ {
		start := time.Now()
		fn()
		latencies = append(latencies, time.Since(start))
	}

	sortDurationsForTest(latencies)

	if len(latencies) == 0 {
		return 0, 0, 0
	}

	p50 = latencies[len(latencies)*50/100]
	p95 = latencies[len(latencies)*95/100]
	p99 = latencies[len(latencies)*99/100]
	return
}

// sortDurationsForTest sorts a slice of durations in place (for testing).
func sortDurationsForTest(d []time.Duration) {
	for i := 1; i < len(d); i++ {
		for j := i; j > 0 && d[j] < d[j-1]; j-- {
			d[j], d[j-1] = d[j-1], d[j]
		}
	}
}

// warmup sends requests to warm up the server.
func warmup(t *testing.T, target string, duration time.Duration) {
	t.Helper()

	client := newTestClient()
	end := time.Now().Add(duration)

	for time.Now().Before(end) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		req, _ := http.NewRequestWithContext(ctx, "GET", target+"/api/health", nil)
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
		cancel()
	}
}
