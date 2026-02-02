# Fazt Implementation State

**Last Updated**: 2026-02-02
**Current Version**: v0.21.0 (released)

## Status

State: **CLEAN** - Markdown-based CLI help system released

---

## Last Session (2026-02-02) - Markdown-Based CLI Help System

**Released**: v0.21.0

### What Was Done

#### 1. Markdown-Based CLI Help System ✅
- Created `internal/help/` package with types, loader, and renderer
- Uses `//go:embed` to embed markdown docs at build time
- Glamour library renders markdown with colors in terminal
- Falls back to hardcoded help when markdown not found
- Piped output returns plain markdown (no ANSI codes)
- Plain `go build` works (no build-time copy needed)

#### 2. Architecture: Single Source of Truth ✅
- `internal/help/cli/` is the source of truth (tracked in git)
- `knowledge-base/cli` symlinks to it for web doc export
- No divergent copies - DRY principle enforced
- Go's embed reads directly from `internal/help/cli/`

#### 3. CLI Documentation ✅
- `internal/help/cli/fazt.md` - Root help
- `internal/help/cli/app/_index.md` - App group help
- `internal/help/cli/app/deploy.md` - Deploy command (updated from original)
- `internal/help/cli/app/list.md` - List command
- `internal/help/cli/peer/_index.md` - Peer group help

#### 4. Migration Log Suppression ✅
- Added `database.SetVerbose(bool)` function
- Modified `database.Init()` and `runMigrations()` to respect verbose flag
- Migration logs now silent by default
- Use `--verbose` flag to see migration logs

### Files Created
- `internal/help/types.go` - CommandDoc, Argument, Flag, Example structs
- `internal/help/loader.go` - Load and parse markdown with YAML frontmatter
- `internal/help/render.go` - Render docs with glamour
- `internal/help/embed.go` - Embed directive for cli/ docs
- `internal/help/loader_test.go` - Unit tests
- `internal/help/cli/fazt.md` - Root help
- `internal/help/cli/app/_index.md` - App group
- `internal/help/cli/app/list.md` - List command
- `internal/help/cli/peer/_index.md` - Peer group

### Files Modified
- `cmd/server/main.go` - Added help import, showCommandHelp, updated printUsage/printPeerHelp
- `cmd/server/app_v2.go` - Added help import, updated printAppHelpV2
- `cmd/server/app.go` - Added help import, updated deploy help handler
- `.gitignore` - No longer ignores internal/help/cli/
- `Makefile` - Removed build-time copy steps
- `internal/database/db.go` - Added SetVerbose(), conditional logging
- `knowledge-base/agent-context/api.md` - Added Global CLI Flags section
- `knowledge-base/agent-context/peer-routing.md` - Added debugging best practice
- `knowledge-base/agent-context/setup.md` - Added Troubleshooting section

### What Works
```bash
# Markdown-based help
fazt --help
fazt app --help
fazt app deploy --help
fazt peer --help

# Fallback to hardcoded
fazt server --help   # No markdown doc yet

# Plain markdown when piped
fazt app --help | head -10

# Migration logs suppressed
fazt @local app list   # Clean output
fazt --verbose @local app list   # Shows migrations

# Plain go build works
go build ./cmd/server
```

### Tests Passed
- ✅ All unit tests pass (`go test ./...`)
- ✅ Help package tests pass
- ✅ Binary compiles with plain `go build`
- ✅ Local server running v0.21.0
- ✅ Remote peer (zyt) upgraded to v0.21.0

### Release
**Released v0.21.0**
- GitHub: https://github.com/fazt-sh/fazt/releases/tag/v0.21.0
- Local binary: v0.21.0
- zyt peer: v0.21.0
- local peer: v0.20.0 (not upgraded yet)

---

## Next Up

### High Priority
1. **Upgrade local peer to v0.21.0** - Currently on v0.20.0
2. **Expand CLI help docs** - Add more commands (server, auth, sql, etc.)
3. **Complete help migration** - Remove all LEGACY_CODE markers once all commands have markdown docs

### Future Work
1. **Web HTML rendering** - docs-rendering-design.md Phase 2
2. **Search/navigation generation** - Auto-generate from CLI docs
3. **Theme system** - Customizable help colors
4. **Full command coverage** - All commands with markdown help

---

## Quick Reference

```bash
# Version
fazt --version                  # v0.21.0

# Database
~/.fazt/data.db                 # Single DB for everything

# Binary
~/.local/bin/fazt               # Installed binary

# Help system
internal/help/cli/              # Source of truth for CLI docs
knowledge-base/cli              # Symlink → ../internal/help/cli

# Local server
systemctl --user status fazt-local
journalctl --user -u fazt-local -f

# Remote operations
fazt peer list
fazt @local upgrade             # Upgrade local peer
fazt @zyt app list
```

---

## Architecture Notes

### CLI Help System
- **Single source of truth**: `internal/help/cli/` (tracked in git)
- **Symlink**: `knowledge-base/cli` → `../internal/help/cli`
- **Embed**: `//go:embed all:cli` in `internal/help/embed.go`
- **No build steps**: Plain `go build` works
- **Web export**: Symlink transparent to web doc system

### YAML Frontmatter Schema
```yaml
command: "app deploy"
description: "Deploy a local directory"
syntax: "fazt [@peer] app deploy <directory> [flags]"
arguments:
  - name: "directory"
    required: true
    description: "Path to deploy"
flags:
  - name: "--name"
    short: "-n"
    type: "string"
    default: "directory name"
    description: "App name"
examples:
  - title: "Deploy to local"
    command: "fazt app deploy ./my-app"
related:
  - command: "app list"
    description: "List deployed apps"
```
