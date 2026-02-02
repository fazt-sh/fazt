# Fazt Docs Rendering System Design

**Version**: 0.18.0
**Status**: Design Draft
**Created**: 2026-02-02
**Updated**: 2026-02-02

## Overview

A generic markdown-to-HTML rendering system for Fazt that supports:

1. **CLI Help** - Terminal rendering with colors/formatting (from markdown)
2. **Web HTML** - Themed static HTML deployed as fazt app
3. **Web TXT** - Plain text files (markdown minus frontmatter) for LLMs/CLI

**Key decisions:**
- ✅ Two formats only: `.html` and `.txt` (no separate `.md` files)
- ✅ TXT file IS markdown content (frontmatter stripped)
- ✅ Copy button fetches `.txt` (which is markdown)
- ✅ Frontmatter uses standard blog fields (`title`, `date`, `tags`)
- ✅ No `version` field in frontmatter (use monorepo version.json)
- ✅ GitHub Flavored Markdown (GFM) supported
- ✅ Blog themes work out of the box

The system dog-foods Fazt's own documentation while remaining generic enough
for any content site (blogs, general docs, etc.).

---

## 1. Blog-Aligned Frontmatter Schema

The frontmatter schema aligns with standard blog platforms (Jekyll, Hugo,
Gatsby, Astro) so existing themes can work with minimal changes.

### Design Principles

1. **Use standard blog field names** - Themes expect `title`, `date`, `tags`
2. **Add CLI fields as extensions** - `command`, `syntax`, `arguments`, etc.
3. **No field conflicts** - CLI fields are namespaced or optional
4. **Support pure blog posts** - No CLI fields required
5. **Support pure CLI docs** - Blog fields optional for commands

### Standard Blog Fields

These fields work with any blog theme out of the box:

```yaml
---
# Required
title: "Page Title"                      # Display title

# Recommended
date: 2026-02-02                         # Publication/update date
description: "Brief summary"             # Meta description, excerpt

# Optional
tags: [cli, deployment]                  # Searchable tags (always array)
category: commands                       # Single category (string)
author: fazt                             # Author name or identifier
image: /images/deploy-guide.png          # Featured/OG image
draft: false                             # Hide from production
---
```

### CLI Extension Fields

Additional fields for command documentation:

```yaml
---
# Command routing
command: "app deploy"                    # CLI routing path
aliases: ["deploy"]                      # Alternative invocations

# Syntax display
syntax: "fazt [@peer] app deploy <directory> [flags]"

# Structured data for rendering
arguments:
  - name: "directory"
    type: "path"
    required: true
    description: "Path to the directory to deploy"
    default: null

flags:
  - name: "--name"
    short: "-n"
    type: "string"
    default: "directory name"
    description: "App name"

# Peer support metadata
peer:
  supported: true
  local: true
  remote: true

# Examples with context
examples:
  - title: "Deploy to local fazt"
    command: "fazt app deploy ./my-app"
    description: "Deploy directory to local instance"

# Related commands
related:
  - command: "app list"
    description: "List deployed apps"

# Common errors
errors:
  - code: "ENOENT"
    message: "directory does not exist"
    solution: "Verify the path exists"
---
```

### Web/Navigation Fields

Additional fields for web documentation:

```yaml
---
# Navigation
order: 10                                # Sort order (lower = first)
parent: "app"                            # Parent doc for breadcrumbs
slug: "app-deploy"                       # URL-friendly identifier

# Content features
toc: true                                # Generate table of contents
toc_depth: 3                             # TOC depth (h1-h3)
reading_time: true                       # Show estimated reading time

# SEO
canonical: "https://docs.zyt.app/deploy" # Canonical URL

# Series/Sequence
series: "getting-started"                # Part of a series
series_order: 3                          # Position in series
prev: "installation"                     # Previous doc
next: "configuration"                    # Next doc

# Blog-specific
featured: true                           # Featured on homepage
---
```

### Field Reference by Content Type

| Field | Blog Post | CLI Command | General Doc | Required |
|-------|:---------:|:-----------:|:-----------:|:--------:|
| `title` | Yes | Yes | Yes | **Yes** |
| `date` | Yes | Yes | Yes | No |
| `description` | Yes | Yes | Yes | No |
| `tags` | Yes | Yes | Yes | No |
| `category` | Yes | Yes | Yes | No |
| `author` | Yes | - | - | No |
| `image` | Yes | - | - | No |
| `draft` | Yes | Yes | Yes | No |
| `command` | - | **Yes** | - | CLI only |
| `syntax` | - | Yes | - | No |
| `arguments` | - | Yes | - | No |
| `flags` | - | Yes | - | No |
| `examples` | - | Yes | - | No |
| `related` | - | Yes | - | No |
| `errors` | - | Yes | - | No |
| `peer` | - | Yes | - | No |
| `toc` | - | - | Yes | No |
| `order` | - | Yes | Yes | No |
| `series` | - | - | Yes | No |

### Example: Blog Post

```yaml
---
title: "Introducing Fazt 0.18"
date: 2026-02-02
description: "New @peer pattern and CLI improvements"
tags: [release, cli, announcement]
category: blog
author: fazt
image: /images/v018-release.png
---

Today we're releasing Fazt 0.18 with major improvements...
```

### Example: CLI Command

```yaml
---
title: "App Deploy Command"
date: 2026-02-02
description: "Deploy apps to fazt instances"
tags: [cli, deployment]
category: commands

# CLI-specific fields
command: "app deploy"
syntax: "fazt [@peer] app deploy <directory> [flags]"
arguments:
  - name: "directory"
    type: "path"
    required: true
    description: "Path to the directory to deploy"
flags:
  - name: "--name"
    short: "-n"
    type: "string"
    default: "directory name"
    description: "App name (overrides directory name)"
  - name: "--spa"
    type: "bool"
    default: false
    description: "Enable SPA routing for clean URLs"
examples:
  - title: "Deploy to local"
    command: "fazt app deploy ./my-app"
  - title: "Deploy to remote"
    command: "fazt @zyt app deploy ./my-app"
related:
  - command: "app list"
    description: "List deployed apps"
  - command: "app remove"
    description: "Remove a deployed app"
---
```

### Example: General Documentation

```yaml
---
title: "Architecture Overview"
date: 2026-02-02
description: "How Fazt works internally"
tags: [architecture, internals]
category: docs
toc: true
toc_depth: 2
order: 10
---

Fazt follows a single-binary architecture...
```

### Key Design Decisions

**Version field removed**: Use monorepo `version.json` instead of per-doc
versioning. This ensures consistency and reduces maintenance.

**Date field**: Uses standard `date` field (not `updated`). Represents
last meaningful update. Blog themes expect this name.

**Category vs categories**: Use `category` (singular string). Simpler for
navigation and consistent with Jekyll defaults. For multiple categories,
use `tags` instead.

**Tags always array**: Even single tags should be `tags: [cli]` not
`tags: cli`. This ensures consistent processing.

**Description field**: Standard blog field, used as:
- Meta description for SEO
- Excerpt for listings
- Summary in search results
- OpenGraph description

**Command field**: Only present in CLI docs. Blog themes ignore unknown
fields, so this doesn't break theme compatibility.

---

## 2. Plain Text (.txt) Rendering

### Rationale

Alongside HTML, generate `.txt` files for:

- **LLM consumption** - Lower token count, cleaner parsing
- **CLI tools** - `curl`, `wget`, terminal browsers
- **Bandwidth** - Smaller files for automated access
- **Readability** - Opens as plain text in browsers (no download)

### URL Pattern

```
docs.zyt.app/cli/app/deploy.html  -> Themed HTML for browsers
docs.zyt.app/cli/app/deploy.txt   -> Plain text for LLMs/CLI
```

### Rendering Pipeline

```
Markdown source
  |
  +---> HTML renderer ---> styled.html (themes, CSS, sidebar)
  |
  +---> TXT renderer  ---> plain.txt (ASCII, <80 chars)
  |
  +---> Binary --help ---> terminal (ANSI colors)
```

**Key insight**: The `.txt` renderer uses THE SAME formatting logic as
the binary `--help` renderer, minus ANSI color codes. This ensures
consistency and code reuse.

### .txt Formatting Rules

| Rule | Value | Rationale |
|------|-------|-----------|
| Line width | 80 chars max | Terminal-friendly |
| Horizontal rules | 80 `=` or `-` chars | Visual separation |
| Section headers | UPPERCASE | Clear hierarchy |
| Subsections | 2-space indent | Readable nesting |
| Code examples | 4-space indent, `$` prompt | Consistent style |
| Tables | ASCII box drawing | Universal support |
| Line wrapping | Hard wrap at 80 | No terminal reflow |
| Character set | ASCII only | Maximum compatibility |
| No ANSI codes | Plain text | Not terminal output |
| No HTML | Pure text | LLM-friendly |

### .txt Structure for Commands

```
================================================================================
COMMAND NAME
================================================================================

Brief description of the command

--------------------------------------------------------------------------------
SYNOPSIS
--------------------------------------------------------------------------------

  fazt [@peer] command <args> [flags]

--------------------------------------------------------------------------------
DESCRIPTION
--------------------------------------------------------------------------------

  Longer description of what the command does, wrapped at 80 characters for
  terminal readability. This section can span multiple paragraphs.

--------------------------------------------------------------------------------
ARGUMENTS
--------------------------------------------------------------------------------

  <argument-name>
    Description of the argument
    Type: type (required|optional)
    Default: default value

--------------------------------------------------------------------------------
FLAGS
--------------------------------------------------------------------------------

  --flag-name, -f <type>
    Description of the flag
    Default: default value

  --another-flag
    Description of boolean flag

--------------------------------------------------------------------------------
EXAMPLES
--------------------------------------------------------------------------------

  Example title:
    $ fazt command ./path

  Another example:
    $ fazt @peer command ./path --flag value

--------------------------------------------------------------------------------
RELATED COMMANDS
--------------------------------------------------------------------------------

  related list       List related items
  related remove     Remove related item

--------------------------------------------------------------------------------
COMMON ERRORS
--------------------------------------------------------------------------------

  Error description
    Error: error message
    Solution: how to fix

================================================================================
```

### Shared Rendering Logic

The binary help and web .txt use the same core renderer:

```go
package render

// TextRenderer renders markdown to plain text
type TextRenderer struct {
    Width       int  // Line width (default 80)
    UseANSI     bool // Enable colors (binary) or not (web .txt)
}

// RenderCommand renders a CommandDoc to text
func (r *TextRenderer) RenderCommand(doc *CommandDoc) string {
    var buf strings.Builder

    // Title banner
    r.writeHeader(&buf, strings.ToUpper(doc.Title))

    // Description
    buf.WriteString("\n")
    buf.WriteString(r.wrap(doc.Description, 0))
    buf.WriteString("\n")

    // Synopsis
    r.writeSection(&buf, "SYNOPSIS")
    r.writeIndented(&buf, doc.Syntax, 2)

    // Arguments
    if len(doc.Arguments) > 0 {
        r.writeSection(&buf, "ARGUMENTS")
        for _, arg := range doc.Arguments {
            r.writeArgument(&buf, arg)
        }
    }

    // Flags
    if len(doc.Flags) > 0 {
        r.writeSection(&buf, "FLAGS")
        for _, flag := range doc.Flags {
            r.writeFlag(&buf, flag)
        }
    }

    // Examples
    if len(doc.Examples) > 0 {
        r.writeSection(&buf, "EXAMPLES")
        for _, ex := range doc.Examples {
            r.writeExample(&buf, ex)
        }
    }

    // Footer
    r.writeFooter(&buf)

    return buf.String()
}

// For binary --help: r := &TextRenderer{Width: 80, UseANSI: true}
// For web .txt:      r := &TextRenderer{Width: 80, UseANSI: false}
```

### Content Parity

The `.txt` and `.html` files contain the **same content** from the same
markdown source. The only differences are presentation:

| Aspect | .html | .txt |
|--------|-------|------|
| Styling | CSS themes | None |
| Navigation | Sidebar, breadcrumbs | None |
| Tables | HTML tables | ASCII tables |
| Code | Syntax highlighting | Plain indented |
| Links | Clickable hyperlinks | Text only |
| Images | Displayed | Alt text only |
| Search | Interactive | None |
| Size | ~15KB typical | ~3KB typical |

---

## 3. Build Process

### Updated Pipeline

```
Input (Markdown)              Process                    Output
--------------------          ----------------------     --------------------
knowledge-base/               1. Scan directory          dist/
├── cli/                      2. Parse frontmatter       ├── cli/
│   ├── app/                  3. Build navigation        │   ├── app/
│   │   └── deploy.md   -->   4. Render markdown    -->  │   │   ├── deploy.html
│   └── commands.md           5. Apply template          │   │   └── deploy.txt
├── blog/                     6. Syntax highlight        │   └── commands.html
│   └── v018-release.md       7. Generate .txt files     │   └── commands.txt
└── _templates/               8. Generate metadata       ├── blog/
    └── page.html             9. Copy assets             │   └── v018-release.html
                                                         ├── _metadata.json
                                                         ├── _search.json
                                                         └── _nav.json
```

### fazt-md Command

```bash
# Basic usage - outputs both .html and .txt
fazt-md build <input-dir> --output <output-dir>

# With options
fazt-md build knowledge-base/ \
  --output dist/ \
  --template templates/default.html \
  --theme fazt-docs \
  --base-url https://docs.zyt.app \
  --txt true \
  --txt-width 80

# Skip .txt generation
fazt-md build knowledge-base/ --output dist/ --txt false
```

### Directory Structure Output

```
dist/
├── index.html
├── index.txt
├── cli/
│   ├── index.html
│   ├── index.txt
│   ├── app/
│   │   ├── deploy.html
│   │   ├── deploy.txt
│   │   ├── list.html
│   │   ├── list.txt
│   │   └── ...
│   └── ...
├── blog/
│   ├── v018-release.html
│   └── v018-release.txt
├── theme.css
├── syntax.css
├── search.js
├── _metadata.json
├── _search.json
└── _nav.json
```

### Content Negotiation

The fazt server supports content negotiation:

```bash
# Request HTML (default)
curl https://docs.zyt.app/cli/app/deploy

# Request plain text
curl -H "Accept: text/plain" https://docs.zyt.app/cli/app/deploy

# Direct file access
curl https://docs.zyt.app/cli/app/deploy.txt
curl https://docs.zyt.app/cli/app/deploy.html
```

Server logic:
1. If URL ends with `.txt` or `.html`, serve that file
2. If `Accept: text/plain`, serve `.txt`
3. Otherwise, serve `.html`

### Generated Metadata

**`_metadata.json`** includes .txt paths:

```json
{
  "generated": "2026-02-02T10:30:00Z",
  "documents": [
    {
      "slug": "cli/app/deploy",
      "title": "App Deploy Command",
      "html_path": "/cli/app/deploy.html",
      "txt_path": "/cli/app/deploy.txt",
      "category": "commands",
      "tags": ["deployment", "apps"],
      "date": "2026-02-02",
      "description": "Deploy apps to fazt instances"
    }
  ]
}
```

---

## 4. Theme System Updates

### Blog Theme Compatibility

Themes must support standard blog frontmatter fields:

```html
<!-- Required theme variables -->
{{.Title}}        <!-- From: title -->
{{.Description}}  <!-- From: description -->
{{.Date}}         <!-- From: date -->
{{.Tags}}         <!-- From: tags -->
{{.Category}}     <!-- From: category -->
{{.Author}}       <!-- From: author -->
{{.Image}}        <!-- From: image -->
{{.Draft}}        <!-- From: draft -->
```

CLI-specific fields are optional in templates:

```html
{{if .Command}}
<section class="command-meta">
  <!-- CLI command rendering -->
</section>
{{end}}
```

### .txt Link in HTML

Every HTML page includes a link to its .txt version:

```html
<footer class="doc-footer">
  <div class="meta">
    <time datetime="{{.Date}}">Updated: {{.DateFormatted}}</time>
    <a href="{{.TxtPath}}" class="txt-link" title="Plain text version">
      View as plain text
    </a>
  </div>
</footer>
```

Or in the `<head>`:

```html
<link rel="alternate" type="text/plain" href="{{.TxtPath}}"
      title="Plain text version">
```

### Theme CSS Classes

Additional classes for .txt-related elements:

```css
/* Link to .txt version */
.txt-link {
  font-size: 0.875rem;
  color: var(--color-text-muted);
}

.txt-link:hover {
  color: var(--color-primary);
}

/* Format indicator */
.format-toggle {
  display: flex;
  gap: 0.5rem;
}

.format-toggle a {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
}

.format-toggle a.active {
  background: var(--color-primary);
  color: white;
}
```

---

## 5. HTML Structure (Theme-Friendly)

Semantic HTML with consistent CSS classes for easy theming.

### Document Classes

```css
/* Layout */
.sidebar { }              /* Left navigation */
.content { }              /* Main content area */
.doc { }                  /* Document wrapper */

/* Document parts */
.doc-header { }           /* Title, description, meta */
.doc-body { }             /* Main content */
.doc-footer { }           /* Timestamps, edit links */

/* Navigation */
.breadcrumb { }           /* Breadcrumb trail */
.toc { }                  /* Table of contents */
.nav-tree { }             /* Sidebar navigation tree */
.nav-item { }             /* Navigation item */
.nav-item.active { }      /* Current page */

/* CLI-specific */
.command-meta { }         /* Command metadata block */
.syntax { }               /* Command syntax */
.arguments { }            /* Arguments list */
.flags { }                /* Flags list */
.examples { }             /* Examples section */
.errors { }               /* Common errors */

/* Content */
.code-block { }           /* Code block wrapper */
.inline-code { }          /* Inline code */
.highlight { }            /* Syntax highlighted block */
.note { }                 /* Note/warning callout */
.tip { }                  /* Tip callout */
.warning { }              /* Warning callout */

/* Interactive */
.copy-button { }          /* Copy code button */
.theme-toggle { }         /* Dark/light toggle */
.search-box { }           /* Search input */
.search-results { }       /* Search results dropdown */
.txt-link { }             /* Link to .txt version */
```

### CLI Command Rendering

For CLI command docs, render structured frontmatter as semantic HTML:

```html
<article class="doc doc-command">
  <header class="doc-header">
    <nav class="breadcrumb">
      <a href="/cli">CLI</a> / <a href="/cli/app">app</a> / <span>deploy</span>
    </nav>
    <h1>fazt app deploy</h1>
    <p class="description">Deploy a local directory to a fazt instance</p>
  </header>

  <section class="command-meta">
    <h2>Synopsis</h2>
    <pre class="syntax"><code>fazt [@peer] app deploy &lt;directory&gt; [flags]</code></pre>
  </section>

  <section class="arguments">
    <h2>Arguments</h2>
    <dl>
      <dt><code>&lt;directory&gt;</code> <span class="required">required</span></dt>
      <dd>Path to the directory to deploy</dd>
    </dl>
  </section>

  <section class="flags">
    <h2>Flags</h2>
    <dl>
      <dt><code>--name, -n</code> <span class="type">string</span></dt>
      <dd>App name (default: directory name)</dd>

      <dt><code>--spa</code> <span class="type">bool</span></dt>
      <dd>Enable SPA routing for clean URLs</dd>
    </dl>
  </section>

  <section class="examples">
    <h2>Examples</h2>
    <div class="example">
      <h3>Deploy to local fazt</h3>
      <pre class="highlight"><code class="language-bash">fazt app deploy ./my-app</code></pre>
      <p>Deploy the ./my-app directory to the local fazt instance</p>
    </div>
  </section>

  <section class="doc-body">
    <!-- Rendered markdown content -->
  </section>

  <section class="related">
    <h2>See Also</h2>
    <ul>
      <li><a href="/cli/app/list">fazt app list</a> - List deployed apps</li>
      <li><a href="/cli/app/remove">fazt app remove</a> - Remove a deployed app</li>
    </ul>
  </section>

  <footer class="doc-footer">
    <time datetime="2026-02-02">Updated: Feb 2, 2026</time>
    <a href="/cli/app/deploy.txt" class="txt-link">View as plain text</a>
  </footer>
</article>
```

---

## 6. Binary Help System (Terminal)

Separate from web rendering - the binary reads markdown directly.

### Embedded Documentation

```go
package docs

import (
    "embed"
    "gopkg.in/yaml.v3"
)

//go:embed cli/**/*.md
var DocsFS embed.FS

type CommandDoc struct {
    Title       string      `yaml:"title"`
    Command     string      `yaml:"command"`
    Syntax      string      `yaml:"syntax"`
    Description string      `yaml:"description"`
    Arguments   []Argument  `yaml:"arguments"`
    Flags       []Flag      `yaml:"flags"`
    Examples    []Example   `yaml:"examples"`
    Errors      []Error     `yaml:"errors"`
    Related     []Related   `yaml:"related"`
    Content     string      // Markdown body
}

func LoadCommand(cmd string) (*CommandDoc, error) {
    // Map command to file path
    // "app deploy" -> "cli/app/deploy.md"
    path := commandToPath(cmd)

    data, err := DocsFS.ReadFile(path)
    if err != nil {
        return nil, err
    }

    return parseCommandDoc(data)
}
```

### Terminal Rendering

Uses shared TextRenderer with ANSI colors enabled:

```go
package docs

import (
    "github.com/fatih/color"
)

func RenderHelp(doc *CommandDoc) string {
    renderer := &render.TextRenderer{
        Width:   80,
        UseANSI: true,
    }

    return renderer.RenderCommand(doc)
}
```

### CLI Integration

```go
// cmd/server/help.go

func showHelp(cmd string) {
    doc, err := docs.LoadCommand(cmd)
    if err != nil {
        // Fall back to cobra's default help
        return
    }

    fmt.Println(docs.RenderHelp(doc))
}

// Wire up to commands
var appDeployCmd = &cobra.Command{
    Use:   "deploy <directory>",
    Short: "Deploy a local directory to fazt",
    Long:  docs.LoadCommand("app deploy").Description,
    Run: func(cmd *cobra.Command, args []string) {
        // ...
    },
}

// Custom help function
func init() {
    rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
        showHelp(cmd.CommandPath())
    })
}
```

---

## 7. Testing .txt Output

### Line Width Validation

```go
func TestTxtLineWidth(t *testing.T) {
    files, _ := filepath.Glob("dist/**/*.txt")
    for _, file := range files {
        content, _ := os.ReadFile(file)
        lines := strings.Split(string(content), "\n")
        for i, line := range lines {
            if len(line) > 80 {
                t.Errorf("%s:%d: line exceeds 80 chars (%d): %s",
                    file, i+1, len(line), line[:50]+"...")
            }
        }
    }
}
```

### ASCII-Only Validation

```go
func TestTxtASCIIOnly(t *testing.T) {
    files, _ := filepath.Glob("dist/**/*.txt")
    for _, file := range files {
        content, _ := os.ReadFile(file)
        for i, b := range content {
            if b > 127 {
                t.Errorf("%s: non-ASCII byte at position %d", file, i)
            }
        }
    }
}
```

### No ANSI Codes Validation

```go
func TestTxtNoANSI(t *testing.T) {
    ansiPattern := regexp.MustCompile(`\x1b\[[0-9;]*m`)
    files, _ := filepath.Glob("dist/**/*.txt")
    for _, file := range files {
        content, _ := os.ReadFile(file)
        if ansiPattern.Match(content) {
            t.Errorf("%s: contains ANSI escape codes", file)
        }
    }
}
```

### CI Integration

```yaml
# .github/workflows/docs.yml
- name: Build docs
  run: fazt-md build knowledge-base/ --output dist/

- name: Validate .txt files
  run: go test ./internal/docs/... -run TestTxt

- name: Check line widths
  run: |
    for f in dist/**/*.txt; do
      if awk 'length > 80' "$f" | grep -q .; then
        echo "ERROR: $f has lines > 80 chars"
        exit 1
      fi
    done
```

---

## 8. Build Integration Options

Three approaches to integrating the build step.

### Option A: Manual Build + Deploy (Recommended Initially)

Explicit two-step process:

```bash
# Step 1: Render markdown to HTML + TXT
fazt-md build knowledge-base/ --output dist/docs/

# Step 2: Deploy the rendered content
fazt app deploy dist/docs/ --name docs
fazt @zyt app deploy dist/docs/ --name docs
```

**Pros**: Simple, debuggable, clear separation
**Cons**: Extra step, easy to forget

### Option B: Docs App Handles Build

The docs-theme app runs the build on deploy:

```bash
# Install docs theme (one-time)
fazt app install github:fazt-sh/docs-theme --name docs

# Configure content directory in manifest
# Then deploy - build happens automatically
fazt app deploy docs/ --name docs
```

The app's `package.json` build script runs `fazt-md`.

**Pros**: Single deploy command
**Cons**: Requires docs-theme app, more complex

### Option C: Integrated Build in fazt CLI

Fazt CLI detects markdown content and builds automatically:

```bash
# Fazt detects .md files, runs fazt-md, deploys HTML + TXT
fazt app deploy knowledge-base/ --name docs --type docs
```

**Pros**: Seamless
**Cons**: Adds complexity to core binary

### Recommendation

Start with **Option A** (manual) for clarity, then implement **Option B**
(docs app) for the streamlined experience. Option C adds too much
complexity to the core binary.

---

## 9. Critical Decisions for CLI Refactor

Decisions that must be made NOW to ensure CLI help and web docs work together.

### 9.1 Single vs Split Frontmatter

**Decision**: Single unified schema with optional sections.

CLI-specific fields (`command`, `syntax`, `arguments`, `flags`) are ignored
by web renderer for non-command docs. Web-specific fields (`description`,
`tags`, `image`) are ignored by CLI help.

### 9.2 CLI Field Rendering in Web

**Decision**: Render structured CLI fields as semantic HTML.

The `arguments`, `flags`, and `examples` frontmatter fields render as styled
HTML sections, not just metadata. This ensures CLI command pages look good
on web.

### 9.3 Web Fields in CLI Help

**Decision**: Ignore web-only fields in CLI.

Fields like `tags`, `image`, `toc`, `series` have no meaning in terminal
output. The CLI renderer simply skips them.

### 9.4 Cross-References

**Decision**: Use relative markdown links, convert to HTML paths.

```markdown
See [app list](../list.md) for listing apps.
```

Converts to:
- Web HTML: `<a href="../list.html">app list</a>`
- Web TXT: `See 'app list' (../list.txt) for listing apps.`
- CLI: `See 'fazt app list' for listing apps.`

### 9.5 Example Format

**Decision**: Structured examples in frontmatter + fenced code blocks in body.

Frontmatter examples render as formatted terminal output (CLI) or styled code
blocks (web). Fenced code blocks in the markdown body work in both contexts.

### 9.6 File Organization

**Decision**: Mirror CLI command structure.

```
knowledge-base/
├── cli/
│   ├── index.md              # CLI overview
│   ├── app/
│   │   ├── index.md          # app command overview
│   │   ├── deploy.md         # fazt app deploy
│   │   ├── list.md           # fazt app list
│   │   └── remove.md         # fazt app remove
│   ├── remote/
│   │   ├── index.md
│   │   └── add.md
│   └── server/
│       ├── index.md
│       └── start.md
├── blog/
│   ├── v018-release.md
│   └── v019-release.md
└── guides/
    ├── getting-started.md
    └── deployment.md
```

### 9.7 Binary Embedding

**Decision**: Embed markdown, not HTML.

The binary embeds raw markdown files at build time. Terminal rendering
happens at runtime. This keeps the binary smaller and allows consistent
source.

### 9.8 .txt Parity with --help

**Decision**: Shared rendering logic, different color settings.

The web `.txt` files use the same text rendering code as binary `--help`,
with ANSI colors disabled. This ensures consistent formatting and reduces
maintenance.

---

## 10. Dog-Fooding Strategy

### Phase 1: CLI Docs (v0.19.0)

Convert existing CLI documentation to the new format.

**Scope**:
- `knowledge-base/cli/` folder
- All command docs follow unified frontmatter
- Binary embeds and renders for `--help`
- Basic web rendering (single theme)
- .txt generation

**Deliverables**:
1. Blog-aligned frontmatter schema implemented
2. `fazt-md` build tool (outputs HTML + TXT)
3. CLI help reads from markdown
4. Basic docs-theme app
5. .txt files for all CLI docs

### Phase 2: Full Knowledge Base (v0.20.0)

Extend to all documentation.

**Scope**:
- `knowledge-base/architecture/`
- `knowledge-base/workflows/`
- `knowledge-base/agent-context/`
- Navigation, search, multiple themes
- Blog posts

**Deliverables**:
1. Full knowledge-base rendered
2. Sidebar navigation
3. Client-side search
4. Theme system (3+ themes)
5. Blog support with standard frontmatter

### Phase 3: External Use (v0.21.0)

Make docs-theme installable for external users.

**Scope**:
- GitHub repository: `fazt-sh/docs-theme`
- Documentation for custom themes
- Blog template variant

**Deliverables**:
1. Public docs-theme repo
2. Theme creation guide
3. Blog variant theme
4. Example sites

---

## 11. Example Rendering

### Input: `knowledge-base/cli/app/deploy.md`

```yaml
---
title: "App Deploy Command"
date: 2026-02-02
description: "Deploy apps to fazt instances"
tags: [cli, deployment, apps]
category: commands

command: "app deploy"
syntax: "fazt [@peer] app deploy <directory> [flags]"

arguments:
  - name: "directory"
    type: "path"
    required: true
    description: "Path to the directory to deploy"

flags:
  - name: "--name"
    short: "-n"
    type: "string"
    default: "directory name"
    description: "App name (overrides directory name)"

  - name: "--spa"
    type: "bool"
    default: false
    description: "Enable SPA routing for clean URLs"

  - name: "--no-build"
    type: "bool"
    default: false
    description: "Skip automatic build step"

  - name: "--include-private"
    type: "bool"
    default: false
    description: "Include gitignored private/ directory"

examples:
  - title: "Deploy to local"
    command: "fazt app deploy ./my-app"
    description: "Deploy directory to local instance"

  - title: "Deploy to remote peer"
    command: "fazt @zyt app deploy ./my-app"
    description: "Deploy to the 'zyt' peer"

  - title: "Deploy SPA"
    command: "fazt @zyt app deploy ./my-spa --spa"
    description: "Deploy with SPA routing enabled"

related:
  - command: "app list"
    description: "List deployed apps"
  - command: "app remove"
    description: "Remove a deployed app"
  - command: "app validate"
    description: "Validate app structure"
  - command: "app info"
    description: "Show app details"

errors:
  - message: "directory does not exist"
    solution: "Verify the path exists and is accessible"
  - message: "No local fazt server running"
    solution: "Start local server with 'fazt server start' or deploy to remote"
---

# Description

The `deploy` command uploads a directory to a fazt instance.

## Build Detection

If the directory contains a `package.json`, fazt automatically runs:

| Lockfile | Build Command |
|----------|---------------|
| `pnpm-lock.yaml` | `pnpm install && pnpm build` |
| `bun.lockb` | `bun install && bun run build` |
| `yarn.lock` | `yarn install && yarn build` |
| `package-lock.json` | `npm install && npm run build` |

## SPA Routing

For single-page applications, use `--spa` to enable clean URL routing.
```

### Terminal Output (`fazt app deploy --help`)

```
APP DEPLOY COMMAND

  fazt [@peer] app deploy <directory> [flags]

Deploy apps to fazt instances

ARGUMENTS:
  <directory>  (required)
      Path to the directory to deploy

FLAGS:
  -n, --name <string>
      App name (overrides directory name) (default: directory name)
  --spa
      Enable SPA routing for clean URLs
  --no-build
      Skip automatic build step
  --include-private
      Include gitignored private/ directory

EXAMPLES:
  Deploy to local:
    $ fazt app deploy ./my-app

  Deploy to remote peer:
    $ fazt @zyt app deploy ./my-app

  Deploy SPA:
    $ fazt @zyt app deploy ./my-spa --spa

See also: fazt app list, fazt app remove, fazt app validate, fazt app info
```

### Plain Text Output (`/cli/app/deploy.txt`)

See `_templates/command.txt` for the full formatted example.

### Web Output (`/cli/app/deploy.html`)

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="description" content="Deploy apps to fazt instances">
  <title>App Deploy Command - Fazt Docs</title>
  <link rel="stylesheet" href="/theme.css">
  <link rel="alternate" type="text/plain" href="/cli/app/deploy.txt">
</head>
<body>
  <nav class="sidebar">
    <div class="nav-tree">
      <a href="/cli/" class="nav-item">CLI</a>
      <div class="nav-group">
        <a href="/cli/app/" class="nav-item">app</a>
        <div class="nav-children">
          <a href="/cli/app/deploy.html" class="nav-item active">deploy</a>
          <a href="/cli/app/list.html" class="nav-item">list</a>
          <a href="/cli/app/remove.html" class="nav-item">remove</a>
        </div>
      </div>
    </div>
  </nav>

  <main class="content">
    <article class="doc doc-command">
      <header class="doc-header">
        <nav class="breadcrumb">
          <a href="/cli/">CLI</a> /
          <a href="/cli/app/">app</a> /
          <span>deploy</span>
        </nav>
        <h1>App Deploy Command</h1>
        <p class="description">Deploy apps to fazt instances</p>
        <div class="meta">
          <time datetime="2026-02-02">Updated: Feb 2, 2026</time>
          <div class="tags">
            <span class="tag">cli</span>
            <span class="tag">deployment</span>
            <span class="tag">apps</span>
          </div>
        </div>
      </header>

      <section class="command-meta">
        <h2>Synopsis</h2>
        <pre class="syntax"><code>fazt [@peer] app deploy &lt;directory&gt; [flags]</code></pre>
      </section>

      <!-- Arguments, Flags, Examples sections -->

      <footer class="doc-footer">
        <a href="/cli/app/deploy.txt" class="txt-link">View as plain text</a>
        <a href="https://github.com/fazt-sh/fazt/edit/master/knowledge-base/cli/app/deploy.md" class="edit-link">
          Edit this page
        </a>
      </footer>
    </article>
  </main>
</body>
</html>
```

---

## 12. Implementation Roadmap

### Milestone 1: Blog-Aligned Frontmatter (1 week)

- Define final YAML schema with blog fields
- Go parser for frontmatter
- Validate existing docs against schema
- Update `deploy.md` as reference

### Milestone 2: Terminal Renderer (1 week)

- Embed markdown in binary
- Parse frontmatter at runtime
- Render to terminal with colors
- Wire up `--help` flag

### Milestone 3: HTML + TXT Renderer (2 weeks)

- `fazt-md` build tool
- goldmark + chroma integration
- Template system
- .txt generation (shared with terminal)
- Navigation generation

### Milestone 4: Docs Theme App (2 weeks)

- Basic theme CSS
- Sidebar navigation
- Search functionality
- Dark mode
- .txt link in footer

### Milestone 5: Polish (1 week)

- Multiple themes
- Full knowledge-base render
- Performance optimization
- Documentation
- Blog template variant

---

## Summary

This design provides:

1. **Blog-Aligned Frontmatter** - Standard fields for theme compatibility
2. **Three Renderers** - Terminal (ANSI), HTML (themed), TXT (plain)
3. **Shared TXT Logic** - Same code for binary help and web .txt
4. **Theme System** - CSS-based theming with standard blog support
5. **LLM-Friendly** - .txt files for automated consumption
6. **Generic Design** - Not coupled to Fazt docs specifically
7. **Dog-Fooding Path** - Use for Fazt's own documentation

The system keeps CLI help fast (embedded markdown) while enabling rich
web documentation (themed HTML) and LLM-friendly plain text (.txt) from
the same source files.
