package egress

import (
	"testing"
)

func TestMatchDomain(t *testing.T) {
	tests := []struct {
		pattern string
		domain  string
		want    bool
	}{
		// Exact match
		{"api.stripe.com", "api.stripe.com", true},
		{"api.stripe.com", "api.example.com", false},

		// Wildcard match
		{"*.googleapis.com", "maps.googleapis.com", true},
		{"*.googleapis.com", "oauth2.googleapis.com", true},
		{"*.googleapis.com", "googleapis.com", false}, // bare domain doesn't match
		{"*.googleapis.com", "evil.example.com", false},

		// Case-insensitive (inputs should already be canonicalized)
		{"api.stripe.com", "api.stripe.com", true},

		// No bare wildcard
		{"*", "anything.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.domain, func(t *testing.T) {
			if got := matchDomain(tt.pattern, tt.domain); got != tt.want {
				t.Errorf("matchDomain(%q, %q) = %v, want %v", tt.pattern, tt.domain, got, tt.want)
			}
		})
	}
}

func TestAllowlistCRUD(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	al := NewAllowlist(db)

	// Initially empty
	entries, err := al.List("")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}

	// Add global
	if err := al.Add("api.stripe.com", "", true); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Add app-scoped
	if err := al.Add("api.openai.com", "myapp", true); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Add wildcard
	if err := al.Add("*.googleapis.com", "", true); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// List all
	entries, err = al.List("")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	// Check IsAllowed
	if !al.IsAllowed("api.stripe.com", "anyapp") {
		t.Error("api.stripe.com should be allowed globally")
	}
	if !al.IsAllowed("api.openai.com", "myapp") {
		t.Error("api.openai.com should be allowed for myapp")
	}
	if al.IsAllowed("api.openai.com", "otherapp") {
		t.Error("api.openai.com should NOT be allowed for otherapp")
	}
	if !al.IsAllowed("maps.googleapis.com", "anyapp") {
		t.Error("maps.googleapis.com should match wildcard *.googleapis.com")
	}
	if al.IsAllowed("googleapis.com", "anyapp") {
		t.Error("googleapis.com should NOT match *.googleapis.com")
	}
	if al.IsAllowed("evil.com", "anyapp") {
		t.Error("evil.com should not be allowed")
	}

	// Remove
	if err := al.Remove("api.stripe.com", ""); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if al.IsAllowed("api.stripe.com", "anyapp") {
		t.Error("api.stripe.com should no longer be allowed after removal")
	}
}

func TestAllowlistRejectsBareWildcard(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	al := NewAllowlist(db)
	err := al.Add("*", "", true)
	if err == nil {
		t.Error("expected error for bare wildcard")
	}
}

func TestAllowlistHTTPSOnly(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	al := NewAllowlist(db)
	al.Add("secure.example.com", "", true)
	al.Add("insecure.example.com", "", false)

	// entryFor should return correct HTTPSOnly setting
	entry := al.entryFor("secure.example.com", "")
	if entry == nil || !entry.HTTPSOnly {
		t.Error("secure.example.com should be HTTPS-only")
	}

	entry = al.entryFor("insecure.example.com", "")
	if entry == nil || entry.HTTPSOnly {
		t.Error("insecure.example.com should allow HTTP")
	}
}

func TestAllowlistCanonicalization(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	al := NewAllowlist(db)
	al.Add("API.Stripe.COM", "", true) // Will be canonicalized

	// Should match regardless of case/trailing dot/port
	if !al.IsAllowed("api.stripe.com", "") {
		t.Error("lowercase should match")
	}
	if !al.IsAllowed("API.Stripe.COM", "") {
		t.Error("uppercase should match (canonicalized)")
	}
	if !al.IsAllowed("api.stripe.com.", "") {
		t.Error("trailing dot should match (canonicalized)")
	}
	if !al.IsAllowed("api.stripe.com:443", "") {
		t.Error("with port should match (canonicalized)")
	}
}
