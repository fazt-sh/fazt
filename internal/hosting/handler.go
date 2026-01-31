package hosting

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// ServeVFSByAppID serves files from the Virtual File System using app_id
func ServeVFSByAppID(w http.ResponseWriter, r *http.Request, appID string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Default to index.html for root or directories
	if path == "/" || strings.HasSuffix(path, "/") {
		path += "index.html"
	}

	// Clean path
	path = filepath.Clean(path)
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "index.html"
	}

	sqlFS, ok := fs.(*SQLFileSystem)
	if !ok {
		http.NotFound(w, r)
		return
	}

	// 1. Try exact match using app_id
	file, err := sqlFS.ReadFileByAppID(appID, path)
	if err != nil {
		// 2. If not found, try with index.html
		if filepath.Ext(path) == "" {
			idxPath := filepath.Join(path, "index.html")
			idxPath = filepath.ToSlash(idxPath)
			file, err = sqlFS.ReadFileByAppID(appID, idxPath)
		}

		// 3. If still not found, 404
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}
	defer file.Content.Close()

	// ETag Caching
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, file.Hash))
	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, file.Hash) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Content Type
	contentType := file.MimeType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(path))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}
	w.Header().Set("Content-Type", contentType)

	// Cache-Control: Smart caching strategy
	// 1. HTML files: Always revalidate (for live reload & version detection)
	// 2. Hashed assets (/assets/*-*.ext): Cache forever (content-addressed)
	// 3. Other files: Short cache (5 minutes)
	if strings.HasSuffix(path, ".html") {
		// HTML: no-cache means "revalidate with server before using cached version"
		// This ensures version checks work reliably
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	} else if strings.HasPrefix(path, "assets/") && strings.Contains(filepath.Base(path), "-") {
		// Hashed assets: cache aggressively (filename changes when content changes)
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		// Everything else: short cache (5 minutes)
		w.Header().Set("Cache-Control", "public, max-age=300")
	}

	// Content Length
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))

	// Serve content
	if _, err := io.Copy(w, file.Content); err != nil {
		// Log error?
	}
}

// ServeVFS serves files from the Virtual File System
func ServeVFS(w http.ResponseWriter, r *http.Request, siteID string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Normalize: redirect trailing slash to non-trailing slash (except root)
	// e.g., /about/ -> /about (301 redirect for SEO consistency)
	if path != "/" && strings.HasSuffix(path, "/") {
		target := strings.TrimSuffix(path, "/")
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
		return
	}

	// Track if original path looks like a route (no extension) for SPA fallback
	isRouteLikePath := filepath.Ext(path) == ""

	// Default to index.html for root
	if path == "/" {
		path = "/index.html"
	}

	// Clean path
	path = filepath.Clean(path)
	// Ensure consistent forward slashes
	path = filepath.ToSlash(path)
	// Remove leading slash for DB lookup if stored without it
	// In deploy.go we used filepath.Clean/ToSlash, which usually removes leading slash for relative paths?
	// zip files usually don't have leading slash.
	// Let's ensure we strip leading slash.
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "index.html"
	}

	// 1. Try exact match
	file, err := fs.ReadFile(siteID, path)
	if err != nil {
		// 2. If not found, and it looks like a directory (no extension), try appending index.html
		if filepath.Ext(path) == "" {
			idxPath := filepath.Join(path, "index.html")
			idxPath = filepath.ToSlash(idxPath)
			file, err = fs.ReadFile(siteID, idxPath)
		}

		// 3. If still not found, check for SPA fallback
		if err != nil {
			// SPA fallback: if original path looked like a route (no extension) and app has SPA enabled
			if isRouteLikePath {
				if sqlFS, ok := fs.(*SQLFileSystem); ok {
					if spa, spaErr := sqlFS.GetAppSPA(siteID); spaErr == nil && spa {
						file, err = fs.ReadFile(siteID, "index.html")
					}
				}
			}

			// 4. If still not found, 404
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
	}
	defer file.Content.Close()

	// ETag Caching
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, file.Hash))
	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, file.Hash) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Content Type
	contentType := file.MimeType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(path))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}
	w.Header().Set("Content-Type", contentType)

	// Cache-Control: Smart caching strategy
	// 1. HTML files: Always revalidate (for live reload & version detection)
	// 2. Hashed assets (/assets/*-*.ext): Cache forever (content-addressed)
	// 3. Other files: Short cache (5 minutes)
	if strings.HasSuffix(path, ".html") {
		// HTML: no-cache means "revalidate with server before using cached version"
		// This ensures version checks work reliably
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
	} else if strings.HasPrefix(path, "assets/") && strings.Contains(filepath.Base(path), "-") {
		// Hashed assets: cache aggressively (filename changes when content changes)
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		// Everything else: short cache (5 minutes)
		w.Header().Set("Cache-Control", "public, max-age=300")
	}

	// Content Length
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))

	// Serve content
	if _, err := io.Copy(w, file.Content); err != nil {
		// Log error?
	}
}
