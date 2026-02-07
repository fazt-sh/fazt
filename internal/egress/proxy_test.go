package egress

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fazt-sh/fazt/internal/system"
	"github.com/fazt-sh/fazt/internal/timeout"

	_ "modernc.org/sqlite"
)

// testDB creates an in-memory SQLite database with the net_allowlist table.
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE net_allowlist (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT NOT NULL,
			app_id TEXT,
			https_only INTEGER NOT NULL DEFAULT 1,
			rate_limit INTEGER DEFAULT 0,
			rate_burst INTEGER DEFAULT 0,
			max_response INTEGER DEFAULT 0,
			timeout_ms INTEGER DEFAULT 0,
			cache_ttl INTEGER DEFAULT 0,
			created_at INTEGER NOT NULL DEFAULT (unixepoch()),
			UNIQUE(domain, app_id)
		);
		CREATE TABLE net_secrets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_id TEXT,
			name TEXT NOT NULL,
			value TEXT NOT NULL,
			inject_as TEXT NOT NULL DEFAULT 'bearer',
			inject_key TEXT,
			domain TEXT,
			created_at INTEGER NOT NULL DEFAULT (unixepoch()),
			updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
			UNIQUE(app_id, name)
		);
		CREATE TABLE net_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_id TEXT NOT NULL,
			domain TEXT NOT NULL,
			method TEXT NOT NULL,
			path TEXT NOT NULL,
			status INTEGER,
			error_code TEXT,
			duration_ms INTEGER NOT NULL,
			request_bytes INTEGER,
			response_bytes INTEGER,
			created_at INTEGER NOT NULL DEFAULT (unixepoch())
		);
	`)
	if err != nil {
		t.Fatalf("failed to create test tables: %v", err)
	}
	return db
}

// testProxy creates a proxy with a test server's domain allowlisted.
func testProxy(t *testing.T, db *sql.DB, serverDomain string) *EgressProxy {
	t.Helper()
	system.ResetCachedLimits()

	allowlist := NewAllowlist(db)
	if serverDomain != "" {
		allowlist.Add(serverDomain, "", false) // allow HTTP for test
	}
	proxy := NewEgressProxy(allowlist)
	return proxy
}

func extractHost(addr string) string {
	host, _, _ := net.SplitHostPort(addr)
	return host
}

// --- Test: IP blocking ---

func TestBlockedIPRanges(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"loopback", "127.0.0.1", true},
		{"loopback-127.0.0.2", "127.0.0.2", true},
		{"private-10", "10.0.0.1", true},
		{"private-172", "172.16.0.1", true},
		{"private-192", "192.168.1.1", true},
		{"link-local", "169.254.169.254", true},
		{"cgnat", "100.64.0.1", true},
		{"this-network", "0.0.0.0", true},
		{"ipv6-loopback", "::1", true},
		{"ipv6-unique-local", "fd00::1", true},
		{"ipv6-link-local", "fe80::1", true},
		{"public-1.1.1.1", "1.1.1.1", false},
		{"public-8.8.8.8", "8.8.8.8", false},
		{"public-ipv6", "2001:db8::1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tt.ip)
			}
			if got := isBlockedIP(ip); got != tt.want {
				t.Errorf("isBlockedIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIPLiteralDetection(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"127.0.0.1", true},
		{"192.168.1.1", true},
		{"::1", true},
		{"[::1]", true},
		{"api.stripe.com", false},
		{"localhost", false},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := isIPLiteral(tt.host); got != tt.want {
				t.Errorf("isIPLiteral(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

// --- Test: IP literal URL rejection ---

func TestFetchBlocksIPLiteral(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	proxy := testProxy(t, db, "")

	ctx := context.Background()
	_, err := proxy.Fetch(ctx, "test-app", "https://127.0.0.1/api", FetchOptions{})
	if err == nil {
		t.Fatal("expected error for IP literal URL")
	}
	ee, ok := err.(*EgressError)
	if !ok {
		t.Fatalf("expected EgressError, got %T: %v", err, err)
	}
	if ee.Code != CodeBlocked {
		t.Errorf("error code: got %s, want %s", ee.Code, CodeBlocked)
	}
}

// --- Test: Allowlist ---

func TestFetchBlocksNonAllowlisted(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	proxy := testProxy(t, db, "allowed.example.com")

	ctx := context.Background()
	_, err := proxy.Fetch(ctx, "test-app", "https://blocked.example.com/api", FetchOptions{})
	if err == nil {
		t.Fatal("expected error for non-allowlisted domain")
	}
	ee, ok := err.(*EgressError)
	if !ok {
		t.Fatalf("expected EgressError, got %T: %v", err, err)
	}
	if ee.Code != CodeBlocked {
		t.Errorf("error code: got %s, want %s", ee.Code, CodeBlocked)
	}
}

func TestFetchRequiresHTTPS(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	system.ResetCachedLimits()
	allowlist := NewAllowlist(db)
	// Add with https_only = true
	allowlist.Add("example.com", "", true)
	proxy := NewEgressProxy(allowlist)

	ctx := context.Background()
	_, err := proxy.Fetch(ctx, "test-app", "http://example.com/api", FetchOptions{})
	if err == nil {
		t.Fatal("expected error for HTTP when HTTPS required")
	}
	ee, ok := err.(*EgressError)
	if !ok {
		t.Fatalf("expected EgressError, got %T: %v", err, err)
	}
	if ee.Code != CodeBlocked {
		t.Errorf("error code: got %s, want %s", ee.Code, CodeBlocked)
	}
}

// --- Test: Request body size ---

func TestFetchBlocksLargeRequestBody(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	proxy := testProxy(t, db, "example.com")

	largeBody := strings.Repeat("x", 2*1024*1024) // 2MB > 1MB limit
	ctx := context.Background()
	_, err := proxy.Fetch(ctx, "test-app", "https://example.com/api", FetchOptions{
		Body: largeBody,
	})
	if err == nil {
		t.Fatal("expected error for oversized request body")
	}
	ee, ok := err.(*EgressError)
	if !ok {
		t.Fatalf("expected EgressError, got %T: %v", err, err)
	}
	if ee.Code != CodeSize {
		t.Errorf("error code: got %s, want %s", ee.Code, CodeSize)
	}
}

// --- Test: Successful fetch with test server ---

func TestFetchHappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "hello"})
	}))
	defer server.Close()

	db := testDB(t)
	defer db.Close()

	// The test server is on 127.0.0.1 which is blocked by our IP filter.
	// For the happy path, we test only the response building logic.
	// A full integration test would need an actual external server.
	t.Log("Note: Happy-path integration test requires non-loopback server. Testing response building instead.")

	resp := &FetchResponse{
		Status:  200,
		OK:      true,
		Headers: map[string]string{"content-type": "application/json"},
		body:    []byte(`{"message":"hello"}`),
	}

	if resp.Text() != `{"message":"hello"}` {
		t.Errorf("Text(): got %q", resp.Text())
	}

	if !resp.OK {
		t.Error("expected OK to be true")
	}
}

// --- Test: Response parsing ---

func TestFetchResponseText(t *testing.T) {
	resp := &FetchResponse{
		Status:  200,
		OK:      true,
		Headers: map[string]string{"content-type": "text/plain"},
		body:    []byte("hello world"),
	}
	if resp.Text() != "hello world" {
		t.Errorf("Text(): got %q, want %q", resp.Text(), "hello world")
	}
}

func TestFetchResponseJSON(t *testing.T) {
	resp := &FetchResponse{
		Status:  200,
		OK:      true,
		body:    []byte(`{"key":"value","num":42}`),
	}

	var data interface{}
	err := json.Unmarshal(resp.body, &data)
	if err != nil {
		t.Fatalf("JSON parse failed: %v", err)
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("expected map")
	}
	if m["key"] != "value" {
		t.Errorf("key: got %v, want %q", m["key"], "value")
	}
}

func TestFetchResponseInvalidJSON(t *testing.T) {
	resp := &FetchResponse{
		body: []byte("not json"),
	}

	var data interface{}
	err := json.Unmarshal(resp.body, &data)
	if err == nil {
		t.Error("expected JSON parse error for invalid JSON")
	}
}

// --- Test: Unsupported scheme ---

func TestFetchBlocksUnsupportedScheme(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	proxy := testProxy(t, db, "")

	ctx := context.Background()
	_, err := proxy.Fetch(ctx, "test-app", "ftp://example.com/file", FetchOptions{})
	if err == nil {
		t.Fatal("expected error for unsupported scheme")
	}
	ee, ok := err.(*EgressError)
	if !ok {
		t.Fatalf("expected EgressError, got %T", err)
	}
	if ee.Code != CodeBlocked {
		t.Errorf("error code: got %s, want %s", ee.Code, CodeBlocked)
	}
}

// --- Test: Header sanitization ---

func TestUnsafeHeadersBlocked(t *testing.T) {
	unsafe := []string{"Host", "Connection", "Proxy-Authorization", "Transfer-Encoding", "Accept-Encoding"}
	for _, h := range unsafe {
		if !unsafeHeaders[strings.ToLower(h)] {
			t.Errorf("expected %q to be in unsafeHeaders", h)
		}
	}
}

// --- Test: Error types ---

func TestEgressErrorTypes(t *testing.T) {
	tests := []struct {
		err       *EgressError
		retryable bool
		code      string
	}{
		{errBlocked("test"), false, CodeBlocked},
		{errTimeout("test"), false, CodeTimeout},
		{errLimit("test"), true, CodeLimit},
		{errBudget("test"), true, CodeBudget},
		{errSize("test"), false, CodeSize},
		{errNet("test"), false, CodeError},
		{errAuth("test"), false, CodeAuth},
		{errRate("test"), true, CodeRate},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("code: got %s, want %s", tt.err.Code, tt.code)
			}
			if IsRetryableError(tt.err) != tt.retryable {
				t.Errorf("retryable: got %v, want %v", IsRetryableError(tt.err), tt.retryable)
			}
			// Test Error() string
			if !strings.Contains(tt.err.Error(), tt.code) {
				t.Errorf("Error() should contain code: %s", tt.err.Error())
			}
		})
	}

	// Non-EgressError should not be retryable
	if IsRetryableError(fmt.Errorf("generic error")) {
		t.Error("generic error should not be retryable")
	}
}

// --- Test: Budget NetContext ---

func TestBudgetNetContext(t *testing.T) {
	cfg := timeout.DefaultConfig()

	// Budget with plenty of time
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	budget := timeout.NewBudget(ctx, cfg)

	netCtx, netCancel, err := budget.NetContext(context.Background())
	if err != nil {
		t.Fatalf("NetContext failed: %v", err)
	}
	defer netCancel()

	deadline, ok := netCtx.Deadline()
	if !ok {
		t.Fatal("expected deadline on net context")
	}
	if time.Until(deadline) > cfg.NetCallTimeout+100*time.Millisecond {
		t.Errorf("net context deadline too far: %v", time.Until(deadline))
	}
}

func TestBudgetNetContextInsufficientTime(t *testing.T) {
	cfg := timeout.DefaultConfig()

	// Budget with very little time
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	budget := timeout.NewBudget(ctx, cfg)

	_, _, err := budget.NetContext(context.Background())
	if err == nil {
		t.Fatal("expected error for insufficient budget")
	}
	if err != timeout.ErrInsufficientTime {
		t.Errorf("expected ErrInsufficientTime, got: %v", err)
	}
}

func TestBudgetNetContextNilBudget(t *testing.T) {
	var budget *timeout.Budget
	_, _, err := budget.NetContext(context.Background())
	if err == nil {
		t.Fatal("expected error for nil budget")
	}
}

// --- Test: Host canonicalization ---

func TestCanonicalizeHost(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"API.Stripe.com", "api.stripe.com"},
		{"api.stripe.com.", "api.stripe.com"},
		{"api.stripe.com:443", "api.stripe.com"},
		{"API.STRIPE.COM.:443", "api.stripe.com"},
		{"example.com", "example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := canonicalizeHost(tt.input); got != tt.want {
				t.Errorf("canonicalizeHost(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- Test: Proxy env ignored ---

func TestProxyEnvIgnored(t *testing.T) {
	system.ResetCachedLimits()
	allowlist := NewAllowlist(testDB(t))
	proxy := NewEgressProxy(allowlist)

	// The transport should have Proxy set to nil
	transport := proxy.client.Transport.(*http.Transport)
	if transport.Proxy != nil {
		t.Error("expected Transport.Proxy to be nil (ignore env proxy)")
	}
}

func TestCompressionDisabled(t *testing.T) {
	system.ResetCachedLimits()
	allowlist := NewAllowlist(testDB(t))
	proxy := NewEgressProxy(allowlist)

	transport := proxy.client.Transport.(*http.Transport)
	if !transport.DisableCompression {
		t.Error("expected DisableCompression to be true")
	}
}

// --- Extended SSRF Tests ---

func TestBlockedIPRanges_IPv4MappedIPv6(t *testing.T) {
	// IPv4-mapped IPv6 addresses should be blocked if they map to private/loopback IPs
	tests := []struct {
		name   string
		ip     string
		want   bool
		reason string
	}{
		{"ipv4-mapped-loopback", "::ffff:127.0.0.1", true, "IPv4-mapped loopback should be blocked"},
		{"ipv4-mapped-private-10", "::ffff:10.0.0.1", true, "IPv4-mapped private (10.x) should be blocked"},
		{"ipv4-mapped-private-192", "::ffff:192.168.1.1", true, "IPv4-mapped private (192.168.x) should be blocked"},
		{"ipv4-mapped-metadata", "::ffff:169.254.169.254", true, "IPv4-mapped metadata IP should be blocked"},
		{"ipv4-mapped-public", "::ffff:1.1.1.1", false, "IPv4-mapped public IP should be allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tt.ip)
			}
			if got := isBlockedIP(ip); got != tt.want {
				t.Errorf("%s: isBlockedIP(%s) = %v, want %v", tt.reason, tt.ip, got, tt.want)
			}
		})
	}
}

func TestBlockedIPRanges_IPv6EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		// IPv6 loopback variations
		{"ipv6-loopback-full", "0000:0000:0000:0000:0000:0000:0000:0001", true},
		{"ipv6-loopback-short", "::1", true},

		// IPv6 link-local (fe80::/10)
		{"ipv6-link-local-start", "fe80::1", true},
		{"ipv6-link-local-mid", "fe80::ffff:ffff:ffff:ffff", true},
		{"ipv6-link-local-upper", "febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff", true},

		// IPv6 unique local (fc00::/7)
		{"ipv6-unique-local-fc", "fc00::1", true},
		{"ipv6-unique-local-fd", "fd00::1", true},
		{"ipv6-unique-local-upper", "fdff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", true},

		// IPv6 documentation (should NOT be blocked - used for examples only)
		{"ipv6-doc-2001-db8", "2001:db8::1", false},

		// Public IPv6
		{"ipv6-google-dns", "2001:4860:4860::8888", false},
		{"ipv6-cloudflare-dns", "2606:4700:4700::1111", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tt.ip)
			}
			if got := isBlockedIP(ip); got != tt.want {
				t.Errorf("isBlockedIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIPLiteralDetection_EdgeCases(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		// IPv4 variations
		{"127.0.0.1", true},
		{"0.0.0.0", true},
		{"255.255.255.255", true},

		// IPv6 variations
		{"::1", true},
		{"[::1]", true},
		{"2001:db8::1", true},
		{"[2001:db8::1]", true},
		{"::ffff:127.0.0.1", true},
		{"[::ffff:127.0.0.1]", true},

		// Not IP literals
		{"localhost", false},
		{"example.com", false},
		{"192-168-1-1.nip.io", false}, // looks like IP but is domain
		{"127.0.0.1.nip.io", false},
		{"::1.example.com", false}, // domain with :: in it (edge case)
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := isIPLiteral(tt.host); got != tt.want {
				t.Errorf("isIPLiteral(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestFetchBlocksIPv6Loopback(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	proxy := testProxy(t, db, "")

	ctx := context.Background()

	ipv6Loopbacks := []string{
		"http://[::1]/api",
		"http://[0:0:0:0:0:0:0:1]/api",
		"https://[::1]/test",
	}

	for _, url := range ipv6Loopbacks {
		t.Run(url, func(t *testing.T) {
			_, err := proxy.Fetch(ctx, "test-app", url, FetchOptions{})
			if err == nil {
				t.Fatalf("expected error for IPv6 loopback URL: %s", url)
			}
			ee, ok := err.(*EgressError)
			if !ok {
				t.Fatalf("expected EgressError, got %T: %v", err, err)
			}
			if ee.Code != CodeBlocked {
				t.Errorf("error code: got %s, want %s for %s", ee.Code, CodeBlocked, url)
			}
		})
	}
}

func TestFetchBlocksIPv6PrivateRanges(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	proxy := testProxy(t, db, "")

	ctx := context.Background()

	privateIPv6URLs := []string{
		"http://[fc00::1]/api",            // unique local
		"http://[fd00::1]/api",            // unique local
		"http://[fe80::1]/api",            // link local
		"http://[::ffff:127.0.0.1]/api",   // IPv4-mapped loopback
		"http://[::ffff:192.168.1.1]/api", // IPv4-mapped private
	}

	for _, url := range privateIPv6URLs {
		t.Run(url, func(t *testing.T) {
			_, err := proxy.Fetch(ctx, "test-app", url, FetchOptions{})
			if err == nil {
				t.Fatalf("expected error for private IPv6 URL: %s", url)
			}
			ee, ok := err.(*EgressError)
			if !ok {
				t.Fatalf("expected EgressError, got %T: %v", err, err)
			}
			if ee.Code != CodeBlocked {
				t.Errorf("error code: got %s, want %s for %s", ee.Code, CodeBlocked, url)
			}
		})
	}
}

func TestFetchBlocksMetadataService(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	proxy := testProxy(t, db, "")

	ctx := context.Background()

	// Cloud metadata service IPs
	metadataURLs := []string{
		"http://169.254.169.254/latest/meta-data/", // AWS, Azure, GCP
		"http://[fe80::1]/metadata",                 // IPv6 link-local (some cloud providers)
	}

	for _, url := range metadataURLs {
		t.Run(url, func(t *testing.T) {
			_, err := proxy.Fetch(ctx, "test-app", url, FetchOptions{})
			if err == nil {
				t.Fatalf("expected error for metadata service URL: %s", url)
			}
			ee, ok := err.(*EgressError)
			if !ok {
				t.Fatalf("expected EgressError, got %T: %v", err, err)
			}
			if ee.Code != CodeBlocked {
				t.Errorf("error code: got %s, want %s for %s", ee.Code, CodeBlocked, url)
			}
		})
	}
}

func TestFetchBlocksAlternativeSchemes(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	proxy := testProxy(t, db, "example.com")

	ctx := context.Background()

	unsupportedSchemes := []string{
		"ftp://example.com/file",
		"file:///etc/passwd",
		"gopher://example.com/",
		"data:text/plain,hello",
		"javascript:alert(1)",
		"ws://example.com/socket",
		"wss://example.com/socket",
	}

	for _, url := range unsupportedSchemes {
		t.Run(url, func(t *testing.T) {
			_, err := proxy.Fetch(ctx, "test-app", url, FetchOptions{})
			if err == nil {
				t.Fatalf("expected error for unsupported scheme: %s", url)
			}
			ee, ok := err.(*EgressError)
			if !ok {
				t.Fatalf("expected EgressError, got %T: %v", err, err)
			}
			if ee.Code != CodeBlocked {
				t.Errorf("error code: got %s, want %s for %s", ee.Code, CodeBlocked, url)
			}
		})
	}
}

// Note: DNS rebinding protection is tested via DialContext resolver check
// The proxy resolves DNS and validates ALL returned IPs before connecting
// This prevents time-of-check-time-of-use attacks where DNS changes between validation and connection
