# Fazt Session Start

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

### 2. Verify Versions

```bash
fazt --version
grep "var Version" internal/config/config.go
fazt remote status zyt | grep -E "Version|Status"
git status --short
```

All three should match. If not, may need to build or release.

### 3. Output

```
## Session Ready

**Version**: vX.Y.Z (source = binary = zyt: ✓)
**State**: CLEAN | IN_PROGRESS | BLOCKED

### From Last Session
[Summary from STATE.md]

### Ready to Work On
[Next task from STATE.md, or ask user]
```

## Reference

| Need | Read |
|------|------|
| Project context | `CLAUDE.md` |
| Current state | `koder/STATE.md` |
| Feature specs | `koder/ideas/specs/` |
| Version history | `CHANGELOG.md` |
