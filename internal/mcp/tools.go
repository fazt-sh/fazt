package mcp

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// registerDefaultTools registers all built-in MCP tools.
func (s *Server) registerDefaultTools() {
	s.RegisterTool(Tool{
		Name:        "fazt_servers_list",
		Description: "List all configured Fazt servers",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: s.handleServersList,
	})

	s.RegisterTool(Tool{
		Name:        "fazt_apps_list",
		Description: "List all apps/sites on a Fazt server",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"server": map[string]interface{}{
					"type":        "string",
					"description": "Server name (omit for default server)",
				},
			},
		},
		Handler: s.handleAppsList,
	})

	s.RegisterTool(Tool{
		Name:        "fazt_deploy",
		Description: "Deploy files to create or update an app",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"server": map[string]interface{}{
					"type":        "string",
					"description": "Server name (omit for default server)",
				},
				"app_name": map[string]interface{}{
					"type":        "string",
					"description": "Name for the app (becomes subdomain)",
				},
				"files": map[string]interface{}{
					"type":        "object",
					"description": "Map of file paths to content",
					"additionalProperties": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"required": []string{"app_name", "files"},
		},
		Handler: s.handleDeploy,
	})

	s.RegisterTool(Tool{
		Name:        "fazt_app_delete",
		Description: "Delete an app from a server",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"server": map[string]interface{}{
					"type":        "string",
					"description": "Server name (omit for default server)",
				},
				"app_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the app to delete",
				},
			},
			"required": []string{"app_name"},
		},
		Handler: s.handleAppDelete,
	})

	s.RegisterTool(Tool{
		Name:        "fazt_system_status",
		Description: "Get server status and health information",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"server": map[string]interface{}{
					"type":        "string",
					"description": "Server name (omit for default server)",
				},
			},
		},
		Handler: s.handleSystemStatus,
	})
}

// handleServersList returns the list of configured servers.
func (s *Server) handleServersList(params map[string]interface{}) (interface{}, error) {
	servers := s.config.ListServers()

	result := make([]map[string]interface{}, len(servers))
	for i, srv := range servers {
		result[i] = map[string]interface{}{
			"name":       srv.Name,
			"url":        srv.URL,
			"is_default": srv.IsDefault,
		}
	}

	return map[string]interface{}{
		"servers": result,
		"count":   len(servers),
	}, nil
}

// handleAppsList returns apps on a server.
func (s *Server) handleAppsList(params map[string]interface{}) (interface{}, error) {
	serverName, _ := params["server"].(string)

	srv, name, err := s.config.GetServer(serverName)
	if err != nil {
		return nil, err
	}

	// Call server's /api/sites endpoint
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", srv.URL+"/api/sites", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+srv.Token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Success bool                     `json:"success"`
		Data    []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return map[string]interface{}{
		"server": name,
		"apps":   apiResp.Data,
		"count":  len(apiResp.Data),
	}, nil
}

// handleDeploy deploys files to create or update an app.
func (s *Server) handleDeploy(params map[string]interface{}) (interface{}, error) {
	serverName, _ := params["server"].(string)
	appName, ok := params["app_name"].(string)
	if !ok || appName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	files, ok := params["files"].(map[string]interface{})
	if !ok || len(files) == 0 {
		return nil, fmt.Errorf("files is required and must not be empty")
	}

	srv, name, err := s.config.GetServer(serverName)
	if err != nil {
		return nil, err
	}

	// Create ZIP in memory
	zipBuffer, fileCount, totalSize, err := createZipFromFiles(files)
	if err != nil {
		return nil, fmt.Errorf("failed to create zip: %w", err)
	}

	// Create multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("site_name", appName); err != nil {
		return nil, fmt.Errorf("failed to write form field: %w", err)
	}

	part, err := writer.CreateFormFile("file", "deploy.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, zipBuffer); err != nil {
		return nil, fmt.Errorf("failed to write zip to form: %w", err)
	}

	writer.Close()

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("POST", srv.URL+"/api/deploy", &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+srv.Token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("deploy failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return map[string]interface{}{
		"server":     name,
		"app_name":   appName,
		"file_count": fileCount,
		"size_bytes": totalSize,
		"success":    true,
	}, nil
}

// handleAppDelete deletes an app from a server.
func (s *Server) handleAppDelete(params map[string]interface{}) (interface{}, error) {
	serverName, _ := params["server"].(string)
	appName, ok := params["app_name"].(string)
	if !ok || appName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	srv, name, err := s.config.GetServer(serverName)
	if err != nil {
		return nil, err
	}

	// First, get the site ID
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", srv.URL+"/api/sites", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+srv.Token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	var sitesResp struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sitesResp); err != nil {
		return nil, fmt.Errorf("failed to parse sites response: %w", err)
	}

	var siteID string
	for _, site := range sitesResp.Data {
		if site.Name == appName {
			siteID = site.ID
			break
		}
	}

	if siteID == "" {
		return nil, fmt.Errorf("app '%s' not found on server '%s'", appName, name)
	}

	// Delete the site
	delReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/sites/%s", srv.URL, siteID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create delete request: %w", err)
	}
	delReq.Header.Set("Authorization", "Bearer "+srv.Token)

	delResp, err := client.Do(delReq)
	if err != nil {
		return nil, fmt.Errorf("failed to delete: %w", err)
	}
	defer delResp.Body.Close()

	if delResp.StatusCode != http.StatusOK && delResp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(delResp.Body)
		return nil, fmt.Errorf("delete failed with status %d: %s", delResp.StatusCode, string(body))
	}

	return map[string]interface{}{
		"server":   name,
		"app_name": appName,
		"deleted":  true,
	}, nil
}

// handleSystemStatus returns server health information.
func (s *Server) handleSystemStatus(params map[string]interface{}) (interface{}, error) {
	serverName, _ := params["server"].(string)

	srv, name, err := s.config.GetServer(serverName)
	if err != nil {
		return nil, err
	}

	// Try health endpoint
	client := &http.Client{Timeout: 10 * time.Second}
	healthResp, err := client.Get(srv.URL + "/health")
	if err != nil {
		return map[string]interface{}{
			"server":  name,
			"url":     srv.URL,
			"healthy": false,
			"error":   err.Error(),
		}, nil
	}
	defer healthResp.Body.Close()

	healthy := healthResp.StatusCode == http.StatusOK

	// Try to get system info if authenticated
	req, _ := http.NewRequest("GET", srv.URL+"/api/system/health", nil)
	req.Header.Set("Authorization", "Bearer "+srv.Token)

	infoResp, err := client.Do(req)
	if err == nil && infoResp.StatusCode == http.StatusOK {
		defer infoResp.Body.Close()
		var info map[string]interface{}
		if json.NewDecoder(infoResp.Body).Decode(&info) == nil {
			info["server"] = name
			info["healthy"] = healthy
			return info, nil
		}
	}

	return map[string]interface{}{
		"server":  name,
		"url":     srv.URL,
		"healthy": healthy,
	}, nil
}

// createZipFromFiles creates a ZIP archive from a map of files.
func createZipFromFiles(files map[string]interface{}) (*bytes.Buffer, int, int64, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	var totalSize int64
	fileCount := 0

	for path, content := range files {
		contentStr, ok := content.(string)
		if !ok {
			continue
		}

		fw, err := w.Create(path)
		if err != nil {
			return nil, 0, 0, err
		}

		n, err := fw.Write([]byte(contentStr))
		if err != nil {
			return nil, 0, 0, err
		}

		totalSize += int64(n)
		fileCount++
	}

	if err := w.Close(); err != nil {
		return nil, 0, 0, err
	}

	return buf, fileCount, totalSize, nil
}
