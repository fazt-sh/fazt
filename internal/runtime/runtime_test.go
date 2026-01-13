package runtime

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewRuntime(t *testing.T) {
	r := NewRuntime(5, 100*time.Millisecond)
	if r == nil {
		t.Fatal("NewRuntime returned nil")
	}
	if r.poolSize != 5 {
		t.Errorf("expected pool size 5, got %d", r.poolSize)
	}
	if r.timeout != 100*time.Millisecond {
		t.Errorf("expected timeout 100ms, got %v", r.timeout)
	}
}

func TestNewRuntime_Defaults(t *testing.T) {
	r := NewRuntime(0, 0)
	if r.poolSize != MaxPoolSize {
		t.Errorf("expected default pool size %d, got %d", MaxPoolSize, r.poolSize)
	}
	if r.timeout != DefaultTimeout {
		t.Errorf("expected default timeout %v, got %v", DefaultTimeout, r.timeout)
	}
}

func TestExecute_SimpleReturn(t *testing.T) {
	r := NewRuntime(1, time.Second)
	ctx := context.Background()

	req := &Request{Method: "GET", Path: "/test"}
	result := r.Execute(ctx, `"hello world"`, req)

	if result.Error != nil {
		t.Fatalf("Execute failed: %v", result.Error)
	}
	if result.Response == nil {
		t.Fatal("Response is nil")
	}
	if result.Response.Body != "hello world" {
		t.Errorf("expected body 'hello world', got %v", result.Response.Body)
	}
	if result.Response.Status != 200 {
		t.Errorf("expected status 200, got %d", result.Response.Status)
	}
}

func TestExecute_RequestObject(t *testing.T) {
	r := NewRuntime(1, time.Second)
	ctx := context.Background()

	req := &Request{
		Method:  "POST",
		Path:    "/api/test",
		Query:   map[string]string{"id": "123"},
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    map[string]interface{}{"name": "Alice"},
	}

	result := r.Execute(ctx, `request.method + " " + request.path`, req)

	if result.Error != nil {
		t.Fatalf("Execute failed: %v", result.Error)
	}
	if result.Response.Body != "POST /api/test" {
		t.Errorf("expected 'POST /api/test', got %v", result.Response.Body)
	}
}

func TestExecute_RequestQuery(t *testing.T) {
	r := NewRuntime(1, time.Second)
	ctx := context.Background()

	req := &Request{
		Method: "GET",
		Path:   "/test",
		Query:  map[string]string{"id": "456", "name": "test"},
	}

	result := r.Execute(ctx, `request.query.id`, req)

	if result.Error != nil {
		t.Fatalf("Execute failed: %v", result.Error)
	}
	if result.Response.Body != "456" {
		t.Errorf("expected '456', got %v", result.Response.Body)
	}
}

func TestExecute_Respond(t *testing.T) {
	r := NewRuntime(1, 5*time.Second) // Use longer timeout for more reliable tests
	ctx := context.Background()

	req := &Request{Method: "GET", Path: "/test"}

	tests := []struct {
		name           string
		code           string
		expectedStatus int
		expectedBody   interface{}
	}{
		{"respond with body", `respond({message: "hello"})`, 200, map[string]interface{}{"message": "hello"}},
		{"respond with status and body", `respond(201, {id: 1})`, 201, map[string]interface{}{"id": int64(1)}},
		{"respond with status only", `respond(204)`, 204, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.Execute(ctx, tt.code, req)
			if result.Error != nil {
				t.Fatalf("Execute failed: %v", result.Error)
			}
			if result.Response.Status != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, result.Response.Status)
			}
		})
	}
}

func TestExecute_ConsoleLog(t *testing.T) {
	r := NewRuntime(1, time.Second)
	ctx := context.Background()

	req := &Request{Method: "GET", Path: "/test"}
	code := `
console.log("info message");
console.warn("warning message");
console.error("error message");
"done"
`
	result := r.Execute(ctx, code, req)

	if result.Error != nil {
		t.Fatalf("Execute failed: %v", result.Error)
	}

	if len(result.Logs) != 3 {
		t.Fatalf("expected 3 log entries, got %d", len(result.Logs))
	}

	expectedLogs := []struct {
		level   string
		message string
	}{
		{"info", "info message"},
		{"warn", "warning message"},
		{"error", "error message"},
	}

	for i, expected := range expectedLogs {
		if result.Logs[i].Level != expected.level {
			t.Errorf("log %d: expected level %s, got %s", i, expected.level, result.Logs[i].Level)
		}
		if result.Logs[i].Message != expected.message {
			t.Errorf("log %d: expected message %s, got %s", i, expected.message, result.Logs[i].Message)
		}
	}
}

func TestExecute_Timeout(t *testing.T) {
	r := NewRuntime(1, 50*time.Millisecond)
	ctx := context.Background()

	req := &Request{Method: "GET", Path: "/test"}
	code := `while(true) {}`

	result := r.Execute(ctx, code, req)

	if result.Error == nil {
		t.Fatal("expected timeout error")
	}
}

func TestExecute_SyntaxError(t *testing.T) {
	r := NewRuntime(1, time.Second)
	ctx := context.Background()

	req := &Request{Method: "GET", Path: "/test"}
	result := r.Execute(ctx, `invalid javascript {{{{`, req)

	if result.Error == nil {
		t.Fatal("expected syntax error")
	}
}

func TestExecute_RuntimeError(t *testing.T) {
	r := NewRuntime(1, time.Second)
	ctx := context.Background()

	req := &Request{Method: "GET", Path: "/test"}
	result := r.Execute(ctx, `undefinedVariable.property`, req)

	if result.Error == nil {
		t.Fatal("expected runtime error")
	}
}

func TestExecuteWithFiles_Require(t *testing.T) {
	r := NewRuntime(1, time.Second)
	ctx := context.Background()

	req := &Request{Method: "GET", Path: "/test"}

	files := map[string]string{
		"api/utils.js": `module.exports = { add: function(a, b) { return a + b; } };`,
	}

	loader := func(path string) (string, error) {
		if content, ok := files[path]; ok {
			return content, nil
		}
		return "", fmt.Errorf("file not found: %s", path)
	}

	code := `
var utils = require('./utils.js');
utils.add(1, 2);
`
	result := r.ExecuteWithFiles(ctx, code, req, loader)

	if result.Error != nil {
		t.Fatalf("Execute failed: %v", result.Error)
	}

	if result.Response.Body != int64(3) {
		t.Errorf("expected 3, got %v", result.Response.Body)
	}
}

func TestExecuteWithFiles_RequireCache(t *testing.T) {
	r := NewRuntime(1, time.Second)
	ctx := context.Background()

	req := &Request{Method: "GET", Path: "/test"}

	loadCount := 0
	files := map[string]string{
		"api/counter.js": `module.exports = { count: 1 };`,
	}

	loader := func(path string) (string, error) {
		loadCount++
		if content, ok := files[path]; ok {
			return content, nil
		}
		return "", fmt.Errorf("file not found: %s", path)
	}

	code := `
var a = require('./counter.js');
var b = require('./counter.js');
a === b;
`
	result := r.ExecuteWithFiles(ctx, code, req, loader)

	if result.Error != nil {
		t.Fatalf("Execute failed: %v", result.Error)
	}

	if loadCount != 1 {
		t.Errorf("expected file to be loaded once, loaded %d times", loadCount)
	}
}

func TestExecuteWithFiles_RequireNotFound(t *testing.T) {
	r := NewRuntime(1, time.Second)
	ctx := context.Background()

	req := &Request{Method: "GET", Path: "/test"}

	loader := func(path string) (string, error) {
		return "", fmt.Errorf("file not found: %s", path)
	}

	code := `require('./nonexistent.js');`
	result := r.ExecuteWithFiles(ctx, code, req, loader)

	if result.Error == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestResolvePath(t *testing.T) {
	tests := []struct {
		base     string
		require  string
		expected string
	}{
		{"api", "./utils.js", "api/utils.js"},
		{"api", "./lib/helper.js", "api/lib/helper.js"},
		{"api", "utils", "api/utils"},
	}

	for _, tt := range tests {
		t.Run(tt.require, func(t *testing.T) {
			result := resolvePath(tt.base, tt.require)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Suppress unused import warning
var _ = fmt.Sprintf
