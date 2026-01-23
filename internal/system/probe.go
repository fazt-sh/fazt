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
	TotalRAM       int64 // Bytes
	AvailableRAM   int64 // Bytes (estimated)
	CPUCount       int
	MaxVFSBytes    int64 // Calculated safe limit
	MaxUploadBytes int64 // Calculated safe limit

	// Capacity estimates (based on stress testing)
	ConcurrentUsers      int `json:"concurrent_users"`       // Conservative estimate
	ConcurrentUsersMax   int `json:"concurrent_users_max"`   // Best-case scenario
	ReadThroughput       int `json:"read_throughput"`        // req/sec
	WriteThroughput      int `json:"write_throughput"`       // req/sec
	MixedThroughput      int `json:"mixed_throughput"`       // req/sec (80R/20W)
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

	// Capacity estimates based on stress testing (fazt v0.10.10, Jan 2026)
	// Baseline: $6 VPS = 1 vCPU, 1GB RAM
	// Scales roughly linearly with CPU cores for read-heavy workloads
	baseUsers := 50
	baseReads := 200
	baseWrites := 100
	baseMixed := 150

	// Scale with CPU cores (diminishing returns after 4 cores for SQLite)
	scaleFactor := cpuCount
	if scaleFactor > 4 {
		scaleFactor = 4 + (cpuCount-4)/2 // Half credit above 4 cores
	}

	cachedLimits = &Limits{
		TotalRAM:       totalRAM,
		AvailableRAM:   totalRAM, // Approximate, we don't track live usage here
		CPUCount:       cpuCount,
		MaxVFSBytes:    maxVFS,
		MaxUploadBytes: maxUpload,

		ConcurrentUsers:    baseUsers * scaleFactor,
		ConcurrentUsersMax: baseUsers * scaleFactor * 2,
		ReadThroughput:     baseReads * scaleFactor,
		WriteThroughput:    baseWrites + (scaleFactor-1)*25, // Writes don't scale as well
		MixedThroughput:    baseMixed * scaleFactor,
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
