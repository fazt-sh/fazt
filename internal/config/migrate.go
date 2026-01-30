package config

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// MigrateFromFile checks for legacy config files and imports them into the database.
// After successful migration, the original files are renamed to .bak.
//
// This should be called after database initialization but before loading config.
func MigrateFromFile(db *sql.DB) error {
	// Try to migrate instance config from ~/.config/fazt/config.json
	if err := migrateInstanceConfig(db); err != nil {
		log.Printf("Warning: failed to migrate instance config: %v", err)
		// Continue anyway - this is not fatal
	}

	return nil
}

// migrateInstanceConfig migrates ~/.config/fazt/config.json to the configurations table
func migrateInstanceConfig(db *sql.DB) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil // Can't find home, skip migration
	}

	configPath := filepath.Join(home, ".config", "fazt", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // No config file, nothing to migrate
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// Parse the JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("Warning: Failed to parse config at %s: %v", configPath, err)
		return nil // Don't fail on parse error
	}

	// Check if we've already migrated (by checking if any config exists in DB)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM configurations WHERE key LIKE 'server.%' OR key LIKE 'auth.%'").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Printf("Config already exists in database, skipping migration")
		return nil
	}

	// Import config into database
	store := NewDBConfigStore(db)
	imported := 0

	// Server config
	if cfg.Server.Port != "" {
		store.Set("server.port", cfg.Server.Port)
		imported++
	}
	if cfg.Server.Domain != "" {
		store.Set("server.domain", cfg.Server.Domain)
		imported++
	}
	if cfg.Server.Env != "" {
		store.Set("server.env", cfg.Server.Env)
		imported++
	}

	// Auth config
	if cfg.Auth.Username != "" {
		store.Set("auth.username", cfg.Auth.Username)
		imported++
	}
	if cfg.Auth.PasswordHash != "" {
		store.Set("auth.password_hash", cfg.Auth.PasswordHash)
		imported++
	}

	// Ntfy config
	if cfg.Ntfy.Topic != "" {
		store.Set("ntfy.topic", cfg.Ntfy.Topic)
		imported++
	}
	if cfg.Ntfy.URL != "" {
		store.Set("ntfy.url", cfg.Ntfy.URL)
		imported++
	}

	// HTTPS config
	if cfg.HTTPS.Enabled {
		store.Set("https.enabled", "true")
		imported++
	}
	if cfg.HTTPS.Email != "" {
		store.Set("https.email", cfg.HTTPS.Email)
		imported++
	}
	if cfg.HTTPS.Staging {
		store.Set("https.staging", "true")
		imported++
	}

	// API Key config
	if cfg.APIKey.Token != "" {
		store.Set("api_key.token", cfg.APIKey.Token)
		imported++
	}
	if cfg.APIKey.Name != "" {
		store.Set("api_key.name", cfg.APIKey.Name)
		imported++
	}

	if imported == 0 {
		return nil // Nothing imported
	}

	// Rename original file to .bak
	bakPath := configPath + ".bak"
	if err := os.Rename(configPath, bakPath); err != nil {
		log.Printf("Warning: Failed to rename config file to .bak: %v", err)
		// Continue anyway, import succeeded
	}

	log.Printf("Migrated %d config values from %s to database", imported, configPath)
	return nil
}

// LoadFromDB loads configuration from the database.
// Returns a Config struct populated from DB values, with defaults for missing values.
func LoadFromDB(db *sql.DB) (*Config, error) {
	store := NewDBConfigStore(db)
	dbConfig, err := store.Load()
	if err != nil {
		return nil, err
	}

	// Start with defaults
	cfg := CreateDefaultConfig()

	// Apply DB values
	applyDBMap(cfg, dbConfig)

	return cfg, nil
}

// SaveToDB saves the current configuration to the database.
func SaveToDB(db *sql.DB, cfg *Config) error {
	store := NewDBConfigStore(db)

	// Server config
	if err := store.Set("server.port", cfg.Server.Port); err != nil {
		return err
	}
	if err := store.Set("server.domain", cfg.Server.Domain); err != nil {
		return err
	}
	if err := store.Set("server.env", cfg.Server.Env); err != nil {
		return err
	}

	// Auth config
	if err := store.Set("auth.username", cfg.Auth.Username); err != nil {
		return err
	}
	if err := store.Set("auth.password_hash", cfg.Auth.PasswordHash); err != nil {
		return err
	}

	// Ntfy config
	if cfg.Ntfy.Topic != "" {
		if err := store.Set("ntfy.topic", cfg.Ntfy.Topic); err != nil {
			return err
		}
	}
	if cfg.Ntfy.URL != "" {
		if err := store.Set("ntfy.url", cfg.Ntfy.URL); err != nil {
			return err
		}
	}

	// HTTPS config
	if cfg.HTTPS.Enabled {
		if err := store.Set("https.enabled", "true"); err != nil {
			return err
		}
	}
	if cfg.HTTPS.Email != "" {
		if err := store.Set("https.email", cfg.HTTPS.Email); err != nil {
			return err
		}
	}
	if cfg.HTTPS.Staging {
		if err := store.Set("https.staging", "true"); err != nil {
			return err
		}
	}

	// API Key config
	if cfg.APIKey.Token != "" {
		if err := store.Set("api_key.token", cfg.APIKey.Token); err != nil {
			return err
		}
	}
	if cfg.APIKey.Name != "" {
		if err := store.Set("api_key.name", cfg.APIKey.Name); err != nil {
			return err
		}
	}

	return nil
}
