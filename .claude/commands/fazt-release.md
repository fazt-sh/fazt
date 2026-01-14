# Fazt Release Skill

Release a new version of fazt with proper versioning, documentation, and deployment.

**Idempotent**: Safe to run multiple times - checks state before each step.

## Pre-flight Checks

```bash
# 1. Check git status
git status

# 2. Run tests
go test ./...

# 3. Get current version
grep "var Version" internal/config/config.go
```

## Release Steps

### 1. Determine Version

Check what version to release:
```bash
# Current version in code
grep "var Version" internal/config/config.go

# Latest git tag
git describe --tags --abbrev=0 2>/dev/null || echo "no tags"

# Pending changes since last tag
git log $(git describe --tags --abbrev=0)..HEAD --oneline
```

If code version matches latest tag and no pending changes, **nothing to release**.

### 2. Update Version (if needed)

Only if version needs bumping:
```go
var Version = "X.Y.Z"  // in internal/config/config.go
```

### 3. Update CHANGELOG.md (if needed)

Check if entry exists:
```bash
grep "## \[X.Y.Z\]" CHANGELOG.md
```

If not, add at top (after header):
```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added/Changed/Fixed
- ...
```

### 4. Update docs/changelog.json (if needed)

Check if entry exists:
```bash
grep '"vX.Y.Z"' docs/changelog.json
```

If not, add at top of array.

### 5. Commit (if needed)

```bash
# Check if there are changes to commit
git status --porcelain

# Only commit if there are changes
git add -A
git diff --cached --quiet || git commit -m "release: vX.Y.Z

<summary>

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

### 6. Tag (if needed)

```bash
# Check if tag exists
git tag -l "vX.Y.Z"

# Only create if doesn't exist
git tag -l "vX.Y.Z" | grep -q . || git tag vX.Y.Z
```

### 7. Build and Install Locally

```bash
go build -ldflags "-X github.com/fazt-sh/fazt/internal/config.Version=X.Y.Z" -o ~/.local/bin/fazt ./cmd/server
fazt --version
```

### 8. Push (if needed)

```bash
# Check if we need to push
git status -sb | grep -q "ahead" && git push origin master --tags || echo "Already pushed"
```

### 9. Wait for CI

```bash
# Poll until complete
curl -s "https://api.github.com/repos/fazt-sh/fazt/actions/runs?per_page=1" | jq '.workflow_runs[0] | {status, conclusion}'
```

Wait for `status: completed` and `conclusion: success`.

### 10. Upgrade Remote Servers

```bash
fazt remote upgrade zyt
```

Returns "already latest" if already upgraded (idempotent).

## Quick Reference

| Check | Command |
|-------|---------|
| Current version | `grep "var Version" internal/config/config.go` |
| Latest tag | `git describe --tags --abbrev=0` |
| Tag exists? | `git tag -l "vX.Y.Z"` |
| Need to push? | `git status -sb` (shows "ahead") |
| Remote version | `fazt remote status zyt` |

## Version Numbering

- **Patch** (0.9.5): Bug fixes
- **Minor** (0.10.0): New features, backwards compatible
- **Major** (1.0.0): Breaking changes
