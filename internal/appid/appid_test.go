package appid

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	id := Generate()

	// Should have prefix
	if !strings.HasPrefix(id, Prefix) {
		t.Errorf("Expected prefix '%s', got '%s'", Prefix, id)
	}

	// Should have correct length (prefix + 8 chars)
	expectedLen := len(Prefix) + DefaultLength
	if len(id) != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, len(id))
	}

	// Should be valid
	if !IsValid(id) {
		t.Errorf("Generated ID should be valid: %s", id)
	}
}

func TestGenerateN(t *testing.T) {
	id := GenerateN(12)

	expectedLen := len(Prefix) + 12
	if len(id) != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, len(id))
	}

	if !IsValid(id) {
		t.Errorf("Generated ID should be valid: %s", id)
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		id       string
		expected bool
	}{
		{"app_abcd1234", true},
		{"app_a1b2c3d4", true},
		{"app_12345678", true},
		{"app_abcdefgh", true},
		{"app_abc", false},                 // Too short
		{"app_", false},                    // Too short
		{"app", false},                     // No prefix
		{"APP_abcd1234", false},            // Wrong case
		{"app_ABCD1234", false},            // Uppercase in random part
		{"app_abcd-1234", false},           // Invalid char
		{"app_abcd_1234", false},           // Invalid char
		{"other_abcd1234", false},          // Wrong prefix
		{"", false},                        // Empty
		{"abcd1234", false},                // No prefix
		{"app_" + strings.Repeat("a", 40), false}, // Too long
	}

	for _, tt := range tests {
		result := IsValid(tt.id)
		if result != tt.expected {
			t.Errorf("IsValid(%q) = %v, expected %v", tt.id, result, tt.expected)
		}
	}
}

func TestGenerateUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		id := Generate()
		if seen[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		seen[id] = true
	}
}

func TestGenerateRandomness(t *testing.T) {
	id1 := Generate()
	id2 := Generate()

	if id1 == id2 {
		t.Errorf("Two consecutive Generate() calls produced same ID: %s", id1)
	}
}
