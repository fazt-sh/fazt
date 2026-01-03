# Fazt Assistant Guide

**Fazt** is sovereign compute infrastructure for individuals—a single Go binary
+ SQLite database that runs anywhere from phones to servers to IoT devices.

**Read `koder/philosophy/VISION.md` first** to understand what Fazt is becoming.

## Quick Start
```
read koder/philosophy/VISION.md   # Understand the vision
read koder/start.md               # Then execute
```

## Core Philosophy
- **Cartridge Model**: One Binary (`fazt`) + One SQLite DB (`data.db`)
- **Pure Go**: `modernc.org/sqlite`, NO CGO, runs everywhere
- **Swarm Ready**: Multiple nodes mesh into personal cloud
- **AI Native**: Lowers floor (anyone can use), raises ceiling (agents)
- **Resilient**: Works when network is denied

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
- **Tables** - Narrow tables OK with aligned columns; wide tables → bullets
- **Minimal HTML** - Avoid inline HTML, use standard markdown

```
# Good: narrow table with aligned columns
| Before     | After              |
|------------|--------------------|
| Web server | Operating system   |
| One VPS    | Swarm of devices   |

# Bad: wide table (convert to bullet list instead)
| Device | Role | Long Description | Another Column | Too Wide |

# Good: wide content as bullets
- **Phone**: Mobile presence (notifications, location)
- **Laptop**: Daily driver (apps, files, local AI)
```

## Current Work
See `koder/NEXT_SESSION.md` for active task and context.