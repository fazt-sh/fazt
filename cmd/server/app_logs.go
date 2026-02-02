package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/remote"
)

// handleAppLogs streams logs from a deployed app
func handleAppLogs(args []string) {
	flags := flag.NewFlagSet("app logs", flag.ExitOnError)
	follow := flags.Bool("f", false, "Follow log output (stream)")
	limit := flags.Int("n", 50, "Number of recent logs to show")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app logs <app> [-f] [-n <count>]")
		fmt.Println("       fazt @<peer> app logs <app> [-f] [-n <count>]")
		fmt.Println()
		fmt.Println("View serverless execution logs for an app.")
		fmt.Println()
		flags.PrintDefaults()
	}

	// Find app name (first non-flag arg)
	var appName string
	var flagArgs []string
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") && appName == "" {
			appName = arg
			flagArgs = args[i+1:]
			break
		}
	}

	if appName == "" {
		fmt.Println("Error: app name is required")
		flags.Usage()
		os.Exit(1)
	}

	flags.Parse(flagArgs)

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, targetPeerName)
	if err != nil {
		if err == remote.ErrNoPeers {
			fmt.Println("No peers configured.")
			fmt.Println("Run: fazt remote add <name> --url <url> --token <token>")
		} else if err == remote.ErrNoDefaultPeer {
			fmt.Println("Multiple peers configured. Specify which peer:")
			fmt.Println("  fazt @<peer> app logs <app>")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		os.Exit(1)
	}

	if *follow {
		streamLogs(peer, appName)
	} else {
		fetchLogs(peer, appName, *limit)
	}
}

// fetchLogs fetches recent logs from the server
func fetchLogs(peer *remote.Peer, appName string, limit int) {
	url := fmt.Sprintf("%s/api/logs?site_id=%s&limit=%d", peer.URL, appName, limit)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+peer.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", resp.Status)
		os.Exit(1)
	}

	var result struct {
		Data struct {
			Logs []LogEntry `json:"logs"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if len(result.Data.Logs) == 0 {
		fmt.Printf("No logs found for %s\n", appName)
		return
	}

	// Print logs in reverse order (oldest first)
	for i := len(result.Data.Logs) - 1; i >= 0; i-- {
		log := result.Data.Logs[i]
		printLog(log)
	}
}

// streamLogs connects to SSE endpoint and streams logs
func streamLogs(peer *remote.Peer, appName string) {
	url := fmt.Sprintf("%s/api/logs/stream?site_id=%s", peer.URL, appName)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+peer.Token)
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: 0} // No timeout for streaming
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error connecting: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", resp.Status)
		os.Exit(1)
	}

	fmt.Printf("Streaming logs for %s (Ctrl+C to stop)...\n\n", appName)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse SSE format
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var log LogEntry
			if err := json.Unmarshal([]byte(data), &log); err == nil {
				printLog(log)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Stream error: %v\n", err)
	}
}

// LogEntry represents a log entry from the server
type LogEntry struct {
	ID        int64  `json:"id"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// printLog prints a log entry with formatting
func printLog(log LogEntry) {
	// Parse timestamp
	timestamp := log.CreatedAt
	if t, err := time.Parse(time.RFC3339, log.CreatedAt); err == nil {
		timestamp = t.Format("15:04:05")
	} else if t, err := time.Parse("2006-01-02T15:04:05Z", log.CreatedAt); err == nil {
		timestamp = t.Format("15:04:05")
	} else if len(log.CreatedAt) > 19 {
		timestamp = log.CreatedAt[11:19]
	}

	// Color by level
	levelStr := log.Level
	switch log.Level {
	case "error":
		levelStr = "\033[31mERROR\033[0m" // Red
	case "warn":
		levelStr = "\033[33mWARN\033[0m" // Yellow
	case "info":
		levelStr = "\033[34mINFO\033[0m" // Blue
	case "debug":
		levelStr = "\033[90mDEBUG\033[0m" // Gray
	}

	fmt.Printf("%s [%s] %s\n", timestamp, levelStr, log.Message)
}
