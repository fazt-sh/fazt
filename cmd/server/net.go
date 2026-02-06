package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/egress"
)

func handleNetCommand(args []string) {
	if len(args) < 1 {
		printNetUsage()
		return
	}

	switch args[0] {
	case "allow":
		handleNetAllow(args[1:])
	case "list":
		handleNetList(args[1:])
	case "remove":
		handleNetRemove(args[1:])
	case "--help", "-h", "help":
		printNetUsage()
	default:
		fmt.Printf("Unknown net subcommand: %s\n", args[0])
		printNetUsage()
		os.Exit(1)
	}
}

func printNetUsage() {
	fmt.Println("fazt net - Network egress management")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  fazt net <command> [options]")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  allow <domain>          Add domain to allowlist")
	fmt.Println("  list                    List allowed domains")
	fmt.Println("  remove <domain>         Remove domain from allowlist")
	fmt.Println()
	fmt.Println("OPTIONS (allow):")
	fmt.Println("  --app <id>              Scope to specific app")
	fmt.Println("  --http                  Allow HTTP (default: HTTPS only)")
	fmt.Println("  --rate <n>              Rate limit (requests/min)")
	fmt.Println("  --burst <n>             Rate burst allowance")
	fmt.Println("  --timeout <ms>          Per-call timeout override")
	fmt.Println("  --max-response <bytes>  Response size limit override")
	fmt.Println("  --cache-ttl <seconds>   Response cache TTL")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  fazt net allow api.stripe.com")
	fmt.Println("  fazt net allow api.openai.com --app myapp")
	fmt.Println("  fazt net allow \"*.googleapis.com\"")
	fmt.Println("  fazt net allow api.stripe.com --rate 60 --burst 10")
	fmt.Println("  fazt net list")
	fmt.Println("  fazt net remove api.old-service.com")
}

func handleNetAllow(args []string) {
	fs := flag.NewFlagSet("net allow", flag.ExitOnError)
	appFlag := fs.String("app", "", "Scope to app ID")
	httpFlag := fs.Bool("http", false, "Allow HTTP (default: HTTPS only)")
	rateFlag := fs.Int("rate", 0, "Rate limit (requests/min)")
	burstFlag := fs.Int("burst", 0, "Rate burst allowance")
	timeoutFlag := fs.Int("timeout", 0, "Per-call timeout (ms)")
	maxRespFlag := fs.Int64("max-response", 0, "Max response size (bytes)")
	cacheTTLFlag := fs.Int("cache-ttl", 0, "Response cache TTL (seconds)")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: domain required")
		fmt.Fprintln(os.Stderr, "Usage: fazt net allow <domain> [options]")
		os.Exit(1)
	}

	domain := fs.Arg(0)
	httpsOnly := !*httpFlag

	db := getClientDB()
	defer database.Close()

	allowlist := egress.NewAllowlist(db)
	if err := allowlist.Add(domain, *appFlag, httpsOnly); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Update extended config if provided
	if *rateFlag > 0 || *burstFlag > 0 || *timeoutFlag > 0 || *maxRespFlag > 0 || *cacheTTLFlag > 0 {
		var appIDVal interface{} = nil
		if *appFlag != "" {
			appIDVal = *appFlag
		}
		_, err := db.Exec(`
			UPDATE net_allowlist
			SET rate_limit = CASE WHEN ? > 0 THEN ? ELSE rate_limit END,
			    rate_burst = CASE WHEN ? > 0 THEN ? ELSE rate_burst END,
			    timeout_ms = CASE WHEN ? > 0 THEN ? ELSE timeout_ms END,
			    max_response = CASE WHEN ? > 0 THEN ? ELSE max_response END,
			    cache_ttl = CASE WHEN ? > 0 THEN ? ELSE cache_ttl END
			WHERE domain = ? AND app_id IS ?
		`, *rateFlag, *rateFlag,
			*burstFlag, *burstFlag,
			*timeoutFlag, *timeoutFlag,
			*maxRespFlag, *maxRespFlag,
			*cacheTTLFlag, *cacheTTLFlag,
			domain, appIDVal)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update config: %v\n", err)
		}
	}

	scope := "global"
	if *appFlag != "" {
		scope = fmt.Sprintf("app:%s", *appFlag)
	}
	proto := "HTTPS"
	if !httpsOnly {
		proto = "HTTP+HTTPS"
	}
	fmt.Printf("Allowed %s (%s, %s)\n", domain, proto, scope)
}

func handleNetList(args []string) {
	fs := flag.NewFlagSet("net list", flag.ExitOnError)
	appFlag := fs.String("app", "", "Filter by app ID")
	fs.Parse(args)

	db := getClientDB()
	defer database.Close()

	allowlist := egress.NewAllowlist(db)
	entries, err := allowlist.List(*appFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("No domains in allowlist")
		fmt.Println()
		fmt.Println("Add domains with: fazt net allow <domain>")
		return
	}

	// Header
	fmt.Printf("%-35s %-10s %-8s %-10s %-10s\n", "Domain", "Scope", "Proto", "Rate", "Created")
	fmt.Println(strings.Repeat("-", 80))

	for _, e := range entries {
		scope := "global"
		if e.AppID != "" {
			scope = e.AppID
		}
		proto := "HTTPS"
		if !e.HTTPSOnly {
			proto = "HTTP+"
		}
		rate := "-"
		if e.RateLimit > 0 {
			rate = fmt.Sprintf("%d/min", e.RateLimit)
		}
		created := time.Unix(e.CreatedAt, 0).Format("2006-01-02")
		fmt.Printf("%-35s %-10s %-8s %-10s %-10s\n", e.Domain, scope, proto, rate, created)
	}

	fmt.Printf("\n%d domain(s)\n", len(entries))
}

func handleNetRemove(args []string) {
	fs := flag.NewFlagSet("net remove", flag.ExitOnError)
	appFlag := fs.String("app", "", "Scope to app ID")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: domain required")
		fmt.Fprintln(os.Stderr, "Usage: fazt net remove <domain> [--app <id>]")
		os.Exit(1)
	}

	domain := fs.Arg(0)

	db := getClientDB()
	defer database.Close()

	allowlist := egress.NewAllowlist(db)
	if err := allowlist.Remove(domain, *appFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Removed %s from allowlist\n", domain)
}
