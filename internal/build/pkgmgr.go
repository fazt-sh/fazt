package build

import (
	"os"
	"os/exec"
	"path/filepath"
)

// PackageManager represents a JavaScript package manager
type PackageManager struct {
	Name       string   // Human-readable name (e.g., "bun")
	Binary     string   // Executable name
	InstallCmd []string // Args for install (e.g., ["install"])
	BuildCmd   []string // Args for build (e.g., ["run", "build"])
	LockFile   string   // Lockfile name for detection
}

// PackageManagers in priority order (fastest first)
var PackageManagers = []PackageManager{
	{
		Name:       "bun",
		Binary:     "bun",
		InstallCmd: []string{"install"},
		BuildCmd:   []string{"run", "build"},
		LockFile:   "bun.lockb",
	},
	{
		Name:       "pnpm",
		Binary:     "pnpm",
		InstallCmd: []string{"install"},
		BuildCmd:   []string{"run", "build"},
		LockFile:   "pnpm-lock.yaml",
	},
	{
		Name:       "yarn",
		Binary:     "yarn",
		InstallCmd: []string{"install"},
		BuildCmd:   []string{"run", "build"},
		LockFile:   "yarn.lock",
	},
	{
		Name:       "npm",
		Binary:     "npm",
		InstallCmd: []string{"install"},
		BuildCmd:   []string{"run", "build"},
		LockFile:   "package-lock.json",
	},
}

// DetectPackageManager finds the best available package manager.
// Priority: lockfile match > first available
func DetectPackageManager(srcDir string) *PackageManager {
	// First, check for lockfiles (indicates project preference)
	for i := range PackageManagers {
		pm := &PackageManagers[i]
		lockPath := filepath.Join(srcDir, pm.LockFile)
		if _, err := os.Stat(lockPath); err == nil {
			// Lockfile exists, check if binary available
			if _, err := exec.LookPath(pm.Binary); err == nil {
				return pm
			}
		}
	}

	// No lockfile match, return first available
	for i := range PackageManagers {
		pm := &PackageManagers[i]
		if _, err := exec.LookPath(pm.Binary); err == nil {
			return pm
		}
	}

	return nil // No package manager available
}

// HasPackageManager checks if any package manager is available
func HasPackageManager() bool {
	for _, pm := range PackageManagers {
		if _, err := exec.LookPath(pm.Binary); err == nil {
			return true
		}
	}
	return false
}
