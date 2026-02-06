---
name: fazt-open
description: Fazt Session Open - Get up to speed quickly at the beginning of a work session. Checks versions, health, and reads previous session state.
---

# Fazt Session Open

Get up to speed quickly at the beginning of a work session.

## Steps

### 1. Read State

```bash
cat koder/STATE.md
```

This is the **handoff from previous session**:
- Current version and status
- What was completed
- What to work on next

### 2. Verify Monorepo Versions & Health

**Unified Versioning**: All components share the same version for guaranteed compatibility.

```bash
# Monorepo version (source of truth)
cat version.json

# Binary version (should match)
fazt --version | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+'

# Latest GitHub release
git describe --tags --abbrev=0 2>/dev/null || echo "no tags"

# Component versions & status
cat admin/version.json | jq -r '"\(.version) | \(.status) | \(.completeness)"'
cat knowledge-base/version.json | jq -r '"\(.version) | \(.status) | \(.completeness)"'

# All configured remotes (shows health)
fazt remote list 2>/dev/null | tail -n +3

# Git status
git status --short
```

### 3. Parse Component Status

Extract status from root version.json and create status table:

```bash
# Show all components with status
cat version.json | jq -r '.components | to_entries[] | "\(.key): \(.value.status) (\(.value.completeness))"'
```

### 4. If Local Server Not Running

Check if `local` appears as unhealthy/unreachable:

```bash
systemctl --user start fazt-local
```

Or if not installed:
```bash
./install.sh  # Select option 2 for Local Development
```

### 5. Output Format

```
## Session Ready

v0.18.0: Source, Binary, Release, local, zyt ✓

Git: clean | X uncommitted

### From Last Session
[Summary from STATE.md]

### Ready to Work On
[Next task from STATE.md, or ask user]
```

**If versions differ, group by version:**
```
v0.18.0: Source, Binary, Release, zyt
v0.17.0: local ⚠️ → fazt remote upgrade local
```

**Fixes:**
- Binary behind → `fazt upgrade` (use canonical upgrade, not manual build)
- Remote behind → `fazt remote upgrade <name>`

## Quick Commands

| Action | Command |
|--------|---------|
| Upgrade binary | `fazt upgrade` (canonical upgrade) |
| Restart local | `systemctl --user restart fazt-local` |
| Local logs | `journalctl --user -u fazt-local -f` |
| Local status | `systemctl --user status fazt-local` |
| List remotes | `fazt remote list` |
| Remote details | `fazt remote status <name>` |

## Reference

| Need | Read |
|------|------|
| Project context | `CLAUDE.md` |
| Current state | `koder/STATE.md` |
| Knowledge-base | `knowledge-base/` |
| Feature specs | `koder/ideas/specs/` |
| Version history | `CHANGELOG.md` |
| Capacity limits | `koder/CAPACITY.md` |
