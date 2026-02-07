package system

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Limits describes all system resource limits.
// Struct tags provide metadata for API schema, admin UI, and future validation.
//
// Tags:
//
//	json     — JSON field name
//	label    — Human-readable label (for UI)
//	desc     — Short description (for UI tooltips / API docs)
//	unit     — Value unit: "bytes", "ms", "count" (for UI formatting)
//	range    — "min,max" accepted values (for UI sliders, future validation)
//	readonly — "true" if hardware-detected, not configurable
type Limits struct {
	Hardware Hardware `json:"hardware"`
	Storage  Storage  `json:"storage"`
	Runtime  Runtime  `json:"runtime"`
	Capacity Capacity `json:"capacity"`
	Net      Net      `json:"net"`
	Media    Media    `json:"media"`
}

// Hardware holds detected hardware characteristics.
type Hardware struct {
	TotalRAM     int64 `json:"total_ram"    label:"Total RAM"     desc:"Detected system memory"     unit:"bytes" readonly:"true"`
	AvailableRAM int64 `json:"available_ram" label:"Available RAM" desc:"Estimated available memory"  unit:"bytes" readonly:"true"`
	CPUCores     int   `json:"cpu_cores"    label:"CPU Cores"     desc:"Detected CPU cores"          readonly:"true"`
}

// Storage holds storage-related limits.
type Storage struct {
	MaxVFS      int64 `json:"max_vfs"      label:"VFS Cache"    desc:"Max VFS cache size"     unit:"bytes" range:"10485760,1073741824"`
	MaxUpload   int64 `json:"max_upload"    label:"Max Upload"   desc:"Max upload size"        unit:"bytes" range:"1048576,104857600"`
	WriteQueue  int   `json:"write_queue"   label:"Write Queue"  desc:"Max pending writes"     range:"100,10000"`
	MaxFileSize int64 `json:"max_file_size" label:"Max File"     desc:"Max single file size"   unit:"bytes" range:"1048576,1073741824"`
	MaxSiteSize int64 `json:"max_site_size" label:"Max Site"     desc:"Max total site size"    unit:"bytes" range:"10485760,5368709120"`
	MaxLogRows  int   `json:"max_log_rows"  label:"Max Log Rows" desc:"Activity log row limit"  range:"10000,5000000"`
}

// Runtime holds serverless execution limits.
type Runtime struct {
	ExecTimeout int   `json:"exec_timeout" label:"Exec Timeout" desc:"Serverless execution timeout" unit:"ms" range:"100,10000"`
	MaxMemory   int64 `json:"max_memory"   label:"Max Memory"   desc:"Per-execution memory limit"   unit:"bytes" range:"1048576,268435456"`
}

// Capacity holds capacity estimates based on stress testing.
type Capacity struct {
	Users       int `json:"users"        label:"Concurrent Users" desc:"Conservative estimate"   readonly:"true"`
	UsersMax    int `json:"users_max"    label:"Max Users"        desc:"Best-case estimate"      readonly:"true"`
	Reads       int `json:"reads"        label:"Read Throughput"  desc:"Requests/sec"            unit:"req/s" readonly:"true"`
	Writes      int `json:"writes"       label:"Write Throughput" desc:"Requests/sec"            unit:"req/s" readonly:"true"`
	Mixed       int `json:"mixed"        label:"Mixed Throughput" desc:"80%% read / 20%% write"  unit:"req/s" readonly:"true"`
	MaxRequests int `json:"max_requests" label:"Max Concurrent"   desc:"Max concurrent requests"  range:"50,5000"`
	Timeout     int `json:"timeout_ms"   label:"Request Timeout"  desc:"Request timeout"          unit:"ms" range:"1000,30000"`
}

// Net holds network egress limits for serverless fetch calls.
type Net struct {
	// Phase 1 — Egress core
	MaxCalls       int   `json:"max_calls"       label:"Max Calls"       desc:"Fetch calls per request"     range:"1,20"`
	CallTimeout    int   `json:"call_timeout"    label:"Call Timeout"    desc:"Per-call timeout"            unit:"ms" range:"1000,10000"`
	Budget         int   `json:"budget"          label:"HTTP Budget"     desc:"Total HTTP time per request"  unit:"ms" range:"1000,10000"`
	AppConcurrency int   `json:"app_concurrency" label:"App Concurrency" desc:"Per-app concurrent outbound"  range:"1,20"`
	Concurrency    int   `json:"concurrency"     label:"Concurrency"     desc:"Global concurrent outbound"   range:"5,100"`
	MaxRequestBody int64 `json:"max_req_body"    label:"Max Request"     desc:"Outgoing body size limit"    unit:"bytes" range:"1024,10485760"`
	MaxResponse    int64 `json:"max_response"    label:"Max Response"    desc:"Response body size limit"    unit:"bytes" range:"1024,10485760"`
	MaxRedirects   int   `json:"max_redirects"   label:"Max Redirects"   desc:"Redirect hop limit"          range:"0,10"`

	// Phase 2 — Rate limiting
	RateLimit int `json:"rate_limit" label:"Rate Limit" desc:"Default requests/min per domain" range:"0,1000"`
	RateBurst int `json:"rate_burst" label:"Rate Burst"  desc:"Burst allowance above rate"      range:"0,100"`

	// Phase 3 — Observability
	LogBufferSize int   `json:"log_buffer"      label:"Log Buffer"  desc:"In-memory log entries before flush" range:"100,10000"`
	LogFlushMs    int   `json:"log_flush"       label:"Log Flush"   desc:"Flush interval"                     unit:"ms" range:"500,10000"`
	CacheMaxItems int   `json:"cache_max_items" label:"Cache Items" desc:"Max cached responses"               range:"0,10000"`
	CacheMaxBytes int64 `json:"cache_max_bytes" label:"Cache Size"  desc:"Max cache memory"                   unit:"bytes" range:"0,104857600"`
}

// Media holds on-demand image processing limits.
type Media struct {
	Concurrency    int   `json:"concurrency"       label:"Concurrency"     desc:"Max concurrent image resizes"   range:"1,16"`
	MaxSourceBytes int64 `json:"max_source_bytes"  label:"Max Source"      desc:"Max input image size"           unit:"bytes" range:"1048576,52428800"`
	WidthStep      int   `json:"width_step"        label:"Width Step"      desc:"Cache widths divisible by this" range:"10,200"`
	CacheMemoryMB  int   `json:"cache_memory_mb"   label:"Cache Memory"    desc:"In-memory LRU for variants"     unit:"MB" range:"0,512"`
}

var cachedLimits *Limits

// GetLimits probes the system and returns resource limits.
func GetLimits() *Limits {
	if cachedLimits != nil {
		return cachedLimits
	}

	totalRAM := getMemoryLimit()
	cpuCount := runtime.NumCPU()

	// VFS Cache: 25% of Total RAM
	maxVFS := totalRAM / 4

	// Max Upload: 10% of RAM, capped at 100MB, min 10MB
	maxUpload := totalRAM / 10
	if maxUpload > 100*1024*1024 {
		maxUpload = 100 * 1024 * 1024
	}
	if maxUpload < 10*1024*1024 {
		maxUpload = 10 * 1024 * 1024
	}

	// Capacity estimates based on stress testing (fazt v0.10.10, Jan 2026)
	// Baseline: $6 VPS = 1 vCPU, 1GB RAM
	baseUsers := 50
	baseReads := 200
	baseWrites := 100
	baseMixed := 150

	// Scale with CPU cores (diminishing returns after 4 cores for SQLite)
	scaleFactor := cpuCount
	if scaleFactor > 4 {
		scaleFactor = 4 + (cpuCount-4)/2
	}

	// Media cache: 2% of RAM, min 16MB, max 256MB
	mediaCacheMB := int(totalRAM / (1024 * 1024) / 50) // 2%
	if mediaCacheMB < 16 {
		mediaCacheMB = 16
	}
	if mediaCacheMB > 256 {
		mediaCacheMB = 256
	}

	// Media concurrency: 2 per CPU core, min 2, max 8
	mediaConcurrency := 2 * cpuCount
	if mediaConcurrency < 2 {
		mediaConcurrency = 2
	}
	if mediaConcurrency > 8 {
		mediaConcurrency = 8
	}

	// Net global concurrency scales with CPU
	netConcurrency := 10 * scaleFactor
	if netConcurrency < 20 {
		netConcurrency = 20
	}
	if netConcurrency > 100 {
		netConcurrency = 100
	}

	cachedLimits = &Limits{
		Hardware: Hardware{
			TotalRAM:     totalRAM,
			AvailableRAM: totalRAM,
			CPUCores:     cpuCount,
		},
		Storage: Storage{
			MaxVFS:      maxVFS,
			MaxUpload:   maxUpload,
			WriteQueue:  1000,
			MaxFileSize: 100 * 1024 * 1024,  // 100MB
			MaxSiteSize: 500 * 1024 * 1024,  // 500MB
			MaxLogRows:  500000,              // ~100MB of activity logs
		},
		Runtime: Runtime{
			ExecTimeout: 5000,            // 5s
			MaxMemory:   50 * 1024 * 1024, // 50MB per execution
		},
		Capacity: Capacity{
			Users:       baseUsers * scaleFactor,
			UsersMax:    baseUsers * scaleFactor * 2,
			Reads:       baseReads * scaleFactor,
			Writes:      baseWrites + (scaleFactor-1)*25,
			Mixed:       baseMixed * scaleFactor,
			MaxRequests: 500,
			Timeout:     5000,
		},
		Media: Media{
			Concurrency:    mediaConcurrency,
			MaxSourceBytes: 20 * 1024 * 1024, // 20MB — large phone photos are ~12MB
			WidthStep:      50,
			CacheMemoryMB:  mediaCacheMB,
		},
		Net: Net{
			MaxCalls:       5,
			CallTimeout:    4000,              // 4s
			Budget:         4000,              // 4s
			AppConcurrency: 5,
			Concurrency:    netConcurrency,
			MaxRequestBody: 1 * 1024 * 1024,   // 1MB
			MaxResponse:    1 * 1024 * 1024,    // 1MB
			MaxRedirects:   3,
			RateLimit:      0,                  // disabled by default
			RateBurst:      0,
			LogBufferSize:  1000,
			LogFlushMs:     1000,
			CacheMaxItems:  0,                  // disabled by default
			CacheMaxBytes:  0,
		},
	}

	return cachedLimits
}

// ResetCachedLimits clears the cached limits (for testing).
func ResetCachedLimits() {
	cachedLimits = nil
}

// getMemoryLimit tries to find the container/host memory limit.
func getMemoryLimit() int64 {
	// 1. Try Cgroup V2
	if limit, err := readInt64("/sys/fs/cgroup/memory.max"); err == nil && limit > 0 {
		return limit
	}

	// 2. Try Cgroup V1
	if limit, err := readInt64("/sys/fs/cgroup/memory/memory.limit_in_bytes"); err == nil && limit > 0 {
		if limit < 1<<50 {
			return limit
		}
	}

	// 3. Fallback to Host RAM (/proc/meminfo)
	if limit, err := getHostTotalRAM(); err == nil && limit > 0 {
		return limit
	}

	// 4. Fallback: safe default for a small VPS (512MB)
	return 512 * 1024 * 1024
}

// getHostTotalRAM reads MemTotal from /proc/meminfo.
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
		return 0, nil
	}
	return strconv.ParseInt(s, 10, 64)
}
