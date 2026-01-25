package provision

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// EnvironmentMatch represents the result of environment detection
type EnvironmentMatch int

const (
	EnvUnconfigured EnvironmentMatch = iota // No domain in DB
	EnvMatch                                // Domain matches current machine
	EnvMismatch                             // Domain doesn't match (different machine)
)

// GetLocalIPs returns all IP addresses of the current machine
func GetLocalIPs() []string {
	var ips []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip IPv6 for now, focus on IPv4
			if ip == nil || ip.To4() == nil {
				continue
			}

			ips = append(ips, ip.String())
		}
	}

	return ips
}

// GetPrimaryLocalIP returns the most likely "main" local IP
// Prefers non-localhost, non-docker IPs
func GetPrimaryLocalIP() string {
	ips := GetLocalIPs()

	for _, ip := range ips {
		// Skip common virtual/docker networks
		if strings.HasPrefix(ip, "172.17.") || strings.HasPrefix(ip, "172.18.") {
			continue
		}
		// Prefer 192.168.x.x or 10.x.x.x
		if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
			return ip
		}
	}

	// Fall back to first available
	if len(ips) > 0 {
		return ips[0]
	}

	return "127.0.0.1"
}

// IsIPAddress checks if the string is an IP address
func IsIPAddress(s string) bool {
	return net.ParseIP(s) != nil
}

// DetectEnvironment checks if the stored domain matches the current machine
func DetectEnvironment(storedDomain string) (EnvironmentMatch, string) {
	// No domain configured
	if storedDomain == "" {
		return EnvUnconfigured, ""
	}

	// Clean the domain (remove protocol, port)
	domain := cleanDomain(storedDomain)

	// Get current machine's IPs
	myIPs := GetLocalIPs()

	// If stored domain is an IP, check directly
	if IsIPAddress(domain) {
		for _, ip := range myIPs {
			if ip == domain {
				return EnvMatch, domain
			}
		}
		return EnvMismatch, domain
	}

	// Domain is a hostname - try to resolve it
	resolvedIPs, err := net.LookupIP(domain)
	if err != nil {
		// Can't resolve - might be offline or invalid domain
		// Treat as mismatch to be safe
		return EnvMismatch, domain
	}

	// Check if any resolved IP matches our IPs
	for _, resolved := range resolvedIPs {
		resolvedStr := resolved.String()
		for _, myIP := range myIPs {
			if resolvedStr == myIP {
				return EnvMatch, domain
			}
		}
	}

	return EnvMismatch, domain
}

// cleanDomain removes protocol, port, and path from a domain string
func cleanDomain(domain string) string {
	d := domain

	// Remove protocol
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimPrefix(d, "https://")

	// Remove port
	if idx := strings.LastIndex(d, ":"); idx != -1 {
		// Make sure it's a port, not part of IPv6
		if !strings.Contains(d[idx:], "]") {
			d = d[:idx]
		}
	}

	// Remove path
	if idx := strings.Index(d, "/"); idx != -1 {
		d = d[:idx]
	}

	// Remove wildcard DNS suffix (e.g., .nip.io, .sslip.io)
	wildcardSuffixes := []string{".nip.io", ".sslip.io", ".xip.io"}
	for _, suffix := range wildcardSuffixes {
		if strings.HasSuffix(d, suffix) {
			d = strings.TrimSuffix(d, suffix)
			break
		}
	}

	return d
}

// GetHostname returns the machine's hostname
func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// DescribeEnvironment returns a human-readable description of the detection result
func DescribeEnvironment(match EnvironmentMatch, storedDomain, detectedIP string) string {
	switch match {
	case EnvUnconfigured:
		return fmt.Sprintf("No domain configured. Detected local IP: %s", detectedIP)
	case EnvMatch:
		return fmt.Sprintf("Domain '%s' matches this machine", storedDomain)
	case EnvMismatch:
		return fmt.Sprintf("Domain '%s' does not match this machine (local IP: %s)", storedDomain, detectedIP)
	default:
		return "Unknown environment state"
	}
}

// wildcardDNSSuffixes are domains that resolve any IP embedded in the hostname
var wildcardDNSSuffixes = []string{".nip.io", ".sslip.io", ".xip.io"}

// IsWildcardDNS checks if a domain uses a wildcard DNS service
func IsWildcardDNS(domain string) bool {
	d := strings.ToLower(domain)
	for _, suffix := range wildcardDNSSuffixes {
		if strings.HasSuffix(d, suffix) {
			return true
		}
	}
	return false
}

// IsPortableDomain returns true if the domain is machine-specific
// (IP address or wildcard DNS) and should be auto-updated when the DB is moved.
// Real domains (zyt.app, example.com) return false and are always trusted.
func IsPortableDomain(domain string) bool {
	cleaned := cleanDomain(domain)
	// IP addresses are portable
	if IsIPAddress(cleaned) {
		return true
	}
	// Wildcard DNS domains are portable
	if IsWildcardDNS(domain) {
		return true
	}
	return false
}

// CleanDomain is the exported version of cleanDomain
func CleanDomain(domain string) string {
	return cleanDomain(domain)
}
