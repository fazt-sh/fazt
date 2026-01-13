package provision

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// GetServiceDBPath returns the database path used by the fazt systemd service.
// It reads the service file to find the WorkingDirectory and returns the path to data.db.
// If the service isn't installed or can't be read, it returns an empty string.
func GetServiceDBPath() string {
	// Read the service file
	serviceFile := "/etc/systemd/system/fazt.service"
	data, err := os.ReadFile(serviceFile)
	if err != nil {
		return ""
	}

	// Extract WorkingDirectory
	re := regexp.MustCompile(`WorkingDirectory=(.+)`)
	matches := re.FindSubmatch(data)
	if len(matches) < 2 {
		return ""
	}

	workDir := strings.TrimSpace(string(matches[1]))
	return filepath.Join(workDir, "data.db")
}

// GetEffectiveDBPath returns the database path to use, with this priority:
// 1. Explicit --db flag (if provided and non-empty)
// 2. FAZT_DB_PATH environment variable
// 3. Service's database path (if fazt service is installed)
// 4. Default ./data.db
func GetEffectiveDBPath(explicitPath string) string {
	// 1. Explicit flag
	if explicitPath != "" {
		return explicitPath
	}

	// 2. Environment variable
	if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
		return envPath
	}

	// 3. Service database path
	if servicePath := GetServiceDBPath(); servicePath != "" {
		// Verify the path exists or the directory exists
		if _, err := os.Stat(servicePath); err == nil {
			return servicePath
		}
		// Check if directory exists (DB might not exist yet)
		dir := filepath.Dir(servicePath)
		if _, err := os.Stat(dir); err == nil {
			return servicePath
		}
	}

	// 4. Default
	return "./data.db"
}

// IsServiceRunning checks if the fazt systemd service is currently running
func IsServiceRunning() bool {
	cmd := exec.Command("systemctl", "is-active", "--quiet", "fazt")
	return cmd.Run() == nil
}
