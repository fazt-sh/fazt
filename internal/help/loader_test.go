package help

import (
	"testing"
)

func TestResolveFilePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "cli/fazt.md"},
		{"fazt", "cli/fazt.md"},
		{"app", "cli/app/_index.md"},
		{"peer", "cli/peer/_index.md"},
		{"app deploy", "cli/app/deploy.md"},
		{"app list", "cli/app/list.md"},
		{"peer add", "cli/peer/add.md"},
	}

	for _, tt := range tests {
		result := resolveFilePath(tt.input)
		if result != tt.expected {
			t.Errorf("resolveFilePath(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestLoad(t *testing.T) {
	// Test loading root help
	doc, err := Load("")
	if err != nil {
		t.Fatalf("Failed to load root help: %v", err)
	}
	if doc.Description == "" {
		t.Error("Root help should have a description")
	}

	// Test loading app group help
	doc, err = Load("app")
	if err != nil {
		t.Fatalf("Failed to load app help: %v", err)
	}
	if doc.Command != "app" {
		t.Errorf("Expected command 'app', got %q", doc.Command)
	}

	// Test loading app deploy help
	doc, err = Load("app deploy")
	if err != nil {
		t.Fatalf("Failed to load app deploy help: %v", err)
	}
	if doc.Command != "app deploy" {
		t.Errorf("Expected command 'app deploy', got %q", doc.Command)
	}
	if len(doc.Arguments) == 0 {
		t.Error("app deploy should have arguments")
	}
	if len(doc.Flags) == 0 {
		t.Error("app deploy should have flags")
	}
}

func TestExists(t *testing.T) {
	if !Exists("") {
		t.Error("Root help should exist")
	}
	if !Exists("app") {
		t.Error("App help should exist")
	}
	if !Exists("app deploy") {
		t.Error("App deploy help should exist")
	}
	if Exists("nonexistent command") {
		t.Error("Nonexistent command should not exist")
	}
}

func TestRender(t *testing.T) {
	doc, err := Load("app deploy")
	if err != nil {
		t.Fatalf("Failed to load app deploy help: %v", err)
	}

	output := Render(doc)
	if output == "" {
		t.Error("Render should produce output")
	}

	// Check that key sections are present
	if !contains(output, "fazt app deploy") {
		t.Error("Output should contain command name")
	}
	if !contains(output, "directory") {
		t.Error("Output should contain directory argument")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(len(s) >= len(substr)) &&
		(s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
