// Package clientconfig handles client-side configuration for the fazt CLI.
// Configuration is stored in ~/.fazt/config.json and supports multiple servers.
package clientconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the client configuration file structure.
type Config struct {
	Version       int               `json:"version"`
	DefaultServer string            `json:"default_server,omitempty"`
	Servers       map[string]Server `json:"servers"`
}

// Server represents a configured Fazt server.
type Server struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// ConfigDir returns the path to the config directory (~/.fazt).
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home dir unavailable
		return ".fazt"
	}
	return filepath.Join(home, ".fazt")
}

// ConfigPath returns the full path to the config file (~/.fazt/config.json).
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// Load reads the config file from disk.
// Returns an empty config if the file doesn't exist.
func Load() (*Config, error) {
	data, err := os.ReadFile(ConfigPath())
	if os.IsNotExist(err) {
		// Return empty config if file doesn't exist
		return &Config{
			Version: 1,
			Servers: make(map[string]Server),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Initialize servers map if nil
	if cfg.Servers == nil {
		cfg.Servers = make(map[string]Server)
	}

	return &cfg, nil
}

// Save writes the config to disk.
// Creates the config directory if it doesn't exist.
func (c *Config) Save() error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(ConfigPath(), data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// AddServer adds or updates a server configuration.
func (c *Config) AddServer(name, url, token string) error {
	if name == "" {
		return fmt.Errorf("server name is required")
	}
	if url == "" {
		return fmt.Errorf("server URL is required")
	}
	if token == "" {
		return fmt.Errorf("server token is required")
	}

	c.Servers[name] = Server{
		URL:   url,
		Token: token,
	}

	// If this is the first server, make it the default
	if len(c.Servers) == 1 {
		c.DefaultServer = name
	}

	return nil
}

// RemoveServer removes a server from the configuration.
func (c *Config) RemoveServer(name string) error {
	if _, exists := c.Servers[name]; !exists {
		return fmt.Errorf("server '%s' not found", name)
	}

	delete(c.Servers, name)

	// Clear default if we removed it
	if c.DefaultServer == name {
		c.DefaultServer = ""
		// If there's only one server left, make it the default
		if len(c.Servers) == 1 {
			for n := range c.Servers {
				c.DefaultServer = n
			}
		}
	}

	return nil
}

// SetDefault sets the default server.
func (c *Config) SetDefault(name string) error {
	if _, exists := c.Servers[name]; !exists {
		return fmt.Errorf("server '%s' not found", name)
	}
	c.DefaultServer = name
	return nil
}

// GetServer returns the server configuration for the given name.
// If name is empty, it uses smart defaults:
//   - If a default is set, use it
//   - If only one server exists, use it
//   - Otherwise, return an error
func (c *Config) GetServer(name string) (*Server, string, error) {
	// Explicit server specified
	if name != "" {
		srv, ok := c.Servers[name]
		if !ok {
			return nil, "", fmt.Errorf("server '%s' not found", name)
		}
		return &srv, name, nil
	}

	// No server specified - try smart defaults

	// Case 1: Default is set -> use it
	if c.DefaultServer != "" {
		srv, ok := c.Servers[c.DefaultServer]
		if ok {
			return &srv, c.DefaultServer, nil
		}
		// Default is set but doesn't exist (stale config)
		// Fall through to other cases
	}

	// Case 2: Only ONE server configured -> use it (no need for --server)
	if len(c.Servers) == 1 {
		for name, srv := range c.Servers {
			return &srv, name, nil
		}
	}

	// Case 3: Zero servers
	if len(c.Servers) == 0 {
		return nil, "", fmt.Errorf("no servers configured\nRun: fazt servers add <name> --url <url> --token <token>")
	}

	// Case 4: Multiple servers, no default
	return nil, "", fmt.Errorf("multiple servers configured, specify --to <server> or set default:\n  fazt servers default <name>")
}

// ListServers returns a list of all configured servers with their names.
func (c *Config) ListServers() []struct {
	Name      string
	URL       string
	IsDefault bool
} {
	result := make([]struct {
		Name      string
		URL       string
		IsDefault bool
	}, 0, len(c.Servers))

	for name, srv := range c.Servers {
		result = append(result, struct {
			Name      string
			URL       string
			IsDefault bool
		}{
			Name:      name,
			URL:       srv.URL,
			IsDefault: name == c.DefaultServer,
		})
	}

	return result
}

// HasServer checks if a server with the given name exists.
func (c *Config) HasServer(name string) bool {
	_, exists := c.Servers[name]
	return exists
}

// ServerCount returns the number of configured servers.
func (c *Config) ServerCount() int {
	return len(c.Servers)
}
