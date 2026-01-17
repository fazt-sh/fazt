# Fazt Implementation State

**Last Updated**: 2026-01-17
**Current Version**: v0.9.23

## Status

```
State: PLAN READY
Plan 18 (App Ecosystem) approved and ready for implementation.
```

---

## Active Plan

**Plan**: `koder/plans/18_app-ecosystem.md`
**Goal**: Unified `fazt app` CLI namespace + git-based app distribution

### What Was Decided

1. **CLI Restructure** (breaking change OK)
   - `fazt remote deploy` → `fazt app deploy --to <peer>`
   - `fazt remote apps` → `fazt app list [peer]`
   - New: `fazt app install/upgrade/pull/remove/info`

2. **Git Integration**
   - go-git for GitHub installation (HTTPS, public repos)
   - Source tracking in DB (source_type, source_url, source_commit)
   - Upgrade detection via commit comparison

3. **`/fazt-app` Skill**
   - Claude builds apps with fazt context
   - Location behavior:
     - In fazt repo → `servers/zyt/{app}/`
     - Elsewhere → `/tmp/fazt-{app}-{hash}/`
     - `--in <dir>` or `--tmp` for explicit control

### Implementation Phases

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | CLI restructure (`fazt app` namespace) | Pending |
| 2 | DB schema + source tracking | Pending |
| 3 | go-git integration | Pending |
| 4 | Install & upgrade commands | Pending |
| 5 | Pull command & API | Pending |
| 6 | `/fazt-app` skill | Pending |

### Next Action

Start Phase 1: Create `cmd/server/app.go` with new command structure.

---

## Apps on zyt.app

| App | URL |
|-----|-----|
| home | https://zyt.app |
| tetris | https://tetris.zyt.app |
| snake | https://snake.zyt.app |

---

## Quick Reference

| Command | Purpose |
|---------|---------|
| `fazt remote status zyt` | Check health/version |
| `fazt remote upgrade zyt` | Upgrade server |
| `fazt remote deploy <dir> zyt` | Deploy app |

---

## Recent Releases

| Version | Date | Summary |
|---------|------|---------|
| v0.9.23 | 2026-01-16 | Verification: auto-restart confirmed |
| v0.9.22 | 2026-01-16 | Fix: systemd-run for auto-restart |

---

## Deferred (Not in Plan 18)

- Private repo auth (needs Notary vault)
- SSH keys (needs Persona)
- Commit signing (needs Persona)
- VFS versioning (`fazt.git` for agents)
- Push to GitHub (agent publishing)

These will be added incrementally without breaking changes.
