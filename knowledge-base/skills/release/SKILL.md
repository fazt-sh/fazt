---
name: fazt-release
description: Release a new version of fazt with proper versioning, documentation, and deployment. Use when asked to release, publish, or deploy a new version.
model: opus
---

# Fazt Release Skill

Release a new version of fazt with proper versioning, documentation, and deployment.

**Idempotent**: Safe to run multiple times - checks state before each step.

## Pre-flight Checks

```bash
# 1. Check git status
git status

# 2. Run tests
go test ./...

# 3. Verify all versions are in sync (see VERSIONING.md)
cat version.json | jq -r '.version'
grep "var Version" internal/config/config.go | grep -oE '[0-9]+\.[0-9]+\.[0-9]+'
cat knowledge-base/version.json | jq -r '.version'
cat admin/version.json | jq -r '.version'
git describe --tags --abbrev=0 2>/dev/null || echo "no tags"

# All should match before proceeding!
```

## Release Steps

### 1. Determine Version

Check what version to release:
```bash
# Current version in all files
cat version.json | jq -r '.version'
grep "var Version" internal/config/config.go | grep -oE '[0-9]+\.[0-9]+\.[0-9]+'
cat knowledge-base/version.json | jq -r '.version'
cat admin/version.json | jq -r '.version'

# Latest git tag
git describe --tags --abbrev=0 2>/dev/null || echo "no tags"

# Pending changes since last tag
git log $(git describe --tags --abbrev=0)..HEAD --oneline
```

If all versions match latest tag and no pending changes, **nothing to release**.

### 2. Update ALL Version Files (if needed)

**CRITICAL**: Update ALL version files together. See `VERSIONING.md` for complete workflow.

```bash
# Set new version
NEW_VERSION="X.Y.Z"
DATE=$(date +%Y-%m-%d)
COMMIT=$(git rev-parse --short HEAD)

# 1. Root version.json
# Edit: "version": "X.Y.Z", "updated_at": "YYYY-MM-DD"

# 2. Binary version
# Edit: internal/config/config.go
# Change: var Version = "X.Y.Z"

# 3. Knowledge-base version
# Edit: knowledge-base/version.json
# Update: version, commit, updated_at, all sections

# 4. Admin version
# Edit: admin/version.json
# Update: version, updated_at (status/completeness only if changed)
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

**IMPORTANT**: Commit all version files together.

```bash
# Check if there are changes to commit
git status --porcelain

# Stage all version files + CHANGELOG
git add version.json internal/config/config.go knowledge-base/version.json admin/version.json CHANGELOG.md docs/changelog.json

# Verify what's staged
git diff --cached --name-only

# Only commit if there are changes
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

### 7. Build and Upload Release

Use `scripts/release.sh` for fast local release (~30s vs ~4min CI):

```bash
source .env  # loads GITHUB_PAT_FAZT
./scripts/release.sh vX.Y.Z
```

This script:
- Builds all 4 platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64)
- Creates GitHub release
- Uploads all assets

### 8. Push

```bash
# Push (Actions will skip since 4 assets already exist)
git push origin master && git push origin vX.Y.Z
```

### 9. Upgrade Local Binary and Service

**IMPORTANT**: Always use the canonical `fazt upgrade` command to test the actual user experience. Never manually build with `go build` during releases.

```bash
# Upgrade local binary using the canonical upgrade command
# This tests the actual user experience and ensures upgrade works elegantly
fazt upgrade
fazt --version

# Restart local systemd service (if running)
systemctl --user restart fazt-local 2>/dev/null && echo "Local service restarted" || echo "No local service running"

# Wait for service to start
sleep 2

# Verify local peer upgraded (if configured as peer)
fazt @local status 2>/dev/null || echo "Local not configured as peer"
```

**Why use `fazt upgrade` instead of `go build`?**
- Tests the same upgrade path users will use
- Verifies GitHub release assets are valid
- Ensures upgrade experience is elegant
- Catches issues users would encounter
- Maintains developer empathy

### 10. Upgrade All Peers

```bash
# List all configured peers
fazt peer list

# Upgrade each peer (idempotent - returns "already latest" if up-to-date)
fazt @local upgrade 2>/dev/null || echo "Local peer not configured"
fazt @zyt upgrade   2>/dev/null || echo "zyt peer not configured"

# Or upgrade all peers automatically:
# fazt peer list --format json | jq -r '.[].name' | xargs -I {} fazt @{} upgrade
```

**Note**: Each upgrade is idempotent and returns "already latest" if the peer is already upgraded.

## Quick Reference

| Check | Command |
|-------|---------|
| All versions in sync? | See VERSIONING.md verification commands |
| Root version | `cat version.json \| jq -r '.version'` |
| Binary version | `grep "var Version" internal/config/config.go` |
| KB version | `cat knowledge-base/version.json \| jq -r '.version'` |
| Admin version | `cat admin/version.json \| jq -r '.version'` |
| Latest tag | `git describe --tags --abbrev=0` |
| Tag exists? | `git tag -l "vX.Y.Z"` |
| Need to push? | `git status -sb` (shows "ahead") |
| All peer versions | `fazt peer list` |
| Specific peer version | `fazt @<name> status` |

## Version Numbering

- **Patch** (0.9.5): Bug fixes
- **Minor** (0.10.0): New features, backwards compatible
- **Major** (1.0.0): Breaking changes
