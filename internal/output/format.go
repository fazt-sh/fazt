package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/glamour"
)

// TimeAgo formats a time as a human-readable relative string
func TimeAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	d := time.Since(t)
	if d < 0 {
		return "just now"
	}

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	case d < 30*24*time.Hour:
		weeks := int(d.Hours() / 24 / 7)
		if weeks == 1 {
			return "1w ago"
		}
		return fmt.Sprintf("%dw ago", weeks)
	case d < 365*24*time.Hour:
		months := int(d.Hours() / 24 / 30)
		if months == 1 {
			return "1mo ago"
		}
		return fmt.Sprintf("%dmo ago", months)
	default:
		years := int(d.Hours() / 24 / 365)
		if years == 1 {
			return "1y ago"
		}
		return fmt.Sprintf("%dy ago", years)
	}
}

// TimeAgoUnix formats a unix timestamp as a human-readable relative string
func TimeAgoUnix(unix int64) string {
	if unix == 0 {
		return "never"
	}
	return TimeAgo(time.Unix(unix, 0))
}

// TimeAgoString parses a datetime string and returns time ago
// Supports formats: "2006-01-02 15:04:05" and "2006-01-02T15:04:05Z"
func TimeAgoString(s string) string {
	if s == "" {
		return "never"
	}

	// Try common formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			return TimeAgo(t)
		}
	}

	return s // Return original if parsing fails
}

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
