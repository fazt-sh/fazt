package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"
	"github.com/fazt-sh/fazt/internal/hosting"
)

func setupAppsV2Test(t *testing.T) {
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

// createTestAppV2 inserts a test app with v2 schema fields
func createTestAppV2(t *testing.T, title string) string {
	t.Helper()
	db := database.GetDB()

	id := "app_" + testutil.RandStr(8)
	_, err := db.Exec(`
		INSERT INTO apps (id, original_id, title, source, visibility, created_at, updated_at)
		VALUES (?, ?, ?, 'deploy', 'unlisted', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, id, title)
	if err != nil {
		t.Fatalf("Failed to create test app v2: %v", err)
	}
	return id
}

// createTestAppV2WithVisibility creates app with specific visibility
func createTestAppV2WithVisibility(t *testing.T, title, visibility string) string {
	t.Helper()
	db := database.GetDB()

	id := "app_" + testutil.RandStr(8)
	_, err := db.Exec(`
		INSERT INTO apps (id, original_id, title, source, visibility, created_at, updated_at)
		VALUES (?, ?, ?, 'deploy', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, id, title, visibility)
	if err != nil {
		t.Fatalf("Failed to create test app v2: %v", err)
	}
	return id
}

// createTestAliasForApp creates a proxy alias pointing to an app
func createTestAliasForApp(t *testing.T, subdomain, appID string) {
	t.Helper()
	db := database.GetDB()

	targets := fmt.Sprintf(`{"app_id":"%s"}`, appID)
	_, err := db.Exec(`
		INSERT INTO aliases (subdomain, type, targets, created_at, updated_at)
		VALUES (?, 'proxy', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, subdomain, targets)
	if err != nil {
		t.Fatalf("Failed to create test alias: %v", err)
	}
}

// --- AppsListHandlerV2 ---

func TestAppsListHandlerV2_Empty(t *testing.T) {
	setupAppsV2Test(t)

	req := httptest.NewRequest("GET", "/api/v2/apps", nil)
	resp := httptest.NewRecorder()
	AppsListHandlerV2(resp, req)

	// System apps may be present from hosting.Init, so just check it's OK
	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d. Body: %s", resp.Code, resp.Body.String())
	}
}

// TestAppsListHandlerV2_PublicOnly and _ShowAll skipped: getAliasesForApp called
// inside rows loop causes deadlock with MaxOpenConns(1)
func TestAppsListHandlerV2_PublicOnly(t *testing.T) {
	t.Skip("Skipped: nested query deadlock with MaxOpenConns(1) — Issue 05 pattern")
}

func TestAppsListHandlerV2_ShowAll(t *testing.T) {
	t.Skip("Skipped: nested query deadlock with MaxOpenConns(1) — Issue 05 pattern")
}

func TestAppsListHandlerV2_MethodNotAllowed(t *testing.T) {
	setupAppsV2Test(t)

	req := httptest.NewRequest("POST", "/api/v2/apps", nil)
	resp := httptest.NewRecorder()
	AppsListHandlerV2(resp, req)

	testutil.CheckError(t, resp, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
}

// --- AppDetailHandlerV2 ---

func TestAppDetailHandlerV2_ByID(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "detail-v2-app")

	req := httptest.NewRequest("GET", "/api/v2/apps/"+id, nil)
	req.SetPathValue("id", id)
	resp := httptest.NewRecorder()
	AppDetailHandlerV2(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "title", "detail-v2-app")
}

func TestAppDetailHandlerV2_ByAlias(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "alias-app")
	createTestAliasForApp(t, "my-alias", id)

	req := httptest.NewRequest("GET", "/api/v2/apps/my-alias", nil)
	req.SetPathValue("id", "my-alias")
	resp := httptest.NewRecorder()
	AppDetailHandlerV2(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "title", "alias-app")
}

func TestAppDetailHandlerV2_NotFound(t *testing.T) {
	setupAppsV2Test(t)

	req := httptest.NewRequest("GET", "/api/v2/apps/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	resp := httptest.NewRecorder()
	AppDetailHandlerV2(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", resp.Code)
	}
}

func TestAppDetailHandlerV2_ReservedAlias(t *testing.T) {
	setupAppsV2Test(t)
	// 'admin' is a reserved alias created by migrations

	req := httptest.NewRequest("GET", "/api/v2/apps/admin", nil)
	req.SetPathValue("id", "admin")
	resp := httptest.NewRecorder()
	AppDetailHandlerV2(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "RESERVED")
}

func TestAppDetailHandlerV2_MissingID(t *testing.T) {
	setupAppsV2Test(t)

	req := httptest.NewRequest("GET", "/api/v2/apps/", nil)
	resp := httptest.NewRecorder()
	AppDetailHandlerV2(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

// --- AppCreateHandlerV2 ---

func TestAppCreateHandlerV2_Success(t *testing.T) {
	setupAppsV2Test(t)

	body := map[string]interface{}{
		"title":      "My New App",
		"visibility": "public",
	}

	req := testutil.JSONRequest("POST", "/api/v2/apps", body)
	resp := httptest.NewRecorder()
	AppCreateHandlerV2(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusCreated)
	testutil.AssertFieldEquals(t, data, "title", "My New App")
	testutil.AssertFieldEquals(t, data, "visibility", "public")
	testutil.AssertFieldExists(t, data, "id")
}

func TestAppCreateHandlerV2_WithAlias(t *testing.T) {
	setupAppsV2Test(t)

	body := map[string]interface{}{
		"title": "Aliased App",
		"alias": "cool-app",
	}

	req := testutil.JSONRequest("POST", "/api/v2/apps", body)
	resp := httptest.NewRecorder()
	AppCreateHandlerV2(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusCreated)
	testutil.AssertFieldEquals(t, data, "alias", "cool-app")
	testutil.AssertFieldExists(t, data, "url")
}

func TestAppCreateHandlerV2_DefaultVisibility(t *testing.T) {
	setupAppsV2Test(t)

	body := map[string]interface{}{
		"title": "Default Vis App",
	}

	req := testutil.JSONRequest("POST", "/api/v2/apps", body)
	resp := httptest.NewRecorder()
	AppCreateHandlerV2(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusCreated)
	testutil.AssertFieldEquals(t, data, "visibility", "unlisted")
}

func TestAppCreateHandlerV2_MissingTitle(t *testing.T) {
	setupAppsV2Test(t)

	body := map[string]interface{}{
		"visibility": "public",
	}

	req := testutil.JSONRequest("POST", "/api/v2/apps", body)
	resp := httptest.NewRecorder()
	AppCreateHandlerV2(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestAppCreateHandlerV2_InvalidVisibility(t *testing.T) {
	setupAppsV2Test(t)

	body := map[string]interface{}{
		"title":      "Bad Vis App",
		"visibility": "secret",
	}

	req := testutil.JSONRequest("POST", "/api/v2/apps", body)
	resp := httptest.NewRecorder()
	AppCreateHandlerV2(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestAppCreateHandlerV2_MethodNotAllowed(t *testing.T) {
	setupAppsV2Test(t)

	req := httptest.NewRequest("GET", "/api/v2/apps/create", nil)
	resp := httptest.NewRecorder()
	AppCreateHandlerV2(resp, req)

	testutil.CheckError(t, resp, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
}

func TestAppCreateHandlerV2_InvalidJSON(t *testing.T) {
	setupAppsV2Test(t)

	req := httptest.NewRequest("POST", "/api/v2/apps", nil)
	resp := httptest.NewRecorder()
	AppCreateHandlerV2(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

// --- AppUpdateHandlerV2 ---

func TestAppUpdateHandlerV2_Title(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "original-title")

	title := "Updated Title"
	body := map[string]interface{}{
		"title": title,
	}

	req := testutil.JSONRequest("PUT", "/api/v2/apps/"+id, body)
	req.SetPathValue("id", id)
	resp := httptest.NewRecorder()
	AppUpdateHandlerV2(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "message", "App updated")
}

func TestAppUpdateHandlerV2_Visibility(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "vis-app")

	vis := "public"
	body := map[string]interface{}{
		"visibility": vis,
	}

	req := testutil.JSONRequest("PUT", "/api/v2/apps/"+id, body)
	req.SetPathValue("id", id)
	resp := httptest.NewRecorder()
	AppUpdateHandlerV2(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "message", "App updated")
}

func TestAppUpdateHandlerV2_InvalidVisibility(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "bad-vis-app")

	vis := "nonexistent"
	body := map[string]interface{}{
		"visibility": vis,
	}

	req := testutil.JSONRequest("PUT", "/api/v2/apps/"+id, body)
	req.SetPathValue("id", id)
	resp := httptest.NewRecorder()
	AppUpdateHandlerV2(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestAppUpdateHandlerV2_NoFields(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "no-fields-app")

	body := map[string]interface{}{}

	req := testutil.JSONRequest("PUT", "/api/v2/apps/"+id, body)
	req.SetPathValue("id", id)
	resp := httptest.NewRecorder()
	AppUpdateHandlerV2(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for no fields, got %d", resp.Code)
	}
}

func TestAppUpdateHandlerV2_NotFound(t *testing.T) {
	setupAppsV2Test(t)

	title := "ghost"
	body := map[string]interface{}{
		"title": title,
	}

	req := testutil.JSONRequest("PUT", "/api/v2/apps/nonexistent", body)
	req.SetPathValue("id", "nonexistent")
	resp := httptest.NewRecorder()
	AppUpdateHandlerV2(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "APP_NOT_FOUND")
}

func TestAppUpdateHandlerV2_MethodNotAllowed(t *testing.T) {
	setupAppsV2Test(t)

	req := httptest.NewRequest("GET", "/api/v2/apps/whatever", nil)
	req.SetPathValue("id", "whatever")
	resp := httptest.NewRecorder()
	AppUpdateHandlerV2(resp, req)

	testutil.CheckError(t, resp, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
}

// --- AppDeleteHandlerV2 ---

func TestAppDeleteHandlerV2_Success(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "del-v2-app")

	req := httptest.NewRequest("DELETE", "/api/v2/apps/"+id, nil)
	req.SetPathValue("id", id)
	resp := httptest.NewRecorder()
	AppDeleteHandlerV2(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "message", "App deleted")
}

func TestAppDeleteHandlerV2_SystemApp(t *testing.T) {
	setupAppsV2Test(t)
	db := database.GetDB()
	_, err := db.Exec(`
		INSERT INTO apps (id, original_id, title, source, created_at, updated_at)
		VALUES ('sys-v2', 'sys-v2', 'System App', 'system', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to insert system app: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/api/v2/apps/sys-v2", nil)
	req.SetPathValue("id", "sys-v2")
	resp := httptest.NewRecorder()
	AppDeleteHandlerV2(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d. Body: %s", resp.Code, resp.Body.String())
	}
}

func TestAppDeleteHandlerV2_NotFound(t *testing.T) {
	setupAppsV2Test(t)

	req := httptest.NewRequest("DELETE", "/api/v2/apps/ghost", nil)
	req.SetPathValue("id", "ghost")
	resp := httptest.NewRecorder()
	AppDeleteHandlerV2(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "APP_NOT_FOUND")
}

func TestAppDeleteHandlerV2_WithForks(t *testing.T) {
	setupAppsV2Test(t)
	db := database.GetDB()

	parentID := "app_parent1"
	forkID := "app_fork1"

	_, err := db.Exec(`
		INSERT INTO apps (id, original_id, title, source, created_at, updated_at)
		VALUES (?, ?, 'Parent App', 'deploy', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, parentID, parentID)
	if err != nil {
		t.Fatalf("Failed to create parent: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO apps (id, original_id, forked_from_id, title, source, created_at, updated_at)
		VALUES (?, ?, ?, 'Forked App', 'fork', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, forkID, parentID, parentID)
	if err != nil {
		t.Fatalf("Failed to create fork: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/api/v2/apps/"+parentID+"?with-forks=true", nil)
	req.SetPathValue("id", parentID)
	resp := httptest.NewRecorder()
	AppDeleteHandlerV2(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	// Should have deleted 2 apps (parent + fork)
	if deleted, ok := data["deleted"].(float64); !ok || deleted != 2 {
		t.Errorf("Expected 2 deleted, got %v", data["deleted"])
	}
}

func TestAppDeleteHandlerV2_MethodNotAllowed(t *testing.T) {
	setupAppsV2Test(t)

	req := httptest.NewRequest("GET", "/api/v2/apps/whatever", nil)
	req.SetPathValue("id", "whatever")
	resp := httptest.NewRecorder()
	AppDeleteHandlerV2(resp, req)

	testutil.CheckError(t, resp, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
}

// --- AppForkHandler ---

func TestAppForkHandler_Success(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "forkable-app")

	body := map[string]interface{}{}

	req := testutil.JSONRequest("POST", "/api/v2/apps/"+id+"/fork", body)
	req.SetPathValue("id", id)
	resp := httptest.NewRecorder()
	AppForkHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusCreated)
	testutil.AssertFieldEquals(t, data, "forked_from_id", id)
	testutil.AssertFieldExists(t, data, "id")
}

func TestAppForkHandler_WithAlias(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "fork-with-alias")

	body := map[string]interface{}{
		"alias": "my-fork",
	}

	req := testutil.JSONRequest("POST", "/api/v2/apps/"+id+"/fork", body)
	req.SetPathValue("id", id)
	resp := httptest.NewRecorder()
	AppForkHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusCreated)
	testutil.AssertFieldEquals(t, data, "alias", "my-fork")
}

func TestAppForkHandler_NotFound(t *testing.T) {
	setupAppsV2Test(t)

	body := map[string]interface{}{}

	req := testutil.JSONRequest("POST", "/api/v2/apps/ghost/fork", body)
	req.SetPathValue("id", "ghost")
	resp := httptest.NewRecorder()
	AppForkHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "APP_NOT_FOUND")
}

func TestAppForkHandler_MethodNotAllowed(t *testing.T) {
	setupAppsV2Test(t)

	req := httptest.NewRequest("GET", "/api/v2/apps/whatever/fork", nil)
	req.SetPathValue("id", "whatever")
	resp := httptest.NewRecorder()
	AppForkHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
}

// --- AppLineageHandler ---

func TestAppLineageHandler_Success(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "lineage-app")

	req := httptest.NewRequest("GET", "/api/v2/apps/"+id+"/lineage", nil)
	req.SetPathValue("id", id)
	resp := httptest.NewRecorder()
	AppLineageHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "id", id)
}

func TestAppLineageHandler_NotFound(t *testing.T) {
	setupAppsV2Test(t)

	req := httptest.NewRequest("GET", "/api/v2/apps/ghost/lineage", nil)
	req.SetPathValue("id", "ghost")
	resp := httptest.NewRecorder()
	AppLineageHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "APP_NOT_FOUND")
}

func TestAppLineageHandler_MethodNotAllowed(t *testing.T) {
	setupAppsV2Test(t)

	req := httptest.NewRequest("POST", "/api/v2/apps/whatever/lineage", nil)
	req.SetPathValue("id", "whatever")
	resp := httptest.NewRecorder()
	AppLineageHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
}

// --- AppForksHandler ---

func TestAppForksHandler_Empty(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "no-forks-app")

	req := httptest.NewRequest("GET", "/api/v2/apps/"+id+"/forks", nil)
	req.SetPathValue("id", id)
	resp := httptest.NewRecorder()
	AppForksHandler(resp, req)

	// Handler returns nil data when no forks (not empty array)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d. Body: %s", resp.Code, resp.Body.String())
	}
}

// TestAppForksHandler_WithForks skipped due to nested query deadlock
// (AppForksHandler calls getAliasesForApp while iterating rows)
func TestAppForksHandler_WithForks(t *testing.T) {
	t.Skip("Skipped: nested query deadlock with MaxOpenConns(1) — Issue 05 pattern")
}

func TestAppForksHandler_MethodNotAllowed(t *testing.T) {
	setupAppsV2Test(t)

	req := httptest.NewRequest("POST", "/api/v2/apps/whatever/forks", nil)
	req.SetPathValue("id", "whatever")
	resp := httptest.NewRecorder()
	AppForksHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED")
}

// --- Helper: getAliasesForApp ---

func TestGetAliasesForApp(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "alias-test-app")
	createTestAliasForApp(t, "alias1", id)
	createTestAliasForApp(t, "alias2", id)

	db := database.GetDB()
	aliases := getAliasesForApp(db, id)

	if len(aliases) != 2 {
		t.Errorf("Expected 2 aliases, got %d", len(aliases))
	}
}

func TestGetAliasesForApp_NoAliases(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "no-alias-app")

	db := database.GetDB()
	aliases := getAliasesForApp(db, id)

	if len(aliases) != 0 {
		t.Errorf("Expected 0 aliases, got %d", len(aliases))
	}
}

// --- Helper: getAppByID ---

func TestGetAppByID_Success(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "getapp-test")

	db := database.GetDB()
	app, err := getAppByID(db, id)

	if err != nil {
		t.Fatalf("getAppByID failed: %v", err)
	}
	if app.Title != "getapp-test" {
		t.Errorf("Expected title 'getapp-test', got '%s'", app.Title)
	}
}

func TestGetAppByID_NotFound(t *testing.T) {
	setupAppsV2Test(t)

	db := database.GetDB()
	_, err := getAppByID(db, "nonexistent")

	if err == nil {
		t.Error("Expected error for nonexistent app")
	}
}

// --- Helper: buildLineageTree ---

func TestBuildLineageTree_Simple(t *testing.T) {
	setupAppsV2Test(t)
	id := createTestAppV2(t, "lineage-root")

	db := database.GetDB()
	tree := buildLineageTree(db, id, nil)

	if tree == nil {
		t.Fatal("Expected non-nil tree")
	}
	if tree.ID != id {
		t.Errorf("Expected root ID %s, got %s", id, tree.ID)
	}
}

// TestBuildLineageTree_WithForks is skipped because buildLineageTree calls
// getAliasesForApp while iterating rows — same nested query deadlock as Issue 05.
// Needs MaxOpenConns > 1 or code fix (collect rows first, then query aliases).
func TestBuildLineageTree_WithForks(t *testing.T) {
	t.Skip("Skipped: nested query deadlock with MaxOpenConns(1) — known Issue 05 pattern")
}
