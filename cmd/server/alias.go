package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/output"
	"github.com/fazt-sh/fazt/internal/remote"
)

// handleAliasCommand handles local alias management
func handleAliasCommand(args []string) {
	if len(args) < 1 {
		handleAliasList(nil)
		return
	}

	subCmd := args[0]
	switch subCmd {
	case "list":
		handleAliasList(args[1:])
	case "info":
		handleAliasInfo(args[1:])
	case "--help", "-h", "help":
		printAliasUsage()
	default:
		fmt.Printf("Unknown alias subcommand: %s\n", subCmd)
		printAliasUsage()
		os.Exit(1)
	}
}

func printAliasUsage() {
	fmt.Println("fazt alias - Alias management commands")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  fazt alias <command> [options]")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  list                    List all aliases")
	fmt.Println("  info                    Show alias details (requires --name)")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  --name <subdomain>      Alias subdomain name")
	fmt.Println("  --offset <n>            Skip first n results (default: 0)")
	fmt.Println("  --limit <n>             Max results to return (default: 20)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  fazt alias list")
	fmt.Println("  fazt alias list --limit 50")
	fmt.Println("  fazt alias info --name myapp")
	fmt.Println("  fazt @zyt alias list")
}

func handleAliasList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	offsetFlag := fs.Int("offset", 0, "Skip first n results")
	limitFlag := fs.Int("limit", 20, "Max results to return")
	fs.Parse(args)

	// Remote peer support
	if targetPeerName != "" {
		handleAliasListRemote(targetPeerName, *offsetFlag, *limitFlag)
		return
	}

	db := getClientDB()
	defer database.Close()

	// Get total count
	var total int
	db.QueryRow(`SELECT COUNT(*) FROM aliases`).Scan(&total)

	// Query with pagination
	rows, err := db.Query(`
		SELECT subdomain, type, targets, datetime(created_at), datetime(updated_at)
		FROM aliases
		ORDER BY subdomain
		LIMIT ? OFFSET ?
	`, *limitFlag, *offsetFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	renderer := getRenderer()
	var tableRows [][]string
	var aliases []map[string]interface{}

	for rows.Next() {
		var subdomain, aliasType, targets, createdAt, updatedAt string
		var targetsPtr *string
		if err := rows.Scan(&subdomain, &aliasType, &targetsPtr, &createdAt, &updatedAt); err != nil {
			continue
		}
		if targetsPtr != nil {
			targets = *targetsPtr
		}

		// Parse target for display
		targetDisplay := ""
		if aliasType == "app" && targets != "" {
			var t struct {
				AppID string `json:"app_id"`
			}
			json.Unmarshal([]byte(targets), &t)
			targetDisplay = t.AppID
		} else if aliasType == "redirect" && targets != "" {
			var t struct {
				URL string `json:"url"`
			}
			json.Unmarshal([]byte(targets), &t)
			targetDisplay = t.URL
		} else if aliasType == "reserved" {
			targetDisplay = "(reserved)"
		}

		tableRows = append(tableRows, []string{subdomain, aliasType, targetDisplay, output.TimeAgoString(updatedAt)})
		aliases = append(aliases, map[string]interface{}{
			"subdomain":  subdomain,
			"type":       aliasType,
			"targets":    targets,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}

	table := &output.Table{
		Headers: []string{"Subdomain", "Type", "Target", "Updated"},
		Rows:    tableRows,
	}

	md := output.NewMarkdown().
		H1("Aliases").
		Table(table)

	// Show pagination info if not showing all
	if total > *limitFlag || *offsetFlag > 0 {
		md.Para(fmt.Sprintf("Showing %d-%d of %d (--offset %d --limit %d)",
			*offsetFlag+1, min(*offsetFlag+len(aliases), total), total, *offsetFlag, *limitFlag))
	}

	renderer.Print(md.String(), map[string]interface{}{
		"aliases": aliases,
		"pagination": map[string]interface{}{
			"offset":   *offsetFlag,
			"limit":    *limitFlag,
			"total":    total,
			"has_more": *offsetFlag+*limitFlag < total,
		},
	})
}

func handleAliasInfo(args []string) {
	fs := flag.NewFlagSet("info", flag.ExitOnError)
	nameFlag := fs.String("name", "", "Alias subdomain name")
	fs.Parse(args)

	if *nameFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: --name flag required")
		fmt.Fprintln(os.Stderr, "Usage: fazt alias info --name <SUBDOMAIN>")
		os.Exit(1)
	}

	// Remote peer support
	if targetPeerName != "" {
		handleAliasInfoRemote(targetPeerName, *nameFlag)
		return
	}

	db := getClientDB()
	defer database.Close()

	var subdomain, aliasType, targets, createdAt, updatedAt string
	var targetsPtr *string
	err := db.QueryRow(`
		SELECT subdomain, type, targets, datetime(created_at), datetime(updated_at)
		FROM aliases WHERE subdomain = ?
	`, *nameFlag).Scan(&subdomain, &aliasType, &targetsPtr, &createdAt, &updatedAt)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Alias not found: %s\n", *nameFlag)
		os.Exit(1)
	}
	if targetsPtr != nil {
		targets = *targetsPtr
	}

	renderer := getRenderer()

	// Format output
	md := output.NewMarkdown().
		H1("Alias Info")

	infoTable := &output.Table{
		Headers: []string{"Field", "Value"},
		Rows: [][]string{
			{"Subdomain", subdomain},
			{"Type", aliasType},
			{"Created", output.TimeAgoString(createdAt)},
			{"Updated", output.TimeAgoString(updatedAt)},
		},
	}
	md.Table(infoTable)

	// Show target details
	if targets != "" {
		md.H2("Target")
		if aliasType == "app" {
			var t struct {
				AppID string `json:"app_id"`
			}
			json.Unmarshal([]byte(targets), &t)
			md.Para(fmt.Sprintf("App ID: %s", t.AppID))
		} else if aliasType == "redirect" {
			var t struct {
				URL string `json:"url"`
			}
			json.Unmarshal([]byte(targets), &t)
			md.Para(fmt.Sprintf("Redirect URL: %s", t.URL))
		} else if aliasType == "split" {
			var splits []struct {
				AppID  string `json:"app_id"`
				Weight int    `json:"weight"`
			}
			json.Unmarshal([]byte(targets), &splits)
			splitRows := make([][]string, len(splits))
			for i, s := range splits {
				splitRows[i] = []string{s.AppID, fmt.Sprintf("%d%%", s.Weight)}
			}
			splitTable := &output.Table{
				Headers: []string{"App ID", "Weight"},
				Rows:    splitRows,
			}
			md.Table(splitTable)
		}
	}

	jsonData := map[string]interface{}{
		"subdomain":  subdomain,
		"type":       aliasType,
		"targets":    targets,
		"created_at": createdAt,
		"updated_at": updatedAt,
	}

	renderer.Print(md.String(), jsonData)
}

// Remote handlers

func handleAliasCommandWithPeer(peerName string, args []string) {
	if len(args) < 1 {
		handleAliasListRemote(peerName, 0, 20)
		return
	}

	subCmd := args[0]
	switch subCmd {
	case "list":
		fs := flag.NewFlagSet("list", flag.ExitOnError)
		offsetFlag := fs.Int("offset", 0, "Skip first n results")
		limitFlag := fs.Int("limit", 20, "Max results to return")
		fs.Parse(args[1:])
		handleAliasListRemote(peerName, *offsetFlag, *limitFlag)
	case "info":
		fs := flag.NewFlagSet("info", flag.ExitOnError)
		nameFlag := fs.String("name", "", "Alias subdomain name")
		fs.Parse(args[1:])
		if *nameFlag == "" {
			fmt.Fprintln(os.Stderr, "Error: --name flag required")
			os.Exit(1)
		}
		handleAliasInfoRemote(peerName, *nameFlag)
	default:
		fmt.Printf("Unknown alias subcommand: %s\n", subCmd)
		printAliasUsage()
		os.Exit(1)
	}
}

func handleAliasListRemote(peerName string, offset, limit int) {
	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, peerName)
	if err != nil {
		handlePeerError(err)
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/api/aliases?offset=%d&limit=%d", peer.URL, offset, limit)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+peer.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var response struct {
		Data []struct {
			Subdomain string          `json:"subdomain"`
			Type      string          `json:"type"`
			Targets   json.RawMessage `json:"targets"`
			CreatedAt string          `json:"created_at"`
			UpdatedAt string          `json:"updated_at"`
		} `json:"data"`
		Meta struct {
			Offset  int  `json:"offset"`
			Limit   int  `json:"limit"`
			Total   int  `json:"total"`
			HasMore bool `json:"has_more"`
		} `json:"meta"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		os.Exit(1)
	}

	if response.Error != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", response.Error.Message)
		os.Exit(1)
	}

	renderer := getRenderer()
	var tableRows [][]string

	for _, a := range response.Data {
		// Parse target for display
		targetDisplay := ""
		if a.Type == "app" {
			var t struct {
				AppID string `json:"app_id"`
			}
			json.Unmarshal(a.Targets, &t)
			targetDisplay = t.AppID
		} else if a.Type == "redirect" {
			var t struct {
				URL string `json:"url"`
			}
			json.Unmarshal(a.Targets, &t)
			targetDisplay = t.URL
		} else if a.Type == "reserved" {
			targetDisplay = "(reserved)"
		}

		tableRows = append(tableRows, []string{a.Subdomain, a.Type, targetDisplay, output.TimeAgoString(a.UpdatedAt)})
	}

	table := &output.Table{
		Headers: []string{"Subdomain", "Type", "Target", "Updated"},
		Rows:    tableRows,
	}

	md := output.NewMarkdown().
		H1(fmt.Sprintf("Aliases (@%s)", peerName)).
		Table(table)

	// Show pagination info
	meta := response.Meta
	if meta.Total > meta.Limit || meta.Offset > 0 {
		md.Para(fmt.Sprintf("Showing %d-%d of %d (--offset %d --limit %d)",
			meta.Offset+1, min(meta.Offset+len(response.Data), meta.Total), meta.Total, meta.Offset, meta.Limit))
	}

	renderer.Print(md.String(), response)
}

func handleAliasInfoRemote(peerName, name string) {
	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, peerName)
	if err != nil {
		handlePeerError(err)
		os.Exit(1)
	}

	req, _ := http.NewRequest("GET", peer.URL+"/api/aliases/"+name, nil)
	req.Header.Set("Authorization", "Bearer "+peer.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var response struct {
		Data struct {
			Subdomain string          `json:"subdomain"`
			Type      string          `json:"type"`
			Targets   json.RawMessage `json:"targets"`
			CreatedAt string          `json:"created_at"`
			UpdatedAt string          `json:"updated_at"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		os.Exit(1)
	}

	if response.Error != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", response.Error.Message)
		os.Exit(1)
	}

	a := response.Data
	renderer := getRenderer()

	md := output.NewMarkdown().
		H1(fmt.Sprintf("Alias Info (@%s)", peerName))

	infoTable := &output.Table{
		Headers: []string{"Field", "Value"},
		Rows: [][]string{
			{"Subdomain", a.Subdomain},
			{"Type", a.Type},
			{"Created", output.TimeAgoString(a.CreatedAt)},
			{"Updated", output.TimeAgoString(a.UpdatedAt)},
		},
	}
	md.Table(infoTable)

	// Show target details
	if len(a.Targets) > 0 {
		md.H2("Target")
		if a.Type == "app" {
			var t struct {
				AppID string `json:"app_id"`
			}
			json.Unmarshal(a.Targets, &t)
			md.Para(fmt.Sprintf("App ID: %s", t.AppID))
		} else if a.Type == "redirect" {
			var t struct {
				URL string `json:"url"`
			}
			json.Unmarshal(a.Targets, &t)
			md.Para(fmt.Sprintf("Redirect URL: %s", t.URL))
		}
	}

	renderer.Print(md.String(), response.Data)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
