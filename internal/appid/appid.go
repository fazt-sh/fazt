// Package appid provides ID generation for fazt entities
package appid

import (
	"crypto/rand"
	"strings"
)

const (
	// FaztPrefix is the global prefix for all fazt IDs
	FaztPrefix = "fazt_"

	// ID type prefixes
	PrefixUser    = "usr"
	PrefixApp     = "app"
	PrefixToken   = "tok"
	PrefixSession = "ses"
	PrefixInvite  = "inv"

	// Base62 alphabet for new IDs (case-sensitive, URL-safe)
	base62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	// IDLength is the random part length for new fazt IDs
	IDLength = 12

	// LEGACY_CODE: Old format constants - remove when all IDs migrated
	legacyAppPrefix = "app_"
	legacyAlphabet  = "0123456789abcdefghijklmnopqrstuvwxyz"
)

// Generate creates a new fazt ID: fazt_<type>_<12 chars>
// Example: fazt_usr_Nf4rFeUfNV2H
func Generate(typePrefix string) string {
	return FaztPrefix + typePrefix + "_" + generateBase62(IDLength)
}

// GenerateUser creates a new user ID: fazt_usr_<12 chars>
func GenerateUser() string {
	return Generate(PrefixUser)
}

// GenerateApp creates a new app ID: fazt_app_<12 chars>
func GenerateApp() string {
	return Generate(PrefixApp)
}

// GenerateToken creates a new token ID: fazt_tok_<12 chars>
func GenerateToken() string {
	return Generate(PrefixToken)
}

// GenerateSession creates a new session ID: fazt_ses_<12 chars>
func GenerateSession() string {
	return Generate(PrefixSession)
}

// GenerateInvite creates a new invite ID: fazt_inv_<12 chars>
func GenerateInvite() string {
	return Generate(PrefixInvite)
}

// IsValid checks if a string is a valid fazt ID
func IsValid(id string) bool {
	if !strings.HasPrefix(id, FaztPrefix) {
		// LEGACY_CODE: Also accept old app_ format
		return isValidLegacy(id)
	}

	// Parse fazt_<type>_<random>
	rest := strings.TrimPrefix(id, FaztPrefix)
	parts := strings.SplitN(rest, "_", 2)
	if len(parts) != 2 {
		return false
	}

	typePrefix := parts[0]
	random := parts[1]

	// Validate type prefix
	validTypes := []string{PrefixUser, PrefixApp, PrefixToken, PrefixSession, PrefixInvite}
	typeValid := false
	for _, t := range validTypes {
		if typePrefix == t {
			typeValid = true
			break
		}
	}
	if !typeValid {
		return false
	}

	// Validate random part (12 chars, base62)
	if len(random) != IDLength {
		return false
	}
	for _, c := range random {
		if !strings.ContainsRune(base62, c) {
			return false
		}
	}
	return true
}

// GetType extracts the type prefix from a fazt ID
// Returns empty string for invalid or legacy IDs
func GetType(id string) string {
	if !strings.HasPrefix(id, FaztPrefix) {
		return ""
	}
	rest := strings.TrimPrefix(id, FaztPrefix)
	parts := strings.SplitN(rest, "_", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// generateBase62 generates a random base62 string
func generateBase62(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	for i := 0; i < length; i++ {
		bytes[i] = base62[bytes[i]%byte(len(base62))]
	}
	return string(bytes)
}

// =============================================================================
// LEGACY_CODE: Old app ID format - remove when all IDs migrated to fazt_ format
// =============================================================================

// GenerateLegacyApp creates an old-format app ID: app_<8 chars>
// LEGACY_CODE: Use GenerateApp() for new code
func GenerateLegacyApp() string {
	return legacyAppPrefix + generateLegacyRandom(8)
}

// isValidLegacy checks if a string is a valid legacy app ID
func isValidLegacy(id string) bool {
	if !strings.HasPrefix(id, legacyAppPrefix) {
		return false
	}
	random := strings.TrimPrefix(id, legacyAppPrefix)
	if len(random) < 4 || len(random) > 32 {
		return false
	}
	for _, c := range random {
		if !strings.ContainsRune(legacyAlphabet, c) {
			return false
		}
	}
	return true
}

// generateLegacyRandom generates a random string using legacy alphabet
func generateLegacyRandom(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	for i := 0; i < length; i++ {
		bytes[i] = legacyAlphabet[bytes[i]%byte(len(legacyAlphabet))]
	}
	return string(bytes)
}
