package system

import (
	"os"
	"runtime"
	"testing"
)

func TestGetLimits(t *testing.T) {
	cachedLimits = nil

	limits := GetLimits()

	if limits == nil {
		t.Fatal("Expected limits to be non-nil")
	}

	// Verify basic constraints
	if limits.Hardware.TotalRAM <= 0 {
		t.Errorf("Expected Hardware.TotalRAM > 0, got %d", limits.Hardware.TotalRAM)
	}

	if limits.Hardware.CPUCores != runtime.NumCPU() {
		t.Errorf("Expected Hardware.CPUCores to match runtime.NumCPU() (%d), got %d", runtime.NumCPU(), limits.Hardware.CPUCores)
	}

	// Verify heuristics
	expectedMaxVFS := limits.Hardware.TotalRAM / 4
	if limits.Storage.MaxVFS != expectedMaxVFS {
		t.Errorf("Expected Storage.MaxVFS to be 25%% of TotalRAM (%d), got %d", expectedMaxVFS, limits.Storage.MaxVFS)
	}

	// MaxUpload should be 10% of RAM, capped at 100MB, min 10MB
	expectedMaxUpload := limits.Hardware.TotalRAM / 10
	if expectedMaxUpload > 100*1024*1024 {
		expectedMaxUpload = 100 * 1024 * 1024
	}
	if expectedMaxUpload < 10*1024*1024 {
		expectedMaxUpload = 10 * 1024 * 1024
	}

	if limits.Storage.MaxUpload != expectedMaxUpload {
		t.Errorf("Expected Storage.MaxUpload to be %d, got %d", expectedMaxUpload, limits.Storage.MaxUpload)
	}
}

func TestGetLimitsNetDefaults(t *testing.T) {
	cachedLimits = nil

	limits := GetLimits()

	if limits.Net.MaxCalls != 5 {
		t.Errorf("Net.MaxCalls: got %d, want 5", limits.Net.MaxCalls)
	}
	if limits.Net.CallTimeout != 4000 {
		t.Errorf("Net.CallTimeout: got %d, want 4000", limits.Net.CallTimeout)
	}
	if limits.Net.MaxRedirects != 3 {
		t.Errorf("Net.MaxRedirects: got %d, want 3", limits.Net.MaxRedirects)
	}
	if limits.Net.Concurrency < 20 {
		t.Errorf("Net.Concurrency should be >= 20, got %d", limits.Net.Concurrency)
	}
	if limits.Net.RateLimit != 0 {
		t.Errorf("Net.RateLimit should default to 0 (disabled), got %d", limits.Net.RateLimit)
	}
	if limits.Net.CacheMaxItems != 0 {
		t.Errorf("Net.CacheMaxItems should default to 0 (disabled), got %d", limits.Net.CacheMaxItems)
	}
}

func TestGetLimitsCaching(t *testing.T) {
	cachedLimits = nil

	limits1 := GetLimits()
	limits2 := GetLimits()

	if limits1 != limits2 {
		t.Error("Expected GetLimits() to return cached value on second call")
	}
}

func TestGetHostTotalRAM(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Test only applicable on Linux")
	}

	ram, err := getHostTotalRAM()
	if err != nil {
		t.Skipf("Could not read /proc/meminfo: %v", err)
	}

	if ram <= 0 {
		t.Errorf("Expected positive RAM value, got %d", ram)
	}

	minExpected := int64(100 * 1024 * 1024)
	if ram < minExpected {
		t.Errorf("RAM value seems too low: %d bytes (expected at least %d)", ram, minExpected)
	}

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
		{"valid number", "1048576\n", 1048576, false},
		{"max string", "max\n", 0, false},
		{"large number", "9223372036854775807\n", 9223372036854775807, false},
		{"zero", "0\n", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test-cgroup-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			tmpfile.Close()

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
	limit := getMemoryLimit()

	if limit <= 0 {
		t.Errorf("Expected positive memory limit, got %d", limit)
	}

	minExpected := int64(512 * 1024 * 1024)
	if limit < minExpected {
		t.Logf("Warning: Memory limit %d is less than minimum fallback %d", limit, minExpected)
	}
}

func TestMaxUploadConstraints(t *testing.T) {
	cachedLimits = nil

	limits := GetLimits()

	minUpload := int64(10 * 1024 * 1024)
	if limits.Storage.MaxUpload < minUpload {
		t.Errorf("Storage.MaxUpload should be at least 10MB, got %d", limits.Storage.MaxUpload)
	}

	maxUpload := int64(100 * 1024 * 1024)
	if limits.Storage.MaxUpload > maxUpload {
		t.Errorf("Storage.MaxUpload should be at most 100MB, got %d", limits.Storage.MaxUpload)
	}
}

func TestAvailableRAMEquality(t *testing.T) {
	cachedLimits = nil

	limits := GetLimits()

	if limits.Hardware.AvailableRAM != limits.Hardware.TotalRAM {
		t.Logf("Note: AvailableRAM (%d) differs from TotalRAM (%d)", limits.Hardware.AvailableRAM, limits.Hardware.TotalRAM)
	}
}
