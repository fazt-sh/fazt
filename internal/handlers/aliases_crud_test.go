package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"
	"github.com/fazt-sh/fazt/internal/hosting"
	"golang.org/x/crypto/bcrypt"
)

func setupAliasTest(t *testing.T) string {
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

	testCfg := &config.Config{
		Server: config.ServerConfig{
			Domain: "test.local",
			Env:    "test",
		},
	}
	config.SetConfig(testCfg)

	// Set up auth with API key
	token := "alias-test-token"
	hash, _ := bcrypt.GenerateFromPassword([]byte(token), bcrypt.MinCost)
	_, err := db.Exec(`INSERT INTO api_keys (name, key_hash, scopes) VALUES (?, ?, ?)`, "test-key", string(hash), "deploy")
	if err != nil {
		t.Fatalf("Failed to insert API key: %v", err)
	}

	// Init auth service
	service := auth.NewService(db, "test.local", false)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() { limiter.Stop() })
	InitAuth(service, limiter, "v0.28.0-test")

	return token
}

// createAliasProxy creates a proxy alias in the test DB
func createAliasProxy(t *testing.T, subdomain, appID string) {
	t.Helper()
	db := database.GetDB()
	targets := fmt.Sprintf(`{"app_id":"%s"}`, appID)
	_, err := db.Exec(`
		INSERT INTO aliases (subdomain, type, targets, created_at, updated_at)
		VALUES (?, 'proxy', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, subdomain, targets)
	if err != nil {
		t.Fatalf("Failed to create alias: %v", err)
	}
}

// createAliasRedirect creates a redirect alias
func createAliasRedirect(t *testing.T, subdomain, url string) {
	t.Helper()
	db := database.GetDB()
	targets := fmt.Sprintf(`{"url":"%s"}`, url)
	_, err := db.Exec(`
		INSERT INTO aliases (subdomain, type, targets, created_at, updated_at)
		VALUES (?, 'redirect', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, subdomain, targets)
	if err != nil {
		t.Fatalf("Failed to create redirect alias: %v", err)
	}
}

// createAppForAlias creates a minimal app for alias tests
func createAppForAlias(t *testing.T, id string) {
	t.Helper()
	db := database.GetDB()
	_, err := db.Exec(`
		INSERT INTO apps (id, original_id, title, source, created_at, updated_at)
		VALUES (?, ?, 'Test App', 'deploy', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, id)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
}

// --- AliasCreateHandler ---

func TestAliasCreateHandler_Proxy(t *testing.T) {
	token := setupAliasTest(t)
	appID := "app_" + testutil.RandStr(8)
	createAppForAlias(t, appID)

	body := map[string]interface{}{
		"subdomain": "test-alias",
		"type":      "proxy",
		"app_id":    appID,
	}

	req := testutil.JSONRequest("POST", "/api/aliases", body)
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasCreateHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusCreated)
	testutil.AssertFieldEquals(t, data, "subdomain", "test-alias")
	testutil.AssertFieldEquals(t, data, "type", "proxy")
}

func TestAliasCreateHandler_Redirect(t *testing.T) {
	token := setupAliasTest(t)

	body := map[string]interface{}{
		"subdomain": "redir-alias",
		"type":      "redirect",
		"url":       "https://example.com",
	}

	req := testutil.JSONRequest("POST", "/api/aliases", body)
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasCreateHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusCreated)
	testutil.AssertFieldEquals(t, data, "type", "redirect")
}

func TestAliasCreateHandler_Reserved(t *testing.T) {
	token := setupAliasTest(t)

	body := map[string]interface{}{
		"subdomain": "reserved-test",
		"type":      "reserved",
	}

	req := testutil.JSONRequest("POST", "/api/aliases", body)
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasCreateHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusCreated)
	testutil.AssertFieldEquals(t, data, "type", "reserved")
}

func TestAliasCreateHandler_MissingSubdomain(t *testing.T) {
	token := setupAliasTest(t)

	body := map[string]interface{}{
		"type": "proxy",
	}

	req := testutil.JSONRequest("POST", "/api/aliases", body)
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasCreateHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestAliasCreateHandler_InvalidSubdomain(t *testing.T) {
	token := setupAliasTest(t)

	body := map[string]interface{}{
		"subdomain": "INVALID_NAME",
		"type":      "proxy",
	}

	req := testutil.JSONRequest("POST", "/api/aliases", body)
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasCreateHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestAliasCreateHandler_InvalidType(t *testing.T) {
	token := setupAliasTest(t)

	body := map[string]interface{}{
		"subdomain": "bad-type",
		"type":      "invalid",
	}

	req := testutil.JSONRequest("POST", "/api/aliases", body)
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasCreateHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestAliasCreateHandler_ProxyMissingAppID(t *testing.T) {
	token := setupAliasTest(t)

	body := map[string]interface{}{
		"subdomain": "no-app-id",
		"type":      "proxy",
	}

	req := testutil.JSONRequest("POST", "/api/aliases", body)
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasCreateHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestAliasCreateHandler_MethodNotAllowed(t *testing.T) {
	setupAliasTest(t)

	req := httptest.NewRequest("GET", "/api/aliases", nil)
	resp := httptest.NewRecorder()
	AliasCreateHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
}

// --- AliasDetailHandler ---

func TestAliasDetailHandler_Success(t *testing.T) {
	token := setupAliasTest(t)
	appID := "app_" + testutil.RandStr(8)
	createAppForAlias(t, appID)
	createAliasProxy(t, "detail-alias", appID)

	req := httptest.NewRequest("GET", "/api/aliases/detail-alias", nil)
	req.SetPathValue("subdomain", "detail-alias")
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasDetailHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "subdomain", "detail-alias")
}

func TestAliasDetailHandler_NotFound(t *testing.T) {
	token := setupAliasTest(t)

	req := httptest.NewRequest("GET", "/api/aliases/ghost", nil)
	req.SetPathValue("subdomain", "ghost")
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasDetailHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "ALIAS_NOT_FOUND")
}

func TestAliasDetailHandler_MissingSubdomain(t *testing.T) {
	token := setupAliasTest(t)

	req := httptest.NewRequest("GET", "/api/aliases/", nil)
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasDetailHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

// --- AliasDeleteHandler ---

func TestAliasDeleteHandler_Success(t *testing.T) {
	token := setupAliasTest(t)
	appID := "app_" + testutil.RandStr(8)
	createAppForAlias(t, appID)
	createAliasProxy(t, "del-alias", appID)

	req := httptest.NewRequest("DELETE", "/api/aliases/del-alias", nil)
	req.SetPathValue("subdomain", "del-alias")
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasDeleteHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "message", "Alias deleted")
}

func TestAliasDeleteHandler_NotFound(t *testing.T) {
	token := setupAliasTest(t)

	req := httptest.NewRequest("DELETE", "/api/aliases/ghost", nil)
	req.SetPathValue("subdomain", "ghost")
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasDeleteHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "ALIAS_NOT_FOUND")
}

func TestAliasDeleteHandler_SystemAlias(t *testing.T) {
	token := setupAliasTest(t)
	// 'admin' is a system reserved alias created by migrations

	req := httptest.NewRequest("DELETE", "/api/aliases/admin", nil)
	req.SetPathValue("subdomain", "admin")
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasDeleteHandler(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Errorf("Expected 403 for system alias, got %d. Body: %s", resp.Code, resp.Body.String())
	}
}

func TestAliasDeleteHandler_MethodNotAllowed(t *testing.T) {
	setupAliasTest(t)

	req := httptest.NewRequest("GET", "/api/aliases/whatever", nil)
	req.SetPathValue("subdomain", "whatever")
	resp := httptest.NewRecorder()
	AliasDeleteHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
}

// --- AliasReserveHandler ---

func TestAliasReserveHandler_Success(t *testing.T) {
	token := setupAliasTest(t)

	req := httptest.NewRequest("POST", "/api/aliases/reserve-me/reserve", nil)
	req.SetPathValue("subdomain", "reserve-me")
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasReserveHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusCreated)
	testutil.AssertFieldEquals(t, data, "type", "reserved")
}

func TestAliasReserveHandler_InvalidSubdomain(t *testing.T) {
	token := setupAliasTest(t)

	req := httptest.NewRequest("POST", "/api/aliases/INVALID/reserve", nil)
	req.SetPathValue("subdomain", "INVALID")
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasReserveHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestAliasReserveHandler_MissingSubdomain(t *testing.T) {
	token := setupAliasTest(t)

	req := httptest.NewRequest("POST", "/api/aliases//reserve", nil)
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasReserveHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

// --- AliasSwapHandler ---

func TestAliasSwapHandler_Success(t *testing.T) {
	token := setupAliasTest(t)
	app1 := "app_" + testutil.RandStr(8)
	app2 := "app_" + testutil.RandStr(8)
	createAppForAlias(t, app1)
	createAppForAlias(t, app2)
	createAliasProxy(t, "swap-a", app1)
	createAliasProxy(t, "swap-b", app2)

	body := map[string]interface{}{
		"alias1": "swap-a",
		"alias2": "swap-b",
	}

	req := testutil.JSONRequest("POST", "/api/aliases/swap", body)
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasSwapHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "message", "Aliases swapped")
}

func TestAliasSwapHandler_MissingAlias(t *testing.T) {
	token := setupAliasTest(t)

	body := map[string]interface{}{
		"alias1": "swap-a",
	}

	req := testutil.JSONRequest("POST", "/api/aliases/swap", body)
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasSwapHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestAliasSwapHandler_AliasNotFound(t *testing.T) {
	token := setupAliasTest(t)
	app1 := "app_" + testutil.RandStr(8)
	createAppForAlias(t, app1)
	createAliasProxy(t, "swap-exists", app1)

	body := map[string]interface{}{
		"alias1": "swap-exists",
		"alias2": "swap-ghost",
	}

	req := testutil.JSONRequest("POST", "/api/aliases/swap", body)
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasSwapHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestAliasSwapHandler_NonProxyType(t *testing.T) {
	token := setupAliasTest(t)
	app1 := "app_" + testutil.RandStr(8)
	createAppForAlias(t, app1)
	createAliasProxy(t, "swap-proxy", app1)
	createAliasRedirect(t, "swap-redir", "https://example.com")

	body := map[string]interface{}{
		"alias1": "swap-proxy",
		"alias2": "swap-redir",
	}

	req := testutil.JSONRequest("POST", "/api/aliases/swap", body)
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasSwapHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for non-proxy swap, got %d", resp.Code)
	}
}

// --- AliasSplitHandler ---

func TestAliasSplitHandler_Success(t *testing.T) {
	token := setupAliasTest(t)
	app1 := "app_" + testutil.RandStr(8)
	app2 := "app_" + testutil.RandStr(8)
	createAppForAlias(t, app1)
	createAppForAlias(t, app2)

	body := map[string]interface{}{
		"targets": []map[string]interface{}{
			{"app_id": app1, "weight": 70},
			{"app_id": app2, "weight": 30},
		},
	}

	req := testutil.JSONRequest("POST", "/api/aliases/split-alias/split", body)
	req.SetPathValue("subdomain", "split-alias")
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasSplitHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusCreated)
	testutil.AssertFieldEquals(t, data, "type", "split")
}

func TestAliasSplitHandler_WeightsDontSumTo100(t *testing.T) {
	token := setupAliasTest(t)

	body := map[string]interface{}{
		"targets": []map[string]interface{}{
			{"app_id": "app1", "weight": 60},
			{"app_id": "app2", "weight": 30},
		},
	}

	req := testutil.JSONRequest("POST", "/api/aliases/bad-split/split", body)
	req.SetPathValue("subdomain", "bad-split")
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasSplitHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestAliasSplitHandler_TooFewTargets(t *testing.T) {
	token := setupAliasTest(t)

	body := map[string]interface{}{
		"targets": []map[string]interface{}{
			{"app_id": "app1", "weight": 100},
		},
	}

	req := testutil.JSONRequest("POST", "/api/aliases/single-split/split", body)
	req.SetPathValue("subdomain", "single-split")
	testutil.WithAuth(req, token)
	resp := httptest.NewRecorder()
	AliasSplitHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

// ResolveAlias and GetRedirectURL tests are in aliases_test.go

// --- GetRedirectURL (additional handler-level tests) ---

func TestGetRedirectURL_WithRedirectAlias(t *testing.T) {
	setupAliasTest(t)
	createAliasRedirect(t, "redir-url-test", "https://example.com/target")

	url, err := GetRedirectURL("redir-url-test")
	if err != nil {
		t.Fatalf("GetRedirectURL failed: %v", err)
	}
	if url != "https://example.com/target" {
		t.Errorf("Expected redirect URL, got '%s'", url)
	}
}

func TestGetRedirectURL_NonRedirectType(t *testing.T) {
	setupAliasTest(t)
	appID := "app_" + testutil.RandStr(8)
	createAppForAlias(t, appID)
	createAliasProxy(t, "not-redirect", appID)

	url, err := GetRedirectURL("not-redirect")
	if err != nil {
		t.Fatalf("GetRedirectURL failed: %v", err)
	}
	if url != "" {
		t.Errorf("Expected empty URL for non-redirect, got '%s'", url)
	}
}

// isValidSubdomain tests are in aliases_handler_test.go
