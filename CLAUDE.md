# Fazt Assistant Guide

**Fazt** is sovereign compute infrastructure for individuals—a single Go binary
+ SQLite database that runs anywhere from phones to servers to IoT devices.

## Quick Start

```
read koder/start.md    # Bootstrap instructions
read koder/STATE.md    # Current progress
```

That's it. `start.md` tells you what to do. `STATE.md` tells you where we are.

## Core Philosophy

- **Cartridge Model**: One Binary (`fazt`) + One SQLite DB (`data.db`)
- **Pure Go**: `modernc.org/sqlite`, NO CGO, runs everywhere
- **Swarm Ready**: Multiple nodes mesh into personal cloud
- **AI Native**: Lowers floor (anyone can use), raises ceiling (agents)
- **Resilient**: Works when network is denied

## Build & Test

```bash
go build -o fazt ./cmd/server           # Build
go test -v -cover ./...                  # Test all with coverage
go run ./cmd/server server start --port 8080  # Run dev
```

**Always run tests before starting new work.**

## Key Directories

```
cmd/server/           # Entry point, CLI commands
internal/
├── handlers/         # HTTP handlers
├── api/              # Response helpers
├── hosting/          # VFS, Deploy logic
├── database/         # SQLite init, migrations
├── config/           # Server configuration
├── clientconfig/     # Client config (~/.fazt/config.json) [NEW]
├── mcp/              # MCP server for Claude Code [NEW]
├── runtime/          # JS serverless runtime [NEW]
└── analytics/        # Event buffering
admin/                # React SPA (Vite + Tailwind)
koder/
├── start.md          # Bootstrap entry point
├── STATE.md          # Implementation progress tracker
├── plans/            # Implementation plans
└── philosophy/       # Vision docs
```

## API Response Format

```go
// Success
api.Success(w, http.StatusOK, data)

// Errors
api.BadRequest(w, "message")
api.ValidationError(w, "message", "field", "constraint")
api.InternalError(w, err)
```

## Implementation Tracking

Progress is tracked in `koder/STATE.md`:
- Current phase and status
- Completed tasks (checkboxes)
- Next actions
- Blockers

**Update STATE.md after every significant step.**

## Current Plan

See `koder/plans/16_implementation-roadmap.md` for the active implementation plan:
- Phase 0: Verify current deploy
- Phase 0.5: Multi-server config
- Phase 1: MCP Server
- Phase 2: Serverless Runtime
- Phase 3: Analytics App
- Phase 4: Sites → Apps Migration
- Phase 5-8: Release, Deploy, Setup

## Environment

- Git credentials available
- Deploy target: Ubuntu 24.04 x86
- Live server: https://zyt.app

## Markdown Style

All markdown files must be readable in raw format (terminal, vim, cat):

- **80 character width** - Wrap prose at 80 chars, code can extend
- **Blank lines** - Before/after headings, between paragraphs
- **Short lines** - One sentence per line when possible (helps diffs)
- **Tables** - Narrow tables OK; wide tables → bullet lists
- **Minimal HTML** - Avoid inline HTML, use standard markdown
