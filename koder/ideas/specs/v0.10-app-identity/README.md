# App Identity and Lineage

**Version**: v0.10
**Status**: Draft
**Created**: 2026-01-19

## Overview

This spec defines a new app identity model that separates permanent identity
(UUID) from mutable labels (subdomains), and introduces lineage tracking for
fork/clone workflows.

## Goals

1. **Stable Identity**: Apps have permanent UUIDs that never change
2. **Flexible Labels**: Subdomains can be reassigned trivially
3. **Lineage Tracking**: Know where an app came from and its history
4. **Agent-Friendly**: Support LLM agent workflows (test, fork, promote)
5. **Safe Testing**: Fork production, test safely, promote when ready

## Non-Goals

- Multi-tenant apps (single owner per fazt instance)
- Complex permission models
- Git-like branching (simpler fork/promote model)

---

## Data Model

### Apps Table

```sql
CREATE TABLE apps (
    -- Identity (immutable)
    id TEXT PRIMARY KEY,              -- "app_7f3k9x2m" (UUID, never changes)

    -- Label (mutable, can be NULL)
    label TEXT UNIQUE,                -- "myapp" (subdomain, reassignable)

    -- Lineage
    original_id TEXT,                 -- Root ancestor (self if original)
    forked_from_id TEXT,              -- Immediate parent (NULL if original)

    -- Metadata
    source TEXT DEFAULT 'deploy',     -- 'deploy', 'git', 'fork', 'system'
    manifest TEXT,                    -- JSON manifest
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    -- Source tracking (for git-installed apps)
    source_url TEXT,
    source_ref TEXT,
    source_commit TEXT,
    installed_at DATETIME
);

CREATE INDEX idx_apps_label ON apps(label);
CREATE INDEX idx_apps_original ON apps(original_id);
CREATE INDEX idx_apps_forked_from ON apps(forked_from_id);
```

### Key Concepts

| Field | Purpose | Mutability |
|-------|---------|------------|
| `id` | Permanent identity | Immutable |
| `label` | Subdomain / public name | Mutable, nullable, unique |
| `original_id` | Root ancestor in lineage | Immutable after creation |
| `forked_from_id` | Immediate parent | Immutable after creation |

### ID Format

```
app_{nanoid}

Examples:
- app_7f3k9x2m
- app_x9p2q4wn
- app_m3n8k1vz
```

Use 8-character nanoid for readability. Collision-resistant for single-instance.

---

## Lineage Model

### Example Lineage Tree

```
app_001 "myapp" (original)
│
├── app_002 "myapp-debug"
│   │   forked_from: app_001
│   │   original: app_001
│   │
│   └── app_003 "myapp-fix"
│           forked_from: app_002
│           original: app_001
│
└── app_004 (no label - unlisted)
        forked_from: app_001
        original: app_001
```

### Lineage Rules

1. **Original apps**: `original_id = id`, `forked_from_id = NULL`
2. **Forked apps**: `original_id` = root ancestor, `forked_from_id` = parent
3. **Deletion**: Deleting an app doesn't affect descendants' lineage fields
4. **Orphans OK**: `forked_from_id` can reference deleted apps

---

## Label Management

Labels are subdomains that point to apps. They are:

- **Mutable**: Can be changed anytime
- **Unique**: Only one app can have a given label
- **Optional**: Apps can exist without labels (unlisted)
- **Swappable**: Two apps can swap labels atomically

### Label Operations

```bash
# Assign label to app
fazt app rename app_7f3k9x2m --to myapp
# → myapp.zyt.app now serves app_7f3k9x2m

# Remove label (app becomes unlisted)
fazt app rename myapp --to ""
# → myapp.zyt.app returns 404, app still accessible by ID

# Reassign label to different app
fazt app rename app_x9p2q4wn --to myapp
# → myapp.zyt.app now serves app_x9p2q4wn

# Swap labels between two apps (atomic)
fazt app swap myapp myapp-v2
# → Labels exchanged, no downtime
```

### Accessing Unlisted Apps

Apps without labels can still be accessed via their ID:

```
http://app_7f3k9x2m.zyt.app    (direct ID access)
```

This enables:
- Preview deployments
- Testing before assigning label
- Keeping old versions accessible

---

## CLI Commands

Labels are managed through standard `fazt app` commands. No separate namespace.

### List and Info

```bash
# List all apps
fazt app list
# ID            LABEL        FORKED-FROM   SOURCE   CREATED
# app_7f3k9x2m  myapp        -             deploy   2026-01-15
# app_m3n8k1vz  myapp-test   app_7f3k9x2m  fork     2026-01-19
# app_x9p2q4wn  othelo       -             deploy   2026-01-10
# app_p4q7r2st  -            app_7f3k9x2m  fork     2026-01-19

# Show app details (by label or ID)
fazt app info myapp
# ID:          app_7f3k9x2m
# Label:       myapp
# URL:         https://myapp.zyt.app
# Original:    app_7f3k9x2m (self)
# Forked from: -
# Source:      deploy
# Created:     2026-01-19
# Storage:     42 keys
```

### Deploy

```bash
# Deploy by label or ID
fazt app deploy ./src --to myapp
fazt app deploy ./src --to app_7f3k9x2m

# Deploy creates app if doesn't exist
fazt app deploy ./src --to newapp
# → Created app_x9p2q4wn with label "newapp"
```

### Rename (Label Assignment)

```bash
# Assign or change label
fazt app rename app_7f3k9x2m --to myapp
# → app_7f3k9x2m accessible at myapp.zyt.app

# Change existing label
fazt app rename myapp --to my-new-app
# → myapp.zyt.app becomes my-new-app.zyt.app

# Remove label (unlisted, accessible by ID only)
fazt app rename myapp --to ""
# → Accessible only at app_7f3k9x2m.zyt.app

# Swap labels between two apps
fazt app swap myapp myapp-v2
# → Labels exchanged instantly
```

### Fork

```bash
# Fork with label (copies files + storage)
fazt app fork myapp --as myapp-test
# → Created app_m3n8k1vz "myapp-test"

# Fork without label
fazt app fork myapp
# → Created app_p4q7r2st (unlisted)

# Fork without storage
fazt app fork myapp --as myapp-clean --no-storage
```

### Lineage

```bash
# Show lineage tree
fazt app lineage myapp
# app_7f3k9x2m "myapp" (original)
# ├── app_m3n8k1vz "myapp-test"
# │   └── app_q2w3e4r5 "myapp-test-v2"
# └── app_p4q7r2st (unlisted)

# Delete app
fazt app delete myapp

# Delete app and all forks
fazt app delete myapp --with-forks

# Delete only forks
fazt app prune myapp
```

---

## Storage Handling

### On Fork

When forking an app, storage is copied by default:

```sql
-- Copy all KV entries from parent to new app
INSERT INTO app_kv (app_id, key, value, expires_at, created_at, updated_at)
SELECT 'app_NEW', key, value, expires_at, created_at, CURRENT_TIMESTAMP
FROM app_kv WHERE app_id = 'app_PARENT';
```

Options:
- `--no-storage`: Fork without copying storage (clean slate)
- `--storage-snapshot`: Copy storage at specific point (future)

### Storage Scoping

Storage is always scoped by `app_id` (the UUID), not the label:

```sql
-- This query works regardless of label changes
SELECT * FROM app_kv WHERE app_id = 'app_7f3k9x2m';
```

---

## Agent Testing Endpoints

Reserved `/_fazt/` endpoints for agent workflows (requires auth).

### Storage Inspection

```
GET  /_fazt/info              → App metadata, lineage, storage stats
GET  /_fazt/storage           → List all storage keys
GET  /_fazt/storage/:key      → Get specific key value
POST /_fazt/snapshot          → Create named snapshot
POST /_fazt/restore/:name     → Restore to snapshot
GET  /_fazt/snapshots         → List available snapshots
```

### Serverless Debugging

```
GET  /_fazt/logs              → Recent serverless execution logs
GET  /_fazt/logs?limit=20     → Last N executions
GET  /_fazt/errors            → Recent errors with stack traces
GET  /_fazt/errors?limit=10   → Last N errors
```

**Log entry example:**
```json
{
  "id": "log_abc123",
  "timestamp": "2026-01-19T15:30:00Z",
  "method": "POST",
  "path": "/api/action",
  "status": 200,
  "duration_ms": 45,
  "storage_ops": ["get:scores", "set:scores"],
  "error": null
}
```

**Error entry example:**
```json
{
  "id": "err_xyz789",
  "timestamp": "2026-01-19T15:31:00Z",
  "method": "POST",
  "path": "/api/broken",
  "error": "TypeError: Cannot read property 'x' of undefined",
  "stack": "at handler (api/main.js:42)\n  at ..."
}
```

### Snapshot Model

```sql
CREATE TABLE app_snapshots (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL,
    name TEXT NOT NULL,
    storage_data TEXT,        -- JSON dump of app_kv entries
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(app_id, name)
);
```

### Example Agent Workflow

```bash
# 1. Fork production for testing
fazt app fork myapp --as myapp-test

# 2. Create snapshot before tests
curl -X POST http://myapp-test.local:8080/_fazt/snapshot \
  -d '{"name":"pre-test"}'

# 3. Run tests, modify data
curl -X POST http://myapp-test.local:8080/api/data -d '{"test":true}'

# 4. Check execution logs
curl http://myapp-test.local:8080/_fazt/logs?limit=1

# 5. If something failed, check errors
curl http://myapp-test.local:8080/_fazt/errors

# 6. Verify storage state
curl http://myapp-test.local:8080/_fazt/storage

# 7. Restore after tests
curl -X POST http://myapp-test.local:8080/_fazt/restore/pre-test

# 8. When ready, swap labels
fazt app swap myapp myapp-test

# 9. Clean up old version
fazt app delete app_OLD_ID
```

---

## Migration

### From Current Schema

Current:
```sql
id TEXT PRIMARY KEY,      -- "othelo" (same as name)
name TEXT NOT NULL UNIQUE -- "othelo"
```

Migration steps:

1. Generate UUIDs for existing apps
2. Rename `name` to `label`
3. Add lineage fields (NULL for existing apps)
4. Update storage references

```sql
-- Step 1: Add new columns
ALTER TABLE apps ADD COLUMN new_id TEXT;
ALTER TABLE apps ADD COLUMN original_id TEXT;
ALTER TABLE apps ADD COLUMN forked_from_id TEXT;

-- Step 2: Generate UUIDs and set as original
UPDATE apps SET
  new_id = 'app_' || lower(hex(randomblob(4))),
  original_id = 'app_' || lower(hex(randomblob(4)));

-- Step 3: Update original_id to match new_id (they're originals)
UPDATE apps SET original_id = new_id;

-- Step 4: Update storage references
UPDATE app_kv SET app_id = (
  SELECT new_id FROM apps WHERE apps.id = app_kv.app_id
);

-- Step 5: Rename columns
ALTER TABLE apps RENAME COLUMN name TO label;
ALTER TABLE apps RENAME COLUMN id TO old_id;
ALTER TABLE apps RENAME COLUMN new_id TO id;

-- Step 6: Drop old column, update primary key
-- (requires table rebuild in SQLite)
```

---

## System Apps

System apps (admin, 404, root) have reserved labels:

| Label | Purpose |
|-------|---------|
| `admin` | Admin dashboard |
| `404` | Error page |
| `root` | Root domain handler |

These cannot be reassigned to user apps.

---

## URL Routing

Request routing with new model:

```
Request: https://myapp.zyt.app/api/hello

1. Extract subdomain: "myapp"
2. Query: SELECT id FROM apps WHERE label = 'myapp'
3. If found: Route to app with that ID
4. If not found: Check if subdomain matches app_* pattern
   - If yes: Route to app by ID directly
   - If no: Return 404
```

This allows both:
- `myapp.zyt.app` (via label)
- `app_7f3k9x2m.zyt.app` (via ID directly)

---

## Open Questions

1. **Label history**: Track when labels were assigned/removed?
2. **Multiple labels**: Allow multiple labels per app (aliases)?
3. **Label reservations**: Reserve labels before app is ready?
4. **Cross-node forks**: Fork from remote to local?
5. **Storage snapshots**: How many to keep? Auto-cleanup?

---

## Implementation Order

1. **Schema migration**: Add UUID, rename name→label, add lineage
2. **ID generation**: Implement nanoid/UUID generation
3. **Label commands**: `fazt app label/unlabel/swap`
4. **Fork command**: `fazt app fork` with storage copy
5. **Lineage commands**: `fazt app lineage/family`
6. **Agent endpoints**: `/_fazt/*` for testing workflows
7. **Update routing**: Support both label and ID-based routing

---

## Related Specs

- `v0.10-runtime/` - Runtime enhancements (async, stdlib)
- `v0.9-storage/` - Storage layer (KV, docs, blobs)
