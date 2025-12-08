package handlers

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"
)

// TestUserMeHandler_Success tests successful user info retrieval
func TestUserMeHandler_Success(t *testing.T) {
	// Setup
	silenceTestLogs(t)
	store, sessionID := setupTestAuth(t)
	setupTestConfig(t)
	InitAuth(store, auth.NewRateLimiter(), "v0.8.0-test")

	// Create request with valid session
	req := testutil.JSONRequest("GET", "/api/user/me", nil)
	req = testutil.WithSession(req, sessionID)

	rr := httptest.NewRecorder()

	// Execute
	UserMeHandler(rr, req)

	// Assert
	data := testutil.CheckSuccess(t, rr, 200)
	testutil.AssertFieldEquals(t, data, "username", "admin")
	testutil.AssertFieldEquals(t, data, "version", "v0.8.0-test")
}

// TestUserMeHandler_Unauthorized tests user info without session
func TestUserMeHandler_Unauthorized(t *testing.T) {
	// Setup
	silenceTestLogs(t)
	store, _ := setupTestAuth(t)
	InitAuth(store, auth.NewRateLimiter(), "v0.8.0-test")

	// Create request WITHOUT session cookie
	req := testutil.JSONRequest("GET", "/api/user/me", nil)

	rr := httptest.NewRecorder()

	// Execute
	UserMeHandler(rr, req)

	// Assert - should return error response
	testutil.CheckError(t, rr, 401, "UNAUTHORIZED")
}

// TestUserMeHandler_ExpiredSession tests user info with expired session
func TestUserMeHandler_ExpiredSession(t *testing.T) {
	// Setup
	silenceTestLogs(t)

	// Create store with very short TTL
	store := auth.NewSessionStore(1 * time.Millisecond)
	sessionID, _ := store.CreateSession("admin")

	// Wait for session to expire
	time.Sleep(10 * time.Millisecond)

	InitAuth(store, auth.NewRateLimiter(), "v0.8.0-test")

	// Create request with expired session
	req := testutil.JSONRequest("GET", "/api/user/me", nil)
	req = testutil.WithSession(req, sessionID)

	rr := httptest.NewRecorder()

	// Execute
	UserMeHandler(rr, req)

	// Assert - should return SESSION_EXPIRED (more specific than UNAUTHORIZED)
	testutil.CheckError(t, rr, 401, "SESSION_EXPIRED")
}

// TestLoginHandler_Success tests successful login
func TestLoginHandler_Success(t *testing.T) {
	// Setup
	silenceTestLogs(t)
	store := auth.NewSessionStore(24 * time.Hour)
	limiter := auth.NewRateLimiter()
	InitAuth(store, limiter, "v0.8.0-test")

	// Setup config with known password
	passwordHash, _ := auth.HashPassword("testpassword123")
	testCfg := &config.Config{
		Server: config.ServerConfig{
			Env: "test",
		},
		Auth: config.AuthConfig{
			Username:     "admin",
			PasswordHash: passwordHash,
		},
	}
	config.SetConfig(testCfg)

	// Create login request
	req := testutil.JSONRequest("POST", "/api/login", map[string]interface{}{
		"username": "admin",
		"password": "testpassword123",
	})

	rr := httptest.NewRecorder()

	// Execute
	LoginHandler(rr, req)

	// Assert
	data := testutil.CheckSuccess(t, rr, 200)
	testutil.AssertFieldExists(t, data, "message")

	// Verify session cookie was set
	cookies := rr.Result().Cookies()
	foundCookie := false
	for _, cookie := range cookies {
		if cookie.Name == "cc_session" { // Correct cookie name
			foundCookie = true
			if cookie.Value == "" {
				t.Error("Session cookie value should not be empty")
			}
		}
	}
	if !foundCookie {
		t.Error("Expected cc_session cookie to be set")
	}
}

// TestLoginHandler_InvalidCredentials tests login with wrong password
func TestLoginHandler_InvalidCredentials(t *testing.T) {
	// Setup
	silenceTestLogs(t)
	store := auth.NewSessionStore(24 * time.Hour)
	limiter := auth.NewRateLimiter()
	InitAuth(store, limiter, "v0.8.0-test")

	// Setup config
	passwordHash, _ := auth.HashPassword("correctpassword")
	testCfg := &config.Config{
		Server: config.ServerConfig{
			Env: "test",
		},
		Auth: config.AuthConfig{
			Username:     "admin",
			PasswordHash: passwordHash,
		},
	}
	config.SetConfig(testCfg)

	// Create login request with wrong password
	req := testutil.JSONRequest("POST", "/api/login", map[string]interface{}{
		"username": "admin",
		"password": "wrongpassword",
	})

	rr := httptest.NewRecorder()

	// Execute
	LoginHandler(rr, req)

	// Assert
	testutil.CheckError(t, rr, 401, "INVALID_CREDENTIALS")
}

// TestLoginHandler_InvalidUsername tests login with wrong username
func TestLoginHandler_InvalidUsername(t *testing.T) {
	// Setup
	silenceTestLogs(t)
	store := auth.NewSessionStore(24 * time.Hour)
	limiter := auth.NewRateLimiter()
	InitAuth(store, limiter, "v0.8.0-test")

	// Setup config
	passwordHash, _ := auth.HashPassword("testpassword")
	testCfg := &config.Config{
		Server: config.ServerConfig{
			Env: "test",
		},
		Auth: config.AuthConfig{
			Username:     "admin",
			PasswordHash: passwordHash,
		},
	}
	config.SetConfig(testCfg)

	// Create login request with wrong username
	req := testutil.JSONRequest("POST", "/api/login", map[string]interface{}{
		"username": "wronguser",
		"password": "testpassword",
	})

	rr := httptest.NewRecorder()

	// Execute
	LoginHandler(rr, req)

	// Assert
	testutil.CheckError(t, rr, 401, "INVALID_CREDENTIALS")
}

// TestLoginHandler_InvalidJSON tests login with malformed JSON
func TestLoginHandler_InvalidJSON(t *testing.T) {
	// Setup
	silenceTestLogs(t)
	store := auth.NewSessionStore(24 * time.Hour)
	limiter := auth.NewRateLimiter()
	InitAuth(store, limiter, "v0.8.0-test")
	setupTestConfig(t)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/api/login", nil)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Execute
	LoginHandler(rr, req)

	// Assert - should return INVALID_JSON (more specific than BAD_REQUEST)
	testutil.CheckError(t, rr, 400, "INVALID_JSON")
}

// TestLogoutHandler_Success tests successful logout
func TestLogoutHandler_Success(t *testing.T) {
	// Setup
	silenceTestLogs(t)
	store, sessionID := setupTestAuth(t)
	InitAuth(store, auth.NewRateLimiter(), "v0.8.0-test")

	// Create logout request with session
	req := testutil.JSONRequest("POST", "/api/logout", nil)
	req = testutil.WithSession(req, sessionID)

	rr := httptest.NewRecorder()

	// Execute
	LogoutHandler(rr, req)

	// Assert
	data := testutil.CheckSuccess(t, rr, 200)
	testutil.AssertFieldExists(t, data, "message")

	// Verify session was deleted
	_, err := store.GetSession(sessionID)
	if err == nil {
		t.Error("Expected session to be deleted after logout")
	}
}

// TestLogoutHandler_NoSession tests logout without session
func TestLogoutHandler_NoSession(t *testing.T) {
	// Setup
	silenceTestLogs(t)
	store := auth.NewSessionStore(24 * time.Hour)
	InitAuth(store, auth.NewRateLimiter(), "v0.8.0-test")

	// Create logout request WITHOUT session
	req := testutil.JSONRequest("POST", "/api/logout", nil)

	rr := httptest.NewRecorder()

	// Execute
	LogoutHandler(rr, req)

	// Assert - logout should still succeed even without session
	data := testutil.CheckSuccess(t, rr, 200)
	testutil.AssertFieldExists(t, data, "message")
}

// TestAuthStatusHandler_Authenticated tests auth status with valid session
func TestAuthStatusHandler_Authenticated(t *testing.T) {
	// Setup
	silenceTestLogs(t)
	store, sessionID := setupTestAuth(t)
	InitAuth(store, auth.NewRateLimiter(), "v0.8.0-test")

	// Create request with valid session
	req := testutil.JSONRequest("GET", "/api/auth/status", nil)
	req = testutil.WithSession(req, sessionID)

	rr := httptest.NewRecorder()

	// Execute
	AuthStatusHandler(rr, req)

	// Assert
	data := testutil.CheckSuccess(t, rr, 200)

	// Check authenticated field
	if auth, ok := data["authenticated"].(bool); !ok || !auth {
		t.Error("Expected authenticated to be true")
	}

	testutil.AssertFieldEquals(t, data, "username", "admin")
	testutil.AssertFieldExists(t, data, "expiresAt")
}

// TestAuthStatusHandler_NotAuthenticated tests auth status without session
func TestAuthStatusHandler_NotAuthenticated(t *testing.T) {
	// Setup
	silenceTestLogs(t)
	store := auth.NewSessionStore(24 * time.Hour)
	InitAuth(store, auth.NewRateLimiter(), "v0.8.0-test")

	// Create request WITHOUT session
	req := testutil.JSONRequest("GET", "/api/auth/status", nil)

	rr := httptest.NewRecorder()

	// Execute
	AuthStatusHandler(rr, req)

	// Assert
	data := testutil.CheckSuccess(t, rr, 200)

	// Check authenticated field
	if auth, ok := data["authenticated"].(bool); !ok || auth {
		t.Error("Expected authenticated to be false")
	}
}
