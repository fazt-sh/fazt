package config

import (
	"database/sql"
	"fmt"
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

// LoadFromDB loads config from SQLite database and applies CLI flag overrides.
// Config priority: CLI flags > Database > Defaults
// The database is the source of truth. CLI flags are for temporary overrides.
func LoadFromDB(db *sql.DB, flags *CLIFlags) error {
	if appConfig == nil {
		return fmt.Errorf("config not initialized")
	}

	store := NewDBConfigStore(db)
	dbConfig, err := store.Load()
	if err != nil {
		return err
	}

	// Apply DB config
	applyDBMap(appConfig, dbConfig)

	// Apply CLI flags (highest priority - for temporary overrides)
	applyCLIFlags(appConfig, flags)

	if err := appConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

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
