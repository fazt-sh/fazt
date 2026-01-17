package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/git"
	"github.com/fazt-sh/fazt/internal/remote"
)

// handleAppCommand routes app subcommands
func handleAppCommand(args []string) {
	if len(args) < 1 {
		// Default: list apps on default peer
		handleAppList(nil)
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		handleAppList(args[1:])
	case "deploy":
		handleAppDeploy(args[1:])
	case "install":
		handleAppInstall(args[1:])
	case "upgrade":
		handleAppUpgrade(args[1:])
	case "pull":
		handleAppPull(args[1:])
	case "info":
		handleAppInfo(args[1:])
	case "remove":
		handleAppRemove(args[1:])
	case "--help", "-h", "help":
		printAppHelp()
	default:
		fmt.Printf("Unknown app command: %s\n\n", subcommand)
		printAppHelp()
		os.Exit(1)
	}
}

// handleAppList lists apps on a peer
func handleAppList(args []string) {
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
			fmt.Println("  fazt app list <peer>")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
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

// handleAppDeploy deploys a directory to a peer
func handleAppDeploy(args []string) {
	flags := flag.NewFlagSet("app deploy", flag.ExitOnError)
	siteName := flags.String("name", "", "App name (defaults to directory name)")
	peerFlag := flags.String("to", "", "Target peer name")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app deploy <directory> [--name <app>] [--to <peer>]")
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

	// Determine app name
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
		if err == remote.ErrNoPeers {
			fmt.Println("No peers configured.")
			fmt.Println("Run: fazt remote add <name> --url <url> --token <token>")
		} else if err == remote.ErrNoDefaultPeer {
			fmt.Println("Multiple peers configured. Specify target peer:")
			fmt.Println("  fazt app deploy <dir> --to <peer>")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
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

// handleAppInfo shows details about an app
func handleAppInfo(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: app name is required")
		fmt.Println("Usage: fazt app info <app> [peer]")
		os.Exit(1)
	}

	appName := args[0]
	var peerName string
	if len(args) > 1 && !strings.HasPrefix(args[1], "-") {
		peerName = args[1]
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

	// Find the app
	for _, app := range apps {
		if app.Name == appName {
			fmt.Printf("App:      %s\n", app.Name)
			fmt.Printf("Peer:     %s\n", peer.Name)
			fmt.Printf("Files:    %d\n", app.FileCount)
			fmt.Printf("Size:     %s\n", formatSize(app.SizeBytes))
			fmt.Printf("Updated:  %s\n", app.UpdatedAt)
			fmt.Printf("URL:      https://%s.%s\n", app.Name, strings.TrimPrefix(peer.URL, "https://admin."))
			return
		}
	}

	fmt.Printf("Error: app '%s' not found on %s\n", appName, peer.Name)
	os.Exit(1)
}

// handleAppRemove removes an app from a peer
func handleAppRemove(args []string) {
	flags := flag.NewFlagSet("app remove", flag.ExitOnError)
	peerFlag := flags.String("from", "", "Target peer name")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app remove <app> [--from <peer>]")
		fmt.Println()
		flags.PrintDefaults()
	}

	if len(args) < 1 {
		fmt.Println("Error: app name is required")
		flags.Usage()
		os.Exit(1)
	}

	appName := args[0]
	flags.Parse(args[1:])

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, *peerFlag)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	client := remote.NewClient(peer)
	err = client.DeleteApp(appName)
	if err != nil {
		fmt.Printf("Error removing app: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Removed '%s' from %s\n", appName, peer.Name)
}

// handleAppInstall installs an app from a git repository
func handleAppInstall(args []string) {
	flags := flag.NewFlagSet("app install", flag.ExitOnError)
	peerFlag := flags.String("to", "", "Target peer name")
	nameFlag := flags.String("name", "", "Override app name")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app install <url> [--to <peer>] [--name <name>]")
		fmt.Println()
		fmt.Println("Install an app from a git repository.")
		fmt.Println()
		fmt.Println("URL formats:")
		fmt.Println("  github.com/user/repo")
		fmt.Println("  github.com/user/repo/path/to/app")
		fmt.Println("  github.com/user/repo@v1.0.0")
		fmt.Println("  github:user/repo (shorthand)")
		fmt.Println()
		flags.PrintDefaults()
	}

	// Find URL arg (first non-flag arg)
	var url string
	var flagArgs []string
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") && url == "" {
			url = arg
			flagArgs = args[i+1:]
			break
		}
	}

	if url == "" {
		fmt.Println("Error: git URL is required")
		flags.Usage()
		os.Exit(1)
	}

	flags.Parse(flagArgs)

	// Parse git URL
	ref, err := git.ParseURL(url)
	if err != nil {
		fmt.Printf("Error: invalid URL: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Installing from %s...\n", ref.String())

	// Clone to temp directory
	tmpDir, err := os.MkdirTemp("", "fazt-install-*")
	if err != nil {
		fmt.Printf("Error creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	result, err := git.Clone(git.CloneOptions{
		URL:       ref.FullURL(),
		Path:      ref.Path,
		Ref:       ref.Ref,
		TargetDir: tmpDir,
	})
	if err != nil {
		fmt.Printf("Error cloning: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Cloned %d files (commit: %s)\n", result.Files, result.CommitSHA[:7])

	// Read manifest to get app name
	appName := *nameFlag
	if appName == "" {
		manifest, err := readManifest(tmpDir)
		if err != nil {
			// Use repo name as fallback
			appName = ref.Repo
			if ref.Path != "" {
				appName = filepath.Base(ref.Path)
			}
		} else {
			appName = manifest.Name
		}
	}

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, *peerFlag)
	if err != nil {
		if err == remote.ErrNoPeers {
			fmt.Println("No peers configured.")
			fmt.Println("Run: fazt remote add <name> --url <url> --token <token>")
		} else if err == remote.ErrNoDefaultPeer {
			fmt.Println("Multiple peers configured. Specify target peer:")
			fmt.Println("  fazt app install <url> --to <peer>")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("Deploying '%s' to %s...\n", appName, peer.Name)

	// Create ZIP from cloned files
	zipBuffer, fileCount, err := createDeployZip(tmpDir)
	if err != nil {
		fmt.Printf("Error creating ZIP: %v\n", err)
		os.Exit(1)
	}

	// Write to temp file
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

	// Deploy with source tracking
	client := remote.NewClient(peer)
	deployResult, err := client.DeployWithSource(tmpFile.Name(), appName, &remote.SourceInfo{
		Type:   "git",
		URL:    ref.String(),
		Ref:    ref.Ref,
		Commit: result.CommitSHA,
	})
	if err != nil {
		fmt.Printf("Error deploying: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("Installed: %s\n", deployResult.Site)
	fmt.Printf("Source:    %s\n", ref.String())
	fmt.Printf("Commit:    %s\n", result.CommitSHA[:7])
	fmt.Printf("Files:     %d\n", fileCount)
}

// handleAppUpgrade upgrades a git-sourced app to the latest version
func handleAppUpgrade(args []string) {
	flags := flag.NewFlagSet("app upgrade", flag.ExitOnError)
	peerFlag := flags.String("from", "", "Target peer name")
	checkOnly := flags.Bool("check", false, "Only check for updates, don't install")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app upgrade <app> [--from <peer>] [--check]")
		fmt.Println()
		flags.PrintDefaults()
	}

	if len(args) < 1 {
		fmt.Println("Error: app name is required")
		flags.Usage()
		os.Exit(1)
	}

	appName := args[0]
	flags.Parse(args[1:])

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, *peerFlag)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Get app source info from peer
	client := remote.NewClient(peer)
	sourceInfo, err := client.GetAppSource(appName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if sourceInfo.Type != "git" {
		fmt.Printf("Error: '%s' is not installed from git (source: %s)\n", appName, sourceInfo.Type)
		fmt.Println("Only git-sourced apps can be upgraded. Use 'fazt app deploy' for manual updates.")
		os.Exit(1)
	}

	fmt.Printf("Checking for updates to %s...\n", appName)
	fmt.Printf("Source: %s\n", sourceInfo.URL)
	fmt.Printf("Current commit: %s\n", sourceInfo.Commit[:7])

	// Parse the source URL
	ref, err := git.ParseURL(sourceInfo.URL)
	if err != nil {
		fmt.Printf("Error parsing source URL: %v\n", err)
		os.Exit(1)
	}

	// Get latest commit
	latest, err := git.GetLatestCommit(ref.FullURL(), ref.Ref)
	if err != nil {
		fmt.Printf("Error checking for updates: %v\n", err)
		os.Exit(1)
	}

	if latest == sourceInfo.Commit {
		fmt.Println("\nAlready up to date.")
		return
	}

	fmt.Printf("Update available: %s → %s\n", sourceInfo.Commit[:7], latest[:7])

	if *checkOnly {
		fmt.Println("\nRun without --check to install the update.")
		return
	}

	// Reinstall with same URL
	fmt.Println("\nUpgrading...")
	handleAppInstall([]string{sourceInfo.URL, "--to", peer.Name, "--name", appName})
}

// handleAppPull downloads an app's files to a local directory
func handleAppPull(args []string) {
	flags := flag.NewFlagSet("app pull", flag.ExitOnError)
	peerFlag := flags.String("from", "", "Source peer name")
	targetFlag := flags.String("to", "", "Target directory (defaults to ./<app>)")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app pull <app> [--from <peer>] [--to <dir>]")
		fmt.Println()
		flags.PrintDefaults()
	}

	if len(args) < 1 {
		fmt.Println("Error: app name is required")
		flags.Usage()
		os.Exit(1)
	}

	appName := args[0]
	flags.Parse(args[1:])

	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, *peerFlag)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Determine target directory
	targetDir := *targetFlag
	if targetDir == "" {
		targetDir = "./" + appName
	}

	// Get app files from peer
	client := remote.NewClient(peer)
	files, err := client.GetAppFiles(appName)
	if err != nil {
		fmt.Printf("Error fetching files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Printf("Error: app '%s' not found or has no files\n", appName)
		os.Exit(1)
	}

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Download and write each file
	fmt.Printf("Pulling %d files from %s to %s...\n", len(files), peer.Name, targetDir)

	for _, f := range files {
		content, err := client.GetAppFileContent(appName, f.Path)
		if err != nil {
			fmt.Printf("  Error fetching %s: %v\n", f.Path, err)
			continue
		}

		targetPath := filepath.Join(targetDir, f.Path)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			fmt.Printf("  Error creating dir for %s: %v\n", f.Path, err)
			continue
		}

		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			fmt.Printf("  Error writing %s: %v\n", f.Path, err)
			continue
		}
	}

	fmt.Printf("\nPulled %d files to %s\n", len(files), targetDir)
}

// Manifest represents an app manifest.json
type Manifest struct {
	Name string `json:"name"`
}

// readManifest reads and parses a manifest.json file
func readManifest(dir string) (*Manifest, error) {
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	if m.Name == "" {
		return nil, fmt.Errorf("manifest missing 'name' field")
	}

	return &m, nil
}

func printAppHelp() {
	fmt.Println(`Fazt.sh - App Management

USAGE:
  fazt app <command> [options]

COMMANDS:
  list [peer]        List apps on a peer
  deploy <dir>       Deploy directory to peer
  install <url>      Install app from git repository
  upgrade <app>      Upgrade git-sourced app
  pull <app>         Download app files to local directory
  info <app> [peer]  Show app details
  remove <app>       Remove an app from peer

OPTIONS:
  --to <peer>        Target peer for deploy/install
  --from <peer>      Source peer for pull/remove/upgrade
  --name <name>      Override app name
  --check            Check for updates only (upgrade)

EXAMPLES:
  # List apps
  fazt app list zyt

  # Deploy local directory
  fazt app deploy ./my-site --to zyt

  # Install from GitHub
  fazt app install github.com/user/repo --to zyt
  fazt app install github:user/repo/apps/blog@v1.0.0

  # Check for updates
  fazt app upgrade myapp --check

  # Download app to local folder
  fazt app pull myapp --to ./local-copy

  # Remove an app
  fazt app remove myapp --from zyt`)
}
