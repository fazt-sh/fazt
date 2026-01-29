// Package harness provides a comprehensive test harness for fazt performance,
// resilience, and security validation.
package harness

import (
	"time"

	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// Config holds harness configuration.
type Config struct {
	// Target URL to test against
	TargetURL string

	// Database path for direct DB tests (optional)
	DatabasePath string

	// Test categories to run
	Categories []string

	// Baseline configuration
	Baseline BaselineConfig

	// Resilience configuration
	Resilience ResilienceConfig

	// Security configuration
	Security SecurityConfig

	// Output configuration
	Output OutputConfig
}

// BaselineConfig configures baseline measurements.
type BaselineConfig struct {
	// Duration for throughput tests
	Duration time.Duration

	// Concurrency levels to test
	ConcurrencyLevels []int

	// Tolerance for baseline comparison (e.g., 0.10 for ±10%)
	Tolerance float64

	// WarmupDuration before measurements
	WarmupDuration time.Duration
}

// ResilienceConfig configures resilience tests.
type ResilienceConfig struct {
	// MemoryPressureMB for memory tests
	MemoryPressureMB int

	// DiskFillPercent for disk tests
	DiskFillPercent int

	// QueueDepth for queue saturation tests
	QueueDepth int
}

// SecurityConfig configures security tests.
type SecurityConfig struct {
	// RateLimitBurst for rate limit tests
	RateLimitBurst int

	// MaxPayloadMB for payload tests
	MaxPayloadMB int

	// SlowClientCount for slowloris tests
	SlowClientCount int

	// SlowClientByteRate (bytes per second)
	SlowClientByteRate int
}

// OutputConfig configures report output.
type OutputConfig struct {
	// Format: "json", "markdown", "text"
	Format string

	// OutputPath for report file (empty for stdout)
	OutputPath string

	// Verbose enables detailed output
	Verbose bool
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		TargetURL: "http://localhost:8080",
		Categories: []string{
			"baseline",
			"requests",
			"resilience",
			"security",
		},
		Baseline: BaselineConfig{
			Duration:          10 * time.Second,
			ConcurrencyLevels: []int{1, 10, 50, 100},
			Tolerance:         0.15,
			WarmupDuration:    2 * time.Second,
		},
		Resilience: ResilienceConfig{
			MemoryPressureMB: 100,
			DiskFillPercent:  80,
			QueueDepth:       500,
		},
		Security: SecurityConfig{
			RateLimitBurst:     20,
			MaxPayloadMB:       10,
			SlowClientCount:    50,
			SlowClientByteRate: 1,
		},
		Output: OutputConfig{
			Format:  "text",
			Verbose: false,
		},
	}
}

// Expected represents expected test outcomes.
type Expected struct {
	Status       int
	MaxLatency   time.Duration
	MinLatency   time.Duration
	BodyContains string
	HeaderKey    string
	HeaderValue  string
	Error        string
}

// Expect creates an Expected with common defaults.
func Expect(status int) Expected {
	return Expected{Status: status}
}

// ExpectWithLatency creates Expected with latency bounds.
func ExpectWithLatency(status int, maxLatency time.Duration) Expected {
	return Expected{
		Status:     status,
		MaxLatency: maxLatency,
	}
}

// ExpectError creates Expected for error cases.
func ExpectError(status int, errorContains string) Expected {
	return Expected{
		Status: status,
		Error:  errorContains,
	}
}

// WithBody adds body content check.
func (e Expected) WithBody(contains string) Expected {
	e.BodyContains = contains
	return e
}

// WithHeader adds header check.
func (e Expected) WithHeader(key, value string) Expected {
	e.HeaderKey = key
	e.HeaderValue = value
	return e
}

// TestResult represents the outcome of a single test.
type TestResult struct {
	Name     string
	Category string
	Passed   bool
	Duration time.Duration
	Expected Expected
	Actual   Actual
	Error    error
	Gap      *gaps.Gap // Non-nil if a gap was discovered
}

// Actual represents actual test outcomes.
type Actual struct {
	Status     int
	Latency    time.Duration
	Body       string
	Headers    map[string]string
	Error      string
	StatusCode int
}

// ThroughputResult holds throughput measurement results.
type ThroughputResult struct {
	Scenario       string
	Duration       time.Duration
	Concurrency    int
	TotalRequests  int64
	SuccessCount   int64
	ErrorCount     int64
	RPS            float64
	ExpectedRPS    float64
	WithinBaseline bool
	Errors         map[string]int // Error type -> count
}

// LatencyResult holds latency measurement results.
type LatencyResult struct {
	Scenario   string
	P50        time.Duration
	P95        time.Duration
	P99        time.Duration
	Min        time.Duration
	Max        time.Duration
	Mean       time.Duration
	SampleSize int
}

// ResourceResult holds resource usage results.
type ResourceResult struct {
	Scenario       string
	HeapAllocMB    uint64
	HeapSysMB      uint64
	GoroutineCount int
	GCPauseNs      uint64
	CPUPercent     float64
}
