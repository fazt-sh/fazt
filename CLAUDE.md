# Fazt Assistant Guide

**Fazt** is sovereign compute infrastructure for individuals—a single Go binary
+ SQLite database that runs anywhere from phones to servers to IoT devices.

**Current Version**: 0.11.9
**This Repo**: Source code for fazt development

## Uniform Peers

Every fazt instance is a first-class peer. There's no "dev" vs "production"
distinction - just peers that happen to run in different locations.

**Domain handling is automatic:**

| Domain Type | Behavior |
|-------------|----------|
| Real domain (`zyt.app`) | Trusted - never modified |
| Wildcard DNS (`*.nip.io`) | Auto-updates if IP changes |
| IP address | Auto-updates if machine changes |
| Empty | Auto-detects local IP |

This means:
- Same binary, same commands everywhere
- Copy `data.db` to another machine - domain auto-adjusts
- No environment variables to remember
- Real domains are always respected

## Environment Context

- **Location**: VM (headless) at `192.168.64.3`
- **OS**: Ubuntu (development machine)
- **Live Instance**: zyt.app (personal fazt server, test bed for development)

### Remote Server Access

Production instances run behind Cloudflare. For SSH access, use the actual
server IP (not the domain). IPs are stored in `.env`:

```bash
source .env
ssh root@$ZYT_IP   # SSH into production server
```

The VM has SSH access to remote servers. Use this for emergency recovery
or manual deployments when the `fazt remote upgrade` command can't reach
the server.

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
read koder/STATE.md      # Current state, what's in progress
read koder/CAPACITY.md   # Performance limits, real-time capacity
read koder/start.md      # Deep implementation protocol
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

**Note**: The local server runs as a **user systemd service** (`fazt-local`).
It auto-starts on boot and persists across sessions. Never stop it manually.

**Local server commands:**
```bash
systemctl --user status fazt-local    # Check status
systemctl --user restart fazt-local   # Restart after rebuild
journalctl --user -u fazt-local -f    # View logs
```

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

**2. Install local server (first time only):**

Use the unified install script:
```bash
./install.sh  # Select option 2: Local Development
```

Or manually:
```bash
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

If installed via `install.sh`, the server runs as a systemd user service:
```bash
systemctl --user start fazt-local   # Start
systemctl --user status fazt-local  # Check status
```

Or manually (for one-off testing):
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

## Current Capabilities (v0.11.x)

| Feature | Status |
|---------|--------|
| Static Hosting | ✅ VFS-backed, subdomain routing |
| Admin Dashboard | ✅ React SPA at admin.* |
| Serverless Runtime | ✅ JavaScript via Goja |
| OAuth (Google) | ✅ App-level user auth |
| Storage API | ✅ KV, Docs, Blobs |
| Analytics | ✅ Event tracking |
| Remote Management | ✅ `fazt remote` CLI |
| Security | ✅ Slowloris protection, rate limiting |

## Roadmap

See `koder/ideas/ROADMAP.md` for future specs (v0.12+).

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

## Releasing

Use `/fazt-release` skill, which calls `scripts/release.sh` for the heavy lifting.

```bash
source .env                    # loads GITHUB_PAT_FAZT
./scripts/release.sh vX.Y.Z    # build all platforms, upload to GitHub
```

Fast local release (~30s) vs GitHub Actions (~4min).

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
├── STATE.md              # Current work (lightweight)
├── THINKING_DIRECTIONS.md # Strategic directions to explore
├── CAPACITY.md           # Performance limits
├── plans/                # Implementation plans
├── ideas/                # Specs for future versions
└── philosophy/           # Vision docs
```

## Documentation Maintenance

| Doc | Purpose | Update When |
|-----|---------|-------------|
| `CLAUDE.md` | Stable reference | Capabilities change |
| `koder/STATE.md` | Current work | Each session |
| `koder/THINKING_DIRECTIONS.md` | Strategic ideas | New directions emerge |
| `koder/CAPACITY.md` | Performance data | After benchmarks |
| `CHANGELOG.md` | Version history | Each release |

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
