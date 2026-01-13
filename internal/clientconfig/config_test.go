package clientconfig

import (
	"os"
	"path/filepath"
	"testing"
)

// testConfigDir creates a temporary config directory for testing
func testConfigDir(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "fazt-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Override the config path for testing
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	cleanup := func() {
		os.Setenv("HOME", origHome)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestConfigPath(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	path := ConfigPath()
	if filepath.Base(path) != "config.json" {
		t.Errorf("ConfigPath() should end with config.json, got %s", path)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("ConfigPath() should be absolute, got %s", path)
	}
}

func TestLoad_Empty(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(cfg.Servers))
	}
	if cfg.DefaultServer != "" {
		t.Errorf("expected empty default server, got %s", cfg.DefaultServer)
	}
}

func TestSaveAndLoad(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	// Create and save config
	cfg := &Config{
		Version:       1,
		DefaultServer: "test",
		Servers: map[string]Server{
			"test": {URL: "https://test.example.com", Token: "token123"},
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Load it back
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if loaded.Version != cfg.Version {
		t.Errorf("version mismatch: got %d, want %d", loaded.Version, cfg.Version)
	}
	if loaded.DefaultServer != cfg.DefaultServer {
		t.Errorf("default server mismatch: got %s, want %s", loaded.DefaultServer, cfg.DefaultServer)
	}
	if len(loaded.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(loaded.Servers))
	}

	srv := loaded.Servers["test"]
	if srv.URL != "https://test.example.com" {
		t.Errorf("URL mismatch: got %s", srv.URL)
	}
	if srv.Token != "token123" {
		t.Errorf("Token mismatch: got %s", srv.Token)
	}
}

func TestAddServer(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()

	// Add first server
	if err := cfg.AddServer("prod", "https://prod.example.com", "token1"); err != nil {
		t.Fatalf("AddServer() failed: %v", err)
	}

	if !cfg.HasServer("prod") {
		t.Error("server 'prod' not found after AddServer")
	}

	// First server should become default
	if cfg.DefaultServer != "prod" {
		t.Errorf("first server should be default, got %s", cfg.DefaultServer)
	}

	// Add second server
	if err := cfg.AddServer("dev", "https://dev.example.com", "token2"); err != nil {
		t.Fatalf("AddServer() failed: %v", err)
	}

	// Default should still be prod
	if cfg.DefaultServer != "prod" {
		t.Errorf("default should still be prod, got %s", cfg.DefaultServer)
	}

	if cfg.ServerCount() != 2 {
		t.Errorf("expected 2 servers, got %d", cfg.ServerCount())
	}
}

func TestAddServer_Validation(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()

	tests := []struct {
		name    string
		srvName string
		url     string
		token   string
		wantErr bool
	}{
		{"empty name", "", "https://example.com", "token", true},
		{"empty url", "test", "", "token", true},
		{"empty token", "test", "https://example.com", "", true},
		{"valid", "test", "https://example.com", "token", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cfg.AddServer(tt.srvName, tt.url, tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveServer(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()
	cfg.AddServer("prod", "https://prod.example.com", "token1")
	cfg.AddServer("dev", "https://dev.example.com", "token2")

	// Remove non-default server
	if err := cfg.RemoveServer("dev"); err != nil {
		t.Fatalf("RemoveServer() failed: %v", err)
	}

	if cfg.HasServer("dev") {
		t.Error("server 'dev' should be removed")
	}
	if cfg.DefaultServer != "prod" {
		t.Errorf("default should still be prod, got %s", cfg.DefaultServer)
	}

	// Remove default server
	cfg.AddServer("staging", "https://staging.example.com", "token3")
	if err := cfg.RemoveServer("prod"); err != nil {
		t.Fatalf("RemoveServer() failed: %v", err)
	}

	// When there's one server left, it should become the default
	if cfg.DefaultServer != "staging" {
		t.Errorf("default should be staging after removing prod, got %s", cfg.DefaultServer)
	}
}

func TestRemoveServer_NotFound(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()

	err := cfg.RemoveServer("nonexistent")
	if err == nil {
		t.Error("RemoveServer() should return error for nonexistent server")
	}
}

func TestSetDefault(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()
	cfg.AddServer("prod", "https://prod.example.com", "token1")
	cfg.AddServer("dev", "https://dev.example.com", "token2")

	if err := cfg.SetDefault("dev"); err != nil {
		t.Fatalf("SetDefault() failed: %v", err)
	}

	if cfg.DefaultServer != "dev" {
		t.Errorf("default should be dev, got %s", cfg.DefaultServer)
	}
}

func TestSetDefault_NotFound(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()

	err := cfg.SetDefault("nonexistent")
	if err == nil {
		t.Error("SetDefault() should return error for nonexistent server")
	}
}

func TestGetServer_Explicit(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()
	cfg.AddServer("prod", "https://prod.example.com", "token1")
	cfg.AddServer("dev", "https://dev.example.com", "token2")

	srv, name, err := cfg.GetServer("dev")
	if err != nil {
		t.Fatalf("GetServer() failed: %v", err)
	}

	if name != "dev" {
		t.Errorf("expected name 'dev', got %s", name)
	}
	if srv.URL != "https://dev.example.com" {
		t.Errorf("expected dev URL, got %s", srv.URL)
	}
}

func TestGetServer_ExplicitNotFound(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()
	cfg.AddServer("prod", "https://prod.example.com", "token1")

	_, _, err := cfg.GetServer("nonexistent")
	if err == nil {
		t.Error("GetServer() should return error for nonexistent server")
	}
}

func TestGetServer_SmartDefault_SingleServer(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()
	cfg.AddServer("prod", "https://prod.example.com", "token1")
	cfg.DefaultServer = "" // Clear default to test auto-selection

	srv, name, err := cfg.GetServer("")
	if err != nil {
		t.Fatalf("GetServer() failed: %v", err)
	}

	if name != "prod" {
		t.Errorf("expected 'prod' (single server), got %s", name)
	}
	if srv.URL != "https://prod.example.com" {
		t.Errorf("expected prod URL, got %s", srv.URL)
	}
}

func TestGetServer_SmartDefault_WithDefault(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()
	cfg.AddServer("prod", "https://prod.example.com", "token1")
	cfg.AddServer("dev", "https://dev.example.com", "token2")
	cfg.SetDefault("dev")

	srv, name, err := cfg.GetServer("")
	if err != nil {
		t.Fatalf("GetServer() failed: %v", err)
	}

	if name != "dev" {
		t.Errorf("expected 'dev' (default), got %s", name)
	}
	if srv.URL != "https://dev.example.com" {
		t.Errorf("expected dev URL, got %s", srv.URL)
	}
}

func TestGetServer_NoServers(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()

	_, _, err := cfg.GetServer("")
	if err == nil {
		t.Error("GetServer() should return error when no servers configured")
	}
}

func TestGetServer_MultipleNoDefault(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()
	cfg.AddServer("prod", "https://prod.example.com", "token1")
	cfg.AddServer("dev", "https://dev.example.com", "token2")
	cfg.DefaultServer = "" // Clear default

	_, _, err := cfg.GetServer("")
	if err == nil {
		t.Error("GetServer() should return error when multiple servers and no default")
	}
}

func TestListServers(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()
	cfg.AddServer("prod", "https://prod.example.com", "token1")
	cfg.AddServer("dev", "https://dev.example.com", "token2")
	cfg.SetDefault("prod")

	servers := cfg.ListServers()
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}

	// Find prod and dev in the list
	var foundProd, foundDev bool
	for _, s := range servers {
		if s.Name == "prod" {
			foundProd = true
			if !s.IsDefault {
				t.Error("prod should be marked as default")
			}
		}
		if s.Name == "dev" {
			foundDev = true
			if s.IsDefault {
				t.Error("dev should not be marked as default")
			}
		}
	}

	if !foundProd || !foundDev {
		t.Error("missing servers in list")
	}
}

func TestServerCount(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()

	if cfg.ServerCount() != 0 {
		t.Errorf("expected 0 servers, got %d", cfg.ServerCount())
	}

	cfg.AddServer("test", "https://test.example.com", "token")
	if cfg.ServerCount() != 1 {
		t.Errorf("expected 1 server, got %d", cfg.ServerCount())
	}

	cfg.AddServer("test2", "https://test2.example.com", "token2")
	if cfg.ServerCount() != 2 {
		t.Errorf("expected 2 servers, got %d", cfg.ServerCount())
	}
}

func TestHasServer(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()
	cfg.AddServer("prod", "https://prod.example.com", "token1")

	if !cfg.HasServer("prod") {
		t.Error("HasServer() should return true for existing server")
	}
	if cfg.HasServer("nonexistent") {
		t.Error("HasServer() should return false for nonexistent server")
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fazt-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Point HOME to a directory without .fazt
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := &Config{
		Version: 1,
		Servers: map[string]Server{
			"test": {URL: "https://test.example.com", Token: "token"},
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Check that directory was created
	faztDir := filepath.Join(tmpDir, ".fazt")
	info, err := os.Stat(faztDir)
	if err != nil {
		t.Fatalf("config directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("config directory is not a directory")
	}

	// Check permissions (should be 0700)
	if info.Mode().Perm() != 0700 {
		t.Errorf("config directory permissions should be 0700, got %o", info.Mode().Perm())
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fazt-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create .fazt directory
	faztDir := filepath.Join(tmpDir, ".fazt")
	os.MkdirAll(faztDir, 0700)

	// Write invalid JSON
	configPath := filepath.Join(faztDir, "config.json")
	os.WriteFile(configPath, []byte("not valid json"), 0600)

	_, err = Load()
	if err == nil {
		t.Error("Load() should return error for invalid JSON")
	}
}

func TestConfig_UpdateServer(t *testing.T) {
	_, cleanup := testConfigDir(t)
	defer cleanup()

	cfg, _ := Load()
	cfg.AddServer("prod", "https://prod.example.com", "token1")

	// Update the server with new values
	cfg.AddServer("prod", "https://new-prod.example.com", "newtoken")

	srv, _, _ := cfg.GetServer("prod")
	if srv.URL != "https://new-prod.example.com" {
		t.Errorf("URL should be updated, got %s", srv.URL)
	}
	if srv.Token != "newtoken" {
		t.Errorf("Token should be updated, got %s", srv.Token)
	}
}
