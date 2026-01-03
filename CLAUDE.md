# Fazt.sh Assistant Guide

**Fazt** is a personal PaaS in a single Go binary - deploy static sites and serverless functions to your own VPS with zero dependencies.

## Quick Start
```
read and execute koder/start.md
```

## Core Philosophy
- **Cartridge Architecture**: One Binary (`fazt`) + One SQLite DB (`data.db`)
- **Zero Dependencies**: Pure Go + `modernc.org/sqlite`. NO CGO.
- **VFS**: Sites/assets live in DB, not filesystem
- **Admin SPA**: Embedded React 18 + Vite + Tailwind CSS app (`admin/`)
- **Safety**: `CGO_ENABLED=0` always

## Build & Test
```bash
go build -o fazt ./cmd/server          # Build
go test ./...                           # Test all
go run ./cmd/server server start --port 8080  # Run dev
```

## Key Directories
- `cmd/server/` - Entry point, CLI
- `internal/handlers/` - HTTP handlers
- `internal/api/` - Response helpers
- `internal/hosting/` - VFS, Deploy
- `admin/` - React SPA source (See `koder/plans/14_admin-spa-complete.md`)
- `koder/plans/` - Implementation plans

## API Response Format
```go
// Success
api.Success(w, http.StatusOK, data)

// Errors
api.BadRequest(w, "message")
api.ValidationError(w, "message", "field", "constraint")
api.InternalError(w, err)
```

## Environment
- Running in podman container (no systemd)
- Git credentials available
- Deploy target: Ubuntu 24.04 x86

## Markdown Style

All markdown files must be readable in raw format (terminal, vim, cat):

- **80 character width** - Wrap prose at 80 chars, code can extend
- **Blank lines** - Before/after headings, between paragraphs, around code blocks
- **Short lines** - One sentence per line when possible (helps diffs)
- **No wide tables** - Prefer bullet lists or multiple smaller tables
- **Minimal HTML** - Avoid inline HTML, use standard markdown

```
# Good: readable raw
This is a paragraph that wraps at 80 characters for easy
reading in any terminal or text editor.

# Bad: unreadable raw
This is a very long line that goes on and on and requires horizontal scrolling to read which makes it hard to review in pull requests or read in a terminal.
```

## Current Work
See `koder/NEXT_SESSION.md` for active task and context.