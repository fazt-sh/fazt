package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/fazt-sh/fazt/internal/assets"
)

// handleAppCreate creates a new app from a template
func handleAppCreate(args []string) {
	// Guard: this is a local-only command
	if targetPeerName != "" {
		fmt.Fprintf(os.Stderr, "Error: 'app create' is a local operation\n")
		fmt.Fprintf(os.Stderr, "This command creates local files, not apps on remote peers.\n")
		fmt.Fprintf(os.Stderr, "Usage: fazt app create <name> [--template <template>]\n")
		os.Exit(1)
	}

	flags := flag.NewFlagSet("app create", flag.ExitOnError)
	templateName := flags.String("template", "minimal", "Template to use (minimal, vite)")
	listTemplates := flags.Bool("list-templates", false, "List available templates")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app create <name> [--template <template>]")
		fmt.Println()
		fmt.Println("Create a new app from a template.")
		fmt.Println()
		flags.PrintDefaults()
		fmt.Println()
		fmt.Println("Templates:")
		for _, t := range assets.ListTemplates() {
			fmt.Printf("  - %s\n", t)
		}
	}

	// Check for --list-templates before parsing
	for _, arg := range args {
		if arg == "--list-templates" || arg == "-list-templates" {
			fmt.Println("Available templates:")
			for _, t := range assets.ListTemplates() {
				fmt.Printf("  - %s\n", t)
			}
			return
		}
	}

	// Find name arg (first non-flag arg)
	var appName string
	var flagArgs []string
	for i, arg := range args {
		if arg[0] != '-' && appName == "" {
			appName = arg
			flagArgs = args[i+1:]
			break
		}
	}

	flags.Parse(flagArgs)

	if *listTemplates {
		fmt.Println("Available templates:")
		for _, t := range assets.ListTemplates() {
			fmt.Printf("  - %s\n", t)
		}
		return
	}

	if appName == "" {
		fmt.Println("Error: app name is required")
		flags.Usage()
		os.Exit(1)
	}

	// Validate app name
	if !isValidAppName(appName) {
		fmt.Printf("Error: invalid app name '%s'\n", appName)
		fmt.Println("App names must:")
		fmt.Println("  - Use only lowercase letters, numbers, and hyphens")
		fmt.Println("  - Not start or end with a hyphen")
		fmt.Println("  - Be 1-63 characters long")
		os.Exit(1)
	}

	// Get template
	tmplFS, err := assets.GetTemplate(*templateName)
	if err != nil {
		fmt.Printf("Error: template '%s' not found\n", *templateName)
		fmt.Println("Available templates:")
		for _, t := range assets.ListTemplates() {
			fmt.Printf("  - %s\n", t)
		}
		os.Exit(1)
	}

	// Check if directory exists
	if _, err := os.Stat(appName); err == nil {
		fmt.Printf("Error: directory '%s' already exists\n", appName)
		os.Exit(1)
	}

	// Template data
	data := map[string]string{
		"Name": appName,
	}

	// Copy template files with substitution
	err = fs.WalkDir(tmplFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip root directory
		if path == "." {
			return nil
		}

		destPath := filepath.Join(appName, path)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		content, err := fs.ReadFile(tmplFS, path)
		if err != nil {
			return err
		}

		// Apply template substitution
		tmpl, err := template.New(path).Parse(string(content))
		if err != nil {
			// Not a valid template, copy as-is
			return writeFile(destPath, content)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			// Template execution failed, copy as-is
			return writeFile(destPath, content)
		}

		return writeFile(destPath, buf.Bytes())
	})

	if err != nil {
		fmt.Printf("Error creating app: %v\n", err)
		os.Exit(1)
	}

	// Success message
	fmt.Printf("Created '%s' from '%s' template\n\n", appName, *templateName)

	switch *templateName {
	case "vue", "vue-api", "vite":
		fmt.Println("Next steps:")
		fmt.Printf("  cd %s\n", appName)
		fmt.Println("  npm install        # Install dev dependencies")
		fmt.Println("  npm run dev        # Start dev server with HMR")
		fmt.Println("  npm run build      # Build for production")
		fmt.Println()
		fmt.Println("Or deploy directly (works without npm):")
		fmt.Printf("  fazt app deploy %s --to zyt\n", appName)
	default:
		fmt.Println("Next steps:")
		fmt.Printf("  fazt app deploy %s --to zyt\n", appName)
	}
}

// isValidAppName validates an app name for subdomain use
func isValidAppName(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}

	// Must start and end with alphanumeric
	if name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}

	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}

	return true
}

// writeFile writes content to a file, creating parent directories as needed
func writeFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}
