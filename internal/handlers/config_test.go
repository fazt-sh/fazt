package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"
)

func setupConfigHandlerTest(t *testing.T) {
	t.Helper()
	silenceTestLogs(t)

	testCfg := &config.Config{
		Server: config.ServerConfig{
			Port:   "8080",
			Domain: "test.local",
			Env:    "test",
		},
		Database: config.DatabaseConfig{
			Path: ":memory:",
		},
		Auth: config.AuthConfig{
			Username: "admin",
		},
		Ntfy: config.NtfyConfig{
			Topic: "test-topic",
			URL:   "https://ntfy.sh",
		},
	}
	config.SetConfig(testCfg)
}

// TestConfigHandler_Success tests getting sanitized config
func TestConfigHandler_Success(t *testing.T) {
	setupConfigHandlerTest(t)

	req := testutil.JSONRequest("GET", "/api/config", nil)
	rr := httptest.NewRecorder()

	ConfigHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, http.StatusOK)
	testutil.AssertFieldExists(t, data, "server")
	testutil.AssertFieldExists(t, data, "database")
	testutil.AssertFieldExists(t, data, "auth")
	testutil.AssertFieldExists(t, data, "ntfy")
}

// TestConfigHandler_MethodNotAllowed tests non-GET request
func TestConfigHandler_MethodNotAllowed(t *testing.T) {
	setupConfigHandlerTest(t)

	req := testutil.JSONRequest("POST", "/api/config", nil)
	rr := httptest.NewRecorder()

	ConfigHandler(rr, req)

	testutil.CheckError(t, rr, http.StatusBadRequest, "BAD_REQUEST")
}

// TestConfigHandler_PasswordNotExposed tests that password hash is not exposed
func TestConfigHandler_PasswordNotExposed(t *testing.T) {
	setupConfigHandlerTest(t)

	req := testutil.JSONRequest("GET", "/api/config", nil)
	rr := httptest.NewRecorder()

	ConfigHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, http.StatusOK)

	// Check that auth section exists
	authSection, ok := data["auth"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected auth section to be a map")
	}

	// Verify password_hash is NOT present
	if _, exists := authSection["password_hash"]; exists {
		t.Error("password_hash should not be exposed in config endpoint")
	}

	// Verify username IS present
	if _, exists := authSection["username"]; !exists {
		t.Error("username should be present in config endpoint")
	}
}
