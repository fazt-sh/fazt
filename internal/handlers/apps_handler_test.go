package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"
	"github.com/fazt-sh/fazt/internal/hosting"
)

func setupAppsTest(t *testing.T) {
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
}

// createTestApp inserts a test app into the database
func createTestApp(t *testing.T, title string) string {
	t.Helper()
	db := database.GetDB()

	id := testutil.RandStr(8)
	_, err := db.Exec(`
		INSERT INTO apps (id, original_id, title, source, visibility, created_at, updated_at)
		VALUES (?, ?, ?, 'deploy', 'unlisted', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, id, title)
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	return id
}

// Note: AppsListHandler and AppDetailHandler use v1 schema (a.name, a.manifest)
// which doesn't exist after migration 012. Testing method guards and error paths only.

func TestAppsListHandler_MethodNotAllowed(t *testing.T) {
	setupAppsTest(t)

	req := httptest.NewRequest("POST", "/api/apps", nil)
	resp := httptest.NewRecorder()
	AppsListHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
}

func TestAppDetailHandler_MissingID(t *testing.T) {
	setupAppsTest(t)

	req := httptest.NewRequest("GET", "/api/apps/", nil)
	resp := httptest.NewRecorder()
	AppDetailHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

// --- AppDeleteHandler ---

func TestAppDeleteHandler_Success(t *testing.T) {
	setupAppsTest(t)
	id := createTestApp(t, "del-app")

	req := httptest.NewRequest("DELETE", "/api/apps/"+id, nil)
	req.SetPathValue("id", id)
	resp := httptest.NewRecorder()
	AppDeleteHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "message", "App deleted")
}

func TestAppDeleteHandler_NotFound(t *testing.T) {
	setupAppsTest(t)

	req := httptest.NewRequest("DELETE", "/api/apps/ghost", nil)
	req.SetPathValue("id", "ghost")
	resp := httptest.NewRecorder()
	AppDeleteHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "APP_NOT_FOUND")
}

func TestAppDeleteHandler_SystemApp(t *testing.T) {
	setupAppsTest(t)
	db := database.GetDB()
	// Insert system app (title = "root")
	_, err := db.Exec(`
		INSERT INTO apps (id, original_id, title, source, created_at, updated_at)
		VALUES ('sys-root', 'sys-root', 'root', 'system', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to create system app: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/api/apps/sys-root", nil)
	req.SetPathValue("id", "sys-root")
	resp := httptest.NewRecorder()
	AppDeleteHandler(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Errorf("Expected 403 for system app, got %d. Body: %s", resp.Code, resp.Body.String())
	}
}

func TestAppDeleteHandler_MethodNotAllowed(t *testing.T) {
	setupAppsTest(t)

	req := httptest.NewRequest("GET", "/api/apps/whatever", nil)
	req.SetPathValue("id", "whatever")
	resp := httptest.NewRecorder()
	AppDeleteHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
}

func TestAppDeleteHandler_MissingID(t *testing.T) {
	setupAppsTest(t)

	req := httptest.NewRequest("DELETE", "/api/apps/", nil)
	resp := httptest.NewRecorder()
	AppDeleteHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

// --- AppSourceHandler ---

func TestAppSourceHandler_Success(t *testing.T) {
	setupAppsTest(t)
	db := database.GetDB()
	_, err := db.Exec(`
		INSERT INTO apps (id, original_id, title, source, source_url, source_ref, source_commit, created_at, updated_at)
		VALUES ('src-app', 'src-app', 'src-app', 'git', 'https://github.com/test/repo', 'main', 'abc1234', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/apps/src-app/source", nil)
	req.SetPathValue("id", "src-app")
	resp := httptest.NewRecorder()
	AppSourceHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "type", "git")
	testutil.AssertFieldEquals(t, data, "url", "https://github.com/test/repo")
	testutil.AssertFieldEquals(t, data, "ref", "main")
	testutil.AssertFieldEquals(t, data, "commit", "abc1234")
}

func TestAppSourceHandler_NotFound(t *testing.T) {
	setupAppsTest(t)

	req := httptest.NewRequest("GET", "/api/apps/ghost/source", nil)
	req.SetPathValue("id", "ghost")
	resp := httptest.NewRecorder()
	AppSourceHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "APP_NOT_FOUND")
}

func TestAppSourceHandler_MissingID(t *testing.T) {
	setupAppsTest(t)

	req := httptest.NewRequest("GET", "/api/apps//source", nil)
	resp := httptest.NewRecorder()
	AppSourceHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

// --- AppFilesHandler ---

func TestAppFilesHandler_NotFound(t *testing.T) {
	setupAppsTest(t)

	req := httptest.NewRequest("GET", "/api/apps/ghost/files", nil)
	req.SetPathValue("id", "ghost")
	resp := httptest.NewRecorder()
	AppFilesHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "APP_NOT_FOUND")
}

func TestAppFilesHandler_MissingID(t *testing.T) {
	setupAppsTest(t)

	req := httptest.NewRequest("GET", "/api/apps//files", nil)
	resp := httptest.NewRecorder()
	AppFilesHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestAppFilesHandler_WithFiles(t *testing.T) {
	setupAppsTest(t)
	id := createTestApp(t, "files-app")
	createTestSite(t, database.GetDB(), "files-app")

	req := httptest.NewRequest("GET", "/api/apps/"+id+"/files", nil)
	req.SetPathValue("id", id)
	resp := httptest.NewRecorder()
	AppFilesHandler(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d. Body: %s", resp.Code, resp.Body.String())
	}
}

// --- formatTime ---

func TestFormatTime(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{"2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z"},
		{nil, ""},
		{42, ""},
	}

	for _, tt := range tests {
		result := formatTime(tt.input)
		if result != tt.expected {
			t.Errorf("formatTime(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// --- isValidAppName ---

func TestIsValidAppName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"my-app", true},
		{"app123", true},
		{"a", true},
		{"", false},
		{"-starts-with-dash", false},
		{"ends-with-dash-", false},
		{"has spaces", false},
		{"UPPERCASE", false},
		{"has.dot", false},
		{"has_underscore", false},
		// 63 chars (max length)
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true},
		// 64 chars (too long)
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},
	}

	for _, tt := range tests {
		result := isValidAppName(tt.name)
		if result != tt.expected {
			t.Errorf("isValidAppName(%q) = %v, want %v", tt.name, result, tt.expected)
		}
	}
}

// --- TemplatesListHandler ---

func TestTemplatesListHandler(t *testing.T) {
	setupAppsTest(t)

	req := httptest.NewRequest("GET", "/api/apps/templates", nil)
	resp := httptest.NewRecorder()
	TemplatesListHandler(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d. Body: %s", resp.Code, resp.Body.String())
	}
}
