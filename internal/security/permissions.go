package security

import (
	"fmt"
	"log"
	"os"
)

// CheckFilePermissions verifies that sensitive files have proper permissions
func CheckFilePermissions(path string, expectedPerms os.FileMode) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's okay
		}
		return fmt.Errorf("failed to check file permissions: %w", err)
	}

	actualPerms := info.Mode().Perm()
	if actualPerms != expectedPerms {
		log.Printf("WARNING: %s has permissions %o, should be %o", path, actualPerms, expectedPerms)
		log.Printf("Attempting to fix permissions...")

		if err := os.Chmod(path, expectedPerms); err != nil {
			return fmt.Errorf("failed to set permissions: %w", err)
		}

		log.Printf("✓ Fixed permissions for %s", path)
	}

	return nil
}

// EnsureSecurePermissions ensures database files have secure permissions
func EnsureSecurePermissions(dbPath string) {
	// Database file should be 0600 (owner read/write only)
	if err := CheckFilePermissions(dbPath, 0600); err != nil {
		log.Printf("Warning: Could not secure database file permissions: %v", err)
	}

	// WAL and SHM files too
	CheckFilePermissions(dbPath+"-wal", 0600)
	CheckFilePermissions(dbPath+"-shm", 0600)
}
