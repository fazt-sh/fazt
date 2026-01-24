package provision

import (
	"testing"
)

func TestGetLocalIPs(t *testing.T) {
	ips := GetLocalIPs()
	if len(ips) == 0 {
		t.Log("No local IPs found (may be expected in some environments)")
	}
	for _, ip := range ips {
		if !IsIPAddress(ip) {
			t.Errorf("GetLocalIPs returned non-IP: %s", ip)
		}
	}
}

func TestGetPrimaryLocalIP(t *testing.T) {
	ip := GetPrimaryLocalIP()
	if ip == "" {
		t.Error("GetPrimaryLocalIP returned empty string")
	}
	if !IsIPAddress(ip) {
		t.Errorf("GetPrimaryLocalIP returned non-IP: %s", ip)
	}
}

func TestIsIPAddress(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"127.0.0.1", true},
		{"::1", true},
		{"2001:db8::1", true},
		{"example.com", false},
		{"not-an-ip", false},
		{"192.168.1", false},
		{"", false},
	}

	for _, test := range tests {
		result := IsIPAddress(test.input)
		if result != test.expected {
			t.Errorf("IsIPAddress(%q) = %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestCleanDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com", "example.com"},
		{"https://example.com", "example.com"},
		{"http://example.com", "example.com"},
		{"example.com:8080", "example.com"},
		{"https://example.com:443", "example.com"},
		{"example.com/path", "example.com"},
		{"192.168.1.1", "192.168.1.1"},
		{"192.168.1.1.nip.io", "192.168.1.1"},
		{"192.168.1.1.sslip.io", "192.168.1.1"},
		{"app.192.168.1.1.nip.io", "app.192.168.1.1"},
	}

	for _, test := range tests {
		result := cleanDomain(test.input)
		if result != test.expected {
			t.Errorf("cleanDomain(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestDetectEnvironment_IPMatch(t *testing.T) {
	// Get a local IP
	localIP := GetPrimaryLocalIP()
	if localIP == "127.0.0.1" {
		t.Skip("Only loopback available, skipping IP match test")
	}

	match, domain := DetectEnvironment(localIP)
	if match != EnvMatch {
		t.Errorf("DetectEnvironment(%q) = %v, want EnvMatch", localIP, match)
	}
	if domain != localIP {
		t.Errorf("DetectEnvironment returned domain %q, want %q", domain, localIP)
	}
}

func TestDetectEnvironment_Unconfigured(t *testing.T) {
	match, _ := DetectEnvironment("")
	if match != EnvUnconfigured {
		t.Errorf("DetectEnvironment(\"\") = %v, want EnvUnconfigured", match)
	}
}

func TestDetectEnvironment_Mismatch(t *testing.T) {
	// Use an IP that's definitely not local
	match, _ := DetectEnvironment("8.8.8.8")
	if match != EnvMismatch {
		t.Errorf("DetectEnvironment(\"8.8.8.8\") = %v, want EnvMismatch", match)
	}
}
