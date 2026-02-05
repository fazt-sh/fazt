# Fazt

**Sovereign compute** - Single Go binary + SQLite database that runs anywhere.

**Version**: 0.25.4 | **State**: `koder/STATE.md`

## Monorepo Structure

Fazt uses **unified versioning** - all components share the same version for guaranteed compatibility.

**Components:**
- **fazt-binary** (`internal/`) - Core Go binary [stable, 100%]
- **admin** (`admin/`) - Web admin UI [alpha, 15%]
- **fazt-sdk** (`admin/packages/fazt-sdk/`) - JavaScript API client [alpha, 20%]
- **knowledge-base/** - Documentation [stable, 80%]

**Versioning:**
- One version (0.25.4) = everything works together
- Status markers track maturity: `stable`, `beta`, `alpha`
- Completeness % shows progress towards full parity
- See `version.json` at repo root for details

## Philosophy

- **No backward compatibility** - Single user, rapidly iterating. Break things freely.
- **No legacy support** - Remove old code paths. Don't maintain deprecated features.
- **LEGACY_CODE markers** - When adding new APIs alongside old ones, mark old code with `// LEGACY_CODE: <reason>` comments. Periodically grep for these and remove them.
- **No hacks** - Build for millions of future users, not the current problem. If you see architectural dissonance, push back hard.
- **Elegant solutions** - Fix the binary, not the instance. The system should work elegantly.
- **Single DB philosophy** - Everything in SQLite. No config files. The database IS the instance.
- **Static hosting first** - Serverless is enhancement, never blocks static deploy.
- **Core vs test apps** - `admin/` is core (tracked in git). `servers/` are test/demo apps (gitignored).

## Architecture: fazt-sdk (The Core Engine)

**fazt-sdk is the foundation** - Admin UI can change tomorrow, but fazt-sdk must be rock solid. It will power:
- Admin web UI (React)
- Mobile apps (React Native)
- CLI tools (Node.js)
- Third-party integrations

**Principles:**
1. **Platform-agnostic core** - HTTP client + cache layer work anywhere (browser, Node, RN)
2. **Smart caching** - Stale-while-revalidate pattern, no count=0 flicker
3. **Request deduplication** - 10 components requesting same data = 1 network call
4. **Type safety** - Full TypeScript, strict types
5. **Framework adapters** - React hooks, Vue composables wrap the core

**Architecture:**
```
fazt-sdk/
├── core/           # Platform-agnostic (client, cache, types)
├── adapters/       # fetch, mock, react-native
└── integrations/   # React hooks (uses @tanstack/react-query)
```

**Critical rule:** Admin UI MUST use fazt-sdk, never duplicate HTTP logic. The SDK owns all API communication.

## Architecture: Config

```
Config Priority:
  CLI flags (--domain, --port)  ← temporary overrides
  ↓
  Database (configurations table) ← source of truth
  ↓
  Defaults
```

- **Database is truth**: `fazt server start --db /path/to/data.db` should be enough
- **CLI flags are overrides**: For debugging/testing, not persistent
- **No config files**: Removed. Everything in SQLite.

## Environment

| What | Value |
|------|-------|
| VM | `192.168.64.3` (headless Ubuntu) |
| Production | `zyt.app` |
| Local server | `fazt-local` systemd service |
| **Database** | `~/.fazt/data.db` (single DB for everything) |
| Binary | `~/.local/bin/fazt` |

**Database contains:** Apps, aliases, storage, auth, events, peer configs - everything.
**Override:** `FAZT_DB_PATH` env var or `--db` flag.

**IMPORTANT - zyt SSH Access:**
- `zyt.app` resolves to Cloudflare IPs (cannot SSH)
- Actual server IP: `ZYT_IP` in `.env` file (currently `165.227.11.46`)
- SSH: `ssh root@165.227.11.46`
- Service DB: `/home/fazt/.fazt/data.db` (NOT `~/.fazt/data.db`)

## Essential Commands

```bash
# Build & Test (development)
go build -o fazt ./cmd/server  # Quick iteration only
go test ./...

# Upgrade (releases & testing)
fazt upgrade                    # ALWAYS use for releases - tests user experience
fazt --version

# Deploy
fazt @zyt app deploy ./my-app
fazt @local app deploy ./my-app

# Local server
systemctl --user restart fazt-local
journalctl --user -u fazt-local -f
```

**IMPORTANT**: When releasing or testing upgrades, ALWAYS use `fazt upgrade` instead of `go build`. This ensures we test the actual user experience and verify the upgrade path works elegantly.

## Key Paths

```
cmd/server/           # Entry point, CLI
internal/
├── runtime/          # Serverless JS (Goja)
├── handlers/         # HTTP handlers
├── hosting/          # VFS, deploy logic
├── auth/             # OAuth, sessions
└── storage/          # KV, Docs, Blobs
admin/                # Official Admin UI (TRACKED)
├── packages/         # fazt-sdk, zap, fazt-ui
├── src/              # Pages, stores, routes
└── version.json      # Version tracking
koder/
├── STATE.md          # Current work
└── plans/            # Implementation plans
knowledge-base/
├── agent-context/    # Detailed dev context
├── workflows/        # Task-oriented development guides
└── skills/app/       # App development patterns
servers/              # Test/demo apps (GITIGNORED)
```

## Deep Context

Read these **as needed**, not every session:

### Architecture & Context
| File | When to Read |
|------|--------------|
| `koder/STATE.md` | Start of session - current work |
| `knowledge-base/agent-context/setup.md` | Local server setup, SSH access |
| `knowledge-base/agent-context/architecture.md` | How fazt works, app model |
| `knowledge-base/agent-context/api.md` | API endpoints, CLI commands |
| `knowledge-base/agent-context/tooling.md` | Skills, knowledge-base, releasing |
| `knowledge-base/skills/app/` | App development patterns |

### Development Workflows
| Task | Read |
|------|------|
| Add Admin UI feature | `knowledge-base/workflows/admin-ui/adding-features.md` |
| Understand UI state | `knowledge-base/workflows/admin-ui/architecture.md` |
| Test mock vs real | `knowledge-base/workflows/admin-ui/testing.md` |
| Pre-implementation check | `knowledge-base/workflows/admin-ui/checklist.md` |
| Extend fazt-sdk | `knowledge-base/workflows/fazt-sdk/extending.md` |
| Add backend API | `knowledge-base/workflows/fazt-binary/adding-apis.md` |

**Rules**:
- Always check workflows **before** implementing features to validate backend support
- Check the `updated:` date in frontmatter - if doc is >2 days old, verify info is still accurate
- If doc seems stale, update it and change the `updated:` date

## Quick Reference

**Peers**: `zyt` (production), `local` (development)

**App structure**:
```
my-app/
├── manifest.json    # { "name": "my-app" }
├── index.html
├── api/main.js      # Serverless
└── private/         # Auth-gated files
```

**Create app**:
```bash
mkdir -p servers/zyt/my-app && cd servers/zyt/my-app
echo '{"name":"my-app"}' > manifest.json
```
