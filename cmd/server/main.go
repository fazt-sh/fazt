package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
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

	"github.com/fazt-sh/fazt/internal/audit"
	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/config"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers"
	"github.com/fazt-sh/fazt/internal/hosting"
	"github.com/fazt-sh/fazt/internal/middleware"
	"github.com/fazt-sh/fazt/internal/provision"
	"github.com/fazt-sh/fazt/internal/security"
	"github.com/fazt-sh/fazt/internal/term"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
	"github.com/caddyserver/certmagic"
)

var Version = "dev"

var (
	showVersion = flag.Bool("version", false, "Show version and exit")
	showHelp    = flag.Bool("help", false, "Show help and exit")
	verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	quiet       = flag.Bool("quiet", false, "Quiet mode (errors only)")
)

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

	// Handle top-level subcommands
	switch command {
	case "server":
		handleServerCommand(os.Args[2:])
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

	// Set configs
	configs := map[string]string{
		"server.port":        port,
		"server.domain":      domain,
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

	// Update domain if provided
	if domain != "" {
		if err := store.Set("server.domain", domain); err != nil {
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
	case "--help", "-h", "help":
		printServerHelp()
	default:
		fmt.Printf("Unknown server command: %s\n\n", subcommand)
		printServerHelp()
		os.Exit(1)
	}
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

// printVersion displays version information
func printVersion() {
	fmt.Printf("fazt.sh %s\n", Version)
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
// Serves files from VFS
// If main.js exists, executes serverless JavaScript instead
// WebSocket connections at /ws are handled by the WebSocket hub
func siteHandler(w http.ResponseWriter, r *http.Request, subdomain string) {
	// Check if site exists
	if !hosting.SiteExists(subdomain) {
		serveSiteNotFound(w, r, subdomain)
		return
	}
	// Handle WebSocket connections at /ws
	if r.URL.Path == "/ws" {
		hosting.HandleWebSocket(w, r, subdomain)
		return
	}

	// Log analytics event for site visits
	logSiteVisit(r, subdomain)

	// Check for serverless (main.js)
	// We check directly in VFS now
	fs := hosting.GetFileSystem()
	hasServerless, _ := fs.Exists(subdomain, "main.js")

	if hasServerless {
		db := database.GetDB()
		if hosting.RunServerless(w, r, subdomain, db, subdomain) {
			return // Serverless handled the request
		}
	}

	// Serve from VFS
	hosting.ServeVFS(w, r, subdomain)
}

// logSiteVisit logs an analytics event for a site visit
func logSiteVisit(r *http.Request, subdomain string) {
	db := database.GetDB()
	if db == nil {
		return
	}

	// Insert event into database
	_, err := db.Exec(`
		INSERT INTO events (domain, source_type, event_type, path, referrer, user_agent, ip_address, query_params)
		VALUES (?, 'hosting', 'pageview', ?, ?, ?, ?, ?)
	`,
		subdomain,
		r.URL.Path,
		r.Referer(),
		r.UserAgent(),
		r.RemoteAddr,
		r.URL.RawQuery,
	)

	if err != nil {
		log.Printf("Failed to log site visit: %v", err)
	}
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

// createDeployZip creates a ZIP archive of the directory
func createDeployZip(dir string) (*bytes.Buffer, int, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	fileCount := 0

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
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
	server := flags.String("server", "", "fazt.sh server URL (default: http://localhost:4698 or from config)")

	flags.Usage = func() {
		fmt.Println("Usage: fazt client deploy --path <PATH> --domain <SUBDOMAIN>")
		fmt.Println()
		fmt.Println("Deploys a directory to a fazt.sh server.")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  fazt client deploy --path . --domain my-site")
		fmt.Println("  fazt client deploy --path ~/Desktop/site --domain example --server https://fazt.example.com")
		fmt.Println("  fazt client deploy --domain my-site --path .")
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

	// Load config to get API key
	flagsConfig := config.ParseFlags()
	cfg, err := config.Load(flagsConfig)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	token := cfg.GetAPIKey()
	if token == "" {
		fmt.Println("Error: No API key found in config")
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/api/logs?site_id=%s&limit=%d", *server, *site, *limit)
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

	// Load config to get API key
	flagsConfig := config.ParseFlags()
	cfg, err := config.Load(flagsConfig)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	token := cfg.GetAPIKey()
	if token == "" {
		fmt.Println("Error: No API key found in config")
		os.Exit(1)
	}

	req, err := http.NewRequest("GET", *server+"/api/sites", nil)
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
		Success bool `json:"success"`
		Sites   []struct {
			Name      string      `json:"Name"`
			FileCount int         `json:"FileCount"`
			SizeBytes int64       `json:"SizeBytes"`
			ModTime   interface{} `json:"ModTime"`
		} `json:"sites"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%-20s %-10s %-10s\n", "SITE", "FILES", "SIZE")
	fmt.Println("──────────────────────────────────────────")
	for _, site := range result.Sites {
		fmt.Printf("%-20s %-10d %-10d\n", site.Name, site.FileCount, site.SizeBytes)
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

	// Load config to get API key
	flagsConfig := config.ParseFlags()
	cfg, err := config.Load(flagsConfig)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	token := cfg.GetAPIKey()
	if token == "" {
		fmt.Println("Error: No API key found in config")
		os.Exit(1)
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/sites?site_id=%s", *server, *site), nil)
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

	if err := provision.Upgrade(Version); err != nil {
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
	if *domain != "" {
		cfg.Server.Domain = *domain
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

	// Ensure secure file permissions
	security.EnsureSecurePermissions(config.ExpandPath(cliFlags.ConfigPath), cfg.Database.Path)

	// Display startup information
	term.Banner()
	term.Section(fmt.Sprintf("fazt.sh %s - Starting Up", Version))

	fmt.Printf("  Environment:  %s\n", cfg.Server.Env)
	fmt.Printf("  Port:         %s\n", cfg.Server.Port)
	fmt.Printf("  Domain:       %s\n", cfg.Server.Domain)
	fmt.Printf("  Database:     %s\n", cfg.Database.Path)
	// fmt.Printf("  Config File:  %s\n", config.ExpandPath(cliFlags.ConfigPath)) // Deprecated display

	// Initialize session store
	sessionStore := auth.NewSessionStore(auth.SessionTTL)
	defer sessionStore.Stop()

	// Initialize rate limiter
	rateLimiter := auth.NewRateLimiter()

	// Initialize auth handlers with session store and rate limiter
	handlers.InitAuth(sessionStore, rateLimiter)

	// Display auth status (v0.4.0: auth always required)
	fmt.Printf("  Authentication: ✓ Enabled (user: %s)\n", cfg.Auth.Username)
	fmt.Println()

	// Initialize audit logging
	if err := audit.Init(database.GetDB()); err != nil {
		log.Fatalf("Failed to initialize audit logging: %v", err)
	}

	// Initialize hosting system
	if err := hosting.Init(database.GetDB()); err != nil {
		log.Fatalf("Failed to initialize hosting: %v", err)
	}
	log.Printf("Hosting initialized (VFS Mode)")

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
	dashboardMux.HandleFunc("/login", handlers.LoginPageHandler)
	dashboardMux.HandleFunc("/api/login", handlers.LoginHandler)
	dashboardMux.HandleFunc("/api/logout", handlers.LogoutHandler)
	dashboardMux.HandleFunc("/api/auth/status", handlers.AuthStatusHandler)

	// API routes - Tracking
	dashboardMux.HandleFunc("/track", handlers.TrackHandler)
	dashboardMux.HandleFunc("/pixel.gif", handlers.PixelHandler)
	dashboardMux.HandleFunc("/r/", handlers.RedirectHandler)
	dashboardMux.HandleFunc("/webhook/", handlers.WebhookHandler)

	// API routes - Dashboard
	dashboardMux.HandleFunc("/api/stats", handlers.StatsHandler)
	dashboardMux.HandleFunc("/api/events", handlers.EventsHandler)
	dashboardMux.HandleFunc("/api/redirects", handlers.RedirectsHandler)
	dashboardMux.HandleFunc("/api/domains", handlers.DomainsHandler)
	dashboardMux.HandleFunc("/api/tags", handlers.TagsHandler)
	dashboardMux.HandleFunc("/api/webhooks", handlers.WebhooksHandler)
	dashboardMux.HandleFunc("/api/config", handlers.ConfigHandler)

	// API routes - Hosting/Deploy
	dashboardMux.HandleFunc("/api/deploy", handlers.DeployHandler)
	dashboardMux.HandleFunc("/api/sites", handlers.SitesHandler)
	dashboardMux.HandleFunc("/api/keys", handlers.APIKeysHandler)
	dashboardMux.HandleFunc("/api/deployments", handlers.DeploymentsHandler)
	dashboardMux.HandleFunc("/api/envvars", handlers.EnvVarsHandler)
	dashboardMux.HandleFunc("/api/logs", handlers.LogsHandler)

	// Hosting management page
	dashboardMux.HandleFunc("/hosting", handlers.HostingPageHandler)

	// Static files
	fs := http.FileServer(http.Dir("./web/static"))
	dashboardMux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Dashboard (root)
	dashboardMux.HandleFunc("/", handlers.DashboardHandler)

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
	fmt.Printf("Fazt.sh %s - Personal Cloud Platform\n", Version)
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  fazt <command> [flags]")
	fmt.Println()
	fmt.Println("MODES:")
	fmt.Println("  service    System Service (install, start, logs)")
	fmt.Println("  client     Client Tool (deploy, logs, tokens)")
	fmt.Println("  server     Manual Control (config, reset-password)")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  deploy     Deploy current directory")
	fmt.Println("  version    Show version info")
	fmt.Println("  help       Show this message")
	fmt.Println()
	fmt.Println("QUICK START:")
	fmt.Println("  1. System Service (Hosting)")
	fmt.Println("     sudo fazt service install --domain example.com --email you@mail.com --https")
	fmt.Println()
	fmt.Println("  2. Client Tool (Deploying)")
	fmt.Println("     fazt client set-auth-token --token <YOUR_TOKEN>")
	fmt.Println("     fazt deploy --domain my-site")
	fmt.Println()
}

// printServiceHelp displays service-specific help
func printServiceHelp() {
	fmt.Printf("fazt.sh %s - Service Commands\n", Version)
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
	fmt.Printf("fazt.sh %s - Server Commands\n", Version)
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
	fmt.Println("  --help, -h       Show this help")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Initialize (required first step for manual run)")
	fmt.Println("  fazt server init --username admin --password secret --domain https://fazt.example.com")
	fmt.Println()
	fmt.Println("  # Start server manually (debugging)")
	fmt.Println("  fazt server start")
	fmt.Println()
	fmt.Println("  # Reset Admin Password")
	fmt.Println("  fazt server set-credentials --username admin --password newsecret")
	fmt.Println()
}

// printClientHelp displays client-specific help
func printClientHelp() {
	fmt.Printf("Fazt.sh %s - Client Commands\n", Version)
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
