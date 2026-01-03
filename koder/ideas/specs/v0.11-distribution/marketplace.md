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

A marketplace is just a Git repo with a specific structure:

```
fazt-marketplace/
├── registry.json           # App index
├── apps/
│   ├── blog/
│   │   ├── app.json
│   │   ├── index.html
│   │   └── api/
│   │       └── main.js
│   └── todo/
│       ├── app.json
│       └── ...
```

## Registry Format

### registry.json

```json
{
  "name": "Official Fazt Apps",
  "url": "https://github.com/fazt-sh/marketplace",
  "apps": [
    {
      "name": "blog",
      "version": "1.2.0",
      "description": "Simple markdown blog",
      "path": "apps/blog"
    },
    {
      "name": "todo",
      "version": "2.0.1",
      "description": "Task management app",
      "path": "apps/todo"
    }
  ]
}
```

## CLI Commands

### Add Marketplace

```bash
# Add official marketplace
fazt marketplace add https://github.com/fazt-sh/marketplace

# Add community marketplace
fazt marketplace add https://github.com/alice/my-apps
```

### Sync Registry

```bash
# Fetch latest registry from all marketplaces
fazt marketplace sync

# Output:
# Syncing fazt-sh/marketplace... 15 apps
# Syncing alice/my-apps... 3 apps
# Total: 18 apps available
```

### Search Apps

```bash
fazt app search blog

# Output:
# fazt-sh/blog     v1.2.0  Simple markdown blog
# alice/my-blog    v0.3.0  Photo blog with gallery
```

### Install App

```bash
# Install from marketplace
fazt app install blog

# Install specific version
fazt app install blog@1.1.0

# Install from specific marketplace
fazt app install alice/my-blog
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
1. fazt app install blog
2. Resolve: Find "blog" in registry cache
3. Fetch: Download app folder from Git repo
4. Validate: Check app.json, verify signatures (future)
5. Provision: Generate UUID, create VFS entries
6. Configure: Prompt for required env vars
7. Deploy: App is live at blog.example.com
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

| Type          | Description         | Example                 |
| ------------- | ------------------- | ----------------------- |
| `personal`    | Deployed via CLI    | `fazt deploy ./my-site` |
| `marketplace` | Installed from repo | `fazt app install blog` |

## Update Detection

```go
func (m *Marketplace) CheckUpdates(app *App) (*Update, error) {
    if app.SourceType != "marketplace" {
        return nil, nil // Personal apps don't have updates
    }

    registry := m.GetRegistry(app.SourceURL)
    latest := registry.GetApp(app.Name)

    if semver.Compare(latest.Version, app.SourceVersion) > 0 {
        return &Update{
            From: app.SourceVersion,
            To:   latest.Version,
        }, nil
    }

    return nil, nil
}
```

## Security Considerations

### Trust Model

Marketplaces are trusted by the user who adds them:

```bash
# This is a trust decision
fazt marketplace add https://github.com/untrusted/apps
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

Apps not from a marketplace are "personal":

```bash
fazt deploy ./my-site --slug my-site

# Source is recorded as:
# source_type: 'personal'
# source_url: NULL
# source_version: NULL
```

Personal apps:
- Don't receive update notifications
- Can be converted to marketplace by publishing
- Are labeled "Sideloaded" in dashboard

## Open Questions

1. **Private Marketplaces**: Support authenticated Git repos?
2. **Dependency Resolution**: Apps depending on other apps?
3. **Rollback**: Revert to previous version on failed update?
