---
name: fazt-close
description: Fazt Session Close - Close session with proper handoff for next time. Updates STATE.md, summarizes work, and leaves repo in clean state.
---

# Fazt Session Close

Close session with proper handoff for next time. Leaves repo in a clean state.

## Steps

### 1. Gather Current State

```bash
# All version files (must be in sync!)
cat version.json | jq -r '.version'
grep "var Version" internal/config/config.go | grep -oE '[0-9]+\.[0-9]+\.[0-9]+'
cat knowledge-base/version.json | jq -r '.version'
cat admin/version.json | jq -r '.version'
fazt --version | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+'

# Latest release tag
git describe --tags --abbrev=0 2>/dev/null || echo "no tags"

# Commits since last release (shows if unreleased changes exist)
git log $(git describe --tags --abbrev=0 2>/dev/null)..HEAD --oneline 2>/dev/null | head -5

# All remotes health (shows version column)
fazt remote list 2>/dev/null | tail -n +3

# Git state
git status --short
git log --oneline -3
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

### 5. Review Knowledge-Base

**After any session with code changes**, check if knowledge-base needs updating:

| Change Type | Update Needed? |
|-------------|----------------|
| New CLI flag/command | Yes - update `knowledge-base/skills/app/fazt/cli-*.md` |
| New serverless API | Yes - update `knowledge-base/skills/app/references/serverless-api.md` |
| New pattern/workflow | Yes - add to `knowledge-base/skills/app/patterns/` |
| Bug fix only | No |
| Internal refactor | No |

**Always bump `knowledge-base/version.json` on release** (even if content unchanged):
- KB version is a sync marker that tracks compatibility with binary version
- Indicates KB has been verified against the released version

```bash
# Get current version and commit
VERSION=$(grep "var Version" internal/config/config.go | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date +%Y-%m-%d)

# Update version.json (use editor or write directly)
cat > knowledge-base/version.json << EOF
{
  "version": "$VERSION",
  "commit": "$COMMIT",
  "updated_at": "$DATE",
  "sections": {
    "skills/app": "$VERSION"
  }
}
EOF
```

### 6. Commit Tickets (if any uncommitted)

Check for uncommitted ticket files and commit them:

```bash
# Check for uncommitted tickets
git status --short .tickets/

# If any exist, commit them
git add .tickets/*.md
git diff --cached --quiet .tickets/ || git commit -m "chore: Update tickets

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

### 7. Commit Docs and Push

```bash
# Stage doc files (only if modified)
git add koder/STATE.md CHANGELOG.md CLAUDE.md knowledge-base/

# Check if anything staged
git diff --cached --quiet || git commit -m "docs: Session close - update STATE.md

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

# Push all commits
git push origin master
```

**Note**: If code changes are uncommitted, either:
- Commit them separately with appropriate message first
- Or leave for next session if incomplete

### 8. Consider Release

**Always explicitly reason about whether to release.** Ask:

1. **Was code committed this session?** If no, skip release.

2. **Can the changes be tested without deployment?**
   - Unit tests only → Maybe defer release
   - Requires server/browser/real APIs → Release needed to test

3. **What's the risk of releasing untested?**
   - Low risk (additive, isolated) → Release and test
   - High risk (breaking changes) → Defer if possible

**Decision matrix:**

| Changes | Testable Locally? | Action |
|---------|-------------------|--------|
| Bug fix | Yes | Release after local test |
| Bug fix | No (needs server) | Release to test |
| New feature | Yes | Test first, then release |
| New feature | No (needs server) | Release to test |
| Refactor | Yes | Test first, then release |

**If releasing:** Use `/fazt-release` skill.

**If not releasing:** Document why in the session output:
```
### Release
Not released: [reason - e.g., "No code changes" or "Can test locally first"]
```

**Common "needs deployment to test" scenarios:**
- OAuth/auth flows (need real redirect URIs)
- Cookie/session handling (need real domains)
- TLS/HTTPS behavior
- Domain routing
- External API integrations

### 9. Verify Clean State

```bash
git status --short
```

Should show nothing (clean working tree).

### 10. Output

```
## Session Closed

v0.18.0: Root, Binary, KB, Admin, Release, local, zyt ✓

Git: clean

### Version Check
[All versions in sync ✓] or [⚠️ Drift detected - see VERSIONING.md]

### Release
[Released vX.Y.Z] or [Not released: reason]
- If unreleased: "⚠️ X commits since vY.Z.W - consider /release"

### This Session
- [What was done]

### Next Session
- [What to work on]
```

## Principles

1. **Clean state** - Repo should have no uncommitted changes after stop
2. **STATE.md is the handoff** - Next session reads what you write
3. **Knowledge-base tracks docs** - Keep in sync with code changes
4. **Be specific** - Future Claude needs concrete details
5. **Tickets are tracked** - Always commit ticket changes
6. **Don't over-document** - Minimal sessions need minimal updates
