# Fazt Session Stop

Close session with proper handoff for next time.

## Steps

### 1. Gather Current State

```bash
# Versions
grep "var Version" internal/config/config.go | grep -oE '[0-9]+\.[0-9]+\.[0-9]+'
fazt --version | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+'

# All remotes health
fazt remote list 2>/dev/null | tail -n +3

# Git state
git status --short
git log --oneline -5
```

### 2. Update STATE.md

Write `koder/STATE.md` with current state. Template:

```markdown
# Fazt Implementation State

**Last Updated**: YYYY-MM-DD
**Current Version**: vX.Y.Z

## Status

State: CLEAN | IN_PROGRESS | BLOCKED
[One line describing current state]

---

## Last Session

**[Session Title]**

1. **[Change 1]**
   - Details...

2. **[Change 2]**
   - Details...

## Next Up

1. **[Task 1]** - Brief description
2. **[Task 2]** - Brief description

---

## Quick Reference

```bash
# Relevant commands for continuity
```
```

### 3. Update CHANGELOG.md (if substantial changes)

If significant work was done but not released, add to `## [Unreleased]`:

```markdown
## [Unreleased]

### Added
- New feature or file

### Changed
- Modified behavior

### Fixed
- Bug fix
```

### 4. Update CLAUDE.md (if capabilities changed)

If new features, commands, or workflows were added:
- Update relevant sections in `CLAUDE.md`
- Keep it concise - CLAUDE.md is reference, not narrative

### 5. Commit and Push

```bash
git add koder/STATE.md CHANGELOG.md CLAUDE.md
git status --short  # Verify what's staged

git commit -m "docs: Update STATE.md and CHANGELOG for session close

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin master
```

**Note**: Only commit doc changes. If code changes are uncommitted, either:
- Commit them separately with appropriate message
- Or leave for next session if incomplete

### 6. Output

```
## Session Closed

**Version**: vX.Y.Z

| Component | Version | Status |
|-----------|---------|--------|
| Source    | X.Y.Z   | -      |
| Binary    | X.Y.Z   | ✓      |

**Remotes**: [all healthy / X unreachable]

### This Session
- [What was done]

### Next Session
- [What to work on]

### Committed
- [files committed, or "no changes"]
```

## Principles

1. **STATE.md is the handoff** - Next session reads what you write
2. **Be specific** - Future Claude needs concrete details
3. **CHANGELOG for unreleased work** - Captures substantial changes between releases
4. **CLAUDE.md for capability changes** - Update when features/workflows change
5. **Don't over-document** - Minimal sessions need minimal updates
6. **Clean commits** - Separate code commits from doc commits
