package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"
	"github.com/fazt-sh/fazt/internal/hosting"
)

func setupSystemHandlerTest(t *testing.T) (*auth.Service, string, string) {
	t.Helper()

	silenceTestLogs(t)

	db := setupTestDB(t)
	database.SetDB(db)
	t.Cleanup(func() {
		db.Close()
		database.SetDB(nil)
	})

	if err := hosting.Init(db); err != nil {
		t.Fatalf("Failed to init hosting: %v", err)
	}

	// Set up auth service
	service := auth.NewService(db, "test.local", false)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	InitAuth(service, limiter, "v0.8.0-test")

	// Create test user and session
	user, err := service.GetOrCreateLocalAdmin("admin")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	token, err := service.CreateSession(user.ID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create API key
	apiKey := "system-token-123"
	insertTestAPIKey(t, db, apiKey)

	// Set test config
	testCfg := &config.Config{
		Server: config.ServerConfig{
			Domain: "test.local",
			Env:    "test",
		},
		Database: config.DatabaseConfig{
			Path: ":memory:",
		},
	}
	config.SetConfig(testCfg)

	return service, token, apiKey
}

// TestSystemHealthHandler_Success tests system health endpoint
func TestSystemHealthHandler_Success(t *testing.T) {
	_, sessionToken, _ := setupSystemHandlerTest(t)

	req := testutil.JSONRequest("GET", "/api/system/health", nil)
	req = testutil.WithSession(req, sessionToken)
	rr := httptest.NewRecorder()

	SystemHealthHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "status", "healthy")
	testutil.AssertFieldExists(t, data, "uptime_seconds")
	testutil.AssertFieldExists(t, data, "memory")
	testutil.AssertFieldExists(t, data, "database")
}

// TestSystemHealthHandler_Unauthorized tests system health without auth
func TestSystemHealthHandler_Unauthorized(t *testing.T) {
	setupSystemHandlerTest(t)

	req := testutil.JSONRequest("GET", "/api/system/health", nil)
	rr := httptest.NewRecorder()

	SystemHealthHandler(rr, req)

	testutil.CheckError(t, rr, http.StatusUnauthorized, "UNAUTHORIZED")
}

// TestSystemHealthHandler_WithAPIKey tests system health with API key auth
func TestSystemHealthHandler_WithAPIKey(t *testing.T) {
	_, _, apiKey := setupSystemHandlerTest(t)

	req := testutil.JSONRequest("GET", "/api/system/health", nil)
	req = testutil.WithAuth(req, apiKey)
	rr := httptest.NewRecorder()

	SystemHealthHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "status", "healthy")
}

// TestSystemLimitsHandler_Success tests system limits endpoint
func TestSystemLimitsHandler_Success(t *testing.T) {
	setupSystemHandlerTest(t)

	req := testutil.JSONRequest("GET", "/api/system/limits", nil)
	rr := httptest.NewRecorder()

	SystemLimitsHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, http.StatusOK)
	if data == nil {
		t.Error("Expected non-nil response data")
	}
}

// TestSystemLimitsSchemaHandler_Success tests limits schema endpoint
func TestSystemLimitsSchemaHandler_Success(t *testing.T) {
	setupSystemHandlerTest(t)

	req := testutil.JSONRequest("GET", "/api/system/limits/schema", nil)
	rr := httptest.NewRecorder()

	SystemLimitsSchemaHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

// TestSystemCacheHandler_Success tests VFS cache stats endpoint
func TestSystemCacheHandler_Success(t *testing.T) {
	setupSystemHandlerTest(t)

	req := testutil.JSONRequest("GET", "/api/system/cache", nil)
	rr := httptest.NewRecorder()

	SystemCacheHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, http.StatusOK)
	if data == nil {
		t.Error("Expected non-nil response data")
	}
}

// TestSystemDBHandler_Success tests database stats endpoint
func TestSystemDBHandler_Success(t *testing.T) {
	setupSystemHandlerTest(t)

	req := testutil.JSONRequest("GET", "/api/system/db", nil)
	rr := httptest.NewRecorder()

	SystemDBHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, http.StatusOK)
	if data == nil {
		t.Error("Expected non-nil response data")
	}
}

// TestSystemConfigHandler_Success tests config endpoint
func TestSystemConfigHandler_Success(t *testing.T) {
	setupSystemHandlerTest(t)

	req := testutil.JSONRequest("GET", "/api/system/config", nil)
	rr := httptest.NewRecorder()

	SystemConfigHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, http.StatusOK)
	testutil.AssertFieldExists(t, data, "version")
	testutil.AssertFieldEquals(t, data, "domain", "test.local")
}
