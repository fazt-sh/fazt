package handlers

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/hosting"
)

// SiteDetailHandler returns details for a single site
func SiteDetailHandler(w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("id")
	if siteID == "" {
		api.BadRequest(w, "site_id required")
		return
	}

	// Reuse ListSites logic but filter? Or add GetSite to hosting package?
	// For now, list and find. Ideally hosting package should have GetSite(id).
	sites, err := hosting.ListSites()
	if err != nil {
		api.ServerError(w, err)
		return
	}

	var site *hosting.SiteInfo
	for _, s := range sites {
		if s.Name == siteID {
			site = &s
			break
		}
	}

	if site == nil {
		api.NotFound(w, "Site not found")
		return
	}

	api.JSON(w, http.StatusOK, site, nil)
}

// SiteFilesHandler returns the file tree for a site
func SiteFilesHandler(w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("id")
	if siteID == "" {
		api.BadRequest(w, "site_id required")
		return
	}

	if !hosting.SiteExists(siteID) {
		api.NotFound(w, "Site not found")
		return
	}

	fs := hosting.GetFileSystem()
	files, err := fs.ListFiles(siteID)
	if err != nil {
		api.ServerError(w, err)
		return
	}

	api.JSON(w, http.StatusOK, files, nil)
}

// SiteFileContentHandler returns the content of a file
func SiteFileContentHandler(w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("id")
	filePath := r.PathValue("path") // This might not capture slashes?
	
	// Go 1.22 wildcard {path...} captures remaining path
	// Route should be /api/sites/{id}/files/{path...}
	
	if siteID == "" || filePath == "" {
		api.BadRequest(w, "site_id and path required")
		return
	}

	fs := hosting.GetFileSystem()
	file, err := fs.ReadFile(siteID, filePath)
	if err != nil {
		api.NotFound(w, "File not found")
		return
	}
	defer file.Content.Close()

	// Detect Content-Type
	contentType := file.MimeType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}
	w.Header().Set("Content-Type", contentType)
	
	io.Copy(w, file.Content)
}
