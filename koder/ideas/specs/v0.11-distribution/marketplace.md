# Marketplace

## Summary

Marketplaces are Git repositories that serve as app sources. Fazt adopts the
"Linux distro" model: decentralized repositories, local package management,
and transparent provenance tracking.

## Rationale

### Rejecting Centralization

We reject the App Store model:
- No single point of control
- No approval process
- No platform fees

### The Git Model

Any Git repo can contain Fazt apps. Install directly via `repo/folder` path:

```bash
# Install from any GitHub repo + subfolder
fazt app install github.com/fazt-sh/store/haikus
fazt app install github.com/someproject/repo/examples/demo

# The official store is just a monorepo
fazt app install github.com/fazt-sh/store/devlog
```

Store structure (no registry.json required):

```
github.com/fazt-sh/store/
├── haikus/                 # Flagship blog example
│   ├── app.json
│   ├── _posts/
│   └── _plugins/
├── devlog/                 # Developer blog
│   ├── app.json
│   └── ...
├── portfolio/              # Portfolio site
└── docs/                   # Documentation template
```

Each folder is a standalone app. No index needed - the folder IS the app.

## Registry Format (Optional)

A `registry.json` at repo root enables search and versioning. Without it,
apps are still installable via direct path.

### registry.json

```json
{
  "name": "Fazt Store",
  "apps": [
    {
      "name": "haikus",
      "version": "1.0.0",
      "description": "Japanese haiku collection",
      "path": "haikus"
    },
    {
      "name": "devlog",
      "version": "2.1.0",
      "description": "Developer blog with code features",
      "path": "devlog"
    }
  ]
}
```

## CLI Commands

### Install App

```bash
# Install from any GitHub repo/folder (primary method)
fazt app install github.com/fazt-sh/store/haikus
fazt app install github.com/someuser/repo/my-app

# Short form for official store
fazt app install fazt-sh/store/haikus
```

### Update App

```bash
# Update single app
fazt app update blog

# Update all apps
fazt app update --all

# Check for updates
fazt app outdated
```

### Remove App

```bash
fazt app remove blog

# Keep data, remove code
fazt app remove blog --keep-data
```

## Installation Flow

```
1. fazt app install github.com/fazt-sh/store/haikus
2. Parse: Extract repo URL and folder path
3. Fetch: Clone/download the specific folder from Git
4. Validate: Check app.json exists
5. Provision: Generate UUID, create VFS entries
6. Configure: Prompt for required env vars (if any)
7. Deploy: App is live at haikus.example.com
```

## Source Tracking

### Database Schema

```sql
CREATE TABLE apps (
    uuid TEXT PRIMARY KEY,
    name TEXT,
    source_type TEXT,       -- 'personal' | 'marketplace'
    source_url TEXT,        -- Git repo URL
    source_version TEXT,    -- Installed version
    installed_at INTEGER,
    updated_at INTEGER
);
```

### Source Types

| Type       | Description              | Example                                      |
|------------|--------------------------|----------------------------------------------|
| `personal` | Deployed from local dir  | `fazt deploy ./my-site`                      |
| `git`      | Installed from Git repo  | `fazt app install github.com/user/repo/app`  |

## Update Detection

```go
func CheckUpdates(app *App) (*Update, error) {
    if app.SourceType != "git" {
        return nil, nil // Personal apps don't have updates
    }

    // Fetch latest commit for the app folder
    latest := git.GetLatestCommit(app.SourceURL, app.SourcePath)

    if latest.SHA != app.SourceVersion {
        return &Update{
            From: app.SourceVersion[:7],
            To:   latest.SHA[:7],
        }, nil
    }

    return nil, nil
}
```

## Security Considerations

### Trust Model

Installing from any Git URL is a trust decision:

```bash
# User decides to trust this source
fazt app install github.com/untrusted/repo/app
```

### Future: Signature Verification

```json
{
  "name": "blog",
  "version": "1.2.0",
  "signature": "base64-encoded-signature",
  "signer": "fazt-sh"
}
```

### Permission Review

During installation, show required permissions:

```
Installing blog v1.2.0...

Permissions requested:
  - storage:kv      Store settings
  - storage:ds      Store blog posts
  - net:fetch       Fetch external images

Proceed? [y/N]
```

## Personal Apps

Apps deployed from local directory are "personal":

```bash
fazt deploy ./my-site

# Source is recorded as:
# source_type: 'personal'
# source_url: NULL
# source_version: NULL
```

Personal apps:
- Don't receive update notifications
- Can be pushed to a Git repo and re-installed as `git` type
- Are labeled "Local" in dashboard

## Open Questions

1. **Private Repos**: Support authenticated Git repos (SSH keys, tokens)?
2. **Rollback**: Revert to previous version on failed update?
3. **Pinning**: Pin to specific commit vs always latest?
