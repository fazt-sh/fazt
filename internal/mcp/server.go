// Package mcp implements the Model Context Protocol server for Fazt.
// This allows Claude Code and other MCP clients to interact with Fazt servers.
package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/fazt-sh/fazt/internal/clientconfig"
)

const (
	ProtocolVersion = "2024-11-05"
	ServerName      = "fazt"
)

// Server is the MCP server instance.
type Server struct {
	config *clientconfig.Config
	tools  map[string]Tool
	mu     sync.RWMutex
}

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Handler     ToolHandler            `json:"-"`
}

// ToolHandler is the function signature for tool implementations.
type ToolHandler func(params map[string]interface{}) (interface{}, error)

// InitializeRequest is the MCP initialize request.
type InitializeRequest struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

// ClientInfo contains information about the MCP client.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResponse is the MCP initialize response.
type InitializeResponse struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ServerCapabilities     `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
}

// ServerCapabilities describes what the server supports.
type ServerCapabilities struct {
	Tools *ToolCapabilities `json:"tools,omitempty"`
}

// ToolCapabilities describes tool-related capabilities.
type ToolCapabilities struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ServerInfo contains server identification.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ToolsListResponse is the response to tools/list.
type ToolsListResponse struct {
	Tools []ToolDefinition `json:"tools"`
}

// ToolDefinition is the tool definition in list response.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolCallRequest is the request to tools/call.
type ToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCallResponse is the response from tools/call.
type ToolCallResponse struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock represents a content item in the response.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// NewServer creates a new MCP server instance.
func NewServer() (*Server, error) {
	cfg, err := clientconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	s := &Server{
		config: cfg,
		tools:  make(map[string]Tool),
	}

	// Register default tools
	s.registerDefaultTools()

	return s, nil
}

// NewServerWithConfig creates a new MCP server with a provided config.
func NewServerWithConfig(cfg *clientconfig.Config) *Server {
	s := &Server{
		config: cfg,
		tools:  make(map[string]Tool),
	}
	s.registerDefaultTools()
	return s
}

// RegisterTool adds a tool to the server.
func (s *Server) RegisterTool(tool Tool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
}

// GetTools returns all registered tools.
func (s *Server) GetTools() []Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]Tool, 0, len(s.tools))
	for _, t := range s.tools {
		tools = append(tools, t)
	}
	return tools
}

// CallTool executes a tool by name.
func (s *Server) CallTool(name string, params map[string]interface{}) (interface{}, error) {
	s.mu.RLock()
	tool, ok := s.tools[name]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}

	if tool.Handler == nil {
		return nil, fmt.Errorf("tool '%s' has no handler", name)
	}

	return tool.Handler(params)
}

// HandleInitialize handles the /mcp/initialize endpoint.
func (s *Server) HandleInitialize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InitializeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp := InitializeResponse{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolCapabilities{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    ServerName,
			Version: "0.8.4",
		},
	}

	writeJSON(w, resp)
}

// HandleToolsList handles the /mcp/tools/list endpoint.
func (s *Server) HandleToolsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := s.GetTools()
	defs := make([]ToolDefinition, len(tools))
	for i, t := range tools {
		defs[i] = ToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		}
	}

	writeJSON(w, ToolsListResponse{Tools: defs})
}

// HandleToolsCall handles the /mcp/tools/call endpoint.
func (s *Server) HandleToolsCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ToolCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	result, err := s.CallTool(req.Name, req.Arguments)
	if err != nil {
		writeJSON(w, ToolCallResponse{
			Content: []ContentBlock{{Type: "text", Text: err.Error()}},
			IsError: true,
		})
		return
	}

	// Convert result to JSON string for text content
	resultJSON, _ := json.MarshalIndent(result, "", "  ")

	writeJSON(w, ToolCallResponse{
		Content: []ContentBlock{{Type: "text", Text: string(resultJSON)}},
	})
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	})
}
