package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fazt-sh/fazt/internal/analytics"
	"github.com/fazt-sh/fazt/internal/audit"
	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/clientconfig"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers"
	"github.com/fazt-sh/fazt/internal/hosting"
	"github.com/fazt-sh/fazt/internal/mcp"
	"github.com/fazt-sh/fazt/internal/middleware"
	"github.com/fazt-sh/fazt/internal/provision"
	"github.com/fazt-sh/fazt/internal/remote"
	jsruntime "github.com/fazt-sh/fazt/internal/runtime"
	"github.com/fazt-sh/fazt/internal/security"
	"github.com/fazt-sh/fazt/internal/storage"
	"github.com/fazt-sh/fazt/internal/term"
	ignore "github.com/sabhiram/go-gitignore"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
	"github.com/caddyserver/certmagic"
)

var (
	showVersion = flag.Bool("version", false, "Show version and exit")
	showHelp    = flag.Bool("help", false, "Show help and exit")
	verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	quiet       = flag.Bool("quiet", false, "Quiet mode (errors only)")
)

// serverlessHandler is the global serverless handler with storage support
var serverlessHandler *jsruntime.ServerlessHandler

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	// Handle help/version flags first
	if command == "--version" || command == "-version" || command == "-v" || command == "version" {
		printVersion()
		return
	}
	if command == "--help" || command == "-help" || command == "-h" {
		printUsage()
		return
	}

	// v0.10: Handle @peer prefix for remote execution
	// e.g., "fazt @zyt app list" -> execute "app list" on peer "zyt"
	if strings.HasPrefix(command, "@") {
		handlePeerCommand(command[1:], os.Args[2:])
		return
	}

	// Handle top-level subcommands
	switch command {
	case "server":
		handleServerCommand(os.Args[2:])
	case "servers":
		handleServersCommand(os.Args[2:])
	case "remote":
		handleRemoteCommand(os.Args[2:])
	case "app":
		handleAppCommandV2(os.Args[2:]) // v0.10: Use new app command handler
	case "service":
		handleServiceCommand(os.Args[2:])
	case "client":
		handleClientCommand(os.Args[2:])
	case "deploy":
		handleDeployCommand() // Alias for client deploy
	case "upgrade":
		handleUpgradeCommand()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// handlePeerCommand handles @peer remote execution
// e.g., "fazt @zyt app list" executes "app list" on peer "zyt"
func handlePeerCommand(peerName string, args []string) {
	if len(args) < 1 {
		fmt.Printf("Error: command required after @%s\n", peerName)
		fmt.Println("Usage: fazt @<peer> <command> [args...]")
		os.Exit(1)
	}

	command := args[0]
	cmdArgs := args[1:]

	// Only certain commands can be executed remotely
	switch command {
	case "app":
		// Inject peer into args for app commands
		// Most app commands already support --to/--from/--on flags
		handleAppCommandV2WithPeer(peerName, cmdArgs)
	case "server":
		// Limited server commands (info only)
		if len(cmdArgs) > 0 && cmdArgs[0] == "info" {
			handleRemoteServerInfo(peerName)
		} else {
			fmt.Printf("Error: only 'server info' can be executed remotely\n")
			os.Exit(1)
		}
	default:
		fmt.Printf("Error: command '%s' cannot be executed remotely\n", command)
		fmt.Println("Remote execution supported for: app, server info")
		os.Exit(1)
	}
}

// handleAppCommandV2WithPeer handles app commands with an explicit peer
func handleAppCommandV2WithPeer(peerName string, args []string) {
	if len(args) < 1 {
		// Default: list apps on specified peer
		handleAppListV2([]string{peerName})
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	// Inject peer name into the args based on the subcommand
	switch subcommand {
	case "list":
		handleAppListV2(append([]string{peerName}, subArgs...))
	case "info":
		handleAppInfoV2(append(subArgs, "--on", peerName))
	case "deploy":
		handleAppDeploy(append(subArgs, "--to", peerName))
	case "install":
		handleAppInstall(append(subArgs, "--to", peerName))
	case "remove":
		handleAppRemoveV2(append(subArgs, "--from", peerName))
	case "link":
		handleAppLink(append(subArgs, "--to", peerName))
	case "unlink":
		handleAppUnlink(append(subArgs, "--from", peerName))
	case "reserve":
		handleAppReserve(append(subArgs, "--on", peerName))
	case "fork":
		handleAppFork(append(subArgs, "--to", peerName))
	case "swap":
		handleAppSwap(append(subArgs, "--on", peerName))
	case "split":
		handleAppSplit(append(subArgs, "--on", peerName))
	case "lineage":
		handleAppLineage(append(subArgs, "--on", peerName))
	case "upgrade":
		handleAppUpgrade(append(subArgs, "--from", peerName))
	case "pull":
		handleAppPull(append(subArgs, "--from", peerName))
	default:
		fmt.Printf("Unknown app command: %s\n", subcommand)
		os.Exit(1)
	}
}

// handleRemoteServerInfo gets server info from a remote peer
func handleRemoteServerInfo(peerName string) {
	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, peerName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	result, err := executeRemoteCmd(peer, "server", []string{"info"})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if info, ok := result.(map[string]interface{}); ok {
		fmt.Printf("Server: %s\n", peerName)
		fmt.Printf("Version: %v\n", info["version"])
		fmt.Printf("Domain:  %v\n", info["domain"])
		fmt.Printf("Env:     %v\n", info["env"])
	}
}

// ===================================================================================
// CLI Command Functions (v0.4.0)
// ===================================================================================

// initCommand initializes server configuration for first-time setup
func initCommand(username, password, domain, port, env, dbPath string) error {
	// Check if DB already exists
	if _, err := os.Stat(dbPath); err == nil {
		return fmt.Errorf("Error: Server already initialized\nDatabase exists at: %s", dbPath)
	}

	// Validate required fields
	if username == "" || password == "" || domain == "" {
		return errors.New("Error: username, password, and domain are required")
	}

	// Validate port
	if port != "" {
		portNum, err := strconv.Atoi(port)
		if err != nil || portNum < 1 || portNum > 65535 {
			return fmt.Errorf("Error: invalid port '%s' (must be 1-65535)", port)
		}
	}

	// Validate environment
	if env != "development" && env != "production" {
		return fmt.Errorf("Error: invalid environment '%s' (must be 'development' or 'production')", env)
	}

	// Hash password with bcrypt cost 12
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("Error: failed to hash password: %v", err)
	}

	// Initialize DB
	if err := database.Init(dbPath); err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}
	defer database.Close()

	store := config.NewDBConfigStore(database.GetDB())

	// Wrap domain with wildcard DNS if it's an IP address
	wrappedDomain := config.WrapWithWildcardDNS(domain)

	// Set configs
	configs := map[string]string{
		"server.port":        port,
		"server.domain":      wrappedDomain,
		"server.env":         env,
		"auth.username":      username,
		"auth.password_hash": string(passwordHash),
		"ntfy.url":           "https://ntfy.sh",
	}

	for k, v := range configs {
		if err := store.Set(k, v); err != nil {
			return fmt.Errorf("failed to set config %s: %w", k, err)
		}
	}

	return nil
}

// setCredentialsCommand updates username and/or password in existing config
func setCredentialsCommand(username, password, dbPath string) error {
	// Validate at least one field is provided
	if username == "" && password == "" {
		return errors.New("Error: at least one of --username or --password is required")
	}

	// Initialize DB
	if err := database.Init(dbPath); err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}
	defer database.Close()
	
	store := config.NewDBConfigStore(database.GetDB())

	// Update provided fields
	if username != "" {
		if err := store.Set("auth.username", username); err != nil {
			return fmt.Errorf("failed to set username: %w", err)
		}
	}
	if password != "" {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			return fmt.Errorf("Error: Failed to hash password: %v", err)
		}
		if err := store.Set("auth.password_hash", string(passwordHash)); err != nil {
			return fmt.Errorf("failed to set password: %w", err)
		}
	}

	return nil
}

// setConfigCommand updates server configuration settings
func setConfigCommand(domain, port, env, dbPath string) error {
	// Validate at least one field is provided
	if domain == "" && port == "" && env == "" {
		return errors.New("Error: at least one of --domain, --port, or --env is required")
	}

	// Initialize DB
	if err := database.Init(dbPath); err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}
	defer database.Close()
	
	store := config.NewDBConfigStore(database.GetDB())

	// Validate and update port if provided
	if port != "" {
		portNum, err := strconv.Atoi(port)
		if err != nil || portNum < 1 || portNum > 65535 {
			return fmt.Errorf("Error: invalid port '%s' (must be 1-65535)", port)
		}
		if err := store.Set("server.port", port); err != nil {
			return fmt.Errorf("failed to set port: %w", err)
		}
	}

	// Validate and update environment if provided
	if env != "" {
		if env != "development" && env != "production" {
			return fmt.Errorf("Error: invalid environment '%s' (must be 'development' or 'production')", env)
		}
		if err := store.Set("server.env", env); err != nil {
			return fmt.Errorf("failed to set env: %w", err)
		}
	}

	// Update domain if provided (wrap with wildcard DNS if IP)
	if domain != "" {
		wrappedDomain := config.WrapWithWildcardDNS(domain)
		if err := store.Set("server.domain", wrappedDomain); err != nil {
			return fmt.Errorf("failed to set domain: %w", err)
		}
	}

	return nil
}

// statusCommand displays current configuration and server status
func statusCommand(dbPath string) (string, error) {
	// Initialize DB
	if err := database.Init(dbPath); err != nil {
		return "", fmt.Errorf("failed to init database at %s: %w", dbPath, err)
	}
	defer database.Close()

	// Manually use the Store to read values for display
	store := config.NewDBConfigStore(database.GetDB())
	dbMap, _ := store.Load()
	
	// Helper to get value or default
	get := func(key, def string) string {
		if v, ok := dbMap[key]; ok {
			return v
		}
		return def
	}

	var output strings.Builder
	output.WriteString("Server Status\n")
	output.WriteString("═══════════════════════════════════════════════════════════\n")
	output.WriteString(fmt.Sprintf("Database:     %s\n", dbPath))
	output.WriteString(fmt.Sprintf("Domain:       %s\n", get("server.domain", "https://fazt.sh")))
	output.WriteString(fmt.Sprintf("Port:         %s\n", get("server.port", "4698")))
	output.WriteString(fmt.Sprintf("Environment:  %s\n", get("server.env", "development")))
	output.WriteString(fmt.Sprintf("Username:     %s\n", get("auth.username", "(not set)")))

	// Check database size
	if stat, err := os.Stat(dbPath); err == nil {
		size := float64(stat.Size()) / (1024 * 1024) // Convert to MB
		output.WriteString(fmt.Sprintf("DB Size:      %.1f MB\n", size))
	}

	// Check VFS Site Count
	var siteCount int
	database.GetDB().QueryRow("SELECT COUNT(DISTINCT site_id) FROM files").Scan(&siteCount)
	output.WriteString(fmt.Sprintf("Sites (VFS):  %d\n", siteCount))

	// Check PID file for server status
	pidFile := filepath.Join(filepath.Dir(dbPath), "cc-server.pid")
	if pidData, err := os.ReadFile(pidFile); err == nil {
		pidStr := strings.TrimSpace(string(pidData))
		output.WriteString(fmt.Sprintf("\nServer:       ● Running (PID: %s)\n", pidStr))
	} else {
		output.WriteString("\nServer:       ○ Not running\n")
	}

	return output.String(), nil
}

// handleServerCommand handles server-related subcommands
func handleServerCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: server command requires a subcommand")
		printServerHelp()
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "init":
		handleInitCommand()
	case "set-credentials":
		handleSetCredentials()
	case "set-config":
		handleSetConfigCommand()
	case "status":
		handleStatusCommand()
	case "start":
		handleStartCommand()
	case "reset-admin":
		handleResetAdminCommand()
	case "create-key":
		handleCreateKeyCommand()
	case "--help", "-h", "help":
		printServerHelp()
	default:
		fmt.Printf("Unknown server command: %s\n\n", subcommand)
		printServerHelp()
		os.Exit(1)
	}
}

// handleServersCommand handles multi-server configuration commands
func handleServersCommand(args []string) {
	if len(args) < 1 {
		handleServersList()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "add":
		handleServersAdd(args[1:])
	case "list":
		handleServersList()
	case "default":
		handleServersDefault(args[1:])
	case "remove":
		handleServersRemove(args[1:])
	case "ping":
		handleServersPing(args[1:])
	case "--help", "-h", "help":
		printServersHelp()
	default:
		fmt.Printf("Unknown servers command: %s\n\n", subcommand)
		printServersHelp()
		os.Exit(1)
	}
}

func handleServersAdd(args []string) {
	// Support: fazt servers add <name> --url <url> --token <token>
	// The name comes first, then flags
	if len(args) < 1 {
		fmt.Println("Error: server name is required")
		fmt.Println("Usage: fazt servers add <name> --url <url> --token <token>")
		os.Exit(1)
	}

	name := args[0]
	flagArgs := args[1:]

	flags := flag.NewFlagSet("servers add", flag.ExitOnError)
	urlFlag := flags.String("url", "", "Server URL (required)")
	tokenFlag := flags.String("token", "", "API token (required)")
	flags.Parse(flagArgs)

	if *urlFlag == "" || *tokenFlag == "" {
		fmt.Println("Error: --url and --token are required")
		fmt.Println("Usage: fazt servers add <name> --url <url> --token <token>")
		os.Exit(1)
	}

	cfg, err := clientconfig.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.AddServer(name, *urlFlag, *tokenFlag); err != nil {
		fmt.Printf("Error adding server: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server '%s' added successfully!\n", name)
	if cfg.DefaultServer == name {
		fmt.Printf("  (set as default)\n")
	}
}

func handleServersList() {
	cfg, err := clientconfig.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	servers := cfg.ListServers()
	if len(servers) == 0 {
		fmt.Println("No servers configured.")
		fmt.Println("Run: fazt servers add <name> --url <url> --token <token>")
		return
	}

	fmt.Println("NAME\t\tURL\t\t\t\tDEFAULT")
	fmt.Println("----\t\t---\t\t\t\t-------")
	for _, srv := range servers {
		defaultMarker := ""
		if srv.IsDefault {
			defaultMarker = "*"
		}
		// Truncate URL for display if too long
		displayURL := srv.URL
		if len(displayURL) > 30 {
			displayURL = displayURL[:27] + "..."
		}
		fmt.Printf("%s\t\t%s\t\t%s\n", srv.Name, displayURL, defaultMarker)
	}
}

func handleServersDefault(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: server name is required")
		fmt.Println("Usage: fazt servers default <name>")
		os.Exit(1)
	}

	name := args[0]

	cfg, err := clientconfig.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.SetDefault(name); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Default server set to '%s'\n", name)
}

func handleServersRemove(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: server name is required")
		fmt.Println("Usage: fazt servers remove <name>")
		os.Exit(1)
	}

	name := args[0]

	cfg, err := clientconfig.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.RemoveServer(name); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server '%s' removed\n", name)
}

func handleServersPing(args []string) {
	var serverName string
	if len(args) > 0 {
		serverName = args[0]
	}

	cfg, err := clientconfig.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	srv, name, err := cfg.GetServer(serverName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Pinging %s (%s)...\n", name, srv.URL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(srv.URL + "/health")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Server '%s' is healthy!\n", name)
	} else {
		fmt.Printf("Server '%s' returned status %d\n", name, resp.StatusCode)
	}
}

func printServersHelp() {
	fmt.Println(`Fazt.sh - Server Management

USAGE:
  fazt servers <command> [options]

COMMANDS:
  add <name>     Add a new server configuration
  list           List configured servers
  default <name> Set the default server
  remove <name>  Remove a server configuration
  ping [name]    Test connection to a server

EXAMPLES:
  # Add a server
  fazt servers add prod --url https://zyt.app --token fzt_abc123...

  # List servers
  fazt servers list

  # Set default server
  fazt servers default prod

  # Test connection
  fazt servers ping prod`)
}

// ===================================================================================
// Remote Commands (v0.9.0) - fazt-to-fazt communication
// ===================================================================================

// handleRemoteCommand handles remote peer commands
func handleRemoteCommand(args []string) {
	if len(args) < 1 {
		handleRemoteList()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "add":
		handleRemoteAdd(args[1:])
	case "list":
		handleRemoteList()
	case "remove":
		handleRemoteRemove(args[1:])
	case "default":
		handleRemoteDefault(args[1:])
	case "status":
		handleRemoteStatus(args[1:])
	case "apps":
		fmt.Fprintln(os.Stderr, "DEPRECATED: Use 'fazt app list [peer]' instead")
		handleRemoteApps(args[1:])
	case "upgrade":
		handleRemoteUpgrade(args[1:])
	case "deploy":
		fmt.Fprintln(os.Stderr, "DEPRECATED: Use 'fazt app deploy <dir> --to <peer>' instead")
		handleRemoteDeploy(args[1:])
	case "--help", "-h", "help":
		printRemoteHelp()
	default:
		fmt.Printf("Unknown remote command: %s\n\n", subcommand)
		printRemoteHelp()
		os.Exit(1)
	}
}

func getClientDB() *sql.DB {
	// Use XDG config path for client database
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	dbPath := filepath.Join(configDir, "fazt", "data.db")

	if err := database.Init(dbPath); err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	return database.GetDB()
}

func handleRemoteAdd(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: peer name is required")
		fmt.Println("Usage: fazt remote add <name> --url <url> --token <token>")
		os.Exit(1)
	}

	name := args[0]
	flagArgs := args[1:]

	flags := flag.NewFlagSet("remote add", flag.ExitOnError)
	urlFlag := flags.String("url", "", "Peer URL (required)")
	tokenFlag := flags.String("token", "", "API token (required)")
	descFlag := flags.String("desc", "", "Description")
	flags.Parse(flagArgs)

	if *urlFlag == "" || *tokenFlag == "" {
		fmt.Println("Error: --url and --token are required")
		fmt.Println("Usage: fazt remote add <name> --url <url> --token <token>")
		os.Exit(1)
	}

	db := getClientDB()
	defer database.Close()

	if err := remote.AddPeer(db, name, *urlFlag, *tokenFlag, *descFlag); err != nil {
		if err == remote.ErrPeerAlreadyExists {
			fmt.Printf("Error: peer '%s' already exists\n", name)
		} else {
			fmt.Printf("Error adding peer: %v\n", err)
		}
		os.Exit(1)
	}

	// Set as default if it's the first peer
	peers, _ := remote.ListPeers(db)
	if len(peers) == 1 {
		remote.SetDefaultPeer(db, name)
		fmt.Printf("Peer '%s' added and set as default.\n", name)
	} else {
		fmt.Printf("Peer '%s' added.\n", name)
	}
}

func handleRemoteList() {
	db := getClientDB()
	defer database.Close()

	peers, err := remote.ListPeers(db)
	if err != nil {
		fmt.Printf("Error listing peers: %v\n", err)
		os.Exit(1)
	}

	if len(peers) == 0 {
		fmt.Println("No peers configured.")
		fmt.Println("Run: fazt remote add <name> --url <url> --token <token>")
		return
	}

	fmt.Printf("%-12s %-30s %-10s %-10s\n", "NAME", "URL", "STATUS", "DEFAULT")
	fmt.Println("────────────────────────────────────────────────────────────────")
	for _, p := range peers {
		defaultMark := ""
		if p.IsDefault {
			defaultMark = "*"
		}
		status := p.LastStatus
		if status == "" {
			status = "-"
		}
		// Truncate URL if too long
		displayURL := p.URL
		if len(displayURL) > 28 {
			displayURL = displayURL[:25] + "..."
		}
		fmt.Printf("%-12s %-30s %-10s %-10s\n", p.Name, displayURL, status, defaultMark)
	}
}

func handleRemoteRemove(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: peer name is required")
		fmt.Println("Usage: fazt remote remove <name>")
		os.Exit(1)
	}

	name := args[0]
	db := getClientDB()
	defer database.Close()

	if err := remote.RemovePeer(db, name); err != nil {
		if err == remote.ErrPeerNotFound {
			fmt.Printf("Error: peer '%s' not found\n", name)
		} else {
			fmt.Printf("Error removing peer: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("Peer '%s' removed.\n", name)
}

func handleRemoteDefault(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: peer name is required")
		fmt.Println("Usage: fazt remote default <name>")
		os.Exit(1)
	}

	name := args[0]
	db := getClientDB()
	defer database.Close()

	if err := remote.SetDefaultPeer(db, name); err != nil {
		if err == remote.ErrPeerNotFound {
			fmt.Printf("Error: peer '%s' not found\n", name)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("Default peer set to '%s'.\n", name)
}

func handleRemoteStatus(args []string) {
	var peerName string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		peerName = args[0]
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, peerName)
	if err != nil {
		if err == remote.ErrNoPeers {
			fmt.Println("No peers configured.")
			fmt.Println("Run: fazt remote add <name> --url <url> --token <token>")
		} else if err == remote.ErrNoDefaultPeer {
			fmt.Println("Multiple peers configured. Specify which peer:")
			fmt.Println("  fazt remote status <name>")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		os.Exit(1)
	}

	client := remote.NewClient(peer)

	// Health check first
	healthy, err := client.HealthCheck()
	if err != nil || !healthy {
		fmt.Printf("Server: %s (%s)\n", peer.Name, peer.URL)
		fmt.Printf("Status: UNREACHABLE\n")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		remote.UpdatePeerStatus(db, peer.Name, "unreachable", "")
		os.Exit(1)
	}

	// Get full status
	status, err := client.Status()
	if err != nil {
		fmt.Printf("Server: %s (%s)\n", peer.Name, peer.URL)
		fmt.Printf("Health: OK (HTTP 200)\n")
		fmt.Printf("Status: Auth required or error: %v\n", err)
		remote.UpdatePeerStatus(db, peer.Name, "healthy", "")
		return
	}

	// Update peer status in DB
	remote.UpdatePeerStatus(db, peer.Name, status.Status, status.Version)

	// Display status
	fmt.Printf("Server: %s\n", peer.Name)
	fmt.Printf("URL:    %s\n", peer.URL)
	fmt.Println()
	fmt.Printf("Health:\n")
	fmt.Printf("  Status:     %s\n", status.Status)
	fmt.Printf("  Version:    %s\n", status.Version)
	fmt.Printf("  Mode:       %s\n", status.Mode)
	fmt.Printf("  Uptime:     %s\n", formatDuration(status.Uptime))
	fmt.Println()
	fmt.Printf("Resources:\n")
	fmt.Printf("  Memory:     %.1f MB / %.0f MB\n", status.Memory.UsedMB, status.Memory.LimitMB)
	fmt.Printf("  Goroutines: %d\n", status.Runtime.Goroutines)
	fmt.Printf("  DB Conns:   %d open, %d in use\n", status.Database.OpenConnections, status.Database.InUse)
}

func handleRemoteApps(args []string) {
	var peerName string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		peerName = args[0]
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, peerName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	client := remote.NewClient(peer)
	apps, err := client.Apps()
	if err != nil {
		fmt.Printf("Error fetching apps: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Apps on %s:\n\n", peer.Name)
	fmt.Printf("%-20s %-10s %-10s %-12s\n", "NAME", "FILES", "SIZE", "UPDATED")
	fmt.Println("──────────────────────────────────────────────────────")
	for _, app := range apps {
		updated := app.UpdatedAt
		if len(updated) > 10 {
			updated = updated[:10]
		}
		fmt.Printf("%-20s %-10d %-10s %-12s\n", app.Name, app.FileCount, formatSize(app.SizeBytes), updated)
	}
}

func handleRemoteUpgrade(args []string) {
	checkOnly := false
	var peerName string

	for _, arg := range args {
		if arg == "check" || arg == "--check" {
			checkOnly = true
		} else if !strings.HasPrefix(arg, "-") {
			peerName = arg
		}
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, peerName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	client := remote.NewClient(peer)
	result, err := client.Upgrade(checkOnly)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server:  %s (%s)\n", peer.Name, peer.URL)
	fmt.Printf("Current: v%s\n", result.CurrentVersion)
	fmt.Printf("Latest:  v%s\n", result.NewVersion)
	fmt.Println()

	switch result.Action {
	case "already_latest":
		fmt.Println("Already running the latest version.")
	case "check_only":
		fmt.Println("Update available!")
		fmt.Println("Run: fazt remote upgrade", peer.Name)
	case "upgraded":
		fmt.Println("Upgraded! Server is restarting...")
	}
}

func handleRemoteDeploy(args []string) {
	flags := flag.NewFlagSet("remote deploy", flag.ExitOnError)
	siteName := flags.String("name", "", "Site name (defaults to directory name)")
	peerFlag := flags.String("to", "", "Target peer name")

	flags.Usage = func() {
		fmt.Println("Usage: fazt remote deploy <directory> [--name <site>] [--to <peer>]")
		fmt.Println()
		flags.PrintDefaults()
	}

	// Find directory arg (first non-flag arg)
	var dir string
	var flagArgs []string
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") && dir == "" {
			dir = arg
			flagArgs = args[i+1:]
			break
		}
	}

	if dir == "" {
		fmt.Println("Error: directory is required")
		flags.Usage()
		os.Exit(1)
	}

	flags.Parse(flagArgs)

	// Validate directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Error: directory '%s' does not exist\n", dir)
		os.Exit(1)
	}

	// Determine site name
	name := *siteName
	if name == "" {
		name = filepath.Base(dir)
		if name == "." {
			wd, _ := os.Getwd()
			name = filepath.Base(wd)
		}
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, *peerFlag)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Deploying '%s' to %s as '%s'...\n", dir, peer.Name, name)

	// Create ZIP
	zipBuffer, fileCount, err := createDeployZip(dir)
	if err != nil {
		fmt.Printf("Error creating ZIP: %v\n", err)
		os.Exit(1)
	}

	// Write to temp file (client expects file path)
	tmpFile, err := os.CreateTemp("", "deploy-*.zip")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(zipBuffer.Bytes()); err != nil {
		fmt.Printf("Error writing ZIP: %v\n", err)
		os.Exit(1)
	}
	tmpFile.Close()

	fmt.Printf("Zipped %d files (%s)\n", fileCount, formatSize(int64(zipBuffer.Len())))

	client := remote.NewClient(peer)
	result, err := client.Deploy(tmpFile.Name(), name)
	if err != nil {
		fmt.Printf("Error deploying: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("Deployed: %s\n", result.Site)
	fmt.Printf("Files:    %d\n", result.FileCount)
	fmt.Printf("Size:     %s\n", formatSize(result.SizeBytes))
}

func formatDuration(seconds float64) string {
	d := time.Duration(seconds) * time.Second
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func printRemoteHelp() {
	fmt.Println(`Fazt.sh - Remote Peer Management

USAGE:
  fazt remote <command> [options]

COMMANDS:
  add <name>       Add a remote peer
  list             List configured peers
  remove <name>    Remove a peer
  default <name>   Set the default peer
  status [name]    Check peer health and status
  upgrade [name]   Check/perform upgrade on peer

DEPRECATED (use 'fazt app' instead):
  apps [name]      → fazt app list [peer]
  deploy <dir>     → fazt app deploy <dir> --to <peer>

EXAMPLES:
  # Add a peer
  fazt remote add zyt --url https://admin.zyt.app --token xxx

  # Check status (uses default if only one peer)
  fazt remote status

  # List apps on specific peer (NEW)
  fazt app list zyt

  # Deploy to peer (NEW)
  fazt app deploy ./my-site --to zyt

  # Check for upgrades
  fazt remote upgrade check

  # Perform upgrade
  fazt remote upgrade zyt`)
}

// handleServiceCommand handles service-related subcommands
func handleServiceCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: service command requires a subcommand")
		printServiceHelp()
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "install":
		handleInstallCommand() // Reuse the install logic but moved here
	case "start":
		if err := provision.Systemctl("start", "fazt"); err != nil {
			fmt.Printf("Error starting service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service started.")
	case "stop":
		if err := provision.Systemctl("stop", "fazt"); err != nil {
			fmt.Printf("Error stopping service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service stopped.")
	case "status":
		if err := provision.Systemctl("status", "fazt"); err != nil {
			// Systemctl status returns non-zero if service is not running, which is fine to show
			// os.Exit(1) 
		}
	case "logs":
		if err := provision.ServiceLogs("fazt"); err != nil {
			fmt.Printf("Error reading logs: %v\n", err)
			os.Exit(1)
		}
	case "--help", "-h", "help":
		printServiceHelp()
	default:
		fmt.Printf("Unknown service command: %s\n\n", subcommand)
		printServiceHelp()
		os.Exit(1)
	}
}

// handleClientCommand handles client-related subcommands
func handleClientCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: client command requires a subcommand")
		printClientHelp()
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "set-auth-token":
		handleSetAuthToken()
	case "deploy":
		handleDeployCommand()
	case "logs":
		handleLogsCommand()
	case "sites":
		handleSitesCommand()
	case "apps":
		handleAppsCommand()
	case "delete":
		handleDeleteCommand()
	case "--help", "-h", "help":
		printClientHelp()
	default:
		fmt.Printf("Unknown client command: %s\n\n", subcommand)
		printClientHelp()
		os.Exit(1)
	}
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		requestID := r.Header.Get("X-Request-ID")
		if requestID != "" {
			log.Printf("[%s] %s %s %d %v", requestID, r.Method, r.URL.Path, wrapped.statusCode, duration)
		} else {
			log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
		}
	})
}

// corsMiddleware adds CORS headers for development
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := config.Get()

		if cfg.IsDevelopment() {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware recovers from panics and logs the error
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Hijack implements http.Hijacker for WebSocket support
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
}

// printVersion displays version information
func printVersion() {
	fmt.Printf("fazt.sh %s\n", config.Version)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// createRootHandler creates a handler that routes based on the Host header

// - Requests to admin.domain go to the dashboard

// - Requests to root.domain or domain go to the "root" site

// - Requests to subdomains go to the site handler

func createRootHandler(cfg *config.Config, dashboardMux *http.ServeMux, sessionStore *auth.SessionStore) http.Handler {

	// Parse the main domain from config

	mainDomain := extractDomain(cfg.Server.Domain)



	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		host := r.Host



		// Remove port from host if present

		if colonIdx := strings.LastIndex(host, ":"); colonIdx != -1 {

			// Check if this is IPv6 (has brackets)

			if !strings.Contains(host, "]") || strings.LastIndex(host, "]") < colonIdx {

				host = host[:colonIdx]

			}

		}

		// Local-only: /_app/<id>/ routes for direct app access by ID
		// Only available from local/private IPs (dev/testing escape hatch)
		if appID, remaining, ok := hosting.ParseAppPath(r.URL.Path); ok {
			if !hosting.IsLocalRequest(r) {
				// Return 404 (not 401) to avoid revealing the route exists
				http.NotFound(w, r)
				return
			}
			// Rewrite the request path and serve the app
			r.URL.Path = remaining
			siteHandler(w, r, appID)
			return
		}

		// Special case: localhost serves Dashboard (for CLI/Dev simplicity)

		// Users can still test sites via Host headers:

		// curl -H "Host: root.localhost" ...

		if host == "localhost" {

			middleware.AuthMiddleware(sessionStore)(dashboardMux).ServeHTTP(w, r)

			return

		}



		// 1. Dashboard Routing (admin.<domain>)

		if host == "admin."+mainDomain {

			middleware.AuthMiddleware(sessionStore)(dashboardMux).ServeHTTP(w, r)

			return

		}



		// 2. Root Domain Routing (root.<domain> or <domain>)

		if host == "root."+mainDomain || host == mainDomain {

			siteHandler(w, r, "root")

			return

		}



		// 3. 404 Domain Routing

		if host == "404."+mainDomain {

			siteHandler(w, r, "404")

			return

		}



		// 4. Subdomain Routing

		subdomain := extractSubdomain(host, mainDomain)

		if subdomain != "" {

			siteHandler(w, r, subdomain)

			return

		}



		// Fallback -> 404

		serveSiteNotFound(w, r, host)

	})

}

// extractDomain extracts the domain from a URL (removes protocol and path)
func extractDomain(rawURL string) string {
	// Handle URLs with protocol
	if strings.Contains(rawURL, "://") {
		if parsed, err := url.Parse(rawURL); err == nil {
			return parsed.Hostname()
		}
	}
	// Handle bare domains
	if colonIdx := strings.Index(rawURL, ":"); colonIdx != -1 {
		return rawURL[:colonIdx]
	}
	return rawURL
}

// isDashboardHost checks if the host should be routed to the dashboard

// extractSubdomain extracts the subdomain from a host
// e.g., "blog.example.com" with mainDomain "example.com" returns "blog"
// e.g., "blog.localhost" returns "blog"
func extractSubdomain(host, mainDomain string) string {
	host = strings.ToLower(host)
	mainDomain = strings.ToLower(mainDomain)

	// Handle *.localhost pattern
	if strings.HasSuffix(host, ".localhost") {
		return strings.TrimSuffix(host, ".localhost")
	}

	// Handle *.127.0.0.1 pattern (rare but possible)
	if strings.HasSuffix(host, ".127.0.0.1") {
		return strings.TrimSuffix(host, ".127.0.0.1")
	}

	// Handle *.mainDomain pattern
	suffix := "." + mainDomain
	if strings.HasSuffix(host, suffix) {
		subdomain := strings.TrimSuffix(host, suffix)
		// Don't return empty subdomain or subdomain with dots (nested subdomains)
		if subdomain != "" && !strings.Contains(subdomain, ".") {
			return subdomain
		}
	}

	return ""
}

// siteHandler handles requests for hosted sites
// v0.10: First resolves alias to app_id, then serves files from VFS
// If main.js exists, executes serverless JavaScript instead
// WebSocket connections at /ws are handled by the WebSocket hub
// API paths (/api or /api/*) are handled by the serverless handler with storage
func siteHandler(w http.ResponseWriter, r *http.Request, subdomain string) {
	// v0.10: Resolve alias to app_id
	appID, aliasType, err := handlers.ResolveAlias(subdomain)
	if err != nil {
		// Alias resolution failed, try legacy lookup by site_id
		appID = subdomain
	}

	// Handle alias types
	switch aliasType {
	case "reserved":
		// Reserved subdomain - return 404
		serveSiteNotFound(w, r, subdomain)
		return
	case "redirect":
		// Get redirect URL and redirect
		redirectURL, err := handlers.GetRedirectURL(subdomain)
		if err == nil && redirectURL != "" {
			http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
			return
		}
	}

	// Use subdomain for file lookups (files are stored with site_id = subdomain)
	// Use appID for analytics and identity tracking
	siteID := subdomain

	// Check if site exists
	if !hosting.SiteExists(subdomain) && appID == "" {
		serveSiteNotFound(w, r, subdomain)
		return
	}

	// Handle WebSocket connections at /_ws
	if r.URL.Path == "/_ws" {
		hosting.HandleWebSocket(w, r, subdomain)
		return
	}

	// Log analytics event for site visits (using app_id for v0.10, fallback to subdomain)
	analyticsID := subdomain
	if appID != "" {
		analyticsID = appID
	}
	logSiteVisit(r, analyticsID)

	// Check for API paths (/api or /api/*)
	// These are handled by the serverless handler with storage support
	if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
		// Check for api/main.js
		fs := hosting.GetFileSystem()
		hasAPI, _ := fs.Exists(siteID, "api/main.js")
		if hasAPI && serverlessHandler != nil {
			serverlessHandler.HandleRequest(w, r, siteID, siteID)
			return
		}
		// No api/main.js found
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}

	// Serve from VFS using subdomain as site_id
	hosting.ServeVFS(w, r, siteID)
}

// logSiteVisit logs an analytics event for a site visit
func logSiteVisit(r *http.Request, subdomain string) {
	analytics.Add(analytics.Event{
		Domain:      subdomain,
		SourceType:  "hosting",
		EventType:   "pageview",
		Path:        r.URL.Path,
		Referrer:    r.Referer(),
		UserAgent:   r.UserAgent(),
		IPAddress:   r.RemoteAddr,
		QueryParams: r.URL.RawQuery,
	})
}

// serveSiteNotFound renders the 404 page for non-existent sites
func serveSiteNotFound(w http.ResponseWriter, r *http.Request, subdomain string) {
	// Try to serve universal 404 site if it exists
	if hosting.SiteExists("404") {
		// Use the 404 site content
		// We pass "404" as the site ID
		w.WriteHeader(http.StatusNotFound) // Ensure we still send 404 status
		hosting.ServeVFS(w, r, "404")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `<!DOCTYPE html><html>
<head>
    <title>Site Not Found</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
               display: flex; justify-content: center; align-items: center;
               height: 100vh; margin: 0; background: #f5f5f5; }
        .container { text-align: center; padding: 40px; background: white;
                     border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; margin-bottom: 10px; }
        p { color: #666; }
        .subdomain { font-family: monospace; background: #f0f0f0; padding: 2px 8px; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>404 - Site Not Found</h1>
        <p>The site <span class="subdomain">%s</span> does not exist.</p>
    </div>
</body>
</html>`, subdomain)
}

// createDeployZip creates a ZIP archive of the directory, respecting .gitignore
func createDeployZip(dir string) (*bytes.Buffer, int, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	fileCount := 0

	// Load .gitignore if present
	var gitignore *ignore.GitIgnore
	gitignorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		gitignore, _ = ignore.CompileIgnoreFile(gitignorePath)
	}

	// Default ignores (always skip these)
	defaultIgnores := []string{
		"node_modules",
		".git",
		".DS_Store",
		"*.log",
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path first (needed for gitignore matching)
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Skip hidden files and directories (except the root)
		if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check default ignores
		for _, pattern := range defaultIgnores {
			if matched, _ := filepath.Match(pattern, info.Name()); matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check .gitignore patterns
		if gitignore != nil && relPath != "." {
			// For directories, append / to match gitignore conventions
			matchPath := relPath
			if info.IsDir() {
				matchPath = relPath + "/"
			}
			if gitignore.MatchesPath(matchPath) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Skip directories (we only store files)
		if info.IsDir() {
			return nil
		}

		// Create ZIP entry
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Copy file contents
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}

		fileCount++
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, 0, err
	}

	return buf, fileCount, nil
}

// formatSize formats bytes to human readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// handleSetCredentials handles the set-credentials subcommand
func handleSetCredentials() {
	flags := flag.NewFlagSet("set-credentials", flag.ExitOnError)
	username := flags.String("username", "", "Username for authentication")
	password := flags.String("password", "", "Password for authentication")
	db := flags.String("db", "", "Database file path")

	flags.Usage = func() {
		fmt.Println("Usage: fazt server set-credentials [flags]")
		fmt.Println()
		fmt.Println("Update authentication credentials for the fazt.sh dashboard.")
		fmt.Println("At least one of --username or --password must be provided.")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  fazt server set-credentials --username newuser")
		fmt.Println("  fazt server set-credentials --password newpass")
		fmt.Println("  fazt server set-credentials --username admin --password secret123")
		fmt.Println("  fazt server set-credentials --db /path/to/data.db")
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Resolve DB Path
	dbPath := "./data.db"
	if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
		dbPath = envPath
	}
	if *db != "" {
		dbPath = config.ExpandPath(*db)
	}

	// Call command function
	if err := setCredentialsCommand(*username, *password, dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Credentials updated successfully")
	if *username != "" {
		fmt.Printf("  Username: %s\n", *username)
	}
	if *password != "" {
		fmt.Println("  Password: [updated and hashed]")
	}
	fmt.Println()
}

// handleInitCommand handles the init subcommand
func handleInitCommand() {
	flags := flag.NewFlagSet("init", flag.ExitOnError)
	username := flags.String("username", "", "Admin username (interactive if empty)")
	password := flags.String("password", "", "Admin password (interactive if empty)")
	domain := flags.String("domain", "", "Server domain (interactive if empty)")
	port := flags.String("port", "4698", "Server port")
	env := flags.String("env", "development", "Environment (development|production)")
	db := flags.String("db", "", "Database file path")

	flags.Usage = func() {
		fmt.Println("Usage: fazt server init [flags]")
		fmt.Println()
		fmt.Println("Initialize fazt.sh server configuration")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  fazt server init")
		fmt.Println("  fazt server init --username admin --password secret --domain localhost")
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Interactive Prompts
	if *username == "" {
		fmt.Print("Admin Username: ")
		fmt.Scanln(username)
	}
	if *password == "" {
		fmt.Print("Admin Password: ")
		fmt.Scanln(password)
	}
	if *domain == "" {
		fmt.Print("Server Domain (default: localhost): ")
		var input string
		fmt.Scanln(&input)
		if input != "" {
			*domain = input
		} else {
			*domain = "localhost"
		}
	}

	// Get DB path
	dbPath := "./data.db"
	if *db != "" {
		dbPath = config.ExpandPath(*db)
	}

	// Call command function
	if err := initCommand(*username, *password, *domain, *port, *env, dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Server initialized successfully")
	fmt.Printf("  Database: %s\n", dbPath)
	fmt.Println()
	fmt.Println("To start the server:")
	fmt.Println("  fazt server start")
	fmt.Println()
}

// handleSetConfigCommand handles the set-config subcommand
func handleSetConfigCommand() {
	flags := flag.NewFlagSet("set-config", flag.ExitOnError)
	domain := flags.String("domain", "", "Server domain")
	port := flags.String("port", "", "Server port")
	env := flags.String("env", "", "Environment (development|production)")
	db := flags.String("db", "", "Database file path")

	flags.Usage = func() {
		fmt.Println("Usage: fazt server set-config [flags]")
		fmt.Println()
		fmt.Println("Update server configuration settings")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  fazt server set-config --domain https://newdomain.com")
		fmt.Println("  fazt server set-config --port 8080")
		fmt.Println("  fazt server set-config --env production")
		fmt.Println("  fazt server set-config --domain https://prod.com --port 443 --env production")
		fmt.Println("  fazt server set-config --domain https://prod.com --db /path/to/data.db")
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Resolve DB Path
	dbPath := "./data.db"
	if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
		dbPath = envPath
	}
	if *db != "" {
		dbPath = config.ExpandPath(*db)
	}

	// Call command function
	if err := setConfigCommand(*domain, *port, *env, dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Configuration updated successfully")
	if *domain != "" {
		fmt.Printf("  Domain: %s\n", *domain)
	}
	if *port != "" {
		fmt.Printf("  Port: %s\n", *port)
	}
	if *env != "" {
		fmt.Printf("  Environment: %s\n", *env)
	}
	fmt.Println()
}

// handleStatusCommand handles the status subcommand
func handleStatusCommand() {
	flags := flag.NewFlagSet("status", flag.ExitOnError)
	db := flags.String("db", "", "Database file path")

	flags.Usage = func() {
		fmt.Println("Usage: fazt server status [flags]")
		fmt.Println()
		fmt.Println("Display server configuration and status")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Shows:")
		fmt.Println("  Server settings (domain, port, environment)")
		fmt.Println("  Authentication status")
		fmt.Println("  Database information")
		fmt.Println("  VFS Status")
		fmt.Println("  Server running status")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  fazt server status")
		fmt.Println("  fazt server status --db /path/to/data.db")
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Resolve DB Path
	dbPath := "./data.db"
	if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
		dbPath = envPath
	}
	if *db != "" {
		dbPath = config.ExpandPath(*db)
	}

	// Call command function
	output, err := statusCommand(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}

// handleSetAuthToken handles the set-auth-token subcommand
func handleSetAuthToken() {
	flags := flag.NewFlagSet("set-auth-token", flag.ExitOnError)
	token := flags.String("token", "", "Authentication token (required)")
	server := flags.String("server", "", "Server URL (optional, defaults to http://localhost:4698)")
	db := flags.String("db", "", "Database file path")

	flags.Usage = func() {
		fmt.Println("Usage: fazt client set-auth-token --token <TOKEN> [options]")
		fmt.Println()
		fmt.Println("Sets the authentication token and optional server URL.")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  fazt client set-auth-token --token abc... --server https://fazt.example.com")
		fmt.Println()
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	if *token == "" {
		fmt.Println("Error: --token is required")
		flags.Usage()
		os.Exit(1)
	}

	// Resolve DB Path
	dbPath := "./data.db"
	if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
		dbPath = envPath
	}
	if *db != "" {
		dbPath = config.ExpandPath(*db)
	}

	// Initialize DB
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer database.Close()
	
	store := config.NewDBConfigStore(database.GetDB())

	// Set token
	if err := store.Set("api_key.token", *token); err != nil {
		log.Fatalf("Failed to save token: %v", err)
	}
	store.Set("api_key.name", "deployment-token")

	// Set server URL if provided
	if *server != "" {
		if err := store.Set("client.server_url", *server); err != nil {
			log.Fatalf("Failed to save server url: %v", err)
		}
	}

	fmt.Println("✓ Configuration saved!")
	if *server != "" {
		fmt.Printf("  Server: %s\n", *server)
	}
	fmt.Printf("  Token:  %s...\n", (*token)[:4])
}
func handleDeployCommand() {
	flags := flag.NewFlagSet("deploy", flag.ExitOnError)
	path := flags.String("path", "", "Directory to deploy (required)")
	domain := flags.String("domain", "", "Domain/subdomain for the site (required)")
	toServer := flags.String("to", "", "Target server name (from 'fazt servers list')")
	server := flags.String("server", "", "Server URL override (deprecated, use --to)")

	flags.Usage = func() {
		fmt.Println("Usage: fazt deploy --path <PATH> --domain <SUBDOMAIN> [--to <SERVER>]")
		fmt.Println()
		fmt.Println("Deploys a directory to a fazt.sh server.")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  fazt deploy --path . --domain my-site")
		fmt.Println("  fazt deploy --path . --domain my-site --to prod")
		fmt.Println("  fazt client deploy --path ~/Desktop/site --domain example")
	}

	// Determine args offset based on whether this is "deploy" or "client deploy"
	argsOffset := 3
	if len(os.Args) > 1 && os.Args[1] == "deploy" {
		argsOffset = 2
	}

	if err := flags.Parse(os.Args[argsOffset:]); err != nil {
		os.Exit(1)
	}

	deployPath := *path
	if deployPath == "" {
		fmt.Println("Error: --path is required")
		flags.Usage()
		os.Exit(1)
	}

	if *domain == "" {
		fmt.Println("Error: --domain is required")
		flags.Usage()
		os.Exit(1)
	}

	// Validate the path exists
	if _, err := os.Stat(deployPath); os.IsNotExist(err) {
		fmt.Printf("Error: Path '%s' does not exist\n", deployPath)
		os.Exit(1)
	}

	// Try new config first (~/.fazt/config.json)
	var serverURL, token string
	cfg, err := clientconfig.Load()
	if err == nil && cfg.ServerCount() > 0 {
		// Use new config system
		srv, _, err := cfg.GetServer(*toServer)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		serverURL = srv.URL
		token = srv.Token
	} else {
		// Fall back to old data.db config for backwards compatibility
		dbPath := "./data.db"
		if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
			dbPath = envPath
		}

		if err := database.Init(dbPath); err != nil {
			fmt.Println("Error: No servers configured")
			fmt.Println("Run: fazt servers add <name> --url <url> --token <token>")
			os.Exit(1)
		}
		defer database.Close()

		store := config.NewDBConfigStore(database.GetDB())
		dbMap, _ := store.Load()
		token = dbMap["api_key.token"]

		if token == "" {
			fmt.Println("Error: No API key found in configuration")
			fmt.Println("Run: fazt servers add <name> --url <url> --token <token>")
			os.Exit(1)
		}

		serverURL = "http://localhost:4698"
		if dbURL, ok := dbMap["client.server_url"]; ok && dbURL != "" {
			serverURL = dbURL
		}
	}

	// Allow --server flag to override (for backwards compat)
	if *server != "" {
		serverURL = *server
	}

	fmt.Printf("Deploying %s to %s as '%s'...\n", deployPath, serverURL, *domain)

	// Change to the deploy directory
	originalDir, _ := os.Getwd()
	if err := os.Chdir(deployPath); err != nil {
		fmt.Printf("Error changing to directory %s: %v\n", deployPath, err)
		os.Exit(1)
	}
	defer os.Chdir(originalDir)

	// Create ZIP of the directory
	zipBuffer, fileCount, err := createDeployZip(".")
	if err != nil {
		fmt.Printf("Error creating ZIP: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Zipped %d files (%d bytes)\n", fileCount, zipBuffer.Len())

	// Create HTTP request
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add domain field
	if err := writer.WriteField("site_name", *domain); err != nil {
		fmt.Printf("Error creating form: %v\n", err)
		os.Exit(1)
	}

	// Add file field
	part, err := writer.CreateFormFile("file", "deploy.zip")
	if err != nil {
		fmt.Printf("Error creating file field: %v\n", err)
		os.Exit(1)
	}
	if _, err := io.Copy(part, zipBuffer); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}
	writer.Close()

	// Make request
	req, err := http.NewRequest("POST", serverURL+"/api/deploy", &body)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error deploying: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Check response
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("✗ Deployment failed!\n")
		fmt.Printf("  Status: %s\n", resp.Status)
		fmt.Printf("  Error: %s\n", string(respBody))
		os.Exit(1)
	}

	// Parse success response
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err == nil {
		if success, ok := result["success"].(bool); ok && success {
			fmt.Printf("✓ Deployment successful!\n")
			if site, ok := result["site"].(string); ok {
				// Extract server URL for display
				serverURL := *server
				serverURL = strings.TrimPrefix(serverURL, "http://")
				serverURL = strings.TrimPrefix(serverURL, "https://")
				fmt.Printf("  Site: http://%s.%s\n", site, serverURL)
			}
			if fileCount, ok := result["file_count"].(float64); ok {
				fmt.Printf("  Files: %.0f\n", fileCount)
			}
			if sizeBytes, ok := result["size_bytes"].(float64); ok {
				fmt.Printf("  Size: %.0f bytes\n", sizeBytes)
			}
			return
		}
	}

	fmt.Printf("✓ Deployment completed! (Status: %s)\n", resp.Status)
}

// handleLogsCommand handles the logs subcommand
func handleLogsCommand() {
	flags := flag.NewFlagSet("logs", flag.ExitOnError)
	site := flags.String("site", "", "Site name (subdomain) (required)")
	limit := flags.Int("limit", 50, "Number of logs to fetch")
	server := flags.String("server", "http://localhost:4698", "fazt.sh server URL")

	flags.Usage = func() {
		fmt.Println("Usage: fazt client logs --site <SITE> [options]")
		fmt.Println()
		fmt.Println("Fetch logs for a specific site.")
		fmt.Println()
		flags.PrintDefaults()
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	if *site == "" {
		fmt.Println("Error: --site is required")
		flags.Usage()
		os.Exit(1)
	}

	// Resolve DB Path
	dbPath := "./data.db"
	if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	// Initialize DB to get token
	if err := database.Init(dbPath); err != nil {
		fmt.Printf("Error: Failed to init database at %s: %v\n", dbPath, err)
		os.Exit(1)
	}
	defer database.Close()

	store := config.NewDBConfigStore(database.GetDB())
	dbMap, _ := store.Load()
	token := dbMap["api_key.token"]

	if token == "" {
		fmt.Println("Error: No API key found in config")
		os.Exit(1)
	}

	// Resolve Server URL
	serverURL := "http://localhost:4698"
	if dbURL, ok := dbMap["client.server_url"]; ok && dbURL != "" {
		serverURL = dbURL
	}
	if *server != "http://localhost:4698" {
		serverURL = *server
	}

	url := fmt.Sprintf("%s/api/logs?site_id=%s&limit=%d", serverURL, *site, *limit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching logs: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	var result struct {
		Success bool `json:"success"`
		Logs    []struct {
			Level     string `json:"level"`
			Message   string `json:"message"`
			CreatedAt string `json:"created_at"`
		} `json:"logs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Logs for %s (last %d):\n", *site, *limit)
	for _, log := range result.Logs {
		fmt.Printf("[%s] [%s] %s\n", log.CreatedAt, log.Level, log.Message)
	}
}

// handleSitesCommand handles the sites list subcommand
func handleSitesCommand() {
	flags := flag.NewFlagSet("sites", flag.ExitOnError)
	server := flags.String("server", "http://localhost:4698", "fazt.sh server URL")

	flags.Usage = func() {
		fmt.Println("Usage: fazt client sites [options]")
		fmt.Println()
		fmt.Println("List all deployed sites.")
		fmt.Println()
		flags.PrintDefaults()
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Resolve DB Path
	dbPath := "./data.db"
	if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	// Initialize DB to get token
	if err := database.Init(dbPath); err != nil {
		fmt.Printf("Error: Failed to init database at %s: %v\n", dbPath, err)
		fmt.Println("Run 'fazt client set-auth-token' to configure your client.")
		os.Exit(1)
	}
	defer database.Close()

	store := config.NewDBConfigStore(database.GetDB())
	dbMap, _ := store.Load()
	token := dbMap["api_key.token"]

	if token == "" {
		fmt.Println("Error: No API key found in configuration")
		fmt.Printf("Database: %s\n", dbPath)
		fmt.Println("Please run: fazt client set-auth-token --token <YOUR_TOKEN>")
		os.Exit(1)
	}

	// Resolve Server URL
	serverURL := "http://localhost:4698"
	if dbURL, ok := dbMap["client.server_url"]; ok && dbURL != "" {
		serverURL = dbURL
	}
	if *server != "http://localhost:4698" {
		serverURL = *server
	}

	req, err := http.NewRequest("GET", serverURL+"/api/sites", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching sites: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	var result struct {
		Data []struct {
			Name      string      `json:"Name"`
			FileCount int         `json:"FileCount"`
			SizeBytes int64       `json:"SizeBytes"`
			ModTime   interface{} `json:"ModTime"`
		} `json:"data"`
		Error interface{} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if result.Error != nil {
		fmt.Printf("API Error: %v\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("%-20s %-10s %-10s\n", "SITE", "FILES", "SIZE")
	fmt.Println("──────────────────────────────────────────")
	for _, site := range result.Data {
		fmt.Printf("%-20s %-10d %-10d\n", site.Name, site.FileCount, site.SizeBytes)
	}
}

// handleAppsCommand handles the apps list subcommand (new API)
func handleAppsCommand() {
	flags := flag.NewFlagSet("apps", flag.ExitOnError)
	toServer := flags.String("to", "", "Target server (or default)")

	flags.Usage = func() {
		fmt.Println("Usage: fazt client apps [options]")
		fmt.Println()
		fmt.Println("List all deployed apps.")
		fmt.Println()
		flags.PrintDefaults()
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Load client config
	cfg, err := clientconfig.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Get target server
	srv, serverName, err := cfg.GetServer(*toServer)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("GET", srv.URL+"/api/apps", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+srv.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching apps: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	var result struct {
		Data []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Source    string `json:"source"`
			FileCount int    `json:"file_count"`
			SizeBytes int64  `json:"size_bytes"`
			CreatedAt string `json:"created_at"`
		} `json:"data"`
		Error interface{} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if result.Error != nil {
		fmt.Printf("API Error: %v\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("Apps on %s:\n\n", serverName)
	fmt.Printf("%-20s %-10s %-10s %-10s\n", "NAME", "SOURCE", "FILES", "SIZE")
	fmt.Println("────────────────────────────────────────────────────")
	for _, app := range result.Data {
		source := app.Source
		if source == "" {
			source = "deploy"
		}
		fmt.Printf("%-20s %-10s %-10d %-10d\n", app.Name, source, app.FileCount, app.SizeBytes)
	}
}

// handleDeleteCommand handles the delete site subcommand
func handleDeleteCommand() {
	flags := flag.NewFlagSet("delete", flag.ExitOnError)
	site := flags.String("site", "", "Site name (subdomain) (required)")
	server := flags.String("server", "http://localhost:4698", "fazt.sh server URL")
	confirm := flags.Bool("yes", false, "Skip confirmation")

	flags.Usage = func() {
		fmt.Println("Usage: fazt client delete --site <SITE> [options]")
		fmt.Println()
		fmt.Println("Delete a deployed site.")
		fmt.Println()
		flags.PrintDefaults()
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	if *site == "" {
		fmt.Println("Error: --site is required")
		flags.Usage()
		os.Exit(1)
	}

	if !*confirm {
		fmt.Printf("Are you sure you want to delete site '%s'? [y/N] ", *site)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled.")
			os.Exit(0)
		}
	}

	// Resolve DB Path
	dbPath := "./data.db"
	if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	// Initialize DB to get token
	if err := database.Init(dbPath); err != nil {
		fmt.Printf("Error: Failed to init database at %s: %v\n", dbPath, err)
		os.Exit(1)
	}
	defer database.Close()

	store := config.NewDBConfigStore(database.GetDB())
	dbMap, _ := store.Load()
	token := dbMap["api_key.token"]

	if token == "" {
		fmt.Println("Error: No API key found in config")
		os.Exit(1)
	}

	// Resolve Server URL
	serverURL := "http://localhost:4698"
	if dbURL, ok := dbMap["client.server_url"]; ok && dbURL != "" {
		serverURL = dbURL
	}
	if *server != "http://localhost:4698" {
		serverURL = *server
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/sites?site_id=%s", serverURL, *site), nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error deleting site: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	fmt.Printf("✓ Site '%s' deleted successfully.\n", *site)
}

// handleUpgradeCommand handles the self-upgrade
func handleUpgradeCommand() {
	if os.Geteuid() != 0 {
		fmt.Println("Warning: upgrading typically requires root privileges (sudo).")
		fmt.Println("If this fails, try: sudo fazt upgrade")
		fmt.Println()
	}

	if err := provision.Upgrade(config.Version); err != nil {
		fmt.Printf("Error upgrading: %v\n", err)
		os.Exit(1)
	}
}

// handleStartCommand handles the start subcommand
func handleStartCommand() {
	flags := flag.NewFlagSet("start", flag.ExitOnError)
	port := flags.String("port", "", "Server port (overrides config)")
	db := flags.String("db", "", "Database file path (overrides config)")
	configFile := flags.String("config", "", "Config file path")
	domain := flags.String("domain", "", "Server domain (overrides config)")

	flags.Usage = func() {
		fmt.Println("Usage: fazt server start [options]")
		fmt.Println()
		fmt.Println("Starts the fazt.sh server.")
		fmt.Println()
		fmt.Println("Domain Configuration:")
		fmt.Println("  Default: https://fazt.sh (for project use)")
		fmt.Println("  Override: --domain yourdomain.com")
		fmt.Println("  Environment: FAZT_DOMAIN environment variable")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  cc-server server start")
		fmt.Println("  cc-server server start --port 8080")
		fmt.Println("  cc-server server start --domain mysite.com")
		fmt.Println("  cc-server server start --config /path/to/config.json")
		fmt.Println()
		fmt.Println("Environment Variables:")
		fmt.Println("  FAZT_DOMAIN=fazt.sh cc-server server start")
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Set up configuration
	if !*quiet {
		log.Println("Starting fazt.sh...")
	}

	// Use default flags structure but override with our specific flags
	cliFlags := config.ParseFlags()
	if *port != "" {
		cliFlags.Port = *port
	}
	if *db != "" {
		cliFlags.DBPath = *db
	}
	if *configFile != "" {
		cliFlags.ConfigPath = *configFile
	}
	// Load configuration
	cfg, err := config.Load(cliFlags)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Apply domain override if provided (highest priority)
	// Wrap with wildcard DNS if it's an IP address
	if *domain != "" {
		cfg.Server.Domain = config.WrapWithWildcardDNS(*domain)
	}

	// Initialize database EARLY to load config
	if err := database.Init(cfg.Database.Path); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Overlay configuration from database
	if err := config.OverlayDB(database.GetDB(), cliFlags); err != nil {
		// Log warning but don't fail
		log.Printf("Warning: Failed to load config from DB: %v", err)
	}

	// Portable database support: auto-adjust machine-specific domains
	// - Real domains (zyt.app, example.com): Always trusted, never touched
	// - Wildcard DNS (*.nip.io): Check if IP matches, update if different machine
	// - IP addresses: Check if matches local, update if not
	// - Empty: Auto-detect local IP
	if *domain == "" {
		localIP := provision.GetPrimaryLocalIP()

		if cfg.Server.Domain == "" {
			// No domain configured - auto-detect
			log.Printf("No domain configured, using detected IP: %s", localIP)
			cfg.Server.Domain = config.WrapWithWildcardDNS(localIP)
		} else if provision.IsPortableDomain(cfg.Server.Domain) {
			// Machine-specific domain (IP or wildcard DNS) - check if needs update
			match, _ := provision.DetectEnvironment(cfg.Server.Domain)
			if match == provision.EnvMismatch {
				log.Printf("Portable DB: Updating domain from '%s' to '%s'",
					cfg.Server.Domain, config.WrapWithWildcardDNS(localIP))
				cfg.Server.Domain = config.WrapWithWildcardDNS(localIP)
			}
		}
		// Real domains are always trusted - no detection needed
	}

	// Validate configuration now that we have loaded everything
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Ensure secure file permissions
	security.EnsureSecurePermissions(config.ExpandPath(cliFlags.ConfigPath), cfg.Database.Path)

	// Display startup information
	term.Banner()
	term.Section(fmt.Sprintf("fazt.sh %s - Starting Up", config.Version))

	fmt.Printf("  Environment:  %s\n", cfg.Server.Env)
	fmt.Printf("  Port:         %s\n", cfg.Server.Port)
	fmt.Printf("  Domain:       %s\n", cfg.Server.Domain)
	fmt.Printf("  Database:     %s\n", cfg.Database.Path)

	// Show accessible URLs
	protocol := "http"
	if cfg.HTTPS.Enabled {
		protocol = "https"
	}
	portSuffix := ""
	if cfg.Server.Port != "80" && cfg.Server.Port != "443" {
		portSuffix = ":" + cfg.Server.Port
	}
	fmt.Println()
	fmt.Printf("  Dashboard:    %s://admin.%s%s\n", protocol, cfg.Server.Domain, portSuffix)
	fmt.Printf("  Apps:         %s://<app>.%s%s\n", protocol, cfg.Server.Domain, portSuffix)

	// Initialize session store
	sessionStore := auth.NewSessionStore(auth.SessionTTL)
	defer sessionStore.Stop()

	// Initialize rate limiter
	rateLimiter := auth.NewRateLimiter()

	// Initialize auth handlers with session store and rate limiter
	handlers.InitAuth(sessionStore, rateLimiter, config.Version)

	// Display auth status (v0.4.0: auth always required)
	fmt.Printf("  Authentication: ✓ Enabled (user: %s)\n", cfg.Auth.Username)
	fmt.Println()

	// Initialize audit logging
	if err := audit.Init(database.GetDB()); err != nil {
		log.Fatalf("Failed to initialize audit logging: %v", err)
	}

	// Initialize global write queue (must come before analytics)
	storage.InitWriter()

	// Initialize analytics buffer
	analytics.Init()

	// Initialize hosting system
	if err := hosting.Init(database.GetDB()); err != nil {
		log.Fatalf("Failed to initialize hosting: %v", err)
	}
	log.Printf("Hosting initialized (VFS Mode)")

	// Initialize serverless handler with storage support
	serverlessHandler = jsruntime.NewServerlessHandler(database.GetDB())

	// Generate mock data in development mode
	if cfg.IsDevelopment() {
		log.Println("Development mode: Checking for existing data...")
		// Only generate mock data if database is empty
		db := database.GetDB()
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
		if err == nil && count == 0 {
			log.Println("Database is empty, generating mock data...")
			if err := database.GenerateMockData(); err != nil {
				log.Printf("Warning: Failed to generate mock data: %v", err)
			}
		} else {
			log.Printf("Database already has %d events, skipping mock data generation", count)
		}
	}

	// Create dashboard router (existing dashboard functionality)
	dashboardMux := http.NewServeMux()

	// Authentication routes
	dashboardMux.HandleFunc("/api/login", handlers.LoginHandler)
	dashboardMux.HandleFunc("/api/logout", handlers.LogoutHandler)
	dashboardMux.HandleFunc("/api/auth/status", handlers.AuthStatusHandler)
	dashboardMux.HandleFunc("/api/user/me", handlers.UserMeHandler)

	// API routes - Tracking
	dashboardMux.HandleFunc("/track", handlers.TrackHandler)
	dashboardMux.HandleFunc("/pixel.gif", handlers.PixelHandler)
	dashboardMux.HandleFunc("/r/", handlers.RedirectHandler)
	dashboardMux.HandleFunc("/webhook/", handlers.WebhookHandler)

	// API routes - Dashboard
	dashboardMux.HandleFunc("/api/stats", handlers.StatsHandler)
	dashboardMux.HandleFunc("/api/events", handlers.EventsHandler)
	dashboardMux.HandleFunc("/api/redirects", handlers.RedirectsHandler)
	dashboardMux.HandleFunc("DELETE /api/redirects/{id}", handlers.DeleteRedirectHandler)
	dashboardMux.HandleFunc("/api/domains", handlers.DomainsHandler)
	dashboardMux.HandleFunc("/api/tags", handlers.TagsHandler)
	dashboardMux.HandleFunc("/api/webhooks", handlers.WebhooksHandler)
	dashboardMux.HandleFunc("DELETE /api/webhooks/{id}", handlers.DeleteWebhookHandler)
	dashboardMux.HandleFunc("PUT /api/webhooks/{id}", handlers.UpdateWebhookHandler)
	dashboardMux.HandleFunc("GET /api/system/limits", handlers.SystemLimitsHandler)
	dashboardMux.HandleFunc("GET /api/system/cache", handlers.SystemCacheHandler)
	dashboardMux.HandleFunc("GET /api/system/db", handlers.SystemDBHandler)
	dashboardMux.HandleFunc("GET /api/system/config", handlers.SystemConfigHandler)
	dashboardMux.HandleFunc("/api/config", handlers.SystemConfigHandler) // Alias
	dashboardMux.HandleFunc("GET /api/system/health", handlers.SystemHealthHandler)
	dashboardMux.HandleFunc("GET /api/system/capacity", handlers.SystemCapacityHandler)

	// API routes - Hosting/Deploy
	dashboardMux.HandleFunc("/api/deploy", handlers.DeployHandler)
	dashboardMux.HandleFunc("/api/sites", handlers.SitesHandler)
	dashboardMux.HandleFunc("GET /api/sites/{id}", handlers.SiteDetailHandler)
	dashboardMux.HandleFunc("GET /api/sites/{id}/files", handlers.SiteFilesHandler)
	dashboardMux.HandleFunc("GET /api/sites/{id}/files/{path...}", handlers.SiteFileContentHandler)

	// Apps API v2 (v0.10 - identity model)
	dashboardMux.HandleFunc("GET /api/apps", handlers.AppsListHandlerV2)
	dashboardMux.HandleFunc("POST /api/apps", handlers.AppCreateHandlerV2)
	dashboardMux.HandleFunc("POST /api/apps/install", handlers.AppInstallHandler)
	dashboardMux.HandleFunc("POST /api/apps/create", handlers.AppCreateHandler) // Legacy
	dashboardMux.HandleFunc("GET /api/templates", handlers.TemplatesListHandler)
	dashboardMux.HandleFunc("GET /api/apps/{id}", handlers.AppDetailHandlerV2)
	dashboardMux.HandleFunc("PUT /api/apps/{id}", handlers.AppUpdateHandlerV2)
	dashboardMux.HandleFunc("DELETE /api/apps/{id}", handlers.AppDeleteHandlerV2)
	dashboardMux.HandleFunc("GET /api/apps/{id}/files", handlers.AppFilesHandler)
	dashboardMux.HandleFunc("GET /api/apps/{id}/source", handlers.AppSourceHandler)
	dashboardMux.HandleFunc("GET /api/apps/{id}/files/{path...}", handlers.AppFileContentHandler)
	dashboardMux.HandleFunc("POST /api/apps/{id}/fork", handlers.AppForkHandler)
	dashboardMux.HandleFunc("GET /api/apps/{id}/lineage", handlers.AppLineageHandler)
	dashboardMux.HandleFunc("GET /api/apps/{id}/forks", handlers.AppForksHandler)

	// Aliases API (v0.10 - routing layer)
	dashboardMux.HandleFunc("GET /api/aliases", handlers.AliasesListHandler)
	dashboardMux.HandleFunc("POST /api/aliases", handlers.AliasCreateHandler)
	dashboardMux.HandleFunc("GET /api/aliases/{subdomain}", handlers.AliasDetailHandler)
	dashboardMux.HandleFunc("PUT /api/aliases/{subdomain}", handlers.AliasUpdateHandler)
	dashboardMux.HandleFunc("DELETE /api/aliases/{subdomain}", handlers.AliasDeleteHandler)
	dashboardMux.HandleFunc("POST /api/aliases/{subdomain}/reserve", handlers.AliasReserveHandler)
	dashboardMux.HandleFunc("POST /api/aliases/{subdomain}/split", handlers.AliasSplitHandler)
	dashboardMux.HandleFunc("POST /api/aliases/swap", handlers.AliasSwapHandler)

	// Command Gateway (v0.10 - for @peer remote execution)
	dashboardMux.HandleFunc("POST /api/cmd", handlers.CmdGatewayHandler)

	// Agent Endpoints (v0.10 - for LLM agent workflows)
	dashboardMux.HandleFunc("GET /_fazt/info", handlers.AgentInfoHandler)
	dashboardMux.HandleFunc("GET /_fazt/storage", handlers.AgentStorageListHandler)
	dashboardMux.HandleFunc("GET /_fazt/storage/{key}", handlers.AgentStorageGetHandler)
	dashboardMux.HandleFunc("POST /_fazt/snapshot", handlers.AgentSnapshotHandler)
	dashboardMux.HandleFunc("POST /_fazt/restore/{name}", handlers.AgentRestoreHandler)
	dashboardMux.HandleFunc("GET /_fazt/snapshots", handlers.AgentSnapshotsListHandler)
	dashboardMux.HandleFunc("GET /_fazt/logs", handlers.AgentLogsHandler)
	dashboardMux.HandleFunc("GET /_fazt/errors", handlers.AgentErrorsHandler)

	dashboardMux.HandleFunc("/api/keys", handlers.APIKeysHandler)
	dashboardMux.HandleFunc("/api/deployments", handlers.DeploymentsHandler)
	dashboardMux.HandleFunc("/api/envvars", handlers.EnvVarsHandler)
	dashboardMux.HandleFunc("/api/logs", handlers.LogsHandler)
	dashboardMux.HandleFunc("/api/logs/stream", handlers.LogStreamHandler)

	// System upgrade endpoint (requires API key auth)
	dashboardMux.HandleFunc("POST /api/upgrade", handlers.UpgradeHandler)

	// MCP (Model Context Protocol) routes
	mcpServer, err := mcp.NewServer()
	if err != nil {
		log.Printf("Warning: MCP server not initialized: %v", err)
	} else {
		dashboardMux.HandleFunc("POST /mcp/initialize", mcpServer.HandleInitialize)
		dashboardMux.HandleFunc("POST /mcp/tools/list", mcpServer.HandleToolsList)
		dashboardMux.HandleFunc("POST /mcp/tools/call", mcpServer.HandleToolsCall)
	}

	// Dashboard (Admin VFS Site)
	dashboardMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		siteHandler(w, r, "admin")
	})

	// Health check (available on both dashboard and sites)
	dashboardMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.HealthCheck(); err != nil {
			http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create the root handler with host-based routing
	rootHandler := createRootHandler(cfg, dashboardMux, sessionStore)

	// Apply middleware (order: tracing -> logging -> body limit -> security -> cors -> recovery -> root)
	handler := middleware.RequestTracing(
		loggingMiddleware(
			middleware.BodySizeLimit(middleware.MaxBodySize)(
				middleware.SecurityHeaders(
					corsMiddleware(
						recoveryMiddleware(rootHandler),
					),
				),
			),
		),
	)

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Write PID file for stop command
	pidFile := filepath.Join(filepath.Dir(cfg.Database.Path), "cc-server.pid")
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		log.Printf("Warning: Failed to write PID file: %v", err)
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on :%s", cfg.Server.Port)
		log.Printf("Dashboard: %s", cfg.Server.Domain)

		if cfg.HTTPS.Enabled {
			// Configure CertMagic
			log.Println("HTTPS Enabled: Using CertMagic")
			
			// Initialize SQL Storage
			certStorage := database.NewSQLCertStorage(database.GetDB())
			
			certmagic.DefaultACME.Email = cfg.HTTPS.Email
			if cfg.HTTPS.Staging {
				certmagic.DefaultACME.CA = certmagic.LetsEncryptStagingCA
			}
			
			// Use our SQL storage
			certmagic.Default.Storage = certStorage

			// Get allowed domains
			// We want to serve the main domain and any hosted subdomains.
			// CertMagic OnDemand allows serving any domain we have permission for.
			// But we need to restrict it to prevent abuse.
			// For personal PaaS, we might only allow *.domain.com and domain.com.
			
			// Configure OnDemand TLS
			cfgDomain := extractDomain(cfg.Server.Domain)
			certmagic.Default.OnDemand = &certmagic.OnDemandConfig{
				DecisionFunc: func(ctx context.Context, name string) error {
					// Allow main domain
					if name == cfgDomain {
						return nil
					}
					// Allow subdomains of main domain
					if strings.HasSuffix(name, "."+cfgDomain) {
						// Optionally check if site exists in DB?
						// if !hosting.SiteExists(extractSubdomain(name, cfgDomain)) { return fmt.Errorf("unknown site") }
						return nil
					}
					return fmt.Errorf("domain not allowed")
				},
			}

			// Start HTTPS server
			// certmagic.HTTPS blocks, so we wrap the handler
			// Note: CertMagic listens on :80 and :443 by default.
			// If cfg.Server.Port is not 443, this might be confusing.
			// CertMagic manages the listeners.
			
			err := certmagic.HTTPS([]string{cfgDomain}, handler)
			if err != nil {
				log.Fatalf("HTTPS Server failed: %v", err)
			}
		} else {
			// Standard HTTP
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server failed to start: %v", err)
			}
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Clean up PID file
	os.Remove(pidFile)

	// Flush analytics buffer
	analytics.Shutdown()

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// handleInstallCommand handles the install subcommand
func handleInstallCommand() {
	flags := flag.NewFlagSet("install", flag.ExitOnError)
	domain := flags.String("domain", "", "Domain for the server (required)")
	email := flags.String("email", "", "Email for Let's Encrypt (required for HTTPS)")
	user := flags.String("user", "fazt", "System user to run as")
	https := flags.Bool("https", false, "Enable automatic HTTPS")
	adminUser := flags.String("username", "admin", "Admin username")
	adminPass := flags.String("password", "", "Admin password (will generate if empty)")

	flags.Usage = func() {
		fmt.Println("Usage: fazt server install [flags]")
		fmt.Println()
		fmt.Println("Auto-installs fazt as a systemd service.")
		fmt.Println("Must be run with sudo.")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  sudo fazt server install --domain example.com --email admin@example.com --https")
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	if *domain == "" {
		fmt.Println("Error: --domain is required")
		flags.Usage()
		os.Exit(1)
	}

	if *https && *email == "" {
		fmt.Println("Error: --email is required when --https is enabled")
		flags.Usage()
		os.Exit(1)
	}

	// Generate password if empty
	if *adminPass == "" {
		// Simple random string generation
		*adminPass = "fazt-" + strconv.FormatInt(time.Now().UnixNano(), 36)
		fmt.Printf("Generated Admin Password: %s\n", *adminPass)
		fmt.Println("(Please save this password!)")
	}

	opts := provision.InstallOptions{
		User:          *user,
		Domain:        *domain,
		Email:         *email,
		AdminUser:     *adminUser,
		AdminPassword: *adminPass,
		HTTPS:         *https,
	}

	if err := provision.RunInstall(opts); err != nil {
		fmt.Printf("Installation failed: %v\n", err)
		os.Exit(1)
	}
}

// printUsage displays the usage information
func printUsage() {
	fmt.Printf("Fazt.sh %s - Personal Cloud Platform\n", config.Version)
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  fazt <command> [flags]")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  app        App management (list, deploy, info, remove)")
	fmt.Println("  remote     Peer management (add, list, status, upgrade)")
	fmt.Println("  service    System service (install, start, logs)")
	fmt.Println("  server     Server control (init, start, config)")
	fmt.Println("  version    Show version info")
	fmt.Println("  help       Show this message")
	fmt.Println()
	fmt.Println("QUICK START:")
	fmt.Println("  # Deploy an app to a peer")
	fmt.Println("  fazt app deploy ./my-site --to zyt")
	fmt.Println()
	fmt.Println("  # List apps on a peer")
	fmt.Println("  fazt app list zyt")
	fmt.Println()
	fmt.Println("  # Check peer status")
	fmt.Println("  fazt remote status zyt")
	fmt.Println()
}

// printServiceHelp displays service-specific help
func printServiceHelp() {
	fmt.Printf("fazt.sh %s - Service Commands\n", config.Version)
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  fazt service <command> [options]")
	fmt.Println()
	fmt.Println("SERVICE COMMANDS:")
	fmt.Println("  install          Auto-install as systemd service (requires sudo)")
	fmt.Println("  start            Start the system service")
	fmt.Println("  stop             Stop the system service")
	fmt.Println("  status           Check status of system service")
	fmt.Println("  logs             Follow service logs")
	fmt.Println("  --help, -h       Show this help")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Install as production service")
	fmt.Println("  sudo fazt service install --domain example.com --email admin@example.com --https")
	fmt.Println()
	fmt.Println("  # Check status")
	fmt.Println("  fazt service status")
	fmt.Println()
}

// printServerHelp displays server-specific help
func printServerHelp() {
	fmt.Printf("fazt.sh %s - Server Commands\n", config.Version)
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  fazt server <command> [options]")
	fmt.Println()
	fmt.Println("SERVER COMMANDS:")
	fmt.Println("  init             Initialize server (creates config & db)")
	fmt.Println("  start            Start the server manually (HTTP or HTTPS)")
	fmt.Println("  status           Show configuration and server status")
	fmt.Println("  set-credentials  Update admin credentials (password reset)")
	fmt.Println("  set-config       Update settings (domain, port, env)")
	fmt.Println("  create-key       Create an API key for deployments")
	fmt.Println("  reset-admin      Reset admin dashboard to embedded version")
	fmt.Println("  --help, -h       Show this help")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Initialize (required first step for manual run)")
	fmt.Println("  fazt server init --username admin --password secret --domain https://fazt.example.com")
	fmt.Println()
	fmt.Println("  # Start server manually (debugging)")
	fmt.Println("  fazt server start")
	fmt.Println()
	fmt.Println("  # Create API key for client deployment")
	fmt.Println("  fazt server create-key --name my-laptop")
	fmt.Println()
	fmt.Println("  # Reset Admin Password")
	fmt.Println("  fazt server set-credentials --username admin --password newsecret")
	fmt.Println()
}

// printClientHelp displays client-specific help
func printClientHelp() {
	fmt.Printf("Fazt.sh %s - Client Commands\n", config.Version)
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  fazt client <command> [options]")
	fmt.Println()
	fmt.Println("CLIENT COMMANDS:")
	fmt.Println("  set-auth-token   Set deployment token (generate at /hosting)")
	fmt.Println("  deploy           Deploy a site/app to the server")
	fmt.Println("  sites            List deployed sites")
	fmt.Println("  logs             View site logs")
	fmt.Println("  delete           Delete a site")
	fmt.Println("  --help, -h       Show this help")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Configure client (Get token from your dashboard)")
	fmt.Println("  fazt client set-auth-token --token <TOKEN>")
	fmt.Println()
	fmt.Println("  # Deploy static site")
	fmt.Println("  fazt client deploy --path . --domain my-site")
	fmt.Println()
	fmt.Println("  # View logs")
	fmt.Println("  fazt client logs --site my-site")
	fmt.Println()
}

// handleResetAdminCommand handles the reset-admin subcommand
func handleResetAdminCommand() {
	flags := flag.NewFlagSet("reset-admin", flag.ExitOnError)
	db := flags.String("db", "", "Database file path")

	flags.Usage = func() {
		fmt.Println("Usage: fazt server reset-admin [flags]")
		fmt.Println()
		fmt.Println("Reset the admin dashboard (VFS) to the version embedded in this binary.")
		fmt.Println("This is useful after upgrading fazt or if the dashboard is corrupted.")
		fmt.Println()
		flags.PrintDefaults()
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	// Resolve DB Path
	dbPath := "./data.db"
	if envPath := os.Getenv("FAZT_DB_PATH"); envPath != "" {
		dbPath = envPath
	}
	if *db != "" {
		dbPath = config.ExpandPath(*db)
	}

	// Initialize DB
	if err := database.Init(dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize Hosting (VFS)
	if err := hosting.Init(database.GetDB()); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init hosting: %v\n", err)
		os.Exit(1)
	}

	// Reset Admin Site
	if err := hosting.ResetAdminSite(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to reset admin site: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Admin dashboard reset successfully.")
}

// handleCreateKeyCommand creates a new API key for deployment.
// This is meant to be run on the server (e.g., via SSH) to generate tokens
// for clients to use when deploying.
func handleCreateKeyCommand() {
	flags := flag.NewFlagSet("create-key", flag.ExitOnError)
	name := flags.String("name", "", "Key name (required)")
	scopes := flags.String("scopes", "deploy", "Key scopes (default: deploy)")
	db := flags.String("db", "", "Database file path")

	flags.Usage = func() {
		fmt.Println("Usage: fazt server create-key --name <NAME> [flags]")
		fmt.Println()
		fmt.Println("Create a new API key for deploying apps to this server.")
		fmt.Println("Run this on your server, then use the token with 'fazt servers add' on your client.")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  fazt server create-key --name my-laptop")
		fmt.Println("  # Then on your laptop:")
		fmt.Println("  fazt servers add prod --url https://your-server.com --token <TOKEN>")
	}

	if err := flags.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	if *name == "" {
		fmt.Println("Error: --name is required")
		flags.Usage()
		os.Exit(1)
	}

	// Resolve DB Path (auto-detect from service if not specified)
	dbPath := provision.GetEffectiveDBPath(*db)
	if *db != "" {
		dbPath = config.ExpandPath(*db)
	}

	// Show which database we're using
	fmt.Printf("Using database: %s\n", dbPath)

	// Initialize DB
	if err := database.Init(dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Create API key
	token, err := hosting.CreateAPIKey(database.GetDB(), *name, *scopes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create API key: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("API Key created successfully!")
	fmt.Println()
	fmt.Printf("  Name:   %s\n", *name)
	fmt.Printf("  Scopes: %s\n", *scopes)
	fmt.Printf("  Token:  %s\n", token)
	fmt.Println()
	fmt.Println("Save this token - it won't be shown again!")
	fmt.Println()
	fmt.Println("To configure your client:")
	fmt.Printf("  fazt servers add <name> --url <YOUR_SERVER_URL> --token %s\n", token)
}
