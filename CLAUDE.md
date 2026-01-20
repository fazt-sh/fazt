# Fazt Assistant Guide

**Fazt** is sovereign compute infrastructure for individuals—a single Go binary
+ SQLite database that runs anywhere from phones to servers to IoT devices.

**Current Version**: 0.10.3
**This Repo**: Source code for fazt development

## Environment Context

- **Location**: VM (headless) at `192.168.64.3`
- **OS**: Ubuntu (development machine)
- **Live Instance**: zyt.app (personal fazt server, test bed for development)

### Local Paths

| What | Path |
|------|------|
| Binary | `~/.local/bin/fazt` (in PATH) |
| Client DB | `~/.config/fazt/data.db` |
| Peers | Stored in client DB (peers table) |
| Apps | `servers/zyt/` (gitignored) |

The `fazt` command is available in PATH. Use `fazt --help` to explore commands:

```bash
fazt --help              # All commands
fazt remote --help       # Remote/peer commands
fazt server --help       # Server commands
```

## Quick Start

This file (`CLAUDE.md`) is the primary context. For deep implementation work:

```
read koder/STATE.md    # Check if a plan is active
read koder/start.md    # Deep implementation protocol
```

## Development Philosophy

**No backward compatibility.** Fazt is rapidly iterating with a single user.
We break things and evolve. Never maintain legacy patterns or compatibility shims.

**Static hosting first.** Fazt's primary goal is being a self-hostable Surge
alternative. Static file hosting must work perfectly with zero build steps.
All other features (serverless, apps model) are progressive enhancements that
should never get in the way of simple static file deployment.

**Apps are throw-away.** Apps in `servers/zyt/` exist only to exercise and
refine fazt. When an app hits a bug or limitation, fix fazt - never work around
it in the app. The app is the test case; fazt is the product.

**Simple nomenclature:**
- A static site is called an **app**
- A subdomain is called an **alias**

## App Development

An **app** in fazt is a website with optional serverless capabilities.

### Where to Store Apps (This Repo)

```
servers/                 # gitignored - NOT part of fazt source
└── zyt/                 # Apps for zyt.app
    ├── xray/            # An app
    └── my-new-app/      # Another app
```

**Why gitignored?**
- Apps are instance-specific, not fazt source code
- Each developer has different instances/apps

Develop in `servers/zyt/`, test locally, deploy to zyt when ready.

### App Structure

```
my-app/
├── manifest.json      # Required: { "name": "my-app" }
├── index.html         # Entry point
├── static/            # Assets (css, js, images)
└── api/               # Serverless functions (*.js)
    └── hello.js       # → GET /api/hello
```

### Configured Peers

| Name | URL | Description |
|------|-----|-------------|
| zyt | https://zyt.app | Personal production instance |
| local | http://192.168.64.3:8080 | Local dev server (always running) |

**Note**: The local server is a persistent instance - always running across all
Claude sessions. Never stop it. Use it to test apps before deploying to zyt.

**Managing peers:**
```bash
fazt remote list                    # List all peers
fazt remote status zyt              # Health, version, uptime
fazt remote upgrade zyt             # Upgrade to latest version
```

**Managing apps:**
```bash
fazt app list zyt                   # List deployed apps
fazt app deploy <dir> --to zyt      # Deploy from local directory
fazt app install <url> --to zyt     # Install from GitHub
fazt app upgrade <app>              # Upgrade git-sourced app
fazt app pull <app> --to ./local    # Download app files
fazt app info <app>                 # Show app details
fazt app remove <app> --from zyt    # Remove an app
```

### Development Workflow

#### Quick Deploy (Static Only)

For apps without serverless (`/api`) endpoints, use a simple HTTP server:

```bash
python3 -m http.server 7780 --directory servers/zyt/my-app
# Access at http://192.168.64.3:7780
```

#### Full Local Server (With Serverless)

To test `/api` endpoints locally, run a local fazt server. This requires
building the binary with embedded system sites.

**1. Build fazt with embedded admin:**
```bash
# Build admin dashboard first
npm run build --prefix admin

# Copy to embed location
cp -r admin/dist internal/assets/system/admin

# Build fazt binary
go build -o fazt ./cmd/server
```

**2. Initialize local server (first time only):**
```bash
# Creates DB at servers/local/data.db (persists across reboots)
mkdir -p servers/local
fazt server init \
  --username dev \
  --password dev \
  --domain 192.168.64.3 \
  --db servers/local/data.db
```

**3. Create API key and add as peer:**
```bash
# Generate API key
fazt server create-key --db servers/local/data.db
# Save the token output

# Add local peer (if not already added)
fazt remote add local \
  --url http://192.168.64.3:8080 \
  --token <API_KEY>
```

**4. Start local server:**
```bash
fazt server start \
  --port 8080 \
  --domain 192.168.64.3 \
  --db servers/local/data.db
```

**5. Deploy and test:**
```bash
fazt remote deploy servers/zyt/my-app local
# Access at http://my-app.192.168.64.3:8080
# Or use curl with Host header:
curl -H "Host: my-app.192.168.64.3" http://192.168.64.3:8080/api/hello
```

#### Deploy to Production

```bash
fazt remote deploy servers/zyt/my-app zyt
```
App available at: `https://my-app.zyt.app`

#### Create New App

```bash
mkdir -p servers/zyt/my-app && cd servers/zyt/my-app
echo '{"name":"my-app"}' > manifest.json
echo '<h1>Hello</h1>' > index.html
```

### Serverless (Current)

JavaScript files in `api/` are executed server-side via Goja:

```javascript
// api/hello.js
function handler(req) {
  return {
    status: 200,
    body: JSON.stringify({ message: "Hello", time: Date.now() })
  };
}
```

Access: `https://my-app.zyt.app/api/hello`

**Limitations**: Basic JS only, no npm modules, no async/await yet.
See `koder/ideas/specs/v0.10-runtime/` for future enhancements.

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
| `/fazt-start` | Begin work session |
| `/fazt-app` | Build and deploy apps with Claude |
| `/fazt-release` | Release new version (full workflow) |

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/deploy` | POST | Deploy ZIP archive |
| `/api/apps` | GET | List apps |
| `/api/apps/{id}` | GET/DELETE | App details/delete |
| `/api/apps/{id}/source` | GET | App source tracking |
| `/api/apps/{id}/files` | GET | List app files |
| `/api/apps/{id}/files/{path}` | GET | Get file content |
| `/api/upgrade` | POST | Upgrade server |
| `/health` | GET | Health check |

## Build & Test

```bash
go build -o fazt ./cmd/server           # Build (basic)
go test -v -cover ./...                  # Test all with coverage
```

**Build with embedded admin** (required for local server):
```bash
npm run build --prefix admin
cp -r admin/dist internal/assets/system/admin
go build -o fazt ./cmd/server
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
