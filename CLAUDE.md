# Fazt Assistant Guide

**Fazt** is sovereign compute infrastructure for individuals—a single Go binary
+ SQLite database that runs anywhere from phones to servers to IoT devices.

**Current Version**: 0.9.0
**This Repo**: Source code for fazt development

## Environment Context

- **Location**: VM (headless) at `192.168.64.3`
- **OS**: Ubuntu (development machine)
- **Fazt Binary**: Should be built locally (`go build -o fazt ./cmd/server`)
- **Live Instance**: zyt.app (personal fazt server, test bed for development)
- **Servers Dir**: `servers/` (gitignored) holds instance configs

## Quick Start

```
read koder/start.md    # Bootstrap instructions
read koder/STATE.md    # Current progress
```

## Current Capabilities (v0.9.x)

| Feature | Status | Description |
|---------|--------|-------------|
| Static Hosting | Done | VFS-backed site hosting |
| Multi-site | Done | Subdomain routing |
| Admin Dashboard | Done | React SPA at admin.* |
| Analytics | Done | Event tracking + dashboard |
| Serverless Runtime | Done | JavaScript via Goja |
| Peers Table | Done | All config in SQLite |
| `fazt remote` | Done | Native node-to-node communication |
| MCP Server | Done | 5 tools for Claude Code |
| Remote Upgrade | Done | `/api/upgrade` endpoint |
| Claude Skills | Done | `/fazt-*` commands |

## Future Roadmap (in `koder/ideas/specs/`)

Many features are spec'd but not implemented:

- **v0.9**: Storage layer (blobs, documents)
- **v0.10**: Runtime enhancements (stdlib, sandbox)
- **v0.11**: Distribution (marketplace, manifest)
- **v0.12**: Agentic (AI harness, ai-shim)
- **v0.13**: Network (domains, VPN)
- **v0.14**: Security (RLS, notary, halt)
- **v0.15**: Identity (persona)
- **v0.16**: Mesh (P2P, protocols)
- **v0.17-v0.20**: WebSocket, Email, Workers, Services

Read `koder/ideas/ROADMAP.md` and `koder/ideas/SURFACE.md` for details.

## Managing Fazt: Skills vs MCP

There are two ways to manage fazt instances:

### Claude Skills (`.claude/commands/fazt-*.md`)
- Human-readable prompts that guide Claude Code
- Work via HTTP/curl to fazt API
- Portable: can be copied to any project

### MCP Server (`internal/mcp/`)
- Machine protocol for tool integration
- Configured via `.mcp.json` (gitignored)
- Tighter integration, type-safe

**Current Approach**: Both exist. Skills are simpler and portable. MCP is more
powerful but requires server running. Evaluate based on use case.

## Instance Management

### Server Config Structure

```
servers/                  # gitignored
├── zyt/
│   ├── config.json       # { url, domain, token }
│   └── xray/             # Sites for this instance
└── local/
    └── config.json
```

### Management Skills

| Skill | Description |
|-------|-------------|
| `/fazt-status` | Check server health |
| `/fazt-apps` | List/manage apps |
| `/fazt-deploy` | Deploy site/app |
| `/fazt-upgrade` | Check/perform upgrades |

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/deploy` | POST | Deploy ZIP archive |
| `/api/apps` | GET | List apps |
| `/api/apps/{id}` | GET/DELETE | App details/delete |
| `/api/upgrade` | POST | Upgrade server |
| `/health` | GET | Health check |

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
├── clientconfig/     # Client config (~/.fazt/config.json)
├── mcp/              # MCP server for Claude Code
├── runtime/          # JS serverless runtime
└── analytics/        # Event buffering
admin/                # React SPA (Vite + Tailwind)
koder/
├── start.md          # Bootstrap entry point
├── STATE.md          # Implementation progress
├── plans/            # Implementation plans
├── ideas/            # Specs for future versions
└── philosophy/       # Vision docs
```

## Documentation Maintenance

**Keep docs up to date**:

| Doc | Purpose | Update When |
|-----|---------|-------------|
| `CHANGELOG.md` | Version history | Every release |
| `koder/STATE.md` | Implementation progress | After significant work |
| `CLAUDE.md` | Assistant context | When capabilities change |
| `koder/ideas/*.md` | Feature specs | When planning new features |

## Core Philosophy

- **Cartridge Model**: One Binary (`fazt`) + One SQLite DB (`data.db`)
- **Pure Go**: `modernc.org/sqlite`, NO CGO, runs everywhere
- **Swarm Ready**: Multiple nodes mesh into personal cloud
- **AI Native**: Lowers floor (anyone can use), raises ceiling (agents)
- **Resilient**: Works when network is denied

## API Response Format

```go
// Success
api.Success(w, http.StatusOK, data)

// Errors
api.BadRequest(w, "message")
api.ValidationError(w, "message", "field", "constraint")
api.InternalError(w, err)
```

## Markdown Style

All markdown files must be readable in raw format (terminal, vim, cat):

- **80 character width** - Wrap prose at 80 chars, code can extend
- **Blank lines** - Before/after headings, between paragraphs
- **Short lines** - One sentence per line when possible (helps diffs)
- **Tables** - Narrow tables OK; wide tables → bullet lists
- **Minimal HTML** - Avoid inline HTML, use standard markdown
