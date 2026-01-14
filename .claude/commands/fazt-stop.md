# Fazt Session Stop

Run this before ending a work session to ensure clean handoff to the next session.

## Steps

### 1. Gather Current State

```bash
# Source code version
grep "var Version" internal/config/config.go

# Installed binary version
fazt --version

# Remote server version
fazt remote status zyt | grep -E "Version|Status"

# Git status
git status --short
```

**Verify versions match**: source = installed = remote. If not, may need to build/release.

### 2. Update koder/STATE.md

This is the **primary handoff document**. Update it with:

```markdown
# Fazt Implementation State

**Last Updated**: YYYY-MM-DD
**Current Version**: vX.Y.Z (local), vX.Y.Z (zyt)

## Status

\```
State: [CLEAN | IN_PROGRESS | BLOCKED]
[Brief description of current state]
\```

## Active Work (if any)

[What was being worked on, what's left to do]

## Recent Changes

[Summary of what was accomplished this session]

## Known Issues

[Any bugs, blockers, or pending items]

## Quick Reference

- **Primary context**: `CLAUDE.md` (root)
- **Deep implementation**: `koder/start.md`
- **Future specs**: `koder/ideas/specs/`
```

### 3. Update CLAUDE.md (if needed)

Only update if:
- New capabilities were added
- New commands/APIs were created
- Workflow changed significantly
- Version number needs updating

Key sections to check:
- `Current Version` in header area
- `Current Capabilities` table
- `Managing Fazt` section
- Any outdated instructions

### 4. Check CHANGELOG.md

Verify all released versions are documented:
- Read `internal/config/config.go` for current version
- Ensure CHANGELOG.md has entry for that version
- Ensure `docs/changelog.json` matches (for website)

### 5. Commit Documentation Updates

```bash
git add -A
git status  # Review what's being committed

# Only if there are doc changes:
git commit -m "docs: Session close - update state and docs

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
git push origin master
```

### 6. Final Status Report

Output a summary for the user:

```
## Session Summary

**Version**: vX.Y.Z
**zyt.app**: vX.Y.Z, [healthy/unhealthy]

### Completed
- [List of completed items]

### Pending (if any)
- [List of items for next session]

### Files Updated
- [List of modified files]

### Next Session Entry Point
Read `koder/STATE.md` for current state.
```

## Important Notes

- **Don't over-update**: Only modify docs that actually changed
- **Be specific**: Future Claude needs concrete details, not vague summaries
- **Version accuracy**: Always verify versions match across all files
- **Git cleanliness**: Ensure no uncommitted changes are left behind
