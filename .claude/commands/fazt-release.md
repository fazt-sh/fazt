# Fazt Release Skill

Release a new version of fazt with proper versioning, documentation, and deployment.

## Pre-flight Checks

1. **Verify clean state**: `git status` should be clean (no uncommitted changes)
2. **Run tests**: `go test ./...` must pass
3. **Check current version**: Read `internal/config/config.go` for `Version`

## Release Steps

### 1. Update Version

Edit `internal/config/config.go`:
```go
var Version = "X.Y.Z"  // Update to new version
```

### 2. Update CHANGELOG.md

Add entry at top (after header):
```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- ...

### Changed
- ...

### Fixed
- ...
```

### 3. Update Website Changelog

Add entry at top of `docs/changelog.json`:
```json
{
    "version": "vX.Y.Z",
    "title": "Short Title",
    "description": "One-line summary of changes.",
    "created_at": "YYYY-MM-DD"
}
```

This powers the changelog on the GitHub Pages website.

### 4. Commit and Tag

```bash
git add -A
git commit -m "release: vX.Y.Z

<summary of changes>

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git tag vX.Y.Z
```

### 5. Build and Install Locally

```bash
go build -ldflags "-X github.com/fazt-sh/fazt/internal/config.Version=X.Y.Z" -o ~/.local/bin/fazt ./cmd/server
```

Verify: `fazt --version`

Note: `~/.local/bin` is in PATH and doesn't require sudo. The `setcap` for
binding ports 80/443 is only needed for local servers, not CLI usage.

### 6. Push to GitHub

```bash
git push origin master --tags
```

### 7. Wait for CI

Poll until complete:
```bash
curl -s "https://api.github.com/repos/fazt-sh/fazt/actions/runs?per_page=1" | jq '.workflow_runs[0] | {status, conclusion}'
```

Wait until `status: completed` and `conclusion: success`.

### 8. Upgrade Remote Servers

For each configured peer:
```bash
fazt remote upgrade <peer-name>
```

If upgrade fails due to service file issues, user needs to SSH and run:
```bash
curl -fsSL https://raw.githubusercontent.com/fazt-sh/fazt/master/install.sh | sudo bash
```

## Install Script URL

The install script should be fetched from:
```
https://raw.githubusercontent.com/fazt-sh/fazt/master/install.sh
```

NOT from fazt.sh (domain not purchased yet).

## Version Numbering

- **Patch** (0.9.1): Bug fixes, no new features
- **Minor** (0.10.0): New features, backwards compatible
- **Major** (1.0.0): Breaking changes

## Post-Release

1. Verify release on GitHub: https://github.com/fazt-sh/fazt/releases
2. Test install script: `curl -fsSL https://raw.githubusercontent.com/fazt-sh/fazt/master/install.sh | bash`
3. Verify remote server versions: `fazt remote status`

## Local Paths (Dev VM)

| What | Path |
|------|------|
| Binary | `~/.local/bin/fazt` |
| Client DB | `~/.config/fazt/data.db` |
| Peers | Stored in client DB (peers table) |

The client DB contains peer configurations. Moving the DB moves everything.
