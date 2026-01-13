package runtime

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ServerlessHandler handles requests to /api/* paths by executing JavaScript.
type ServerlessHandler struct {
	runtime *Runtime
	db      *sql.DB
}

// NewServerlessHandler creates a new serverless handler.
func NewServerlessHandler(db *sql.DB) *ServerlessHandler {
	return &ServerlessHandler{
		runtime: NewRuntime(MaxPoolSize, DefaultTimeout),
		db:      db,
	}
}

// NewServerlessHandlerWithRuntime creates a handler with a custom runtime.
func NewServerlessHandlerWithRuntime(db *sql.DB, rt *Runtime) *ServerlessHandler {
	return &ServerlessHandler{
		runtime: rt,
		db:      db,
	}
}

// HandleRequest handles a serverless request for a specific app.
func (h *ServerlessHandler) HandleRequest(w http.ResponseWriter, r *http.Request, appID, appName string) {
	ctx := r.Context()

	// Load api/main.js from the app's files
	mainJS, err := h.loadFile(appID, "api/main.js")
	if err != nil {
		// No serverless handler found
		http.Error(w, "No serverless handler found", http.StatusNotFound)
		return
	}

	// Build request object
	req := buildRequest(r)

	// Create file loader for require()
	loader := func(path string) (string, error) {
		return h.loadFile(appID, path)
	}

	// Load environment variables for the app
	env := h.loadEnvVars(appID)

	// Create app context
	app := &AppContext{
		ID:   appID,
		Name: appName,
	}

	// Execute with a timeout
	execCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := h.executeWithFazt(execCtx, mainJS, req, loader, app, env)

	// Handle errors
	if result.Error != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": result.Error.Error(),
			"logs":  result.Logs,
		})
		return
	}

	// Write response
	if result.Response == nil {
		result.Response = &Response{Status: 200}
	}

	// Set headers
	for k, v := range result.Response.Headers {
		w.Header().Set(k, v)
	}

	// Set content type if not set
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	w.WriteHeader(result.Response.Status)

	// Write body
	if result.Response.Body != nil {
		if str, ok := result.Response.Body.(string); ok {
			w.Write([]byte(str))
		} else {
			json.NewEncoder(w).Encode(result.Response.Body)
		}
	}
}

// executeWithFazt executes code with the fazt namespace injected.
func (h *ServerlessHandler) executeWithFazt(ctx context.Context, code string, req *Request, loader FileLoader, app *AppContext, env EnvVars) *ExecuteResult {
	// For now, use ExecuteWithFiles which has the core functionality
	// The fazt namespace needs to be added to the runtime
	return h.runtime.ExecuteWithFiles(ctx, code, req, loader)
}

// loadFile loads a file from the VFS for a given app.
func (h *ServerlessHandler) loadFile(appID, path string) (string, error) {
	var content string
	err := h.db.QueryRow(`
		SELECT content FROM files
		WHERE site_id = ? AND path = ?
	`, appID, path).Scan(&content)
	if err != nil {
		return "", err
	}
	return content, nil
}

// loadEnvVars loads environment variables for an app.
func (h *ServerlessHandler) loadEnvVars(appID string) EnvVars {
	env := make(EnvVars)
	rows, err := h.db.Query(`
		SELECT key, value FROM env_vars
		WHERE site_id = ?
	`, appID)
	if err != nil {
		return env
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err == nil {
			env[key] = value
		}
	}
	return env
}

// buildRequest creates a Request from an HTTP request.
func buildRequest(r *http.Request) *Request {
	// Parse query parameters
	query := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			query[k] = v[0]
		}
	}

	// Parse headers
	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// Parse body for JSON requests
	var body interface{}
	if r.Method != "GET" && r.Method != "HEAD" {
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			json.NewDecoder(r.Body).Decode(&body)
		}
	}

	return &Request{
		Method:  r.Method,
		Path:    r.URL.Path,
		Query:   query,
		Headers: headers,
		Body:    body,
	}
}

// IsServerlessPath returns true if the path should be handled by serverless.
func IsServerlessPath(path string) bool {
	return strings.HasPrefix(path, "/api/")
}
