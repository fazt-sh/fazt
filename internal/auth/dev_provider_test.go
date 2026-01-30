package auth

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestIsLocalMode(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		tls      bool
		expected bool
	}{
		{"no TLS", "example.com", false, true},
		{"localhost no TLS", "localhost:8080", false, true},
		{"127.0.0.1 no TLS", "127.0.0.1:8080", false, true},
		{"nip.io no TLS", "app.192.168.1.1.nip.io", false, true},
		{"HTTPS production", "example.com", true, false},
		{"HTTPS with localhost", "localhost:443", true, true},
		{"HTTPS with .local", "app.local:443", true, true},
		{"HTTPS with .internal", "app.internal:443", true, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "http://"+tc.host+"/", nil)
			r.Host = tc.host
			if tc.tls {
				r.TLS = &tls.ConnectionState{}
			}

			result := IsLocalMode(r)
			if result != tc.expected {
				t.Errorf("IsLocalMode() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestHashEmail(t *testing.T) {
	// Same email should produce same hash
	hash1 := hashEmail("test@example.com")
	hash2 := hashEmail("test@example.com")
	if hash1 != hash2 {
		t.Error("same email should produce same hash")
	}

	// Different emails should produce different hashes
	hash3 := hashEmail("other@example.com")
	if hash1 == hash3 {
		t.Error("different emails should produce different hashes")
	}

	// Hash should be 16 characters (8 bytes hex encoded)
	if len(hash1) != 16 {
		t.Errorf("hash length = %d, want 16", len(hash1))
	}
}

func TestDevLoginCallback(t *testing.T) {
	// Create test service
	db := setupTestDB(t)
	service := NewService(db, "localhost", false)
	handler := NewHandler(service)

	// Create form data
	form := url.Values{}
	form.Set("email", "test@example.com")
	form.Set("name", "Test User")
	form.Set("role", "user")
	form.Set("redirect", "/dashboard")

	// Create request (no TLS = local mode)
	req := httptest.NewRequest("POST", "http://localhost:8080/auth/dev/callback",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "localhost:8080"

	rr := httptest.NewRecorder()
	handler.DevLoginCallback(rr, req)

	// Should redirect
	if rr.Code != http.StatusFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusFound)
	}

	// Should set session cookie
	cookies := rr.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "fazt_session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Error("session cookie not set")
	}
	if sessionCookie.Value == "" {
		t.Error("session cookie value empty")
	}

	// Should redirect to dashboard
	location := rr.Header().Get("Location")
	if location != "/dashboard" {
		t.Errorf("redirect = %q, want /dashboard", location)
	}

	// Verify user was created
	user, err := service.GetUserByEmail("test@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail error: %v", err)
	}
	if user.Name != "Test User" {
		t.Errorf("user.Name = %q, want 'Test User'", user.Name)
	}
	if user.Provider != "dev" {
		t.Errorf("user.Provider = %q, want 'dev'", user.Provider)
	}

	// Verify session is valid
	validUser, err := service.ValidateSession(sessionCookie.Value)
	if err != nil {
		t.Fatalf("ValidateSession error: %v", err)
	}
	if validUser.Email != "test@example.com" {
		t.Errorf("session user email = %q, want 'test@example.com'", validUser.Email)
	}
}

func TestDevLoginBlockedOnHTTPS(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "example.com", true)
	handler := NewHandler(service)

	// Create request with TLS (production mode)
	req := httptest.NewRequest("GET", "https://example.com/auth/dev/login", nil)
	req.TLS = &tls.ConnectionState{}
	req.Host = "example.com"

	rr := httptest.NewRecorder()
	handler.DevLoginForm(rr, req)

	// Should be forbidden
	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d (forbidden)", rr.Code, http.StatusForbidden)
	}
}

func TestDevLoginWithRoleUpdate(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db, "localhost", false)
	handler := NewHandler(service)

	// First login as user
	form := url.Values{}
	form.Set("email", "admin@example.com")
	form.Set("name", "Admin User")
	form.Set("role", "user")
	form.Set("redirect", "/")

	req := httptest.NewRequest("POST", "http://localhost:8080/auth/dev/callback",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "localhost:8080"

	rr := httptest.NewRecorder()
	handler.DevLoginCallback(rr, req)

	// Verify user is regular user
	user, _ := service.GetUserByEmail("admin@example.com")
	if user.Role != "user" {
		t.Errorf("initial role = %q, want 'user'", user.Role)
	}

	// Login again as admin
	form.Set("role", "admin")
	req = httptest.NewRequest("POST", "http://localhost:8080/auth/dev/callback",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "localhost:8080"

	rr = httptest.NewRecorder()
	handler.DevLoginCallback(rr, req)

	// Verify role was updated
	user, _ = service.GetUserByEmail("admin@example.com")
	if user.Role != "admin" {
		t.Errorf("updated role = %q, want 'admin'", user.Role)
	}
}
