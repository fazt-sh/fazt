package provision

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/term"
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

	term.Section("fazt.sh System Installer")
	term.Info("Installing for domain: %s", opts.Domain)

	// 1. Check Ports
	if opts.HTTPS {
		if err := checkPortsAvailable([]string{"80", "443"}); err != nil {
			return err
		}
	} else {
		// Even if not HTTPS, we install on port 80 by default in production
		if err := checkPortsAvailable([]string{"80"}); err != nil {
			return err
		}
	}

	// 2. Ensure User
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
	
	term.Step("Configuring environment...")

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

	// Port logic
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

	term.Section("Installation Complete")
	term.Success("Fazt is now running at https://%s", opts.Domain)
	fmt.Println()
	
	// Display Credentials Box
	fmt.Println(term.Yellow + "╔══════════════════════════════════════════════════════════╗" + term.Reset)
	fmt.Printf(term.Yellow+"║ %-56s ║"+term.Reset+"\n", "ADMIN CREDENTIALS (SAVE THESE!)")
	fmt.Println(term.Yellow + "╠══════════════════════════════════════════════════════════╣" + term.Reset)
	fmt.Printf(term.Yellow+"║ %-10s %-45s ║"+term.Reset+"\n", "Username:", opts.AdminUser)
	fmt.Printf(term.Yellow+"║ %-10s %-45s ║"+term.Reset+"\n", "Password:", opts.AdminPassword)
	fmt.Println(term.Yellow + "╚══════════════════════════════════════════════════════════╝" + term.Reset)
	fmt.Println()
	term.Print(term.Dim + "Login at: " + term.Reset + "https://" + opts.Domain + "/login")
	fmt.Println()
	
	return nil
}

// checkPortsAvailable checks if the required ports are free
func checkPortsAvailable(ports []string) error {
	term.Step("Checking port availability...")
	for _, port := range ports {
		ln, err := net.Listen("tcp", ":"+port)
		if err != nil {
			term.Error("Port %s is already in use!", port)
			term.Warn("Common culprits: nginx, apache2, caddy")
			term.Warn("Try stopping them: systemctl stop nginx")
			return fmt.Errorf("port %s unavailable: %w", port, err)
		}
		ln.Close()
	}
	term.Success("Ports %v are available", ports)
	return nil
}
