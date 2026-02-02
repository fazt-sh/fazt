package output

import "strings"

// Table represents a markdown table
type Table struct {
	Headers []string
	Rows    [][]string
}

// Markdown converts the table to markdown format
func (t *Table) Markdown() string {
	if len(t.Rows) == 0 {
		return ""
	}

	var b strings.Builder

	// Headers
	b.WriteString("| ")
	b.WriteString(strings.Join(t.Headers, " | "))
	b.WriteString(" |\n")

	// Separator
	b.WriteString("|")
	for range t.Headers {
		b.WriteString("---|")
	}
	b.WriteString("\n")

	// Rows
	for _, row := range t.Rows {
		b.WriteString("| ")
		b.WriteString(strings.Join(row, " | "))
		b.WriteString(" |\n")
	}

	return b.String()
}
