package appid

import (
	"strings"
	"testing"
)

func TestGenerateUser(t *testing.T) {
	id := GenerateUser()

	// Should have fazt_usr_ prefix
	if !strings.HasPrefix(id, "fazt_usr_") {
		t.Errorf("Expected prefix 'fazt_usr_', got '%s'", id)
	}

	// Should have correct length: fazt_usr_ (9) + 12 = 21
	if len(id) != 21 {
		t.Errorf("Expected length 21, got %d for %s", len(id), id)
	}

	// Should be valid
	if !IsValid(id) {
		t.Errorf("Generated ID should be valid: %s", id)
	}

	// Should have correct type
	if GetType(id) != PrefixUser {
		t.Errorf("Expected type 'usr', got '%s'", GetType(id))
	}
}

func TestGenerateApp(t *testing.T) {
	id := GenerateApp()

	if !strings.HasPrefix(id, "fazt_app_") {
		t.Errorf("Expected prefix 'fazt_app_', got '%s'", id)
	}

	if len(id) != 21 {
		t.Errorf("Expected length 21, got %d for %s", len(id), id)
	}

	if !IsValid(id) {
		t.Errorf("Generated ID should be valid: %s", id)
	}
}

func TestGenerateSession(t *testing.T) {
	id := GenerateSession()

	if !strings.HasPrefix(id, "fazt_ses_") {
		t.Errorf("Expected prefix 'fazt_ses_', got '%s'", id)
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
		// New fazt format
		{"fazt_usr_Nf4rFeUfNV2H", true},
		{"fazt_app_qW8n4P1zXy3m", true},
		{"fazt_tok_Ab3dEf6gHi9j", true},
		{"fazt_ses_Kl0mNoPqRs2t", true},
		{"fazt_inv_Mn7oPqRs3tUv", true},

		// Invalid new format
		{"fazt_usr_short", false},      // Too short
		{"fazt_usr_", false},           // No random part
		{"fazt_xxx_Nf4rFeUfNV2H", false}, // Invalid type
		{"fazt_Nf4rFeUfNV2H", false},   // Missing type
		{"fazt_usr_Nf4rFeUfNV2!", false}, // Invalid char

		// LEGACY_CODE: Old app_ format should still be valid
		{"app_abcd1234", true},
		{"app_a1b2c3d4", true},
		{"app_12345678", true},

		// Invalid legacy format
		{"app_abc", false},             // Too short
		{"app_", false},                // No random
		{"APP_abcd1234", false},        // Wrong case

		// Invalid general
		{"", false},
		{"random_string", false},
		{"user_123", false},
	}

	for _, tt := range tests {
		result := IsValid(tt.id)
		if result != tt.expected {
			t.Errorf("IsValid(%q) = %v, expected %v", tt.id, result, tt.expected)
		}
	}
}

func TestGetType(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"fazt_usr_Nf4rFeUfNV2H", "usr"},
		{"fazt_app_qW8n4P1zXy3m", "app"},
		{"fazt_tok_Ab3dEf6gHi9j", "tok"},
		{"app_abcd1234", ""},           // Legacy has no type
		{"invalid", ""},
	}

	for _, tt := range tests {
		result := GetType(tt.id)
		if result != tt.expected {
			t.Errorf("GetType(%q) = %q, expected %q", tt.id, result, tt.expected)
		}
	}
}

func TestGenerateUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		id := GenerateUser()
		if seen[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		seen[id] = true
	}
}

func TestBase62Alphabet(t *testing.T) {
	// Generate many IDs and verify they only contain base62 chars
	for i := 0; i < 100; i++ {
		id := GenerateUser()
		random := strings.TrimPrefix(id, "fazt_usr_")
		for _, c := range random {
			if !strings.ContainsRune(base62, c) {
				t.Errorf("ID contains invalid char '%c': %s", c, id)
			}
		}
	}
}

// LEGACY_CODE: Remove this test when old app_ format is removed
func TestLegacy_OldAppIDFormat(t *testing.T) {
	// Test legacy generation
	id := GenerateLegacyApp()
	if !strings.HasPrefix(id, "app_") {
		t.Errorf("Expected prefix 'app_', got '%s'", id)
	}
	if len(id) != 12 {
		t.Errorf("Expected length 12, got %d for %s", len(id), id)
	}
	if !IsValid(id) {
		t.Errorf("Legacy ID should be valid: %s", id)
	}

	// Test legacy validation
	legacyIDs := []string{
		"app_abcd1234",
		"app_a1b2c3d4",
		"app_12345678",
	}
	for _, lid := range legacyIDs {
		if !IsValid(lid) {
			t.Errorf("Legacy ID should be valid: %s", lid)
		}
	}
}
