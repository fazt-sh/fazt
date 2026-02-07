package remote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Client provides HTTP communication with a remote fazt peer
type Client struct {
	peer   *Peer
	client *http.Client
}

// NewClient creates a new client for communicating with a peer
func NewClient(peer *Peer) *Client {
	return &Client{
		peer: peer,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// StatusResponse represents the /api/system/health response
type StatusResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Mode    string `json:"mode"`
	Uptime  float64 `json:"uptime_seconds"`
	Memory  struct {
		UsedMB  float64 `json:"used_mb"`
		LimitMB float64 `json:"limit_mb"`
	} `json:"memory"`
	Database struct {
		OpenConnections int `json:"open_connections"`
		InUse           int `json:"in_use"`
	} `json:"database"`
	Runtime struct {
		Goroutines int `json:"goroutines"`
	} `json:"runtime"`
}

// App represents an app on the remote server
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

// UpgradeResponse represents the /api/upgrade response
type UpgradeResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	CurrentVersion string `json:"current_version"`
	NewVersion     string `json:"new_version,omitempty"`
	Action         string `json:"action,omitempty"`
}

// DeployResponse represents the /api/deploy response
type DeployResponse struct {
	Site      string `json:"site"`
	FileCount int    `json:"file_count"`
	SizeBytes int64  `json:"size_bytes"`
	Message   string `json:"message"`
}

// APIResponse wraps the standard API response format
type APIResponse struct {
	Data  json.RawMessage `json:"data"`
	Error *APIError       `json:"error,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Status checks the health of the remote peer
func (c *Client) Status() (*StatusResponse, error) {
	resp, err := c.doRequest("GET", "/api/system/health", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
	}

	var status StatusResponse
	if err := json.Unmarshal(apiResp.Data, &status); err != nil {
		return nil, fmt.Errorf("failed to decode status: %w", err)
	}

	return &status, nil
}

// HealthCheck performs a simple health check
func (c *Client) HealthCheck() (bool, error) {
	resp, err := c.doRequest("GET", "/api/system/health", nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// Apps lists all apps on the remote peer
func (c *Client) Apps() ([]App, error) {
	resp, err := c.doRequest("GET", "/api/apps", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
	}

	var apps []App
	if err := json.Unmarshal(apiResp.Data, &apps); err != nil {
		return nil, fmt.Errorf("failed to decode apps: %w", err)
	}

	return apps, nil
}

// Upgrade checks for or performs an upgrade
func (c *Client) Upgrade(checkOnly bool) (*UpgradeResponse, error) {
	return c.UpgradeWithURL(checkOnly, "")
}

// UpgradeWithURL checks for or performs an upgrade, optionally from a custom URL
func (c *Client) UpgradeWithURL(checkOnly bool, customURL string) (*UpgradeResponse, error) {
	path := "/api/upgrade"
	params := []string{}

	if checkOnly {
		params = append(params, "check=true")
	}
	if customURL != "" {
		params = append(params, "url="+url.QueryEscape(customURL))
	}

	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}

	resp, err := c.doRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
	}

	var upgrade UpgradeResponse
	if err := json.Unmarshal(apiResp.Data, &upgrade); err != nil {
		return nil, fmt.Errorf("failed to decode upgrade response: %w", err)
	}

	return &upgrade, nil
}

// Deploy deploys a ZIP file to the remote peer
func (c *Client) Deploy(zipPath, siteName string) (*DeployResponse, error) {
	// Open the zip file
	file, err := os.Open(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add site_name field
	if err := writer.WriteField("site_name", siteName); err != nil {
		return nil, fmt.Errorf("failed to write site_name: %w", err)
	}

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(zipPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.peer.URL+"/api/deploy", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.peer.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.peer.Token)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
	}

	var deploy DeployResponse
	if err := json.Unmarshal(apiResp.Data, &deploy); err != nil {
		return nil, fmt.Errorf("failed to decode deploy response: %w", err)
	}

	return &deploy, nil
}

// DeployOptions configures deployment behavior
type DeployOptions struct {
	SPA bool // Enable SPA routing (clean URLs)
}

// DeployWithOptions deploys a ZIP file with additional options
func (c *Client) DeployWithOptions(zipPath, siteName string, opts *DeployOptions) (*DeployResponse, error) {
	// Open the zip file
	file, err := os.Open(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add site_name field
	if err := writer.WriteField("site_name", siteName); err != nil {
		return nil, fmt.Errorf("failed to write site_name: %w", err)
	}

	// Add spa field if enabled
	if opts != nil && opts.SPA {
		if err := writer.WriteField("spa", "true"); err != nil {
			return nil, fmt.Errorf("failed to write spa: %w", err)
		}
	}

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(zipPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.peer.URL+"/api/deploy", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.peer.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.peer.Token)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
	}

	var deploy DeployResponse
	if err := json.Unmarshal(apiResp.Data, &deploy); err != nil {
		return nil, fmt.Errorf("failed to decode deploy response: %w", err)
	}

	return &deploy, nil
}

// DeleteApp deletes an app from the remote peer
func (c *Client) DeleteApp(name string) error {
	resp, err := c.doRequest("DELETE", "/api/apps/"+name, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
	}

	return nil
}

// SourceInfo contains source tracking information
type SourceInfo struct {
	Type   string `json:"type"`
	URL    string `json:"url"`
	Ref    string `json:"ref"`
	Commit string `json:"commit"`
}

// FileEntry represents a file in an app
type FileEntry struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

// DeployWithSource deploys a ZIP file with source tracking info
func (c *Client) DeployWithSource(zipPath, siteName string, source *SourceInfo) (*DeployResponse, error) {
	// Open the zip file
	file, err := os.Open(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add site_name field
	if err := writer.WriteField("site_name", siteName); err != nil {
		return nil, fmt.Errorf("failed to write site_name: %w", err)
	}

	// Add source tracking fields if provided
	if source != nil {
		writer.WriteField("source_type", source.Type)
		writer.WriteField("source_url", source.URL)
		writer.WriteField("source_ref", source.Ref)
		writer.WriteField("source_commit", source.Commit)
	}

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(zipPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.peer.URL+"/api/deploy", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.peer.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.peer.Token)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
	}

	var deploy DeployResponse
	if err := json.Unmarshal(apiResp.Data, &deploy); err != nil {
		return nil, fmt.Errorf("failed to decode deploy response: %w", err)
	}

	return &deploy, nil
}

// GetAppSource gets the source tracking info for an app
func (c *Client) GetAppSource(name string) (*SourceInfo, error) {
	resp, err := c.doRequest("GET", "/api/apps/"+name+"/source", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
	}

	var source SourceInfo
	if err := json.Unmarshal(apiResp.Data, &source); err != nil {
		return nil, fmt.Errorf("failed to decode source: %w", err)
	}

	return &source, nil
}

// GetAppFiles lists files in an app
func (c *Client) GetAppFiles(name string) ([]FileEntry, error) {
	resp, err := c.doRequest("GET", "/api/apps/"+name+"/files", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
	}

	var files []FileEntry
	if err := json.Unmarshal(apiResp.Data, &files); err != nil {
		return nil, fmt.Errorf("failed to decode files: %w", err)
	}

	return files, nil
}

// GetAppFileContent gets the content of a specific file
func (c *Client) GetAppFileContent(appName, filePath string) ([]byte, error) {
	resp, err := c.doRequest("GET", "/api/apps/"+appName+"/files/"+filePath, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// File content is returned directly, not wrapped in API response
	if resp.StatusCode != http.StatusOK {
		var apiResp APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err == nil && apiResp.Error != nil {
			return nil, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
		}
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// ProviderConfig represents an OAuth provider configuration
type ProviderConfig struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Enabled     bool   `json:"enabled"`
	ClientID    string `json:"client_id,omitempty"`
	Configured  bool   `json:"configured,omitempty"`
}

// ListAuthProviders lists configured auth providers
func (c *Client) ListAuthProviders() ([]ProviderConfig, error) {
	resp, err := c.doRequest("GET", "/api/auth/providers", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
	}

	var providers []ProviderConfig
	if err := json.Unmarshal(apiResp.Data, &providers); err != nil {
		return nil, fmt.Errorf("failed to decode providers: %w", err)
	}

	return providers, nil
}

// ConfigureAuthProvider configures an OAuth provider
func (c *Client) ConfigureAuthProvider(name, clientID, clientSecret string, enable *bool) (*ProviderConfig, error) {
	reqBody := map[string]interface{}{}
	if clientID != "" && clientSecret != "" {
		reqBody["client_id"] = clientID
		reqBody["client_secret"] = clientSecret
	}
	if enable != nil {
		reqBody["enable"] = *enable
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest("PUT", "/api/auth/providers/"+name, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("%s: %s", apiResp.Error.Code, apiResp.Error.Message)
	}

	var cfg ProviderConfig
	if err := json.Unmarshal(apiResp.Data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode provider config: %w", err)
	}

	return &cfg, nil
}

// doRequest performs an authenticated HTTP request
func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.peer.URL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.peer.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.peer.Token)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// SQLRequest represents a SQL query request
type SQLRequest struct {
	Query string `json:"query"`
	Write bool   `json:"write"`
	Limit int    `json:"limit"`
}

// SQLResponse represents a SQL query response
type SQLResponse struct {
	Columns  []string        `json:"columns,omitempty"`
	Rows     [][]interface{} `json:"rows,omitempty"`
	Count    int             `json:"count,omitempty"`
	Affected int64           `json:"affected,omitempty"`
	TimeMS   int64           `json:"time_ms"`
}

// SQL executes a SQL query on the remote peer
func (c *Client) SQL(query string, write bool, limit int) (*SQLResponse, error) {
	reqBody := SQLRequest{
		Query: query,
		Write: write,
		Limit: limit,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", "/api/sql", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SQL request failed: %s", string(bodyBytes))
	}

	var sqlResp SQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&sqlResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &sqlResp, nil
}
