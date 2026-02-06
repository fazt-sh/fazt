package egress

import (
	"net/http"
	"net/url"
	"testing"
)

func TestSecretsSetAndLookup(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	store := NewSecretsStore(db)

	// Set a bearer secret
	err := store.Set("STRIPE_KEY", "sk_live_xxx", "bearer", "", "", "")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Lookup
	secret, err := store.Lookup("STRIPE_KEY", "anyapp")
	if err != nil {
		t.Fatalf("Lookup failed: %v", err)
	}
	if secret.Name != "STRIPE_KEY" {
		t.Errorf("Name: got %q, want %q", secret.Name, "STRIPE_KEY")
	}
	if secret.Value != "sk_live_xxx" {
		t.Errorf("Value: got %q, want %q", secret.Value, "sk_live_xxx")
	}
	if secret.InjectAs != "bearer" {
		t.Errorf("InjectAs: got %q, want %q", secret.InjectAs, "bearer")
	}
}

func TestSecretsAppScoped(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	store := NewSecretsStore(db)

	// Global secret
	store.Set("GLOBAL_KEY", "global_val", "bearer", "", "", "")
	// App-scoped secret
	store.Set("APP_KEY", "app_val", "bearer", "", "", "myapp")

	// App should see both
	sec, err := store.Lookup("GLOBAL_KEY", "myapp")
	if err != nil || sec.Value != "global_val" {
		t.Error("app should see global secret")
	}

	sec, err = store.Lookup("APP_KEY", "myapp")
	if err != nil || sec.Value != "app_val" {
		t.Error("app should see its own secret")
	}

	// Other app should not see app-scoped secret
	_, err = store.Lookup("APP_KEY", "otherapp")
	if err == nil {
		t.Error("other app should not see myapp's secret")
	}
}

func TestSecretsInjectBearer(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	store := NewSecretsStore(db)
	store.Set("TOKEN", "mytoken", "bearer", "", "", "")

	secret, _ := store.Lookup("TOKEN", "")
	req, _ := http.NewRequest("GET", "https://api.example.com/data", nil)

	err := store.InjectAuth(req, secret, "api.example.com")
	if err != nil {
		t.Fatalf("InjectAuth failed: %v", err)
	}

	auth := req.Header.Get("Authorization")
	if auth != "Bearer mytoken" {
		t.Errorf("Authorization: got %q, want %q", auth, "Bearer mytoken")
	}
}

func TestSecretsInjectHeader(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	store := NewSecretsStore(db)
	store.Set("API_KEY", "sk-xxx", "header", "X-API-Key", "", "")

	secret, _ := store.Lookup("API_KEY", "")
	req, _ := http.NewRequest("GET", "https://api.example.com/data", nil)

	err := store.InjectAuth(req, secret, "api.example.com")
	if err != nil {
		t.Fatalf("InjectAuth failed: %v", err)
	}

	val := req.Header.Get("X-API-Key")
	if val != "sk-xxx" {
		t.Errorf("X-API-Key: got %q, want %q", val, "sk-xxx")
	}
}

func TestSecretsInjectQuery(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	store := NewSecretsStore(db)
	store.Set("TOKEN", "abc123", "query", "token", "", "")

	secret, _ := store.Lookup("TOKEN", "")
	u, _ := url.Parse("https://api.example.com/data")
	req := &http.Request{URL: u, Header: http.Header{}}

	err := store.InjectAuth(req, secret, "api.example.com")
	if err != nil {
		t.Fatalf("InjectAuth failed: %v", err)
	}

	if req.URL.Query().Get("token") != "abc123" {
		t.Errorf("query param 'token': got %q, want %q", req.URL.Query().Get("token"), "abc123")
	}
}

func TestSecretsDomainRestriction(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	store := NewSecretsStore(db)
	store.Set("KEY", "val", "bearer", "", "api.stripe.com", "")

	secret, _ := store.Lookup("KEY", "")

	// Should work for matching domain
	req, _ := http.NewRequest("GET", "https://api.stripe.com/v1", nil)
	err := store.InjectAuth(req, secret, "api.stripe.com")
	if err != nil {
		t.Errorf("expected success for matching domain, got: %v", err)
	}

	// Should fail for different domain
	req, _ = http.NewRequest("GET", "https://api.evil.com/steal", nil)
	err = store.InjectAuth(req, secret, "api.evil.com")
	if err == nil {
		t.Error("expected error for non-matching domain")
	}
}

func TestSecretsInvalidInjectAs(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	store := NewSecretsStore(db)
	err := store.Set("KEY", "val", "invalid", "", "", "")
	if err == nil {
		t.Error("expected error for invalid inject_as")
	}
}

func TestSecretsMissingInjectKey(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	store := NewSecretsStore(db)
	err := store.Set("KEY", "val", "header", "", "", "")
	if err == nil {
		t.Error("expected error when inject_key missing for header type")
	}
}

func TestSecretsRemove(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	store := NewSecretsStore(db)
	store.Set("KEY", "val", "bearer", "", "", "")

	err := store.Remove("KEY", "")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	_, err = store.Lookup("KEY", "")
	if err == nil {
		t.Error("expected error after removal")
	}
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abc", "***"},
		{"abcdef", "******"},
		{"sk_live_abcdef123", "sk_***********123"},
	}
	for _, tt := range tests {
		got := MaskValue(tt.input)
		if got != tt.want {
			t.Errorf("MaskValue(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSecretsList(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	store := NewSecretsStore(db)
	store.Set("KEY1", "val1", "bearer", "", "", "")
	store.Set("KEY2", "val2", "header", "X-Key", "api.com", "myapp")

	secrets, err := store.List("")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(secrets))
	}
}
