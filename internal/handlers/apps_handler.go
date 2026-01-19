package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/assets"
	"github.com/fazt-sh/fazt/internal/build"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/git"
	"github.com/fazt-sh/fazt/internal/hosting"
)

// App represents an app with metadata
type App struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Source    string      `json:"source"`
	Manifest  interface{} `json:"manifest,omitempty"`
	FileCount int         `json:"file_count"`
	SizeBytes int64       `json:"size_bytes"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
}

// AppsListHandler returns the list of apps with metadata
func AppsListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Query apps with file stats
	query := `
		SELECT
			a.id,
			a.name,
			a.source,
			a.manifest,
			a.created_at,
			a.updated_at,
			COALESCE(COUNT(f.path), 0) as file_count,
			COALESCE(SUM(f.size_bytes), 0) as size_bytes
		FROM apps a
		LEFT JOIN files f ON a.name = f.site_id
		GROUP BY a.id
		ORDER BY a.updated_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer rows.Close()

	var apps []App
	for rows.Next() {
		var app App
		var manifest *string
		var createdAt, updatedAt interface{}

		err := rows.Scan(
			&app.ID,
			&app.Name,
			&app.Source,
			&manifest,
			&createdAt,
			&updatedAt,
			&app.FileCount,
			&app.SizeBytes,
		)
		if err != nil {
			continue
		}

		// Parse manifest if present
		if manifest != nil && *manifest != "" {
			json.Unmarshal([]byte(*manifest), &app.Manifest)
		}

		// Format timestamps
		if createdAt != nil {
			app.CreatedAt = formatTime(createdAt)
		}
		if updatedAt != nil {
			app.UpdatedAt = formatTime(updatedAt)
		}

		apps = append(apps, app)
	}

	api.Success(w, http.StatusOK, apps)
}

// AppDetailHandler returns details for a single app
func AppDetailHandler(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	if appID == "" {
		api.BadRequest(w, "app_id required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Query app by id or name
	query := `
		SELECT
			a.id,
			a.name,
			a.source,
			a.manifest,
			a.created_at,
			a.updated_at,
			COALESCE(COUNT(f.path), 0) as file_count,
			COALESCE(SUM(f.size_bytes), 0) as size_bytes
		FROM apps a
		LEFT JOIN files f ON a.name = f.site_id
		WHERE a.id = ? OR a.name = ?
		GROUP BY a.id
	`

	var app App
	var manifest *string
	var createdAt, updatedAt interface{}

	err := db.QueryRow(query, appID, appID).Scan(
		&app.ID,
		&app.Name,
		&app.Source,
		&manifest,
		&createdAt,
		&updatedAt,
		&app.FileCount,
		&app.SizeBytes,
	)
	if err != nil {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}

	// Parse manifest if present
	if manifest != nil && *manifest != "" {
		json.Unmarshal([]byte(*manifest), &app.Manifest)
	}

	// Format timestamps
	if createdAt != nil {
		app.CreatedAt = formatTime(createdAt)
	}
	if updatedAt != nil {
		app.UpdatedAt = formatTime(updatedAt)
	}

	api.Success(w, http.StatusOK, app)
}

// AppDeleteHandler deletes an app
func AppDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	appID := r.PathValue("id")
	if appID == "" {
		api.BadRequest(w, "app_id required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Get app title for file deletion
	var title string
	err := db.QueryRow("SELECT title FROM apps WHERE id = ? OR title = ?", appID, appID).Scan(&title)
	if err != nil {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}

	// Don't allow deleting system apps
	if title == "root" || title == "404" || title == "admin" {
		api.ErrorResponse(w, http.StatusForbidden, "SYSTEM_APP", "Cannot delete system app", "")
		return
	}

	// Delete files via hosting
	if err := hosting.DeleteSite(title); err != nil {
		api.InternalError(w, err)
		return
	}

	// Delete from apps table
	_, err = db.Exec("DELETE FROM apps WHERE id = ? OR title = ?", appID, appID)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, map[string]interface{}{
		"message": "App deleted",
		"name":    title,
	})
}

// AppFilesHandler returns the file tree for an app
func AppFilesHandler(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	if appID == "" {
		api.BadRequest(w, "app_id required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Get app title
	var title string
	err := db.QueryRow("SELECT title FROM apps WHERE id = ? OR title = ?", appID, appID).Scan(&title)
	if err != nil {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}

	fs := hosting.GetFileSystem()
	files, err := fs.ListFiles(title)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.Success(w, http.StatusOK, files)
}

// formatTime converts interface{} time to string
func formatTime(t interface{}) string {
	if s, ok := t.(string); ok {
		return s
	}
	return ""
}

// AppSourceHandler returns source tracking info for an app
func AppSourceHandler(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	if appID == "" {
		api.BadRequest(w, "app_id required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	query := `
		SELECT source, source_url, source_ref, source_commit
		FROM apps WHERE title = ? OR id = ?
	`

	var sourceType string
	var sourceURL, sourceRef, sourceCommit *string

	err := db.QueryRow(query, appID, appID).Scan(&sourceType, &sourceURL, &sourceRef, &sourceCommit)
	if err != nil {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}

	result := map[string]string{
		"type": sourceType,
	}
	if sourceURL != nil {
		result["url"] = *sourceURL
	}
	if sourceRef != nil {
		result["ref"] = *sourceRef
	}
	if sourceCommit != nil {
		result["commit"] = *sourceCommit
	}

	api.Success(w, http.StatusOK, result)
}

// AppFileContentHandler returns the content of a specific file
func AppFileContentHandler(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	filePath := r.PathValue("path")

	if appID == "" || filePath == "" {
		api.BadRequest(w, "app_id and path required")
		return
	}

	db := database.GetDB()
	if db == nil {
		api.InternalError(w, nil)
		return
	}

	// Get app title
	var title string
	err := db.QueryRow("SELECT title FROM apps WHERE id = ? OR title = ?", appID, appID).Scan(&title)
	if err != nil {
		api.NotFound(w, "APP_NOT_FOUND", "App not found")
		return
	}

	fs := hosting.GetFileSystem()
	file, err := fs.ReadFile(title, filePath)
	if err != nil {
		api.NotFound(w, "FILE_NOT_FOUND", "File not found")
		return
	}
	defer file.Content.Close()

	// Return raw content with correct MIME type
	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))
	w.WriteHeader(http.StatusOK)

	buf := make([]byte, file.Size)
	file.Content.Read(buf)
	w.Write(buf)
}

// InstallRequest is the request body for POST /api/apps/install
type InstallRequest struct {
	URL  string `json:"url"`  // GitHub URL
	Name string `json:"name"` // Optional name override
}

// AppInstallHandler installs an app from a git repository
func AppInstallHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	var req InstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if req.URL == "" {
		api.BadRequest(w, "url is required")
		return
	}

	// Parse git URL
	ref, err := git.ParseURL(req.URL)
	if err != nil {
		api.BadRequest(w, "invalid GitHub URL: "+err.Error())
		return
	}

	// Clone to temp directory
	tmpDir, err := os.MkdirTemp("", "fazt-install-*")
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer os.RemoveAll(tmpDir)

	result, err := git.Clone(git.CloneOptions{
		URL:       ref.FullURL(),
		Path:      ref.Path,
		Ref:       ref.Ref,
		TargetDir: tmpDir,
	})
	if err != nil {
		api.BadRequest(w, "clone failed: "+err.Error())
		return
	}

	// Determine app name
	appName := req.Name
	if appName == "" {
		manifest, err := readManifestFile(tmpDir)
		if err != nil {
			appName = ref.Repo
			if ref.Path != "" {
				appName = filepath.Base(ref.Path)
			}
		} else {
			appName = manifest.Name
		}
	}

	// Build step (server likely has no npm - will use existing or source)
	deployDir := tmpDir
	buildResult, err := build.Build(tmpDir, nil)
	if err != nil {
		if err == build.ErrBuildRequired {
			// Try pre-built branch
			prebuilt := git.FindPrebuiltBranch(ref.FullURL())
			if prebuilt != "" {
				os.RemoveAll(tmpDir)
				tmpDir, _ = os.MkdirTemp("", "fazt-install-*")
				result, err = git.Clone(git.CloneOptions{
					URL:       ref.FullURL(),
					Path:      ref.Path,
					Ref:       prebuilt,
					TargetDir: tmpDir,
				})
				if err != nil {
					api.BadRequest(w, "clone of pre-built branch failed: "+err.Error())
					return
				}
				ref.Ref = prebuilt
				deployDir = tmpDir
			} else {
				api.BadRequest(w, "app requires building; no package manager available and no pre-built branch found")
				return
			}
		} else {
			api.BadRequest(w, "build failed: "+err.Error())
			return
		}
	} else {
		deployDir = buildResult.OutputDir
	}

	// Create zip from directory
	zipData, err := createZipFromDir(deployDir)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Deploy locally
	sourceInfo := &hosting.SourceInfo{
		Type:   "git",
		URL:    req.URL,
		Ref:    ref.Ref,
		Commit: result.CommitSHA,
	}

	_, err = hosting.DeploySiteWithSource(zipReader, appName, sourceInfo)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	cfg := config.Get()
	api.Success(w, http.StatusCreated, map[string]interface{}{
		"name":   appName,
		"url":    fmt.Sprintf("https://%s.%s", appName, cfg.Server.Domain),
		"source": req.URL,
		"commit": result.CommitSHA[:7],
	})
}

// CreateRequest is the request body for POST /api/apps/create
type CreateRequest struct {
	Name     string `json:"name"`
	Template string `json:"template"` // "minimal" or "vite"
}

// AppCreateHandler creates a new app from a template
func AppCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		api.BadRequest(w, "name is required")
		return
	}

	// Validate app name
	if !isValidAppName(req.Name) {
		api.BadRequest(w, "invalid app name: use lowercase letters, numbers, and hyphens only")
		return
	}

	// Default template
	if req.Template == "" {
		req.Template = "minimal"
	}

	// Get template
	tmplFS, err := assets.GetTemplate(req.Template)
	if err != nil {
		api.BadRequest(w, "unknown template: "+req.Template)
		return
	}

	// Create in temp directory
	tmpDir, err := os.MkdirTemp("", "fazt-create-*")
	if err != nil {
		api.InternalError(w, err)
		return
	}
	defer os.RemoveAll(tmpDir)

	// Copy template with substitution
	data := map[string]string{"Name": req.Name}
	if err := copyTemplateFiles(tmplFS, tmpDir, data); err != nil {
		api.InternalError(w, err)
		return
	}

	// Create zip from directory
	zipData, err := createZipFromDir(tmpDir)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		api.InternalError(w, err)
		return
	}

	// Deploy locally
	sourceInfo := &hosting.SourceInfo{
		Type: "template",
		URL:  req.Template,
	}

	_, err = hosting.DeploySiteWithSource(zipReader, req.Name, sourceInfo)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	cfg := config.Get()
	api.Success(w, http.StatusCreated, map[string]interface{}{
		"name":     req.Name,
		"template": req.Template,
		"url":      fmt.Sprintf("https://%s.%s", req.Name, cfg.Server.Domain),
	})
}

// TemplatesListHandler returns available templates
func TemplatesListHandler(w http.ResponseWriter, r *http.Request) {
	templates := assets.ListTemplates()
	api.Success(w, http.StatusOK, templates)
}

// Manifest for reading manifest.json
type manifestFile struct {
	Name string `json:"name"`
}

func readManifestFile(dir string) (*manifestFile, error) {
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m manifestFile
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Name == "" {
		return nil, fmt.Errorf("manifest missing 'name' field")
	}
	return &m, nil
}

func isValidAppName(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}
	if name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

func copyTemplateFiles(tmplFS fs.FS, destDir string, data map[string]string) error {
	return fs.WalkDir(tmplFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}

		destPath := filepath.Join(destDir, path)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		content, err := fs.ReadFile(tmplFS, path)
		if err != nil {
			return err
		}

		// Apply template substitution
		tmpl, err := template.New(path).Parse(string(content))
		if err != nil {
			return os.WriteFile(destPath, content, 0644)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return os.WriteFile(destPath, content, 0644)
		}

		return os.WriteFile(destPath, buf.Bytes(), 0644)
	})
}

// createZipFromDir creates a zip file from a directory
func createZipFromDir(srcDir string) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and hidden files
		if info.IsDir() || info.Name()[0] == '.' {
			return nil
		}

		// Skip node_modules
		relPath, _ := filepath.Rel(srcDir, path)
		if len(relPath) > 12 && relPath[:12] == "node_modules" {
			return nil
		}

		// Create zip entry
		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Read and write file content
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
