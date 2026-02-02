package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
)

// Format represents the output format type
type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatJSON     Format = "json"
)

// Renderer handles output formatting and rendering
type Renderer struct {
	Format Format
	theme  string
}

// NewRenderer creates a new output renderer
func NewRenderer(format Format) *Renderer {
	return &Renderer{
		Format: format,
		theme:  "auto", // auto-detect terminal theme
	}
}

// Print renders output in the specified format
// For markdown: renders the markdown string with glamour
// For JSON: marshals and prints the data object
func (r *Renderer) Print(markdown string, data interface{}) error {
	if r.Format == FormatJSON {
		return r.printJSON(data)
	}
	return r.printMarkdown(markdown)
}

// printMarkdown renders markdown with glamour
func (r *Renderer) printMarkdown(md string) error {
	// Check if output is a terminal
	if !isTerminal() {
		// Not a terminal, output plain markdown
		fmt.Print(md)
		return nil
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		// Fallback to plain output
		fmt.Print(md)
		return nil
	}

	out, err := renderer.Render(md)
	if err != nil {
		// Fallback to plain output
		fmt.Print(md)
		return nil
	}

	fmt.Print(out)
	return nil
}

// printJSON outputs data as formatted JSON
func (r *Renderer) printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
