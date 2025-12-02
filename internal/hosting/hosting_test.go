package hosting

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"io"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// setupTestDB creates a temporary in-memory database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}

	// Enable WAL (though not strictly needed for :memory:)
	db.Exec("PRAGMA journal_mode=WAL")

	// Create schema
	schema := `
	CREATE TABLE files (
		site_id TEXT NOT NULL,
		path TEXT NOT NULL,
		content BLOB,
		size_bytes INTEGER NOT NULL,
		mime_type TEXT,
		hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (site_id, path)
	);
	CREATE TABLE api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		key_hash TEXT NOT NULL,
		scopes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME
	);
	CREATE TABLE deployments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		site_id TEXT NOT NULL,
		size_bytes INTEGER,
		file_count INTEGER,
		deployed_by TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestVFS_WriteAndRead(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	fs := NewSQLFileSystem(db)
	
	// Test Write
	content := []byte("Hello World")
	err := fs.WriteFile("site1", "index.html", bytes.NewReader(content), int64(len(content)), "text/html")
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Test Read
	file, err := fs.ReadFile("site1", "index.html")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	defer file.Content.Close()

	readContent, _ := io.ReadAll(file.Content)
	if string(readContent) != string(content) {
		t.Errorf("Content mismatch. Got %s, want %s", readContent, content)
	}
	if file.MimeType != "text/html" {
		t.Errorf("MimeType mismatch. Got %s, want text/html", file.MimeType)
	}
}

func TestDeploySite(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Initialize hosting with DB
	Init(db)

	// Create a mock zip file
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	
	files := map[string]string{
		"index.html": "<h1>Hello</h1>",
		"css/style.css": "body { color: red; }",
		"main.js": "console.log('test');",
	}

	for name, content := range files {
		f, _ := zipWriter.Create(name)
		f.Write([]byte(content))
	}
	zipWriter.Close()

	// Create reader from buffer
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Failed to create zip reader: %v", err)
	}

	// Deploy
	res, err := DeploySite(zipReader, "test-site")
	if err != nil {
		t.Fatalf("DeploySite failed: %v", err)
	}

	if res.FileCount != 3 {
		t.Errorf("Expected 3 files, got %d", res.FileCount)
	}

	// Verify files in DB
	fs := GetFileSystem()
	
	// Check index.html
	exists, err := fs.Exists("test-site", "index.html")
	if !exists || err != nil {
		t.Error("index.html not found in VFS")
	}

	// Check subdirectory file
	exists, err = fs.Exists("test-site", "css/style.css")
	if !exists || err != nil {
		t.Error("css/style.css not found in VFS")
	}
}

func TestDeploySitePathTraversal(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	Init(db)

	// Create ZIP with path traversal attempt
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// Try to create a file with ..
	f, _ := w.Create("../../../etc/passwd")
	f.Write([]byte("malicious content"))

	// Also add a normal file
	f2, _ := w.Create("index.html")
	f2.Write([]byte("<h1>Normal</h1>"))

	w.Close()

	zipReader, _ := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))

	// Deploy should succeed but skip the malicious file
	result, err := DeploySite(zipReader, "safeside")
	if err != nil {
		t.Fatalf("DeploySite() failed: %v", err)
	}

	// Only the safe file should be deployed
	if result.FileCount != 1 {
		t.Errorf("result.FileCount = %d, want 1 (malicious file should be skipped)", result.FileCount)
	}

	// Verify malicious file was not created in VFS (implicitly handled by file count, but we can check existence)
	fs := GetFileSystem()
	exists, _ := fs.Exists("safeside", "../../../etc/passwd")
	if exists {
		t.Error("Malicious file was created in VFS")
	}
}

func TestDeploySiteInvalidSubdomain(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	Init(db)

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	w.Create("index.html")
	w.Close()
	zipReader, _ := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))

	// Note: "My-Site" is actually valid because it's lowercased to "my-site"
	invalidNames := []string{"", "../bad", "test.site", "test_site"}
	for _, name := range invalidNames {
		_, err := DeploySite(zipReader, name)
		if err == nil {
			t.Errorf("DeploySite(%q) should have failed", name)
		}
	}
}

func TestAPIKeyOperations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create API key
	token, err := CreateAPIKey(db, "test-key", "deploy")
	if err != nil {
		t.Fatalf("CreateAPIKey() failed: %v", err)
	}
	if token == "" {
		t.Error("CreateAPIKey() returned empty token")
	}

	// Validate the key
	id, name, err := ValidateAPIKey(db, token)
	if err != nil {
		t.Fatalf("ValidateAPIKey() failed: %v", err)
	}
	if name != "test-key" {
		t.Errorf("name = %q, want %q", name, "test-key")
	}
	if id == 0 {
		t.Error("id should not be 0")
	}

	// Validate with wrong key
	_, _, err = ValidateAPIKey(db, "wrong-token")
	if err == nil {
		t.Error("ValidateAPIKey() should fail with wrong token")
	}

	// List keys
	keys, err := ListAPIKeys(db)
	if err != nil {
		t.Fatalf("ListAPIKeys() failed: %v", err)
	}
	if len(keys) != 1 {
		t.Errorf("len(keys) = %d, want 1", len(keys))
	}

	// Delete key
	if err := DeleteAPIKey(db, id); err != nil {
		t.Fatalf("DeleteAPIKey() failed: %v", err)
	}

	// Verify deleted
	keys, _ = ListAPIKeys(db)
	if len(keys) != 0 {
		t.Errorf("len(keys) = %d, want 0 after delete", len(keys))
	}
}

func TestSiteExists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	Init(db)
	fs := GetFileSystem()

	// Case 1: No site
	if SiteExists("ghost") {
		t.Error("SiteExists returned true for non-existent site")
	}

	// Case 2: Static site (index.html)
	fs.WriteFile("static", "index.html", strings.NewReader("hi"), 2, "text/html")
	if !SiteExists("static") {
		t.Error("SiteExists returned false for static site")
	}

	// Case 3: Serverless site (main.js only)
	fs.WriteFile("app", "main.js", strings.NewReader("code"), 4, "text/javascript")
	if !SiteExists("app") {
		t.Error("SiteExists returned false for serverless app")
	}
}

func TestValidateSubdomain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
	}{
		{"valid simple", "mysite", false},
		{"valid with numbers", "site123", false},
		{"valid with hyphen", "my-site", false},
		{"empty", "", true},
		{"too long", "this-is-a-very-long-subdomain-name-that-exceeds-sixty-three-characters-limit", true},
		{"starts with hyphen", "-mysite", true},
		{"ends with hyphen", "mysite-", true},
		{"double hyphen", "my--site", false}, // double hyphens are allowed
		{"uppercase", "MySite", false},     // converted to lowercase, so valid
		{"underscore", "my_site", true},
		{"dot", "my.site", true},
		{"space", "my site", true},
		{"special chars", "my@site", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSubdomain(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSubdomain(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestSiteOperations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	Init(db)
	fs := GetFileSystem()

	siteName := "testsite"
	// CreateSite just validates subdomain now, but let's emulate creation by writing a file
	if err := CreateSite(siteName); err != nil {
		t.Fatalf("CreateSite() failed: %v", err)
	}
	fs.WriteFile(siteName, "index.html", strings.NewReader("hi"), 2, "text/html")

	if !SiteExists(siteName) {
		t.Error("SiteExists() returned false for created site")
	}

	sites, err := ListSites()
	if err != nil {
		t.Fatalf("ListSites() failed: %v", err)
	}
	if len(sites) != 1 || sites[0].Name != siteName {
		t.Errorf("ListSites() = %v, want [%s]", sites, siteName)
	}

	if err := DeleteSite(siteName); err != nil {
		t.Fatalf("DeleteSite() failed: %v", err)
	}

	if SiteExists(siteName) {
		t.Error("SiteExists() returned true for deleted site")
	}
}
