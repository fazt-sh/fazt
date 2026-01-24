package provision

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"text/template"

	"github.com/fazt-sh/fazt/internal/term"
)

// System service template (production, requires root)
const serviceTemplate = `[Unit]
Description=Fazt PaaS
After=network.target

[Service]
Type=simple
User={{.User}}
WorkingDirectory=/home/{{.User}}/.config/fazt
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
ExecStart={{.BinaryPath}} server start
Restart=always
LimitNOFILE=4096
Environment=FAZT_ENV=production
# Security hardening
ProtectSystem=strict
ReadWritePaths=/usr/local/bin
PrivateTmp=true

[Install]
WantedBy=multi-user.target
`

// User service template (local development, no root needed)
const userServiceTemplate = `[Unit]
Description=Fazt Local Development Server
After=network.target

[Service]
Type=simple
ExecStart={{.BinaryPath}} server start --port {{.Port}} --domain {{.Domain}} --db {{.DBPath}}
Restart=always
RestartSec=5
WorkingDirectory={{.WorkDir}}
Environment=FAZT_ENV=development

[Install]
WantedBy=default.target
`

type ServiceConfig struct {
	User       string
	BinaryPath string
}

// UserServiceConfig holds config for user-level services
type UserServiceConfig struct {
	BinaryPath string
	Port       string
	Domain     string
	DBPath     string
	WorkDir    string
}

// InstallSystemdService generates and installs the systemd unit file
func InstallSystemdService(serviceName string, config ServiceConfig) error {
	// check if systemd exists
	if _, err := exec.LookPath("systemctl"); err != nil {
		return fmt.Errorf("systemd (systemctl) not found")
	}

	tmpl, err := template.New("service").Parse(serviceTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse service template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return fmt.Errorf("failed to execute service template: %w", err)
	}

	servicePath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	term.Info("Writing service file to %s...", servicePath)

	if err := os.WriteFile(servicePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	return nil
}

// EnableAndStartService reloads systemd, enables and starts the service
func EnableAndStartService(serviceName string) error {
	commands := [][]string{
		{"systemctl", "daemon-reload"},
		{"systemctl", "enable", serviceName},
		{"systemctl", "start", serviceName},
	}

	term.Step("Starting system service...")

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("command failed: %s\nOutput: %s", args, string(output))
		}
	}

	return nil
}

// Systemctl runs a systemctl command for the service
func Systemctl(command, serviceName string) error {
	cmd := exec.Command("systemctl", command, serviceName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ServiceLogs follows the service logs
func ServiceLogs(serviceName string) error {
	cmd := exec.Command("journalctl", "-u", serviceName, "-f")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// UserServiceLogs follows user service logs
func UserServiceLogs(serviceName string) error {
	cmd := exec.Command("journalctl", "--user", "-u", serviceName, "-f")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// InstallUserService creates a user-level systemd service (no root required)
func InstallUserService(serviceName string, config UserServiceConfig) error {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return fmt.Errorf("systemd (systemctl) not found")
	}

	// Get user's systemd directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	serviceDir := filepath.Join(homeDir, ".config", "systemd", "user")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd user directory: %w", err)
	}

	tmpl, err := template.New("userservice").Parse(userServiceTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse user service template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return fmt.Errorf("failed to execute user service template: %w", err)
	}

	servicePath := filepath.Join(serviceDir, serviceName+".service")
	term.Info("Writing user service file to %s...", servicePath)

	if err := os.WriteFile(servicePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write user service file: %w", err)
	}

	return nil
}

// EnableAndStartUserService reloads user systemd, enables and starts the service
func EnableAndStartUserService(serviceName string) error {
	commands := [][]string{
		{"systemctl", "--user", "daemon-reload"},
		{"systemctl", "--user", "enable", serviceName},
		{"systemctl", "--user", "start", serviceName},
	}

	term.Step("Starting user service...")

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("command failed: %s\nOutput: %s", args, string(output))
		}
	}

	return nil
}

// EnableLinger enables systemd linger for the current user
// This allows user services to run even when the user is not logged in
func EnableLinger() error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	cmd := exec.Command("loginctl", "enable-linger", currentUser.Username)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to enable linger: %s\nOutput: %s", err, string(output))
	}

	term.Success("Enabled linger for user %s (service will run on boot)", currentUser.Username)
	return nil
}

// UserSystemctl runs a systemctl --user command
func UserSystemctl(command, serviceName string) error {
	cmd := exec.Command("systemctl", "--user", command, serviceName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// IsUserServiceRunning checks if a user service is active
func IsUserServiceRunning(serviceName string) bool {
	cmd := exec.Command("systemctl", "--user", "is-active", "--quiet", serviceName)
	return cmd.Run() == nil
}

// IsSystemServiceRunning checks if a system service is active
func IsSystemServiceRunning(serviceName string) bool {
	cmd := exec.Command("systemctl", "is-active", "--quiet", serviceName)
	return cmd.Run() == nil
}

// StopUserService stops a user service
func StopUserService(serviceName string) error {
	return UserSystemctl("stop", serviceName)
}

// RestartUserService restarts a user service
func RestartUserService(serviceName string) error {
	return UserSystemctl("restart", serviceName)
}

// GetUserServiceStatus returns the status of a user service
func GetUserServiceStatus(serviceName string) (string, error) {
	cmd := exec.Command("systemctl", "--user", "is-active", serviceName)
	output, _ := cmd.Output()
	return string(output), nil
}

