package handlers

import (
	"testing"
	"time"
)

func TestSchema_ForeignKeyConstraints(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO auth_sessions (token_hash, user_id, created_at, expires_at, last_seen)
		VALUES (?, ?, ?, ?, ?)
	`, "token_hash_1", "missing-user", now, now+3600, now)
	if err == nil {
		t.Fatal("expected foreign key constraint failure for auth_sessions.user_id")
	}
}

func TestSchema_UniqueConstraints(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`INSERT INTO auth_users (id, email, provider) VALUES (?, ?, ?)`, "user_1", "dup@test.local", "test")
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO auth_users (id, email, provider) VALUES (?, ?, ?)`, "user_2", "dup@test.local", "test")
	if err == nil {
		t.Fatal("expected unique constraint failure for auth_users.email")
	}
}

func TestSchema_DefaultValues(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`INSERT INTO apps (id) VALUES (?)`, "app_default_1")
	if err != nil {
		t.Fatalf("failed to insert app: %v", err)
	}

	var visibility, source string
	err = db.QueryRow(`SELECT visibility, source FROM apps WHERE id = ?`, "app_default_1").Scan(&visibility, &source)
	if err != nil {
		t.Fatalf("failed to query app defaults: %v", err)
	}

	if visibility != "unlisted" {
		t.Fatalf("visibility = %q, want %q", visibility, "unlisted")
	}
	if source != "deploy" {
		t.Fatalf("source = %q, want %q", source, "deploy")
	}
}
