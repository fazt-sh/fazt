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

### 9. Install Locally and Restart Service

```bash
# Build and install binary
go build -o ~/.local/bin/fazt ./cmd/server
fazt --version

# Restart local systemd service (if running)
systemctl --user restart fazt-local 2>/dev/null && echo "Local service restarted" || echo "No local service running"

# Wait for service to start
sleep 2

# Verify local peer upgraded (if configured as peer)
fazt peer status local 2>/dev/null || echo "Local not configured as peer"
```

### 10. Upgrade All Peers

```bash
# List all configured peers
fazt peer list

# Upgrade each peer (idempotent - returns "already latest" if up-to-date)
fazt peer upgrade local 2>/dev/null || echo "Local peer not configured"
fazt peer upgrade zyt   2>/dev/null || echo "zyt peer not configured"

# Or upgrade all peers automatically:
# fazt peer list --format json | jq -r '.[].name' | xargs -I {} fazt peer upgrade {}
```

**Note**: Each upgrade is idempotent and returns "already latest" if the peer is already upgraded.

## Quick Reference

| Check | Command |
|-------|---------|
| Current version | `grep "var Version" internal/config/config.go` |
| Latest tag | `git describe --tags --abbrev=0` |
| Tag exists? | `git tag -l "vX.Y.Z"` |
| Need to push? | `git status -sb` (shows "ahead") |
| All peer versions | `fazt peer list` |
| Specific peer version | `fazt peer status <name>` |

## Version Numbering

- **Patch** (0.9.5): Bug fixes
- **Minor** (0.10.0): New features, backwards compatible
- **Major** (1.0.0): Breaking changes
