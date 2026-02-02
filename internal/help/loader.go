package help

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"gopkg.in/yaml.v3"
)

// docs is populated by embed.go with the CLI documentation files
var docs embed.FS

// SetDocs sets the embedded filesystem (called from embed.go init)
func SetDocs(f embed.FS) {
	docs = f
}

// Load loads and parses a command's help documentation
// cmdPath examples:
//   - "" or "fazt" -> cli/fazt.md (root help)
//   - "app" -> cli/app/_index.md (group help)
//   - "app deploy" -> cli/app/deploy.md
//   - "peer" -> cli/peer/_index.md
func Load(cmdPath string) (*CommandDoc, error) {
	filePath := resolveFilePath(cmdPath)

	content, err := fs.ReadFile(docs, filePath)
	if err != nil {
		return nil, fmt.Errorf("help not found for '%s': %w", cmdPath, err)
	}

	return parse(content)
}

// Exists checks if help documentation exists for a command
func Exists(cmdPath string) bool {
	filePath := resolveFilePath(cmdPath)
	_, err := fs.ReadFile(docs, filePath)
	return err == nil
}

// resolveFilePath converts a command path to a file path
func resolveFilePath(cmdPath string) string {
	cmdPath = strings.TrimSpace(cmdPath)

	// Root help
	if cmdPath == "" || cmdPath == "fazt" {
		return "cli/fazt.md"
	}

	parts := strings.Fields(cmdPath)

	// Single word = group help (e.g., "app" -> cli/app/_index.md)
	if len(parts) == 1 {
		return fmt.Sprintf("cli/%s/_index.md", parts[0])
	}

	// Multi-word = specific command (e.g., "app deploy" -> cli/app/deploy.md)
	return fmt.Sprintf("cli/%s/%s.md", parts[0], parts[1])
}

// parse parses markdown content with YAML frontmatter
func parse(content []byte) (*CommandDoc, error) {
	// Check for frontmatter delimiter
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return nil, fmt.Errorf("invalid help file: missing YAML frontmatter")
	}

	// Find the end of frontmatter
	rest := content[4:] // Skip opening "---\n"
	endIdx := bytes.Index(rest, []byte("\n---"))
	if endIdx == -1 {
		return nil, fmt.Errorf("invalid help file: unclosed YAML frontmatter")
	}

	frontmatter := rest[:endIdx]
	body := rest[endIdx+4:] // Skip "\n---"

	// Parse YAML frontmatter
	doc := &CommandDoc{}
	if err := yaml.Unmarshal(frontmatter, doc); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Store the markdown body
	doc.Body = strings.TrimSpace(string(body))

	return doc, nil
}
