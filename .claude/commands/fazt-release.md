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

### 7. Build All Platforms Locally

```bash
# Build admin if needed
[ -d "internal/assets/system/admin" ] || (npm run build --prefix admin && cp -r admin/dist internal/assets/system/admin)

# Build all platforms
for PLATFORM in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do
  OS=${PLATFORM%/*}
  ARCH=${PLATFORM#*/}
  echo "Building $OS/$ARCH..."
  CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build -ldflags="-w -s" -o fazt ./cmd/server
  tar -czvf "fazt-vX.Y.Z-${OS}-${ARCH}.tar.gz" fazt
done
```

### 8. Create Release and Upload Assets

Uses `GITHUB_PAT_FAZT` from `.env` for fast local release (skips waiting for CI):

```bash
source .env

# Create release
RELEASE=$(curl -s -X POST \
  -H "Authorization: token $GITHUB_PAT_FAZT" \
  -H "Accept: application/vnd.github.v3+json" \
  "https://api.github.com/repos/fazt-sh/fazt/releases" \
  -d '{"tag_name":"vX.Y.Z","name":"vX.Y.Z","generate_release_notes":true}')

RELEASE_ID=$(echo "$RELEASE" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")

# Upload each asset
for ASSET in fazt-vX.Y.Z-*.tar.gz; do
  echo "Uploading $ASSET..."
  curl -s -X POST \
    -H "Authorization: token $GITHUB_PAT_FAZT" \
    -H "Content-Type: application/gzip" \
    "https://uploads.github.com/repos/fazt-sh/fazt/releases/$RELEASE_ID/assets?name=$ASSET" \
    --data-binary "@$ASSET"
done

# Cleanup
rm -f fazt-*.tar.gz fazt
```

### 9. Push and Install

```bash
# Push (Actions will skip since 4 assets already exist)
git push origin master && git push origin vX.Y.Z

# Install locally
cp fazt ~/.local/bin/fazt
fazt --version
```

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
