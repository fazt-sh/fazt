package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
)

// ValidationResult holds the result of validating an app
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationError
}

// ValidationError represents a validation error or warning
type ValidationError struct {
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Message string `json:"message"`
}

// handleAppValidate validates an app directory before deploy
func handleAppValidate(args []string) {
	// Guard: this is a local-only command
	if targetPeerName != "" {
		fmt.Fprintf(os.Stderr, "Error: 'app validate' is a local operation\n")
		fmt.Fprintf(os.Stderr, "This command validates local files, not apps on remote peers.\n")
		fmt.Fprintf(os.Stderr, "Usage: fazt app validate <directory> [--json]\n")
		os.Exit(1)
	}

	flags := flag.NewFlagSet("app validate", flag.ExitOnError)
	jsonOutput := flags.Bool("json", false, "Output as JSON")

	flags.Usage = func() {
		fmt.Println("Usage: fazt app validate <directory> [--json]")
		fmt.Println()
		fmt.Println("Validate an app before deployment.")
		fmt.Println()
		flags.PrintDefaults()
	}

	// Find directory arg
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
		dir = "."
	}

	flags.Parse(flagArgs)

	// Validate directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Error: directory '%s' does not exist\n", dir)
		os.Exit(1)
	}

	result := validateApp(dir)

	if *jsonOutput {
		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))
	} else {
		printValidationResult(result, dir)
	}

	if !result.Valid {
		os.Exit(1)
	}
}

// validateApp performs all validation checks
func validateApp(dir string) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// 1. Check manifest.json
	validateManifest(dir, result)

	// 2. Check required files
	validateRequiredFiles(dir, result)

	// 3. Validate JavaScript files in api/
	validateAPIFiles(dir, result)

	return result
}

// validateManifest checks manifest.json
func validateManifest(dir string, result *ValidationResult) {
	manifestPath := filepath.Join(dir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			File:    "manifest.json",
			Message: "manifest.json not found (required)",
		})
		result.Valid = false
		return
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			File:    "manifest.json",
			Message: fmt.Sprintf("invalid JSON: %v", err),
		})
		result.Valid = false
		return
	}

	// Check required fields
	if _, ok := manifest["name"]; !ok {
		result.Errors = append(result.Errors, ValidationError{
			File:    "manifest.json",
			Message: "missing required field: name",
		})
		result.Valid = false
	} else if name, ok := manifest["name"].(string); ok {
		if !isValidAppName(name) {
			result.Errors = append(result.Errors, ValidationError{
				File:    "manifest.json",
				Message: "invalid name: must be lowercase letters, numbers, and hyphens only",
			})
			result.Valid = false
		}
	}
}

// validateRequiredFiles checks for required files
func validateRequiredFiles(dir string, result *ValidationResult) {
	// Check for index.html
	indexPath := filepath.Join(dir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		result.Warnings = append(result.Warnings, ValidationError{
			File:    "index.html",
			Message: "index.html not found (recommended for static sites)",
		})
	}
}

// validateAPIFiles validates JavaScript files in the api/ directory
func validateAPIFiles(dir string, result *ValidationResult) {
	apiDir := filepath.Join(dir, "api")
	if _, err := os.Stat(apiDir); os.IsNotExist(err) {
		// No api/ directory, that's fine
		return
	}

	filepath.Walk(apiDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".js") {
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)
		validateJSFile(path, relPath, result)
		return nil
	})
}

// validateJSFile parses a JS file to check for syntax errors
func validateJSFile(path, relPath string, result *ValidationResult) {
	content, err := os.ReadFile(path)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			File:    relPath,
			Message: fmt.Sprintf("cannot read file: %v", err),
		})
		result.Valid = false
		return
	}

	// Try to run the code in a VM to check for syntax errors
	// We wrap it in a function to avoid actually executing it
	vm := goja.New()
	testCode := fmt.Sprintf("(function(){%s})", string(content))
	_, err = vm.RunString(testCode)
	if err != nil {
		// Parse the error to extract line/column info
		errMsg := err.Error()
		verr := ValidationError{
			File:    relPath,
			Message: errMsg,
		}

		// Try to extract line number from error
		if syntaxErr, ok := err.(*goja.CompilerSyntaxError); ok {
			verr.Message = syntaxErr.Error()
		}

		result.Errors = append(result.Errors, verr)
		result.Valid = false
		return
	}

	// Check for handler function
	code := string(content)
	if !strings.Contains(code, "function handler") && !strings.Contains(code, "handler =") {
		result.Warnings = append(result.Warnings, ValidationError{
			File:    relPath,
			Message: "no handler function found (required for serverless)",
		})
	}
}

// printValidationResult prints validation results in a human-readable format
func printValidationResult(result *ValidationResult, dir string) {
	fmt.Printf("Validating %s...\n\n", dir)

	if len(result.Errors) == 0 && len(result.Warnings) == 0 {
		fmt.Println("✓ All checks passed")
		return
	}

	for _, err := range result.Errors {
		location := err.File
		if err.Line > 0 {
			location = fmt.Sprintf("%s:%d", err.File, err.Line)
		}
		fmt.Printf("✗ %s - %s\n", location, err.Message)
	}

	for _, warn := range result.Warnings {
		location := warn.File
		if warn.Line > 0 {
			location = fmt.Sprintf("%s:%d", warn.File, warn.Line)
		}
		fmt.Printf("⚠ %s - %s\n", location, warn.Message)
	}

	fmt.Println()
	if result.Valid {
		fmt.Printf("✓ Valid (%d warnings)\n", len(result.Warnings))
	} else {
		fmt.Printf("✗ Invalid (%d errors, %d warnings)\n", len(result.Errors), len(result.Warnings))
	}
}
