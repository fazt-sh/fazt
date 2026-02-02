package output

import (
	"fmt"
	"strings"
)

// MarkdownBuilder helps build markdown documents
type MarkdownBuilder struct {
	sections []string
}

// NewMarkdown creates a new markdown builder
func NewMarkdown() *MarkdownBuilder {
	return &MarkdownBuilder{}
}

// H1 adds a level 1 header
func (b *MarkdownBuilder) H1(text string) *MarkdownBuilder {
	b.sections = append(b.sections, "# "+text)
	return b
}

// H2 adds a level 2 header
func (b *MarkdownBuilder) H2(text string) *MarkdownBuilder {
	b.sections = append(b.sections, "## "+text)
	return b
}

// H3 adds a level 3 header
func (b *MarkdownBuilder) H3(text string) *MarkdownBuilder {
	b.sections = append(b.sections, "### "+text)
	return b
}

// Para adds a paragraph
func (b *MarkdownBuilder) Para(text string) *MarkdownBuilder {
	b.sections = append(b.sections, text)
	return b
}

// Table adds a table
func (b *MarkdownBuilder) Table(t *Table) *MarkdownBuilder {
	b.sections = append(b.sections, t.Markdown())
	return b
}

// List adds an unordered list
func (b *MarkdownBuilder) List(items []string) *MarkdownBuilder {
	for _, item := range items {
		b.sections = append(b.sections, "- "+item)
	}
	return b
}

// OrderedList adds an ordered list
func (b *MarkdownBuilder) OrderedList(items []string) *MarkdownBuilder {
	for i, item := range items {
		b.sections = append(b.sections, fmt.Sprintf("%d. %s", i+1, item))
	}
	return b
}

// Code adds a code block
func (b *MarkdownBuilder) Code(code, lang string) *MarkdownBuilder {
	b.sections = append(b.sections, fmt.Sprintf("```%s\n%s\n```", lang, code))
	return b
}

// Bold returns bold text
func Bold(text string) string {
	return "**" + text + "**"
}

// Italic returns italic text
func Italic(text string) string {
	return "*" + text + "*"
}

// Code returns inline code
func Code(text string) string {
	return "`" + text + "`"
}

// Rule adds a horizontal rule
func (b *MarkdownBuilder) Rule() *MarkdownBuilder {
	b.sections = append(b.sections, "---")
	return b
}

// String returns the final markdown document
func (b *MarkdownBuilder) String() string {
	return strings.Join(b.sections, "\n\n")
}
