# Everything is an App

## Summary

v0.8 replaces the concept of "Sites" with "Apps". An App is a first-class
entity with a stable UUID, independent of its subdomain or domain mapping.

## Rationale

### The Problem with Sites

In v0.7, sites are identified by subdomain:
- `blog` lives at `blog.example.com`
- Rename the subdomain? Lose analytics, logs, settings
- Multiple domains for one site? Not possible

### The App Model

Apps are identified by UUID:
- `app_x9z2k` is the stable identity
- `blog.example.com` is just a routing pointer
- Rename, add domains, move—data stays intact

## Data Model

### Sites Table (v0.7)

```sql
CREATE TABLE sites (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE,      -- "blog" (also the subdomain)
    created_at TIMESTAMP
);
```

### Apps Table (v0.8)

```sql
CREATE TABLE apps (
    id INTEGER PRIMARY KEY,
    uuid TEXT UNIQUE,       -- "app_x9z2k" (stable identity)
    name TEXT,              -- "My Blog" (display name)
    slug TEXT,              -- "blog" (subdomain, can change)
    manifest TEXT,          -- JSON: app.json contents
    source TEXT,            -- "personal" | "marketplace:repo-url"
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE app_domains (
    app_uuid TEXT,
    domain TEXT UNIQUE,     -- "blog.example.com"
    is_primary BOOLEAN,
    created_at TIMESTAMP
);
```

## Benefits

### 1. Stable Identity

```
# Rename without data loss
fazt app rename app_x9z2k --slug new-blog

# Data stays keyed to app_x9z2k
analytics, logs, storage all preserved
```

### 2. Multi-Domain Support

```
# Multiple domains point to same app
fazt net domain map blog.example.com app_x9z2k
fazt net domain map writings.me app_x9z2k
```

### 3. Private Routing

Derive sensitive endpoints from UUID:
- Webhook: `https://example.com/hook/x9z2k`
- Email: `x9z2k@example.com` (future)
- IPC: App A can reference App B by UUID

### 4. Marketplace Ready

```json
{
  "uuid": "app_x9z2k",
  "source": "marketplace:github.com/fazt-apps/blog",
  "version": "1.2.0",
  "installed_at": "2025-01-15T10:00:00Z"
}
```

## System Apps

The kernel itself provides apps:

| App   | UUID               | Subdomain | Purpose      |
| ----- | ------------------ | --------- | ------------ |
| Admin | `app_system_admin` | `admin.*` | Dashboard    |
| Root  | `app_system_root`  | `root.*`  | Landing page |
| 404   | `app_system_404`   | `404.*`   | Error page   |

System apps cannot be deleted. They can be customized by replacing their VFS
contents.

## Migration

### From v0.7 Sites

```sql
-- Migration 008: Convert sites to apps
ALTER TABLE sites ADD COLUMN uuid TEXT;
ALTER TABLE sites ADD COLUMN manifest TEXT;
ALTER TABLE sites ADD COLUMN source TEXT DEFAULT 'personal';

-- Generate UUIDs for existing sites
UPDATE sites SET uuid = 'app_' || hex(randomblob(4))
WHERE uuid IS NULL;
```

### CLI Compatibility

```bash
# v0.7 style still works (translates to app operations)
fazt deploy ./my-site --name blog

# v0.8 style
fazt app deploy ./my-site --slug blog
fazt app deploy ./my-site --uuid app_x9z2k  # Update existing
```

## App Lifecycle

```
Created                    Active                    Archived
   │                          │                          │
   │  fazt app deploy         │  Running                 │  Soft delete
   │  fazt app install        │  Serving traffic         │  Data preserved
   │                          │  Analytics collecting    │  No traffic
   ├──────────────────────────┼──────────────────────────┤
   │                          │                          │
   └──── uuid generated ──────┴──── uuid immutable ──────┘
```

## Open Questions

1. **Cascading Deletes**: When app is deleted, delete files + analytics?
2. **UUID Format**: `app_` prefix + 8 hex chars? Or full UUIDv4?
3. **System App Customization**: Allow replacing system app content?
