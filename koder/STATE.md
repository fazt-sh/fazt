# Fazt Implementation State

**Last Updated**: 2026-02-02
**Current Version**: v0.21.0 (released)

## Status

State: **CLEAN** - Universal @peer pattern implemented (4 commits since v0.21.0)

---

## Last Session (2026-02-02) - Universal @Peer Routing Pattern

### What Was Done

#### 1. Universal @Peer Pattern ✅
Implemented universal `fazt @<target> <command>` syntax for ALL commands:
- `fazt @zyt status` (was: `fazt peer status zyt`)
- `fazt @zyt upgrade` (was: `fazt peer upgrade zyt`)
- All commands use same pattern - no exceptions to remember

#### 2. Helpful Error Messages ✅
Commands that can't work remotely show SSH guidance:
```
$ fazt @zyt service install
Error: 'service' requires local system access (systemd/sudo).

To manage the service on zyt, SSH into the machine:
  ssh user@zyt.app
  fazt service install
```

#### 3. Documentation Updates ✅
- Comprehensive rewrite of `peer-routing.md`
- Updated `api.md`, `setup.md`, `cli-peer.md`
- Updated all embedded help docs (`fazt.md`, `deploy.md`, `peer/_index.md`)
- Updated release skill with new syntax
- Moved `analysis.md` design doc out of embedded help

### Key Files Changed
- `cmd/server/main.go` - Universal @peer router with `handleAtPeerRouting()`
- `knowledge-base/agent-context/peer-routing.md` - Complete rewrite
- `internal/help/cli/peer/_index.md` - Updated for new pattern
- `.claude/commands/release.md` - Updated upgrade steps

### Tests Passed
- ✅ All unit tests pass (`go test ./...`)
- ✅ Binary compiles with plain `go build`
- ✅ Documentation consistent across all files

### Release
**Not released** - 4 commits since v0.21.0 (documentation + CLI routing)
- Consider releasing v0.22.0 when ready to test @peer pattern on production

---

## Next Up

### High Priority
1. **Test universal @peer pattern** - Verify `fazt @zyt status` and `fazt @zyt upgrade` work
2. **Release v0.22.0** - Deploy universal @peer pattern changes

### Future Work
1. **Expand CLI help docs** - Add more commands (server, auth, sql, etc.)
2. **Web HTML rendering** - docs-rendering-design.md Phase 2
3. **Full command coverage** - All commands with markdown help

---

## Quick Reference

```bash
# Version
fazt --version                  # v0.21.0

# Universal @peer syntax (NEW)
fazt @zyt status                # Check peer health
fazt @zyt upgrade               # Upgrade peer binary
fazt @zyt app list              # List apps
fazt @local status              # Check local peer

# Local peer config (no @peer)
fazt peer list                  # List configured peers
fazt peer add prod ...          # Add peer

# Local server
systemctl --user status fazt-local
journalctl --user -u fazt-local -f

# Build
go build -o ~/.local/bin/fazt ./cmd/server
```

---

## Architecture Notes

### CLI Help System
- **Single source of truth**: `internal/help/cli/` (tracked in git)
- **Symlink**: `knowledge-base/cli` → `../internal/help/cli`
- **Embed**: `//go:embed all:cli` in `internal/help/embed.go`
- **No build steps**: Plain `go build` works

### Universal @Peer Pattern
```
fazt @<target> <command> [args...]
     ^^^^^^^^  ^^^^^^^^^
     peer name  any command

Commands that work remotely: app, status, upgrade, sql, auth providers
Commands with helpful errors: service, config, server init, app create
```
