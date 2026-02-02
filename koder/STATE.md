# Fazt Implementation State

**Last Updated**: 2026-02-02
**Current Version**: v0.20.0 (released)

## Status

State: **WORKING** - Markdown-based CLI help system (unreleased)

---

## Current Session (2026-02-02) - Markdown-Based CLI Help System

**Completed**: CLI help now reads from embedded markdown files

### What Was Done

#### Markdown-Based Help System ✅
- Created `internal/help/` package with types, loader, and renderer
- Uses `//go:embed` to embed markdown docs at build time
- Glamour library renders markdown with colors in terminal
- Falls back to hardcoded help when markdown not found
- Piped output returns plain markdown (no ANSI codes)

#### Files Created
- `internal/help/types.go` - CommandDoc, Argument, Flag, Example, Related structs
- `internal/help/loader.go` - Load and parse markdown with YAML frontmatter
- `internal/help/render.go` - Render docs with glamour
- `internal/help/embed.go` - Embed directive for cli/ docs
- `internal/help/loader_test.go` - Unit tests

#### CLI Documentation (embedded)
- `internal/help/cli/fazt.md` - Root help
- `internal/help/cli/app/_index.md` - App group help
- `internal/help/cli/app/deploy.md` - Deploy command
- `internal/help/cli/app/list.md` - List command
- `internal/help/cli/peer/_index.md` - Peer group help

### What Works
```bash
# All these commands render markdown help
fazt --help
fazt app --help
fazt app deploy --help
fazt peer --help

# Fallback for commands without markdown
fazt server --help   # Uses hardcoded help
fazt service --help  # Uses hardcoded help

# Piped output is plain markdown
fazt app --help | head -10
```

### Tests Passed
- ✅ `TestResolveFilePath` - Path resolution works
- ✅ `TestLoad` - Loading markdown docs works
- ✅ `TestExists` - Existence check works
- ✅ `TestRender` - Rendering produces output
- ✅ Binary compiles and runs
- ✅ Local server running with updated binary

---

## Previous Session (2026-02-02) - Migration Log Suppression

**Completed**: Suppress migration logs unless --verbose flag is provided

### What Was Done

#### Migration Log Suppression ✅
- Added `database.SetVerbose(bool)` function to control migration logging
- Modified `database.Init()` and `runMigrations()` to respect verbose flag
- Added `--verbose` flag extraction in main.go
- Updated help text to document `--verbose` and `--format` flags
- Migration logs now silent by default (clean CLI output)
- Logs show when `--verbose` flag is provided (for debugging)

### Files Modified
- `internal/database/db.go` - Added SetVerbose(), conditional logging
- `cmd/server/main.go` - Extract --verbose flag, call SetVerbose(), updated help (main + peer)
- `cmd/server/app.go` - Added GLOBAL FLAGS section to app help
- `knowledge-base/cli/commands.md` - Added Global Flags section, documented --verbose
- `knowledge-base/agent-context/api.md` - Added Global CLI Flags section
- `knowledge-base/agent-context/setup.md` - Added Troubleshooting section with --verbose usage
- `knowledge-base/agent-context/peer-routing.md` - Added debugging best practice

### What Works
```bash
# Clean output (no migration logs)
fazt @local app list
fazt peer list
fazt sql "SELECT * FROM apps"

# Verbose output (shows migrations)
fazt --verbose @local app list
fazt --verbose peer status

# Help documents the flag
fazt --help  # Shows --verbose in GLOBAL FLAGS section
```

### Tests Passed
- ✅ `fazt @local app list` produces clean output (no migration logs)
- ✅ `fazt --verbose @local app list` shows all migration logs
- ✅ `fazt peer list` clean output
- ✅ `fazt sql "query"` clean output
- ✅ Help text documents --verbose flag
- ✅ Binary compiles and runs
- ✅ Local server running with updated binary

---

## Previous Session (2026-02-02) - @Peer Pattern Audit Phase 3 & 4

**Completed**: Phase 3 & 4 of @peer pattern audit plan

### What Was Done

#### Phase 3.1: `app files` Command ✅
- Added `handleAppFiles()` CLI handler in `cmd/server/app_v2.go`
- Registered `files` case in app command router
- Implemented file listing with output system integration
- Fixed `remote.FileEntry` struct to match API response format
- Command shows path, size (formatted), and modified timestamp
- Supports both `--alias` and `--id` flags for app identification
- Full JSON output support with `--format json`

#### Phase 3.2: Enhanced `peer status` Command ✅
- Converted `handlePeerStatus()` to use output system
- Added support for global `targetPeerName` context
- Built structured markdown tables for health and resources
- Implemented JSON output with full status data
- Clean separation of concerns (health vs resources)

#### Phase 4: Documentation ✅
- Created `knowledge-base/agent-context/peer-routing.md`
  - Complete @peer pattern reference
  - Lists all remote-capable vs local-only commands
  - Error handling patterns and usage examples
  - Implementation details for developers
- Updated `CHANGELOG.md` with Phase 3 & 4 changes
- Updated `knowledge-base/skills/app/fazt/cli-app.md` with `app files` command

### Files Modified
- `cmd/server/app_v2.go` - Added `handleAppFiles()`, updated switch and help
- `cmd/server/main.go` - Enhanced `handlePeerStatus()` with output system
- `internal/remote/client.go` - Fixed FileEntry JSON tags
- `CHANGELOG.md` - Added Phase 3 & 4 to Unreleased section
- `knowledge-base/skills/app/fazt/cli-app.md` - Added app files documentation

### Files Created
- `knowledge-base/agent-context/peer-routing.md` - Comprehensive @peer guide

### What Works
```bash
# New app files command
fazt @local app files admin-ui
fazt @zyt app files tetris --format json
fazt app files my-app --id app_abc123

# Enhanced peer status
fazt peer status
fazt peer status local
fazt peer status --format json

# All output formatted consistently
- Markdown tables with proper headers
- JSON output with structured data
- File sizes formatted (KB, MB)
- Timestamps in ISO 8601 format
```

### Tests Passed
- ✅ `fazt @local app files admin-ui` shows 10 files with sizes/dates
- ✅ `fazt app files admin-ui --format json` returns structured JSON
- ✅ `fazt peer status` renders markdown tables
- ✅ `fazt peer status --format json` returns full status data
- ✅ Help text includes `files` command and examples
- ✅ All tests in `/tmp/test-phase3.sh` pass
- ✅ Remote package tests pass
- ✅ Binary compiles without errors

### Release Complete ✅
**Version**: v0.20.0
**Released**: 2026-02-02
**GitHub**: https://github.com/fazt-sh/fazt/releases/tag/v0.20.0

All Phase 3 & 4 work committed and released:
- ✅ Committed all changes
- ✅ Version bumped to v0.20.0
- ✅ CHANGELOG.md and docs/changelog.json updated
- ✅ Git tag created and pushed
- ✅ GitHub release with all 4 platform binaries
- ✅ Local binary updated
- ✅ Remote peer (zyt) upgraded to v0.20.0
- ✅ All tests passing

---

## Previous Session (2026-02-02) - Complete CLI Refactor & Release

**Completed Plans: 31 (FULL), 33, 25**
**Released**: v0.19.0

### What Was Done

#### Plan 31: CLI Refactor (COMPLETE) ✅
- Renamed `remote` → `peer` command throughout codebase
- **Removed ALL** `--to`, `--from`, `--on` flags from app commands
- Added global peer context (`targetPeerName`) for @peer routing
- Refactored ALL app handlers (deploy, remove, install, upgrade, pull, etc.)
- Updated help messages to reflect @peer pattern
- @peer is now PRIMARY pattern: `fazt @zyt app deploy ./app`
- Local operations by default: `fazt app list` (no peer needed)
- **BREAKING CHANGE** - Old flag syntax no longer works

#### Plan 33: Standardized CLI Output ✅
- Created `internal/output/` package with glamour rendering
- Added `--format` flag (markdown, json)
- Implemented markdown rendering with tables
- Updated `peer list` command to use new output system
- Both markdown and JSON formats working

#### Plan 25: SQL Command ✅
- Implemented `fazt sql "query"` for local execution
- Added `/api/sql` endpoint for remote queries
- Implemented `fazt @peer sql "query"` for remote execution
- Write protection with `--write` flag
- Both markdown and JSON output formats

### Files Created
- `internal/output/format.go` - Output renderer with glamour
- `internal/output/table.go` - Markdown table helper
- `internal/output/builder.go` - Markdown builder utilities
- `cmd/server/sql.go` - SQL command implementation
- `internal/handlers/sql.go` - SQL API endpoint
- `knowledge-base/skills/app/fazt/cli-peer.md` - Renamed from cli-remote.md

### Files Modified
- `cmd/server/main.go` - Global peer context, @peer routing without flag injection
- `cmd/server/app.go` - Removed ALL peer flags from handlers
- `cmd/server/app_v2.go` - Removed ALL peer flags from v2 handlers
- `cmd/server/app_logs.go` - Removed --peer flag
- `cmd/server/auth.go` - Renamed remote functions to peer
- `internal/remote/client.go` - Added SQL() method
- `internal/config/config.go` - Version bump to 0.19.0
- `version.json` - Version bump to 0.19.0
- `CHANGELOG.md` - Added v0.19.0 entry
- `docs/changelog.json` - Added v0.19.0 entry
- `go.mod` - Added glamour dependency
- **20+ documentation files** across knowledge-base updated

### What Works
```bash
# Peer management (renamed from remote)
fazt peer list
fazt peer list --format json
fazt peer add <name> --url <url> --token <token>
fazt peer upgrade zyt

# @peer pattern (PRIMARY)
fazt @zyt app list
fazt @zyt app deploy ./myapp
fazt @zyt app remove myapp
fazt @local app list  # explicit local

# Local by default (NEW)
fazt app list         # uses default peer (or local if one peer)
fazt app deploy ./app # no flags needed!

# SQL queries (local)
fazt sql "SELECT * FROM apps LIMIT 5"
fazt sql "SELECT * FROM apps" --format json
fazt sql "UPDATE ..." --write

# SQL queries (remote)
fazt @zyt sql "SELECT * FROM apps LIMIT 5"
fazt @local sql "SELECT COUNT(*) FROM peers"

# Output formatting
--format markdown  (default, beautiful terminal rendering)
--format json      (machine-readable)
```

### Tests Passed
- ✅ `fazt peer list` returns peers with markdown/json formats
- ✅ `fazt peer upgrade zyt` successfully upgraded to v0.19.0
- ✅ `fazt sql "SELECT..."` works locally
- ✅ `fazt @zyt sql "SELECT..."` works remotely
- ✅ `fazt @zyt app list` works (no --to flag needed)
- ✅ `fazt app list` defaults to configured peer
- ✅ `fazt @local app list` explicit local works
- ✅ Write protection requires --write flag
- ✅ Binary compiles and runs
- ✅ Local server restarted with new binary
- ✅ GitHub release v0.19.0 published
- ✅ All 4 platform binaries uploaded

---

## Next Up

**Session Complete!** All requested work finished:
- ✅ Plan 31 fully implemented
- ✅ All documentation updated
- ✅ v0.19.0 released and deployed
- ✅ Everything verified working

**Future work (optional):**

1. **Expand output system**: Convert more commands to use output package
   - Currently only peer list and sql use new output system
   - Could add to: app list, app info, etc.

2. **New features**: Continue with `koder/THINKING_DIRECTIONS.md`
   - Explore remaining enhancement ideas
   - Pick next high-impact feature

3. **Polish**: Minor improvements
   - Suppress migration logs in CLI output
   - Add progress indicators for long operations
   - Improve error messages

---

## Quick Reference

```bash
# New commands this session
fazt peer list                    # Renamed from 'remote list'
fazt sql "SELECT * FROM apps"     # New: local SQL queries
fazt @zyt sql "SELECT ..."        # New: remote SQL queries
--format json                     # New: JSON output for any command

# Database location
~/.fazt/data.db

# Binary
~/.local/bin/fazt
```

---

## Post-Release Improvements (unreleased)

### 1. Peer List Live Status (b36c18f)
- Changed `fazt peer list` to always show live status/version
- Previously showed stale cached data from database
- Now makes API calls to all peers (~1.2s for 2 peers)
- More accurate, acceptable performance

### 2. @Peer Pattern Audit & Output Standardization (da853aa)

**Phase 1: @Peer Pattern Consistency** ✅
- Added guards to local-only commands (`app create`, `app validate`)
- These commands now error clearly when used with @peer prefix
- Improved auth remote error messages with SSH guidance
- Updated help text to separate remote vs local commands

**Phase 2: Output Standardization** ✅
- Converted `app list` to use `internal/output` system
- Converted `auth providers`, `auth users`, `auth invites` to output system
- All converted commands now support `--format json`
- Markdown output uses glamour for beautiful terminal rendering

**What Works:**
```bash
# Guards prevent misuse
fazt @zyt app create test-app       # ❌ Errors clearly
fazt @zyt app validate ./app        # ❌ Errors clearly
fazt @zyt auth users                # ❌ Errors with SSH guidance

# Correct usage
fazt app create test-app            # ✅ Creates locally
fazt app validate ./app             # ✅ Validates locally
fazt auth users                     # ✅ Lists local users

# JSON output for scripting
fazt @zyt app list --format json    # ✅ Machine-readable apps
fazt auth providers --format json   # ✅ Machine-readable providers
fazt auth users --format json       # ✅ Machine-readable users
```

**Commands with Output System (6 total):**
- `peer list` (markdown, json)
- `sql` (markdown, json)
- `app list` (markdown, json) ← NEW
- `auth providers` (markdown, json) ← NEW
- `auth users` (markdown, json) ← NEW
- `auth invites` (markdown, json) ← NEW

**Release recommendation:** Can batch with v0.19.1 or next feature release. All improvements are additive/clarifying.

