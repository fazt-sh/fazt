# Plan 32: Docs Rendering System

**Status**: Design Complete
**Created**: 2026-02-02
**Depends On**: Plan 31 (CLI Refactor - for finalized markdown docs)

## Summary

Build a generic markdown→HTML rendering system that deploys documentation as a fazt app with themes. Supports both CLI docs and general content (blog posts, guides). Dog-food Fazt's own documentation while keeping it generic for any content site.

## Motivation

**Current state:**
- Documentation scattered (README, knowledge-base/, help text in code)
- No web docs for CLI commands
- Help text hardcoded in Go strings
- No single source of truth

**With docs rendering:**
- knowledge-base/ markdown files are source of truth
- Render to beautiful web docs (themed HTML)
- Deploy as fazt app at docs.fazt.app
- Multiple themes (light/dark/minimal)
- Works for any markdown content (not just Fazt docs)

## Design

### Two-Format Output

```
Markdown file (source)
  ├→ HTML: Themed, styled, for browsers
  └→ TXT: Plain text, for LLMs/CLI (markdown minus frontmatter)
```

**URL structure:**
```
docs.fazt.app/cli/app/deploy        → HTML (default)
docs.fazt.app/cli/app/deploy.html   → HTML (explicit)
docs.fazt.app/cli/app/deploy.txt    → Plain text (markdown content)
```

### Frontmatter Schema (Blog-Aligned)

**Standard blog fields** (works with Jekyll/Hugo/Gatsby themes):
```yaml
---
title: "App Deploy Command"
date: 2026-02-02
description: "Deploy apps to fazt instances"
tags: [cli, deployment, apps]
category: commands
author: fazt
draft: false
---
```

**CLI-specific extensions** (optional, ignored by blog themes):
```yaml
command: "app deploy"
syntax: "fazt [@peer] app deploy <directory> [flags]"
peer:
  supported: true
  local: true
  remote: true
```

**Key decisions:**
- No `version` field (use monorepo version.json)
- Use `date` for last update (standard blog field)
- Singular `category`, array `tags`
- Standard fields enable theme reuse

### Markdown Guidelines

- GitHub Flavored Markdown (GFM)
- Keep lines <80 chars where reasonable (for readability)
- Use code blocks liberally (fine for TXT files)
- Clean formatting (works as plain text)
- Primary purpose: beautiful docs website

**The TXT file IS the markdown content** (minus frontmatter). No separate rendering needed.

### Build Process

```bash
# Render markdown to HTML + TXT
fazt-md build knowledge-base/ --output dist/

# For each .md file:
#   1. Parse frontmatter
#   2. Render markdown → HTML (goldmark)
#   3. Strip frontmatter → TXT (copy content)
#   4. Apply theme to HTML
```

**Output:**
```
dist/
├── cli/
│   ├── app/
│   │   ├── deploy.html    # Themed HTML
│   │   └── deploy.txt     # Markdown content (no frontmatter)
├── _metadata.json         # Navigation structure
├── _nav.json              # Sidebar navigation
└── _search.json           # Search index
```

### HTML Features

**Page structure:**
```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>App Deploy - Fazt Docs</title>
  <meta name="description" content="Deploy apps to fazt">
  <link rel="stylesheet" href="/theme.css">
</head>
<body>
  <nav class="sidebar"><!-- Auto-generated --></nav>

  <main class="content">
    <article class="doc">
      <!-- Rendered markdown -->
    </article>

    <footer class="doc-footer">
      <a href="deploy.txt">View as plain text</a>
      <button onclick="copyMarkdown()">📋 Copy markdown</button>
      <a href="https://github.com/fazt-sh/fazt/blob/master/knowledge-base/cli/app/deploy.md">
        View source
      </a>
    </footer>
  </main>
</body>
</html>
```

**Copy markdown button:**
```javascript
async function copyMarkdown() {
  const url = window.location.pathname.replace('.html', '.txt');
  const response = await fetch(url);
  const markdown = await response.text();
  await navigator.clipboard.writeText(markdown);
  showToast('Copied to clipboard!');
}
```

The button fetches the `.txt` file, which IS the markdown content!

### Docs Theme App

**Installation:**
```bash
fazt app install github:fazt-sh/docs-theme --name docs
```

**Configuration (manifest.json):**
```json
{
  "name": "docs-theme",
  "type": "docs",
  "build": {
    "command": "npm run build",
    "input": "../knowledge-base",
    "output": "dist"
  },
  "theme": {
    "name": "fazt-docs",
    "dark_mode": true,
    "syntax_theme": "monokai",
    "sidebar": true,
    "search": true
  }
}
```

**Theme system:**
```
docs-theme/
├── manifest.json
├── package.json
├── src/
│   ├── build.js           # Markdown → HTML renderer
│   ├── nav.js             # Navigation generator
│   └── search.js          # Client-side search
├── themes/
│   ├── fazt-docs.css
│   ├── minimal.css
│   └── dark.css
└── templates/
    └── page.html          # HTML template
```

**Multiple themes available:**
- fazt-docs (default)
- minimal (clean, simple)
- dark (dark mode first)
- blog (blog-style layout)

### TXT File Details

**Format:** Markdown content with frontmatter stripped

**Why TXT, not MD?**
- Opens as plain text in browsers (not downloaded)
- Signals "readable text" vs "source file"
- LLMs/CLI tools expect .txt
- Lower token cost for AI agents

**The TXT file IS markdown!** Just without the YAML frontmatter block.

**Users can:**
- View in browser
- Curl for CLI reading
- Give to LLMs (efficient)
- Copy for their docs (via copy button or direct download)

## Implementation

### Phase 1: Markdown Renderer (2-3 days)

**Create `fazt-md` CLI tool:**
```bash
fazt-md build <input> --output <output> [--theme <name>]
```

**Features:**
- Parse YAML frontmatter (gopkg.in/yaml.v3)
- Render markdown → HTML (goldmark)
- Syntax highlighting (chroma)
- Strip frontmatter → TXT
- Generate navigation metadata
- Apply theme CSS

### Phase 2: Docs Theme App (2-3 days)

**Create GitHub repo: fazt-sh/docs-theme**

**Features:**
- Multiple CSS themes
- Sidebar navigation
- Client-side search
- Dark mode toggle
- Copy markdown button
- Responsive design

**Build integration:**
- Detects markdown files
- Runs fazt-md renderer
- Outputs HTML + TXT
- Copies theme assets

### Phase 3: Deploy Fazt Docs (1-2 days)

**Dog-fooding:**
```bash
cd docs-theme
npm install

# Point to Fazt knowledge-base
npm run build -- --input ../fazt/knowledge-base --output dist

# Deploy
fazt app deploy dist --name docs
fazt @zyt app deploy dist --name docs
```

**Result:**
- http://docs.192.168.64.3.nip.io:8080 (local)
- https://docs.zyt.app (production)

### Phase 4: Theme Polish (1-2 days)

- Multiple themes
- Improved navigation
- Search functionality
- Mobile responsive
- Performance optimization

**Total: 6-10 days**

## Use Cases

### CLI Documentation
```bash
# Build CLI docs
fazt-md build knowledge-base/cli/ --output dist/cli/

# Deploy
fazt app deploy dist/cli/ --name cli-docs
```

### General Documentation
```bash
# Build full knowledge-base
fazt-md build knowledge-base/ --output dist/

# Deploy
fazt app deploy dist/ --name docs
```

### Blog Content
```bash
# User's blog content
fazt-md build ./posts/ --output dist/ --theme blog

# Deploy
fazt @prod app deploy dist/ --name blog
```

### External Use
Anyone can:
1. Install docs-theme app
2. Point it at their markdown folder
3. Get beautiful docs site
4. Choose from multiple themes

## Success Criteria

- [ ] fazt-md renderer built and working
- [ ] Outputs HTML + TXT from markdown
- [ ] Docs theme app installable from GitHub
- [ ] Multiple themes available
- [ ] Fazt docs running on fazt (dog-fooding)
- [ ] Copy markdown button works
- [ ] Navigation and search functional
- [ ] Mobile responsive
- [ ] Generic enough for any markdown content

## Related

- **Plan 31**: CLI Refactor (creates knowledge-base/cli/ docs)
- **THINKING_DIRECTIONS.md**: D1 - Documentation Overhaul

## Dog-Fooding Benefits

**Validates Fazt capabilities:**
- Static hosting (HTML files)
- Build detection (npm run build)
- Custom domains (docs.zyt.app)
- Deployment workflow

**Showcases features:**
- Fast static hosting
- Clean URLs
- Theme system
- Multi-format content

## Future Enhancements

- Server-side search (serverless function)
- Versioned docs (v0.17, v0.18, etc.)
- API reference auto-generation
- Interactive examples
- PDF export
- Dark mode persistence
- Table of contents
- Reading time estimates
- Last updated timestamps

## Notes

- Generic system, not coupled to Fazt docs
- Blog themes should work with minimal changes
- TXT files are markdown minus frontmatter
- Copy button fetches TXT (which is markdown!)
- Frontmatter optional for general content
- CLI fields ignored by blog themes
