package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
)

func TestStatus(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "data.db")

	// Initialize DB and set config
	if err := database.Init(dbPath); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	store := config.NewDBConfigStore(database.GetDB())
	store.Set("server.domain", "https://test.example.com")
	store.Set("server.port", "4698")
	store.Set("server.env", "production")
	store.Set("auth.username", "admin")
	database.Close()

	output, err := statusCommand(dbPath)
	if err != nil {
		t.Fatalf("statusCommand failed: %v", err)
	}

	// Verify output contains expected information
	expectedStrings := []string{
		"Server Status",
		"Domain:",
		"https://test.example.com",
		"Port:",
		"4698",
		"Environment:",
		"production",
		"Username:",
		"admin",
		"Database:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Status output missing '%s'\nGot output:\n%s", expected, output)
		}
	}
}

func TestFullWorkflow_InitSetConfigStatus(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "data.db")

	// 1. Init
	err := initCommand("admin", "pass123", "https://test.com", "4698", "development", dbPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 2. Verify init worked
	if err := database.Init(dbPath); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	store := config.NewDBConfigStore(database.GetDB())
	dbMap, _ := store.Load()
	database.Close()

	if dbMap["server.domain"] != "https://test.com" {
		t.Error("Init didn't set domain correctly")
	}

	// 3. Update credentials
	err = setCredentialsCommand("newadmin", "newpass", dbPath)
	if err != nil {
		t.Fatalf("set-credentials failed: %v", err)
	}

	// 4. Update config
	err = setConfigCommand("https://new.com", "8080", "production", dbPath)
	if err != nil {
		t.Fatalf("set-config failed: %v", err)
	}

	// 5. Verify updates worked
	if err := database.Init(dbPath); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	store = config.NewDBConfigStore(database.GetDB())
	dbMap, _ = store.Load()
	database.Close()

	if dbMap["server.domain"] != "https://new.com" {
		t.Errorf("Domain not updated. Got: %s", dbMap["server.domain"])
	}
	if dbMap["server.port"] != "8080" {
		t.Errorf("Port not updated. Got: %s", dbMap["server.port"])
	}
	if dbMap["auth.username"] != "newadmin" {
		t.Errorf("Username not updated. Got: %s", dbMap["auth.username"])
	}

	// 6. Status check
	output, err := statusCommand(dbPath)
	if err != nil {
		t.Fatalf("statusCommand failed: %v", err)
	}

	if !strings.Contains(output, "https://new.com") {
		t.Error("Status output doesn't reflect new domain")
	}
}