package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/handlers"
	"golang.org/x/crypto/bcrypt"
)

// setupAdversarialTest extends setupIntegrationTest with real bcrypt admin password
// and initialized rate limiter — needed by timing, rate limit, and cookie tests.
func setupAdversarialTest(t *testing.T) *integrationTestServer {
	t.Helper()

	s := setupIntegrationTest(t)

	// Set a real bcrypt password in both the configurations table and config.Get()
	password := "testpassword123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Insert admin credentials into configurations table (used by VerifyAdminCredentials)
	s.db.Exec(`INSERT OR REPLACE INTO configurations (key, value) VALUES ('auth.username', 'admin')`)
	s.db.Exec(`INSERT OR REPLACE INTO configurations (key, value) VALUES ('auth.password_hash', ?)`, string(hash))

	// Update config.Get() so LoginHandler can verify credentials
	cfg := config.Get()
	cfg.Auth.Username = "admin"
	cfg.Auth.PasswordHash = string(hash)
	config.SetConfig(cfg)

	// Initialize rate limiter for LoginHandler
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	handlers.InitAuth(s.authService, limiter, "test")

	return s
}

// --- Category: Cookie & Session Attacks ---

// TestAdversarial_SessionFixationViaCookieInjection simulates a subdomain cookie
// injection attack. An attacker on evil.domain.com sets a fazt_session cookie scoped
// to .domain.com. When the victim visits admin.domain.com, the browser sends both
// cookies. Go's r.Cookie() returns the first match — the attacker's invalid token.
//
// Result: auth denial-of-service (victim gets 401), NOT privilege escalation.
func TestAdversarial_SessionFixationViaCookieInjection(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create a legitimate session
	validToken := s.createSession(t, "victim@test.com", "admin")

	// Attacker's invalid token placed FIRST in the cookie header
	// Go's http.Request.Cookie() returns the first match by name
	attackerToken := "attacker-planted-invalid-token"

	req, err := http.NewRequest("GET", s.server.URL+"/api/auth/me", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Host = "admin.testdomain.com"

	// Simulate browser sending two cookies with same name:
	// Attacker's cookie (from .domain.com scope) comes FIRST
	req.Header.Set("Cookie",
		fmt.Sprintf("fazt_session=%s; fazt_session=%s", attackerToken, validToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Go's r.Cookie() returns the FIRST match — attacker's invalid token
	// This means the valid session is shadowed → 401
	if resp.StatusCode == http.StatusOK {
		t.Error("Cookie injection should NOT grant access — attacker's token evaluated first should fail")
	}

	t.Logf("Cookie injection result: status=%d (attacker causes auth DoS, not privilege escalation)", resp.StatusCode)

	// Verify the legitimate token still works when sent alone
	resp2 := s.makeRequest(t, "GET", "/api/auth/me", nil,
		withHost("admin.testdomain.com"),
		withSession(validToken),
	)
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Valid token alone should work, got %d", resp2.StatusCode)
	}
}

// TestAdversarial_CookieAttributeVerification performs a real login and inspects
// the Set-Cookie header for security attributes. This is a regression guard.
func TestAdversarial_CookieAttributeVerification(t *testing.T) {
	s := setupAdversarialTest(t)

	// Perform a real login via localhost (admin.* routes through AdminMiddleware
	// which blocks unauthenticated /api/login)
	body := strings.NewReader(`{"username":"admin","password":"testpassword123"}`)
	resp := s.makeRequest(t, "POST", "/api/login", body,
		withHost("localhost"),
		withJSON(),
	)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed with status %d, body: %s", resp.StatusCode, readBody(t, resp))
	}

	// Find the session cookie in Set-Cookie headers
	var sessionCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "fazt_session" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("No fazt_session cookie in login response")
	}

	// Check required security attributes
	if !sessionCookie.HttpOnly {
		t.Error("Cookie MUST have HttpOnly flag to prevent XSS theft")
	}

	if sessionCookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("Cookie SameSite should be Lax, got %v", sessionCookie.SameSite)
	}

	if sessionCookie.Path != "/" {
		t.Errorf("Cookie Path should be '/', got %q", sessionCookie.Path)
	}

	if sessionCookie.Value == "" {
		t.Error("Cookie value must not be empty")
	}

	if sessionCookie.MaxAge <= 0 {
		t.Errorf("Cookie MaxAge should be positive, got %d", sessionCookie.MaxAge)
	}

	// Note: Secure flag is false in test mode (not HTTPS) — expected behavior
	// In production (secure=true), this would be set
	t.Logf("Cookie attributes verified: HttpOnly=%v, SameSite=%v, Path=%s, MaxAge=%d, Secure=%v",
		sessionCookie.HttpOnly, sessionCookie.SameSite, sessionCookie.Path,
		sessionCookie.MaxAge, sessionCookie.Secure)
}

// --- Category: Race Conditions ---

// TestAdversarial_InviteCodeDoubleSpend races 10 goroutines to redeem a single-use
// invite code. RedeemInvite has a TOCTOU gap: GetInvite → IsValid → CreateUser →
// UPDATE use_count (no wrapping transaction). Even with MaxOpenConns(1), Go's
// database/sql releases the connection between individual statements, allowing
// interleaving: goroutine A does SELECT (sees use_count=0) → goroutine B does
// SELECT (also sees use_count=0) → both proceed to CreateUser and UPDATE.
// This documents the vulnerability.
func TestAdversarial_InviteCodeDoubleSpend(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create admin user who creates the invite
	adminID := "user-inviter@test.com"
	s.db.Exec(`
		INSERT INTO auth_users (id, email, name, picture, provider, role, created_at)
		VALUES (?, ?, ?, '', 'test', 'admin', ?)
	`, adminID, "inviter@test.com", "Inviter", time.Now().Unix())

	// Create a single-use invite
	invite, err := s.authService.CreateInvite("user", adminID, 1, nil)
	if err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}

	const racers = 10
	barrier := make(chan struct{})
	var successes int32
	var failures int32
	var wg sync.WaitGroup

	for i := 0; i < racers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-barrier // Wait for barrier release

			email := fmt.Sprintf("racer%d@test.com", idx)
			_, err := s.authService.RedeemInvite(invite.Code, email, "Racer", "password1234")
			if err == nil {
				atomic.AddInt32(&successes, 1)
			} else {
				atomic.AddInt32(&failures, 1)
			}
		}(i)
	}

	// Release all goroutines simultaneously
	close(barrier)
	wg.Wait()

	successCount := atomic.LoadInt32(&successes)
	failureCount := atomic.LoadInt32(&failures)

	t.Logf("Invite double-spend race: %d successes, %d failures (out of %d racers)",
		successCount, failureCount, racers)

	// TOCTOU vulnerability: RedeemInvite reads use_count, checks IsValid, creates
	// user, then increments use_count — all without a transaction. Multiple goroutines
	// can read use_count=0 before any writes, allowing multiple redemptions.
	// With a proper transaction (BEGIN → SELECT FOR UPDATE → ... → COMMIT), only 1
	// would succeed. Document the actual behavior.
	if successCount > 1 {
		t.Logf("TOCTOU CONFIRMED: %d goroutines redeemed a single-use invite (expected 1)", successCount)
		t.Logf("Fix: wrap GetInvite → IsValid → CreateUser → UPDATE in a transaction")
	}

	// Verify invite use_count
	redeemed, err := s.authService.GetInvite(invite.Code)
	if err != nil {
		t.Fatalf("Failed to get invite: %v", err)
	}

	t.Logf("Invite use_count after race: %d (max_uses: %d)", redeemed.UseCount, redeemed.MaxUses)
}

// TestAdversarial_OAuthStateRaceCondition races 5 goroutines to validate the same
// OAuth state token simultaneously. ValidateState does SELECT → DELETE → check expiry
// as separate statements. Go's database/sql releases the connection between calls,
// allowing interleaving: all goroutines SELECT (find the row) before any DELETE runs.
// This documents the TOCTOU in state validation.
func TestAdversarial_OAuthStateRaceCondition(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create an OAuth state token
	state, err := s.authService.CreateState("google", "/dashboard", "")
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	const racers = 5
	barrier := make(chan struct{})
	var successes int32
	var failures int32
	var wg sync.WaitGroup

	for i := 0; i < racers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-barrier

			_, err := s.authService.ValidateState(state)
			if err == nil {
				atomic.AddInt32(&successes, 1)
			} else {
				atomic.AddInt32(&failures, 1)
			}
		}()
	}

	close(barrier)
	wg.Wait()

	successCount := atomic.LoadInt32(&successes)
	failureCount := atomic.LoadInt32(&failures)

	t.Logf("OAuth state race: %d successes, %d failures", successCount, failureCount)

	// ValidateState: SELECT (finds row) → DELETE (removes it) → return success
	// Without a transaction, multiple goroutines can SELECT before any DELETE executes.
	// With a proper atomic DELETE...RETURNING, only 1 would succeed.
	if successCount > 1 {
		t.Logf("TOCTOU CONFIRMED: %d goroutines validated same state token (expected 1)", successCount)
		t.Logf("Fix: use DELETE...RETURNING or wrap in a transaction")
	} else if successCount == 1 {
		t.Logf("State validation was serialized correctly in this run")
	}

	// At least one must succeed
	if successCount == 0 {
		t.Error("No goroutine successfully validated the state token")
	}
}

// TestAdversarial_ConcurrentAuthAndInvalidation hammers /api/auth/me with 5
// goroutines while the session is deleted mid-flight. Verifies: zero 500s,
// clean transition from 200→401, no panics.
func TestAdversarial_ConcurrentAuthAndInvalidation(t *testing.T) {
	s := setupIntegrationTest(t)

	token := s.createSession(t, "concurrent@test.com", "admin")

	const workers = 5
	const requestsPerWorker = 20
	barrier := make(chan struct{})
	var status200 int32
	var status401 int32
	var status500 int32
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-barrier

			for j := 0; j < requestsPerWorker; j++ {
				resp := s.makeRequest(t, "GET", "/api/auth/me", nil,
					withHost("admin.testdomain.com"),
					withSession(token),
				)
				resp.Body.Close()

				switch resp.StatusCode {
				case 200:
					atomic.AddInt32(&status200, 1)
				case 401:
					atomic.AddInt32(&status401, 1)
				default:
					atomic.AddInt32(&status500, 1)
				}
			}
		}()
	}

	// Start all workers
	close(barrier)

	// Delete the session mid-flight after a brief delay
	time.Sleep(2 * time.Millisecond)
	s.authService.DeleteSession(token)

	wg.Wait()

	total := atomic.LoadInt32(&status200) + atomic.LoadInt32(&status401) + atomic.LoadInt32(&status500)

	t.Logf("Concurrent auth race: 200=%d, 401=%d, 5xx=%d (total=%d)",
		atomic.LoadInt32(&status200), atomic.LoadInt32(&status401),
		atomic.LoadInt32(&status500), total)

	// Zero 500s — the server must handle concurrent session invalidation gracefully
	if s500 := atomic.LoadInt32(&status500); s500 > 0 {
		t.Errorf("Got %d 500 responses during concurrent auth/invalidation — server should handle this gracefully", s500)
	}

	// Should see at least some 401s after deletion
	if s401 := atomic.LoadInt32(&status401); s401 == 0 {
		t.Logf("Warning: no 401s observed — session deletion may not have overlapped with requests")
	}
}

// --- Category: Timing & Side Channels ---

// TestAdversarial_LoginTimingSideChannel measures the timing difference between
// invalid username (fast return ~1ms) and valid username + wrong password (bcrypt
// ~80ms+). LoginHandler returns immediately for wrong username without calling bcrypt.
// This documents a username enumeration side channel.
func TestAdversarial_LoginTimingSideChannel(t *testing.T) {
	s := setupAdversarialTest(t)

	const iterations = 3

	// Use localhost for login (admin.* routes through AdminMiddleware)

	// Measure: invalid username (fast path — no bcrypt)
	var invalidUserTotal time.Duration
	for i := 0; i < iterations; i++ {
		body := strings.NewReader(`{"username":"nonexistent","password":"wrong"}`)
		start := time.Now()
		resp := s.makeRequest(t, "POST", "/api/login", body,
			withHost("localhost"),
			withJSON(),
		)
		invalidUserTotal += time.Since(start)
		resp.Body.Close()
	}
	avgInvalidUser := invalidUserTotal / time.Duration(iterations)

	// Measure: valid username + wrong password (slow path — bcrypt verify)
	var validUserTotal time.Duration
	for i := 0; i < iterations; i++ {
		body := strings.NewReader(`{"username":"admin","password":"wrongpassword"}`)
		start := time.Now()
		resp := s.makeRequest(t, "POST", "/api/login", body,
			withHost("localhost"),
			withJSON(),
		)
		validUserTotal += time.Since(start)
		resp.Body.Close()
	}
	avgValidUser := validUserTotal / time.Duration(iterations)

	t.Logf("Timing side channel: invalid_user=%v, valid_user+wrong_pw=%v",
		avgInvalidUser, avgValidUser)

	// Document the side channel — valid username triggers bcrypt, which is measurably slower
	if avgValidUser > avgInvalidUser*2 {
		t.Logf("SIDE CHANNEL CONFIRMED: valid username is %dx slower than invalid username",
			avgValidUser/max(avgInvalidUser, 1))
		t.Logf("Attacker can enumerate valid usernames by measuring response latency")
	} else {
		t.Logf("Timing difference not significant in this run (may vary by CPU load)")
	}
}

// TestAdversarial_RateLimitBypassViaIPSpoofing demonstrates that getClientIP()
// trusts X-Forwarded-For unconditionally. An attacker can exhaust the rate limit
// for their real IP, then bypass it by spoofing different IPs.
func TestAdversarial_RateLimitBypassViaIPSpoofing(t *testing.T) {
	s := setupAdversarialTest(t)

	// Use localhost for login requests (admin.* routes through AdminMiddleware
	// which blocks unauthenticated /api/login)

	// Exhaust rate limit for real IP (5 failed attempts)
	for i := 0; i < 5; i++ {
		body := strings.NewReader(`{"username":"admin","password":"wrong"}`)
		resp := s.makeRequest(t, "POST", "/api/login", body,
			withHost("localhost"),
			withJSON(),
		)
		resp.Body.Close()
	}

	// 6th attempt from real IP should be rate-limited
	body := strings.NewReader(`{"username":"admin","password":"wrong"}`)
	resp := s.makeRequest(t, "POST", "/api/login", body,
		withHost("localhost"),
		withJSON(),
	)
	resp.Body.Close()

	rateLimited := resp.StatusCode == http.StatusTooManyRequests
	t.Logf("After 5 failures, 6th attempt: status=%d (rate_limited=%v)", resp.StatusCode, rateLimited)

	if !rateLimited {
		t.Error("Expected rate limiting after 5 failed attempts")
	}

	// Now spoof different IPs via X-Forwarded-For to bypass
	var bypassed int32
	for i := 0; i < 20; i++ {
		spoofedIP := fmt.Sprintf("10.0.%d.%d", i/256, i%256)
		body := strings.NewReader(`{"username":"admin","password":"wrong"}`)

		req, err := http.NewRequest("POST", s.server.URL+"/api/login", body)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Host = "localhost"
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", spoofedIP)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusTooManyRequests {
			atomic.AddInt32(&bypassed, 1)
		}
	}

	bypassCount := atomic.LoadInt32(&bypassed)
	t.Logf("IP spoofing bypass: %d/20 requests bypassed rate limit via X-Forwarded-For", bypassCount)

	if bypassCount > 0 {
		t.Logf("VULNERABILITY: getClientIP() trusts X-Forwarded-For unconditionally")
		t.Logf("Attacker can send unlimited login attempts by rotating spoofed IPs")
	}
}

// --- Category: Middleware & Routing Bypass ---

// TestAdversarial_HostHeaderAdminBypass attempts to access admin routes using
// variations of the admin hostname. createRootHandler uses exact match
// `host == "admin."+mainDomain` without lowercasing.
func TestAdversarial_HostHeaderAdminBypass(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create a non-admin user
	nonAdminToken := s.createSession(t, "regular@test.com", "user")

	hostVariations := []struct {
		name string
		host string
	}{
		{"uppercase", "ADMIN.testdomain.com"},
		{"mixed_case", "Admin.testdomain.com"},
		{"suffix_attack", "admin.testdomain.com.evil.com"},
		{"prefix_attack", "xadmin.testdomain.com"},
		{"trailing_dot", "admin.testdomain.com."},
		{"double_dot", "admin..testdomain.com"},
		{"null_byte", "admin.testdomain.com\x00evil"},
		{"unicode_confusable", "аdmin.testdomain.com"}, // Cyrillic 'а'
	}

	for _, tc := range hostVariations {
		t.Run(tc.name, func(t *testing.T) {
			resp := s.makeRequest(t, "GET", "/api/system/config", nil,
				withHost(tc.host),
				withSession(nonAdminToken),
			)
			defer resp.Body.Close()

			// None of these should return 200 OK for a non-admin user
			if resp.StatusCode == http.StatusOK {
				t.Errorf("Host %q allowed access to admin API with non-admin session (status=%d)",
					tc.host, resp.StatusCode)
			}

			t.Logf("Host %q → status=%d", tc.name, resp.StatusCode)
		})
	}
}

// TestAdversarial_PathConfusionAuthMiddleware attempts path traversal and case
// manipulation to bypass the auth middleware's prefix-based path matching.
// Go's http.ServeMux cleans paths before routing, preventing these attacks.
func TestAdversarial_PathConfusionAuthMiddleware(t *testing.T) {
	s := setupIntegrationTest(t)

	pathAttacks := []struct {
		name string
		path string
	}{
		{"dot_dot_traversal", "/auth/../api/system/config"},
		{"static_traversal", "/static/../api/system/config"},
		{"login_suffix", "/api/login/../system/config"},
		{"case_static", "/STATIC/../../api/system/config"},
		{"case_auth", "/Auth/../api/system/config"},
		{"case_api", "/API/system/config"},
		{"double_slash", "//api/system/config"},
		{"encoded_slash", "/api%2fsystem%2fconfig"},
		{"null_in_path", "/api/system/config%00.html"},
		{"semicolon", "/api/system/config;.html"},
	}

	for _, tc := range pathAttacks {
		t.Run(tc.name, func(t *testing.T) {
			// Use raw request to prevent Go client from normalizing
			req, err := http.NewRequest("GET", s.server.URL+tc.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Host = "admin.testdomain.com"
			// No session — should be unauthorized

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			// Should never get 200 without authentication
			if resp.StatusCode == http.StatusOK {
				t.Errorf("Path %q bypassed auth and returned 200", tc.path)
			}

			t.Logf("Path %q → status=%d", tc.name, resp.StatusCode)
		})
	}
}

// --- Category: Session Integrity ---

// TestAdversarial_DeletedUserOrphanedSession verifies that deleting a user
// cascades to their sessions via FK ON DELETE CASCADE. An orphaned session
// must not grant access.
func TestAdversarial_DeletedUserOrphanedSession(t *testing.T) {
	s := setupIntegrationTest(t)

	token := s.createSession(t, "doomed@test.com", "admin")
	userID := "user-doomed@test.com"

	// Verify session works before deletion
	resp := s.makeRequest(t, "GET", "/api/auth/me", nil,
		withHost("admin.testdomain.com"),
		withSession(token),
	)
	assertStatus(t, resp, http.StatusOK)

	// Delete the user directly from DB (simulating admin action)
	_, err := s.db.Exec(`DELETE FROM auth_users WHERE id = ?`, userID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify FK CASCADE removed the session
	var sessionCount int
	h := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(h[:])
	s.db.QueryRow(`SELECT COUNT(*) FROM auth_sessions WHERE token_hash = ?`, tokenHash).Scan(&sessionCount)

	if sessionCount > 0 {
		t.Errorf("Session still exists after user deletion (FK CASCADE not working): count=%d", sessionCount)
	} else {
		t.Log("FK CASCADE confirmed: session deleted with user")
	}

	// Verify HTTP returns 401
	resp2 := s.makeRequest(t, "GET", "/api/auth/me", nil,
		withHost("admin.testdomain.com"),
		withSession(token),
	)
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusUnauthorized {
		t.Errorf("Deleted user's session should return 401, got %d", resp2.StatusCode)
	}
}

// TestAdversarial_SessionAfterRoleDowngrade verifies that downgrading a user's
// role is immediately reflected in subsequent requests. ValidateSession calls
// GetUserByID on every request, which fetches the CURRENT role — no stale cache.
func TestAdversarial_SessionAfterRoleDowngrade(t *testing.T) {
	s := setupIntegrationTest(t)

	token := s.createSession(t, "demoted@test.com", "admin")
	userID := "user-demoted@test.com"

	// Admin session works on admin subdomain
	resp := s.makeRequest(t, "GET", "/api/system/config", nil,
		withHost("admin.testdomain.com"),
		withSession(token),
	)
	assertStatus(t, resp, http.StatusOK)

	// Downgrade role to "user" directly in DB (simulating admin action)
	_, err := s.db.Exec(`UPDATE auth_users SET role = 'user' WHERE id = ?`, userID)
	if err != nil {
		t.Fatalf("Failed to downgrade role: %v", err)
	}

	// Same session, same endpoint — should now be 403 (authenticated but not admin)
	resp2 := s.makeRequest(t, "GET", "/api/system/config", nil,
		withHost("admin.testdomain.com"),
		withSession(token),
	)
	defer resp2.Body.Close()

	if resp2.StatusCode == http.StatusOK {
		t.Error("Role downgrade not enforced — user still has admin access after demotion")
	}

	if resp2.StatusCode == http.StatusForbidden {
		t.Log("Role downgrade immediately enforced (no stale cache)")
	} else {
		t.Logf("Unexpected status after role downgrade: %d", resp2.StatusCode)
	}
}

// TestAdversarial_SessionTokenEntropy creates 50 sessions and verifies token quality:
// no collisions, sufficient length, good character distribution.
func TestAdversarial_SessionTokenEntropy(t *testing.T) {
	s := setupIntegrationTest(t)

	// Create a user
	userID := "user-entropy@test.com"
	s.db.Exec(`
		INSERT INTO auth_users (id, email, name, picture, provider, role, created_at)
		VALUES (?, ?, ?, '', 'test', 'user', ?)
	`, userID, "entropy@test.com", "Entropy Tester", time.Now().Unix())

	const tokenCount = 50
	tokens := make([]string, 0, tokenCount)
	hashes := make(map[string]bool, tokenCount)

	for i := 0; i < tokenCount; i++ {
		token, err := s.authService.CreateSession(userID)
		if err != nil {
			t.Fatalf("Failed to create session %d: %v", i, err)
		}
		tokens = append(tokens, token)

		// Track hash collisions
		h := sha256.Sum256([]byte(token))
		hashHex := hex.EncodeToString(h[:])
		if hashes[hashHex] {
			t.Errorf("Hash collision on session %d!", i)
		}
		hashes[hashHex] = true
	}

	// Check for token collisions
	tokenSet := make(map[string]bool, tokenCount)
	for _, token := range tokens {
		if tokenSet[token] {
			t.Error("Token collision detected — generateToken is not sufficiently random")
		}
		tokenSet[token] = true
	}

	// Check token length (32 bytes → base64url → ~44 chars)
	for i, token := range tokens {
		if len(token) < 40 {
			t.Errorf("Token %d too short: %d chars (expected ~44 for 32 bytes base64url)", i, len(token))
		}
	}

	// Check character distribution (should use many unique chars)
	charCounts := make(map[byte]int)
	for _, token := range tokens {
		for i := 0; i < len(token); i++ {
			charCounts[token[i]]++
		}
	}

	uniqueChars := len(charCounts)
	if uniqueChars < 20 {
		t.Errorf("Poor character distribution: only %d unique chars across %d tokens", uniqueChars, tokenCount)
	}

	t.Logf("Token entropy: %d tokens, %d unique chars, avg length=%d, zero collisions",
		tokenCount, uniqueChars, len(tokens[0]))
}
