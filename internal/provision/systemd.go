package provision

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

const serviceTemplate = `[Unit]
Description=Fazt PaaS
After=network.target

[Service]
Type=simple
User={{.User}}
ExecStart={{.BinaryPath}} server start
Restart=always
LimitNOFILE=4096
Environment=FAZT_ENV=production
# Security hardening
ProtectSystem=full
PrivateTmp=true

[Install]
WantedBy=multi-user.target
`

type ServiceConfig struct {
	User       string
	BinaryPath string
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
	fmt.Printf("Writing service file to %s...\n", servicePath)

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

	for _, args := range commands {
		fmt.Printf("Running: %s %s...\n", args[0], args[1])
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
		
		
