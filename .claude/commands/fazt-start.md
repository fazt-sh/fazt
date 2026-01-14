# Fazt Session Start

Run this at the beginning of a work session to get up to speed quickly.

## Steps

### 1. Read Current State

```bash
cat koder/STATE.md
```

This is the **primary handoff document** from the previous session. It contains:
- Current version
- What was completed
- Any pending work or blockers
- Quick reference to other docs

### 2. Verify Environment

```bash
# Check local version
fazt --version

# Check remote server
fazt remote status zyt

# Check git status
git status
```

### 3. Review Recent Changes (if needed)

```bash
# Recent commits
git log --oneline -10

# What changed recently
git diff HEAD~5 --stat
```

### 4. Sync Understanding

If STATE.md mentions pending work or blockers, investigate:
- Read referenced files
- Check mentioned issues
- Understand context before proceeding

### 5. Ready Check

Before starting work, confirm:
- [ ] Read STATE.md
- [ ] Local and remote versions match expectations
- [ ] Git is clean (or understand uncommitted changes)
- [ ] Know what to work on

## Quick Reference

| Doc | When to Read |
|-----|--------------|
| `koder/STATE.md` | Always - primary state |
| `CLAUDE.md` | If unfamiliar with project |
| `koder/start.md` | For deep implementation work |
| `koder/ideas/specs/` | When implementing new features |

## Output

After running session-start, output:

```
## Session Ready

**Version**: vX.Y.Z (local), vX.Y.Z (zyt)
**State**: [CLEAN | IN_PROGRESS | BLOCKED]

### Context
[Brief summary from STATE.md]

### Ready to work on
[What the user likely wants to do, or ask them]
```
