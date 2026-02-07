package handlers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"
	"github.com/fazt-sh/fazt/internal/hosting"
)

var deployIPCounter uint32

func setupDeployHandlerTest(t *testing.T) string {
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

	token := "deploy-token-123"
	insertTestAPIKey(t, db, token)

	return token
}

func buildZip(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	file, err := writer.Create("index.html")
	if err != nil {
		t.Fatalf("Failed to create zip entry: %v", err)
	}

	if _, err := io.WriteString(file, "<html>ok</html>"); err != nil {
		t.Fatalf("Failed to write zip entry: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

func newDeployRequest(t *testing.T, siteName, filename string, fileData []byte) (*http.Request, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if siteName != "" {
		if err := writer.WriteField("site_name", siteName); err != nil {
			t.Fatalf("Failed to write site_name field: %v", err)
		}
	}

	if filename != "" {
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			t.Fatalf("Failed to create file part: %v", err)
		}
		if _, err := part.Write(fileData); err != nil {
			t.Fatalf("Failed to write file data: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/deploy", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// Use unique IP per request to avoid rate limiting interference across tests
	ipNum := atomic.AddUint32(&deployIPCounter, 1)
	req.RemoteAddr = fmt.Sprintf("10.%d.%d.%d:1234", (ipNum>>16)&0xFF, (ipNum>>8)&0xFF, ipNum&0xFF)
	return req, writer.FormDataContentType()
}

func TestDeployHandler_MethodNotAllowed(t *testing.T) {
	setupDeployHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/deploy", nil)
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)

	testutil.CheckError(t, rr, http.StatusBadRequest, "BAD_REQUEST")
}

func TestDeployHandler_MissingAuthorization(t *testing.T) {
	setupDeployHandlerTest(t)

	req, _ := newDeployRequest(t, "test", "site.zip", buildZip(t))
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)

	testutil.CheckError(t, rr, http.StatusUnauthorized, "UNAUTHORIZED")
}

func TestDeployHandler_InvalidAuthorizationFormat(t *testing.T) {
	setupDeployHandlerTest(t)

	req, _ := newDeployRequest(t, "test", "site.zip", buildZip(t))
	req.Header.Set("Authorization", "Token abc")
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)

	testutil.CheckError(t, rr, http.StatusUnauthorized, "UNAUTHORIZED")
}

func TestDeployHandler_InvalidAPIKey(t *testing.T) {
	setupDeployHandlerTest(t)

	req, _ := newDeployRequest(t, "test", "site.zip", buildZip(t))
	req.Header.Set("Authorization", "Bearer invalid")
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)

	testutil.CheckError(t, rr, http.StatusUnauthorized, "INVALID_API_KEY")
}

func TestDeployHandler_MissingSiteName(t *testing.T) {
	token := setupDeployHandlerTest(t)

	req, _ := newDeployRequest(t, "", "site.zip", buildZip(t))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)

	testutil.CheckError(t, rr, http.StatusBadRequest, "BAD_REQUEST")
}

func TestDeployHandler_InvalidFileType(t *testing.T) {
	token := setupDeployHandlerTest(t)

	req, _ := newDeployRequest(t, "test", "site.txt", []byte("not-zip"))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)

	testutil.CheckError(t, rr, http.StatusBadRequest, "BAD_REQUEST")
}

func TestDeployHandler_Success(t *testing.T) {
	token := setupDeployHandlerTest(t)

	req, _ := newDeployRequest(t, "my-site", "site.zip", buildZip(t))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "site", "my-site")
	testutil.AssertFieldEquals(t, data, "message", "Deployment successful")
}

// TestDeployHandler_RateLimitExhaustion tests deploy rate limiting (5 per minute)
func TestDeployHandler_RateLimitExhaustion(t *testing.T) {
	token := setupDeployHandlerTest(t)

	// Make 5 successful deploy requests
	for i := 0; i < 5; i++ {
		req, _ := newDeployRequest(t, "site"+testutil.RandStr(5), "site.zip", buildZip(t))
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "192.168.1.100:12345"
		rr := httptest.NewRecorder()

		DeployHandler(rr, req)
		testutil.CheckSuccess(t, rr, http.StatusOK)
	}

	// 6th request should hit rate limit
	req, _ := newDeployRequest(t, "another-site", "site.zip", buildZip(t))
	req.Header.Set("Authorization", "Bearer "+token)
	req.RemoteAddr = "192.168.1.100:12345"
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)
	testutil.CheckError(t, rr, 429, "RATE_LIMIT_EXCEEDED")
}

// TestDeployHandler_RateLimitPerIP tests that rate limiting is per IP
func TestDeployHandler_RateLimitPerIP(t *testing.T) {
	token := setupDeployHandlerTest(t)

	// Exhaust rate limit for IP1
	for i := 0; i < 5; i++ {
		req, _ := newDeployRequest(t, "site1-"+testutil.RandStr(5), "site.zip", buildZip(t))
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "192.168.1.10:12345"
		rr := httptest.NewRecorder()

		DeployHandler(rr, req)
		testutil.CheckSuccess(t, rr, http.StatusOK)
	}

	// IP1 should be rate limited
	req1, _ := newDeployRequest(t, "blocked-site", "site.zip", buildZip(t))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.RemoteAddr = "192.168.1.10:12345"
	rr1 := httptest.NewRecorder()

	DeployHandler(rr1, req1)
	testutil.CheckError(t, rr1, 429, "RATE_LIMIT_EXCEEDED")

	// IP2 should NOT be rate limited
	req2, _ := newDeployRequest(t, "allowed-site", "site.zip", buildZip(t))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.RemoteAddr = "192.168.1.20:12345"
	rr2 := httptest.NewRecorder()

	DeployHandler(rr2, req2)
	testutil.CheckSuccess(t, rr2, http.StatusOK)
}

// TestDeployHandler_MalformedZIP tests deployment with corrupted ZIP data
func TestDeployHandler_MalformedZIP(t *testing.T) {
	token := setupDeployHandlerTest(t)

	// Create invalid ZIP data
	invalidZip := []byte("PK\x03\x04\x00\x00CORRUPTED")

	req, _ := newDeployRequest(t, "test-site", "site.zip", invalidZip)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)

	testutil.CheckError(t, rr, http.StatusBadRequest, "BAD_REQUEST")
}

// TestDeployHandler_EmptyZIP tests deployment with empty ZIP file
func TestDeployHandler_EmptyZIP(t *testing.T) {
	token := setupDeployHandlerTest(t)

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	req, _ := newDeployRequest(t, "empty-site", "site.zip", buf.Bytes())
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)

	// Empty ZIP should succeed (no files to deploy)
	testutil.CheckSuccess(t, rr, http.StatusOK)
}

// TestDeployHandler_VeryLongSiteName tests deployment with excessively long site name
func TestDeployHandler_VeryLongSiteName(t *testing.T) {
	token := setupDeployHandlerTest(t)

	longName := testutil.RandStr(300)

	req, _ := newDeployRequest(t, longName, "site.zip", buildZip(t))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)

	// Should fail validation
	testutil.CheckError(t, rr, http.StatusBadRequest, "BAD_REQUEST")
}

// TestDeployHandler_InvalidSiteName tests deployment with invalid characters
func TestDeployHandler_InvalidSiteName(t *testing.T) {
	token := setupDeployHandlerTest(t)

	tests := []struct {
		name     string
		siteName string
	}{
		{"dots", "test.site"},
		{"underscore", "test_site"},
		{"special chars", "test@site!"},
		{"empty", ""},
		{"leading dash", "-test"},
		{"trailing dash", "test-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := newDeployRequest(t, tt.siteName, "site.zip", buildZip(t))
			req.Header.Set("Authorization", "Bearer "+token)
			rr := httptest.NewRecorder()

			DeployHandler(rr, req)

			testutil.CheckError(t, rr, http.StatusBadRequest, "BAD_REQUEST")
		})
	}
}

// TestDeployHandler_DomainStripping tests smart domain handling
func TestDeployHandler_DomainStripping(t *testing.T) {
	token := setupDeployHandlerTest(t)

	// Deploy with full domain name (should strip .test.local)
	req, _ := newDeployRequest(t, "my-site.test.local", "site.zip", buildZip(t))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)

	data := testutil.CheckSuccess(t, rr, http.StatusOK)
	// Should have stripped the domain suffix
	testutil.AssertFieldEquals(t, data, "site", "my-site")
}

// TestDeployHandler_ZIPWithManyFiles tests deployment with large number of files
func TestDeployHandler_ZIPWithManyFiles(t *testing.T) {
	token := setupDeployHandlerTest(t)

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	// Create 100 files
	for i := 0; i < 100; i++ {
		file, err := writer.Create("file" + testutil.RandStr(10) + ".html")
		if err != nil {
			t.Fatalf("Failed to create zip entry: %v", err)
		}
		if _, err := io.WriteString(file, "<html>content</html>"); err != nil {
			t.Fatalf("Failed to write zip entry: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	req, _ := newDeployRequest(t, "many-files", "site.zip", buf.Bytes())
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)

	testutil.CheckSuccess(t, rr, http.StatusOK)
}

// TestDeployHandler_ZIPWithSpecialCharFilenames tests files with unicode/special chars
func TestDeployHandler_ZIPWithSpecialCharFilenames(t *testing.T) {
	token := setupDeployHandlerTest(t)

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	// Create files with special characters
	specialNames := []string{
		"index.html",
		"hello-world.html",
		"about_us.html",
	}

	for _, name := range specialNames {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Failed to create zip entry: %v", err)
		}
		if _, err := io.WriteString(file, "<html>ok</html>"); err != nil {
			t.Fatalf("Failed to write zip entry: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	req, _ := newDeployRequest(t, "special-chars", "site.zip", buf.Bytes())
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)

	testutil.CheckSuccess(t, rr, http.StatusOK)
}

// TestDeployHandler_XForwardedFor tests IP extraction from X-Forwarded-For header
func TestDeployHandler_XForwardedFor(t *testing.T) {
	token := setupDeployHandlerTest(t)

	// Make 5 deploys with X-Forwarded-For header
	for i := 0; i < 5; i++ {
		req, _ := newDeployRequest(t, "xff-"+testutil.RandStr(5), "site.zip", buildZip(t))
		req.Header.Set("Authorization", "Bearer "+token)
		req.RemoteAddr = "10.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.50, 192.168.1.1")
		rr := httptest.NewRecorder()

		DeployHandler(rr, req)
		testutil.CheckSuccess(t, rr, http.StatusOK)
	}

	// Should be rate limited based on X-Forwarded-For IP
	req, _ := newDeployRequest(t, "xff-blocked", "site.zip", buildZip(t))
	req.Header.Set("Authorization", "Bearer "+token)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 192.168.1.1")
	rr := httptest.NewRecorder()

	DeployHandler(rr, req)
	testutil.CheckError(t, rr, 429, "RATE_LIMIT_EXCEEDED")
}
