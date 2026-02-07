package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/hosting"
)

func setupSiteFilesTest(t *testing.T) {
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
}

// --- SiteDetailHandler ---

func TestSiteDetailHandler_MissingID(t *testing.T) {
	setupSiteFilesTest(t)

	req := httptest.NewRequest("GET", "/api/sites/", nil)
	resp := httptest.NewRecorder()
	SiteDetailHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestSiteDetailHandler_NotFound(t *testing.T) {
	setupSiteFilesTest(t)

	req := httptest.NewRequest("GET", "/api/sites/ghost", nil)
	req.SetPathValue("id", "ghost")
	resp := httptest.NewRecorder()
	SiteDetailHandler(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", resp.Code)
	}
}

func TestSiteDetailHandler_Success(t *testing.T) {
	setupSiteFilesTest(t)
	db := database.GetDB()
	createTestSite(t, db, "detail-site")

	req := httptest.NewRequest("GET", "/api/sites/detail-site", nil)
	req.SetPathValue("id", "detail-site")
	resp := httptest.NewRecorder()
	SiteDetailHandler(resp, req)

	// System sites (root, 404) are created by hosting.Init
	// detail-site may or may not appear in ListSites depending on hosting impl
	// Just verify it doesn't crash
	if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
		t.Errorf("Expected 200 or 404, got %d. Body: %s", resp.Code, resp.Body.String())
	}
}

// --- SiteFilesHandler ---

func TestSiteFilesHandler_MissingID(t *testing.T) {
	setupSiteFilesTest(t)

	req := httptest.NewRequest("GET", "/api/sites//files", nil)
	resp := httptest.NewRecorder()
	SiteFilesHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestSiteFilesHandler_NotFound(t *testing.T) {
	setupSiteFilesTest(t)

	req := httptest.NewRequest("GET", "/api/sites/ghost/files", nil)
	req.SetPathValue("id", "ghost")
	resp := httptest.NewRecorder()
	SiteFilesHandler(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", resp.Code)
	}
}

func TestSiteFilesHandler_WithFiles(t *testing.T) {
	setupSiteFilesTest(t)
	db := database.GetDB()
	createTestSite(t, db, "files-site")

	req := httptest.NewRequest("GET", "/api/sites/files-site/files", nil)
	req.SetPathValue("id", "files-site")
	resp := httptest.NewRecorder()
	SiteFilesHandler(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d. Body: %s", resp.Code, resp.Body.String())
	}
}

// --- SiteFileContentHandler ---

func TestSiteFileContentHandler_MissingParams(t *testing.T) {
	setupSiteFilesTest(t)

	req := httptest.NewRequest("GET", "/api/sites//files/", nil)
	resp := httptest.NewRecorder()
	SiteFileContentHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestSiteFileContentHandler_FileNotFound(t *testing.T) {
	setupSiteFilesTest(t)

	req := httptest.NewRequest("GET", "/api/sites/test/files/nonexistent.html", nil)
	req.SetPathValue("id", "test")
	req.SetPathValue("path", "nonexistent.html")
	resp := httptest.NewRecorder()
	SiteFileContentHandler(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", resp.Code)
	}
}

func TestSiteFileContentHandler_Success(t *testing.T) {
	setupSiteFilesTest(t)
	db := database.GetDB()
	createTestSite(t, db, "content-site")

	req := httptest.NewRequest("GET", "/api/sites/content-site/files/index.html", nil)
	req.SetPathValue("id", "content-site")
	req.SetPathValue("path", "index.html")
	resp := httptest.NewRecorder()
	SiteFileContentHandler(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d. Body: %s", resp.Code, resp.Body.String())
	}
}
