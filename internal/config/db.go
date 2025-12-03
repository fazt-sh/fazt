package config

import (
	"database/sql"
	"fmt"
	"log"
)

// DBConfigStore handles database operations for configuration
type DBConfigStore struct {
	db *sql.DB
}

// NewDBConfigStore creates a new store
func NewDBConfigStore(db *sql.DB) *DBConfigStore {
	return &DBConfigStore{db: db}
}

// Load reads all configurations from the database
func (s *DBConfigStore) Load() (map[string]string, error) {
	query := "SELECT key, value FROM configurations"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query configurations: %w", err)
	}
	defer rows.Close()

	config := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		config[key] = value
	}
	return config, nil
}

// Set updates or inserts a configuration value
func (s *DBConfigStore) Set(key, value string) error {
	query := `
		INSERT INTO configurations (key, value, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := s.db.Exec(query, key, value)
	return err
}

// OverlayDB loads config from DB and overlays it onto the current appConfig.
// It effectively merges: Defaults < ConfigFile < Env < DB < Flags
// Note: We re-apply flags after DB load to ensure Flags have highest priority.
func OverlayDB(db *sql.DB, flags *CLIFlags) error {
	if appConfig == nil {
		return fmt.Errorf("config not initialized, call Load() first")
	}

	store := NewDBConfigStore(db)
	dbConfig, err := store.Load()
	if err != nil {
		// If table doesn't exist yet (very first run before migration), this might fail.
		// However, OverlayDB should be called AFTER database.Init(), which runs migrations.
		return err
	}

	applyDBMap(appConfig, dbConfig)
	
	// RE-APPLY flags to ensure they override DB values
	applyCLIFlags(appConfig, flags)
	
	// Validate again to ensure the combination is valid
	if err := appConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration after DB overlay: %w", err)
	}

	log.Printf("Configuration overlaid from Database")
	return nil
}

// applyDBMap maps flat keys to Config struct fields
func applyDBMap(cfg *Config, data map[string]string) {
	for k, v := range data {
		switch k {
		// Server
		case "server.port":
			cfg.Server.Port = v
		case "server.domain":
			cfg.Server.Domain = v
		case "server.env":
			cfg.Server.Env = v
		
		// Auth
		case "auth.username":
			cfg.Auth.Username = v
		case "auth.password_hash":
			cfg.Auth.PasswordHash = v
			
		// Ntfy
		case "ntfy.topic":
			cfg.Ntfy.Topic = v
		case "ntfy.url":
			cfg.Ntfy.URL = v

		// HTTPS
		case "https.enabled":
			cfg.HTTPS.Enabled = (v == "true")
		case "https.email":
			cfg.HTTPS.Email = v
		case "https.staging":
			cfg.HTTPS.Staging = (v == "true")

		// API Key
		case "api_key.token":
			cfg.APIKey.Token = v
		case "api_key.name":
			cfg.APIKey.Name = v
		}
	}
}
