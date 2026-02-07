package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"
)

// TestLoginToAPIAccess tests the full auth flow: Login → Session → Middleware → Handler
func TestLoginToAPIAccess(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create a test user with admin role
	sessionID := s.createSession(t, "admin@test.com", "admin")

	// Make authenticated request to admin-protected endpoint
	resp := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
		withSession(sessionID),
	)
	defer resp.Body.Close()

	// Should succeed with valid session
	if resp.StatusCode != http.StatusOK {
		body := readBody(t, resp)
		t.Fatalf("Expected status 200 for authenticated admin request, got %d. Body: %s",
			resp.StatusCode, body)
	}
}

// TestLoginToAPIAccess_InvalidSession tests that invalid sessions are rejected
func TestLoginToAPIAccess_InvalidSession(t *testing.T) {
	s := setupIntegrationTest(t)

	// Make request with invalid session ID
	resp := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
		withSession("invalid-session-id"),
	)
	defer resp.Body.Close()

	// Should be rejected
	if resp.StatusCode != http.StatusUnauthorized {
		body := readBody(t, resp)
		t.Fatalf("Expected status 401 for invalid session, got %d. Body: %s",
			resp.StatusCode, body)
	}
}

// TestLoginToAPIAccess_NoSession tests that missing sessions are rejected
func TestLoginToAPIAccess_NoSession(t *testing.T) {
	s := setupIntegrationTest(t)

	// Make request without session
	resp := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
	)
	defer resp.Body.Close()

	// Should be rejected
	if resp.StatusCode != http.StatusUnauthorized {
		body := readBody(t, resp)
		t.Fatalf("Expected status 401 for missing session, got %d. Body: %s",
			resp.StatusCode, body)
	}
}

// TestSessionExpiry tests that expired sessions are rejected
func TestSessionExpiry(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create user
	userID := "user-expiry-test"
	_, err := s.db.Exec(`
		INSERT INTO auth_users (id, email, name, picture, provider, role, created_at)
		VALUES (?, ?, ?, ?, 'test', 'user', ?)
	`, userID, "expiry@test.com", "Expiry Test", "", time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create expired session (expired 1 hour ago)
	expiredSessionToken := "test-session-expired"

	// Hash the token (same as createSession helper)
	h := sha256.Sum256([]byte(expiredSessionToken))
	tokenHash := hex.EncodeToString(h[:])

	now := time.Now().Unix()
	_, err = s.db.Exec(`
		INSERT INTO auth_sessions (token_hash, user_id, created_at, expires_at, last_seen)
		VALUES (?, ?, ?, ?, ?)
	`, tokenHash, userID, now-7200, now-3600, now-7200)
	if err != nil {
		t.Fatalf("Failed to create expired session: %v", err)
	}

	// Try to use expired session
	resp := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
		withSession(expiredSessionToken),
	)
	defer resp.Body.Close()

	// Should be rejected
	if resp.StatusCode != http.StatusUnauthorized {
		body := readBody(t, resp)
		t.Fatalf("Expected status 401 for expired session, got %d. Body: %s",
			resp.StatusCode, body)
	}
}

// TestRoleEscalation tests that regular users cannot access admin endpoints
func TestRoleEscalation(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create a test user with "user" role (not admin)
	sessionID := s.createSession(t, "user@test.com", "user")

	// Try to access admin-only endpoint on admin subdomain
	// Admin subdomain requires admin role for /api/* endpoints
	resp := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("admin.testdomain.com"),
		withSession(sessionID),
	)
	defer resp.Body.Close()

	// Should be forbidden (403) or unauthorized (401)
	if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusUnauthorized {
		body := readBody(t, resp)
		t.Fatalf("Expected status 403 or 401 for non-admin accessing admin endpoint, got %d. Body: %s",
			resp.StatusCode, body)
	}
}

// TestRoleEscalation_AdminAccess tests that admin users CAN access admin endpoints
func TestRoleEscalation_AdminAccess(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create a test user with "admin" role
	sessionID := s.createSession(t, "admin@test.com", "admin")

	// Access admin-only endpoint on admin subdomain
	resp := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("admin.testdomain.com"),
		withSession(sessionID),
	)
	defer resp.Body.Close()

	// Should succeed
	if resp.StatusCode != http.StatusOK {
		body := readBody(t, resp)
		t.Fatalf("Expected status 200 for admin accessing admin endpoint, got %d. Body: %s",
			resp.StatusCode, body)
	}
}

// TestAuthBypassEndpoints tests that certain endpoints bypass auth middleware
func TestAuthBypassEndpoints(t *testing.T) {
	s := setupIntegrationTest(t)

	// Test public endpoints that should work without auth
	publicEndpoints := []string{
		"/track",
		"/pixel.gif",
	}

	for _, endpoint := range publicEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			resp := s.makeRequest(t, "GET", endpoint,
				nil,
				withHost("admin.testdomain.com"),
			)
			defer resp.Body.Close()

			// Should not return 401 (may return 404 or other errors, but not auth failure)
			if resp.StatusCode == http.StatusUnauthorized {
				body := readBody(t, resp)
				t.Fatalf("Public endpoint %s should not require auth, got 401. Body: %s",
					endpoint, body)
			}
		})
	}
}

// TestSessionPersistence tests that sessions work across multiple requests
func TestSessionPersistence(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create a session
	sessionID := s.createSession(t, "persist@test.com", "admin")

	// Make multiple requests with same session
	for i := 0; i < 3; i++ {
		resp := s.makeRequest(t, "GET", "/api/system/config",
			nil,
			withHost("localhost"),
			withSession(sessionID),
		)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Request %d failed with status %d", i+1, resp.StatusCode)
		}
	}
}

// TestConcurrentSessions tests that multiple users can have concurrent sessions
func TestConcurrentSessions(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create multiple sessions
	session1 := s.createSession(t, "user1@test.com", "admin")
	session2 := s.createSession(t, "user2@test.com", "admin")
	session3 := s.createSession(t, "user3@test.com", "user")

	// Each session should work independently
	sessions := []struct {
		id           string
		expectStatus int
	}{
		{session1, http.StatusOK},
		{session2, http.StatusOK},
		{session3, http.StatusForbidden}, // user role can't access admin endpoint
	}

	for i, sess := range sessions {
		resp := s.makeRequest(t, "GET", "/api/system/config",
			nil,
			withHost("admin.testdomain.com"), // Use admin host for role-based access
			withSession(sess.id),
		)
		resp.Body.Close()

		if resp.StatusCode != sess.expectStatus {
			t.Fatalf("Session %d: expected status %d, got %d",
				i+1, sess.expectStatus, resp.StatusCode)
		}
	}
}

// TestAuthMe tests the /api/auth/me endpoint
func TestAuthMe(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create a session
	sessionID := s.createSession(t, "me@test.com", "admin")

	// Request user info
	resp := s.makeRequest(t, "GET", "/api/auth/me",
		nil,
		withHost("localhost"),
		withSession(sessionID),
	)
	defer resp.Body.Close()

	// Should succeed
	if resp.StatusCode != http.StatusOK {
		body := readBody(t, resp)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, body)
	}

	// Parse response
	var result map[string]interface{}
	parseJSON(t, resp, &result)

	// Verify user info
	if result["email"] != "me@test.com" {
		t.Fatalf("Expected email 'me@test.com', got %v", result["email"])
	}

	if result["role"] != "admin" {
		t.Fatalf("Expected role 'admin', got %v", result["role"])
	}
}

// TestAuthMe_Unauthorized tests that /api/auth/me requires authentication
func TestAuthMe_Unauthorized(t *testing.T) {
	s := setupIntegrationTest(t)

	// Request without session
	resp := s.makeRequest(t, "GET", "/api/auth/me",
		nil,
		withHost("localhost"),
	)
	defer resp.Body.Close()

	// Should be unauthorized
	if resp.StatusCode != http.StatusUnauthorized {
		body := readBody(t, resp)
		t.Fatalf("Expected status 401, got %d. Body: %s", resp.StatusCode, body)
	}
}
