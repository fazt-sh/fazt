# Markdown Service

## Summary

Compile Markdown to HTML with syntax highlighting, shortcodes, and optional
classless CSS. Uses Goldmark (same engine as Hugo). Enables `.md` files to
be served as styled HTML pages.

## Capabilities

| Feature             | Description                          |
| ------------------- | ------------------------------------ |
| Compile             | Markdown → HTML                      |
| Syntax highlighting | Code blocks with Chroma              |
| Shortcodes          | Embed components `{{name ...}}`      |
| Classless CSS       | Optional styling (Pico, Water, etc.) |
| Auto-serve          | `.md` files served as HTML           |

## Usage

### Compile API

```javascript
const html = await fazt.services.markdown.render(content, {
  css: 'pico',              // Optional: classless CSS
  highlight: true,          // Syntax highlighting (default: true)
  toc: true,                // Generate table of contents
  shortcodes: true          // Enable shortcodes (default: true)
});
```

### Compile File

```javascript
const html = await fazt.services.markdown.renderFile('docs/guide.md', {
  css: 'pico'
});
```

### Auto-Serve (Zero Config)

Any `.md` file in VFS is automatically served as HTML:

```
Request: GET /docs/guide.md
Response: Compiled HTML with default styling
```

To customize, add `_markdown.json` in the directory:

```json
{
  "css": "pico",
  "highlight": true,
  "toc": true,
  "template": "_layout.html"
}
```

## Classless CSS Options

Pre-bundled CSS that styles semantic HTML:

| Theme    | Description                |
| -------- | -------------------------- |
| `pico`   | Minimal, elegant (default) |
| `water`  | Lightweight, clean         |
| `simple` | Basic, readable            |
| `none`   | No CSS, raw HTML           |

```javascript
await fazt.services.markdown.render(content, { css: 'water' });
```

Or custom CSS:

```javascript
await fazt.services.markdown.render(content, {
  css: '/styles/custom.css'
});
```

## Shortcodes

Embed dynamic content without JS runtime:

### Form

```markdown
# Contact Us

Fill out this form:

{{form name="contact" redirect="/thanks"}}
```

Expands to form HTML pointing at `/_services/forms/contact`.

### YouTube

```markdown
{{youtube id="dQw4w9WgXcQ"}}

{{youtube id="dQw4w9WgXcQ" width="560" height="315"}}
```

### Image with Processing

```markdown
{{image src="/photos/hero.jpg" width="800"}}

{{image src="/photos/avatar.jpg" thumb="100"}}
```

Uses media service for on-the-fly processing.

### QR Code

```markdown
Scan to visit:

{{qr data="https://example.com" size="200"}}
```

### Include

```markdown
{{include "_partials/header.md"}}

Main content here.

{{include "_partials/footer.md"}}
```

### Table of Contents

```markdown
# My Document

{{toc}}

## Section One
...
```

Generates linked TOC from headings.

## Shortcode Syntax

```
{{name}}                     Simple
{{name param="value"}}       With params
{{name param1="a" param2="b"}}  Multiple params
```

All params are strings. Shortcodes are processed before Markdown compilation.

## Custom Shortcodes

Define in `_shortcodes/` directory:

```html
<!-- _shortcodes/callout.html -->
<div class="callout callout-{{type}}">
  <strong>{{title}}</strong>
  <p>{{content}}</p>
</div>
```

Usage:

```markdown
{{callout type="warning" title="Note" content="This is important."}}
```

## Template Layout

Wrap compiled Markdown in a layout:

```html
<!-- _layout.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{title}}</title>
  {{css}}
</head>
<body>
  <main class="container">
    {{toc}}
    {{content}}
  </main>
</body>
</html>
```

Template variables:

| Variable      | Description                  |
| ------------- | ---------------------------- |
| `{{title}}`   | From frontmatter or first H1 |
| `{{content}}` | Compiled HTML                |
| `{{toc}}`     | Table of contents            |
| `{{css}}`     | CSS link tag                 |

## Frontmatter

YAML frontmatter for metadata:

```markdown
---
title: My Guide
css: water
toc: true
template: _docs-layout.html
---

# My Guide

Content here...
```

## HTTP Endpoint

### Compile on request

```
POST /_services/markdown/render
Content-Type: text/markdown

# Hello World
This is **bold**.
```

Response:

```html
<h1>Hello World</h1>
<p>This is <strong>bold</strong>.</p>
```

### With options

```
POST /_services/markdown/render?css=pico&toc=true
```

## JS API

```javascript
fazt.services.markdown.render(content, options?)
// options: { css, highlight, toc, shortcodes, template }
// Returns: string (HTML)

fazt.services.markdown.renderFile(path, options?)
// Returns: string (HTML)

fazt.services.markdown.extract(content)
// Returns: { frontmatter, body }
// Separates YAML frontmatter from content
```

## Go Libraries

```go
import (
    "github.com/yuin/goldmark"
    "github.com/alecthomas/chroma"  // Syntax highlighting
)
```

## Directory Structure Example

```
app/
├── _markdown.json           # Default options
├── _layout.html             # Default template
├── _shortcodes/
│   └── callout.html
├── index.md                 # Served as /index (or /)
├── about.md                 # Served as /about
└── docs/
    ├── _markdown.json       # Override for /docs/
    ├── guide.md             # Served as /docs/guide
    └── api.md               # Served as /docs/api
```

## Example: Documentation Site

**docs/index.md:**

```markdown
---
title: Documentation
---

# Welcome

{{toc}}

## Getting Started

See the [installation guide](./install.md).

## API Reference

See the [API docs](./api.md).
```

**docs/_markdown.json:**

```json
{
  "css": "pico",
  "highlight": true,
  "toc": true,
  "template": "_docs-layout.html"
}
```

**docs/_docs-layout.html:**

```html
<!DOCTYPE html>
<html>
<head>
  <title>{{title}} - Docs</title>
  {{css}}
</head>
<body>
  <nav>
    <a href="/">Home</a>
    <a href="/docs/">Docs</a>
  </nav>
  <main class="container">
    <aside>{{toc}}</aside>
    <article>{{content}}</article>
  </main>
</body>
</html>
```

## Comparison with Hugo

| Feature       | Hugo              | Fazt Markdown      |
| ------------- | ----------------- | ------------------ |
| Compile speed | Fast              | Fast (Goldmark)    |
| Templates     | Go templates      | Simple `{{var}}`   |
| Shortcodes    | Yes               | Yes (simpler)      |
| Themes        | Full theme system | Classless CSS only |
| Build step    | Required          | None (on-request)  |
| Complexity    | High              | Low                |

Fazt Markdown is not a Hugo replacement. It's for simple sites that
don't need a full static site generator.
