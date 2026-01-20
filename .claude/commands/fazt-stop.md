# Fazt Session Stop

Close session with proper handoff for next time.

## Steps

### 1. Check State

```bash
fazt --version
grep "var Version" internal/config/config.go
fazt remote status zyt | grep -E "Version|Status"
git status --short
git log --oneline -5
```

### 2. Update STATE.md

Write `koder/STATE.md` with:

```markdown
# Fazt Implementation State

**Last Updated**: YYYY-MM-DD
**Current Version**: vX.Y.Z

## Status

State: CLEAN | IN_PROGRESS | BLOCKED
[One line describing current state]

---

## Last Session
[What was accomplished - be specific]

## Next Up
[Clear entry point for next session]

---

## Quick Reference
[Relevant commands or notes for continuity]
```

### 3. Update CHANGELOG.md (if substantial changes)

If significant work was done but not released, add to `## [Unreleased]` section:

```markdown
## [Unreleased]

### Added/Changed/Fixed
- [Description of change]
```

This ensures work is documented even without a release.

### 4. Commit and Push

```bash
git add -A
git commit -m "docs: Update STATE.md for session close

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
git push origin master
```

### 5. Output

```
## Session Closed

**Version**: vX.Y.Z (source = binary = zyt: ✓)

### This Session
- [What was done]

### Next Session
- [What to work on]
```

## Principles

1. **STATE.md is the handoff** - Next session reads what you write
2. **Be specific** - Future Claude needs concrete details
3. **Document unreleased work** - CHANGELOG [Unreleased] captures substantial changes
4. **Don't over-document** - Minimal sessions need minimal updates
