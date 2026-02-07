package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/fazt-sh/fazt/internal/hosting"
)

// --- Helpers ---

// buildTestZip creates a ZIP file in memory from a map of path->content
func buildTestZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("Failed to create zip entry %s: %v", name, err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatalf("Failed to write zip entry %s: %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}
	return buf.Bytes()
}

// withHeader sets an arbitrary request header
func withHeader(key, value string) requestOption {
	return func(r *http.Request) {
		r.Header.Set(key, value)
	}
}

// deployFile inserts a single file into the VFS directly
func (s *integrationTestServer) deployFile(t *testing.T, siteID, path, content, mimeType string) {
	t.Helper()
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	_, err := s.db.Exec(`
		INSERT INTO files (site_id, path, content, size_bytes, mime_type, hash, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
	`, siteID, path, []byte(content), len(content), mimeType, hash)
	if err != nil {
		t.Fatalf("Failed to deploy file %s/%s: %v", siteID, path, err)
	}
}

// --- TestDeployToServing ---
// Tests the full deploy pipeline: ZIP -> extract -> VFS -> alias creation -> static file serving
// Touches: hosting, VFS, aliases, file system, routing

func TestDeployToServing(t *testing.T) {
	s := setupIntegrationTest(t)

	// Build a test ZIP with multiple file types
	zipData := buildTestZip(t, map[string]string{
		"index.html":            `<!DOCTYPE html><html><head><title>Deploy Test</title></head><body><h1>Hello Deploy</h1></body></html>`,
		"css/style.css":         `body { color: red; font-size: 16px; }`,
		"js/app.js":             `console.log("deploy test");`,
		"assets/main-abc123.js": `var x = "hashed asset content";`,
		"about/index.html":      `<!DOCTYPE html><html><body><h1>About Page</h1></body></html>`,
		"manifest.json":         `{"name":"deploy-test"}`,
	})

	// Deploy via hosting.DeploySiteWithSource (the production deploy function)
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Fatalf("Failed to create zip reader: %v", err)
	}

	result, err := hosting.DeploySiteWithSource(zipReader, "deploy-test", &hosting.SourceInfo{
		Type: "deploy",
	})
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	// --- Verify deploy result ---

	if result.FileCount != 6 {
		t.Errorf("Expected 6 deployed files, got %d", result.FileCount)
	}
	if result.SiteID != "deploy-test" {
		t.Errorf("Expected SiteID 'deploy-test', got %q", result.SiteID)
	}
	if result.SizeBytes <= 0 {
		t.Error("Expected positive SizeBytes")
	}

	// --- Verify DB records ---

	t.Run("app record created in database", func(t *testing.T) {
		var appID, title, source string
		err := s.db.QueryRow("SELECT id, title, source FROM apps WHERE title = ?", "deploy-test").
			Scan(&appID, &title, &source)
		if err != nil {
			t.Fatalf("App record not found: %v", err)
		}
		if !strings.HasPrefix(appID, "fazt_app_") {
			t.Errorf("Expected app ID prefix 'fazt_app_', got %q", appID)
		}
		if title != "deploy-test" {
			t.Errorf("Expected title 'deploy-test', got %q", title)
		}
		if source != "deploy" {
			t.Errorf("Expected source 'deploy', got %q", source)
		}
	})

	t.Run("alias auto-created for subdomain", func(t *testing.T) {
		var aliasType, targets string
		err := s.db.QueryRow("SELECT type, targets FROM aliases WHERE subdomain = ?", "deploy-test").
			Scan(&aliasType, &targets)
		if err != nil {
			t.Fatalf("Alias not found: %v", err)
		}
		if aliasType != "app" {
			t.Errorf("Expected alias type 'app', got %q", aliasType)
		}
		// Verify targets JSON contains the app_id
		var appID string
		s.db.QueryRow("SELECT id FROM apps WHERE title = ?", "deploy-test").Scan(&appID)
		if !strings.Contains(targets, appID) {
			t.Errorf("Alias targets %q should contain app_id %q", targets, appID)
		}
	})

	t.Run("files stored in VFS with correct hashes", func(t *testing.T) {
		var count int
		s.db.QueryRow("SELECT COUNT(*) FROM files WHERE site_id = ?", "deploy-test").Scan(&count)
		if count != 6 {
			t.Errorf("Expected 6 files in VFS, got %d", count)
		}

		// Verify a specific file hash matches expected SHA-256
		cssContent := `body { color: red; font-size: 16px; }`
		expectedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(cssContent)))
		var actualHash string
		err := s.db.QueryRow("SELECT hash FROM files WHERE site_id = ? AND path = ?",
			"deploy-test", "css/style.css").Scan(&actualHash)
		if err != nil {
			t.Fatalf("CSS file not found in VFS: %v", err)
		}
		if actualHash != expectedHash {
			t.Errorf("Hash mismatch for css/style.css:\n  got  %q\n  want %q", actualHash, expectedHash)
		}
	})

	// --- Verify HTTP serving ---

	t.Run("serve index.html at root", func(t *testing.T) {
		resp := s.makeRequest(t, "GET", "/", nil,
			withHost("deploy-test.testdomain.com"),
		)
		body := readBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d. Body: %s", resp.StatusCode, body)
		}
		assertContains(t, body, "Hello Deploy")

		// Content-Type
		ct := resp.Header.Get("Content-Type")
		if !strings.Contains(ct, "text/html") {
			t.Errorf("Expected text/html Content-Type, got %q", ct)
		}

		// Cache-Control for HTML: no-cache
		cc := resp.Header.Get("Cache-Control")
		if !strings.Contains(cc, "no-cache") {
			t.Errorf("Expected no-cache for HTML, got %q", cc)
		}

		// ETag present
		if resp.Header.Get("ETag") == "" {
			t.Error("Expected ETag header to be set")
		}
	})

	t.Run("serve CSS with correct MIME and caching", func(t *testing.T) {
		resp := s.makeRequest(t, "GET", "/css/style.css", nil,
			withHost("deploy-test.testdomain.com"),
		)
		body := readBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		assertContains(t, body, "color: red")

		ct := resp.Header.Get("Content-Type")
		if !strings.Contains(ct, "text/css") {
			t.Errorf("Expected text/css, got %q", ct)
		}

		// Non-hashed asset: 5 minute cache
		cc := resp.Header.Get("Cache-Control")
		if !strings.Contains(cc, "max-age=300") {
			t.Errorf("Expected max-age=300 for regular CSS, got %q", cc)
		}
	})

	t.Run("serve JS with correct MIME", func(t *testing.T) {
		resp := s.makeRequest(t, "GET", "/js/app.js", nil,
			withHost("deploy-test.testdomain.com"),
		)
		body := readBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		assertContains(t, body, `console.log("deploy test")`)

		ct := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "application/javascript") && !strings.HasPrefix(ct, "text/javascript") {
			t.Errorf("Expected JavaScript MIME type, got %q", ct)
		}
	})

	t.Run("serve hashed asset with immutable cache", func(t *testing.T) {
		resp := s.makeRequest(t, "GET", "/assets/main-abc123.js", nil,
			withHost("deploy-test.testdomain.com"),
		)
		body := readBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		assertContains(t, body, "hashed asset content")

		cc := resp.Header.Get("Cache-Control")
		if !strings.Contains(cc, "immutable") {
			t.Errorf("Expected immutable for hashed asset, got %q", cc)
		}
		if !strings.Contains(cc, "max-age=31536000") {
			t.Errorf("Expected max-age=31536000 for hashed asset, got %q", cc)
		}
	})

	t.Run("directory index fallback serves about/index.html", func(t *testing.T) {
		resp := s.makeRequest(t, "GET", "/about", nil,
			withHost("deploy-test.testdomain.com"),
		)
		body := readBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 for /about (directory index), got %d. Body: %s",
				resp.StatusCode, body)
		}
		assertContains(t, body, "About Page")
	})

	t.Run("ETag enables 304 Not Modified", func(t *testing.T) {
		// First request to get ETag
		resp1 := s.makeRequest(t, "GET", "/css/style.css", nil,
			withHost("deploy-test.testdomain.com"),
		)
		etag := resp1.Header.Get("ETag")
		resp1.Body.Close()

		if etag == "" {
			t.Fatal("No ETag in first response")
		}

		// Second request with If-None-Match
		resp2 := s.makeRequest(t, "GET", "/css/style.css", nil,
			withHost("deploy-test.testdomain.com"),
			withHeader("If-None-Match", etag),
		)
		defer resp2.Body.Close()

		if resp2.StatusCode != http.StatusNotModified {
			t.Fatalf("Expected 304 Not Modified, got %d", resp2.StatusCode)
		}
	})

	t.Run("non-existent file returns 404", func(t *testing.T) {
		resp := s.makeRequest(t, "GET", "/does-not-exist.txt", nil,
			withHost("deploy-test.testdomain.com"),
		)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("redeploy cleans old files and serves new content", func(t *testing.T) {
		// Deploy v2 with only 2 files (old files should be cleaned)
		zipData2 := buildTestZip(t, map[string]string{
			"index.html":    `<!DOCTYPE html><html><body><h1>Version Two</h1></body></html>`,
			"manifest.json": `{"name":"deploy-test"}`,
		})

		zipReader2, err := zip.NewReader(bytes.NewReader(zipData2), int64(len(zipData2)))
		if err != nil {
			t.Fatalf("Failed to create v2 zip reader: %v", err)
		}

		result2, err := hosting.DeploySiteWithSource(zipReader2, "deploy-test", nil)
		if err != nil {
			t.Fatalf("Redeploy failed: %v", err)
		}
		if result2.FileCount != 2 {
			t.Errorf("Expected 2 files in v2, got %d", result2.FileCount)
		}

		// Verify old CSS file is gone (clean deploy)
		resp := s.makeRequest(t, "GET", "/css/style.css", nil,
			withHost("deploy-test.testdomain.com"),
		)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 for removed file after redeploy, got %d", resp.StatusCode)
		}

		// Verify new content is served
		resp2 := s.makeRequest(t, "GET", "/", nil,
			withHost("deploy-test.testdomain.com"),
		)
		body := readBody(t, resp2)
		assertContains(t, body, "Version Two")
	})
}

// --- TestAliasToAppResolution ---
// Tests how aliases map to apps, how site_id relates to subdomains,
// and VFS serving mechanics including SPA, private files, and path handling.

func TestAliasToAppResolution(t *testing.T) {
	s := setupIntegrationTest(t)

	// Set up two independent sites with different content
	s.createTestApp(t, "app_alpha", "Alpha Site")
	s.createTestAlias(t, "alpha", "app_alpha", "app")
	s.deployFiles(t, "alpha", "Alpha Site")
	s.deployFile(t, "alpha", "css/alpha.css", `body { color: blue; }`, "text/css")

	s.createTestApp(t, "app_beta", "Beta Site")
	s.createTestAlias(t, "beta", "app_beta", "app")
	s.deployFiles(t, "beta", "Beta Site")
	s.deployFile(t, "beta", "css/beta.css", `body { color: green; }`, "text/css")

	t.Run("multiple subdomains serve isolated content", func(t *testing.T) {
		// Alpha serves alpha content
		resp := s.makeRequest(t, "GET", "/", nil,
			withHost("alpha.testdomain.com"),
		)
		body := readBody(t, resp)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 for alpha, got %d. Body: %s", resp.StatusCode, body)
		}
		assertContains(t, body, "Alpha Site")

		// Beta serves beta content
		resp2 := s.makeRequest(t, "GET", "/", nil,
			withHost("beta.testdomain.com"),
		)
		body2 := readBody(t, resp2)
		if resp2.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 for beta, got %d. Body: %s", resp2.StatusCode, body2)
		}
		assertContains(t, body2, "Beta Site")

		// Alpha CSS only available on alpha subdomain
		resp3 := s.makeRequest(t, "GET", "/css/alpha.css", nil,
			withHost("alpha.testdomain.com"),
		)
		body3 := readBody(t, resp3)
		if resp3.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 for alpha CSS, got %d", resp3.StatusCode)
		}
		assertContains(t, body3, "color: blue")

		// Alpha CSS NOT available on beta subdomain
		resp4 := s.makeRequest(t, "GET", "/css/alpha.css", nil,
			withHost("beta.testdomain.com"),
		)
		defer resp4.Body.Close()
		if resp4.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 for alpha CSS on beta, got %d", resp4.StatusCode)
		}
	})

	t.Run("files stored by site_id (subdomain) not app_id", func(t *testing.T) {
		// This is a critical architectural invariant:
		// Files are keyed by subdomain in the VFS, not by app_id.
		// The app_id from alias resolution is only used for analytics/identity.
		var countBySubdomain int
		s.db.QueryRow("SELECT COUNT(*) FROM files WHERE site_id = ?", "alpha").Scan(&countBySubdomain)
		if countBySubdomain == 0 {
			t.Error("Expected files stored under site_id='alpha' (the subdomain)")
		}

		var countByAppID int
		s.db.QueryRow("SELECT COUNT(*) FROM files WHERE site_id = ?", "app_alpha").Scan(&countByAppID)
		if countByAppID != 0 {
			t.Error("Files should NOT be stored under site_id='app_alpha' (the app ID)")
		}
	})

	t.Run("trailing slash redirects to non-trailing path", func(t *testing.T) {
		// Deploy about/index.html for directory index
		s.deployFile(t, "alpha", "about/index.html",
			`<!DOCTYPE html><html><body>Alpha About</body></html>`, "text/html")

		// Use no-redirect client to capture the 301
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		req, _ := http.NewRequest("GET", s.server.URL+"/about/", nil)
		req.Host = "alpha.testdomain.com"

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMovedPermanently {
			t.Fatalf("Expected 301 for trailing slash, got %d", resp.StatusCode)
		}

		location := resp.Header.Get("Location")
		if location != "/about" {
			t.Errorf("Expected redirect to /about, got %q", location)
		}
	})

	t.Run("SPA fallback serves index.html for route-like paths", func(t *testing.T) {
		// Create a separate SPA site
		// Title must match subdomain for GetAppSPA to find it
		s.createTestApp(t, "app_spa", "spa-app")
		s.createTestAlias(t, "spa-app", "app_spa", "app")
		s.deployFile(t, "spa-app", "index.html",
			`<!DOCTYPE html><html><body><div id="app">SPA Root</div></body></html>`, "text/html")
		s.deployFile(t, "spa-app", "manifest.json", `{"name":"spa-app"}`, "application/json")
		s.deployFile(t, "spa-app", "css/style.css", `body { margin: 0; }`, "text/css")

		// Enable SPA routing on this app
		sqlFS, ok := hosting.GetFileSystem().(*hosting.SQLFileSystem)
		if !ok {
			t.Fatal("Expected SQLFileSystem")
		}
		if err := sqlFS.SetAppSPA("spa-app", true); err != nil {
			t.Fatalf("Failed to enable SPA: %v", err)
		}

		// Route-like path (no extension) should fall back to index.html
		resp := s.makeRequest(t, "GET", "/dashboard", nil,
			withHost("spa-app.testdomain.com"),
		)
		body := readBody(t, resp)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 for SPA route /dashboard, got %d. Body: %s",
				resp.StatusCode, body)
		}
		assertContains(t, body, "SPA Root")

		// Nested route also falls back to index.html
		resp2 := s.makeRequest(t, "GET", "/settings/profile", nil,
			withHost("spa-app.testdomain.com"),
		)
		body2 := readBody(t, resp2)
		if resp2.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 for SPA route /settings/profile, got %d", resp2.StatusCode)
		}
		assertContains(t, body2, "SPA Root")

		// Real files still served directly (not SPA fallback)
		resp3 := s.makeRequest(t, "GET", "/css/style.css", nil,
			withHost("spa-app.testdomain.com"),
		)
		body3 := readBody(t, resp3)
		if resp3.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 for real CSS file under SPA site, got %d", resp3.StatusCode)
		}
		assertContains(t, body3, "margin: 0")

		// File with extension that doesn't exist should NOT fall back (SPA only for route-like paths)
		resp4 := s.makeRequest(t, "GET", "/missing.js", nil,
			withHost("spa-app.testdomain.com"),
		)
		defer resp4.Body.Close()
		if resp4.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 for missing .js file (SPA should not catch file extensions), got %d",
				resp4.StatusCode)
		}
	})

	t.Run("private files require authentication", func(t *testing.T) {
		// Deploy a private file
		s.deployFile(t, "alpha", "private/secret.txt", "top secret data", "text/plain")

		// Without auth -> 401
		resp := s.makeRequest(t, "GET", "/private/secret.txt", nil,
			withHost("alpha.testdomain.com"),
		)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("Expected 401 for private file without auth, got %d", resp.StatusCode)
		}

		// With valid session -> 200
		sessionID := s.createSession(t, "private-access@test.com", "user")
		resp2 := s.makeRequest(t, "GET", "/private/secret.txt", nil,
			withHost("alpha.testdomain.com"),
			withSession(sessionID),
		)
		body := readBody(t, resp2)
		if resp2.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 for private file with auth, got %d. Body: %s",
				resp2.StatusCode, body)
		}
		assertContains(t, body, "top secret data")
	})

	t.Run("API path without serverless returns 404", func(t *testing.T) {
		// Sites without api/main.js should return 404 for /api/ paths
		resp := s.makeRequest(t, "GET", "/api/data", nil,
			withHost("alpha.testdomain.com"),
		)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected 404 for /api/ without serverless handler, got %d", resp.StatusCode)
		}
	})

	t.Run("analytics script injected into HTML responses", func(t *testing.T) {
		resp := s.makeRequest(t, "GET", "/", nil,
			withHost("alpha.testdomain.com"),
		)
		body := readBody(t, resp)

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		// Body should contain both the original content AND the injected analytics script
		assertContains(t, body, "Alpha Site")
		assertContains(t, body, "sendBeacon")
	})
}
