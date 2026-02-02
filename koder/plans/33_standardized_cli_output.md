# Plan 33: Standardized CLI Output (Markdown + JSON)

**Status**: Ready to Implement
**Created**: 2026-02-02
**Depends On**: Plan 31 (CLI Refactor - should be stable first)

## Summary

Standardize all CLI output to use markdown format by default (beautifully rendered in terminal) with optional JSON format for machine consumption. Creates consistency across all commands and aligns output format with documentation format.

## Motivation

**Current state:**
- Mixed output formats (tables, plain text, custom formatting)
- Inconsistent styling across commands
- Hard to parse for machines
- Output doesn't match documentation style

**With standardized output:**
```bash
# Beautiful markdown rendering (default)
$ fazt app list

# Apps

| ID         | Name   | Created    |
|------------|--------|------------|
| app_abc123 | tetris | 2026-02-01 |
| app_def456 | notes  | 2026-02-02 |

2 apps

# Machine-readable JSON (opt-in)
$ fazt app list --format json
{
  "apps": [
    {"id": "app_abc123", "name": "tetris", "created": "2026-02-01"},
    {"id": "app_def456", "name": "notes", "created": "2026-02-02"}
  ],
  "count": 2
}
```

## Design

### Output Formats

| Format | Use Case | Example |
|--------|----------|---------|
| `markdown` | Human reading (default) | Rendered with colors, tables, headers |
| `json` | Machine parsing, scripts | Structured JSON objects |

**Future formats** (not in scope): `yaml`, `csv`, `table`

### Global Flag

```bash
fazt <command> [--format <format>]
fazt <command> [-f <format>]

# Examples
fazt app list                    # markdown (default)
fazt app list --format json      # json
fazt app list -f json            # json (short)
```

### Markdown Rendering

**Library:** [glamour](https://github.com/charmbracelet/glamour)
- Terminal markdown renderer with syntax highlighting
- Supports tables, code blocks, headers, lists, bold, italic
- Multiple themes (dark, light, auto-detect)
- Used by GitHub's `gh` CLI

**Features:**
- ✅ Colored headers
- ✅ Tables with borders
- ✅ Syntax-highlighted code blocks
- ✅ Bold/italic text
- ✅ Lists (ordered, unordered)
- ✅ Horizontal rules
- ✅ Auto line wrapping

### Output Structure

**Markdown template:**
```markdown
# <Command Title>

[Optional description or context]

| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| value1   | value2   | value3   |

[Summary line: "N items"]

[Optional additional sections]
```

**JSON template:**
```json
{
  "data": [...],
  "count": N,
  "metadata": {...}
}
```

## Implementation

### Phase 1: Output Package (1 day)

Create `internal/output/` package:

**output/format.go:**
```go
package output

import (
    "encoding/json"
    "fmt"
    "os"
    "strings"

    "github.com/charmbracelet/glamour"
)

type Format string

const (
    FormatMarkdown Format = "markdown"
    FormatJSON     Format = "json"
)

type Renderer struct {
    Format Format
    theme  string
}

func NewRenderer(format Format) *Renderer {
    return &Renderer{
        Format: format,
        theme:  detectTheme(), // auto-detect terminal theme
    }
}

func (r *Renderer) Print(markdown string, data interface{}) error {
    if r.Format == FormatJSON {
        return r.printJSON(data)
    }
    return r.printMarkdown(markdown)
}

func (r *Renderer) printMarkdown(md string) error {
    renderer, err := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(80),
    )
    if err != nil {
        return err
    }

    out, err := renderer.Render(md)
    if err != nil {
        return err
    }

    fmt.Print(out)
    return nil
}

func (r *Renderer) printJSON(data interface{}) error {
    encoder := json.NewEncoder(os.Stdout)
    encoder.SetIndent("", "  ")
    return encoder.Encode(data)
}
```

**output/table.go:**
```go
package output

import "strings"

type Table struct {
    Headers []string
    Rows    [][]string
}

func (t *Table) Markdown() string {
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
```

**output/builder.go:**
```go
package output

import (
    "fmt"
    "strings"
)

type MarkdownBuilder struct {
    sections []string
}

func NewMarkdown() *MarkdownBuilder {
    return &MarkdownBuilder{}
}

func (b *MarkdownBuilder) H1(text string) *MarkdownBuilder {
    b.sections = append(b.sections, "# "+text)
    return b
}

func (b *MarkdownBuilder) H2(text string) *MarkdownBuilder {
    b.sections = append(b.sections, "## "+text)
    return b
}

func (b *MarkdownBuilder) Para(text string) *MarkdownBuilder {
    b.sections = append(b.sections, text)
    return b
}

func (b *MarkdownBuilder) Table(t *Table) *MarkdownBuilder {
    b.sections = append(b.sections, t.Markdown())
    return b
}

func (b *MarkdownBuilder) List(items []string) *MarkdownBuilder {
    for _, item := range items {
        b.sections = append(b.sections, "- "+item)
    }
    return b
}

func (b *MarkdownBuilder) Code(code, lang string) *MarkdownBuilder {
    b.sections = append(b.sections, fmt.Sprintf("```%s\n%s\n```", lang, code))
    return b
}

func (b *MarkdownBuilder) String() string {
    return strings.Join(b.sections, "\n\n")
}
```

### Phase 2: Global Flag (0.5 days)

Add to main.go:

```go
var (
    outputFormat = flag.String("format", "markdown", "Output format: markdown, json")
)

func getRenderer() *output.Renderer {
    format := output.Format(*outputFormat)
    if format != output.FormatMarkdown && format != output.FormatJSON {
        format = output.FormatMarkdown
    }
    return output.NewRenderer(format)
}
```

### Phase 3: Update Commands (2-3 days)

Update each command to use output package:

**Example: app list**

**Before:**
```go
func handleAppList(args []string) {
    apps, _ := getApps(db)
    fmt.Println("ID          NAME       CREATED")
    fmt.Println("-----------------------------------")
    for _, app := range apps {
        fmt.Printf("%-12s%-11s%s\n", app.ID, app.Name, app.Created)
    }
}
```

**After:**
```go
func handleAppList(args []string) {
    apps, _ := getApps(db)

    renderer := getRenderer()

    // Prepare data for JSON
    data := map[string]interface{}{
        "apps":  apps,
        "count": len(apps),
    }

    // Build markdown
    table := &output.Table{
        Headers: []string{"ID", "Name", "Created"},
        Rows:    make([][]string, len(apps)),
    }
    for i, app := range apps {
        table.Rows[i] = []string{app.ID, app.Name, app.Created}
    }

    md := output.NewMarkdown().
        H1("Apps").
        Table(table).
        Para(fmt.Sprintf("%d apps", len(apps))).
        String()

    renderer.Print(md, data)
}
```

**Example: peer list**

```go
func handlePeerList(args []string) {
    peers, _ := remote.ListPeers(db)

    renderer := getRenderer()

    data := map[string]interface{}{
        "peers": peers,
        "count": len(peers),
    }

    table := &output.Table{
        Headers: []string{"Name", "URL", "Status", "Version"},
        Rows:    make([][]string, len(peers)),
    }
    for i, peer := range peers {
        table.Rows[i] = []string{
            peer.Name,
            peer.URL,
            peer.Status,
            peer.Version,
        }
    }

    md := output.NewMarkdown().
        H1("Configured Peers").
        Table(table).
        Para(fmt.Sprintf("%d peers", len(peers))).
        String()

    renderer.Print(md, data)
}
```

**Example: app info**

```go
func handleAppInfo(args []string) {
    app, _ := getApp(db, identifier)

    renderer := getRenderer()

    data := app // Full app object for JSON

    details := []string{
        fmt.Sprintf("**ID**: %s", app.ID),
        fmt.Sprintf("**Name**: %s", app.Name),
        fmt.Sprintf("**Created**: %s", app.Created),
        fmt.Sprintf("**Size**: %s", app.Size),
    }

    md := output.NewMarkdown().
        H1(fmt.Sprintf("App: %s", app.Name)).
        List(details).
        H2("Aliases").
        List(app.Aliases).
        H2("Storage").
        List([]string{
            fmt.Sprintf("KV keys: %d", app.KVKeys),
            fmt.Sprintf("Blobs: %d files (%s)", app.BlobCount, app.BlobSize),
        }).
        String()

    renderer.Print(md, data)
}
```

**Example: SQL query**

```go
func handleSQL(query string, write bool) {
    result, _ := executeQuery(db, query)

    renderer := getRenderer()

    data := map[string]interface{}{
        "columns": result.Columns,
        "rows":    result.Rows,
        "count":   len(result.Rows),
        "time_ms": result.TimeMS,
    }

    table := &output.Table{
        Headers: result.Columns,
        Rows:    result.Rows,
    }

    md := output.NewMarkdown().
        H1("Query Results").
        Table(table).
        Para(fmt.Sprintf("%d rows (%dms)", len(result.Rows), result.TimeMS)).
        String()

    renderer.Print(md, data)
}
```

### Phase 4: Testing (1 day)

**Test cases:**
- Markdown rendering (visual inspection)
- JSON output validation
- Table formatting (various column widths)
- Empty results (0 items)
- Large results (pagination?)
- Special characters in output
- Color detection (light/dark terminal)

**Test script:**
```bash
#!/bin/bash
# Test all commands with both formats

echo "=== Testing app list ==="
fazt app list
fazt app list --format json

echo "=== Testing peer list ==="
fazt peer list
fazt peer list -f json

echo "=== Testing app info ==="
fazt app info myapp
fazt app info myapp --format json

echo "=== Testing SQL ==="
fazt sql "SELECT * FROM apps"
fazt sql "SELECT * FROM apps" -f json
```

## Commands to Update

### High Priority (core commands)
- [x] `app list` - Table of apps
- [x] `app info` - App details
- [x] `peer list` - Table of peers
- [x] `peer status` - Peer health info
- [x] `sql` - Query results (tables)
- [x] `server status` - Server info

### Medium Priority
- [ ] `app logs` - Log output (maybe keep as-is?)
- [ ] `app lineage` - Tree visualization
- [ ] `remote upgrade` - Upgrade status

### Low Priority (already simple)
- [ ] Success messages (keep simple: "Deployed successfully")
- [ ] Error messages (keep stderr, not formatted)

## CSV to Markdown

Simple conversion for SQL and other tabular data:

```go
func CSVToMarkdown(rows [][]string) string {
    if len(rows) == 0 {
        return ""
    }

    table := &output.Table{
        Headers: rows[0],
        Rows:    rows[1:],
    }
    return table.Markdown()
}
```

## Examples: Before & After

### App List

**Before:**
```
ID          NAME       CREATED
-----------------------------------
app_abc123  tetris     2026-02-01
app_def456  notes      2026-02-02
```

**After (Markdown rendered):**
```
 Apps

  ID           Name     Created
  ────────────────────────────────
  app_abc123   tetris   2026-02-01
  app_def456   notes    2026-02-02

2 apps
```
(With colors, bold headers, borders)

**After (JSON):**
```json
{
  "apps": [
    {
      "id": "app_abc123",
      "name": "tetris",
      "created": "2026-02-01"
    },
    {
      "id": "app_def456",
      "name": "notes",
      "created": "2026-02-02"
    }
  ],
  "count": 2
}
```

### SQL Query

**Before:**
```
ID          NAME
-----------------------
app_abc123  tetris
app_def456  notes

2 rows (3ms)
```

**After (Markdown rendered):**
```
 Query Results

  ID           Name
  ────────────────────
  app_abc123   tetris
  app_def456   notes

2 rows (3ms)
```
(With syntax highlighting, colors)

## Success Criteria

- [ ] `internal/output/` package created
- [ ] glamour dependency added
- [ ] Global `--format` flag works
- [ ] All high-priority commands output markdown
- [ ] JSON format works for all commands
- [ ] Tables render correctly in terminal
- [ ] Colors auto-detect terminal theme
- [ ] Documentation updated with examples
- [ ] Tests verify both formats

## Dependencies

**Go packages:**
```bash
go get github.com/charmbracelet/glamour
```

**Depends on:**
- Plan 31 (CLI Refactor) - should be stable first
- Commands should be refactored before changing output

## Migration Notes

- Old output format removed (no compatibility)
- Scripts using CLI output should switch to `--format json`
- Visual output will look different (better!)
- Breaking change acceptable (single user)

## Future Enhancements

- `--format yaml` - YAML output
- `--format csv` - CSV output (for SQL)
- `--format table` - Plain table (no markdown)
- `--no-color` - Disable colors in markdown
- `--theme <name>` - Choose glamour theme
- Pagination for large results
- Export to file: `--output file.md`

## Related

- **Plan 31**: CLI Refactor (@peer-primary)
- **Plan 32**: Docs Rendering System (markdown alignment)

## Notes

- Markdown output = beautiful, consistent, aligned with docs
- JSON output = machine-parseable, scriptable
- glamour library = battle-tested (used by GitHub CLI)
- All commands eventually use same output system
- Breaking change OK (single user, rapid iteration)
