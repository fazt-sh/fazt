package runtime

import (
	"testing"

	"github.com/dop251/goja"
)

func TestInjectFaztNamespace_App(t *testing.T) {
	vm := goja.New()
	result := &ExecuteResult{Logs: make([]LogEntry, 0)}

	app := &AppContext{
		ID:   "app_123",
		Name: "my-app",
	}

	if err := InjectFaztNamespace(vm, app, nil, result); err != nil {
		t.Fatalf("InjectFaztNamespace failed: %v", err)
	}

	// Test fazt.app.id
	val, err := vm.RunString("fazt.app.id")
	if err != nil {
		t.Fatalf("failed to access fazt.app.id: %v", err)
	}
	if val.Export() != "app_123" {
		t.Errorf("expected app_123, got %v", val.Export())
	}

	// Test fazt.app.name
	val, err = vm.RunString("fazt.app.name")
	if err != nil {
		t.Fatalf("failed to access fazt.app.name: %v", err)
	}
	if val.Export() != "my-app" {
		t.Errorf("expected my-app, got %v", val.Export())
	}
}

func TestInjectFaztNamespace_Env(t *testing.T) {
	vm := goja.New()
	result := &ExecuteResult{Logs: make([]LogEntry, 0)}

	env := EnvVars{
		"API_KEY":  "secret123",
		"NODE_ENV": "production",
	}

	if err := InjectFaztNamespace(vm, nil, env, result); err != nil {
		t.Fatalf("InjectFaztNamespace failed: %v", err)
	}

	// Test fazt.env.get()
	val, err := vm.RunString("fazt.env.get('API_KEY')")
	if err != nil {
		t.Fatalf("failed to call fazt.env.get: %v", err)
	}
	if val.Export() != "secret123" {
		t.Errorf("expected secret123, got %v", val.Export())
	}

	// Test fazt.env.get() with default
	val, err = vm.RunString("fazt.env.get('MISSING', 'default_value')")
	if err != nil {
		t.Fatalf("failed to call fazt.env.get with default: %v", err)
	}
	if val.Export() != "default_value" {
		t.Errorf("expected default_value, got %v", val.Export())
	}

	// Test fazt.env.has()
	val, err = vm.RunString("fazt.env.has('API_KEY')")
	if err != nil {
		t.Fatalf("failed to call fazt.env.has: %v", err)
	}
	if val.Export() != true {
		t.Errorf("expected true, got %v", val.Export())
	}

	// Test fazt.env.has() for missing key
	val, err = vm.RunString("fazt.env.has('MISSING')")
	if err != nil {
		t.Fatalf("failed to call fazt.env.has: %v", err)
	}
	if val.Export() != false {
		t.Errorf("expected false, got %v", val.Export())
	}
}

func TestInjectFaztNamespace_Log(t *testing.T) {
	vm := goja.New()
	result := &ExecuteResult{Logs: make([]LogEntry, 0)}

	if err := InjectFaztNamespace(vm, nil, nil, result); err != nil {
		t.Fatalf("InjectFaztNamespace failed: %v", err)
	}

	// Test fazt.log.info
	_, err := vm.RunString("fazt.log.info('info message')")
	if err != nil {
		t.Fatalf("failed to call fazt.log.info: %v", err)
	}

	// Test fazt.log.warn
	_, err = vm.RunString("fazt.log.warn('warning message')")
	if err != nil {
		t.Fatalf("failed to call fazt.log.warn: %v", err)
	}

	// Test fazt.log.error
	_, err = vm.RunString("fazt.log.error('error message')")
	if err != nil {
		t.Fatalf("failed to call fazt.log.error: %v", err)
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

func TestInjectFaztNamespace_Version(t *testing.T) {
	vm := goja.New()
	result := &ExecuteResult{Logs: make([]LogEntry, 0)}

	if err := InjectFaztNamespace(vm, nil, nil, result); err != nil {
		t.Fatalf("InjectFaztNamespace failed: %v", err)
	}

	val, err := vm.RunString("fazt.version")
	if err != nil {
		t.Fatalf("failed to access fazt.version: %v", err)
	}

	if val.Export() != "0.8.0" {
		t.Errorf("expected 0.8.0, got %v", val.Export())
	}
}

func TestInjectFaztNamespace_NilApp(t *testing.T) {
	vm := goja.New()
	result := &ExecuteResult{Logs: make([]LogEntry, 0)}

	if err := InjectFaztNamespace(vm, nil, nil, result); err != nil {
		t.Fatalf("InjectFaztNamespace failed: %v", err)
	}

	// Should have empty strings for nil app
	val, err := vm.RunString("fazt.app.id")
	if err != nil {
		t.Fatalf("failed to access fazt.app.id: %v", err)
	}
	if val.Export() != "" {
		t.Errorf("expected empty string, got %v", val.Export())
	}
}

func TestInjectFaztNamespace_EnvGetMissing(t *testing.T) {
	vm := goja.New()
	result := &ExecuteResult{Logs: make([]LogEntry, 0)}

	if err := InjectFaztNamespace(vm, nil, EnvVars{}, result); err != nil {
		t.Fatalf("InjectFaztNamespace failed: %v", err)
	}

	// Missing key with no default should return undefined
	val, err := vm.RunString("fazt.env.get('MISSING')")
	if err != nil {
		t.Fatalf("failed to call fazt.env.get: %v", err)
	}
	if !goja.IsUndefined(val) {
		t.Errorf("expected undefined, got %v", val.Export())
	}
}
