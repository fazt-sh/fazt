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

### 2. Verify Versions & Health

Check all components are in sync and healthy:

```bash
# Source version
grep "var Version" internal/config/config.go | grep -oE '[0-9]+\.[0-9]+\.[0-9]+'

# Binary version
fazt --version | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+'

# All configured remotes (shows status and default)
fazt remote list 2>/dev/null | tail -n +3

# Git status
git status --short
```

The remote list shows:
- All configured peers with health status
- Default remote marked with `*`
- Use `fazt remote status <name>` for detailed info on any remote

### 3. If Local Server Not Running

Check if `local` appears as unhealthy/unreachable:

```bash
systemctl --user start fazt-local
```

Or if not installed:
```bash
./install.sh  # Select option 2 for Local Development
```

### 4. Output Format

```
## Session Ready

| Component | Version | Status  |
|-----------|---------|---------|
| Source    | X.Y.Z   | -       |
| Local     | X.Y.Z   | healthy |
| Remote    | X.Y.Z   | healthy |

**Git**: clean | X uncommitted changes

### From Last Session
[Summary from STATE.md]

### Ready to Work On
[Next task from STATE.md, or ask user]
```

If versions mismatch:
- Source != Local → Rebuild: `go build -o ~/.local/bin/fazt ./cmd/server`
- Source != Remote → Consider release

## Quick Commands

| Action | Command |
|--------|---------|
| Rebuild binary | `go build -o ~/.local/bin/fazt ./cmd/server` |
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
| Feature specs | `koder/ideas/specs/` |
| Version history | `CHANGELOG.md` |
| Capacity limits | `koder/CAPACITY.md` |
