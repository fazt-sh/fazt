package harness

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fazt-sh/fazt/internal/harness/gaps"
)

// Report represents the complete harness report.
type Report struct {
	// Metadata
	Timestamp  time.Time `json:"timestamp"`
	Version    string    `json:"version"`
	TargetURL  string    `json:"target_url"`
	Duration   time.Duration `json:"duration"`

	// Summary
	Summary ReportSummary `json:"summary"`

	// Baseline measurements
	Throughput []ThroughputResult `json:"throughput,omitempty"`
	Latency    []LatencyResult    `json:"latency,omitempty"`
	Resources  []ResourceResult   `json:"resources,omitempty"`

	// Test results by category
	RequestTests    []TestResult `json:"request_tests,omitempty"`
	ResilienceTests []TestResult `json:"resilience_tests,omitempty"`
	SecurityTests   []TestResult `json:"security_tests,omitempty"`

	// Gaps discovered
	Gaps []*gaps.Gap `json:"gaps,omitempty"`

	// Comparison to baseline (if provided)
	Regressions []Regression `json:"regressions,omitempty"`
}

// ReportSummary holds high-level statistics.
type ReportSummary struct {
	TotalTests   int `json:"total_tests"`
	PassedTests  int `json:"passed_tests"`
	FailedTests  int `json:"failed_tests"`
	SkippedTests int `json:"skipped_tests"`

	GapsCritical int `json:"gaps_critical"`
	GapsHigh     int `json:"gaps_high"`
	GapsMedium   int `json:"gaps_medium"`
	GapsLow      int `json:"gaps_low"`

	RegressionCount int `json:"regression_count"`

	// Quick health indicators
	HasBlockers bool `json:"has_blockers"`
	PassRate    float64 `json:"pass_rate"`
}

// Regression represents a performance regression from baseline.
type Regression struct {
	Metric       string  `json:"metric"`
	Scenario     string  `json:"scenario"`
	BaselineVal  float64 `json:"baseline_value"`
	CurrentVal   float64 `json:"current_value"`
	DeltaPercent float64 `json:"delta_percent"`
	Severity     string  `json:"severity"` // "warning" or "critical"
}

// NewReport creates a new report.
func NewReport(version, targetURL string) *Report {
	return &Report{
		Timestamp: time.Now(),
		Version:   version,
		TargetURL: targetURL,
	}
}

// AddTestResult adds a test result to the appropriate category.
func (r *Report) AddTestResult(result TestResult) {
	switch result.Category {
	case "requests", "static", "api_read", "api_write", "serverless", "auth":
		r.RequestTests = append(r.RequestTests, result)
	case "resilience", "memory", "disk", "queue", "timeout":
		r.ResilienceTests = append(r.ResilienceTests, result)
	case "security", "ratelimit", "payload", "slowloris":
		r.SecurityTests = append(r.SecurityTests, result)
	default:
		r.RequestTests = append(r.RequestTests, result)
	}
}

// AddThroughput adds throughput measurement.
func (r *Report) AddThroughput(result ThroughputResult) {
	r.Throughput = append(r.Throughput, result)
}

// AddLatency adds latency measurement.
func (r *Report) AddLatency(result LatencyResult) {
	r.Latency = append(r.Latency, result)
}

// AddResource adds resource measurement.
func (r *Report) AddResource(result ResourceResult) {
	r.Resources = append(r.Resources, result)
}

// SetGaps sets discovered gaps.
func (r *Report) SetGaps(gapList []*gaps.Gap) {
	r.Gaps = gapList
}

// Finalize calculates summary statistics.
func (r *Report) Finalize(gapTracker *gaps.Tracker) {
	// Count test results
	allTests := append(append(r.RequestTests, r.ResilienceTests...), r.SecurityTests...)
	r.Summary.TotalTests = len(allTests)
	for _, t := range allTests {
		if t.Passed {
			r.Summary.PassedTests++
		} else if t.Error != nil {
			r.Summary.FailedTests++
		} else {
			r.Summary.SkippedTests++
		}
	}

	if r.Summary.TotalTests > 0 {
		r.Summary.PassRate = float64(r.Summary.PassedTests) / float64(r.Summary.TotalTests)
	}

	// Count gaps
	if gapTracker != nil {
		counts := gapTracker.Count()
		r.Summary.GapsCritical = counts[gaps.SeverityCritical]
		r.Summary.GapsHigh = counts[gaps.SeverityHigh]
		r.Summary.GapsMedium = counts[gaps.SeverityMedium]
		r.Summary.GapsLow = counts[gaps.SeverityLow]
		r.Summary.HasBlockers = gapTracker.HasBlockers()
		r.Gaps = gapTracker.All()
	}

	r.Summary.RegressionCount = len(r.Regressions)
}

// CompareToBaseline compares current results to a baseline report.
func (r *Report) CompareToBaseline(baseline *Report, tolerance float64) {
	if baseline == nil {
		return
	}

	// Compare throughput
	for _, current := range r.Throughput {
		for _, base := range baseline.Throughput {
			if current.Scenario == base.Scenario && current.Concurrency == base.Concurrency {
				delta := (current.RPS - base.RPS) / base.RPS
				if delta < -tolerance {
					severity := "warning"
					if delta < -2*tolerance {
						severity = "critical"
					}
					r.Regressions = append(r.Regressions, Regression{
						Metric:       "throughput_rps",
						Scenario:     fmt.Sprintf("%s@%d", current.Scenario, current.Concurrency),
						BaselineVal:  base.RPS,
						CurrentVal:   current.RPS,
						DeltaPercent: delta * 100,
						Severity:     severity,
					})
				}
			}
		}
	}

	// Compare latency (P99)
	for _, current := range r.Latency {
		for _, base := range baseline.Latency {
			if current.Scenario == base.Scenario {
				delta := float64(current.P99-base.P99) / float64(base.P99)
				if delta > tolerance {
					severity := "warning"
					if delta > 2*tolerance {
						severity = "critical"
					}
					r.Regressions = append(r.Regressions, Regression{
						Metric:       "latency_p99",
						Scenario:     current.Scenario,
						BaselineVal:  float64(base.P99.Milliseconds()),
						CurrentVal:   float64(current.P99.Milliseconds()),
						DeltaPercent: delta * 100,
						Severity:     severity,
					})
				}
			}
		}
	}
}

// ToJSON serializes the report to JSON.
func (r *Report) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// ToMarkdown generates a markdown report.
func (r *Report) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString("# Fazt Harness Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n", r.Timestamp.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Version:** %s\n", r.Version))
	sb.WriteString(fmt.Sprintf("**Target:** %s\n", r.TargetURL))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n\n", r.Duration))

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("| Metric | Value |\n"))
	sb.WriteString(fmt.Sprintf("|--------|-------|\n"))
	sb.WriteString(fmt.Sprintf("| Total Tests | %d |\n", r.Summary.TotalTests))
	sb.WriteString(fmt.Sprintf("| Passed | %d |\n", r.Summary.PassedTests))
	sb.WriteString(fmt.Sprintf("| Failed | %d |\n", r.Summary.FailedTests))
	sb.WriteString(fmt.Sprintf("| Pass Rate | %.1f%% |\n", r.Summary.PassRate*100))
	sb.WriteString(fmt.Sprintf("| Regressions | %d |\n", r.Summary.RegressionCount))
	sb.WriteString(fmt.Sprintf("| Blockers | %v |\n\n", r.Summary.HasBlockers))

	// Gaps summary
	if r.Summary.GapsCritical+r.Summary.GapsHigh+r.Summary.GapsMedium+r.Summary.GapsLow > 0 {
		sb.WriteString("### Gaps by Severity\n\n")
		sb.WriteString("| Severity | Count |\n")
		sb.WriteString("|----------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Critical | %d |\n", r.Summary.GapsCritical))
		sb.WriteString(fmt.Sprintf("| High | %d |\n", r.Summary.GapsHigh))
		sb.WriteString(fmt.Sprintf("| Medium | %d |\n", r.Summary.GapsMedium))
		sb.WriteString(fmt.Sprintf("| Low | %d |\n\n", r.Summary.GapsLow))
	}

	// Throughput results
	if len(r.Throughput) > 0 {
		sb.WriteString("## Throughput Baselines\n\n")
		sb.WriteString("| Scenario | Concurrency | RPS | Expected | Status |\n")
		sb.WriteString("|----------|-------------|-----|----------|--------|\n")
		for _, t := range r.Throughput {
			status := "PASS"
			if !t.WithinBaseline {
				status = "FAIL"
			}
			sb.WriteString(fmt.Sprintf("| %s | %d | %.0f | %.0f | %s |\n",
				t.Scenario, t.Concurrency, t.RPS, t.ExpectedRPS, status))
		}
		sb.WriteString("\n")
	}

	// Latency results
	if len(r.Latency) > 0 {
		sb.WriteString("## Latency Baselines\n\n")
		sb.WriteString("| Scenario | P50 | P95 | P99 | Max |\n")
		sb.WriteString("|----------|-----|-----|-----|-----|\n")
		for _, l := range r.Latency {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				l.Scenario,
				formatDuration(l.P50),
				formatDuration(l.P95),
				formatDuration(l.P99),
				formatDuration(l.Max)))
		}
		sb.WriteString("\n")
	}

	// Resource results
	if len(r.Resources) > 0 {
		sb.WriteString("## Resource Usage\n\n")
		sb.WriteString("| Scenario | Heap (MB) | Goroutines | GC Pause |\n")
		sb.WriteString("|----------|-----------|------------|----------|\n")
		for _, res := range r.Resources {
			sb.WriteString(fmt.Sprintf("| %s | %d | %d | %s |\n",
				res.Scenario, res.HeapAllocMB, res.GoroutineCount,
				formatDuration(time.Duration(res.GCPauseNs))))
		}
		sb.WriteString("\n")
	}

	// Test results
	if len(r.RequestTests) > 0 {
		sb.WriteString("## Request Tests\n\n")
		writeTestResults(&sb, r.RequestTests)
	}

	if len(r.ResilienceTests) > 0 {
		sb.WriteString("## Resilience Tests\n\n")
		writeTestResults(&sb, r.ResilienceTests)
	}

	if len(r.SecurityTests) > 0 {
		sb.WriteString("## Security Tests\n\n")
		writeTestResults(&sb, r.SecurityTests)
	}

	// Regressions
	if len(r.Regressions) > 0 {
		sb.WriteString("## Regressions\n\n")
		sb.WriteString("| Metric | Scenario | Baseline | Current | Delta | Severity |\n")
		sb.WriteString("|--------|----------|----------|---------|-------|----------|\n")
		for _, reg := range r.Regressions {
			sb.WriteString(fmt.Sprintf("| %s | %s | %.2f | %.2f | %.1f%% | %s |\n",
				reg.Metric, reg.Scenario, reg.BaselineVal, reg.CurrentVal,
				reg.DeltaPercent, reg.Severity))
		}
		sb.WriteString("\n")
	}

	// Gaps detail
	if len(r.Gaps) > 0 {
		sb.WriteString("## Discovered Gaps\n\n")
		for _, g := range r.Gaps {
			checkbox := "[ ]"
			if g.Resolved {
				checkbox = "[x]"
			}
			sb.WriteString(fmt.Sprintf("- %s **%s** [%s/%s]: %s\n",
				checkbox, g.ID, g.Category, g.Severity, g.Description))
			if g.Remediation != "" {
				sb.WriteString(fmt.Sprintf("  - Fix: %s\n", g.Remediation))
			}
		}
	}

	return sb.String()
}

// ToText generates a plain text report for terminal output.
func (r *Report) ToText() string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	sb.WriteString("                    FAZT HARNESS REPORT                        \n")
	sb.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	sb.WriteString(fmt.Sprintf("Version:  %s\n", r.Version))
	sb.WriteString(fmt.Sprintf("Target:   %s\n", r.TargetURL))
	sb.WriteString(fmt.Sprintf("Duration: %s\n\n", r.Duration))

	// Summary box
	sb.WriteString("┌─────────────────────────────────────┐\n")
	sb.WriteString("│             SUMMARY                 │\n")
	sb.WriteString("├─────────────────────────────────────┤\n")
	sb.WriteString(fmt.Sprintf("│ Tests:     %3d passed / %3d total   │\n",
		r.Summary.PassedTests, r.Summary.TotalTests))
	sb.WriteString(fmt.Sprintf("│ Pass Rate: %5.1f%%                   │\n", r.Summary.PassRate*100))
	sb.WriteString(fmt.Sprintf("│ Gaps:      %d critical, %d high      │\n",
		r.Summary.GapsCritical, r.Summary.GapsHigh))

	if r.Summary.HasBlockers {
		sb.WriteString("│ Status:    BLOCKED                  │\n")
	} else if r.Summary.FailedTests > 0 {
		sb.WriteString("│ Status:    ISSUES FOUND             │\n")
	} else {
		sb.WriteString("│ Status:    PASS                     │\n")
	}
	sb.WriteString("└─────────────────────────────────────┘\n\n")

	// Test results
	allTests := append(append(r.RequestTests, r.ResilienceTests...), r.SecurityTests...)
	if len(allTests) > 0 {
		sb.WriteString("Test Results:\n")
		sb.WriteString("─────────────\n")
		for _, t := range allTests {
			status := "PASS"
			if !t.Passed {
				status = "FAIL"
			}
			sb.WriteString(fmt.Sprintf("  [%s] %s (%s)\n", status, t.Name, t.Duration))
			if !t.Passed && t.Error != nil {
				sb.WriteString(fmt.Sprintf("         Error: %v\n", t.Error))
			}
		}
		sb.WriteString("\n")
	}

	// Gaps
	if len(r.Gaps) > 0 {
		sb.WriteString("Discovered Gaps:\n")
		sb.WriteString("────────────────\n")
		for _, g := range r.Gaps {
			sb.WriteString(fmt.Sprintf("  [%s] %s: %s\n", g.Severity, g.ID, g.Description))
		}
	}

	return sb.String()
}

func writeTestResults(sb *strings.Builder, tests []TestResult) {
	sb.WriteString("| Test | Status | Duration | Notes |\n")
	sb.WriteString("|------|--------|----------|-------|\n")
	for _, t := range tests {
		status := "PASS"
		notes := ""
		if !t.Passed {
			status = "FAIL"
			if t.Error != nil {
				notes = t.Error.Error()
			}
		}
		if t.Gap != nil {
			notes = fmt.Sprintf("Gap: %s", t.Gap.ID)
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			t.Name, status, formatDuration(t.Duration), notes))
	}
	sb.WriteString("\n")
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
