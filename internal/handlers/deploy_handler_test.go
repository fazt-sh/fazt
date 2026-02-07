package handlers

import (
	"archive/zip"
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"
	"github.com/fazt-sh/fazt/internal/hosting"
)

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
	req.RemoteAddr = "127.0.0.1:1234"
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
