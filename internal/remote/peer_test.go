package remote

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "fazt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create peers table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS peers (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(8)))),
			name TEXT UNIQUE NOT NULL,
			url TEXT NOT NULL,
			token TEXT,
			description TEXT,
			is_default INTEGER DEFAULT 0,
			last_seen_at TEXT,
			last_version TEXT,
			last_status TEXT,
			node_id TEXT,
			public_key TEXT,
			created_at TEXT DEFAULT (datetime('now')),
			updated_at TEXT DEFAULT (datetime('now'))
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_peers_default
			ON peers(is_default) WHERE is_default = 1;
	`)
	if err != nil {
		db.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create table: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

func TestAddPeer(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Add a peer
	err := AddPeer(db, "test", "https://test.example.com", "token123", "Test peer")
	if err != nil {
		t.Fatalf("AddPeer failed: %v", err)
	}

	// Verify it was added
	peer, err := GetPeer(db, "test")
	if err != nil {
		t.Fatalf("GetPeer failed: %v", err)
	}

	if peer.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", peer.Name)
	}
	if peer.URL != "https://test.example.com" {
		t.Errorf("Expected URL 'https://test.example.com', got '%s'", peer.URL)
	}
	if peer.Token != "token123" {
		t.Errorf("Expected token 'token123', got '%s'", peer.Token)
	}
}

func TestAddPeerDuplicate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Add a peer
	err := AddPeer(db, "test", "https://test.example.com", "token123", "")
	if err != nil {
		t.Fatalf("AddPeer failed: %v", err)
	}

	// Try to add duplicate
	err = AddPeer(db, "test", "https://other.example.com", "token456", "")
	if err != ErrPeerAlreadyExists {
		t.Errorf("Expected ErrPeerAlreadyExists, got %v", err)
	}
}

func TestRemovePeer(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Add and remove
	AddPeer(db, "test", "https://test.example.com", "token123", "")

	err := RemovePeer(db, "test")
	if err != nil {
		t.Fatalf("RemovePeer failed: %v", err)
	}

	// Verify removed
	_, err = GetPeer(db, "test")
	if err != ErrPeerNotFound {
		t.Errorf("Expected ErrPeerNotFound, got %v", err)
	}
}

func TestSetDefaultPeer(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Add two peers
	AddPeer(db, "peer1", "https://peer1.example.com", "token1", "")
	AddPeer(db, "peer2", "https://peer2.example.com", "token2", "")

	// Set peer1 as default
	err := SetDefaultPeer(db, "peer1")
	if err != nil {
		t.Fatalf("SetDefaultPeer failed: %v", err)
	}

	// Verify
	peer, err := GetDefaultPeer(db)
	if err != nil {
		t.Fatalf("GetDefaultPeer failed: %v", err)
	}
	if peer.Name != "peer1" {
		t.Errorf("Expected default 'peer1', got '%s'", peer.Name)
	}

	// Change default
	SetDefaultPeer(db, "peer2")
	peer, _ = GetDefaultPeer(db)
	if peer.Name != "peer2" {
		t.Errorf("Expected default 'peer2', got '%s'", peer.Name)
	}
}

func TestListPeers(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Add peers
	AddPeer(db, "alpha", "https://alpha.example.com", "token1", "")
	AddPeer(db, "beta", "https://beta.example.com", "token2", "")
	AddPeer(db, "gamma", "https://gamma.example.com", "token3", "")

	peers, err := ListPeers(db)
	if err != nil {
		t.Fatalf("ListPeers failed: %v", err)
	}

	if len(peers) != 3 {
		t.Errorf("Expected 3 peers, got %d", len(peers))
	}

	// Should be sorted by name
	if peers[0].Name != "alpha" {
		t.Errorf("Expected first peer 'alpha', got '%s'", peers[0].Name)
	}
}

func TestResolvePeer(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// No peers - should fail
	_, err := ResolvePeer(db, "")
	if err != ErrNoPeers {
		t.Errorf("Expected ErrNoPeers, got %v", err)
	}

	// One peer - should auto-select
	AddPeer(db, "only", "https://only.example.com", "token", "")
	peer, err := ResolvePeer(db, "")
	if err != nil {
		t.Fatalf("ResolvePeer failed: %v", err)
	}
	if peer.Name != "only" {
		t.Errorf("Expected 'only', got '%s'", peer.Name)
	}

	// Multiple peers, no default - should fail
	AddPeer(db, "another", "https://another.example.com", "token2", "")
	_, err = ResolvePeer(db, "")
	if err != ErrNoDefaultPeer {
		t.Errorf("Expected ErrNoDefaultPeer, got %v", err)
	}

	// With explicit name
	peer, err = ResolvePeer(db, "another")
	if err != nil {
		t.Fatalf("ResolvePeer with name failed: %v", err)
	}
	if peer.Name != "another" {
		t.Errorf("Expected 'another', got '%s'", peer.Name)
	}
}
