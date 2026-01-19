# App Identity, Aliases, and Lineage

**Version**: v0.10
**Status**: Draft
**Created**: 2026-01-19
**Updated**: 2026-01-19

## Overview

This spec separates app identity (content + metadata) from routing (aliases).
Apps have permanent UUIDs and metadata. Aliases handle all subdomain routing,
enabling flexible workflows like traffic splitting, reserved subdomains, and
atomic swaps.

## Goals

1. **Stable Identity**: Apps have permanent UUIDs that never change
2. **Flexible Routing**: Aliases handle subdomains, can be retargeted freely
3. **Lineage Tracking**: Know where an app came from and its history
4. **Agent-Friendly**: Support LLM agent workflows (test, fork, promote)
5. **Organizational Metadata**: Title, description, tags persist across forks

## Non-Goals

- Multi-tenant apps (single owner per fazt instance)
- Complex permission models
- Git-like branching (simpler fork/promote model)

---

## Data Model

### Apps Table (Content + Identity)

```sql
CREATE TABLE apps (
    -- Identity (immutable)
    id TEXT PRIMARY KEY,              -- "app_7f3k9x2m" (UUID, never changes)

    -- Lineage
    original_id TEXT,                 -- Root ancestor (self if original)
    forked_from_id TEXT,              -- Immediate parent (NULL if original)

    -- Metadata (inherited on fork)
    title TEXT,                       -- "Tetris" - what it is
    description TEXT,                 -- "Classic block-stacking game"
    tags TEXT,                        -- JSON array: ["game", "arcade"]
    visibility TEXT DEFAULT 'unlisted', -- public|unlisted|private

    -- Source tracking
    source TEXT DEFAULT 'deploy',     -- 'deploy', 'git', 'fork', 'system'
    source_url TEXT,
    source_ref TEXT,
    source_commit TEXT,

    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_apps_original ON apps(original_id);
CREATE INDEX idx_apps_forked_from ON apps(forked_from_id);
CREATE INDEX idx_apps_visibility ON apps(visibility);
```

### Aliases Table (Routing)

```sql
CREATE TABLE aliases (
    -- Routing key
    subdomain TEXT PRIMARY KEY,       -- "tetris" → tetris.zyt.app

    -- Routing behavior
    type TEXT DEFAULT 'proxy',        -- proxy|redirect|reserved|split

    -- Target(s)
    targets TEXT,                     -- JSON (structure depends on type)

    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Alias Types

| Type | Targets Format | Behavior |
|------|----------------|----------|
| `proxy` | `{"app_id": "app_abc"}` | Serve app content, URL unchanged |
| `redirect` | `{"url": "https://..."}` | 301/302 redirect |
| `reserved` | `null` | Returns 404 (blocks subdomain) |
| `split` | `[{"app_id": "x", "weight": 50}, ...]` | Weighted + sticky session |

### Key Concepts

| Field | Table | Purpose | Mutability |
|-------|-------|---------|------------|
| `id` | apps | Permanent identity | Immutable |
| `title` | apps | Human-readable name | Mutable |
| `subdomain` | aliases | Routing key | Mutable (retargetable) |
| `visibility` | apps | Discoverability | Mutable |

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

## Visibility Model

Controls discoverability, not access:

| Visibility | Direct URL | Public API | Admin API |
|------------|------------|------------|-----------|
| `public` | Yes | Yes | Yes |
| `unlisted` | Yes | No | Yes |
| `private` | Auth only | No | Yes |

- **public**: Listed on homepage, accessible to all
- **unlisted**: Accessible via direct URL, not discoverable
- **private**: Requires authentication (future)

---

## Lineage Model

### Example Lineage Tree

```
app_001 (original)
│   title: "Tetris"
│   original_id: app_001
│
├── app_002
│   │   title: "Tetris" (inherited)
│   │   forked_from: app_001
│   │   original: app_001
│   │
│   └── app_003
│           title: "Tetris" (inherited)
│           forked_from: app_002
│           original: app_001
│
└── app_004
        title: "Tetris" (inherited)
        forked_from: app_001
        original: app_001
```

Aliases (separate):
```
tetris       → app_001 (production)
tetris-staging → app_002
tetris-v2    → app_003
(app_004 has no alias - accessible by ID only)
```

### Lineage Rules

1. **Original apps**: `original_id = id`, `forked_from_id = NULL`
2. **Forked apps**: `original_id` = root ancestor, `forked_from_id` = parent
3. **Metadata inheritance**: title, description, tags copied on fork
4. **Deletion**: Deleting an app doesn't affect descendants' lineage fields
5. **Orphans OK**: `forked_from_id` can reference deleted apps

---

## Alias Operations

Aliases are the routing layer. Multiple aliases can point to same app.

### Basic Operations

```bash
# Create alias pointing to app
fazt app link tetris --id app_abc123 --to zyt
# → tetris.zyt.app now serves app_abc123

# Retarget alias to different app
fazt app link tetris --id app_def456 --to zyt
# → tetris.zyt.app now serves app_def456

# Remove alias
fazt app unlink tetris --from zyt
# → tetris.zyt.app returns 404

# Reserve subdomain (block it)
fazt app reserve admin --on zyt
# → admin.zyt.app returns 404, can't be used
```

### Multiple Aliases Per App

```bash
fazt app link tetris --id app_abc123 --to zyt
fazt app link games-tetris --id app_abc123 --to zyt
fazt app link classic-blocks --id app_abc123 --to zyt
# → All three subdomains serve the same app
```

### Traffic Splitting

```bash
fazt app split tetris --ids app_v1:50,app_v2:50 --on zyt
# → 50% traffic to each version, sticky sessions
```

Traffic splitting without URL change:
- Server picks target based on weights
- Sets cookie for session affinity (same user → same version)
- Browser URL stays `tetris.zyt.app`

---

## CLI Commands

### Reference by --alias or --id

Most commands accept either flag:

```bash
# Info
fazt app info --alias tetris          # resolves alias → shows app
fazt app info --id app_abc123         # direct lookup

# Deploy
fazt app deploy ./dir --alias tetris --to zyt
fazt app deploy ./dir --id app_abc123 --to zyt

# Remove
fazt app remove --alias tetris --from zyt   # removes alias only
fazt app remove --id app_abc123 --from zyt  # removes app + orphans aliases
```

### List

```bash
# List all apps
fazt app list zyt
# ID            TITLE     VISIBILITY  TAGS          FORKED-FROM
# app_7f3k9x2m  Tetris    public      game,arcade   -
# app_m3n8k1vz  Tetris    unlisted    game,arcade   app_7f3k9x2m
# app_x9p2q4wn  Notes     public      tool          -

# List aliases
fazt app list --aliases zyt
# SUBDOMAIN      TYPE    TARGET
# tetris         proxy   app_7f3k9x2m
# tetris-dev     proxy   app_m3n8k1vz
# notes          proxy   app_x9p2q4wn
# admin          reserved -
```

### Info

```bash
fazt app info --alias tetris --on zyt
# Alias:       tetris
# Type:        proxy
# App ID:      app_7f3k9x2m
# Title:       Tetris
# Description: Classic block-stacking game
# Tags:        game, arcade
# Visibility:  public
# URL:         https://tetris.zyt.app
# Original:    app_7f3k9x2m (self)
# Forked from: -
# Source:      deploy
```

### Deploy

```bash
# Deploy creates app + alias from manifest
fazt app deploy ./src --alias tetris --to zyt
# → Created app_x9p2q4wn
# → Created alias tetris → app_x9p2q4wn
# → URL: https://tetris.zyt.app

# Deploy to existing alias (updates app behind it)
fazt app deploy ./src --alias tetris --to zyt
# → Updated app_x9p2q4wn

# Deploy to app directly (no alias)
fazt app deploy ./src --id app_x9p2q4wn --to zyt
```

### Fork

```bash
# Fork with new alias
fazt app fork --alias tetris --as tetris-staging --to zyt
# → Created app_m3n8k1vz (forked from app_7f3k9x2m)
# → Created alias tetris-staging → app_m3n8k1vz

# Fork without alias (unlisted)
fazt app fork --id app_7f3k9x2m --to zyt
# → Created app_p4q7r2st (no alias, accessible by ID)

# Fork without copying storage
fazt app fork --alias tetris --as tetris-clean --no-storage --to zyt
```

### Link/Unlink (Alias Management)

```bash
# Create or update alias
fazt app link tetris --id app_abc123 --to zyt

# Remove alias
fazt app unlink tetris --from zyt

# Reserve subdomain
fazt app reserve admin --on zyt

# Traffic split
fazt app split tetris --ids app_v1:50,app_v2:50 --on zyt

# Atomic swap (exchange two aliases' targets)
fazt app swap tetris tetris-v2 --on zyt
```

### Lineage

```bash
# Show lineage tree
fazt app lineage --id app_7f3k9x2m --on zyt
# app_7f3k9x2m "Tetris" (original)
# ├── app_m3n8k1vz "Tetris" [tetris-staging]
# │   └── app_q2w3e4r5 "Tetris" [tetris-v2]
# └── app_p4q7r2st "Tetris" (no alias)

# Delete app and all forks
fazt app remove --id app_7f3k9x2m --with-forks --from zyt
```

---

## Manifest Extension

`manifest.json` now supports metadata:

```json
{
  "name": "tetris",
  "title": "Tetris",
  "description": "Classic block-stacking game",
  "tags": ["game", "arcade"],
  "visibility": "public"
}
```

On deploy:
- `name` → used as alias subdomain (if --alias not specified)
- `title`, `description`, `tags`, `visibility` → stored on app

---

## Logging and Analytics

All logs use `app_id`, not subdomain:

```sql
-- Events table
CREATE TABLE events (
    ...
    app_id TEXT,    -- app_abc123 (stable across renames)
    domain TEXT,    -- tetris.zyt.app (for reference)
    ...
);

-- Audit logs
CREATE TABLE audit_logs (
    ...
    app_id TEXT,    -- app_abc123
    ...
);

-- Deployments
CREATE TABLE deployments (
    ...
    app_id TEXT,    -- app_abc123
    ...
);
```

This ensures historical data stays connected even after alias changes.

---

## URL Routing

Request routing with aliases:

```
Request: https://tetris.zyt.app/api/hello

1. Extract subdomain: "tetris"
2. Query aliases: SELECT type, targets FROM aliases WHERE subdomain = 'tetris'
3. If found:
   - proxy: Serve content from target app_id
   - redirect: 301/302 to target URL
   - reserved: Return 404
   - split: Pick target by weight, set sticky cookie, serve
4. If not found:
   - Check if subdomain matches app_* pattern
   - If yes: Serve app by ID directly
   - If no: Return 404
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

### Example Agent Workflow

```bash
# 1. Fork production for testing
fazt app fork --alias tetris --as tetris-test --to local

# 2. Create snapshot before tests
curl -X POST http://tetris-test.local:8080/_fazt/snapshot \
  -d '{"name":"pre-test"}'

# 3. Run tests
curl -X POST http://tetris-test.local:8080/api/action

# 4. Check logs
curl http://tetris-test.local:8080/_fazt/logs?limit=1

# 5. Check errors if needed
curl http://tetris-test.local:8080/_fazt/errors

# 6. Verify storage
curl http://tetris-test.local:8080/_fazt/storage

# 7. Restore after tests
curl -X POST http://tetris-test.local:8080/_fazt/restore/pre-test

# 8. When ready, swap to production
fazt app swap tetris tetris-test --on zyt

# 9. Clean up old version
fazt app remove --id app_OLD_ID --from zyt
```

---

## Migration

### From Current Schema

Current:
```sql
-- apps table uses name as both id and subdomain
id TEXT PRIMARY KEY,      -- "othelo"
name TEXT NOT NULL UNIQUE -- "othelo"
```

Migration steps:

1. Create new `apps` table with UUID schema
2. Create `aliases` table
3. Migrate existing apps:
   - Generate UUID for each app
   - Copy name to alias pointing to new UUID
   - Set original_id = id (they're originals)
4. Update VFS to key by app_id instead of subdomain
5. Update all logs/analytics to reference app_id

```sql
-- Create new tables
CREATE TABLE apps_new (...);
CREATE TABLE aliases (...);

-- Migrate apps
INSERT INTO apps_new (id, original_id, title, source, created_at, updated_at)
SELECT
  'app_' || lower(hex(randomblob(4))),
  'app_' || lower(hex(randomblob(4))),
  name,
  source,
  created_at,
  updated_at
FROM apps;

-- Fix original_id to match id
UPDATE apps_new SET original_id = id;

-- Create aliases for each app
INSERT INTO aliases (subdomain, type, targets)
SELECT
  old.name,
  'proxy',
  json_object('app_id', new.id)
FROM apps old
JOIN apps_new new ON old.name = new.title;

-- Update VFS references
UPDATE vfs_files SET app_id = (
  SELECT a.id FROM apps_new a
  JOIN aliases al ON json_extract(al.targets, '$.app_id') = a.id
  WHERE al.subdomain = vfs_files.site_id
);
```

---

## Reserved Subdomains

System subdomains that cannot be used by user apps:

| Subdomain | Purpose |
|-----------|---------|
| `admin` | Admin dashboard |
| `api` | API endpoints |
| `404` | Error page |
| `root` | Root domain handler |

These are created as `reserved` type aliases on server init.

---

## Remote Execution (`@peer`)

Any command can be executed on a remote fazt node using the `@peer` prefix:

```bash
# Local execution
fazt app list

# Remote execution (equivalent)
fazt @zyt app list
fazt @local app deploy ./dir --alias tetris
```

### How It Works

```
┌─────────────────┐         ┌─────────────────┐
│  Local CLI      │         │  Remote fazt    │
│                 │         │                 │
│ fazt @zyt app   │───API──→│ POST /api/cmd   │
│      list       │         │ {cmd: "app",    │
│                 │←──JSON──│  args: ["list"]}│
└─────────────────┘         └─────────────────┘
```

1. CLI detects `@peer` prefix
2. Looks up peer URL and token from local config
3. Serializes command to API request
4. Remote fazt executes, returns JSON response
5. Local CLI formats and displays result

### Syntax

```bash
fazt @<peer> <command> [args...]

# Examples
fazt @zyt app list
fazt @zyt app info --alias tetris
fazt @zyt app deploy ./dir --alias myapp
fazt @local app fork --alias tetris --as tetris-dev
```

### Authentication

Uses existing peer tokens stored in local client DB:

```bash
# Peer already configured with token
fazt remote add zyt --url https://admin.zyt.app --token <TOKEN>

# Now @zyt works automatically
fazt @zyt app list
```

### Commands Available Remotely

| Command | Remote API | Notes |
|---------|------------|-------|
| `app list` | Yes | List apps/aliases |
| `app info` | Yes | App details |
| `app deploy` | Yes | Upload and deploy |
| `app remove` | Yes | Delete app/alias |
| `app fork` | Yes | Fork with metadata |
| `app link` | Yes | Create/update alias |
| `app unlink` | Yes | Remove alias |
| `app reserve` | Yes | Block subdomain |
| `app split` | Yes | Traffic splitting |
| `app swap` | Yes | Atomic swap |
| `app lineage` | Yes | Show fork tree |
| `server info` | Yes | Server metadata |
| `server stop` | Yes | Graceful shutdown |
| `server create-key` | Yes | Create API key |
| `server start` | No | Local only |
| `server init` | No | Local only |
| `remote *` | No | Manages local peers |

### Error Handling

```bash
fazt @zyt app list
# Error: peer 'zyt' not found (check 'fazt remote list')

fazt @zyt app list
# Error: connection refused (is the server running?)

fazt @zyt app list
# Error: 401 unauthorized (check API token)
```

---

## API Endpoints

Every CLI command has a corresponding API endpoint. All admin endpoints require
authentication via API token in `Authorization: Bearer <token>` header.

### Public (no auth)

```
GET /api/apps                    → visibility=public apps only
GET /api/health                  → Health check
```

### Apps (authenticated)

```
GET    /api/apps?all=true        → All apps (admin)
GET    /api/apps/:id             → App by ID
POST   /api/apps                 → Create app
PUT    /api/apps/:id             → Update app metadata
DELETE /api/apps/:id             → Delete app
DELETE /api/apps/:id?with-forks  → Delete app and all forks
```

### Aliases (authenticated)

```
GET    /api/aliases              → All aliases
GET    /api/aliases/:subdomain   → Alias details
POST   /api/aliases              → Create alias (link)
PUT    /api/aliases/:subdomain   → Update alias (retarget)
DELETE /api/aliases/:subdomain   → Delete alias (unlink)
```

### Alias Operations (authenticated)

```
POST /api/aliases/:subdomain/reserve
  → Reserve subdomain (block it)

POST /api/aliases/swap
  Body: {"alias1": "tetris", "alias2": "tetris-v2"}
  → Atomic swap of two aliases' targets

POST /api/aliases/:subdomain/split
  Body: {"targets": [{"app_id": "app_x", "weight": 50}, ...]}
  → Configure traffic splitting
```

### Fork (authenticated)

```
POST /api/apps/:id/fork
  Body: {"alias": "tetris-dev", "copy_storage": true}
  → Fork app with optional new alias
```

### Lineage (authenticated)

```
GET /api/apps/:id/lineage        → Lineage tree for app
GET /api/apps/:id/forks          → Direct forks of app
```

### Deploy (authenticated)

```
POST /api/deploy
  Content-Type: multipart/form-data
  Body: zip file + metadata
  → Deploy app (creates app + alias if needed)

POST /api/deploy?alias=tetris
  → Deploy to specific alias

POST /api/deploy?id=app_abc123
  → Deploy to specific app ID
```

### Server (authenticated)

```
GET  /api/server/info            → Server metadata, version
POST /api/server/stop            → Graceful shutdown
POST /api/server/keys            → Create new API key
GET  /api/server/keys            → List API keys
DELETE /api/server/keys/:id      → Revoke API key
```

### Command Gateway (for @peer)

Generic endpoint that accepts any CLI command:

```
POST /api/cmd
  Body: {"command": "app", "args": ["list", "--aliases"]}
  Response: {"success": true, "output": "...", "data": {...}}
```

This enables `@peer` to work with any command without needing
explicit endpoint mapping in the client.

---

## Implementation Order

### Phase 1: Core Data Model
1. **Schema migration**: Create apps/aliases tables, migrate data
2. **ID generation**: Implement nanoid generation
3. **Routing update**: Check aliases table first

### Phase 2: CLI + API (1:1)
4. **API endpoints**: All app/alias CRUD operations
5. **CLI --alias/--id flags**: Update all app commands
6. **Link/unlink/reserve**: Alias management commands + APIs
7. **Fork with metadata**: Copy title/description/tags

### Phase 3: Remote Execution
8. **Command gateway**: `POST /api/cmd` endpoint
9. **@peer parsing**: CLI detects and routes to remote
10. **Error handling**: Network, auth, command errors

### Phase 4: Advanced Features
11. **Visibility filter**: Public API only returns public apps
12. **Traffic splitting**: Weighted routing + sticky sessions
13. **Agent endpoints**: `/_fazt/*` for testing workflows

---

## Related Specs

- `v0.10-runtime/` - Runtime enhancements (async, stdlib)
- `v0.9-storage/` - Storage layer (KV, docs, blobs)
