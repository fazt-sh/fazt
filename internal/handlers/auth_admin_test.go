package handlers

import (
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"

	"golang.org/x/crypto/bcrypt"
)

func setupAdminAuthTest(t *testing.T) *auth.Service {
	t.Helper()

	silenceTestLogs(t)

	db := setupTestDB(t)
	database.SetDB(db)
	t.Cleanup(func() {
		db.Close()
		database.SetDB(nil)
	})

	service := auth.NewService(db, "test.local", false)
	limiter := auth.NewRateLimiter()
	t.Cleanup(func() {
		limiter.Stop()
	})
	InitAuth(service, limiter, "v0.0.0-test")

	return service
}

func createTestUserWithRole(t *testing.T, service *auth.Service, role string) *auth.User {
	t.Helper()

	user, err := service.CreateUser(role+"@test.local", "Test User", "", "test", nil)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	if role != "user" {
		if err := service.UpdateUserRole(user.ID, role); err != nil {
			t.Fatalf("Failed to update user role: %v", err)
		}
		user.Role = role
	}

	return user
}

func createTestSessionToken(t *testing.T, service *auth.Service, userID string) string {
	t.Helper()

	token, err := service.CreateSession(userID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	return token
}

func insertTestAPIKey(t *testing.T, db *sql.DB, token string) {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("Failed to hash API key: %v", err)
	}

	_, err = db.Exec(`INSERT INTO api_keys (name, key_hash, scopes) VALUES (?, ?, ?)`, "test-key", string(hash), "deploy")
	if err != nil {
		t.Fatalf("Failed to insert API key: %v", err)
	}
}

func TestRequireAPIKeyAuth_Missing(t *testing.T) {
	setupAdminAuthTest(t)

	req := httptest.NewRequest("GET", "/api/apps", nil)
	rr := httptest.NewRecorder()

	ok := requireAPIKeyAuth(rr, req)
	if ok {
		t.Fatal("Expected missing API key auth to fail")
	}

	testutil.CheckError(t, rr, 401, "UNAUTHORIZED")
}

func TestRequireAPIKeyAuth_Invalid(t *testing.T) {
	setupAdminAuthTest(t)

	req := httptest.NewRequest("GET", "/api/apps", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	ok := requireAPIKeyAuth(rr, req)
	if ok {
		t.Fatal("Expected invalid API key auth to fail")
	}

	testutil.CheckError(t, rr, 401, "INVALID_API_KEY")
}

func TestRequireAPIKeyAuth_Valid(t *testing.T) {
	setupAdminAuthTest(t)
	token := "valid-token-123"
	insertTestAPIKey(t, database.GetDB(), token)

	req := httptest.NewRequest("GET", "/api/apps", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	ok := requireAPIKeyAuth(rr, req)
	if !ok {
		t.Fatal("Expected valid API key auth to succeed")
	}
	if database.GetDB() == nil {
		t.Fatal("Expected database to be initialized")
	}
}

func TestRequireAdminAuth_NoSession(t *testing.T) {
	setupAdminAuthTest(t)

	req := httptest.NewRequest("GET", "/api/apps", nil)
	rr := httptest.NewRecorder()

	_, ok := requireAdminAuth(rr, req)
	if ok {
		t.Fatal("Expected admin auth without session to fail")
	}

	testutil.CheckError(t, rr, 401, "UNAUTHORIZED")
}

func TestRequireAdminAuth_UserRole(t *testing.T) {
	service := setupAdminAuthTest(t)

	user := createTestUserWithRole(t, service, "user")
	session := createTestSessionToken(t, service, user.ID)

	req := httptest.NewRequest("GET", "/api/apps", nil)
	req = testutil.WithSession(req, session)
	rr := httptest.NewRecorder()

	_, ok := requireAdminAuth(rr, req)
	if ok {
		t.Fatal("Expected admin auth with user role to fail")
	}

	testutil.CheckError(t, rr, 403, "FORBIDDEN")
}

func TestRequireAdminAuth_AdminRole(t *testing.T) {
	service := setupAdminAuthTest(t)

	user := createTestUserWithRole(t, service, "admin")
	session := createTestSessionToken(t, service, user.ID)

	req := httptest.NewRequest("GET", "/api/apps", nil)
	req = testutil.WithSession(req, session)
	rr := httptest.NewRecorder()

	role, ok := requireAdminAuth(rr, req)
	if !ok {
		t.Fatal("Expected admin auth with admin role to succeed")
	}
	if role != "admin" {
		t.Fatalf("role = %q, want %q", role, "admin")
	}
}

func TestRequireAdminAuth_OwnerRole(t *testing.T) {
	service := setupAdminAuthTest(t)

	user := createTestUserWithRole(t, service, "owner")
	session := createTestSessionToken(t, service, user.ID)

	req := httptest.NewRequest("GET", "/api/apps", nil)
	req = testutil.WithSession(req, session)
	rr := httptest.NewRecorder()

	role, ok := requireAdminAuth(rr, req)
	if !ok {
		t.Fatal("Expected admin auth with owner role to succeed")
	}
	if role != "owner" {
		t.Fatalf("role = %q, want %q", role, "owner")
	}
}

func TestRequireAdminAuth_APIKey(t *testing.T) {
	setupAdminAuthTest(t)
	token := "valid-token-456"
	insertTestAPIKey(t, database.GetDB(), token)

	req := httptest.NewRequest("GET", "/api/apps", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	role, ok := requireAdminAuth(rr, req)
	if !ok {
		t.Fatal("Expected admin auth with API key to succeed")
	}
	if role != "owner" {
		t.Fatalf("role = %q, want %q", role, "owner")
	}
}
