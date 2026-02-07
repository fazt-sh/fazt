package handlers

import (
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"
	_ "modernc.org/sqlite"
)

// setupAuthTestDB creates a test database with auth tables
func setupAuthTestDB(t *testing.T) *sql.DB {
	return setupTestDB(t)
}

// setupTestAuthService creates a test auth service with a test user
func setupTestAuthService(t *testing.T) (*auth.Service, string) {
	db := setupAuthTestDB(t)
	service := auth.NewService(db, "test.local", false)

	// Create a test user
	user, err := service.GetOrCreateLocalAdmin("admin")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a session for the user
	token, err := service.CreateSession(user.ID)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	return service, token
}

// TestUserMeHandler_Success tests successful user info retrieval
func TestUserMeHandler_Success(t *testing.T) {
	silenceTestLogs(t)
	service, token := setupTestAuthService(t)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	InitAuth(service, limiter, "v0.8.0-test")

	req := testutil.JSONRequest("GET", "/api/user/me", nil)
	req = testutil.WithSession(req, token)

	rr := httptest.NewRecorder()
	UserMeHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, 200)
	testutil.AssertFieldEquals(t, data, "username", "admin")
	testutil.AssertFieldEquals(t, data, "version", "v0.8.0-test")
}

// TestUserMeHandler_Unauthorized tests user info without session
func TestUserMeHandler_Unauthorized(t *testing.T) {
	silenceTestLogs(t)
	service, _ := setupTestAuthService(t)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	InitAuth(service, limiter, "v0.8.0-test")

	req := testutil.JSONRequest("GET", "/api/user/me", nil)

	rr := httptest.NewRecorder()
	UserMeHandler(rr, req)

	testutil.CheckError(t, rr, 401, "UNAUTHORIZED")
}

// TestUserMeHandler_InvalidSession tests user info with invalid session
func TestUserMeHandler_InvalidSession(t *testing.T) {
	silenceTestLogs(t)
	service, _ := setupTestAuthService(t)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	InitAuth(service, limiter, "v0.8.0-test")

	req := testutil.JSONRequest("GET", "/api/user/me", nil)
	req = testutil.WithSession(req, "invalid-session-token")

	rr := httptest.NewRecorder()
	UserMeHandler(rr, req)

	testutil.CheckError(t, rr, 401, "UNAUTHORIZED")
}

// TestLoginHandler_Success tests successful login
func TestLoginHandler_Success(t *testing.T) {
	silenceTestLogs(t)
	db := setupAuthTestDB(t)
	service := auth.NewService(db, "test.local", false)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	InitAuth(service, limiter, "v0.8.0-test")

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

	req := testutil.JSONRequest("POST", "/api/login", map[string]interface{}{
		"username": "admin",
		"password": "testpassword123",
	})

	rr := httptest.NewRecorder()
	LoginHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, 200)
	testutil.AssertFieldExists(t, data, "message")

	// Verify session cookie was set
	cookies := rr.Result().Cookies()
	foundCookie := false
	for _, cookie := range cookies {
		if cookie.Name == "fazt_session" {
			foundCookie = true
			if cookie.Value == "" {
				t.Error("Session cookie value should not be empty")
			}
		}
	}
	if !foundCookie {
		t.Error("Expected fazt_session cookie to be set")
	}
}

// TestLoginHandler_InvalidCredentials tests login with wrong password
func TestLoginHandler_InvalidCredentials(t *testing.T) {
	silenceTestLogs(t)
	db := setupAuthTestDB(t)
	service := auth.NewService(db, "test.local", false)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	InitAuth(service, limiter, "v0.8.0-test")

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

	req := testutil.JSONRequest("POST", "/api/login", map[string]interface{}{
		"username": "admin",
		"password": "wrongpassword",
	})

	rr := httptest.NewRecorder()
	LoginHandler(rr, req)

	testutil.CheckError(t, rr, 401, "INVALID_CREDENTIALS")
}

// TestLoginHandler_InvalidUsername tests login with wrong username
func TestLoginHandler_InvalidUsername(t *testing.T) {
	silenceTestLogs(t)
	db := setupAuthTestDB(t)
	service := auth.NewService(db, "test.local", false)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	InitAuth(service, limiter, "v0.8.0-test")

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

	req := testutil.JSONRequest("POST", "/api/login", map[string]interface{}{
		"username": "wronguser",
		"password": "testpassword",
	})

	rr := httptest.NewRecorder()
	LoginHandler(rr, req)

	testutil.CheckError(t, rr, 401, "INVALID_CREDENTIALS")
}

// TestLoginHandler_InvalidJSON tests login with malformed JSON
func TestLoginHandler_InvalidJSON(t *testing.T) {
	silenceTestLogs(t)
	db := setupAuthTestDB(t)
	service := auth.NewService(db, "test.local", false)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	InitAuth(service, limiter, "v0.8.0-test")
	setupTestConfig(t)

	req := httptest.NewRequest("POST", "/api/login", nil)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	LoginHandler(rr, req)

	testutil.CheckError(t, rr, 400, "INVALID_JSON")
}

// TestLogoutHandler_Success tests successful logout
func TestLogoutHandler_Success(t *testing.T) {
	silenceTestLogs(t)
	service, token := setupTestAuthService(t)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	InitAuth(service, limiter, "v0.8.0-test")

	req := testutil.JSONRequest("POST", "/api/logout", nil)
	req = testutil.WithSession(req, token)

	rr := httptest.NewRecorder()
	LogoutHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, 200)
	testutil.AssertFieldExists(t, data, "message")

	// Verify session was deleted (next request should fail)
	req2 := testutil.JSONRequest("GET", "/api/user/me", nil)
	req2 = testutil.WithSession(req2, token)
	rr2 := httptest.NewRecorder()
	UserMeHandler(rr2, req2)
	testutil.CheckError(t, rr2, 401, "UNAUTHORIZED")
}

// TestLogoutHandler_NoSession tests logout without session
func TestLogoutHandler_NoSession(t *testing.T) {
	silenceTestLogs(t)
	db := setupAuthTestDB(t)
	service := auth.NewService(db, "test.local", false)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	InitAuth(service, limiter, "v0.8.0-test")

	req := testutil.JSONRequest("POST", "/api/logout", nil)

	rr := httptest.NewRecorder()
	LogoutHandler(rr, req)

	// Logout should still succeed even without session
	data := testutil.CheckSuccess(t, rr, 200)
	testutil.AssertFieldExists(t, data, "message")
}

// TestAuthStatusHandler_Authenticated tests auth status with valid session
func TestAuthStatusHandler_Authenticated(t *testing.T) {
	silenceTestLogs(t)
	service, token := setupTestAuthService(t)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	InitAuth(service, limiter, "v0.8.0-test")

	req := testutil.JSONRequest("GET", "/api/auth/status", nil)
	req = testutil.WithSession(req, token)

	rr := httptest.NewRecorder()
	AuthStatusHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, 200)

	if authenticated, ok := data["authenticated"].(bool); !ok || !authenticated {
		t.Error("Expected authenticated to be true")
	}

	testutil.AssertFieldEquals(t, data, "username", "admin")
}

// TestAuthStatusHandler_NotAuthenticated tests auth status without session
func TestAuthStatusHandler_NotAuthenticated(t *testing.T) {
	silenceTestLogs(t)
	db := setupAuthTestDB(t)
	service := auth.NewService(db, "test.local", false)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	InitAuth(service, limiter, "v0.8.0-test")

	req := testutil.JSONRequest("GET", "/api/auth/status", nil)

	rr := httptest.NewRecorder()
	AuthStatusHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, 200)

	if authenticated, ok := data["authenticated"].(bool); !ok || authenticated {
		t.Error("Expected authenticated to be false")
	}
}
