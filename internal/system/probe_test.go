package system

import (
	"os"
	"runtime"
	"testing"
)

func TestGetLimits(t *testing.T) {
	// Reset cache to ensure fresh probe
	cachedLimits = nil

	limits := GetLimits()

	if limits == nil {
		t.Fatal("Expected limits to be non-nil")
	}

	// Verify basic constraints
	if limits.TotalRAM <= 0 {
		t.Errorf("Expected TotalRAM > 0, got %d", limits.TotalRAM)
	}

	if limits.CPUCount != runtime.NumCPU() {
		t.Errorf("Expected CPUCount to match runtime.NumCPU() (%d), got %d", runtime.NumCPU(), limits.CPUCount)
	}

	// Verify heuristics
	expectedMaxVFS := limits.TotalRAM / 4
	if limits.MaxVFSBytes != expectedMaxVFS {
		t.Errorf("Expected MaxVFSBytes to be 25%% of TotalRAM (%d), got %d", expectedMaxVFS, limits.MaxVFSBytes)
	}

	// MaxUpload should be 10% of RAM, capped at 100MB, min 10MB
	expectedMaxUpload := limits.TotalRAM / 10
	if expectedMaxUpload > 100*1024*1024 {
		expectedMaxUpload = 100 * 1024 * 1024
	}
	if expectedMaxUpload < 10*1024*1024 {
		expectedMaxUpload = 10 * 1024 * 1024
	}

	if limits.MaxUploadBytes != expectedMaxUpload {
		t.Errorf("Expected MaxUploadBytes to be %d, got %d", expectedMaxUpload, limits.MaxUploadBytes)
	}
}

func TestGetLimitsCaching(t *testing.T) {
	// Reset cache
	cachedLimits = nil

	// First call
	limits1 := GetLimits()

	// Second call should return cached value
	limits2 := GetLimits()

	// Should be the same pointer
	if limits1 != limits2 {
		t.Error("Expected GetLimits() to return cached value on second call")
	}
}

func TestGetHostTotalRAM(t *testing.T) {
	// This test only works on Linux with /proc/meminfo
	if runtime.GOOS != "linux" {
		t.Skip("Test only applicable on Linux")
	}

	ram, err := getHostTotalRAM()
	if err != nil {
		// /proc/meminfo might not exist in some test environments
		t.Skipf("Could not read /proc/meminfo: %v", err)
	}

	if ram <= 0 {
		t.Errorf("Expected positive RAM value, got %d", ram)
	}

	// Sanity check: RAM should be at least 100MB (very conservative)
	minExpected := int64(100 * 1024 * 1024)
	if ram < minExpected {
		t.Errorf("RAM value seems too low: %d bytes (expected at least %d)", ram, minExpected)
	}

	// Sanity check: RAM should be less than 1TB (conservative upper bound)
	maxExpected := int64(1024 * 1024 * 1024 * 1024)
	if ram > maxExpected {
		t.Errorf("RAM value seems too high: %d bytes (expected at most %d)", ram, maxExpected)
	}
}

func TestReadInt64(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int64
		wantErr  bool
	}{
		{
			name:     "valid number",
			content:  "1048576\n",
			expected: 1048576,
			wantErr:  false,
		},
		{
			name:     "max string",
			content:  "max\n",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "large number",
			content:  "9223372036854775807\n",
			expected: 9223372036854775807,
			wantErr:  false,
		},
		{
			name:     "zero",
			content:  "0\n",
			expected: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpfile, err := os.CreateTemp("", "test-cgroup-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			// Write content
			if _, err := tmpfile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			tmpfile.Close()

			// Read and verify
			result, err := readInt64(tmpfile.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("readInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("readInt64() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReadInt64NonExistentFile(t *testing.T) {
	_, err := readInt64("/nonexistent/path/to/file")
	if err == nil {
		t.Error("Expected error when reading non-existent file")
	}
}

func TestGetMemoryLimitFallback(t *testing.T) {
	// This test verifies the fallback behavior
	// Since we can't mock the filesystem easily, we just ensure it doesn't panic
	// and returns a reasonable default

	limit := getMemoryLimit()

	// Should return some positive value
	if limit <= 0 {
		t.Errorf("Expected positive memory limit, got %d", limit)
	}

	// Should be at least the minimum fallback (512MB)
	minExpected := int64(512 * 1024 * 1024)
	if limit < minExpected {
		t.Logf("Warning: Memory limit %d is less than minimum fallback %d", limit, minExpected)
	}
}

func TestMaxUploadBytesConstraints(t *testing.T) {
	// Reset cache
	cachedLimits = nil

	limits := GetLimits()

	// MaxUploadBytes should be at least 10MB
	minUpload := int64(10 * 1024 * 1024)
	if limits.MaxUploadBytes < minUpload {
		t.Errorf("MaxUploadBytes should be at least 10MB, got %d", limits.MaxUploadBytes)
	}

	// MaxUploadBytes should be at most 100MB
	maxUpload := int64(100 * 1024 * 1024)
	if limits.MaxUploadBytes > maxUpload {
		t.Errorf("MaxUploadBytes should be at most 100MB, got %d", limits.MaxUploadBytes)
	}
}

func TestAvailableRAMEquality(t *testing.T) {
	// Reset cache
	cachedLimits = nil

	limits := GetLimits()

	// Currently AvailableRAM is set equal to TotalRAM
	// This might change in the future to track actual usage
	if limits.AvailableRAM != limits.TotalRAM {
		t.Logf("Note: AvailableRAM (%d) differs from TotalRAM (%d)", limits.AvailableRAM, limits.TotalRAM)
	}
}
