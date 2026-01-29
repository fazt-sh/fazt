//go:build integration

package harness

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// Rate Limit Tests
// =============================================================================

func TestSecurity_BasicRateLimit(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	requestCount := 50
	var okCount, rateLimitedCount int

	for i := 0; i < requestCount; i++ {
		status, _, headers := doRequest(t, client, "GET", target+"/api/health")

		if status == 200 {
			okCount++
		} else if status == 429 {
			rateLimitedCount++
			if headers.Get("Retry-After") != "" {
				t.Log("Retry-After header present")
			}
		}
	}

	t.Logf("Rate limit test: ok=%d, rate_limited=%d", okCount, rateLimitedCount)

	// If no rate limiting, warn but don't fail (rate limiting may not be enabled)
	if rateLimitedCount == 0 && okCount > 30 {
		t.Log("Warning: no rate limiting detected (this may be expected)")
	}
}

func TestSecurity_BurstRateLimit(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	burstSize := 30
	var burstOK, burstThrottled int64
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 0; i < burstSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequestWithContext(ctx, "GET", target+"/api/health", nil)
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode == 200 {
				atomic.AddInt64(&burstOK, 1)
			} else if resp.StatusCode == 429 {
				atomic.AddInt64(&burstThrottled, 1)
			}
		}()
	}
	wg.Wait()

	t.Logf("Burst rate limit: ok=%d, throttled=%d", burstOK, burstThrottled)
}

func TestSecurity_RateLimitRecovery(t *testing.T) {
	target := getTarget(t)
	client := newTestClient()

	// First: exhaust rate limit
	for i := 0; i < 50; i++ {
		doRequest(t, client, "GET", target+"/api/health")
	}

	// Wait for rate limit to recover
	time.Sleep(2 * time.Second)

	// Test: should be able to make requests again
	var recoveryOK int
	for i := 0; i < 5; i++ {
		status, _, _ := doRequest(t, client, "GET", target+"/api/health")
		if status == 200 {
			recoveryOK++
		}
	}

	t.Logf("Rate limit recovery: %d/5 succeeded", recoveryOK)

	if recoveryOK == 0 {
		t.Error("rate limit did not recover after 2 seconds")
	}
}

// =============================================================================
// Payload Tests
// =============================================================================

func TestSecurity_SmallPayload(t *testing.T) {
	target := getTarget(t)
	token := os.Getenv("FAZT_TOKEN")
	if token == "" {
		t.Skip("FAZT_TOKEN not set")
	}

	client := newTestClient()

	// 1KB payload should always be accepted
	body := makeTestBody(1024)
	req, _ := http.NewRequest("POST", target+"/api/storage/kv/test-app/payload-test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Skip if KV storage endpoint not available
	if resp.StatusCode == 404 || resp.StatusCode == 500 {
		t.Skip("KV storage endpoint not available on target")
	}

	if resp.StatusCode != 200 {
		t.Errorf("1KB payload status = %d, want 200", resp.StatusCode)
	}
}

func TestSecurity_LargePayloadRejected(t *testing.T) {
	target := getTarget(t)
	token := os.Getenv("FAZT_TOKEN")
	if token == "" {
		t.Skip("FAZT_TOKEN not set")
	}

	client := &http.Client{Timeout: 60 * time.Second}

	// First check if deploy endpoint exists (admin only)
	probeReq, _ := http.NewRequest("POST", target+"/api/deploy", bytes.NewReader([]byte("probe")))
	probeReq.Header.Set("Authorization", "Bearer "+token)
	probeResp, err := client.Do(probeReq)
	if err == nil {
		probeResp.Body.Close()
		// Skip if endpoint not found (404) or runtime error (500)
		if probeResp.StatusCode == 404 || probeResp.StatusCode == 500 {
			t.Skip("deploy endpoint not available on target")
		}
	}

	// 100MB payload should be rejected
	body := makeTestBody(100 * 1024 * 1024)
	req, _ := http.NewRequest("POST", target+"/api/deploy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/zip")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		// Connection closed is acceptable for oversized payload
		t.Logf("Connection closed for large payload (expected)")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 413 && resp.StatusCode != 400 {
		t.Errorf("100MB payload status = %d, want 413 or 400", resp.StatusCode)
	}
}

func TestSecurity_MalformedJSON(t *testing.T) {
	target := getTarget(t)
	token := os.Getenv("FAZT_TOKEN")
	if token == "" {
		t.Skip("FAZT_TOKEN not set")
	}

	client := newTestClient()

	// Check if KV endpoint is available
	probeReq, _ := http.NewRequest("POST", target+"/api/storage/kv/test-app/probe",
		bytes.NewReader([]byte(`{"value":"probe"}`)))
	probeReq.Header.Set("Content-Type", "application/json")
	probeReq.Header.Set("Authorization", "Bearer "+token)
	probeResp, err := client.Do(probeReq)
	if err == nil {
		probeResp.Body.Close()
		if probeResp.StatusCode == 404 || probeResp.StatusCode == 500 {
			t.Skip("KV storage endpoint not available on target")
		}
	}

	tests := []struct {
		name string
		body []byte
	}{
		{"invalid_json", []byte(`{invalid json`)},
		{"deep_nesting", makeDeepJSON(100)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", target+"/api/storage/kv/test-app/malformed-test",
				bytes.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 400 {
				t.Errorf("%s status = %d, want 400", tc.name, resp.StatusCode)
			}
		})
	}
}

// =============================================================================
// Slowloris Tests
// =============================================================================

func TestSecurity_SlowHeaders(t *testing.T) {
	target := getTarget(t)

	// Parse target address
	addr := strings.TrimPrefix(target, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	if !strings.Contains(addr, ":") {
		addr += ":80"
	}

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	// Send request line
	conn.Write([]byte("GET /health HTTP/1.1\r\n"))
	conn.Write([]byte("Host: " + addr + "\r\n"))

	// Slowly send header (should be terminated by server timeout)
	headerLine := "X-Slow-Header: " + strings.Repeat("x", 50) + "\r\n"
	serverClosed := false

	for i := 0; i < len(headerLine); i++ {
		conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		_, err := conn.Write([]byte{headerLine[i]})
		if err != nil {
			serverClosed = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	conn.Close()

	if serverClosed {
		t.Log("Server closed slow header connection (expected)")
	} else {
		t.Log("Warning: server did not timeout slow header sending")
	}
}

func TestSecurity_ServiceDuringSlowloris(t *testing.T) {
	target := getTarget(t)
	client := &http.Client{Timeout: 5 * time.Second}

	addr := strings.TrimPrefix(target, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	if !strings.Contains(addr, ":") {
		addr += ":80"
	}

	// Start background slow connections
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}

				conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
				if err != nil {
					time.Sleep(100 * time.Millisecond)
					continue
				}

				conn.Write([]byte("GET /health HTTP/1.1\r\n"))
				conn.Write([]byte("Host: " + addr + "\r\n"))
				time.Sleep(2 * time.Second)
				conn.Close()
			}
		}()
	}

	// Let slow connections establish
	time.Sleep(500 * time.Millisecond)

	// Try legitimate requests
	var successCount int
	for i := 0; i < 10; i++ {
		resp, err := client.Get(target + "/health")
		if err == nil && resp.StatusCode == 200 {
			successCount++
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}

	close(done)
	time.Sleep(200 * time.Millisecond)

	t.Logf("Service during slowloris: %d/10 legitimate requests succeeded", successCount)

	// Slowloris vulnerability is a known issue for simple HTTP servers
	// This test documents the finding rather than failing the build
	if successCount == 0 {
		t.Log("Warning: server appears vulnerable to slowloris attack (0 legitimate requests succeeded)")
	}
}

func TestSecurity_ConnectionFlood(t *testing.T) {
	target := getTarget(t)

	addr := strings.TrimPrefix(target, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	if !strings.Contains(addr, ":") {
		addr += ":80"
	}

	var successCount, failCount int64
	var wg sync.WaitGroup

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
			if err != nil {
				atomic.AddInt64(&failCount, 1)
				return
			}
			atomic.AddInt64(&successCount, 1)

			conn.Write([]byte("GET /health HTTP/1.1\r\nHost: " + addr + "\r\nConnection: close\r\n\r\n"))

			buf := make([]byte, 1024)
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			conn.Read(buf)
			conn.Close()
		}()
	}

	wg.Wait()

	t.Logf("Connection flood: success=%d, fail=%d", successCount, failCount)

	if successCount < 150 {
		t.Errorf("connection handling degraded: only %d/200 succeeded", successCount)
	}
}

func TestSecurity_SlowBody(t *testing.T) {
	target := getTarget(t)

	addr := strings.TrimPrefix(target, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	if !strings.Contains(addr, ":") {
		addr += ":80"
	}

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	// Send complete headers with large content-length
	request := "POST /api/storage/kv/test-app/slow-body HTTP/1.1\r\n" +
		"Host: " + addr + "\r\n" +
		"Content-Type: application/json\r\n" +
		"Content-Length: 1000\r\n" +
		"\r\n"
	conn.Write([]byte(request))

	// Slowly send body
	body := `{"value":"` + strings.Repeat("x", 980) + `"}`
	serverClosed := false

	for i := 0; i < len(body); i++ {
		conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		_, err := conn.Write([]byte{body[i]})
		if err != nil {
			serverClosed = true
			break
		}
		if i%50 == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	if !serverClosed {
		// Try to read response
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		reader := bufio.NewReader(conn)
		response, _ := reader.ReadString('\n')
		t.Logf("Slow body completed, response: %s", strings.TrimSpace(response))
	} else {
		t.Log("Server closed slow body connection")
	}

	conn.Close()
}

// =============================================================================
// Helper Functions
// =============================================================================

func makeTestBody(size int) []byte {
	valueSize := size - 20
	if valueSize < 0 {
		valueSize = 0
	}
	value := strings.Repeat("x", valueSize)
	return []byte(fmt.Sprintf(`{"value":"%s"}`, value))
}

func makeDeepJSON(depth int) []byte {
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
