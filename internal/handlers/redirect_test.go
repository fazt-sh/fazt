package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/database"
)

func setupRedirectTest(t *testing.T) {
	t.Helper()
	silenceTestLogs(t)

	db := setupTestDB(t)
	database.SetDB(db)
	t.Cleanup(func() {
		db.Close()
		database.SetDB(nil)
	})
}

// --- RedirectHandler ---

func TestRedirectHandler_Success(t *testing.T) {
	setupRedirectTest(t)
	db := database.GetDB()
	createTestRedirect(t, db, "gh", "https://github.com/test")

	req := httptest.NewRequest("GET", "/r/gh", nil)
	resp := httptest.NewRecorder()
	RedirectHandler(resp, req)

	if resp.Code != http.StatusFound {
		t.Errorf("Expected 302, got %d", resp.Code)
	}
	loc := resp.Header().Get("Location")
	if loc != "https://github.com/test" {
		t.Errorf("Expected redirect to https://github.com/test, got %s", loc)
	}
}

func TestRedirectHandler_NotFound(t *testing.T) {
	setupRedirectTest(t)

	req := httptest.NewRequest("GET", "/r/nonexistent", nil)
	resp := httptest.NewRecorder()
	RedirectHandler(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", resp.Code)
	}
}

func TestRedirectHandler_EmptySlug(t *testing.T) {
	setupRedirectTest(t)

	req := httptest.NewRequest("GET", "/r/", nil)
	resp := httptest.NewRecorder()
	RedirectHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestRedirectHandler_WithExtraTags(t *testing.T) {
	setupRedirectTest(t)
	db := database.GetDB()
	createTestRedirect(t, db, "tagged", "https://example.com")

	req := httptest.NewRequest("GET", "/r/tagged?tags=email,campaign", nil)
	resp := httptest.NewRecorder()
	RedirectHandler(resp, req)

	if resp.Code != http.StatusFound {
		t.Errorf("Expected 302, got %d", resp.Code)
	}
}

func TestRedirectHandler_ClickCountIncrement(t *testing.T) {
	setupRedirectTest(t)
	db := database.GetDB()
	createTestRedirect(t, db, "counter", "https://example.com")

	// First click
	req := httptest.NewRequest("GET", "/r/counter", nil)
	resp := httptest.NewRecorder()
	RedirectHandler(resp, req)

	if resp.Code != http.StatusFound {
		t.Fatalf("Expected 302, got %d", resp.Code)
	}

	// Second click
	req = httptest.NewRequest("GET", "/r/counter", nil)
	resp = httptest.NewRecorder()
	RedirectHandler(resp, req)

	// Check click count in DB
	var count int
	err := db.QueryRow("SELECT click_count FROM redirects WHERE slug = ?", "counter").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query click count: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected click_count=2, got %d", count)
	}
}
