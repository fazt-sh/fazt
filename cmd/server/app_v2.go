package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/help"
	"github.com/fazt-sh/fazt/internal/output"
	"github.com/fazt-sh/fazt/internal/remote"
)

// handleAppCommandV2 routes app subcommands with v0.10 features
func handleAppCommandV2(args []string) {
	if len(args) < 1 {
		// Default: list apps on default peer
		handleAppListV2(nil)
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		handleAppListV2(args[1:])
	case "info":
		handleAppInfoV2(args[1:])
	case "deploy":
		handleAppDeploy(args[1:]) // Use existing deploy
	case "create":
		handleAppCreate(args[1:]) // Use existing create
	case "validate":
		handleAppValidate(args[1:]) // Use existing validate
	case "logs":
		handleAppLogs(args[1:]) // Use existing logs
	case "install":
		handleAppInstall(args[1:]) // Use existing install
	case "remove":
		handleAppRemoveV2(args[1:])
	case "link":
		handleAppLink(args[1:])
	case "unlink":
		handleAppUnlink(args[1:])
	case "reserve":
		handleAppReserve(args[1:])
	case "fork":
		handleAppFork(args[1:])
	case "swap":
		handleAppSwap(args[1:])
	case "split":
		handleAppSplit(args[1:])
	case "lineage":
		handleAppLineage(args[1:])
	case "upgrade":
		handleAppUpgrade(args[1:])
	case "pull":
		handleAppPull(args[1:])
	case "files":
		handleAppFiles(args[1:])
	case "--help", "-h", "help":
		printAppHelpV2()
	default:
		fmt.Printf("Unknown app command: %s\n\n", subcommand)
		printAppHelpV2()
		os.Exit(1)
	}
}

// handleAppListV2 lists apps on a peer with v0.10 format
func handleAppListV2(args []string) {
	showAliases := false

	// Use positional peer argument if provided, otherwise use global context
	var peerName string
	for i, arg := range args {
		if arg == "--aliases" {
			showAliases = true
		} else if !strings.HasPrefix(arg, "-") && peerName == "" {
			peerName = args[i]
		}
	}

	// If no positional peer, use global context
	if peerName == "" {
		peerName = targetPeerName
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, peerName)
	if err != nil {
		handlePeerError(err)
		os.Exit(1)
	}

	// Use command gateway
	result, err := executeRemoteCmd(peer, "app", []string{"list"})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	renderer := getRenderer()

	if showAliases {
		// Fetch aliases instead
		result, err = executeRemoteCmd(peer, "app", []string{"list", "--aliases"})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Build table for aliases
		table := &output.Table{
			Headers: []string{"Subdomain", "Type", "Target"},
			Rows:    [][]string{},
		}

		aliasData := []interface{}{}
		if aliases, ok := result.([]interface{}); ok {
			for _, a := range aliases {
				if alias, ok := a.(map[string]interface{}); ok {
					subdomain := getString(alias, "subdomain")
					aliasType := getString(alias, "type")
					target := "-"
					if targets, ok := alias["targets"].(map[string]interface{}); ok {
						if appID, ok := targets["app_id"].(string); ok {
							target = appID
						}
					}
					table.Rows = append(table.Rows, []string{subdomain, aliasType, target})
					aliasData = append(aliasData, alias)
				}
			}
		}

		data := map[string]interface{}{
			"peer":    peer.Name,
			"aliases": aliasData,
			"count":   len(table.Rows),
		}

		md := output.NewMarkdown().
			H1(fmt.Sprintf("Aliases on %s", peer.Name)).
			Table(table).
			Para(fmt.Sprintf("%d aliases", len(table.Rows))).
			String()

		renderer.Print(md, data)
		return
	}

	// Build table for apps
	table := &output.Table{
		Headers: []string{"ID", "Title", "Visibility", "Aliases"},
		Rows:    [][]string{},
	}

	appsData := []interface{}{}
	if apps, ok := result.([]interface{}); ok {
		for _, a := range apps {
			if app, ok := a.(map[string]interface{}); ok {
				id := getString(app, "id")
				title := getString(app, "title")
				visibility := getString(app, "visibility")
				aliases := "-"
				if aliasArr, ok := app["aliases"].([]interface{}); ok && len(aliasArr) > 0 {
					var aliasStrs []string
					for _, al := range aliasArr {
						if s, ok := al.(string); ok {
							aliasStrs = append(aliasStrs, s)
						}
					}
					aliases = strings.Join(aliasStrs, ", ")
				}

				// Truncate for display
				displayID := id
				displayTitle := title
				displayAliases := aliases
				if len(displayID) > 14 {
					displayID = displayID[:14] + ".."
				}
				if len(displayTitle) > 18 {
					displayTitle = displayTitle[:18] + ".."
				}
				if len(displayAliases) > 18 {
					displayAliases = displayAliases[:18] + ".."
				}

				table.Rows = append(table.Rows, []string{displayID, displayTitle, visibility, displayAliases})
				appsData = append(appsData, app)
			}
		}
	}

	data := map[string]interface{}{
		"peer":  peer.Name,
		"apps":  appsData,
		"count": len(table.Rows),
	}

	md := output.NewMarkdown().
		H1(fmt.Sprintf("Apps on %s", peer.Name)).
		Table(table).
		Para(fmt.Sprintf("%d apps", len(table.Rows))).
		String()

	renderer.Print(md, data)
}

// handleAppInfoV2 shows app info with v0.10 format
func handleAppInfoV2(args []string) {
	flags := flag.NewFlagSet("app info", flag.ExitOnError)
	aliasFlag := flags.String("alias", "", "Lookup by alias")
	idFlag := flags.String("id", "", "Lookup by app ID")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app info [--alias <alias> | --id <id>]")
		fmt.Println("       fazt @<peer> app info [--alias <alias> | --id <id>]")
		fmt.Println()
		flags.PrintDefaults()
	}

	// Find identifier (first non-flag arg)
	var identifier string
	var flagArgs []string
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") && identifier == "" {
			identifier = arg
			flagArgs = args[i+1:]
			break
		}
	}

	if len(flagArgs) == 0 {
		flagArgs = args
	}
	flags.Parse(flagArgs)

	// Determine identifier
	if *aliasFlag != "" {
		identifier = *aliasFlag
	} else if *idFlag != "" {
		identifier = *idFlag
	}

	if identifier == "" {
		fmt.Println("Error: app identifier required")
		flags.Usage()
		os.Exit(1)
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, targetPeerName)
	if err != nil {
		handlePeerError(err)
		os.Exit(1)
	}

	cmdArgs := []string{"info"}
	if *aliasFlag != "" {
		cmdArgs = append(cmdArgs, "--alias", identifier)
	} else if *idFlag != "" {
		cmdArgs = append(cmdArgs, "--id", identifier)
	} else {
		cmdArgs = append(cmdArgs, identifier)
	}

	result, err := executeRemoteCmd(peer, "app", cmdArgs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if app, ok := result.(map[string]interface{}); ok {
		fmt.Printf("ID:          %s\n", getString(app, "id"))
		fmt.Printf("Title:       %s\n", getString(app, "title"))
		if desc := getString(app, "description"); desc != "" {
			fmt.Printf("Description: %s\n", desc)
		}
		fmt.Printf("Visibility:  %s\n", getString(app, "visibility"))
		fmt.Printf("Source:      %s\n", getString(app, "source"))
		fmt.Printf("Files:       %v\n", app["file_count"])
		fmt.Printf("Size:        %s\n", formatSize(int64(getFloat(app, "size_bytes"))))

		if aliases, ok := app["aliases"].([]interface{}); ok && len(aliases) > 0 {
			var aliasStrs []string
			for _, a := range aliases {
				if s, ok := a.(string); ok {
					aliasStrs = append(aliasStrs, s)
				}
			}
			fmt.Printf("Aliases:     %s\n", strings.Join(aliasStrs, ", "))
		}

		if url := getString(app, "url"); url != "" {
			fmt.Printf("URL:         %s\n", url)
		}

		if origID := getString(app, "original_id"); origID != "" && origID != getString(app, "id") {
			fmt.Printf("Original:    %s\n", origID)
		}
		if forkedFrom := getString(app, "forked_from_id"); forkedFrom != "" {
			fmt.Printf("Forked from: %s\n", forkedFrom)
		}
	}
}

// handleAppFiles lists files in a deployed app
func handleAppFiles(args []string) {
	flags := flag.NewFlagSet("app files", flag.ExitOnError)
	aliasFlag := flags.String("alias", "", "Lookup by alias")
	idFlag := flags.String("id", "", "Lookup by app ID")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app files <app> [--alias | --id]")
		fmt.Println("       fazt @<peer> app files <app> [--alias | --id]")
		fmt.Println()
		fmt.Println("Lists all files in a deployed app.")
		fmt.Println()
		flags.PrintDefaults()
	}

	// Find identifier (first non-flag arg)
	var identifier string
	var flagArgs []string
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") && identifier == "" {
			identifier = arg
			flagArgs = args[i+1:]
			break
		}
	}

	if len(flagArgs) == 0 {
		flagArgs = args
	}
	flags.Parse(flagArgs)

	// Determine identifier
	if *aliasFlag != "" {
		identifier = *aliasFlag
	} else if *idFlag != "" {
		identifier = *idFlag
	}

	if identifier == "" {
		fmt.Println("Error: app identifier required")
		flags.Usage()
		os.Exit(1)
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, targetPeerName)
	if err != nil {
		handlePeerError(err)
		os.Exit(1)
	}

	// Use remote client to get files
	client := remote.NewClient(peer)
	files, err := client.GetAppFiles(identifier)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Build output using output system
	renderer := getRenderer()

	table := &output.Table{
		Headers: []string{"Path", "Size", "Modified"},
		Rows:    make([][]string, len(files)),
	}

	for i, file := range files {
		table.Rows[i] = []string{
			file.Path,
			formatSize(file.Size),
			file.ModTime,
		}
	}

	data := map[string]interface{}{
		"peer":  peer.Name,
		"app":   identifier,
		"files": files,
		"count": len(files),
	}

	md := output.NewMarkdown().
		H1(fmt.Sprintf("Files in %s", identifier)).
		Table(table).
		Para(fmt.Sprintf("%d files", len(files))).
		String()

	renderer.Print(md, data)
}

// handleAppRemoveV2 removes an app with v0.10 options
func handleAppRemoveV2(args []string) {
	flags := flag.NewFlagSet("app remove", flag.ExitOnError)
	aliasFlag := flags.String("alias", "", "Remove alias only")
	idFlag := flags.String("id", "", "Remove app by ID")
	withForks := flags.Bool("with-forks", false, "Also delete all forks")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app remove [--alias <alias> | --id <id>] [--with-forks]")
		fmt.Println("       fazt @<peer> app remove [--alias <alias> | --id <id>] [--with-forks]")
		fmt.Println()
		flags.PrintDefaults()
	}

	// Find identifier
	var identifier string
	var flagArgs []string
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") && identifier == "" {
			identifier = arg
			flagArgs = args[i+1:]
			break
		}
	}

	if len(flagArgs) == 0 {
		flagArgs = args
	}
	flags.Parse(flagArgs)

	if *aliasFlag != "" {
		identifier = *aliasFlag
	} else if *idFlag != "" {
		identifier = *idFlag
	}

	if identifier == "" {
		fmt.Println("Error: app identifier required")
		flags.Usage()
		os.Exit(1)
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, targetPeerName)
	if err != nil {
		handlePeerError(err)
		os.Exit(1)
	}

	cmdArgs := []string{"remove"}
	if *aliasFlag != "" {
		cmdArgs = append(cmdArgs, "--alias", identifier)
	} else if *idFlag != "" {
		cmdArgs = append(cmdArgs, "--id", identifier)
	} else {
		cmdArgs = append(cmdArgs, identifier)
	}
	if *withForks {
		cmdArgs = append(cmdArgs, "--with-forks")
	}

	result, err := executeRemoteCmd(peer, "app", cmdArgs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp, ok := result.(map[string]interface{}); ok {
		if msg := getString(resp, "message"); msg != "" {
			fmt.Println(msg)
		}
		if deleted := getFloat(resp, "deleted"); deleted > 1 {
			fmt.Printf("Deleted %d apps (including forks)\n", int(deleted))
		}
	}
}

// handleAppLink creates or updates an alias
func handleAppLink(args []string) {
	flags := flag.NewFlagSet("app link", flag.ExitOnError)
	idFlag := flags.String("id", "", "App ID to link (required)")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app link <subdomain> --id <app_id>")
		fmt.Println("       fazt @<peer> app link <subdomain> --id <app_id>")
		fmt.Println()
		flags.PrintDefaults()
	}

	var subdomain string
	var flagArgs []string
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") && subdomain == "" {
			subdomain = arg
			flagArgs = args[i+1:]
			break
		}
	}

	if len(flagArgs) == 0 {
		flagArgs = args
	}
	flags.Parse(flagArgs)

	if subdomain == "" || *idFlag == "" {
		fmt.Println("Error: subdomain and --id are required")
		flags.Usage()
		os.Exit(1)
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, targetPeerName)
	if err != nil {
		handlePeerError(err)
		os.Exit(1)
	}

	result, err := executeRemoteCmd(peer, "app", []string{"link", subdomain, "--id", *idFlag})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp, ok := result.(map[string]interface{}); ok {
		fmt.Printf("Linked %s → %s\n", subdomain, *idFlag)
		if url := getString(resp, "url"); url != "" {
			fmt.Printf("URL: %s\n", url)
		}
	}
}

// handleAppUnlink removes an alias
func handleAppUnlink(args []string) {
	flags := flag.NewFlagSet("app unlink", flag.ExitOnError)

	flags.Usage = func() {
		fmt.Println("Usage: fazt app unlink <subdomain>")
		fmt.Println("       fazt @<peer> app unlink <subdomain>")
		fmt.Println()
		flags.PrintDefaults()
	}

	var subdomain string
	var flagArgs []string
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") && subdomain == "" {
			subdomain = arg
			flagArgs = args[i+1:]
			break
		}
	}

	if len(flagArgs) == 0 {
		flagArgs = args
	}
	flags.Parse(flagArgs)

	if subdomain == "" {
		fmt.Println("Error: subdomain is required")
		flags.Usage()
		os.Exit(1)
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, targetPeerName)
	if err != nil {
		handlePeerError(err)
		os.Exit(1)
	}

	_, err = executeRemoteCmd(peer, "app", []string{"unlink", subdomain})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Unlinked %s\n", subdomain)
}

// handleAppReserve reserves a subdomain
func handleAppReserve(args []string) {
	flags := flag.NewFlagSet("app reserve", flag.ExitOnError)

	flags.Usage = func() {
		fmt.Println("Usage: fazt app reserve <subdomain>")
		fmt.Println("       fazt @<peer> app reserve <subdomain>")
		fmt.Println()
		flags.PrintDefaults()
	}

	var subdomain string
	var flagArgs []string
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") && subdomain == "" {
			subdomain = arg
			flagArgs = args[i+1:]
			break
		}
	}

	if len(flagArgs) == 0 {
		flagArgs = args
	}
	flags.Parse(flagArgs)

	if subdomain == "" {
		fmt.Println("Error: subdomain is required")
		flags.Usage()
		os.Exit(1)
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, targetPeerName)
	if err != nil {
		handlePeerError(err)
		os.Exit(1)
	}

	_, err = executeRemoteCmd(peer, "app", []string{"reserve", subdomain})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Reserved %s\n", subdomain)
}

// handleAppFork forks an app
func handleAppFork(args []string) {
	flags := flag.NewFlagSet("app fork", flag.ExitOnError)
	aliasFlag := flags.String("alias", "", "Source alias")
	idFlag := flags.String("id", "", "Source app ID")
	asFlag := flags.String("as", "", "New alias for fork")
	noStorage := flags.Bool("no-storage", false, "Don't copy storage")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app fork [--alias <alias> | --id <id>] [--as <new-alias>] [--no-storage]")
		fmt.Println("       fazt @<peer> app fork [--alias <alias> | --id <id>] [--as <new-alias>] [--no-storage]")
		fmt.Println()
		flags.PrintDefaults()
	}

	flags.Parse(args)

	identifier := *aliasFlag
	if *idFlag != "" {
		identifier = *idFlag
	}

	if identifier == "" {
		fmt.Println("Error: --alias or --id is required")
		flags.Usage()
		os.Exit(1)
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, targetPeerName)
	if err != nil {
		handlePeerError(err)
		os.Exit(1)
	}

	cmdArgs := []string{"fork"}
	if *aliasFlag != "" {
		cmdArgs = append(cmdArgs, "--alias", *aliasFlag)
	} else {
		cmdArgs = append(cmdArgs, "--id", *idFlag)
	}
	if *asFlag != "" {
		cmdArgs = append(cmdArgs, "--as", *asFlag)
	}
	if *noStorage {
		cmdArgs = append(cmdArgs, "--no-storage")
	}

	result, err := executeRemoteCmd(peer, "app", cmdArgs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if resp, ok := result.(map[string]interface{}); ok {
		fmt.Printf("Forked: %s\n", getString(resp, "id"))
		fmt.Printf("Title:  %s\n", getString(resp, "title"))
		fmt.Printf("From:   %s\n", getString(resp, "forked_from_id"))
		if alias := getString(resp, "alias"); alias != "" {
			fmt.Printf("Alias:  %s\n", alias)
		}
		if url := getString(resp, "url"); url != "" {
			fmt.Printf("URL:    %s\n", url)
		}
	}
}

// handleAppSwap swaps two aliases
func handleAppSwap(args []string) {
	flags := flag.NewFlagSet("app swap", flag.ExitOnError)

	flags.Usage = func() {
		fmt.Println("Usage: fazt app swap <alias1> <alias2>")
		fmt.Println("       fazt @<peer> app swap <alias1> <alias2>")
		fmt.Println()
		flags.PrintDefaults()
	}

	var aliases []string
	var flagArgs []string
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flagArgs = args[i:]
			break
		}
		aliases = append(aliases, arg)
	}

	flags.Parse(flagArgs)

	if len(aliases) < 2 {
		fmt.Println("Error: two aliases required")
		flags.Usage()
		os.Exit(1)
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, targetPeerName)
	if err != nil {
		handlePeerError(err)
		os.Exit(1)
	}

	// Use direct API call for swap
	httpClient := &http.Client{}
	body := map[string]string{"alias1": aliases[0], "alias2": aliases[1]}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", peer.URL+"/api/aliases/swap", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+peer.Token)

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", string(bodyBytes))
		os.Exit(1)
	}

	fmt.Printf("Swapped %s ↔ %s\n", aliases[0], aliases[1])
}

// handleAppSplit configures traffic splitting
func handleAppSplit(args []string) {
	flags := flag.NewFlagSet("app split", flag.ExitOnError)
	idsFlag := flags.String("ids", "", "Comma-separated app_id:weight pairs (e.g., app_abc:50,app_def:50)")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app split <subdomain> --ids <id1:weight1,id2:weight2>")
		fmt.Println("       fazt @<peer> app split <subdomain> --ids <id1:weight1,id2:weight2>")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  fazt @zyt app split tetris --ids app_v1:50,app_v2:50")
		fmt.Println()
		flags.PrintDefaults()
	}

	var subdomain string
	var flagArgs []string
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") && subdomain == "" {
			subdomain = arg
			flagArgs = args[i+1:]
			break
		}
	}

	if len(flagArgs) == 0 {
		flagArgs = args
	}
	flags.Parse(flagArgs)

	if subdomain == "" || *idsFlag == "" {
		fmt.Println("Error: subdomain and --ids are required")
		flags.Usage()
		os.Exit(1)
	}

	// Parse ids
	var targets []map[string]interface{}
	pairs := strings.Split(*idsFlag, ",")
	for _, pair := range pairs {
		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			fmt.Printf("Error: invalid format '%s', expected 'app_id:weight'\n", pair)
			os.Exit(1)
		}
		var weight int
		fmt.Sscanf(parts[1], "%d", &weight)
		targets = append(targets, map[string]interface{}{
			"app_id": parts[0],
			"weight": weight,
		})
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, targetPeerName)
	if err != nil {
		handlePeerError(err)
		os.Exit(1)
	}

	// Use direct API call for split
	httpClient := &http.Client{}
	body := map[string]interface{}{"targets": targets}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", peer.URL+"/api/aliases/"+subdomain+"/split", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+peer.Token)

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", string(bodyBytes))
		os.Exit(1)
	}

	fmt.Printf("Configured traffic split for %s\n", subdomain)
	for _, t := range targets {
		fmt.Printf("  %s: %d%%\n", t["app_id"], t["weight"])
	}
}

// handleAppLineage shows the lineage tree for an app
func handleAppLineage(args []string) {
	flags := flag.NewFlagSet("app lineage", flag.ExitOnError)
	aliasFlag := flags.String("alias", "", "Lookup by alias")
	idFlag := flags.String("id", "", "Lookup by app ID")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app lineage [--alias <alias> | --id <id>]")
		fmt.Println("       fazt @<peer> app lineage [--alias <alias> | --id <id>]")
		fmt.Println()
		flags.PrintDefaults()
	}

	flags.Parse(args)

	identifier := *aliasFlag
	if *idFlag != "" {
		identifier = *idFlag
	}

	if identifier == "" {
		fmt.Println("Error: --alias or --id is required")
		flags.Usage()
		os.Exit(1)
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, targetPeerName)
	if err != nil {
		handlePeerError(err)
		os.Exit(1)
	}

	cmdArgs := []string{"lineage"}
	if *aliasFlag != "" {
		cmdArgs = append(cmdArgs, "--alias", *aliasFlag)
	} else {
		cmdArgs = append(cmdArgs, "--id", *idFlag)
	}

	result, err := executeRemoteCmd(peer, "app", cmdArgs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print tree
	printLineageTree(result, "", true)
}

func printLineageTree(node interface{}, prefix string, isLast bool) {
	if node == nil {
		return
	}

	nodeMap, ok := node.(map[string]interface{})
	if !ok {
		return
	}

	// Print connector
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if prefix == "" {
		connector = ""
	}

	id := getString(nodeMap, "id")
	title := getString(nodeMap, "title")
	aliases := ""
	if aliasArr, ok := nodeMap["aliases"].([]interface{}); ok && len(aliasArr) > 0 {
		var aliasStrs []string
		for _, a := range aliasArr {
			if s, ok := a.(string); ok {
				aliasStrs = append(aliasStrs, s)
			}
		}
		aliases = " [" + strings.Join(aliasStrs, ", ") + "]"
	}

	fmt.Printf("%s%s%s \"%s\"%s\n", prefix, connector, id, title, aliases)

	// Print forks
	if forks, ok := nodeMap["forks"].([]interface{}); ok {
		newPrefix := prefix
		if prefix != "" {
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
		}

		for i, fork := range forks {
			printLineageTree(fork, newPrefix, i == len(forks)-1)
		}
	}
}

// executeRemoteCmd executes a command via the command gateway
func executeRemoteCmd(peer *remote.Peer, command string, args []string) (interface{}, error) {
	client := &http.Client{}

	body := map[string]interface{}{
		"command": command,
		"args":    args,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", peer.URL+"/api/cmd", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+peer.Token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Success bool        `json:"success"`
			Data    interface{} `json:"data"`
			Error   string      `json:"error"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Data.Success {
		return nil, fmt.Errorf("%s", result.Data.Error)
	}

	return result.Data.Data, nil
}

func handlePeerError(err error) {
	if err == remote.ErrNoPeers {
		fmt.Println("No peers configured.")
		fmt.Println("Run: fazt remote add <name> --url <url> --token <token>")
	} else if err == remote.ErrNoDefaultPeer {
		fmt.Println("Multiple peers configured. Specify which peer:")
		fmt.Println("  fazt app list <peer>")
	} else {
		fmt.Printf("Error: %v\n", err)
	}
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

func printAppHelpV2() {
	// Try markdown-based help first
	if help.Exists("app") {
		doc, _ := help.Load("app")
		fmt.Print(help.Render(doc))
		return
	}

	// LEGACY_CODE: migrate to cli/app/_index.md
	fmt.Println(`Fazt.sh - App Management (v0.10)

USAGE:
  fazt app <command> [options]
  fazt @<peer> app <command> [options]  (remote execution)

REMOTE COMMANDS (support @peer):
  list [peer]           List apps (--aliases for alias list)
  info [identifier]     Show app details (--alias or --id)
  files <app>           List files in a deployed app (--alias or --id)
  deploy <dir>          Deploy directory to peer
  logs <app>            View serverless execution logs (-f to follow)
  install <url>         Install app from git repository
  remove [identifier]   Remove app (--alias, --id, --with-forks)
  upgrade <app>         Upgrade git-sourced app
  link <subdomain>      Link subdomain to app (--id required)
  unlink <subdomain>    Remove alias
  reserve <subdomain>   Reserve/block subdomain
  swap <a1> <a2>        Atomically swap two aliases
  split <subdomain>     Configure traffic splitting (--ids)
  fork                  Fork an app (--alias/--id, --as, --no-storage)
  lineage               Show fork tree (--alias/--id)

LOCAL COMMANDS (no @peer support):
  create <name>         Create local app from template (static, vue, vue-api)
  validate <dir>        Validate local directory before deployment

OPTIONS:
  --alias <name>        Reference app by alias
  --id <app_id>         Reference app by ID
  --with-forks          Delete app and all its forks

GLOBAL FLAGS:
  --verbose             Show detailed output (migrations, debug info)
  --format <fmt>        Output format: markdown (default) or json

PEER SELECTION:
  Use @<peer> prefix for remote operations:
    fazt @zyt app list              # List apps on zyt peer
    fazt @local app deploy ./myapp  # Deploy to local peer

EXAMPLES:
  fazt @zyt app list
  fazt @zyt app deploy ./my-site
  fazt @zyt app info --alias tetris`)
}
