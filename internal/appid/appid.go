// Package appid provides app ID generation using nanoid-style identifiers
package appid

import (
	"crypto/rand"
	"fmt"
	"strings"
)

const (
	// Prefix for app IDs
	Prefix = "app_"

	// Alphabet for nanoid (URL-safe, no ambiguous chars)
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

	// DefaultLength is the default length of the random part
	DefaultLength = 8
)

// Generate creates a new app ID with the format "app_xxxxxxxx"
func Generate() string {
	return Prefix + generateRandom(DefaultLength)
}

// GenerateN creates a new app ID with custom length
func GenerateN(length int) string {
	return Prefix + generateRandom(length)
}

// IsValid checks if a string is a valid app ID
func IsValid(id string) bool {
	if !strings.HasPrefix(id, Prefix) {
		return false
	}
	random := strings.TrimPrefix(id, Prefix)
	if len(random) < 4 || len(random) > 32 {
		return false
	}
	for _, c := range random {
		if !strings.ContainsRune(alphabet, c) {
			return false
		}
	}
	return true
}

// generateRandom generates a random string of the given length
func generateRandom(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based if crypto/rand fails
		return fmt.Sprintf("%08x", uint32(length))
	}

	for i := 0; i < length; i++ {
		bytes[i] = alphabet[bytes[i]%byte(len(alphabet))]
	}
	return string(bytes)
}
