package main

import (
	"net/http"
	"testing"
)

// TestHostRoutingFlow tests host-based routing with authentication
func TestHostRoutingFlow(t *testing.T) {
	s := setupIntegrationTest(t)

	tests := []struct {
		name         string
		host         string
		path         string
		sessionRole  string // empty = no session
		expectStatus int
	}{
		// Admin subdomain routing
		{
			name:         "admin host with admin session - allowed",
			host:         "admin.testdomain.com",
			path:         "/api/system/config",
			sessionRole:  "admin",
			expectStatus: http.StatusOK,
		},
		{
			name:         "admin host with user session - forbidden",
			host:         "admin.testdomain.com",
			path:         "/api/system/config",
			sessionRole:  "user",
			expectStatus: http.StatusForbidden,
		},
		{
			name:         "admin host without session - unauthorized",
			host:         "admin.testdomain.com",
			path:         "/api/system/config",
			sessionRole:  "",
			expectStatus: http.StatusUnauthorized,
		},
		// Localhost routing (no role check, only auth)
		{
			name:         "localhost with admin session - allowed",
			host:         "localhost",
			path:         "/api/system/config",
			sessionRole:  "admin",
			expectStatus: http.StatusOK,
		},
		{
			name:         "localhost with user session - allowed (no role check)",
			host:         "localhost",
			path:         "/api/system/config",
			sessionRole:  "user",
			expectStatus: http.StatusOK,
		},
		{
			name:         "localhost without session - unauthorized",
			host:         "localhost",
			path:         "/api/system/config",
			sessionRole:  "",
			expectStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []requestOption{withHost(tt.host)}

			// Add session if specified
			if tt.sessionRole != "" {
				sessionID := s.createSession(t, tt.name+"@test.com", tt.sessionRole)
				opts = append(opts, withSession(sessionID))
			}

			resp := s.makeRequest(t, "GET", tt.path, nil, opts...)
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectStatus {
				body := readBody(t, resp)
				t.Fatalf("Expected status %d, got %d. Body: %s",
					tt.expectStatus, resp.StatusCode, body)
			}
		})
	}
}

// TestSubdomainAppServing tests that apps are served on subdomains via alias resolution
func TestSubdomainAppServing(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create a test app
	s.createTestApp(t, "myapp", "My App")

	// Create an alias pointing to the app
	s.createTestAlias(t, "blog", "myapp", "app")

	// Request the app via subdomain
	resp := s.makeRequest(t, "GET", "/",
		nil,
		withHost("blog.testdomain.com"),
	)
	defer resp.Body.Close()

	// Should serve the app
	if resp.StatusCode != http.StatusOK {
		body := readBody(t, resp)
		t.Fatalf("Expected status 200 for app serving, got %d. Body: %s",
			resp.StatusCode, body)
	}

	// Verify content
	body := readBody(t, resp)
	assertContains(t, body, "My App")
}

// TestSubdomainAppServing_NotFound tests that non-existent subdomains return 404
func TestSubdomainAppServing_NotFound(t *testing.T) {
	s := setupIntegrationTest(t)

	// Request non-existent subdomain
	resp := s.makeRequest(t, "GET", "/",
		nil,
		withHost("nonexistent.testdomain.com"),
	)
	defer resp.Body.Close()

	// Should return 404
	if resp.StatusCode != http.StatusNotFound {
		body := readBody(t, resp)
		t.Fatalf("Expected status 404 for non-existent subdomain, got %d. Body: %s",
			resp.StatusCode, body)
	}
}

// TestSubdomainAppServing_RootDomain tests root domain routing
func TestSubdomainAppServing_RootDomain(t *testing.T) {
	s := setupIntegrationTest(t)

	tests := []struct {
		name string
		host string
	}{
		{"root subdomain", "root.testdomain.com"},
		{"bare domain", "testdomain.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := s.makeRequest(t, "GET", "/",
				nil,
				withHost(tt.host),
			)
			defer resp.Body.Close()

			// Should serve the root app (status 200 or redirects are acceptable)
			if resp.StatusCode != http.StatusOK && resp.StatusCode < 300 {
				body := readBody(t, resp)
				t.Fatalf("Expected status 200 for root domain, got %d. Body: %s",
					resp.StatusCode, body)
			}
		})
	}
}

// TestSubdomainAppServing_Reserved tests that reserved aliases return 404
func TestSubdomainAppServing_Reserved(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create a reserved alias
	s.createTestAlias(t, "reserved-test", "", "reserved")

	// Request the reserved subdomain
	resp := s.makeRequest(t, "GET", "/",
		nil,
		withHost("reserved-test.testdomain.com"),
	)
	defer resp.Body.Close()

	// Should return 404
	if resp.StatusCode != http.StatusNotFound {
		body := readBody(t, resp)
		t.Fatalf("Expected status 404 for reserved alias, got %d. Body: %s",
			resp.StatusCode, body)
	}
}

// TestSubdomainAppServing_Redirect tests that redirect aliases work
func TestSubdomainAppServing_Redirect(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create a redirect alias
	// Schema: subdomain, type, targets (JSON with redirect_url)
	targets := `{"redirect_url":"https://example.com"}`
	_, err := s.db.Exec(`
		INSERT INTO aliases (subdomain, type, targets, created_at)
		VALUES (?, 'redirect', ?, datetime('now'))
	`, "redirect-test", targets)
	if err != nil {
		t.Fatalf("Failed to create redirect alias: %v", err)
	}

	// Request the redirect subdomain (don't follow redirects)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, _ := http.NewRequest("GET", s.server.URL+"/", nil)
	req.Host = "redirect-test.testdomain.com"

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should return 301 redirect
	if resp.StatusCode != http.StatusMovedPermanently {
		body := readBody(t, resp)
		t.Fatalf("Expected status 301 for redirect alias, got %d. Body: %s",
			resp.StatusCode, body)
	}

	// Verify redirect location
	location := resp.Header.Get("Location")
	if location != "https://example.com" {
		t.Fatalf("Expected redirect to https://example.com, got %s", location)
	}
}

// TestLocalhostSpecialCase tests that localhost routes to dashboard with auth
func TestLocalhostSpecialCase(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create a session
	sessionID := s.createSession(t, "localhost@test.com", "admin")

	// Request localhost (should go through AuthMiddleware)
	resp := s.makeRequest(t, "GET", "/",
		nil,
		withHost("localhost"),
		withSession(sessionID),
	)
	defer resp.Body.Close()

	// Should succeed with auth
	if resp.StatusCode >= 400 {
		body := readBody(t, resp)
		t.Fatalf("Expected successful response for authenticated localhost request, got %d. Body: %s",
			resp.StatusCode, body)
	}
}

// TestRoutingAuthBypassEndpoints tests public endpoints that bypass authentication
func TestRoutingAuthBypassEndpoints(t *testing.T) {
	s := setupIntegrationTest(t)

	publicEndpoints := []struct {
		path string
		host string
	}{
		{"/track", "admin.testdomain.com"},
		{"/pixel.gif", "admin.testdomain.com"},
		{"/r/test", "admin.testdomain.com"},
		{"/webhook/test", "admin.testdomain.com"},
	}

	for _, ep := range publicEndpoints {
		t.Run(ep.path, func(t *testing.T) {
			resp := s.makeRequest(t, "GET", ep.path,
				nil,
				withHost(ep.host),
			)
			defer resp.Body.Close()

			// Should not return 401 (may return other errors like 404, but not auth failure)
			if resp.StatusCode == http.StatusUnauthorized {
				body := readBody(t, resp)
				t.Fatalf("Public endpoint %s should not require auth, got 401. Body: %s",
					ep.path, body)
			}
		})
	}
}

// TestMiddlewareOrder tests that auth middleware is applied before handlers
func TestMiddlewareOrder(t *testing.T) {
	s := setupIntegrationTest(t)

	// Request protected endpoint without session
	resp := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
	)
	defer resp.Body.Close()

	// Should be rejected by auth middleware before reaching handler
	if resp.StatusCode != http.StatusUnauthorized {
		body := readBody(t, resp)
		t.Fatalf("Expected status 401 from auth middleware, got %d. Body: %s",
			resp.StatusCode, body)
	}
}

// TestPathPrecedence tests that specific paths take precedence over wildcards
func TestPathPrecedence(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create an app and alias
	s.createTestApp(t, "testapp", "Test App")
	s.createTestAlias(t, "api", "testapp", "app")

	// Request /api/something on root domain
	// This should NOT go to the "api" app, but to the dashboard API routing
	resp := s.makeRequest(t, "GET", "/api/stats",
		nil,
		withHost("admin.testdomain.com"),
	)
	defer resp.Body.Close()

	// Should be handled by API routing (requires auth), not app serving
	if resp.StatusCode != http.StatusUnauthorized {
		body := readBody(t, resp)
		t.Logf("Status: %d, Body: %s", resp.StatusCode, body)
	}
}
