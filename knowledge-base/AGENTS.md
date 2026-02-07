# Fazt

**Sovereign compute** - Single Go binary + SQLite database that runs anywhere.

**Version**: 0.27.0 | **State**: `koder/STATE.md`

## Monorepo Structure

Fazt uses **unified versioning** - all components share the same version for guaranteed compatibility.

**Components:**
- **fazt-binary** (`internal/`) - Core Go binary [stable, 100%]
- **admin** (`admin/`) - Vue 3 + Pinia + Vite admin UI, deployed as fazt-app (BFBB) [alpha, 15%]
- **fazt-sdk** (`admin/packages/fazt-sdk/`) - Framework-agnostic JS API client [alpha, 20%]
- **knowledge-base/** - Documentation [stable, 80%]

**Versioning:**
- One version (0.26.0) = everything works together
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

## Architecture: fazt-sdk

**fazt-sdk is the foundation** — framework-agnostic JavaScript API client.
Currently powers the Admin UI; planned to also serve fazt-apps (see Plan 43).

**Current state** (`admin/packages/fazt-sdk/`):
- 737 lines, zero dependencies, pure JS
- HTTP client + mock adapter + JSDoc types
- Admin API namespaces: auth, apps, aliases, system, stats, events, logs
- Admin UI uses it exclusively via Pinia stores — no direct `fetch()` anywhere

**Architecture:**
```
fazt-sdk/
├── index.js        # Client factory + admin API namespaces
├── client.js       # HTTP adapter (fetch-based, framework-agnostic)
├── mock.js         # Mock adapter for development (?mock=true)
├── types.js        # JSDoc type definitions
└── fixtures/       # Mock response data (7 JSON files)
```

**Planned evolution** (Plan 43): Extend with `app.js` namespace for fazt-apps
(auth, uploads with progress, pagination helpers). One SDK, two surfaces — admin
and app — gated by server-side credentials.

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
- default path: "~/.fazt/data.db"
- **CLI flags are overrides**: For debugging/testing, not persistent
- **No config files**: Removed. Everything in SQLite.

## Environment

 | What         | Value                                        |
 | ------       | -------                                      |
 | VM           | `192.168.64.3` (headless Ubuntu)             |
 | Production   | `zyt.app`                                    |
 | Local server | `fazt-local` systemd service                 |
 | **Database** | `~/.fazt/data.db` (single DB for everything) |
 | Binary       | `~/.local/bin/fazt`                          |

**Database contains:** Apps, aliases, storage, auth, events, peer configs - everything.
**Override:** `FAZT_DB_PATH` env var or `--db` flag.

**IMPORTANT - zyt SSH Access:**
- `zyt.app` resolves to Cloudflare IPs (cannot SSH)
- Actual server IP: `ZYT_IP` in `.env` file
- SSH: `ssh root@<ZYT_IP>`
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
admin/                # Official Admin UI — Vue 3 + Pinia (TRACKED)
├── packages/         # fazt-sdk, fazt-ui
├── src/              # Vue pages, Pinia stores, router
└── version.json      # Version tracking
koder/
├── STATE.md          # Current work
├── issues/           # tracks issues
└── plans/            # Implementation plans
knowledge-base/
├── AGENTS.md         # This file (symlinked to root as CLAUDE.md & AGENTS.md)
├── agent-context/    # Detailed dev context
├── workflows/        # Task-oriented development guides
└── skills/           # All skills (git-tracked)
    ├── app/          # fazt-app (global skill)
    ├── release/      # fazt-release (local skill)
    ├── open/         # fazt-open (local skill)
    ├── close/        # fazt-close (local skill)
    ├── ideate/       # fazt-ideate (local skill)
    └── lite-extract/ # fazt-lite-extract (local skill)
.claude/skills/       # Symlinks to knowledge-base/skills/ (Claude discovery)
.agents/skills/       # Symlinks to knowledge-base/skills/ (Codex discovery)
servers/              # Test/demo apps (GITIGNORED)
```

## Deep Context

Read these **as needed**, not every session:

### Architecture & Context
 | File                                           | When to Read                      |
 | ------                                         | --------------                    |
 | `koder/STATE.md`                               | Start of session - current work   |
 | `knowledge-base/agent-context/setup.md`        | Local server setup, SSH access    |
 | `knowledge-base/agent-context/architecture.md` | How fazt works, app model         |
 | `knowledge-base/agent-context/api.md`          | API endpoints, CLI commands       |
 | `knowledge-base/agent-context/tooling.md`      | Skills, knowledge-base, releasing |
 | `knowledge-base/skills/app/`                   | App development patterns          |

### Development Workflows
 | Task                     | Read                                                   |
 | ------                   | ------                                                 |
 | Add Admin UI feature     | `knowledge-base/workflows/admin-ui/adding-features.md` |
 | Understand UI state      | `knowledge-base/workflows/admin-ui/architecture.md`    |
 | Test mock vs real        | `knowledge-base/workflows/admin-ui/testing.md`         |
 | Pre-implementation check | `knowledge-base/workflows/admin-ui/checklist.md`       |
 | Extend fazt-sdk          | `knowledge-base/workflows/fazt-sdk/extending.md`       |
 | Add backend API          | `knowledge-base/workflows/fazt-binary/adding-apis.md`  |

**Rules**:
- Always check workflows **before** implementing features to validate backend support
- Check the `updated:` date in frontmatter - if doc is >2 days old, verify info is still accurate
- If doc seems stale, update it and change the `updated:` date
- Plans use sequential numbering: `koder/plans/NN_short_label.md` (check highest existing number, increment)
- Issues use same pattern: `koder/issues/NN_short_label.md`

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
