package help

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
)

// Render renders a CommandDoc as terminal-formatted output
func Render(doc *CommandDoc) string {
	md := buildMarkdown(doc)
	return renderMarkdown(md)
}

// RenderBrief renders just the brief help (no extended body)
func RenderBrief(doc *CommandDoc) string {
	md := buildBriefMarkdown(doc)
	return renderMarkdown(md)
}

// buildMarkdown constructs markdown from the structured CommandDoc
func buildMarkdown(doc *CommandDoc) string {
	var b strings.Builder

	// Title and description
	b.WriteString(fmt.Sprintf("# fazt %s\n\n", doc.Command))
	b.WriteString(fmt.Sprintf("%s\n\n", doc.Description))

	// Syntax
	b.WriteString("## Usage\n\n")
	b.WriteString(fmt.Sprintf("```\n%s\n```\n\n", doc.Syntax))

	// Arguments
	if len(doc.Arguments) > 0 {
		b.WriteString("## Arguments\n\n")
		for _, arg := range doc.Arguments {
			required := ""
			if arg.Required {
				required = " (required)"
			}
			b.WriteString(fmt.Sprintf("**`<%s>`**%s\n\n", arg.Name, required))
			b.WriteString(fmt.Sprintf("%s\n\n", arg.Description))
		}
	}

	// Flags
	if len(doc.Flags) > 0 {
		b.WriteString("## Flags\n\n")
		for _, flag := range doc.Flags {
			name := flag.Name
			if flag.Short != "" {
				name = fmt.Sprintf("%s, %s", flag.Short, flag.Name)
			}
			b.WriteString(fmt.Sprintf("**`%s`**", name))
			if flag.Type != "bool" && flag.Type != "" {
				b.WriteString(fmt.Sprintf(" `<%s>`", flag.Type))
			}
			b.WriteString("\n\n")
			b.WriteString(fmt.Sprintf("%s", flag.Description))
			if flag.Default != "" && flag.Default != "false" {
				b.WriteString(fmt.Sprintf(" (default: %s)", flag.Default))
			}
			b.WriteString("\n\n")
		}
	}

	// Examples
	if len(doc.Examples) > 0 {
		b.WriteString("## Examples\n\n")
		for _, ex := range doc.Examples {
			b.WriteString(fmt.Sprintf("**%s**\n\n", ex.Title))
			b.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", ex.Command))
			if ex.Description != "" {
				b.WriteString(fmt.Sprintf("%s\n\n", ex.Description))
			}
		}
	}

	// Related commands
	if len(doc.Related) > 0 {
		b.WriteString("## See Also\n\n")
		for _, rel := range doc.Related {
			b.WriteString(fmt.Sprintf("- **`fazt %s`** - %s\n", rel.Command, rel.Description))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// buildBriefMarkdown constructs a shorter version without extended docs
func buildBriefMarkdown(doc *CommandDoc) string {
	var b strings.Builder

	// Title and description
	b.WriteString(fmt.Sprintf("# fazt %s\n\n", doc.Command))
	b.WriteString(fmt.Sprintf("%s\n\n", doc.Description))

	// Syntax
	b.WriteString("## Usage\n\n")
	b.WriteString(fmt.Sprintf("```\n%s\n```\n\n", doc.Syntax))

	// Arguments (brief)
	if len(doc.Arguments) > 0 {
		b.WriteString("## Arguments\n\n")
		for _, arg := range doc.Arguments {
			required := ""
			if arg.Required {
				required = " (required)"
			}
			b.WriteString(fmt.Sprintf("- `<%s>`%s - %s\n", arg.Name, required, arg.Description))
		}
		b.WriteString("\n")
	}

	// Flags (brief)
	if len(doc.Flags) > 0 {
		b.WriteString("## Flags\n\n")
		for _, flag := range doc.Flags {
			name := flag.Name
			if flag.Short != "" {
				name = fmt.Sprintf("%s, %s", flag.Short, flag.Name)
			}
			b.WriteString(fmt.Sprintf("- `%s` - %s\n", name, flag.Description))
		}
		b.WriteString("\n")
	}

	// Examples (first 3 only)
	if len(doc.Examples) > 0 {
		b.WriteString("## Examples\n\n")
		count := len(doc.Examples)
		if count > 3 {
			count = 3
		}
		for i := 0; i < count; i++ {
			ex := doc.Examples[i]
			b.WriteString(fmt.Sprintf("```bash\n%s\n```\n", ex.Command))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderMarkdown renders markdown with glamour for terminal output
func renderMarkdown(md string) string {
	// Check if output is a terminal
	if !isTerminal() {
		// Not a terminal, return plain markdown
		return md
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		// Fallback to plain output
		return md
	}

	out, err := renderer.Render(md)
	if err != nil {
		// Fallback to plain output
		return md
	}

	return out
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
