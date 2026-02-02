package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fazt-sh/fazt/internal/auth"
	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/remote"
)

// handleAuthCommand handles auth-related subcommands
func handleAuthCommand(args []string) {
	if len(args) < 1 {
		printAuthHelp()
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "provider":
		handleAuthProvider(args[1:])
	case "providers":
		handleAuthProviders()
	case "users":
		handleAuthUsers()
	case "user":
		handleAuthUser(args[1:])
	case "invite":
		handleAuthInvite(args[1:])
	case "invites":
		handleAuthInvites()
	case "--help", "-h", "help":
		printAuthHelp()
	default:
		fmt.Printf("Unknown auth command: %s\n\n", subcommand)
		printAuthHelp()
		os.Exit(1)
	}
}

// handleAuthProvider manages OAuth provider configuration
func handleAuthProvider(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: provider name is required")
		fmt.Println("Usage: fazt auth provider <name> [options]")
		fmt.Println("Providers: google, github, discord, microsoft")
		os.Exit(1)
	}

	providerName := strings.ToLower(args[0])

	// Validate provider name
	if _, ok := auth.Providers[providerName]; !ok {
		fmt.Printf("Unknown provider: %s\n", providerName)
		fmt.Println("Available providers: google, github, discord, microsoft")
		os.Exit(1)
	}

	flags := flag.NewFlagSet("auth provider", flag.ExitOnError)
	clientID := flags.String("client-id", "", "OAuth client ID")
	clientSecret := flags.String("client-secret", "", "OAuth client secret")
	enable := flags.Bool("enable", false, "Enable the provider")
	disable := flags.Bool("disable", false, "Disable the provider")
	dbPath := flags.String("db", getDefaultDBPath(), "Database path")
	flags.Parse(args[1:])

	// Initialize database
	if err := database.Init(*dbPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	service := auth.NewService(database.GetDB(), "", false)

	// If client ID and secret provided, configure the provider
	if *clientID != "" && *clientSecret != "" {
		if err := service.SetProviderConfig(providerName, *clientID, *clientSecret); err != nil {
			fmt.Printf("Error configuring provider: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Provider '%s' configured.\n", providerName)
	}

	// Handle enable/disable
	if *enable {
		if err := service.EnableProvider(providerName); err != nil {
			fmt.Printf("Error enabling provider: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Provider '%s' enabled.\n", providerName)
	} else if *disable {
		if err := service.DisableProvider(providerName); err != nil {
			fmt.Printf("Error disabling provider: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Provider '%s' disabled.\n", providerName)
	}

	// Show current status
	cfg, err := service.GetProviderConfig(providerName)
	if err == auth.ErrProviderDisabled {
		fmt.Printf("\nProvider '%s': not configured\n", providerName)
	} else if err != nil {
		fmt.Printf("\nError getting provider status: %v\n", err)
	} else {
		status := "disabled"
		if cfg.Enabled {
			status = "enabled"
		}
		clientIDDisplay := cfg.ClientID
		if len(clientIDDisplay) > 20 {
			clientIDDisplay = clientIDDisplay[:17] + "..."
		}
		fmt.Printf("\nProvider '%s':\n", providerName)
		fmt.Printf("  Status:    %s\n", status)
		fmt.Printf("  Client ID: %s\n", clientIDDisplay)
	}
}

// handleAuthProviders lists all configured providers
func handleAuthProviders() {
	dbPath := getDefaultDBPath()

	if err := database.Init(dbPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	service := auth.NewService(database.GetDB(), "", false)

	providers, err := service.ListProviders()
	if err != nil {
		fmt.Printf("Error listing providers: %v\n", err)
		os.Exit(1)
	}

	if len(providers) == 0 {
		fmt.Println("No providers configured.")
		fmt.Println("Run: fazt auth provider <name> --client-id <id> --client-secret <secret>")
		return
	}

	fmt.Printf("%-12s %-10s %-40s\n", "PROVIDER", "STATUS", "CLIENT ID")
	fmt.Println(strings.Repeat("-", 65))

	for _, cfg := range providers {
		status := "disabled"
		if cfg.Enabled {
			status = "enabled"
		}
		clientID := cfg.ClientID
		if len(clientID) > 38 {
			clientID = clientID[:35] + "..."
		}
		fmt.Printf("%-12s %-10s %-40s\n", cfg.Name, status, clientID)
	}
}

// handleAuthUsers lists all users
func handleAuthUsers() {
	dbPath := getDefaultDBPath()

	if err := database.Init(dbPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	service := auth.NewService(database.GetDB(), "", false)

	users, err := service.ListUsers()
	if err != nil {
		fmt.Printf("Error listing users: %v\n", err)
		os.Exit(1)
	}

	if len(users) == 0 {
		fmt.Println("No users found.")
		return
	}

	fmt.Printf("%-36s %-30s %-10s %-10s\n", "ID", "EMAIL", "ROLE", "PROVIDER")
	fmt.Println(strings.Repeat("-", 90))

	for _, u := range users {
		email := u.Email
		if len(email) > 28 {
			email = email[:25] + "..."
		}
		fmt.Printf("%-36s %-30s %-10s %-10s\n", u.ID, email, u.Role, u.Provider)
	}
}

// handleAuthUser shows/modifies a specific user
func handleAuthUser(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: user ID is required")
		fmt.Println("Usage: fazt auth user <id> [options]")
		os.Exit(1)
	}

	userID := args[0]

	flags := flag.NewFlagSet("auth user", flag.ExitOnError)
	role := flags.String("role", "", "Set user role (owner, admin, user)")
	del := flags.Bool("delete", false, "Delete the user")
	dbPath := flags.String("db", getDefaultDBPath(), "Database path")
	flags.Parse(args[1:])

	if err := database.Init(*dbPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	service := auth.NewService(database.GetDB(), "", false)

	// Get user first
	user, err := service.GetUserByID(userID)
	if err == auth.ErrUserNotFound {
		// Try by email
		user, err = service.GetUserByEmail(userID)
	}
	if err != nil {
		fmt.Printf("User not found: %s\n", userID)
		os.Exit(1)
	}

	// Handle delete
	if *del {
		if user.IsOwner() {
			fmt.Println("Error: cannot delete the owner")
			os.Exit(1)
		}
		if err := service.DeleteUser(user.ID); err != nil {
			fmt.Printf("Error deleting user: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("User '%s' deleted.\n", user.Email)
		return
	}

	// Handle role change
	if *role != "" {
		if user.IsOwner() && *role != "owner" {
			fmt.Println("Error: cannot demote the owner")
			os.Exit(1)
		}
		if err := service.UpdateUserRole(user.ID, *role); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("User '%s' role updated to '%s'.\n", user.Email, *role)
		user.Role = *role
	}

	// Show user details
	fmt.Printf("\nUser: %s\n", user.Email)
	fmt.Printf("  ID:        %s\n", user.ID)
	fmt.Printf("  Name:      %s\n", user.Name)
	fmt.Printf("  Role:      %s\n", user.Role)
	fmt.Printf("  Provider:  %s\n", user.Provider)
	fmt.Printf("  Created:   %s\n", time.Unix(user.CreatedAt, 0).Format(time.RFC3339))
	if user.LastLogin != nil {
		fmt.Printf("  Last Login: %s\n", time.Unix(*user.LastLogin, 0).Format(time.RFC3339))
	}
}

// handleAuthInvite creates or shows an invite
func handleAuthInvite(args []string) {
	flags := flag.NewFlagSet("auth invite", flag.ExitOnError)
	role := flags.String("role", "user", "Role for invited user")
	maxUses := flags.Int("max-uses", 1, "Maximum number of uses (0 = unlimited)")
	expiryDays := flags.Int("expiry", 7, "Days until expiry (0 = no expiry)")
	dbPath := flags.String("db", getDefaultDBPath(), "Database path")
	flags.Parse(args)

	if err := database.Init(*dbPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	service := auth.NewService(database.GetDB(), "", false)

	var expiry *time.Duration
	if *expiryDays > 0 {
		d := time.Duration(*expiryDays) * 24 * time.Hour
		expiry = &d
	}

	invite, err := service.CreateInvite(*role, "owner", *maxUses, expiry)
	if err != nil {
		fmt.Printf("Error creating invite: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nInvite code created!\n")
	fmt.Printf("  Code:     %s\n", invite.Code)
	fmt.Printf("  Role:     %s\n", invite.Role)
	fmt.Printf("  Max Uses: %d\n", invite.MaxUses)
	if invite.ExpiresAt != nil {
		fmt.Printf("  Expires:  %s\n", time.Unix(*invite.ExpiresAt, 0).Format(time.RFC3339))
	}
	fmt.Printf("\nShare this URL with the invitee:\n")
	fmt.Printf("  /auth/invite/%s\n", invite.Code)
}

// handleAuthInvites lists all invites
func handleAuthInvites() {
	dbPath := getDefaultDBPath()

	if err := database.Init(dbPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	service := auth.NewService(database.GetDB(), "", false)

	invites, err := service.ListInvites()
	if err != nil {
		fmt.Printf("Error listing invites: %v\n", err)
		os.Exit(1)
	}

	if len(invites) == 0 {
		fmt.Println("No invites found.")
		fmt.Println("Run: fazt auth invite --role user")
		return
	}

	fmt.Printf("%-10s %-8s %-6s %-6s %-10s\n", "CODE", "ROLE", "USES", "MAX", "STATUS")
	fmt.Println(strings.Repeat("-", 50))

	now := time.Now().Unix()
	for _, inv := range invites {
		status := "active"
		if inv.MaxUses > 0 && inv.UseCount >= inv.MaxUses {
			status = "used"
		} else if inv.ExpiresAt != nil && now > *inv.ExpiresAt {
			status = "expired"
		}
		fmt.Printf("%-10s %-8s %-6d %-6d %-10s\n", inv.Code, inv.Role, inv.UseCount, inv.MaxUses, status)
	}
}

func getDefaultDBPath() string {
	// Check for explicit path first
	if dbPath := os.Getenv("FAZT_DB"); dbPath != "" {
		return dbPath
	}

	// Try default server location
	home, _ := os.UserHomeDir()
	defaultPath := home + "/.config/fazt/data.db"
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}

	// Fallback to current directory
	return "data.db"
}

func printAuthHelp() {
	fmt.Println(`Fazt - Multi-user Authentication

USAGE:
  fazt auth <command> [options]

COMMANDS:
  provider <name>  Configure an OAuth provider
  providers        List configured providers
  users            List all users
  user <id>        Show/modify a user
  invite           Create an invite code
  invites          List all invites

PROVIDER SETUP:
  # Configure Google OAuth
  fazt auth provider google \
    --client-id "xxx.apps.googleusercontent.com" \
    --client-secret "GOCSPX-xxx"

  # Enable the provider
  fazt auth provider google --enable

  # List configured providers
  fazt auth providers

USER MANAGEMENT:
  # List all users
  fazt auth users

  # Show user details
  fazt auth user <id>

  # Change user role
  fazt auth user <id> --role admin

  # Delete a user
  fazt auth user <id> --delete

INVITES:
  # Create an invite code
  fazt auth invite --role user

  # Create admin invite (7 day expiry)
  fazt auth invite --role admin --expiry 7

  # List invites
  fazt auth invites

SUPPORTED PROVIDERS:
  google     Google OAuth 2.0
  github     GitHub OAuth
  discord    Discord OAuth
  microsoft  Microsoft (Personal accounts)`)
}

// handleAuthCommandWithPeer handles auth commands for a remote peer via API
func handleAuthCommandWithPeer(peerName string, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: auth subcommand required")
		fmt.Println("Usage: fazt @<peer> auth <command> [options]")
		fmt.Println("Commands: provider, providers")
		os.Exit(1)
	}

	// Load peer configuration
	db := getClientDB()
	defer database.Close()

	peer, err := remote.ResolvePeer(db, peerName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	client := remote.NewClient(peer)
	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "provider":
		handlePeerAuthProvider(client, subArgs)
	case "providers":
		handlePeerAuthProviders(client)
	default:
		fmt.Printf("Error: auth command '%s' cannot be executed remotely\n", subcommand)
		fmt.Println("Remote commands: provider, providers")
		os.Exit(1)
	}
}

// handleRemoteAuthProvider configures a provider on a remote peer
func handlePeerAuthProvider(client *remote.Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: provider name is required")
		fmt.Println("Usage: fazt @<peer> auth provider <name> [options]")
		fmt.Println("Providers: google, github, discord, microsoft")
		os.Exit(1)
	}

	providerName := strings.ToLower(args[0])

	// Validate provider name
	if _, ok := auth.Providers[providerName]; !ok {
		fmt.Printf("Unknown provider: %s\n", providerName)
		fmt.Println("Available providers: google, github, discord, microsoft")
		os.Exit(1)
	}

	flags := flag.NewFlagSet("auth provider", flag.ExitOnError)
	clientID := flags.String("client-id", "", "OAuth client ID")
	clientSecret := flags.String("client-secret", "", "OAuth client secret")
	enable := flags.Bool("enable", false, "Enable the provider")
	disable := flags.Bool("disable", false, "Disable the provider")
	flags.Parse(args[1:])

	// Build enable flag
	var enablePtr *bool
	if *enable {
		t := true
		enablePtr = &t
	} else if *disable {
		f := false
		enablePtr = &f
	}

	// Call remote API
	cfg, err := client.ConfigureAuthProvider(providerName, *clientID, *clientSecret, enablePtr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Show results
	if *clientID != "" && *clientSecret != "" {
		fmt.Printf("Provider '%s' configured.\n", providerName)
	}
	if *enable {
		fmt.Printf("Provider '%s' enabled.\n", providerName)
	} else if *disable {
		fmt.Printf("Provider '%s' disabled.\n", providerName)
	}

	// Show current status
	if !cfg.Configured {
		fmt.Printf("\nProvider '%s': not configured\n", providerName)
	} else {
		status := "disabled"
		if cfg.Enabled {
			status = "enabled"
		}
		fmt.Printf("\nProvider '%s':\n", providerName)
		fmt.Printf("  Status:    %s\n", status)
	}
}

// handleRemoteAuthProviders lists providers on a remote peer
func handlePeerAuthProviders(client *remote.Client) {
	providers, err := client.ListAuthProviders()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(providers) == 0 {
		fmt.Println("No providers configured.")
		fmt.Println("Run: fazt @<peer> auth provider <name> --client-id <id> --client-secret <secret>")
		return
	}

	fmt.Printf("%-12s %-10s %-40s\n", "PROVIDER", "STATUS", "CLIENT ID")
	fmt.Println(strings.Repeat("-", 65))
	for _, cfg := range providers {
		status := "disabled"
		if cfg.Enabled {
			status = "enabled"
		}
		fmt.Printf("%-12s %-10s %-40s\n", cfg.Name, status, cfg.ClientID)
	}
}
