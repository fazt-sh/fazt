package provision

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/fazt-sh/fazt/internal/config"
	"golang.org/x/crypto/bcrypt"
)

type InstallOptions struct {
	User          string
	Domain        string
	Email         string
	AdminUser     string
	AdminPassword string
	HTTPS         bool
}

// RunInstall orchestrates the installation process
func RunInstall(opts InstallOptions) error {
	// 0. Check Root
	if os.Geteuid() != 0 {
		return fmt.Errorf("this command must be run as root (use sudo)")
	}

	fmt.Println("Starting installation...")

	// 1. Ensure User
	if err := EnsureUser(opts.User); err != nil {
		return err
	}

	targetUser, err := user.Lookup(opts.User)
	if err != nil {
		return fmt.Errorf("failed to lookup user %s: %w", opts.User, err)
	}
	uid, _ := strconv.Atoi(targetUser.Uid)
	gid, _ := strconv.Atoi(targetUser.Gid)

	// 2. Install Binary
	targetBin := "/usr/local/bin/fazt"
	if err := InstallBinary(targetBin); err != nil {
		return err
	}

	// 3. Set Capabilities
	if err := SetCapabilities(targetBin); err != nil {
		return err
	}

	// 4. Configure
	configDir := filepath.Join(targetUser.HomeDir, ".config", "fazt")
	configPath := filepath.Join(configDir, "config.json")
	
	fmt.Printf("Creating configuration at %s...\n", configPath)

	// Create directory with correct permissions
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	if err := os.Chown(configDir, uid, gid); err != nil {
		return fmt.Errorf("failed to chown config dir: %w", err)
	}

	// Generate Config
	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(opts.AdminPassword), 12)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:   "80", // Default for installed service
			Domain: opts.Domain,
			Env:    "production",
		},
		Database: config.DatabaseConfig{
			Path: filepath.Join(configDir, "data.db"),
		},
		Auth: config.AuthConfig{
			Username:     opts.AdminUser,
			PasswordHash: string(passwordHash),
		},
		HTTPS: config.HTTPSConfig{
			Enabled: opts.HTTPS,
			Email:   opts.Email,
		},
		Ntfy: config.NtfyConfig{
			URL: "https://ntfy.sh",
		},
	}

	// If HTTPS is disabled, use default port 4698?
	// The plan says "Bind to ports 80 and 443 automatically".
	// If HTTPS is false, maybe we still want port 80?
	// Or maybe the user wants to run behind a proxy.
	// But "install" usually implies "take over the machine".
	// Let's stick to 80 if HTTPS is false, or maybe just respect the passed config.
	// For now, let's hardcode 80/443 logic:
	if opts.HTTPS {
		cfg.Server.Port = "443" // CertMagic handles 80 too
	} else {
		cfg.Server.Port = "80"
	}

	// Save Config
	if err := config.SaveToFile(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Chown the config file
	if err := os.Chown(configPath, uid, gid); err != nil {
		return fmt.Errorf("failed to chown config file: %w", err)
	}

	// 5. Systemd Service
	svcConfig := ServiceConfig{
		User:       opts.User,
		BinaryPath: targetBin,
	}
	if err := InstallSystemdService("fazt", svcConfig); err != nil {
		return err
	}

	// 6. Start Service
	if err := EnableAndStartService("fazt"); err != nil {
		return err
	}

	fmt.Println("Installation complete!")
	fmt.Printf("Server running at %s\n", opts.Domain)
	return nil
}
