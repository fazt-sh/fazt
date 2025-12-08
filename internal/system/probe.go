package system

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Limits holds the detected system resource limits
type Limits struct {
	TotalRAM      int64 // Bytes
	AvailableRAM  int64 // Bytes (estimated)
	CPUCount      int
	MaxVFSBytes   int64 // Calculated safe limit
	MaxUploadBytes int64 // Calculated safe limit
}

var cachedLimits *Limits

// GetLimits probes the system and returns resource limits
func GetLimits() *Limits {
	if cachedLimits != nil {
		return cachedLimits
	}

	totalRAM := getMemoryLimit()
	cpuCount := runtime.NumCPU()

	// Heuristics for "Safe" limits
	// VFS Cache: 25% of Total RAM
	maxVFS := totalRAM / 4
	
	// Max Upload: 10% of RAM (to allow in-memory buffer before streaming)
	// But capped at 100MB to be reasonable
	maxUpload := totalRAM / 10
	if maxUpload > 100*1024*1024 {
		maxUpload = 100 * 1024 * 1024
	}
	// Min 10MB
	if maxUpload < 10*1024*1024 {
		maxUpload = 10 * 1024 * 1024
	}

	cachedLimits = &Limits{
		TotalRAM:       totalRAM,
		AvailableRAM:   totalRAM, // Approximate, we don't track live usage here
		CPUCount:       cpuCount,
		MaxVFSBytes:    maxVFS,
		MaxUploadBytes: maxUpload,
	}

	return cachedLimits
}

// getMemoryLimit tries to find the container/host memory limit
func getMemoryLimit() int64 {
	// 1. Try Cgroup V2
	if limit, err := readInt64("/sys/fs/cgroup/memory.max"); err == nil && limit > 0 {
		// "max" in cgroup v2 is usually "max", which we can't parse as int
		// logic moved to readInt64 helper to handle "max" string or huge numbers
		return limit
	}

	// 2. Try Cgroup V1
	if limit, err := readInt64("/sys/fs/cgroup/memory/memory.limit_in_bytes"); err == nil && limit > 0 {
		// Check for "unlimited" large numbers (often > 1PB in cgroups)
		if limit < 1<<50 { // If less than 1 Petabyte, believe it
			return limit
		}
	}

	// 3. Fallback to Host RAM (/proc/meminfo)
	if limit, err := getHostTotalRAM(); err == nil && limit > 0 {
		return limit
	}

	// 4. Fallback to runtime (Last resort, unreliable for Total)
	// Just return a safe default for a small VPS: 512MB
	return 512 * 1024 * 1024
}

// getHostTotalRAM reads MemTotal from /proc/meminfo
func getHostTotalRAM() (int64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				kb, err := strconv.ParseInt(parts[1], 10, 64)
				if err == nil {
					return kb * 1024, nil
				}
			}
		}
	}
	return 0, nil
}

func readInt64(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	if s == "max" {
		return 0, nil // Effectively unlimited/unknown
	}
	return strconv.ParseInt(s, 10, 64)
}
