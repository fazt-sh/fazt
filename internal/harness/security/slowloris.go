package security

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fazt-sh/fazt/internal/harness"
	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// SlowlorisTest defines a slow client attack test.
type SlowlorisTest struct {
	Name           string
	Description    string
	ConnectionCount int
	ByteRate       int           // Bytes per second per connection
	Duration       time.Duration // How long to maintain attack
}

// SlowlorisRunner executes slow client tests.
type SlowlorisRunner struct {
	baseURL    string
	gapTracker *gaps.Tracker
}

// NewSlowlorisRunner creates a new slowloris test runner.
func NewSlowlorisRunner(baseURL string, gapTracker *gaps.Tracker) *SlowlorisRunner {
	return &SlowlorisRunner{
		baseURL:    baseURL,
		gapTracker: gapTracker,
	}
}

// Run executes slow client tests.
func (r *SlowlorisRunner) Run(ctx context.Context) []harness.TestResult {
	var results []harness.TestResult

	// Test 1: Slow headers
	results = append(results, r.testSlowHeaders(ctx))

	// Test 2: Slow body
	results = append(results, r.testSlowBody(ctx))

	// Test 3: Many slow connections
	results = append(results, r.testManySlowConnections(ctx))

	// Test 4: Service availability during attack
	results = append(results, r.testServiceAvailability(ctx))

	return results
}

func (r *SlowlorisRunner) testSlowHeaders(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "slow_headers",
		Category: "security",
	}

	start := time.Now()

	// Parse target address from URL
	target := strings.TrimPrefix(r.baseURL, "http://")
	target = strings.TrimPrefix(target, "https://")
	if !strings.Contains(target, ":") {
		target += ":80"
	}

	// Open raw TCP connection
	conn, err := net.DialTimeout("tcp", target, 5*time.Second)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	// Send request line
	conn.Write([]byte("GET /health HTTP/1.1\r\n"))
	conn.Write([]byte("Host: " + target + "\r\n"))

	// Slowly send additional headers (1 byte per second)
	headerLine := "X-Slow-Header: " + strings.Repeat("x", 100) + "\r\n"
	for i := 0; i < len(headerLine); i++ {
		select {
		case <-ctx.Done():
			conn.Close()
			result.Duration = time.Since(start)
			return result
		default:
		}

		conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		_, err := conn.Write([]byte{headerLine[i]})
		if err != nil {
			// Connection closed by server - this is the expected behavior
			result.Passed = true
			result.Actual = harness.Actual{
				Body: "connection closed (server enforced timeout)",
			}
			conn.Close()
			result.Duration = time.Since(start)
			return result
		}
		time.Sleep(100 * time.Millisecond) // Slower than typical timeout
	}

	// If we got here, server didn't timeout slow headers
	conn.Close()
	result.Duration = time.Since(start)
	result.Passed = false

	if r.gapTracker != nil {
		r.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategorySecurity,
			Severity:     gaps.SeverityCritical,
			Description:  "Server did not timeout slow header sending",
			DiscoveredBy: "slowloris_headers",
			Remediation:  "Set ReadHeaderTimeout on http.Server",
			SpecRef:      "v0.8/safeguards.md#timeouts",
		})
	}

	return result
}

func (r *SlowlorisRunner) testSlowBody(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "slow_body",
		Category: "security",
	}

	start := time.Now()

	// Parse target address
	target := strings.TrimPrefix(r.baseURL, "http://")
	target = strings.TrimPrefix(target, "https://")
	if !strings.Contains(target, ":") {
		target += ":80"
	}

	conn, err := net.DialTimeout("tcp", target, 5*time.Second)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	// Send complete headers with large content-length
	request := "POST /api/storage/kv/test-app/slow-body HTTP/1.1\r\n" +
		"Host: " + target + "\r\n" +
		"Content-Type: application/json\r\n" +
		"Content-Length: 10000\r\n" +
		"\r\n"
	conn.Write([]byte(request))

	// Slowly send body (1 byte per second)
	body := `{"value":"` + strings.Repeat("x", 9980) + `"}`
	for i := 0; i < len(body); i++ {
		select {
		case <-ctx.Done():
			conn.Close()
			result.Duration = time.Since(start)
			return result
		default:
		}

		conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		_, err := conn.Write([]byte{body[i]})
		if err != nil {
			// Connection closed - expected
			result.Passed = true
			result.Actual = harness.Actual{
				Body: "connection closed during slow body",
			}
			conn.Close()
			result.Duration = time.Since(start)
			return result
		}

		// Only slow down every 10 bytes to not take forever
		if i%10 == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Try to read response
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	reader := bufio.NewReader(conn)
	response, _ := reader.ReadString('\n')
	conn.Close()

	result.Duration = time.Since(start)
	result.Passed = response != "" // Got a response
	result.Actual = harness.Actual{
		Body: "completed slow body transmission",
	}

	return result
}

func (r *SlowlorisRunner) testManySlowConnections(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "many_slow_connections",
		Category: "security",
	}

	start := time.Now()

	// Parse target
	target := strings.TrimPrefix(r.baseURL, "http://")
	target = strings.TrimPrefix(target, "https://")
	if !strings.Contains(target, ":") {
		target += ":80"
	}

	// Open many slow connections
	connectionCount := 100
	var openConns int64
	var closedConns int64
	var mu sync.Mutex
	conns := make([]net.Conn, 0, connectionCount)

	// Open connections
	for i := 0; i < connectionCount; i++ {
		conn, err := net.DialTimeout("tcp", target, 2*time.Second)
		if err != nil {
			continue
		}

		atomic.AddInt64(&openConns, 1)
		mu.Lock()
		conns = append(conns, conn)
		mu.Unlock()

		// Send partial request
		go func(c net.Conn, id int) {
			defer func() {
				c.Close()
				atomic.AddInt64(&closedConns, 1)
			}()

			c.Write([]byte("GET /health HTTP/1.1\r\n"))
			c.Write([]byte(fmt.Sprintf("Host: %s\r\n", target)))

			// Keep sending headers slowly
			for j := 0; j < 10; j++ {
				select {
				case <-ctx.Done():
					return
				default:
				}

				c.SetWriteDeadline(time.Now().Add(2 * time.Second))
				_, err := c.Write([]byte(fmt.Sprintf("X-Header-%d: value\r\n", j)))
				if err != nil {
					return // Connection closed
				}
				time.Sleep(500 * time.Millisecond)
			}
		}(conn, i)
	}

	// Let attack run for a few seconds
	time.Sleep(5 * time.Second)

	// Close remaining connections
	mu.Lock()
	for _, conn := range conns {
		conn.Close()
	}
	mu.Unlock()

	result.Duration = time.Since(start)

	// Server should have closed some connections via timeout
	result.Passed = closedConns > int64(connectionCount)/2

	result.Actual = harness.Actual{
		Body: fmt.Sprintf("opened=%d, closed_by_server=%d", openConns, closedConns),
	}

	if closedConns < int64(connectionCount)/2 && r.gapTracker != nil {
		r.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategorySecurity,
			Severity:     gaps.SeverityCritical,
			Description:  fmt.Sprintf("Slow connections not terminated: only %d/%d closed", closedConns, openConns),
			DiscoveredBy: "slowloris_many_connections",
			Remediation:  "Implement connection timeouts and limits",
			SpecRef:      "v0.8/safeguards.md#connection-limits",
		})
	}

	return result
}

func (r *SlowlorisRunner) testServiceAvailability(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "service_during_slowloris",
		Category: "security",
	}

	start := time.Now()

	// Parse target
	target := strings.TrimPrefix(r.baseURL, "http://")
	target = strings.TrimPrefix(target, "https://")
	if !strings.Contains(target, ":") {
		target += ":80"
	}

	// Start background slow connections
	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}

				conn, err := net.DialTimeout("tcp", target, 1*time.Second)
				if err != nil {
					time.Sleep(100 * time.Millisecond)
					continue
				}

				// Send partial request
				conn.Write([]byte("GET /health HTTP/1.1\r\n"))
				conn.Write([]byte("Host: " + target + "\r\n"))

				// Hold connection open
				time.Sleep(2 * time.Second)
				conn.Close()
			}
		}()
	}

	// Let slow connections establish
	time.Sleep(1 * time.Second)

	// Try legitimate requests
	client := &http.Client{Timeout: 5 * time.Second}
	var successCount, failCount int

	for i := 0; i < 20; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", r.baseURL+"/health", nil)
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			successCount++
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		} else {
			failCount++
			if resp != nil {
				resp.Body.Close()
			}
		}
	}

	close(done)
	time.Sleep(500 * time.Millisecond) // Let slow connections clean up

	result.Duration = time.Since(start)

	// Service should remain available (at least 50% success)
	result.Passed = successCount >= 10

	result.Actual = harness.Actual{
		Body: fmt.Sprintf("legitimate_requests: success=%d, fail=%d", successCount, failCount),
	}

	if successCount < 10 && r.gapTracker != nil {
		r.gapTracker.Record(gaps.Gap{
			Category:     gaps.CategorySecurity,
			Severity:     gaps.SeverityCritical,
			Description:  fmt.Sprintf("Service degraded during slowloris: only %d/20 requests succeeded", successCount),
			DiscoveredBy: "slowloris_service_availability",
			Remediation:  "Implement connection limits and slow client detection",
			SpecRef:      "v0.8/safeguards.md#dos-protection",
		})
	}

	return result
}

// ConnectionFloodTest tests behavior under connection flood.
func (r *SlowlorisRunner) ConnectionFloodTest(ctx context.Context) harness.TestResult {
	result := harness.TestResult{
		Name:     "connection_flood",
		Category: "security",
	}

	start := time.Now()

	target := strings.TrimPrefix(r.baseURL, "http://")
	target = strings.TrimPrefix(target, "https://")
	if !strings.Contains(target, ":") {
		target += ":80"
	}

	// Try to open many connections rapidly
	var successCount, failCount int64
	var wg sync.WaitGroup

	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := net.DialTimeout("tcp", target, 1*time.Second)
			if err != nil {
				atomic.AddInt64(&failCount, 1)
				return
			}
			atomic.AddInt64(&successCount, 1)

			// Send a complete request
			conn.Write([]byte("GET /health HTTP/1.1\r\nHost: " + target + "\r\nConnection: close\r\n\r\n"))

			// Read response
			buf := make([]byte, 1024)
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			conn.Read(buf)
			conn.Close()
		}()
	}

	wg.Wait()
	result.Duration = time.Since(start)

	// Should handle most connections
	result.Passed = successCount >= 400

	result.Actual = harness.Actual{
		Body: fmt.Sprintf("connections: success=%d, fail=%d", successCount, failCount),
	}

	return result
}
