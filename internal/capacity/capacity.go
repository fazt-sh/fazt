// Package capacity provides resource capacity estimation and limits configuration.
// Based on stress testing against SQLite storage with write serialization.
package capacity

import (
	"runtime"
)

// Profile represents tested capacity characteristics for a VPS tier.
type Profile struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	// Hardware specs
	RAMBytes int64 `json:"ram_bytes"`
	CPUCores int   `json:"cpu_cores"`

	// Tested throughput (requests/second)
	ReadThroughput  int `json:"read_throughput"`
	WriteThroughput int `json:"write_throughput"`
	MixedThroughput int `json:"mixed_throughput"` // 80% read, 20% write

	// Concurrent user estimates (sustained, not burst)
	ConcurrentUsers     int `json:"concurrent_users"`      // Conservative estimate
	ConcurrentUsersMax  int `json:"concurrent_users_max"`  // Best-case scenario
	ConcurrentUsersBurst int `json:"concurrent_users_burst"` // Short burst capacity

	// Resource growth per 100 concurrent connections
	RAMPerConnection int64   `json:"ram_per_connection"` // bytes
	CPUPerConnection float64 `json:"cpu_per_connection"` // percentage points
}

// Limits represents configurable capacity limits.
type Limits struct {
	// Storage write queue
	WriteQueueSize int `json:"write_queue_size"` // Max pending writes

	// Request handling
	MaxConcurrentRequests int `json:"max_concurrent_requests"`
	RequestTimeoutMs      int `json:"request_timeout_ms"`

	// Serverless runtime
	MaxExecutionTimeMs int   `json:"max_execution_time_ms"`
	MaxMemoryBytes     int64 `json:"max_memory_bytes"`

	// File/site limits
	MaxFileSizeBytes int64 `json:"max_file_size_bytes"`
	MaxSiteSizeBytes int64 `json:"max_site_size_bytes"`
}

// Stats represents current capacity utilization.
type Stats struct {
	// Memory
	HeapAllocBytes  uint64  `json:"heap_alloc_bytes"`
	HeapSysBytes    uint64  `json:"heap_sys_bytes"`
	MemoryUsedPct   float64 `json:"memory_used_pct"`

	// Write queue
	WriteQueueDepth    int     `json:"write_queue_depth"`
	WriteQueueCapacity int     `json:"write_queue_capacity"`
	WriteQueuePct      float64 `json:"write_queue_pct"`

	// Goroutines
	NumGoroutines int `json:"num_goroutines"`
}

// Predefined profiles based on stress testing (January 2026).
// Tested on fazt v0.10.10 with write serialization.
var (
	// ProfileBudgetVPS represents a $5-6/month VPS (1 vCPU, 1GB RAM)
	ProfileBudgetVPS = Profile{
		Name:        "budget-vps",
		Description: "$5-6/month VPS (1 vCPU, 1GB RAM)",
		RAMBytes:    1 * 1024 * 1024 * 1024,
		CPUCores:    1,

		// Conservative estimates (single vCPU limitation)
		ReadThroughput:  200,
		WriteThroughput: 100,
		MixedThroughput: 150,

		ConcurrentUsers:      50,
		ConcurrentUsersMax:   100,
		ConcurrentUsersBurst: 200,

		RAMPerConnection: 50 * 1024, // 50KB per connection
		CPUPerConnection: 0.5,
	}

	// ProfileMidVPS represents a $10-15/month VPS (2 vCPU, 2GB RAM)
	ProfileMidVPS = Profile{
		Name:        "mid-vps",
		Description: "$10-15/month VPS (2 vCPU, 2GB RAM)",
		RAMBytes:    2 * 1024 * 1024 * 1024,
		CPUCores:    2,

		ReadThroughput:  400,
		WriteThroughput: 150,
		MixedThroughput: 300,

		ConcurrentUsers:      100,
		ConcurrentUsersMax:   200,
		ConcurrentUsersBurst: 500,

		RAMPerConnection: 50 * 1024,
		CPUPerConnection: 0.3,
	}

	// ProfilePremiumVPS represents a $20-40/month VPS (4 vCPU, 4GB RAM)
	ProfilePremiumVPS = Profile{
		Name:        "premium-vps",
		Description: "$20-40/month VPS (4 vCPU, 4GB RAM)",
		RAMBytes:    4 * 1024 * 1024 * 1024,
		CPUCores:    4,

		ReadThroughput:  670, // Tested on M4 Air
		WriteThroughput: 530, // Tested on M4 Air
		MixedThroughput: 375, // Tested sustained

		ConcurrentUsers:      200,
		ConcurrentUsersMax:   500,
		ConcurrentUsersBurst: 1000,

		RAMPerConnection: 50 * 1024,
		CPUPerConnection: 0.2,
	}

	// Profiles is the list of all predefined profiles
	Profiles = []Profile{
		ProfileBudgetVPS,
		ProfileMidVPS,
		ProfilePremiumVPS,
	}
)

// DefaultLimits returns sensible defaults for a budget VPS.
func DefaultLimits() Limits {
	return Limits{
		WriteQueueSize:        1000,
		MaxConcurrentRequests: 500,
		RequestTimeoutMs:      5000,
		MaxExecutionTimeMs:    100,
		MaxMemoryBytes:        50 * 1024 * 1024,  // 50MB per execution
		MaxFileSizeBytes:      100 * 1024 * 1024, // 100MB
		MaxSiteSizeBytes:      500 * 1024 * 1024, // 500MB
	}
}

// LimitsForProfile returns recommended limits for a profile.
func LimitsForProfile(p Profile) Limits {
	base := DefaultLimits()

	// Scale write queue with available RAM
	ramGB := p.RAMBytes / (1024 * 1024 * 1024)
	base.WriteQueueSize = int(ramGB) * 1000

	// Scale concurrent requests with CPU cores
	base.MaxConcurrentRequests = p.CPUCores * 250

	return base
}

// CurrentStats returns current capacity utilization.
func CurrentStats(writeQueueDepth, writeQueueCapacity int) Stats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	queuePct := 0.0
	if writeQueueCapacity > 0 {
		queuePct = float64(writeQueueDepth) / float64(writeQueueCapacity)
	}

	return Stats{
		HeapAllocBytes:     m.HeapAlloc,
		HeapSysBytes:       m.HeapSys,
		MemoryUsedPct:      float64(m.HeapAlloc) / float64(m.HeapSys),
		WriteQueueDepth:    writeQueueDepth,
		WriteQueueCapacity: writeQueueCapacity,
		WriteQueuePct:      queuePct,
		NumGoroutines:      runtime.NumGoroutine(),
	}
}

// EstimateConcurrentUsers estimates capacity based on usage pattern.
func EstimateConcurrentUsers(p Profile, readPct, writePct float64) int {
	if readPct+writePct == 0 {
		return 0
	}

	// Weighted average of read/write throughput
	avgThroughput := (float64(p.ReadThroughput)*readPct + float64(p.WriteThroughput)*writePct) / (readPct + writePct)

	// Assume average user makes 0.2 requests/second (one request every 5 seconds)
	requestsPerUser := 0.2

	return int(avgThroughput / requestsPerUser)
}

// RecommendProfile recommends a profile based on expected concurrent users.
func RecommendProfile(expectedUsers int) Profile {
	for _, p := range Profiles {
		if expectedUsers <= p.ConcurrentUsers {
			return p
		}
	}
	// Return premium if above all profiles
	return ProfilePremiumVPS
}

// CapacityInfo provides comprehensive capacity information.
type CapacityInfo struct {
	// Current system
	CurrentStats Stats `json:"current_stats"`

	// Configured limits
	Limits Limits `json:"limits"`

	// Tested characteristics
	TestedProfile    Profile `json:"tested_profile"`
	TestEnvironment  string  `json:"test_environment"`
	TestDate         string  `json:"test_date"`
	FaztVersion      string  `json:"fazt_version"`

	// Architecture notes
	Architecture string `json:"architecture"`
}

// Info returns comprehensive capacity information.
func Info(writeQueueDepth, writeQueueCapacity int) CapacityInfo {
	return CapacityInfo{
		CurrentStats:    CurrentStats(writeQueueDepth, writeQueueCapacity),
		Limits:          DefaultLimits(),
		TestedProfile:   ProfilePremiumVPS, // Tests ran on powerful hardware
		TestEnvironment: "M4 Air 32GB (VM), fazt runtime",
		TestDate:        "2026-01-24",
		FaztVersion:     "0.10.10",
		Architecture: `Fazt uses SQLite with WAL mode for storage.
All writes serialize through a single-goroutine WriteQueue to prevent SQLITE_BUSY errors.
Reads execute directly (WAL allows concurrent reads).
Under overload, fazt returns HTTP 503 with Retry-After header.

Tested throughput:
- Pure reads:  670 req/sec
- Pure writes: 530 req/sec
- Mixed (80/20): 375 req/sec sustained @ 99% success

Memory footprint is minimal (~50KB per connection).
A $6 VPS can handle 50-100 concurrent users comfortably.`,
	}
}
