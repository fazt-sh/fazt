package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/harness"
	"github.com/fazt-sh/fazt/internal/harness/baseline"
	"github.com/fazt-sh/fazt/internal/harness/gaps"
	"github.com/fazt-sh/fazt/internal/harness/requests"
	"github.com/fazt-sh/fazt/internal/harness/resilience"
	"github.com/fazt-sh/fazt/internal/harness/security"
)

// handleHarnessCommand handles the 'fazt harness' subcommand
func handleHarnessCommand(args []string) {
	if len(args) < 1 {
		printHarnessHelp()
		return
	}

	cmd := args[0]
	switch cmd {
	case "run":
		handleHarnessRun(args[1:])
	case "smoke":
		handleHarnessSmoke(args[1:])
	case "report":
		handleHarnessReport(args[1:])
	case "compare":
		handleHarnessCompare(args[1:])
	case "--help", "-h", "help":
		printHarnessHelp()
	default:
		fmt.Printf("Unknown harness command: %s\n\n", cmd)
		printHarnessHelp()
		os.Exit(1)
	}
}

func handleHarnessRun(args []string) {
	fs := flag.NewFlagSet("harness run", flag.ExitOnError)
	targetURL := fs.String("target", "http://localhost:8080", "Target URL to test")
	category := fs.String("category", "", "Run specific category (baseline, requests, resilience, security)")
	format := fs.String("format", "text", "Output format (text, json, markdown)")
	output := fs.String("output", "", "Output file (default: stdout)")
	duration := fs.Duration("duration", 10*time.Second, "Baseline test duration")
	concurrency := fs.Int("concurrency", 100, "Max concurrency for load tests")
	authToken := fs.String("auth", "", "API auth token")
	fs.Parse(args)

	fmt.Printf("Fazt Test Harness v%s\n", config.Version)
	fmt.Printf("Target: %s\n\n", *targetURL)

	cfg := harness.DefaultConfig()
	cfg.TargetURL = *targetURL
	cfg.Baseline.Duration = *duration
	cfg.Output.Format = *format
	if *output != "" {
		cfg.Output.OutputPath = *output
	}

	if *category != "" {
		cfg.Categories = []string{*category}
	}

	h := harness.New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Run the harness
	report, err := h.Run(ctx, config.Version)
	if err != nil {
		fmt.Printf("Error running harness: %v\n", err)
		os.Exit(1)
	}

	// Also run individual test suites if specific categories requested
	if *category == "" || *category == "requests" {
		fmt.Println("Running request lifecycle tests...")
		runRequestTests(ctx, *targetURL, *authToken, h.GapTracker(), report)
	}

	if *category == "" || *category == "baseline" {
		fmt.Println("Running baseline measurements...")
		runBaselineTests(ctx, *targetURL, h.GapTracker(), report, *concurrency)
	}

	if *category == "" || *category == "resilience" {
		fmt.Println("Running resilience tests...")
		runResilienceTests(ctx, *targetURL, *authToken, h.GapTracker(), report)
	}

	if *category == "" || *category == "security" {
		fmt.Println("Running security tests...")
		runSecurityTests(ctx, *targetURL, *authToken, h.GapTracker(), report)
	}

	// Finalize report
	report.Finalize(h.GapTracker())

	// Output
	outputReport(report, cfg.Output)
}

func handleHarnessSmoke(args []string) {
	fs := flag.NewFlagSet("harness smoke", flag.ExitOnError)
	targetURL := fs.String("target", "http://localhost:8080", "Target URL to test")
	fs.Parse(args)

	fmt.Printf("Fazt Smoke Test - Target: %s\n", *targetURL)

	cfg := harness.DefaultConfig()
	cfg.TargetURL = *targetURL

	h := harness.New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	report, err := h.RunSmoke(ctx, config.Version)
	if err != nil {
		fmt.Printf("Smoke test failed: %v\n", err)
		os.Exit(1)
	}

	// Quick summary
	fmt.Printf("\nResults: %d/%d tests passed\n", report.Summary.PassedTests, report.Summary.TotalTests)

	if report.Summary.HasBlockers {
		fmt.Println("Status: BLOCKED (critical gaps found)")
		os.Exit(1)
	} else if report.Summary.FailedTests > 0 {
		fmt.Println("Status: ISSUES FOUND")
		os.Exit(1)
	}
	fmt.Println("Status: PASS")
}

func handleHarnessReport(args []string) {
	fs := flag.NewFlagSet("harness report", flag.ExitOnError)
	format := fs.String("format", "text", "Output format (text, json, markdown)")
	inputFile := fs.String("input", "", "Input JSON results file")
	fs.Parse(args)

	if *inputFile == "" {
		fmt.Println("Error: --input file required")
		os.Exit(1)
	}

	// Read and parse existing results
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	report := &harness.Report{}
	if err := parseJSON(data, report); err != nil {
		fmt.Printf("Error parsing report: %v\n", err)
		os.Exit(1)
	}

	outputReport(report, harness.OutputConfig{Format: *format})
}

func handleHarnessCompare(args []string) {
	fs := flag.NewFlagSet("harness compare", flag.ExitOnError)
	baselineFile := fs.String("baseline", "", "Baseline results JSON file")
	currentFile := fs.String("current", "", "Current results JSON file")
	tolerance := fs.Float64("tolerance", 0.15, "Tolerance for regression detection (0.15 = 15%)")
	fs.Parse(args)

	if *baselineFile == "" || *currentFile == "" {
		fmt.Println("Error: --baseline and --current files required")
		os.Exit(1)
	}

	// Load both reports
	baseData, err := os.ReadFile(*baselineFile)
	if err != nil {
		fmt.Printf("Error reading baseline: %v\n", err)
		os.Exit(1)
	}

	currData, err := os.ReadFile(*currentFile)
	if err != nil {
		fmt.Printf("Error reading current: %v\n", err)
		os.Exit(1)
	}

	baseReport := &harness.Report{}
	currReport := &harness.Report{}

	if err := parseJSON(baseData, baseReport); err != nil {
		fmt.Printf("Error parsing baseline: %v\n", err)
		os.Exit(1)
	}
	if err := parseJSON(currData, currReport); err != nil {
		fmt.Printf("Error parsing current: %v\n", err)
		os.Exit(1)
	}

	// Compare
	currReport.CompareToBaseline(baseReport, *tolerance)

	fmt.Printf("Comparing %s (baseline) vs %s (current)\n\n", baseReport.Version, currReport.Version)

	if len(currReport.Regressions) == 0 {
		fmt.Println("No regressions detected.")
	} else {
		fmt.Printf("Found %d regressions:\n\n", len(currReport.Regressions))
		for _, r := range currReport.Regressions {
			fmt.Printf("  [%s] %s: %s\n", r.Severity, r.Metric, r.Scenario)
			fmt.Printf("         Baseline: %.2f -> Current: %.2f (%.1f%%)\n\n",
				r.BaselineVal, r.CurrentVal, r.DeltaPercent)
		}
		os.Exit(1)
	}
}

func printHarnessHelp() {
	fmt.Printf("fazt.sh %s - Test Harness\n", config.Version)
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  fazt harness <command> [options]")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  run        Run full test harness")
	fmt.Println("  smoke      Quick smoke test")
	fmt.Println("  report     Generate report from results")
	fmt.Println("  compare    Compare current vs baseline")
	fmt.Println()
	fmt.Println("OPTIONS (for run):")
	fmt.Println("  --target URL        Target URL (default: http://localhost:8080)")
	fmt.Println("  --category NAME     Run specific category")
	fmt.Println("  --format FORMAT     Output format: text, json, markdown")
	fmt.Println("  --output FILE       Write to file instead of stdout")
	fmt.Println("  --duration DUR      Baseline test duration (default: 10s)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Run full harness against local server")
	fmt.Println("  fazt harness run --target http://192.168.64.3:8080")
	fmt.Println()
	fmt.Println("  # Quick smoke test")
	fmt.Println("  fazt harness smoke --target http://localhost:8080")
	fmt.Println()
	fmt.Println("  # Run only baseline measurements")
	fmt.Println("  fazt harness run --category baseline")
	fmt.Println()
	fmt.Println("  # Compare against baseline")
	fmt.Println("  fazt harness compare --baseline v0.10.json --current v0.11.json")
	fmt.Println()
}

// Helper functions

func runRequestTests(ctx context.Context, baseURL, authToken string, gapTracker *gaps.Tracker, report *harness.Report) {
	// Static tests
	staticRunner := requests.NewStaticRunner(baseURL, gapTracker)
	for _, result := range staticRunner.Run(ctx) {
		report.AddTestResult(result)
	}

	// API read tests
	apiReadRunner := requests.NewAPIReadRunner(baseURL, authToken, gapTracker)
	for _, result := range apiReadRunner.Run(ctx) {
		report.AddTestResult(result)
	}

	// Auth tests (if credentials available)
	// authRunner := requests.NewAuthRunner(baseURL, "admin", "password", gapTracker)
	// for _, result := range authRunner.Run(ctx) {
	// 	report.AddTestResult(result)
	// }
}

func runBaselineTests(ctx context.Context, baseURL string, gapTracker *gaps.Tracker, report *harness.Report, maxConcurrency int) {
	// Throughput measurements
	throughputRunner := baseline.NewThroughputRunner(baseURL, gapTracker)
	for _, result := range throughputRunner.Run(ctx, 10*time.Second, 2*time.Second) {
		report.AddThroughput(result)
	}

	// Latency measurements
	latencyRunner := baseline.NewLatencyRunner(baseURL, gapTracker)
	for _, result := range latencyRunner.Run(ctx, 1000) {
		report.AddLatency(result)
	}

	// Resource measurements
	resourceRunner := baseline.NewResourceRunner(baseURL, gapTracker)
	for _, result := range resourceRunner.Run(ctx) {
		report.AddResource(result)
	}
}

func runResilienceTests(ctx context.Context, baseURL, authToken string, gapTracker *gaps.Tracker, report *harness.Report) {
	// Memory tests
	memoryRunner := resilience.NewMemoryRunner(baseURL, gapTracker)
	for _, result := range memoryRunner.Run(ctx) {
		report.AddTestResult(result)
	}

	// Queue tests
	queueRunner := resilience.NewQueueRunner(baseURL, authToken, gapTracker)
	for _, result := range queueRunner.Run(ctx) {
		report.AddTestResult(result)
	}

	// Timeout tests
	timeoutRunner := resilience.NewTimeoutRunner(baseURL, authToken, gapTracker)
	for _, result := range timeoutRunner.Run(ctx) {
		report.AddTestResult(result)
	}
}

func runSecurityTests(ctx context.Context, baseURL, authToken string, gapTracker *gaps.Tracker, report *harness.Report) {
	// Rate limit tests
	rateLimitRunner := security.NewRateLimitRunner(baseURL, gapTracker)
	for _, result := range rateLimitRunner.Run(ctx) {
		report.AddTestResult(result)
	}

	// Payload tests
	payloadRunner := security.NewPayloadRunner(baseURL, authToken, gapTracker)
	for _, result := range payloadRunner.Run(ctx) {
		report.AddTestResult(result)
	}

	// Slowloris tests
	slowlorisRunner := security.NewSlowlorisRunner(baseURL, gapTracker)
	for _, result := range slowlorisRunner.Run(ctx) {
		report.AddTestResult(result)
	}
}

func outputReport(report *harness.Report, cfg harness.OutputConfig) {
	var output string

	switch cfg.Format {
	case "json":
		data, _ := report.ToJSON()
		output = string(data)
	case "markdown":
		output = report.ToMarkdown()
	default:
		output = report.ToText()
	}

	if cfg.OutputPath != "" {
		if err := os.WriteFile(cfg.OutputPath, []byte(output), 0644); err != nil {
			fmt.Printf("Error writing output: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Report written to %s\n", cfg.OutputPath)
	} else {
		fmt.Println(output)
	}
}

func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
