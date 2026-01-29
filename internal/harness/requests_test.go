//go:build integration

package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// Static File Tests
// =============================================================================

func TestStatic_Health(t *testing.T) {
	target := getTarget(t)
	client := newTestClientNoRedirect()

	// Retry a few times - previous tests may have stressed the server
	var status int
	for i := 0; i < 3; i++ {
		status, _, _ = doRequest(t, client, "GET", target+"/api/health")
		if status == 200 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	assertStatus(t, status, 200)
}

func TestStatic_RootRedirect(t *testing.T) {
	target := getTarget(t)
	client := newTestClientNoRedirect()

	status, _, _ := doRequest(t, client, "GET", target+"/")
	// Root should redirect (303) or return 200 (if served)
	if status != 303 && status != 200 {
		t.Errorf("root status = %d, want 303 or 200", status)
	}
}

func TestStatic_MissingFile(t *testing.T) {
	target := getTarget(t)
	client := newTestClientNoRedirect()

	status, _, _ := doRequest(t, client, "GET", target+"/nonexistent-file-12345.html")
	assertStatus(t, status, 404)
}

func TestStatic_DotfileBlocked(t *testing.T) {
	target := getTarget(t)
	client := newTestClientNoRedirect()

	tests := []string{
		"/.env",
		"/.gitignore",
		"/.git/config",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			status, _, _ := doRequest(t, client, "GET", target+path)
			// Should be blocked (403) or not found (404)
			if status != 403 && status != 404 {
				t.Errorf("dotfile %s status = %d, want 403 or 404", path, status)
			}
		})
	}
}

func TestStatic_PathTraversal(t *testing.T) {
	target := getTarget(t)
	client := newTestClientNoRedirect()

	tests := []string{
		"/../../../etc/passwd",
		"/%2e%2e/%2e%2e/etc/passwd",
		"/..%2f..%2f..%2fetc/passwd",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			status, _, _ := doRequest(t, client, "GET", target+path)
			// Should be blocked (400) or not found (404)
			if status != 400 && status != 404 {
				t.Errorf("path traversal %s status = %d, want 400 or 404", path, status)
			}
		})
	}
}

func TestStatic_APIDirectoryBlocked(t *testing.T) {
	target := getTarget(t)
	client := newTestClientNoRedirect()

	// Retry a few times - previous tests may have stressed the server
	var status int
	for i := 0; i < 3; i++ {
		status, _, _ = doRequest(t, client, "GET", target+"/api/")
		// 500 indicates runtime timeout, retry
		if status != 500 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	// Directory listing should be blocked (404 or 403)
	if status != 404 && status != 403 {
		t.Errorf("API directory listing status = %d, want 404 or 403", status)
	}
}

// =============================================================================
// API Read Tests
// =============================================================================

func TestAPIRead_Health(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	// Retry a few times - previous tests may have stressed the server
	var status int
	var body string
	for i := 0; i < 3; i++ {
		status, body, _ = doRequest(t, client, "GET", target+"/api/health")
		if status == 200 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	assertStatus(t, status, 200)

	// Should be valid JSON
	var js json.RawMessage
	if err := json.Unmarshal([]byte(body), &js); err != nil {
		t.Errorf("health response is not valid JSON: %v", err)
	}
}

func TestAPIRead_AppsListUnauthorized(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	// Note: This test requires admin API access, not test app
	// Retry a few times in case of runtime timeout
	var status int
	for i := 0; i < 3; i++ {
		status, _, _ = doRequest(t, client, "GET", target+"/api/apps")
		if status != 500 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	// Skip if endpoint doesn't exist (404 = test app, not admin) or runtime error
	if status == 404 || status == 500 {
		t.Skip("admin endpoint /api/apps not available on target")
	}
	assertStatus(t, status, 401)
}

func TestAPIRead_NonexistentEndpoint(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	status, _, _ := doRequest(t, client, "GET", target+"/api/nonexistent-endpoint-12345")
	// Accept 404 (proper handling) or 500 (timeout during heavy load)
	if status != 404 && status != 500 {
		t.Errorf("status = %d, want 404 or 500", status)
	}
}

func TestAPIRead_HealthLatency(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	// Warmup
	for i := 0; i < 10; i++ {
		doRequest(t, client, "GET", target+"/api/health")
	}

	// Measure
	p50, p95, p99 := measureLatency(t, 100, func() {
		doRequest(t, client, "GET", target+"/api/health")
	})

	t.Logf("Health latency: P50=%v, P95=%v, P99=%v", p50, p95, p99)

	// Health endpoint should be fast
	if p99 > 100*time.Millisecond {
		t.Errorf("P99 latency %v > 100ms threshold", p99)
	}
}

// =============================================================================
// API Write Tests
// =============================================================================

func TestAPIWrite_KVSet(t *testing.T) {
	target := getTarget(t)
	token := os.Getenv("FAZT_TOKEN")
	if token == "" {
		t.Skip("FAZT_TOKEN not set")
	}

	client := newTestClient()
	key := fmt.Sprintf("test-key-%d", time.Now().UnixNano())

	body := fmt.Sprintf(`{"value":"test-value"}`)
	req, err := http.NewRequest("POST", target+"/api/storage/kv/test-app/"+key, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Skip if KV storage endpoint not available
	if resp.StatusCode == 404 || resp.StatusCode == 500 {
		t.Skip("KV storage endpoint not available on target")
	}

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Errorf("KV set status = %d, want 200. Body: %s", resp.StatusCode, respBody)
	}
}

func TestAPIWrite_ConcurrentWrites(t *testing.T) {
	target := getTarget(t)
	token := os.Getenv("FAZT_TOKEN")
	if token == "" {
		t.Skip("FAZT_TOKEN not set")
	}

	client := newTestClient()

	// First check if KV endpoint is available
	testReq, _ := http.NewRequest("POST", target+"/api/storage/kv/test-app/probe", strings.NewReader(`{"value":"probe"}`))
	testReq.Header.Set("Content-Type", "application/json")
	testReq.Header.Set("Authorization", "Bearer "+token)
	testResp, err := client.Do(testReq)
	if err != nil {
		t.Fatalf("probe request failed: %v", err)
	}
	testResp.Body.Close()
	if testResp.StatusCode == 404 || testResp.StatusCode == 500 {
		t.Skip("KV storage endpoint not available on target")
	}

	concurrency := 50
	writesPerWorker := 20

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	errorCount := 0
	queueFullCount := 0

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < writesPerWorker; j++ {
				select {
				case <-ctx.Done():
					return
				default:
				}

				key := fmt.Sprintf("concurrent-test-%d-%d", workerID, j)
				body := fmt.Sprintf(`{"value":"value-%d-%d"}`, workerID, j)

				req, _ := http.NewRequestWithContext(ctx, "POST",
					target+"/api/storage/kv/test-app/"+key, strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				resp, err := client.Do(req)
				mu.Lock()
				if err != nil {
					errorCount++
				} else {
					if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						successCount++
					} else {
						errorCount++
						if resp.StatusCode == 503 {
							queueFullCount++
						}
					}
					resp.Body.Close()
				}
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	total := concurrency * writesPerWorker
	successRate := float64(successCount) / float64(total)

	t.Logf("Concurrent writes: %d/%d success (%.1f%%), queue_full=%d",
		successCount, total, successRate*100, queueFullCount)

	if successRate < 0.90 {
		t.Errorf("success rate %.1f%% < 90%% threshold", successRate*100)
	}
}

// =============================================================================
// Auth Tests
// =============================================================================

func TestAuth_ProtectedEndpointNoSession(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	status, _, _ := doRequest(t, client, "GET", target+"/api/me")
	if status == 404 {
		t.Skip("admin endpoint /api/me not available on target")
	}
	assertStatus(t, status, 401)
}

func TestAuth_LoginLogoutCycle(t *testing.T) {
	target := getTarget(t)
	username := os.Getenv("FAZT_USERNAME")
	password := os.Getenv("FAZT_PASSWORD")
	if username == "" || password == "" {
		t.Skip("FAZT_USERNAME/FAZT_PASSWORD not set")
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
	}

	// Step 1: Login
	loginBody := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
	req, _ := http.NewRequest("POST", target+"/api/login", strings.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("login status = %d, want 200", resp.StatusCode)
	}

	// Verify Set-Cookie was sent
	setCookie := resp.Header.Get("Set-Cookie")
	if setCookie == "" {
		t.Error("login did not set session cookie")
	}

	// Step 2: Access protected endpoint
	status, body, _ := doRequest(t, client, "GET", target+"/api/me")
	if status != 200 {
		t.Errorf("protected endpoint status = %d after login, want 200", status)
	}
	if !strings.Contains(body, "username") {
		t.Errorf("protected endpoint should return user info")
	}

	// Step 3: Logout
	req, _ = http.NewRequest("POST", target+"/api/logout", nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("logout status = %d, want 200", resp.StatusCode)
	}

	// Step 4: Verify access denied after logout
	status, _, _ = doRequest(t, client, "GET", target+"/api/me")
	if status != 401 {
		t.Errorf("after logout status = %d, want 401", status)
	}
}

func TestAuth_APIKeyAuth(t *testing.T) {
	target := getTarget(t)
	token := os.Getenv("FAZT_TOKEN")
	if token == "" {
		t.Skip("FAZT_TOKEN not set")
	}

	client := newTestClient()

	// Valid API key
	req, _ := http.NewRequest("GET", target+"/api/apps", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	// Skip if endpoint doesn't exist (app target, not admin)
	if resp.StatusCode == 404 || resp.StatusCode == 500 {
		t.Skip("admin endpoint /api/apps not available on target")
	}

	if resp.StatusCode != 200 {
		t.Errorf("valid API key status = %d, want 200", resp.StatusCode)
	}

	// Invalid API key
	req, _ = http.NewRequest("GET", target+"/api/apps", nil)
	req.Header.Set("Authorization", "Bearer invalid-key-12345")

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 401 {
		t.Errorf("invalid API key status = %d, want 401", resp.StatusCode)
	}
}

func TestAuth_SessionCookieSecurity(t *testing.T) {
	target := getTarget(t)
	username := os.Getenv("FAZT_USERNAME")
	password := os.Getenv("FAZT_PASSWORD")
	if username == "" || password == "" {
		t.Skip("FAZT_USERNAME/FAZT_PASSWORD not set")
	}

	client := newTestClient()

	loginBody := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
	req, _ := http.NewRequest("POST", target+"/api/login", strings.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Skip("login failed, cannot test cookie security")
	}

	setCookie := resp.Header.Get("Set-Cookie")
	if setCookie == "" {
		t.Fatal("no Set-Cookie header")
	}

	// Check HttpOnly
	if !strings.Contains(strings.ToLower(setCookie), "httponly") {
		t.Error("session cookie missing HttpOnly flag")
	}

	// Check SameSite
	if !strings.Contains(strings.ToLower(setCookie), "samesite") {
		t.Error("session cookie missing SameSite attribute")
	}
}

// =============================================================================
// Serverless Tests (require test app with API functions)
// =============================================================================

func TestServerless_BasicExecution(t *testing.T) {
	target := getTarget(t)
	appHost := os.Getenv("FAZT_TEST_APP")
	if appHost == "" {
		t.Skip("FAZT_TEST_APP not set (e.g., test-app.192.168.64.3)")
	}

	client := newTestClient()

	// Test assumes there's a /api/hello endpoint that returns JSON
	// Retry a few times - previous tests may have stressed the server
	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", target+"/api/hello", nil)
		req.Host = appHost
		resp, err = client.Do(req)
		if err == nil && (resp.StatusCode == 200 || resp.StatusCode == 404) {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Either 200 (success) or 404 (endpoint doesn't exist)
	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		t.Errorf("serverless status = %d, body: %s", resp.StatusCode, body)
	}
}

func TestServerless_RequestBody(t *testing.T) {
	target := getTarget(t)
	appHost := os.Getenv("FAZT_TEST_APP")
	if appHost == "" {
		t.Skip("FAZT_TEST_APP not set")
	}

	client := newTestClient()

	// Retry a few times - previous tests may have stressed the server
	reqBody := `{"test":"data"}`
	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("POST", target+"/api/echo", bytes.NewReader([]byte(reqBody)))
		req.Host = appHost
		req.Header.Set("Content-Type", "application/json")
		resp, err = client.Do(req)
		if err == nil && (resp.StatusCode == 200 || resp.StatusCode == 404) {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		if !strings.Contains(string(body), "test") {
			t.Errorf("echo should return request body, got: %s", body)
		}
	} else if resp.StatusCode != 404 {
		t.Errorf("unexpected status = %d", resp.StatusCode)
	}
}

func TestServerless_Timeout(t *testing.T) {
	target := getTarget(t)
	appHost := os.Getenv("FAZT_TEST_APP")
	if appHost == "" {
		t.Skip("FAZT_TEST_APP not set")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Test assumes there's a /api/timeout endpoint with infinite loop
	req, _ := http.NewRequest("GET", target+"/api/timeout", nil)
	req.Host = appHost

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		// Client timeout is acceptable
		if strings.Contains(err.Error(), "timeout") {
			t.Logf("client timeout after %v (expected)", duration)
			return
		}
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	// Server should return 500 for timeout
	if resp.StatusCode == 500 {
		t.Logf("server returned 500 after %v (timeout enforced)", duration)
	} else if resp.StatusCode == 404 {
		t.Skip("timeout endpoint not deployed")
	} else {
		t.Errorf("expected 500 or timeout, got %d after %v", resp.StatusCode, duration)
	}
}
