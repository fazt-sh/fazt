// Package remote provides fazt-to-fazt communication capabilities.
// All peer configuration is stored in SQLite - no external files.
package remote

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Peer represents a known remote fazt node
type Peer struct {
	ID          string
	Name        string
	URL         string
	Token       string
	Description string
	IsDefault   bool

	// Connection state
	LastSeenAt  *time.Time
	LastVersion string
	LastStatus  string

	// Future: Mesh identity
	NodeID    string
	PublicKey string

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrPeerNotFound      = errors.New("peer not found")
	ErrPeerAlreadyExists = errors.New("peer already exists")
	ErrNoPeers           = errors.New("no peers configured")
	ErrNoDefaultPeer     = errors.New("no default peer set")
)

// AddPeer adds a new remote peer
func AddPeer(db *sql.DB, name, url, token, description string) error {
	_, err := db.Exec(`
		INSERT INTO peers (name, url, token, description)
		VALUES (?, ?, ?, ?)
	`, name, url, token, description)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrPeerAlreadyExists
		}
		return fmt.Errorf("failed to add peer: %w", err)
	}
	return nil
}

// RemovePeer removes a peer by name
func RemovePeer(db *sql.DB, name string) error {
	result, err := db.Exec("DELETE FROM peers WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to remove peer: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrPeerNotFound
	}
	return nil
}

// GetPeer retrieves a peer by name
func GetPeer(db *sql.DB, name string) (*Peer, error) {
	peer := &Peer{}
	var lastSeenAt, createdAt, updatedAt sql.NullString
	var token, description, lastVersion, lastStatus, nodeID, publicKey sql.NullString

	err := db.QueryRow(`
		SELECT id, name, url, token, description, is_default,
		       last_seen_at, last_version, last_status,
		       node_id, public_key, created_at, updated_at
		FROM peers WHERE name = ?
	`, name).Scan(
		&peer.ID, &peer.Name, &peer.URL, &token, &description, &peer.IsDefault,
		&lastSeenAt, &lastVersion, &lastStatus,
		&nodeID, &publicKey, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrPeerNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get peer: %w", err)
	}

	// Handle nullable fields
	if token.Valid {
		peer.Token = token.String
	}
	if description.Valid {
		peer.Description = description.String
	}
	if lastSeenAt.Valid {
		t, _ := time.Parse(time.RFC3339, lastSeenAt.String)
		peer.LastSeenAt = &t
	}
	if lastVersion.Valid {
		peer.LastVersion = lastVersion.String
	}
	if lastStatus.Valid {
		peer.LastStatus = lastStatus.String
	}
	if nodeID.Valid {
		peer.NodeID = nodeID.String
	}
	if publicKey.Valid {
		peer.PublicKey = publicKey.String
	}
	if createdAt.Valid {
		peer.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
	}
	if updatedAt.Valid {
		peer.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt.String)
	}

	return peer, nil
}

// GetDefaultPeer retrieves the default peer
func GetDefaultPeer(db *sql.DB) (*Peer, error) {
	var name string
	err := db.QueryRow("SELECT name FROM peers WHERE is_default = 1").Scan(&name)
	if err == sql.ErrNoRows {
		// If no default, try to get the only peer
		var count int
		db.QueryRow("SELECT COUNT(*) FROM peers").Scan(&count)
		if count == 0 {
			return nil, ErrNoPeers
		}
		if count == 1 {
			db.QueryRow("SELECT name FROM peers LIMIT 1").Scan(&name)
			return GetPeer(db, name)
		}
		return nil, ErrNoDefaultPeer
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get default peer: %w", err)
	}
	return GetPeer(db, name)
}

// ListPeers returns all known peers
func ListPeers(db *sql.DB) ([]Peer, error) {
	rows, err := db.Query(`
		SELECT id, name, url, token, description, is_default,
		       last_seen_at, last_version, last_status,
		       node_id, public_key, created_at, updated_at
		FROM peers ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list peers: %w", err)
	}
	defer rows.Close()

	var peers []Peer
	for rows.Next() {
		var peer Peer
		var lastSeenAt, createdAt, updatedAt sql.NullString
		var token, description, lastVersion, lastStatus, nodeID, publicKey sql.NullString

		err := rows.Scan(
			&peer.ID, &peer.Name, &peer.URL, &token, &description, &peer.IsDefault,
			&lastSeenAt, &lastVersion, &lastStatus,
			&nodeID, &publicKey, &createdAt, &updatedAt,
		)
		if err != nil {
			continue
		}

		// Handle nullable fields
		if token.Valid {
			peer.Token = token.String
		}
		if description.Valid {
			peer.Description = description.String
		}
		if lastSeenAt.Valid {
			t, _ := time.Parse(time.RFC3339, lastSeenAt.String)
			peer.LastSeenAt = &t
		}
		if lastVersion.Valid {
			peer.LastVersion = lastVersion.String
		}
		if lastStatus.Valid {
			peer.LastStatus = lastStatus.String
		}
		if nodeID.Valid {
			peer.NodeID = nodeID.String
		}
		if publicKey.Valid {
			peer.PublicKey = publicKey.String
		}
		if createdAt.Valid {
			peer.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
		}
		if updatedAt.Valid {
			peer.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt.String)
		}

		peers = append(peers, peer)
	}

	return peers, nil
}

// SetDefaultPeer sets a peer as the default
func SetDefaultPeer(db *sql.DB, name string) error {
	// First verify the peer exists
	_, err := GetPeer(db, name)
	if err != nil {
		return err
	}

	// Clear existing default
	_, err = db.Exec("UPDATE peers SET is_default = 0 WHERE is_default = 1")
	if err != nil {
		return fmt.Errorf("failed to clear default: %w", err)
	}

	// Set new default
	_, err = db.Exec("UPDATE peers SET is_default = 1, updated_at = datetime('now') WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	return nil
}

// UpdatePeerStatus updates connection state after contacting a peer
func UpdatePeerStatus(db *sql.DB, name, status, version string) error {
	_, err := db.Exec(`
		UPDATE peers
		SET last_seen_at = datetime('now'),
		    last_status = ?,
		    last_version = ?,
		    updated_at = datetime('now')
		WHERE name = ?
	`, status, version, name)
	if err != nil {
		return fmt.Errorf("failed to update peer status: %w", err)
	}
	return nil
}

// UpdatePeerToken updates a peer's authentication token
func UpdatePeerToken(db *sql.DB, name, token string) error {
	result, err := db.Exec(`
		UPDATE peers SET token = ?, updated_at = datetime('now')
		WHERE name = ?
	`, token, name)
	if err != nil {
		return fmt.Errorf("failed to update token: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrPeerNotFound
	}
	return nil
}

// ResolvePeer gets a peer by name, or returns default if name is empty
func ResolvePeer(db *sql.DB, name string) (*Peer, error) {
	if name == "" {
		return GetDefaultPeer(db)
	}
	return GetPeer(db, name)
}

// isUniqueViolation checks if an error is a SQLite unique constraint violation
func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "UNIQUE constraint failed") ||
		contains(err.Error(), "unique constraint"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
