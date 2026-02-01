package database

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultDBPath is the default database path when none is specified.
// Uses ~/.fazt/data.db for a consistent location regardless of CWD.
// Override with --db flag or FAZT_DB_PATH environment variable.
const DefaultDBPath = "~/.fazt/data.db"

// ResolvePath determines the database path using the priority:
// 1. Explicit path argument (--db flag)
// 2. FAZT_DB_PATH environment variable
// 3. Default: ~/.fazt/data.db
//
// This is the single source of truth for DB path resolution.
// All CLI commands and server startup should use this function.
func ResolvePath(explicit string) string {
	// 1. Explicit path has highest priority
	if explicit != "" {
		return expandPath(explicit)
	}

	// 2. Environment variable
	if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
		return expandPath(envPath)
	}

	// 3. Default (expand ~ in default path)
	return expandPath(DefaultDBPath)
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
