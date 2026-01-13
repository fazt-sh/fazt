# Fazt - Session Bootstrap

**Fazt**: Sovereign compute infrastructure in one Go binary + SQLite database.
Runs anywhere from phones to servers—your personal cloud OS.

---

## Quick Start

```
1. Read this file (you're doing it)
2. Read koder/STATE.md → know where we are
3. Read the plan referenced in STATE.md
4. Execute the current phase
5. Update STATE.md after each step
```

---

## Session Protocol

### Step 1: Read State

```
Read: koder/STATE.md
```

This tells you:
- Current phase and status
- What's completed
- What to do next
- Any blockers or user actions needed

### Step 2: Load Context

Based on current phase, read relevant files:

| Phase | Context Files |
|-------|---------------|
| 0-4 | `koder/plans/16_implementation-roadmap.md` |
| 5 (Release) | Plan + `CHANGELOG.md` + `.github/workflows/` |
| 6-8 (Deploy/Setup) | Plan + STATE.md blockers |

### Step 3: Verify Existing Tests Pass

**Before ANY new work**, run the existing test suite:

```bash
go test -v ./...
```

If tests fail:
1. Fix them FIRST
2. Update STATE.md with what was fixed
3. Then proceed with new work

This ensures we never build on a broken foundation.

### Step 4: Execute

Follow the plan for current phase:
1. Implement code
2. Write tests
3. Run tests: `go test ./...`
4. Verify coverage is good
5. Update STATE.md with progress

### Step 5: Update State

After EVERY significant step, update `koder/STATE.md`:
- Mark completed items with `[x]`
- Update "Current Task"
- Update "Next Actions"
- Add to "Session Log"

### Step 6: Handle Blockers

If blocked on user action:
1. Update STATE.md "User Actions Required"
2. Update "Status" to indicate blocked
3. Tell user what they need to do
4. Stop and wait

---

## Architecture Reference

```
fazt (binary)
├── cmd/server/main.go      # Entry point, CLI
├── internal/
│   ├── handlers/           # HTTP handlers
│   ├── hosting/            # VFS, deploy logic
│   ├── database/           # SQLite init, migrations
│   ├── config/             # Server config
│   ├── clientconfig/       # Client config (NEW - Phase 0.5)
│   ├── mcp/                # MCP server (NEW - Phase 1)
│   ├── runtime/            # JS runtime (NEW - Phase 2)
│   └── api/                # Response helpers
├── admin/                  # React SPA (Vite + Tailwind)
└── data.db                 # SQLite database (server mode)
```

**Client config**: `~/.fazt/config.json` (multi-server)

---

## Commands

```bash
# Build
go build -o fazt ./cmd/server

# Test (with coverage)
go test -v -cover ./...

# Run locally
./fazt server start --port 8080

# Admin SPA dev
cd admin && npm run dev -- --port 37180

# Admin SPA build (embeds into binary)
cd admin && npm run build
```

---

## The Plan

The implementation is tracked in:
- **Plan**: `koder/plans/16_implementation-roadmap.md`
- **State**: `koder/STATE.md`

Phases:
```
0     Verify current deploy
0.5   Multi-server config
1     MCP Server
2     Serverless Runtime
3     Analytics App
4     Sites → Apps Migration
5     Release
6     Deploy to Production
7     Local Setup
8     MCP Setup
```

---

## Resuming After Context Reset

When starting a new session:

1. User says: "read & execute koder/start.md"
2. You read this file
3. You read STATE.md
4. You continue from where we left off

The state file is the source of truth. Trust it.

---

## Documentation Management

### CHANGELOG.md

Update CHANGELOG.md for every release. Follow existing format:

```markdown
## [0.8.0] - 2026-01-XX

### Added
- **Multi-Server Config**: Client configuration via `~/.fazt/config.json`
- **Server Management**: `fazt servers add/list/default/remove` commands
- **MCP Server**: Claude Code integration via Model Context Protocol
  - `fazt_servers_list`, `fazt_apps_list`, `fazt_deploy`, etc.
- **Serverless Runtime**: Execute `api/main.js` on HTTP requests
  - `require()` shim for local imports
  - `fazt.*` namespace (app, env, log)
- **CLI**: `fazt server create-key` for headless API key creation

### Changed
- `fazt deploy` now uses `~/.fazt/config.json` instead of `data.db`
- Deploy supports `--to <server>` for multi-server targeting

### Fixed
- (list any bugs fixed)

### Documentation
- Added `koder/plans/16_implementation-roadmap.md`
- Updated README with new commands
```

**Categories used**: Added, Changed, Fixed, Documentation, Removed

### README.md

Update README.md when:
- New commands are added
- Usage patterns change
- Quick start needs updating

### Other Docs

| Doc | When to Update |
|-----|----------------|
| `CONFIGURATION.md` | New config options |
| `koder/plans/*.md` | When plan changes |
| `koder/STATE.md` | After every step |

---

## Critical Rules

1. **Update STATE.md frequently** - After every significant step
2. **Tests before commits** - Never commit failing tests
3. **One phase at a time** - Complete current phase before moving on
4. **User blockers are blockers** - If user action needed, stop and ask
5. **Don't skip phases** - Each builds on the previous

---

## Current Status

Read `koder/STATE.md` for current status.

**GO.**
