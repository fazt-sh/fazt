package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// WildcardDNSProviders is the list of wildcard DNS services to try, in order.
// These services resolve any subdomain containing an IP to that IP.
// e.g., app.192.168.64.3.nip.io resolves to 192.168.64.3
var WildcardDNSProviders = []string{
	"nip.io",
	"sslip.io",
	// "wdns.fazt.sh", // future: self-hosted
}

// ipv4Pattern matches IPv4 addresses
var ipv4Pattern = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)

// IsIPAddress checks if the given string is an IPv4 address
func IsIPAddress(s string) bool {
	// Remove port if present
	host := s
	if idx := strings.LastIndex(s, ":"); idx != -1 {
		// Check if it's not IPv6
		if !strings.Contains(s, "[") {
			host = s[:idx]
		}
	}
	return ipv4Pattern.MatchString(host) && net.ParseIP(host) != nil
}

// WrapWithWildcardDNS wraps an IP address with a working wildcard DNS provider.
// Returns the original domain if it's not an IP or if no provider works.
func WrapWithWildcardDNS(domain string) string {
	// Extract just the IP/hostname part (remove protocol and port)
	cleanDomain := domain
	if strings.HasPrefix(cleanDomain, "http://") {
		cleanDomain = strings.TrimPrefix(cleanDomain, "http://")
	}
	if strings.HasPrefix(cleanDomain, "https://") {
		cleanDomain = strings.TrimPrefix(cleanDomain, "https://")
	}
	if idx := strings.Index(cleanDomain, "/"); idx != -1 {
		cleanDomain = cleanDomain[:idx]
	}
	if idx := strings.LastIndex(cleanDomain, ":"); idx != -1 {
		cleanDomain = cleanDomain[:idx]
	}

	// Only wrap if it's an IP address
	if !IsIPAddress(cleanDomain) {
		return domain
	}

	// Try each provider until one resolves
	for _, provider := range WildcardDNSProviders {
		testDomain := fmt.Sprintf("test.%s.%s", cleanDomain, provider)
		ips, err := net.LookupIP(testDomain)
		if err == nil && len(ips) > 0 {
			wrapped := fmt.Sprintf("%s.%s", cleanDomain, provider)
			log.Printf("Wildcard DNS: %s → %s (via %s)", cleanDomain, wrapped, provider)
			return wrapped
		}
	}

	// No provider worked, return original
	log.Printf("Warning: No wildcard DNS provider available for %s", cleanDomain)
	return domain
}

// Version holds the current application version
var Version = "0.29.0"

// Config holds all configuration for the application
type Config struct {
	Server ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Auth AuthConfig     `json:"auth"`
	Ntfy NtfyConfig     `json:"ntfy"`
	APIKey APIKeyConfig `json:"api_key,omitempty"`
	HTTPS  HTTPSConfig  `json:"https"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port   string `json:"port"`
	Domain string `json:"domain"`
	Env    string `json:"env"` // development/production
}

// HTTPSConfig holds automatic HTTPS configuration
type HTTPSConfig struct {
	Enabled bool   `json:"enabled"`
	Email   string `json:"email"` // ACME contact email
	Staging bool   `json:"staging"` // Use Let's Encrypt Staging
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string `json:"path"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"` // bcrypt hash
}

// NtfyConfig holds notification configuration
type NtfyConfig struct {
	Topic string `json:"topic"`
	URL   string `json:"url"`
}

// APIKeyConfig holds API key configuration for deployment
type APIKeyConfig struct {
	Token string `json:"token,omitempty"`
	Name  string `json:"name,omitempty"`
}

var appConfig *Config

// CLIFlags holds command-line flags for temporary overrides.
// The database is the source of truth. CLI flags override DB values.
type CLIFlags struct {
	DBPath   string
	Port     string
	Domain   string
	Username string
	Password string
}

// ParseFlags parses command-line flags
func ParseFlags() *CLIFlags {
	flags := &CLIFlags{}
	flag.StringVar(&flags.DBPath, "db", "", "Database file path")
	flag.StringVar(&flags.Port, "port", "", "Server port")
	flag.StringVar(&flags.Username, "username", "", "Admin username")
	flag.StringVar(&flags.Password, "password", "", "Admin password")
	flag.Parse()
	return flags
}

// Load initializes config with defaults and resolves DB path.
// Config priority: CLI flags > Database > Defaults
// The database is the source of truth. Use LoadFromDB after database init.
func Load(flags *CLIFlags) (*Config, error) {
	if appConfig != nil {
		return appConfig, nil
	}

	// Start with defaults
	cfg := CreateDefaultConfig()

	// Resolve database path: Flag > Env > Default
	dbPath := "./data.db"
	if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
		dbPath = envPath
	}
	if flags.DBPath != "" {
		dbPath = ExpandPath(flags.DBPath)
	}
	cfg.Database.Path = ExpandPath(dbPath)

	appConfig = cfg
	return appConfig, nil
}

// CreateDefaultConfig creates a default configuration (exported for use in main.go)
func CreateDefaultConfig() *Config {
	// Default to local directory for "Cartridge" experience
	defaultDBPath := "./data.db"

	return &Config{
		Server: ServerConfig{
			Port:   "4698",
			Domain: "https://fazt.sh",
			Env:    "development",
		},
		Database: DatabaseConfig{
			Path: defaultDBPath,
		},
		Auth: AuthConfig{
			Username:     "",
			PasswordHash: "",
		},
		Ntfy: NtfyConfig{
			Topic: "",
			URL:   "https://ntfy.sh",
		},
		HTTPS: HTTPSConfig{
			Enabled: false,
			Email:   "",
			Staging: true,
		},
	}
}

// applyCLIFlags applies CLI flags to config (highest priority)
func applyCLIFlags(cfg *Config, flags *CLIFlags) {
	if flags.Port != "" {
		cfg.Server.Port = flags.Port
	}
	if flags.Domain != "" {
		cfg.Server.Domain = flags.Domain
	}
	if flags.DBPath != "" {
		cfg.Database.Path = ExpandPath(flags.DBPath)
	}
}

// ExpandPath expands ~ to home directory (exported for use in main.go)
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}
	return path
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate port is a number between 1-65535
	port, err := strconv.Atoi(c.Server.Port)
	if err != nil {
		return fmt.Errorf("invalid port: %s (must be a number)", c.Server.Port)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", port)
	}

	// Validate environment
	if c.Server.Env != "development" && c.Server.Env != "production" {
		return fmt.Errorf("invalid environment: %s (must be 'development' or 'production')", c.Server.Env)
	}

	// Ensure DB path is set
	if c.Database.Path == "" {
		return errors.New("database path cannot be empty")
	}

	// Expand database path
	c.Database.Path = ExpandPath(c.Database.Path)

	// Validate auth config (v0.4.0: auth always required)
	if c.Auth.Username == "" {
		return errors.New("auth username is required")
	}
	if c.Auth.PasswordHash == "" {
		return errors.New("auth password hash is required")
	}

	// Validate HTTPS
	if c.HTTPS.Enabled {
		if c.HTTPS.Email == "" {
			return errors.New("https email is required when https is enabled")
		}
	}

	return nil
}

// Get returns the loaded configuration
func Get() *Config {
	if appConfig == nil {
		log.Fatal("Configuration not loaded. Call Load() first.")
	}
	return appConfig
}

// SetConfig sets the application configuration (primarily for testing)
func SetConfig(cfg *Config) {
	appConfig = cfg
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// DebugMode returns true if debug logging is enabled.
// Debug mode is enabled when:
// - FAZT_DEBUG=1 environment variable is set (explicit)
// - OR running in development mode (implicit)
func (c *Config) DebugMode() bool {
	// Explicit env var takes precedence
	if debug := os.Getenv("FAZT_DEBUG"); debug != "" {
		return debug == "1" || debug == "true"
	}
	// Default to on in development mode
	return c.IsDevelopment()
}

// Legacy helpers for backward compatibility
func (c *Config) Port() string {
	return c.Server.Port
}

func (c *Config) DBPath() string {
	return c.Database.Path
}

func (c *Config) NtfyTopic() string {
	return c.Ntfy.Topic
}

func (c *Config) NtfyURL() string {
	return c.Ntfy.URL
}

func (c *Config) Environment() string {
	return c.Server.Env
}

// GetAPIKey returns the stored API key token
func (c *Config) GetAPIKey() string {
	return c.APIKey.Token
}

// SetAPIKey stores the API key token and name in config
func (c *Config) SetAPIKey(token, name string) {
	c.APIKey.Token = token
	c.APIKey.Name = name
}
