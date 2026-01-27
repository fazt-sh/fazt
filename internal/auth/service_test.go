package auth

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Create temp file for test DB
	tmpFile, err := os.CreateTemp("", "auth_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	db, err := sql.Open("sqlite", tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE auth_providers (
			name TEXT PRIMARY KEY,
			enabled INTEGER DEFAULT 0,
			client_id TEXT,
			client_secret TEXT,
			created_at INTEGER NOT NULL DEFAULT (unixepoch())
		);
		CREATE TABLE auth_users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			name TEXT,
			picture TEXT,
			provider TEXT NOT NULL,
			provider_id TEXT,
			password_hash TEXT,
			role TEXT DEFAULT 'user',
			invited_by TEXT,
			created_at INTEGER NOT NULL DEFAULT (unixepoch()),
			last_login INTEGER
		);
		CREATE TABLE auth_sessions (
			token_hash TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			created_at INTEGER NOT NULL DEFAULT (unixepoch()),
			expires_at INTEGER NOT NULL,
			last_seen INTEGER
		);
		CREATE TABLE auth_states (
			state TEXT PRIMARY KEY,
			provider TEXT NOT NULL,
			redirect_to TEXT,
			app_id TEXT,
			created_at INTEGER NOT NULL DEFAULT (unixepoch()),
			expires_at INTEGER NOT NULL
		);
		CREATE TABLE auth_invites (
			code TEXT PRIMARY KEY,
			role TEXT DEFAULT 'user',
			created_by TEXT NOT NULL,
			created_at INTEGER NOT NULL DEFAULT (unixepoch()),
			expires_at INTEGER,
			max_uses INTEGER DEFAULT 1,
			use_count INTEGER DEFAULT 0,
			used_by TEXT,
			used_at INTEGER
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	return db
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "test.com", false)

	// First user should become owner
	user, err := service.CreateUser("owner@test.com", "Owner", "", "google", nil)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	if user.Role != "owner" {
		t.Errorf("Expected first user to be owner, got %s", user.Role)
	}

	// Second user should be regular user
	user2, err := service.CreateUser("user@test.com", "User", "", "github", nil)
	if err != nil {
		t.Fatalf("Failed to create second user: %v", err)
	}
	if user2.Role != "user" {
		t.Errorf("Expected second user to be user, got %s", user2.Role)
	}
}

func TestCreateDuplicateUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "test.com", false)

	_, err := service.CreateUser("test@test.com", "Test", "", "google", nil)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = service.CreateUser("test@test.com", "Test 2", "", "github", nil)
	if err != ErrUserExists {
		t.Errorf("Expected ErrUserExists, got %v", err)
	}
}

func TestGetUserByID(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "test.com", false)

	created, _ := service.CreateUser("test@test.com", "Test", "", "google", nil)

	user, err := service.GetUserByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if user.Email != "test@test.com" {
		t.Errorf("Expected email test@test.com, got %s", user.Email)
	}
}

func TestGetUserNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "test.com", false)

	_, err := service.GetUserByID("nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestServiceCreateSession(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "test.com", false)

	user, _ := service.CreateUser("test@test.com", "Test", "", "google", nil)

	token, err := service.CreateSession(user.ID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	if token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestServiceValidateSession(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "test.com", false)

	user, _ := service.CreateUser("test@test.com", "Test", "", "google", nil)
	token, _ := service.CreateSession(user.ID)

	validatedUser, err := service.ValidateSession(token)
	if err != nil {
		t.Fatalf("Failed to validate session: %v", err)
	}
	if validatedUser.ID != user.ID {
		t.Errorf("User ID mismatch: expected %s, got %s", user.ID, validatedUser.ID)
	}
}

func TestValidateInvalidSession(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "test.com", false)

	_, err := service.ValidateSession("invalid-token")
	if err != ErrInvalidSession {
		t.Errorf("Expected ErrInvalidSession, got %v", err)
	}
}

func TestServiceDeleteSession(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "test.com", false)

	user, _ := service.CreateUser("test@test.com", "Test", "", "google", nil)
	token, _ := service.CreateSession(user.ID)

	err := service.DeleteSession(token)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	_, err = service.ValidateSession(token)
	if err != ErrInvalidSession {
		t.Error("Session should be invalid after deletion")
	}
}

func TestCreateInvite(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "test.com", false)

	invite, err := service.CreateInvite("user", "owner", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create invite: %v", err)
	}
	if invite.Code == "" {
		t.Error("Expected non-empty invite code")
	}
	if invite.Role != "user" {
		t.Errorf("Expected role 'user', got %s", invite.Role)
	}
}

func TestRedeemInvite(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "test.com", false)

	invite, _ := service.CreateInvite("user", "owner", 1, nil)

	user, err := service.RedeemInvite(invite.Code, "new@test.com", "New User", "password123")
	if err != nil {
		t.Fatalf("Failed to redeem invite: %v", err)
	}
	if user.Email != "new@test.com" {
		t.Errorf("Expected email new@test.com, got %s", user.Email)
	}

	// Try to redeem again - should fail
	_, err = service.RedeemInvite(invite.Code, "another@test.com", "Another", "password123")
	if err != ErrInviteUsed {
		t.Errorf("Expected ErrInviteUsed, got %v", err)
	}
}

func TestOAuthState(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "test.com", false)

	state, err := service.CreateState("google", "/dashboard", "app1")
	if err != nil {
		t.Fatalf("Failed to create state: %v", err)
	}

	validated, err := service.ValidateState(state)
	if err != nil {
		t.Fatalf("Failed to validate state: %v", err)
	}
	if validated.Provider != "google" {
		t.Errorf("Expected provider 'google', got %s", validated.Provider)
	}
	if validated.RedirectTo != "/dashboard" {
		t.Errorf("Expected redirect '/dashboard', got %s", validated.RedirectTo)
	}

	// State should be consumed
	_, err = service.ValidateState(state)
	if err != ErrInvalidState {
		t.Error("State should be invalid after use")
	}
}

func TestProviderConfig(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "test.com", false)

	err := service.SetProviderConfig("google", "client123", "secret456")
	if err != nil {
		t.Fatalf("Failed to set provider config: %v", err)
	}

	cfg, err := service.GetProviderConfig("google")
	if err != nil {
		t.Fatalf("Failed to get provider config: %v", err)
	}
	if cfg.ClientID != "client123" {
		t.Errorf("Expected client ID 'client123', got %s", cfg.ClientID)
	}
	if !cfg.Enabled {
		// Should still be disabled until explicitly enabled
	}

	err = service.EnableProvider("google")
	if err != nil {
		t.Fatalf("Failed to enable provider: %v", err)
	}

	cfg, _ = service.GetProviderConfig("google")
	if !cfg.Enabled {
		t.Error("Provider should be enabled")
	}
}

func TestSessionCookie(t *testing.T) {
	service := NewService(nil, "test.com", true)

	cookie := service.SessionCookie("token123", 3600)
	if cookie.Name != "fazt_session" {
		t.Errorf("Expected cookie name 'fazt_session', got %s", cookie.Name)
	}
	if cookie.Value != "token123" {
		t.Errorf("Expected cookie value 'token123', got %s", cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Error("Cookie should be HttpOnly")
	}
	if !cookie.Secure {
		t.Error("Cookie should be Secure in production")
	}
}

func TestInviteExpiry(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "test.com", false)

	// Create invite with past expiry by manipulating the DB directly
	code := "TESTEXPIRED"
	now := time.Now().Unix()
	pastExpiry := now - 100 // 100 seconds in the past
	_, err := db.Exec(`
		INSERT INTO auth_invites (code, role, created_by, created_at, expires_at, max_uses)
		VALUES (?, 'user', 'owner', ?, ?, 1)
	`, code, now, pastExpiry)
	if err != nil {
		t.Fatalf("Failed to insert expired invite: %v", err)
	}

	// Try to get invite - should exist but be invalid
	inv, err := service.GetInvite(code)
	if err != nil {
		t.Fatalf("Expected to find invite: %v", err)
	}
	if inv.IsValid() {
		t.Error("Invite should be invalid after expiry")
	}
}
