package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/egress"
)

func handleSecretCommand(args []string) {
	if len(args) < 1 {
		printSecretUsage()
		return
	}

	switch args[0] {
	case "set":
		handleSecretSet(args[1:])
	case "list":
		handleSecretList(args[1:])
	case "remove":
		handleSecretRemove(args[1:])
	case "--help", "-h", "help":
		printSecretUsage()
	default:
		fmt.Printf("Unknown secret subcommand: %s\n", args[0])
		printSecretUsage()
		os.Exit(1)
	}
}

func printSecretUsage() {
	fmt.Println("fazt secret - Secret management for outbound HTTP auth")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  fazt secret <command> [options]")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  set <name> <value>      Create or update a secret")
	fmt.Println("  list                    List secrets (values masked)")
	fmt.Println("  remove <name>           Remove a secret")
	fmt.Println()
	fmt.Println("OPTIONS (set):")
	fmt.Println("  --as <type>             Injection type: bearer (default), header, query")
	fmt.Println("  --key <name>            Header name or query param (required for header/query)")
	fmt.Println("  --domain <domain>       Only inject for this domain")
	fmt.Println("  --app <id>              Scope to specific app")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  fazt secret set STRIPE_KEY sk_live_xxx")
	fmt.Println("  fazt secret set OPENAI_KEY sk-xxx --as header --key Authorization")
	fmt.Println("  fazt secret set TOKEN abc --as query --key token --app myapp")
	fmt.Println("  fazt secret list")
	fmt.Println("  fazt secret remove STRIPE_KEY")
}

func handleSecretSet(args []string) {
	fs := flag.NewFlagSet("secret set", flag.ExitOnError)
	asFlag := fs.String("as", "bearer", "Injection type: bearer, header, query")
	keyFlag := fs.String("key", "", "Header name or query param")
	domainFlag := fs.String("domain", "", "Only inject for this domain")
	appFlag := fs.String("app", "", "Scope to app ID")
	fs.Parse(args)

	if fs.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "Error: name and value required")
		fmt.Fprintln(os.Stderr, "Usage: fazt secret set <name> <value> [options]")
		os.Exit(1)
	}

	name := fs.Arg(0)
	value := fs.Arg(1)

	db := getClientDB()
	defer database.Close()

	store := egress.NewSecretsStore(db)
	if err := store.Set(name, value, *asFlag, *keyFlag, *domainFlag, *appFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	scope := "global"
	if *appFlag != "" {
		scope = fmt.Sprintf("app:%s", *appFlag)
	}
	fmt.Printf("Secret %s set (%s, %s)\n", name, *asFlag, scope)
}

func handleSecretList(args []string) {
	fs := flag.NewFlagSet("secret list", flag.ExitOnError)
	appFlag := fs.String("app", "", "Filter by app ID")
	fs.Parse(args)

	db := getClientDB()
	defer database.Close()

	store := egress.NewSecretsStore(db)
	secrets, err := store.List(*appFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(secrets) == 0 {
		fmt.Println("No secrets configured")
		fmt.Println()
		fmt.Println("Add secrets with: fazt secret set <name> <value>")
		return
	}

	fmt.Printf("%-20s %-30s %-10s %-10s %-15s\n", "Name", "Value", "Type", "Scope", "Domain")
	fmt.Println(strings.Repeat("-", 90))

	for _, s := range secrets {
		scope := "global"
		if s.AppID != "" {
			scope = s.AppID
		}
		domain := "*"
		if s.Domain != "" {
			domain = s.Domain
		}
		masked := egress.MaskValue(s.Value)
		fmt.Printf("%-20s %-30s %-10s %-10s %-15s\n", s.Name, masked, s.InjectAs, scope, domain)
	}

	fmt.Printf("\n%d secret(s)\n", len(secrets))
}

func handleSecretRemove(args []string) {
	fs := flag.NewFlagSet("secret remove", flag.ExitOnError)
	appFlag := fs.String("app", "", "Scope to app ID")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: name required")
		fmt.Fprintln(os.Stderr, "Usage: fazt secret remove <name> [--app <id>]")
		os.Exit(1)
	}

	name := fs.Arg(0)

	db := getClientDB()
	defer database.Close()

	store := egress.NewSecretsStore(db)
	if err := store.Remove(name, *appFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Secret %s removed\n", name)
}
