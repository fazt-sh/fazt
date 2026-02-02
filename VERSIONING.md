# Fazt Versioning System

**Purpose**: Ensure all parallel systems evolve together with tight synchronization.

## Overview

Fazt uses **unified versioning** across all components. When the version changes, ALL version markers must be updated together.

**Current Version**: Check `version.json` at repo root (source of truth).

---

## Version Files

All version files must stay in sync. On each release, verify and update ALL of these:

| File | Purpose | Format |
|------|---------|--------|
| `version.json` | Monorepo source of truth | Unified version + component metadata |
| `internal/config/config.go` | Binary version (embedded) | `var Version = "X.Y.Z"` |
| `knowledge-base/version.json` | Docs compatibility marker | Version + commit + sections |
| `admin/version.json` | Admin UI version + maturity | Version + status + completeness |
| Git tags | Release version | `vX.Y.Z` |
| Remote peers | Deployed versions | Check via `fazt @<peer> status` |

### 1. Root `version.json` (Source of Truth)

**Location**: `./version.json`

```json
{
  "version": "0.22.0",
  "updated_at": "2026-02-02",
  "repository": "monorepo",
  "versioning": "unified",
  "components": {
    "fazt-binary": {
      "path": "internal/",
      "status": "stable|beta|alpha",
      "completeness": "0-100%"
    },
    "admin": { ... },
    "fazt-sdk": { ... },
    "knowledge-base": { ... }
  }
}
```

**Purpose**:
- Single source of truth for monorepo version
- Tracks all components and their maturity status
- Used by tooling to verify consistency

### 2. Binary Version (`internal/config/config.go`)

**Location**: `internal/config/config.go:81`

```go
var Version = "0.22.0"
```

**Purpose**:
- Embedded in the compiled binary
- Returned by `fazt --version`
- Used for peer version comparison

### 3. Knowledge Base (`knowledge-base/version.json`)

**Location**: `./knowledge-base/version.json`

```json
{
  "version": "0.22.0",
  "commit": "e21569d",
  "updated_at": "2026-02-02",
  "sections": {
    "skills/app": "0.22.0",
    "agent-context": "0.22.0",
    "workflows": "0.22.0"
  }
}
```

**Purpose**:
- Compatibility marker: docs verified against this binary version
- Tracks which sections are up-to-date
- Bump on every release (even if content unchanged)

### 4. Admin UI (`admin/version.json`)

**Location**: `./admin/version.json`

```json
{
  "version": "0.22.0",
  "status": "alpha",
  "completeness": "15%",
  "updated_at": "2026-02-02"
}
```

**Purpose**:
- Tracks admin UI version
- Shows maturity status and progress
- Independent status marker from binary

### 5. Git Tags

**Location**: Git repository

```bash
git tag -l "v*"
```

**Purpose**:
- Marks release points
- Triggers GitHub release workflow
- Used for changelogs and version comparison

### 6. Remote Peers

**Check**: `fazt @<peer> status`

**Purpose**:
- Deployed binary versions
- Runtime version verification
- Upgrade tracking

---

## Status System

Components use status markers to indicate maturity:

| Status | Meaning | Backward Compat | Breaking Changes |
|--------|---------|-----------------|------------------|
| **stable** | Production-ready, complete | Required | Major version only |
| **beta** | Feature-complete, testing | Best effort | Minor version OK |
| **alpha** | In development | Not guaranteed | Anytime |

**Current Component Status** (from `version.json`):

| Component | Status | Completeness | Notes |
|-----------|--------|--------------|-------|
| fazt-binary | stable | 100% | Core is production-ready |
| admin | alpha | 15% | Dashboard functional, rest in progress |
| fazt-sdk | alpha | 20% | Basic API coverage with mocks |
| knowledge-base | stable | 80% | Comprehensive, ongoing updates |

**Completeness** indicates progress toward full API/feature parity.

---

## Release Workflow

### Pre-Release Checklist

Before bumping version, verify:

1. **All tests pass**
   ```bash
   go test ./...
   ```

2. **Binary compiles**
   ```bash
   go build -o fazt ./cmd/server
   ```

3. **No uncommitted changes**
   ```bash
   git status --short
   ```

### Version Update Steps

**CRITICAL**: Update ALL version files in a single commit.

```bash
# 1. Decide new version (X.Y.Z)
NEW_VERSION="0.23.0"
DATE=$(date +%Y-%m-%d)
COMMIT=$(git rev-parse --short HEAD)

# 2. Update root version.json
# Edit: version, updated_at

# 3. Update binary version
# Edit: internal/config/config.go (var Version)

# 4. Update knowledge-base/version.json
# Edit: version, commit, updated_at, sections

# 5. Update admin/version.json
# Edit: version, updated_at (status/completeness only if changed)

# 6. Update CHANGELOG.md (add release entry)

# 7. Commit everything together
git add version.json internal/config/config.go knowledge-base/version.json admin/version.json CHANGELOG.md
git commit -m "release: v${NEW_VERSION}

<release notes>

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"

# 8. Tag
git tag v${NEW_VERSION}

# 9. Build and release
./scripts/release.sh v${NEW_VERSION}

# 10. Push
git push origin master && git push origin v${NEW_VERSION}
```

### Post-Release Steps

```bash
# 1. Install locally
go build -o ~/.local/bin/fazt ./cmd/server
systemctl --user restart fazt-local

# 2. Upgrade all peers
fazt @local upgrade
fazt @zyt upgrade

# 3. Verify versions
fazt --version
fazt @local status
fazt @zyt status
```

---

## Verification Commands

Use these to check version consistency:

```bash
# All version files
cat version.json | jq -r '.version'
grep "var Version" internal/config/config.go | grep -oE '[0-9]+\.[0-9]+\.[0-9]+'
cat knowledge-base/version.json | jq -r '.version'
cat admin/version.json | jq -r '.version'
git describe --tags --abbrev=0
fazt --version | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+'

# All should show same version!

# Remote peers
fazt @local status | grep Version
fazt @zyt status | grep Version

# Component status
cat version.json | jq -r '.components | to_entries[] | "\(.key): \(.value.status) (\(.value.completeness))"'
```

**Expected output**: All version numbers match.

**If drift detected**:
1. Identify which file is out of sync
2. Update it to match source of truth (`version.json`)
3. Commit with descriptive message
4. Consider why drift occurred and fix the process

---

## Testing Requirements

### On Version Bump

| Component | Test Required |
|-----------|---------------|
| **fazt-binary** | `go test ./...` |
| **CLI commands** | Manual smoke test (app deploy, list, status) |
| **Admin UI** | Load in browser, verify dashboard |
| **Knowledge-base** | Spot-check docs accuracy |
| **Remote peers** | Upgrade and verify health |

### Smoke Test Checklist

After deploying new version:

```bash
# 1. Binary works
fazt --version

# 2. Local server healthy
fazt @local status

# 3. App operations
fazt @local app list
fazt @local app deploy ./servers/local/hello

# 4. Remote peer healthy
fazt @zyt status

# 5. Admin UI loads
# Visit: http://admin.192.168.64.3.nip.io:8080
# Visit: https://admin.zyt.app
```

---

## Skills Integration

All session management skills must check version consistency:

### `/open` (Session Start)

**Must check**:
- ✅ Root `version.json`
- ✅ Binary version (`fazt --version`)
- ✅ Git tag (`git describe --tags`)
- ✅ `knowledge-base/version.json`
- ✅ `admin/version.json`
- ✅ Remote peers (`fazt @<peer> status`)

**Output**: Report any version drift with fix commands.

### `/close` (Session End)

**Must check**:
- ✅ Root `version.json`
- ✅ Binary version (`config.go`)
- ✅ Git tag
- ✅ `knowledge-base/version.json`
- ✅ `admin/version.json`
- ✅ Remote peers

**Must update**: `knowledge-base/version.json` if releasing.

### `/release` (Release Process)

**Must update ALL**:
1. Root `version.json`
2. `internal/config/config.go`
3. `knowledge-base/version.json`
4. `admin/version.json`
5. Git tag
6. CHANGELOG.md

**Must verify**: All files updated in single commit before tagging.

---

## Common Issues

### Version Drift

**Symptom**: `version.json` shows v0.20.0, but binary is v0.22.0

**Cause**: Version file not updated during release

**Fix**:
```bash
# Update to match binary
NEW_VERSION=$(fazt --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')
# Edit version.json manually or via script
git commit -m "fix: Sync version.json to v${NEW_VERSION}"
```

### Peer Version Mismatch

**Symptom**: `fazt @zyt status` shows old version after release

**Cause**: Peer not upgraded after release

**Fix**:
```bash
fazt @zyt upgrade
fazt @zyt status  # Verify
```

### Knowledge Base Out of Sync

**Symptom**: `knowledge-base/version.json` behind current release

**Cause**: Forgot to bump on release

**Fix**:
```bash
# Update to current version
# Edit knowledge-base/version.json
git commit -m "docs: Sync knowledge-base version to v${VERSION}"
```

---

## Version Numbering

Fazt uses semantic versioning (while in 0.x):

- **Patch** (0.22.1): Bug fixes, no breaking changes
- **Minor** (0.23.0): New features, may have breaking changes (0.x allows this)
- **Major** (1.0.0): Stable API, breaking changes only with major bump

**Current state**: v0.x series - rapid iteration, breaking changes allowed in minor versions.

---

## Automation Opportunities

Future improvements to enforce consistency:

1. **Pre-commit hook**: Check all version files match
2. **CI check**: Fail build if versions drift
3. **Release script**: Update all files atomically
4. **Version command**: `fazt version sync` to fix drift

---

## References

- Release workflow: `.claude/commands/release.md`
- Session management: `.claude/commands/open.md`, `.claude/commands/close.md`
- Monorepo structure: `CLAUDE.md`
- Component details: `version.json` (components section)
