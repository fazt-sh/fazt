package remote

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// OldConfig represents the legacy ~/.fazt/config.json format
type OldConfig struct {
	Version       int                  `json:"version"`
	DefaultServer string               `json:"default_server"`
	Servers       map[string]OldServer `json:"servers"`
}

// OldServer represents a server in the old config format
type OldServer struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// MigrateOldConfig imports servers from ~/.fazt/config.json into the peers table.
// This should be called on startup after the database is initialized.
// If migration succeeds, the old file is renamed to config.json.migrated.
func MigrateOldConfig(db *sql.DB) error {
	// Find old config file
	home := os.Getenv("HOME")
	if home == "" {
		return nil // Can't find home, skip migration
	}

	oldPath := filepath.Join(home, ".fazt", "config.json")
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil // No old config, nothing to migrate
	}

	// Read old config
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return err
	}

	var oldCfg OldConfig
	if err := json.Unmarshal(data, &oldCfg); err != nil {
		log.Printf("Warning: Failed to parse old config at %s: %v", oldPath, err)
		return nil // Don't fail on parse error, just skip
	}

	if len(oldCfg.Servers) == 0 {
		return nil // Nothing to migrate
	}

	// Import each server as a peer
	imported := 0
	for name, server := range oldCfg.Servers {
		// Check if peer already exists
		_, err := GetPeer(db, name)
		if err == nil {
			// Already exists, skip
			continue
		}

		// Add peer
		if err := AddPeer(db, name, server.URL, server.Token, "Migrated from ~/.fazt/config.json"); err != nil {
			log.Printf("Warning: Failed to migrate server '%s': %v", name, err)
			continue
		}

		// Set as default if it was the default
		if name == oldCfg.DefaultServer {
			SetDefaultPeer(db, name)
		}

		imported++
	}

	if imported == 0 {
		return nil // Nothing new imported
	}

	// Rename old file to indicate migration completed
	migratedPath := oldPath + ".migrated"
	if err := os.Rename(oldPath, migratedPath); err != nil {
		log.Printf("Warning: Failed to rename old config: %v", err)
		// Continue anyway, import succeeded
	}

	log.Printf("Migrated %d servers from %s to peers table", imported, oldPath)
	return nil
}
