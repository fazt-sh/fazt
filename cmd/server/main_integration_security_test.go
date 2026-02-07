package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"testing"
	"time"
)

// --- Auth Bypass Tests ---

func TestAuthBypass_InvalidSessionToken(t *testing.T) {
	s := setupIntegrationTest(t)

	invalidTokens := []string{
		"",                              // empty
		"invalid-token",                 // not in database
		"x",                             // too short
		strings.Repeat("a", 1000),       // too long
		"../../../etc/passwd",           // path traversal
		"<script>alert('xss')</script>", // XSS
		"'; DROP TABLE auth_sessions--", // SQL injection
	}

	for _, token := range invalidTokens {
		t.Run("token="+token, func(t *testing.T) {
			resp := s.makeRequest(t, "GET", "/api/system/config",
				nil,
				withHost("localhost"),
				withSession(token),
			)
			defer resp.Body.Close()

			// Should be rejected as unauthorized
			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected 401 for invalid token, got %d", resp.StatusCode)
			}
		})
	}
}

func TestAuthBypass_ExpiredSession(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create user
	userID := "user-expired"
	_, err := s.db.Exec(`
		INSERT INTO auth_users (id, email, name, picture, provider, role, created_at)
		VALUES (?, ?, ?, ?, 'test', 'user', ?)
	`, userID, "expired@test.com", "Expired User", "", time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create expired session (expired 1 hour ago)
	expiredToken := "test-session-expired"
	h := sha256.Sum256([]byte(expiredToken))
	tokenHash := hex.EncodeToString(h[:])

	now := time.Now().Unix()
	_, err = s.db.Exec(`
		INSERT INTO auth_sessions (token_hash, user_id, created_at, expires_at, last_seen)
		VALUES (?, ?, ?, ?, ?)
	`, tokenHash, userID, now-7200, now-3600, now-7200)
	if err != nil {
		t.Fatalf("Failed to create expired session: %v", err)
	}

	// Try to access protected endpoint with expired session
	resp := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
		withSession(expiredToken),
	)
	defer resp.Body.Close()

	// Should be rejected
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 for expired session, got %d", resp.StatusCode)
	}
}

func TestAuthBypass_TamperedSessionToken(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create valid session
	sessionToken := s.createSession(t, "tamper@test.com", "admin")

	// Tamper with the token
	tamperedTokens := []string{
		sessionToken + "x",                      // appended char
		"x" + sessionToken,                      // prepended char
		sessionToken[:len(sessionToken)-1],      // truncated
		strings.ToUpper(sessionToken),           // case changed
		strings.ReplaceAll(sessionToken, "-", "_"), // char substitution
	}

	for _, tampered := range tamperedTokens {
		t.Run("tampered="+tampered[:20], func(t *testing.T) {
			resp := s.makeRequest(t, "GET", "/api/system/config",
				nil,
				withHost("localhost"),
				withSession(tampered),
			)
			defer resp.Body.Close()

			// Should be rejected
			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected 401 for tampered token, got %d", resp.StatusCode)
			}
		})
	}
}

func TestAuthBypass_TokenReplayAfterLogout(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create session
	sessionToken := s.createSession(t, "replay@test.com", "admin")

	// Verify session works
	resp1 := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
		withSession(sessionToken),
	)
	resp1.Body.Close()
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("Initial auth failed: %d", resp1.StatusCode)
	}

	// Simulate logout by deleting session from database
	h := sha256.Sum256([]byte(sessionToken))
	tokenHash := hex.EncodeToString(h[:])
	_, err := s.db.Exec("DELETE FROM auth_sessions WHERE token_hash = ?", tokenHash)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Try to replay the token
	resp2 := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
		withSession(sessionToken),
	)
	defer resp2.Body.Close()

	// Should be rejected (token no longer valid)
	if resp2.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 for replayed token after logout, got %d", resp2.StatusCode)
	}
}

func TestAuthBypass_MissingCookie(t *testing.T) {
	s := setupIntegrationTest(t)

	// Try to access protected endpoint without cookie
	resp := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
		// No session cookie
	)
	defer resp.Body.Close()

	// Should be rejected
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 for missing cookie, got %d", resp.StatusCode)
	}
}

// TestAuthBypass_SessionWithoutUser is not possible due to foreign key constraints
// The database enforces referential integrity between auth_sessions and auth_users
// This is a good security property - orphan sessions cannot exist

func TestAuthBypass_ConcurrentSessionInvalidation(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create session
	sessionToken := s.createSession(t, "concurrent@test.com", "admin")

	// First request succeeds
	resp1 := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
		withSession(sessionToken),
	)
	resp1.Body.Close()
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("Initial auth failed: %d", resp1.StatusCode)
	}

	// Invalidate session (e.g., admin force-logout)
	h := sha256.Sum256([]byte(sessionToken))
	tokenHash := hex.EncodeToString(h[:])
	_, err := s.db.Exec("DELETE FROM auth_sessions WHERE token_hash = ?", tokenHash)
	if err != nil {
		t.Fatalf("Failed to invalidate session: %v", err)
	}

	// Second request (concurrent with invalidation) should fail
	resp2 := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
		withSession(sessionToken),
	)
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 after session invalidation, got %d", resp2.StatusCode)
	}
}

func TestAuthBypass_SessionHijacking_IPChange(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create session
	sessionToken := s.createSession(t, "hijack@test.com", "admin")

	// Request from "original" IP
	resp1 := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
		withSession(sessionToken),
	)
	resp1.Body.Close()
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("Initial auth failed: %d", resp1.StatusCode)
	}

	// Note: Current implementation doesn't bind sessions to IPs
	// This test documents the current behavior - sessions are portable across IPs
	// If IP binding is added in future, this test should be updated to expect 401

	// Request from "different" IP (simulated via different X-Forwarded-For header)
	resp2 := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
		withSession(sessionToken),
		func(r *http.Request) {
			r.Header.Set("X-Forwarded-For", "1.2.3.4")
		},
	)
	defer resp2.Body.Close()

	// Currently allows (no IP binding)
	if resp2.StatusCode != http.StatusOK {
		t.Logf("Session rejected from different IP: %d (IP binding may be implemented)", resp2.StatusCode)
	}
}

func TestAuthBypass_MultipleInvalidAttempts(t *testing.T) {
	s := setupIntegrationTest(t)

	// Make multiple requests with invalid tokens
	for i := 0; i < 10; i++ {
		resp := s.makeRequest(t, "GET", "/api/system/config",
			nil,
			withHost("localhost"),
			withSession("invalid-token-"+string(rune(i))),
		)
		resp.Body.Close()

		// All should be rejected consistently
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Attempt %d: Expected 401, got %d", i+1, resp.StatusCode)
		}
	}

	// Note: Current implementation doesn't have rate limiting on auth checks
	// If rate limiting is added, this test should be updated
}

func TestAuthBypass_SessionCreatedInFuture(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create user
	userID := "user-future"
	_, err := s.db.Exec(`
		INSERT INTO auth_users (id, email, name, picture, provider, role, created_at)
		VALUES (?, ?, ?, ?, 'test', 'user', ?)
	`, userID, "future@test.com", "Future User", "", time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create session with future timestamps (potential time manipulation)
	futureToken := "test-session-future"
	h := sha256.Sum256([]byte(futureToken))
	tokenHash := hex.EncodeToString(h[:])

	future := time.Now().Add(24 * time.Hour).Unix()
	_, err = s.db.Exec(`
		INSERT INTO auth_sessions (token_hash, user_id, created_at, expires_at, last_seen)
		VALUES (?, ?, ?, ?, ?)
	`, tokenHash, userID, future, future+86400, future)
	if err != nil {
		t.Fatalf("Failed to create future session: %v", err)
	}

	// Try to use future session
	resp := s.makeRequest(t, "GET", "/api/system/config",
		nil,
		withHost("localhost"),
		withSession(futureToken),
	)
	defer resp.Body.Close()

	// Current implementation checks expires_at > now, not created_at
	// Future-dated sessions are currently accepted
	if resp.StatusCode != http.StatusOK {
		t.Logf("Future session rejected: %d (timestamp validation may be implemented)", resp.StatusCode)
	}
}

func TestAuthBypass_SQLInjectionInSessionToken(t *testing.T) {
	s := setupIntegrationTest(t)

	sqlInjectionTokens := []string{
		"' OR '1'='1",
		"'; DROP TABLE auth_sessions--",
		"' UNION SELECT * FROM auth_users--",
		"admin'--",
	}

	for _, token := range sqlInjectionTokens {
		t.Run("sql_injection="+token, func(t *testing.T) {
			resp := s.makeRequest(t, "GET", "/api/system/config",
				nil,
				withHost("localhost"),
				withSession(token),
			)
			defer resp.Body.Close()

			// Should NOT cause 500 (SQL injection prevented by parameterized queries)
			if resp.StatusCode == http.StatusInternalServerError {
				body := readBody(t, resp)
				t.Errorf("SQL injection in session token caused 500: %s", body)
			}

			// Should be rejected as invalid
			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected 401 for SQL injection token, got %d", resp.StatusCode)
			}
		})
	}
}

// Note: Rate limiting tests are commented out pending implementation
// Uncomment when rate limiting is added to login/auth endpoints

/*
func TestAuthBypass_BruteForceRateLimit(t *testing.T) {
	s := setupIntegrationTest(t)

	// Make many rapid login attempts
	for i := 0; i < 100; i++ {
		resp := s.makeRequest(t, "POST", "/api/login",
			strings.NewReader(`{"email":"brute@test.com","password":"wrong"}`),
			withHost("localhost"),
			withJSON(),
		)
		resp.Body.Close()

		// After some threshold (e.g., 10 attempts), should be rate limited
		if i > 10 && resp.StatusCode == http.StatusTooManyRequests {
			t.Logf("Rate limit triggered after %d attempts", i+1)
			return
		}
	}

	t.Logf("WARNING: No rate limiting detected on login endpoint")
}
*/
