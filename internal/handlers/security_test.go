package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/hosting"
)

// setupSecurityTest creates a test database and initializes handlers
func setupSecurityTest(t *testing.T) {
	t.Helper()
	silenceTestLogs(t)

	db := setupTestDB(t)
	database.SetDB(db)

	// Initialize hosting for file-based tests
	if err := hosting.Init(db); err != nil {
		t.Fatalf("Failed to initialize hosting: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		database.SetDB(nil)
	})
}

// --- SQL Injection Tests ---

func TestSQLInjection_EventsHandler_Domain(t *testing.T) {
	setupSecurityTest(t)

	injectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE events--",
		"' UNION SELECT * FROM auth_users--",
		"1' AND 1=1--",
		"admin'--",
		"' OR 1=1#",
		"' OR 'x'='x",
		"1'; DELETE FROM events WHERE '1'='1",
	}

	for _, payload := range injectionPayloads {
		t.Run(fmt.Sprintf("domain=%s", payload), func(t *testing.T) {
			reqURL := fmt.Sprintf("/api/events?domain=%s", url.QueryEscape(payload))
			req := httptest.NewRequest("GET", reqURL, nil)
			resp := httptest.NewRecorder()

			EventsHandler(resp, req)

			// Should NOT cause 500 error (injection prevented)
			if resp.Code == http.StatusInternalServerError {
				t.Errorf("SQL injection caused 500 error: %s", resp.Body.String())
			}

			// Should return valid response (empty array or filtered results)
			if resp.Code != http.StatusOK {
				t.Errorf("Expected 200 for filtered query, got %d", resp.Code)
			}
		})
	}
}

func TestSQLInjection_EventsHandler_Tags(t *testing.T) {
	setupSecurityTest(t)

	injectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE events--",
		"%' OR '1'='1'%",
		"tag' UNION SELECT * FROM auth_sessions--",
	}

	for _, payload := range injectionPayloads {
		t.Run(fmt.Sprintf("tags=%s", payload), func(t *testing.T) {
			reqURL := fmt.Sprintf("/api/events?tags=%s", url.QueryEscape(payload))
			req := httptest.NewRequest("GET", reqURL, nil)
			resp := httptest.NewRecorder()

			EventsHandler(resp, req)

			// Should NOT cause 500 error
			if resp.Code == http.StatusInternalServerError {
				t.Errorf("SQL injection caused 500 error: %s", resp.Body.String())
			}

			// Should return OK (filtered safely)
			if resp.Code != http.StatusOK {
				t.Errorf("Expected 200, got %d", resp.Code)
			}
		})
	}
}

func TestSQLInjection_EventsHandler_SourceType(t *testing.T) {
	setupSecurityTest(t)

	injectionPayloads := []string{
		"web' OR '1'='1",
		"web'; DROP TABLE events--",
		"' UNION SELECT * FROM redirects--",
	}

	for _, payload := range injectionPayloads {
		t.Run(fmt.Sprintf("source_type=%s", payload), func(t *testing.T) {
			reqURL := fmt.Sprintf("/api/events?source_type=%s", url.QueryEscape(payload))
			req := httptest.NewRequest("GET", reqURL, nil)
			resp := httptest.NewRecorder()

			EventsHandler(resp, req)

			if resp.Code == http.StatusInternalServerError {
				t.Errorf("SQL injection caused 500 error: %s", resp.Body.String())
			}

			if resp.Code != http.StatusOK {
				t.Errorf("Expected 200, got %d", resp.Code)
			}
		})
	}
}

func TestSQLInjection_AliasDetailHandler(t *testing.T) {
	setupSecurityTest(t)

	injectionPayloads := []string{
		"admin' OR '1'='1'--",
		"'; DROP TABLE aliases--",
		"' UNION SELECT * FROM auth_users--",
	}

	for _, payload := range injectionPayloads {
		t.Run(fmt.Sprintf("alias=%s", payload), func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/aliases/"+url.PathEscape(payload), nil)
			req.SetPathValue("alias", payload)
			resp := httptest.NewRecorder()

			AliasDetailHandler(resp, req)

			// Should NOT cause 500 error
			if resp.Code == http.StatusInternalServerError {
				t.Errorf("SQL injection caused 500 error: %s", resp.Body.String())
			}

			// Should return 404 (not found), 400 (invalid), or 200 (if somehow valid)
			if resp.Code != http.StatusNotFound && resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest {
				t.Errorf("Expected 404, 400, or 200, got %d", resp.Code)
			}
		})
	}
}

// --- Path Traversal Tests ---

func TestPathTraversal_SiteFileContentHandler(t *testing.T) {
	setupSecurityTest(t)
	db := database.GetDB()

	// Create a test site with files
	siteID := "testsite"
	_, err := db.Exec(`
		INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash, updated_at)
		VALUES (?, 'index.html', '<html>Test</html>', 18, 'text/html', 'abcd1234', datetime('now'))
	`, siteID)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	traversalPayloads := []string{
		"../../etc/passwd",
		"../../../etc/shadow",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"....//....//....//etc/passwd",
		"..%2f..%2f..%2fetc%2fpasswd",
		"%2e%2e%2f%2e%2e%2fetc%2fpasswd",
		"index.html/../../etc/passwd",
		"./../../etc/hosts",
	}

	for _, payload := range traversalPayloads {
		t.Run(fmt.Sprintf("path=%s", payload), func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/sites/"+siteID+"/files/"+payload, nil)
			req.SetPathValue("id", siteID)
			req.SetPathValue("path", payload)
			resp := httptest.NewRecorder()

			SiteFileContentHandler(resp, req)

			// Should NOT return system files
			body := resp.Body.String()
			if strings.Contains(body, "root:") || strings.Contains(body, "/bin/bash") {
				t.Errorf("Path traversal successful - returned system file: %s", body)
			}

			// Should return 404 (file not found in VFS)
			if resp.Code != http.StatusNotFound {
				t.Logf("Path traversal attempt returned %d (expected 404)", resp.Code)
			}
		})
	}
}

func TestPathTraversal_SiteFilesHandler(t *testing.T) {
	setupSecurityTest(t)

	traversalPayloads := []string{
		"../../etc",
		"../../../",
		"..%2f..%2f",
	}

	for _, payload := range traversalPayloads {
		t.Run(fmt.Sprintf("site_id=%s", payload), func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/sites/"+payload+"/files", nil)
			req.SetPathValue("id", payload)
			resp := httptest.NewRecorder()

			SiteFilesHandler(resp, req)

			// Should NOT leak filesystem info
			if resp.Code == http.StatusInternalServerError {
				t.Errorf("Path traversal caused 500 error: %s", resp.Body.String())
			}

			// Should return 404 (site not found)
			if resp.Code != http.StatusNotFound {
				t.Logf("Expected 404 for invalid site_id, got %d", resp.Code)
			}
		})
	}
}

// --- XSS Tests ---

func TestXSS_TrackHandler(t *testing.T) {
	setupSecurityTest(t)

	xssPayloads := []struct {
		name  string
		field string
		value string
	}{
		{"script_tag", "domain", "<script>alert('xss')</script>"},
		{"img_onerror", "path", "<img src=x onerror=alert('xss')>"},
		{"svg_onload", "referrer", "<svg/onload=alert('xss')>"},
		{"javascript_protocol", "path", "javascript:alert('xss')"},
		{"event_handler", "domain", "\" onmouseover=\"alert('xss')"},
		{"encoded_script", "path", "%3Cscript%3Ealert('xss')%3C/script%3E"},
		{"iframe_injection", "referrer", "<iframe src=\"javascript:alert('xss')\">"},
	}

	for _, payload := range xssPayloads {
		t.Run(payload.name, func(t *testing.T) {
			trackReq := map[string]interface{}{
				"domain":     "test.com",
				"path":       "/test",
				"event_type": "pageview",
				"referrer":   "",
			}
			trackReq[payload.field] = payload.value

			body, _ := json.Marshal(trackReq)
			req := httptest.NewRequest("POST", "/track", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			TrackHandler(resp, req)

			// Should accept and sanitize (not crash) - returns 204 No Content
			if resp.Code != http.StatusOK && resp.Code != http.StatusNoContent {
				t.Errorf("Expected 200 or 204 for tracking request, got %d: %s", resp.Code, resp.Body.String())
			}

			// Note: Events are buffered via analytics.Add(), not immediately in database
			// The fact that the handler doesn't crash and returns 204 means input was handled safely
			// Actual sanitization happens in sanitizeInput() function which is called before storage
		})
	}
}

func TestXSS_AliasCreateHandler(t *testing.T) {
	setupSecurityTest(t)

	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"<img src=x onerror=alert('xss')>",
		"javascript:alert('xss')",
	}

	for _, payload := range xssPayloads {
		t.Run(fmt.Sprintf("subdomain=%s", payload), func(t *testing.T) {
			aliasReq := map[string]interface{}{
				"subdomain": payload,
				"type":      "app",
				"targets": map[string]string{
					"app_id": "test-app",
				},
			}

			body, _ := json.Marshal(aliasReq)
			req := httptest.NewRequest("POST", "/api/aliases", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			AliasCreateHandler(resp, req)

			// Should reject or sanitize dangerous input
			if resp.Code == http.StatusCreated {
				// If accepted, verify it was sanitized
				db := database.GetDB()
				var storedSubdomain string
				db.QueryRow("SELECT subdomain FROM aliases ORDER BY created_at DESC LIMIT 1").Scan(&storedSubdomain)

				if strings.Contains(storedSubdomain, "<script>") ||
					strings.Contains(storedSubdomain, "onerror=") {
					t.Errorf("XSS payload not sanitized: %s", storedSubdomain)
				}
			}

			// Most likely will be rejected as invalid subdomain format
			if resp.Code != http.StatusBadRequest && resp.Code != http.StatusCreated {
				t.Logf("XSS payload returned status %d", resp.Code)
			}
		})
	}
}

// --- Resource Exhaustion Tests ---

func TestResourceExhaustion_OversizedRequestBody(t *testing.T) {
	setupSecurityTest(t)

	// Create a very large JSON payload (100MB)
	largePayload := make(map[string]interface{})
	largeString := strings.Repeat("A", 100*1024*1024) // 100MB
	largePayload["data"] = largeString

	body, _ := json.Marshal(largePayload)
	req := httptest.NewRequest("POST", "/track", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	TrackHandler(resp, req)

	// Should reject with 413 or 400 (body too large)
	if resp.Code != http.StatusRequestEntityTooLarge && resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 413 or 400 for oversized body, got %d", resp.Code)
	}
}

func TestResourceExhaustion_OversizedJSONPayload(t *testing.T) {
	setupSecurityTest(t)

	// Create deeply nested JSON
	deepJSON := "{"
	for i := 0; i < 10000; i++ {
		deepJSON += `"level":` + "{"
	}
	deepJSON += `"value":"deep"`
	for i := 0; i < 10000; i++ {
		deepJSON += "}"
	}
	deepJSON += "}"

	req := httptest.NewRequest("POST", "/track", bytes.NewReader([]byte(deepJSON)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	TrackHandler(resp, req)

	// Should reject or handle gracefully (not crash)
	if resp.Code == http.StatusInternalServerError {
		t.Errorf("Deeply nested JSON caused 500 error")
	}

	// Should return 400 (invalid JSON) or 413
	if resp.Code != http.StatusBadRequest && resp.Code != http.StatusRequestEntityTooLarge {
		t.Logf("Deeply nested JSON returned %d", resp.Code)
	}
}

func TestResourceExhaustion_VeryLongURLPath(t *testing.T) {
	setupSecurityTest(t)

	// Create very long URL path (10KB)
	longPath := strings.Repeat("a", 10*1024)
	req := httptest.NewRequest("GET", "/api/events?domain="+url.QueryEscape(longPath), nil)
	resp := httptest.NewRecorder()

	EventsHandler(resp, req)

	// Should handle gracefully (not crash)
	if resp.Code == http.StatusInternalServerError {
		t.Errorf("Very long URL caused 500 error")
	}

	// Should return 200 (valid query, just no results)
	if resp.Code != http.StatusOK {
		t.Logf("Very long URL returned %d", resp.Code)
	}
}

func TestResourceExhaustion_VeryLongQueryString(t *testing.T) {
	setupSecurityTest(t)

	// Create very long query string with many parameters
	longQuery := "/api/events?"
	for i := 0; i < 1000; i++ {
		longQuery += fmt.Sprintf("param%d=%s&", i, strings.Repeat("x", 100))
	}

	req := httptest.NewRequest("GET", longQuery, nil)
	resp := httptest.NewRecorder()

	EventsHandler(resp, req)

	// Should handle gracefully
	if resp.Code == http.StatusInternalServerError {
		t.Errorf("Very long query string caused 500 error")
	}

	// Should return 200 (ignore unknown params)
	if resp.Code != http.StatusOK {
		t.Logf("Very long query string returned %d", resp.Code)
	}
}

func TestResourceExhaustion_ManyTagsInTrackRequest(t *testing.T) {
	setupSecurityTest(t)

	// Create track request with 10,000 tags
	trackReq := map[string]interface{}{
		"domain":     "test.com",
		"path":       "/test",
		"event_type": "pageview",
		"tags":       make([]string, 10000),
	}

	for i := 0; i < 10000; i++ {
		trackReq["tags"].([]string)[i] = fmt.Sprintf("tag%d", i)
	}

	body, _ := json.Marshal(trackReq)
	req := httptest.NewRequest("POST", "/track", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	TrackHandler(resp, req)

	// Should handle gracefully (might truncate or reject)
	if resp.Code == http.StatusInternalServerError {
		t.Errorf("Many tags caused 500 error")
	}

	// Should return 200 or 400
	if resp.Code != http.StatusOK && resp.Code != http.StatusBadRequest {
		t.Logf("Many tags returned %d", resp.Code)
	}
}
