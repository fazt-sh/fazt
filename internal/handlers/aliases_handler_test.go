package handlers

import (
	"testing"
)

func TestIsValidSubdomain(t *testing.T) {
	tests := []struct {
		subdomain string
		expected  bool
	}{
		{"tetris", true},
		{"my-app", true},
		{"app123", true},
		{"a", true},
		{"abc-def-123", true},

		{"", false},           // Empty
		{"-tetris", false},    // Starts with hyphen
		{"tetris-", false},    // Ends with hyphen
		{"Tetris", false},     // Uppercase (will be lowercased)
		{"my_app", false},     // Underscore not allowed
		{"my.app", false},     // Dot not allowed
		{"my app", false},     // Space not allowed
		{"app@123", false},    // Special char
	}

	for _, tt := range tests {
		result := isValidSubdomain(tt.subdomain)
		if result != tt.expected {
			t.Errorf("isValidSubdomain(%q) = %v, expected %v", tt.subdomain, result, tt.expected)
		}
	}
}

func TestIsValidSubdomainLength(t *testing.T) {
	// Test max length (63 chars)
	maxLen := "a12345678901234567890123456789012345678901234567890123456789012" // 63 chars
	if !isValidSubdomain(maxLen) {
		t.Errorf("Expected 63-char subdomain to be valid")
	}

	// Test over max length
	overMax := maxLen + "3"
	if isValidSubdomain(overMax) {
		t.Errorf("Expected 64-char subdomain to be invalid")
	}
}
