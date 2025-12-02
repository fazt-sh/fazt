package hosting

import (
	"database/sql"
	"fmt"
	"mime"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fazt-sh/fazt/internal/assets"
)

var (
	// fs is the active file system
	fs FileSystem
	
	// db is the database connection
	database *sql.DB

	// validSubdomainRegex matches valid subdomain names
	validSubdomainRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
)

// Init initializes the hosting system
func Init(db *sql.DB) error {
	database = db

	// Initialize VFS
	fs = NewSQLFileSystem(db)

	// Seed system sites (root, 404)
	if err := EnsureSystemSites(); err != nil {
		return fmt.Errorf("failed to seed system sites: %w", err)
	}

	return nil
}

// EnsureSystemSites checks and seeds reserved sites from embedded assets
func EnsureSystemSites() error {
	sites := map[string]string{
		"root": "system/root",
		"404":  "system/404",
	}

	for siteID, assetDir := range sites {
		// Check if site exists (simple check for index.html)
		// We use fs.Exists directly to avoid overhead/ambiguity of SiteExists
		exists, _ := fs.Exists(siteID, "index.html")
		if !exists {
			// Seed from assets
			entries, err := assets.SystemFS.ReadDir(assetDir)
			if err != nil {
				return fmt.Errorf("failed to read asset dir %s: %w", assetDir, err)
			}

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				path := assetDir + "/" + entry.Name()
				content, err := assets.SystemFS.Open(path)
				if err != nil {
					return fmt.Errorf("failed to open asset %s: %w", path, err)
				}

				info, _ := entry.Info()
				mimeType := mime.TypeByExtension(filepath.Ext(entry.Name()))
				if mimeType == "" {
					mimeType = "application/octet-stream"
				}

				// Write to VFS
				// Use entry.Name() as path (flat structure for now)
				if err := fs.WriteFile(siteID, entry.Name(), content, info.Size(), mimeType); err != nil {
					content.Close()
					return fmt.Errorf("failed to write asset %s to VFS: %w", entry.Name(), err)
				}
				content.Close()
			}
			// Log to stdout so user knows
			fmt.Printf("✓ Seeded system site: %s\n", siteID)
		}
	}
	return nil
}

// GetFileSystem returns the active file system
func GetFileSystem() FileSystem {
	return fs
}

// SiteExists checks if a site directory exists
func SiteExists(subdomain string) bool {
	// Check VFS first
	exists, err := fs.Exists(subdomain, "index.html")
	if err == nil && exists {
		return true
	}
	// Check for main.js (serverless)
	exists, err = fs.Exists(subdomain, "main.js")
	if err == nil && exists {
		return true
	}
	
	return false
}

// ValidateSubdomain checks if a subdomain name is valid
func ValidateSubdomain(subdomain string) error {
	subdomain = strings.ToLower(subdomain)

	if len(subdomain) < 1 || len(subdomain) > 63 {
		return fmt.Errorf("subdomain must be 1-63 characters")
	}

	if !validSubdomainRegex.MatchString(subdomain) {
		return fmt.Errorf("subdomain must contain only lowercase letters, numbers, and hyphens, and cannot start or end with a hyphen")
	}

	// Reserved subdomains
	reserved := []string{"www", "api", "admin", "mail", "ftp", "smtp", "pop", "imap", "ns1", "ns2", "localhost"}
	for _, r := range reserved {
		if subdomain == r {
			return fmt.Errorf("'%s' is a reserved subdomain", subdomain)
		}
	}

	return nil
}

// ListSites returns all hosted sites
func ListSites() ([]SiteInfo, error) {
	if database == nil {
		return nil, fmt.Errorf("hosting not initialized")
	}

	query := `
		SELECT site_id, COUNT(*) as file_count, SUM(size_bytes) as total_size, MAX(updated_at) as last_mod
		FROM files
		GROUP BY site_id
		ORDER BY last_mod DESC
	`

	rows, err := database.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sites: %w", err)
	}
	defer rows.Close()

	var sites []SiteInfo
	for rows.Next() {
		var site SiteInfo
		var lastMod interface{}
		if err := rows.Scan(&site.Name, &site.FileCount, &site.SizeBytes, &lastMod); err != nil {
			continue
		}
		site.ModTime = lastMod
		site.Path = "vfs://" + site.Name
		sites = append(sites, site)
	}

	return sites, nil
}

// SiteInfo contains information about a hosted site
type SiteInfo struct {
	Name      string
	Path      string
	FileCount int
	SizeBytes int64
	ModTime   interface{} // time.Time
}

// CreateSite creates a new site (placeholder for VFS)
func CreateSite(subdomain string) error {
	return ValidateSubdomain(subdomain)
}

// DeleteSite removes a site and all its contents
func DeleteSite(subdomain string) error {
	// Clean up WebSocket hub
	RemoveHub(subdomain)

	// Delete from VFS
	return fs.DeleteSite(subdomain)
}
