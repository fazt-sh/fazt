package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/fazt-sh/fazt/internal/activity"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/output"
)

// Common filter flags shared across list, cleanup, and export
type filterFlags struct {
	minWeight    *int
	maxWeight    *int
	resourceType *string
	resourceID   *string
	appID        *string
	userID       *string
	actorType    *string
	action       *string
	result       *string
	since        *string
	until        *string
	limit        *int
	offset       *int
	dbPath       *string
}

// addFilterFlags adds common filter flags to a FlagSet
func addFilterFlags(flags *flag.FlagSet) *filterFlags {
	f := &filterFlags{}
	f.minWeight = flags.Int("min-weight", -1, "Minimum weight (0-9)")
	f.maxWeight = flags.Int("max-weight", -1, "Maximum weight (0-9)")
	f.resourceType = flags.String("type", "", "Filter by resource type (app/user/session/kv/doc/page/config)")
	f.resourceID = flags.String("resource", "", "Filter by resource ID")
	f.appID = flags.String("app", "", "Filter by app ID (matches app resources, KV, pages)")
	f.userID = flags.String("user", "", "Filter by user ID")
	f.actorType = flags.String("actor-type", "", "Filter by actor type (user/system/api_key/anonymous)")
	f.action = flags.String("action", "", "Filter by action")
	f.result = flags.String("result", "", "Filter by result (success/failure)")
	f.since = flags.String("since", "", "Show entries since (e.g., '24h', '7d', '2024-01-15')")
	f.until = flags.String("until", "", "Show entries until (e.g., '24h', '7d', '2024-01-15')")
	f.limit = flags.Int("limit", activity.DefaultLimit, "Max results to return")
	f.offset = flags.Int("offset", 0, "Skip first n results")
	f.dbPath = flags.String("db", "", "Database path")
	return f
}

// toQueryParams converts filter flags to QueryParams
func (f *filterFlags) toQueryParams() (activity.QueryParams, error) {
	params := activity.QueryParams{
		Limit:  *f.limit,
		Offset: *f.offset,
	}

	if *f.minWeight >= 0 {
		params.MinWeight = f.minWeight
	}
	if *f.maxWeight >= 0 {
		params.MaxWeight = f.maxWeight
	}
	if *f.resourceType != "" {
		params.ResourceType = *f.resourceType
	}
	if *f.resourceID != "" {
		params.ResourceID = *f.resourceID
	}
	if *f.appID != "" {
		params.AppID = *f.appID
	}
	if *f.userID != "" {
		params.UserID = *f.userID
	}
	if *f.actorType != "" {
		params.ActorType = *f.actorType
	}
	if *f.action != "" {
		params.Action = *f.action
	}
	if *f.result != "" {
		params.Result = *f.result
	}
	if *f.since != "" {
		t, err := parseDuration(*f.since)
		if err != nil {
			return params, fmt.Errorf("invalid --since: %v", err)
		}
		params.Since = &t
	}
	if *f.until != "" {
		t, err := parseDuration(*f.until)
		if err != nil {
			return params, fmt.Errorf("invalid --until: %v", err)
		}
		params.Until = &t
	}

	return params, nil
}

// handleLogsCommand handles activity log subcommands
func handleLogsCommand(args []string) {
	if len(args) < 1 {
		printLogsHelp()
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "list":
		handleLogsList(args[1:])
	case "cleanup":
		handleLogsCleanup(args[1:])
	case "export":
		handleLogsExport(args[1:])
	case "stats":
		handleLogsStats(args[1:])
	case "--help", "-h", "help":
		printLogsHelp()
	default:
		fmt.Printf("Unknown logs command: %s\n\n", subcommand)
		printLogsHelp()
		os.Exit(1)
	}
}

// handleLogsList lists activity log entries
func handleLogsList(args []string) {
	flags := flag.NewFlagSet("logs list", flag.ExitOnError)
	f := addFilterFlags(flags)
	flags.Parse(args)

	params, err := f.toQueryParams()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Resolve DB path and initialize
	resolvedDB := resolveDBPath(*f.dbPath)
	if err := database.Init(resolvedDB); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	entries, total, err := activity.Query(database.GetDB(), params)
	if err != nil {
		fmt.Printf("Error querying logs: %v\n", err)
		os.Exit(1)
	}

	renderer := getRenderer()

	if len(entries) == 0 {
		renderer.Print("No activity logs found.", map[string]interface{}{"entries": []interface{}{}, "total": 0})
		return
	}

	// Build table
	table := &output.Table{
		Headers: []string{"Time", "W", "Actor", "Resource", "Action", "Result"},
		Rows:    make([][]string, len(entries)),
	}

	entriesData := make([]map[string]interface{}, len(entries))
	for i, e := range entries {
		timeStr := formatTime(e.Timestamp)

		actor := e.ActorType
		if e.ActorID != "" {
			actor = e.ActorID
		}
		if len(actor) > 12 {
			actor = actor[:9] + "..."
		}

		resource := e.ResourceType
		if e.ResourceID != "" {
			resource = fmt.Sprintf("%s:%s", e.ResourceType, truncate(e.ResourceID, 15))
		}

		table.Rows[i] = []string{
			timeStr,
			fmt.Sprintf("%d", e.Weight),
			actor,
			resource,
			e.Action,
			e.Result,
		}

		entriesData[i] = map[string]interface{}{
			"id":            e.ID,
			"timestamp":     e.Timestamp.Format(time.RFC3339),
			"actor_type":    e.ActorType,
			"actor_id":      e.ActorID,
			"actor_ip":      e.ActorIP,
			"resource_type": e.ResourceType,
			"resource_id":   e.ResourceID,
			"action":        e.Action,
			"result":        e.Result,
			"weight":        e.Weight,
			"details":       e.Details,
		}
	}

	data := map[string]interface{}{
		"entries": entriesData,
		"total":   total,
		"showing": len(entries),
	}

	md := output.NewMarkdown().
		H1("Activity Logs").
		Table(table).
		Para(fmt.Sprintf("Showing %d of %d entries", len(entries), total)).
		String()

	renderer.Print(md, data)
}

// handleLogsCleanup cleans up activity logs with full filter support
// By default shows what would be deleted (safe). Use --force to actually delete.
func handleLogsCleanup(args []string) {
	flags := flag.NewFlagSet("logs cleanup", flag.ExitOnError)
	f := addFilterFlags(flags)
	force := flags.Bool("force", false, "Actually delete entries (default: preview only)")
	flags.Parse(args)

	params, err := f.toQueryParams()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Resolve DB path and initialize
	resolvedDB := resolveDBPath(*f.dbPath)
	if err := database.Init(resolvedDB); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	dryRun := !*force
	count, err := activity.Cleanup(database.GetDB(), params, dryRun)
	if err != nil {
		fmt.Printf("Error during cleanup: %v\n", err)
		os.Exit(1)
	}

	renderer := getRenderer()
	filterDesc := activity.DescribeFilters(params)

	data := map[string]interface{}{
		"count":   count,
		"filters": filterDesc,
		"force":   *force,
	}

	if dryRun {
		md := fmt.Sprintf("Would delete **%d** entries matching: %s\n\nUse --force to actually delete.", count, filterDesc)
		renderer.Print(md, data)
	} else {
		md := fmt.Sprintf("Deleted **%d** entries matching: %s", count, filterDesc)
		renderer.Print(md, data)
	}
}

// handleLogsExport exports activity logs to JSON or CSV
func handleLogsExport(args []string) {
	flags := flag.NewFlagSet("logs export", flag.ExitOnError)
	f := addFilterFlags(flags)
	exportFormat := flags.String("f", "json", "Export format: json or csv")
	outputFile := flags.String("o", "", "Output file (default: stdout)")
	flags.Parse(args)

	// For export, allow higher limits
	if *f.limit == activity.DefaultLimit {
		*f.limit = activity.MaxLimit
	}

	params, err := f.toQueryParams()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Resolve DB path and initialize
	resolvedDB := resolveDBPath(*f.dbPath)
	if err := database.Init(resolvedDB); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	entries, total, err := activity.Query(database.GetDB(), params)
	if err != nil {
		fmt.Printf("Error querying logs: %v\n", err)
		os.Exit(1)
	}

	// Determine output writer
	var out *os.File
	if *outputFile == "" {
		out = os.Stdout
	} else {
		out, err = os.Create(*outputFile)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer out.Close()
	}

	switch *exportFormat {
	case "json":
		exportJSON(out, entries, total)
	case "csv":
		exportCSV(out, entries)
	default:
		fmt.Printf("Unknown format: %s (use json or csv)\n", *exportFormat)
		os.Exit(1)
	}

	if *outputFile != "" {
		fmt.Fprintf(os.Stderr, "Exported %d entries to %s\n", len(entries), *outputFile)
	}
}

func exportJSON(out *os.File, entries []activity.LogEntry, total int) {
	data := map[string]interface{}{
		"entries":  entries,
		"total":    total,
		"exported": len(entries),
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}

func exportCSV(out *os.File, entries []activity.LogEntry) {
	w := csv.NewWriter(out)
	defer w.Flush()

	// Header
	w.Write([]string{"id", "timestamp", "actor_type", "actor_id", "actor_ip", "resource_type", "resource_id", "action", "result", "weight", "details"})

	for _, e := range entries {
		details := ""
		if e.Details != nil {
			b, _ := json.Marshal(e.Details)
			details = string(b)
		}
		w.Write([]string{
			fmt.Sprintf("%d", e.ID),
			e.Timestamp.Format(time.RFC3339),
			e.ActorType,
			e.ActorID,
			e.ActorIP,
			e.ResourceType,
			e.ResourceID,
			e.Action,
			e.Result,
			fmt.Sprintf("%d", e.Weight),
			details,
		})
	}
}

// handleLogsStats shows activity log statistics with optional filters
func handleLogsStats(args []string) {
	flags := flag.NewFlagSet("logs stats", flag.ExitOnError)
	f := addFilterFlags(flags)
	flags.Parse(args)

	params, err := f.toQueryParams()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resolvedDB := resolveDBPath(*f.dbPath)
	if err := database.Init(resolvedDB); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	stats, err := activity.GetStatsFiltered(database.GetDB(), params)
	if err != nil {
		fmt.Printf("Error getting stats: %v\n", err)
		os.Exit(1)
	}

	renderer := getRenderer()
	filterDesc := activity.DescribeFilters(params)

	weightLabels := map[int]string{
		0: "Debug",
		1: "System",
		2: "Analytics",
		3: "Navigation",
		4: "User Action",
		5: "Data Mutation",
		6: "Deployment",
		7: "Config",
		8: "Auth",
		9: "Security",
	}

	table := &output.Table{
		Headers: []string{"Weight", "Category", "Count"},
		Rows:    make([][]string, 0),
	}

	for w := 9; w >= 0; w-- {
		count := stats.CountByWeight[w]
		if count > 0 {
			table.Rows = append(table.Rows, []string{
				fmt.Sprintf("%d", w),
				weightLabels[w],
				fmt.Sprintf("%d", count),
			})
		}
	}

	var timeRange string
	if stats.OldestEntry != nil && stats.NewestEntry != nil {
		timeRange = fmt.Sprintf("%s to %s", stats.OldestEntry.Format("2006-01-02"), stats.NewestEntry.Format("2006-01-02"))
	} else {
		timeRange = "N/A"
	}

	data := map[string]interface{}{
		"total_count":     stats.TotalCount,
		"count_by_weight": stats.CountByWeight,
		"size_estimate":   stats.SizeEstimate,
		"oldest_entry":    stats.OldestEntry,
		"newest_entry":    stats.NewestEntry,
		"filters":         filterDesc,
	}

	mdBuilder := output.NewMarkdown().
		H1("Activity Log Stats")

	if filterDesc != "all entries" {
		mdBuilder = mdBuilder.Para(fmt.Sprintf("**Filters:** %s", filterDesc))
	}

	md := mdBuilder.
		Para(fmt.Sprintf("**Total entries:** %d", stats.TotalCount)).
		Para(fmt.Sprintf("**Size estimate:** %s", formatBytes(stats.SizeEstimate))).
		Para(fmt.Sprintf("**Time range:** %s", timeRange)).
		H2("Entries by Weight").
		Table(table).
		String()

	renderer.Print(md, data)
}

// printLogsHelp prints help for the logs command
func printLogsHelp() {
	fmt.Println(`Usage: fazt logs <command> [filters...]

Commands:
  list      List activity log entries
  cleanup   Delete entries matching filters
  export    Export entries to JSON or CSV
  stats     Show activity log statistics

Filters (work with all commands):
  --min-weight N    Minimum weight (0-9)
  --max-weight N    Maximum weight (0-9)
  --app ID          Filter by app ID
  --user ID         Filter by user ID
  --actor-type T    Filter by actor type (user/system/api_key/anonymous)
  --type T          Filter by resource type (app/session/kv/doc/page/config)
  --resource ID     Filter by resource ID
  --action A        Filter by action
  --result R        Filter by result (success/failure)
  --since TIME      Show entries since (e.g., '24h', '7d', '2024-01-15')
  --until TIME      Show entries until
  --limit N         Number of entries (default: 20)
  --offset N        Offset for pagination

Command-specific options:
  cleanup:
    --force         Actually delete (default: preview only)

  export:
    -f FORMAT       Output format: json or csv (default: json)
    -o FILE         Output file (default: stdout)

Examples:
  fazt logs list                                     # Recent activity
  fazt logs list --app my-app --since 24h           # App activity last 24h
  fazt logs list --user abc123 --min-weight 5       # User's important events
  fazt logs cleanup --app old-app                    # Preview cleanup (dry-run)
  fazt logs cleanup --max-weight 2 --until 7d --force  # Delete old analytics
  fazt logs stats --app my-app                       # Stats for specific app
  fazt logs export --app my-app --limit 1000 -f csv -o logs.csv

Weight Scale (0-9):
  9: Security    (API key changes, role changes)
  8: Auth        (login/logout, sessions)
  7: Config      (alias/redirect changes)
  6: Deployment  (site deploy/delete)
  5: Data        (KV/doc mutations)
  4: User Action (form submissions)
  3: Navigation  (page views)
  2: Analytics   (pageviews, clicks)
  1: System      (health checks, server start/stop)
  0: Debug       (timing, cache hits)`)
}

// Helper functions

func resolveDBPath(dbPath string) string {
	if dbPath != "" {
		return dbPath
	}
	if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
		return envPath
	}
	return "./data.db"
}

func parseDuration(s string) (time.Time, error) {
	duration, err := parseDurationValue(s)
	if err != nil {
		// Try parsing as date
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid duration or date: %s", s)
		}
		return t, nil
	}
	return time.Now().Add(-duration), nil
}

func parseDurationValue(s string) (time.Duration, error) {
	// Support 'd' for days
	if len(s) > 1 && s[len(s)-1] == 'd' {
		var days int
		_, err := fmt.Sscanf(s, "%dd", &days)
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

func formatTime(t time.Time) string {
	diff := time.Since(t)
	switch {
	case diff < time.Minute:
		return fmt.Sprintf("%ds ago", int(diff.Seconds()))
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	case diff < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	default:
		return t.Format("Jan 02")
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// formatBytes is defined in app_v2.go
