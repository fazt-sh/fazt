package mcp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/clientconfig"
)

func testConfig() *clientconfig.Config {
	return &clientconfig.Config{
		Version:       1,
		DefaultServer: "test",
		Servers: map[string]clientconfig.Server{
			"test": {
				URL:   "https://test.example.com",
				Token: "test-token-123",
			},
			"prod": {
				URL:   "https://prod.example.com",
				Token: "prod-token-456",
			},
		},
	}
}

func TestNewServerWithConfig(t *testing.T) {
	cfg := testConfig()
	s := NewServerWithConfig(cfg)

	if s == nil {
		t.Fatal("NewServerWithConfig returned nil")
	}

	if s.config != cfg {
		t.Error("config not set correctly")
	}

	// Check that default tools are registered
	tools := s.GetTools()
	if len(tools) == 0 {
		t.Error("no tools registered")
	}
}

func TestServer_RegisterTool(t *testing.T) {
	cfg := testConfig()
	s := NewServerWithConfig(cfg)

	initialCount := len(s.GetTools())

	s.RegisterTool(Tool{
		Name:        "custom_tool",
		Description: "A custom test tool",
		Handler: func(params map[string]interface{}) (interface{}, error) {
			return "success", nil
		},
	})

	if len(s.GetTools()) != initialCount+1 {
		t.Errorf("expected %d tools, got %d", initialCount+1, len(s.GetTools()))
	}
}

func TestServer_CallTool(t *testing.T) {
	cfg := testConfig()
	s := NewServerWithConfig(cfg)

	// Test calling fazt_servers_list (always available)
	result, err := s.CallTool("fazt_servers_list", nil)
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}

	if resultMap["count"] != 2 {
		t.Errorf("expected 2 servers, got %v", resultMap["count"])
	}
}

func TestServer_CallTool_NotFound(t *testing.T) {
	cfg := testConfig()
	s := NewServerWithConfig(cfg)

	_, err := s.CallTool("nonexistent_tool", nil)
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}
}

func TestServer_HandleInitialize(t *testing.T) {
	cfg := testConfig()
	s := NewServerWithConfig(cfg)

	reqBody := InitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities:    map[string]interface{}{},
		ClientInfo: ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/mcp/initialize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.HandleInitialize(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp InitializeResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ProtocolVersion != ProtocolVersion {
		t.Errorf("expected protocol version %s, got %s", ProtocolVersion, resp.ProtocolVersion)
	}

	if resp.ServerInfo.Name != ServerName {
		t.Errorf("expected server name %s, got %s", ServerName, resp.ServerInfo.Name)
	}
}

func TestServer_HandleInitialize_WrongMethod(t *testing.T) {
	cfg := testConfig()
	s := NewServerWithConfig(cfg)

	req := httptest.NewRequest("GET", "/mcp/initialize", nil)
	rr := httptest.NewRecorder()

	s.HandleInitialize(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestServer_HandleToolsList(t *testing.T) {
	cfg := testConfig()
	s := NewServerWithConfig(cfg)

	req := httptest.NewRequest("POST", "/mcp/tools/list", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.HandleToolsList(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp ToolsListResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Tools) == 0 {
		t.Error("expected at least one tool")
	}

	// Check for expected tools
	toolNames := make(map[string]bool)
	for _, tool := range resp.Tools {
		toolNames[tool.Name] = true
	}

	expectedTools := []string{
		"fazt_servers_list",
		"fazt_apps_list",
		"fazt_deploy",
		"fazt_app_delete",
		"fazt_system_status",
	}

	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Errorf("missing expected tool: %s", name)
		}
	}
}

func TestServer_HandleToolsCall(t *testing.T) {
	cfg := testConfig()
	s := NewServerWithConfig(cfg)

	reqBody := ToolCallRequest{
		Name:      "fazt_servers_list",
		Arguments: map[string]interface{}{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/mcp/tools/call", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.HandleToolsCall(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp ToolCallResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.IsError {
		t.Error("expected success, got error")
	}

	if len(resp.Content) == 0 {
		t.Error("expected content in response")
	}

	if resp.Content[0].Type != "text" {
		t.Errorf("expected text content, got %s", resp.Content[0].Type)
	}
}

func TestServer_HandleToolsCall_InvalidTool(t *testing.T) {
	cfg := testConfig()
	s := NewServerWithConfig(cfg)

	reqBody := ToolCallRequest{
		Name:      "nonexistent_tool",
		Arguments: map[string]interface{}{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/mcp/tools/call", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.HandleToolsCall(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 even for error, got %d", rr.Code)
	}

	var resp ToolCallResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.IsError {
		t.Error("expected error response")
	}
}

func TestCreateZipFromFiles(t *testing.T) {
	files := map[string]interface{}{
		"index.html":  "<html><body>Hello</body></html>",
		"styles.css":  "body { color: red; }",
		"script.js":   "console.log('hello');",
	}

	buf, count, size, err := createZipFromFiles(files)
	if err != nil {
		t.Fatalf("createZipFromFiles failed: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 files, got %d", count)
	}

	if size == 0 {
		t.Error("expected non-zero size")
	}

	if buf.Len() == 0 {
		t.Error("expected non-empty buffer")
	}
}

func TestHandleServersList(t *testing.T) {
	cfg := testConfig()
	s := NewServerWithConfig(cfg)

	result, err := s.handleServersList(nil)
	if err != nil {
		t.Fatalf("handleServersList failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}

	servers, ok := resultMap["servers"].([]map[string]interface{})
	if !ok {
		t.Fatal("servers is not a slice")
	}

	if len(servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(servers))
	}

	// Check that one server is marked as default
	hasDefault := false
	for _, srv := range servers {
		if srv["is_default"] == true {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		t.Error("expected one server to be marked as default")
	}
}
