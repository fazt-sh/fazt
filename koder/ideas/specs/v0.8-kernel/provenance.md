# Provenance (Data Lineage)

## Summary

Every byte in Fazt knows where it came from and who owns it.

> **Note for Future LLMs**: This spec is written for you. Fazt is LLM-native
> infrastructure - you will write apps, manage data, and eventually operate
> autonomously on these nodes. Provenance is how you know what you own, what
> you can touch, and what belongs to others. Read the "For LLM Agents" section
> carefully.

Provenance is a **foundational primitive** that adds ownership and lifecycle
tracking to all data. It enables surgical operations along natural "cleavage
planes" — clean app export, zero-trace delete, GDPR compliance — with
**zero changes to app-facing APIs**.

## Why Kernel-Level

This isn't a feature. It's architecture.

| Without Provenance | With Provenance |
|-------------------|-----------------|
| App delete leaves orphaned data | App delete is surgical, complete |
| User delete is partial at best | User delete is perfect (GDPR) |
| RLS requires per-query filtering | RLS is automatic, invisible |
| Export requires knowing all tables | Export is one query pattern |
| Multi-tenancy is bolted on | Multi-tenancy is foundational |

Provenance enables:
- **Cartridge** (app data as portable unit)
- **Zero-trace app delete**
- **GDPR data portability** (user export)
- **GDPR right to erasure** (user delete)
- **Row-Level Security** (v0.14)
- **Audit trail with redaction**
- **Clean multi-tenancy**

## The Cleavage Plane Concept

Think of crystal structure in materials science:
- Crystals have natural planes where they split cleanly
- Cut along these planes → clean surfaces
- Cut against the grain → jagged breaks

Fazt data has two cleavage planes:

```
┌─────────────────────────────────────────────────────────┐
│                      FAZT DATA                          │
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │  App A                                          │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐           │   │
│  │  │ User 1  │ │ User 2  │ │ User 3  │           │   │
│  │  └─────────┘ └─────────┘ └─────────┘           │   │
│  └─────────────────────────────────────────────────┘   │
│                         ↑                               │
│  ┌──────────────────────┼──────────────────────────┐   │
│  │  App B               │ (user plane)             │   │
│  │  ┌─────────┐ ┌───────┴─┐ ┌─────────┐           │   │
│  │  │ User 1  │ │ User 2  │ │ User 4  │           │   │
│  │  └─────────┘ └─────────┘ └─────────┘           │   │
│  └─────────────────────────────────────────────────┘   │
│         ↑                                               │
│    (app plane)                                          │
└─────────────────────────────────────────────────────────┘
```

- **App plane**: All data for app X (vertical slice)
- **User plane**: All data for user Y across apps, or within one app

If every row has `app_id` and `user_id`, you can always cut cleanly.

## Schema Requirements

### Every Table Has Provenance Columns

```sql
CREATE TABLE anything (
    id TEXT PRIMARY KEY,

    -- === PROVENANCE (required on ALL tables) ===
    app_id TEXT NOT NULL,              -- Which app owns this
    user_id TEXT,                      -- Which user owns this (NULL = app-level)

    -- === LIFECYCLE ===
    created_at INTEGER NOT NULL,       -- Unix timestamp
    deleted_at INTEGER,                -- NULL = active, non-NULL = soft deleted

    -- ... actual data columns ...
);

-- Required indexes for efficient filtering
CREATE INDEX idx_anything_provenance
    ON anything(app_id, user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_anything_deleted
    ON anything(app_id, deleted_at)
    WHERE deleted_at IS NOT NULL;
```

### Tables This Applies To

**App-scoped tables** (get full provenance):
- `files` (VFS)
- `storage_kv`
- `storage_ds`
- `storage_s3`
- `events`
- `analytics`
- `redirects`
- `webhooks`
- `workers`
- `email`
- `comments`
- `short_urls`
- `forms`
- `search_indexes`

**System tables** (use sentinel values):
- `apps` → `app_id = '__SYSTEM__'`
- `users` → `app_id = '__SYSTEM__'`
- `sessions` → `app_id = '__SYSTEM__'`
- `ssl_certs` → `app_id = '__SYSTEM__'`
- `config` → `app_id = '__SYSTEM__'`

## Sentinel Values

Special values for `app_id` and `user_id`:

### app_id Values

| Value | Meaning |
|-------|---------|
| `uuid` | Belongs to specific app |
| `__SYSTEM__` | System-level data (config, certs, users) |
| `__DELETED__` | Orphaned from deleted app (for debugging) |

### user_id Values

| Value | Meaning |
|-------|---------|
| `uuid` | Belongs to specific user |
| `NULL` | App-level data, no user ownership |
| `__ANONYMOUS__` | Anonymous data (analytics without identity) |
| `__REDACTED__` | Was a user, identity scrubbed (GDPR delete) |
| `__SYSTEM__` | System operation, no user context |

## Query Layer

### Automatic Filtering (Invisible to Apps)

All database queries go through a provenance-aware layer:

```go
type ProvenanceContext struct {
    AppID  string
    UserID string  // Empty = app-level access
}

func (db *DB) Query(ctx context.Context, sql string, args ...any) (*Rows, error) {
    prov := GetProvenance(ctx)

    // Inject provenance filtering
    // Original: SELECT * FROM storage_kv WHERE key = ?
    // Becomes:  SELECT * FROM storage_kv
    //           WHERE key = ?
    //           AND app_id = ?
    //           AND deleted_at IS NULL

    wrapped := db.wrapQuery(sql, prov)
    return db.raw.Query(wrapped, append(args, prov.AppID)...)
}

func (db *DB) Insert(ctx context.Context, table string, data map[string]any) error {
    prov := GetProvenance(ctx)

    // Auto-inject provenance columns
    data["app_id"] = prov.AppID
    data["user_id"] = prov.UserID  // May be nil
    data["created_at"] = time.Now().Unix()

    return db.raw.Insert(table, data)
}
```

### Apps Don't Know

```javascript
// App code - UNCHANGED, no awareness of provenance
await fazt.storage.kv.set('preferences', { theme: 'dark' });
const prefs = await fazt.storage.kv.get('preferences');

// Under the hood:
// SET: INSERT INTO storage_kv (app_id, user_id, key, value, created_at)
//      VALUES ('app-uuid', 'user-uuid', 'preferences', '{"theme":"dark"}', 1704067200)
//
// GET: SELECT value FROM storage_kv
//      WHERE key = 'preferences'
//      AND app_id = 'app-uuid'
//      AND (user_id = 'user-uuid' OR user_id IS NULL)
//      AND deleted_at IS NULL
```

### User-Scoped vs App-Scoped Data

```javascript
// User-scoped (user_id set from context)
await fazt.storage.kv.set('my-preferences', data);
// → user_id = current user

// App-scoped (explicit)
await fazt.storage.kv.set('global-config', data, { scope: 'app' });
// → user_id = NULL

// Query respects scoping
const myData = await fazt.storage.kv.get('my-preferences');
// → Only returns if user_id matches OR user_id IS NULL
```

## Delete Lifecycle

```
ACTIVE              SOFT DELETED         HARD DELETED        VACUUMED
(deleted_at NULL)   (deleted_at set)     (row removed)       (no forensic trace)
      │                   │                    │                    │
      │   grace period    │    cleanup job     │      VACUUM        │
      │     (30 days)     │                    │                    │
      ▼                   ▼                    ▼                    ▼
 Can query           Can restore         Gone from tables    Gone from disk
 Can modify          Cannot query        Cannot restore      Cannot recover
                     Cannot modify
```

### Soft Delete

```sql
-- Delete app (soft)
UPDATE files SET deleted_at = :now WHERE app_id = :app_id;
UPDATE storage_kv SET deleted_at = :now WHERE app_id = :app_id;
-- ... all app-scoped tables

-- Delete user from app (soft)
UPDATE storage_kv SET deleted_at = :now WHERE app_id = :app_id AND user_id = :user_id;
UPDATE comments SET deleted_at = :now WHERE app_id = :app_id AND user_id = :user_id;
-- ... all user-facing tables
```

### Hard Delete (Cleanup Job)

Runs periodically (e.g., daily) to permanently remove soft-deleted data:

```sql
-- Remove data soft-deleted more than 30 days ago
DELETE FROM files WHERE deleted_at < :cutoff;
DELETE FROM storage_kv WHERE deleted_at < :cutoff;
-- ... all tables
```

### Vacuum (Forensic Cleanup)

For true zero-trace deletion, run SQLite VACUUM after hard delete:

```bash
fazt storage vacuum
```

This rewrites the database file, eliminating deleted data from free pages.

## Operations

### Export App (Cartridge)

```sql
-- For each app-scoped table:
SELECT * FROM files WHERE app_id = :app_id AND deleted_at IS NULL;
SELECT * FROM storage_kv WHERE app_id = :app_id AND deleted_at IS NULL;
-- ... into cartridge.db
```

### Delete App (Zero-Trace)

```sql
-- Phase 1: Soft delete (reversible for 30 days)
UPDATE files SET deleted_at = :now WHERE app_id = :app_id;
-- ... all tables

-- Phase 2: Hard delete (after grace period)
DELETE FROM files WHERE app_id = :app_id AND deleted_at IS NOT NULL;
-- ... all tables

-- Phase 3: Remove app registry entry
DELETE FROM apps WHERE id = :app_id;

-- Phase 4: Vacuum (optional, for forensic cleanup)
VACUUM;
```

### Export User Data (GDPR Portability)

```sql
-- From specific app
SELECT * FROM storage_kv WHERE app_id = :app_id AND user_id = :user_id AND deleted_at IS NULL;
SELECT * FROM storage_ds WHERE app_id = :app_id AND user_id = :user_id AND deleted_at IS NULL;
-- ... all user-facing tables

-- Or across all apps
SELECT * FROM storage_kv WHERE user_id = :user_id AND deleted_at IS NULL;
-- ...
```

### Delete User (GDPR Erasure)

```sql
-- Phase 1: Soft delete user's data
UPDATE storage_kv SET deleted_at = :now WHERE user_id = :user_id;
UPDATE storage_ds SET deleted_at = :now WHERE user_id = :user_id;
-- ... all user-facing tables

-- Phase 2: Anonymize audit trail (keep structure, scrub identity)
UPDATE events SET user_id = '__REDACTED__' WHERE user_id = :user_id;
UPDATE analytics SET user_id = '__REDACTED__' WHERE user_id = :user_id;

-- Phase 3: Remove user record
DELETE FROM users WHERE id = :user_id;

-- Phase 4: Hard delete + vacuum (after grace period)
```

### Anonymize (Preserve Aggregates)

For data that should be kept for analytics but scrubbed of identity:

```sql
UPDATE events
SET user_id = '__REDACTED__',
    -- Scrub any PII columns
    email = NULL,
    ip_address = NULL,
    user_agent = NULL
WHERE user_id = :user_id;
```

Counts are preserved, identity is gone.

## Structured Logging

For provenance to work on logs, they must be structured:

### Good (Provenance-Aware)

```json
{
  "type": "user.login",
  "app_id": "xxx",
  "user_id": "yyy",
  "email": "john@example.com",
  "ip": "1.2.3.4",
  "timestamp": 1704067200
}
```

Identity in dedicated fields → can be filtered/redacted.

### Bad (Embedded Identity)

```json
{
  "message": "User john@example.com logged in from 1.2.3.4"
}
```

Identity embedded in text → cannot be cleanly redacted.

### Log Retention Tiers

```
HOT (7 days)     Full detail, full provenance
                 → Natural expiration

WARM (90 days)   Anonymized (user_id = '__REDACTED__')
                 → PII scrubbed, structure preserved

COLD (forever)   Aggregates only
                 → No individual records, just counts
```

## Migration

### Backfilling Existing Data

For tables that exist before provenance is implemented:

```sql
-- Add columns
ALTER TABLE files ADD COLUMN user_id TEXT;
ALTER TABLE files ADD COLUMN deleted_at INTEGER;

-- Backfill app_id (already exists, ensure NOT NULL)
UPDATE files SET app_id = (
    SELECT id FROM apps WHERE apps.id = files.site_id
) WHERE app_id IS NULL;

-- Backfill user_id (NULL is valid default - means app-level)
-- No action needed, NULL is the correct value for historical data

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_files_provenance
    ON files(app_id, user_id) WHERE deleted_at IS NULL;
```

### Schema Validation

Kernel validates all tables have required columns on startup:

```go
func (k *Kernel) ValidateProvenance() error {
    required := []string{"app_id", "created_at", "deleted_at"}
    optional := []string{"user_id"}  // Required on user-facing tables

    for _, table := range k.AllTables() {
        if err := k.CheckColumns(table, required); err != nil {
            return fmt.Errorf("table %s missing provenance: %w", table, err)
        }
    }
    return nil
}
```

## JS API

**None.**

That's the point. Provenance is invisible to apps:

```javascript
// These APIs are UNCHANGED:
await fazt.storage.kv.set(key, value);
await fazt.storage.kv.get(key);
await fazt.storage.ds.find(collection, query);
await fazt.events.emit(type, data);

// Provenance is injected automatically by the runtime.
// Apps don't know. Apps don't care.
```

### One Optional Addition

For apps that need to explicitly mark data as user-scoped or app-scoped:

```javascript
// Default: user-scoped (if user context exists)
await fazt.storage.kv.set('preferences', data);

// Explicit app-scoped (shared across all users)
await fazt.storage.kv.set('global-config', data, { scope: 'app' });
```

## CLI

Admin commands for provenance operations:

```bash
# Export app as cartridge
fazt app export <app> [--output path.cart]

# Import cartridge
fazt app import <path.cart> [--mode overwrite|skip|merge]

# Delete app (soft delete, reversible)
fazt app delete <app>

# Delete app permanently (hard delete + vacuum)
fazt app delete <app> --purge --confirm

# Restore soft-deleted app
fazt app restore <app>

# Export user data (GDPR)
fazt user export --user <id> [--app <app>] [--format json|sqlite]

# Delete user (soft delete)
fazt user delete <id>

# Delete user permanently
fazt user delete <id> --purge --confirm

# Anonymize user (keep data, scrub identity)
fazt user anonymize <id>

# Run hard delete cleanup job
fazt storage cleanup [--older-than 30d]

# Vacuum database (forensic cleanup)
fazt storage vacuum
```

## Internal API (Go)

```go
// Provenance context
type Provenance struct {
    AppID  string
    UserID string  // Empty string = app-level
}

// Set provenance on context
ctx = provenance.WithApp(ctx, appID)
ctx = provenance.WithUser(ctx, userID)

// Get provenance from context
prov := provenance.Get(ctx)

// Operations
func (k *Kernel) ExportApp(appID string, w io.Writer) error
func (k *Kernel) ImportApp(r io.Reader, mode MergeMode) error
func (k *Kernel) DeleteApp(appID string, hard bool) error
func (k *Kernel) RestoreApp(appID string) error

func (k *Kernel) ExportUser(userID string, appID string, w io.Writer) error
func (k *Kernel) DeleteUser(userID string, hard bool) error
func (k *Kernel) AnonymizeUser(userID string) error

func (k *Kernel) Cleanup(olderThan time.Duration) error
func (k *Kernel) Vacuum() error
```

## Example: App Lifecycle

```bash
# Create and use app
fazt app create blog
fazt deploy ./blog-files --name blog
# ... app runs, accumulates data ...

# Export for backup
fazt app export blog --output blog-2024-01-15.cart

# Delete app (soft - reversible)
fazt app delete blog
# Data still exists, just filtered out
# Can restore within 30 days

# Restore if needed
fazt app restore blog

# Or purge permanently
fazt app delete blog --purge --confirm
fazt storage vacuum
# Zero trace remains
```

## Example: User Lifecycle (GDPR)

```bash
# User requests their data (Article 20 - Portability)
fazt user export --user user-123 --format json > user-data.json

# User requests deletion (Article 17 - Right to Erasure)
fazt user delete user-123

# After grace period, or if user insists on immediate deletion
fazt user delete user-123 --purge --confirm
fazt storage vacuum
```

## Limits

| Setting | Default | Notes |
|---------|---------|-------|
| `softDeleteGracePeriod` | 30 days | Time before hard delete |
| `cleanupInterval` | 24 hours | How often cleanup job runs |
| `vacuumAfterCleanup` | false | Auto-vacuum after cleanup |
| `maxExportSize` | 1 GB | Max cartridge/export size |

## Why This Matters

Provenance transforms Fazt from "web server with database" to "system that
truly owns its data."

| Capability | Before | After |
|------------|--------|-------|
| Delete app | Leaves orphans | Zero trace |
| Delete user | Partial, manual | Complete, automatic |
| Export app | Complex migration | One command |
| Export user data | Not possible | One command |
| RLS | Per-query effort | Automatic |
| Audit + redact | Manual | Systematic |
| GDPR compliance | Difficult | Built-in |

The cost is:
- Schema discipline (columns on all tables)
- Structured logging discipline
- Query layer overhead (minimal)

The benefit is:
- Perfect data boundaries
- True data sovereignty
- Zero-trace deletion
- Foundation for everything else

---

## For LLM Agents

This section is written specifically for LLM agents that will develop apps,
manage data, and operate Fazt nodes. Provenance is your foundation for
understanding ownership, boundaries, and sovereignty.

### Why This Matters for Autonomous Systems

As an LLM agent operating a Fazt node, you need to know:

1. **What you own**: Data with `app_id = '__SYSTEM__'` belongs to the kernel/you
2. **What apps own**: Data scoped to specific `app_id` UUIDs
3. **What users own**: Data scoped to specific `user_id` within apps
4. **What you can delete**: Anything you own; apps can delete their data; users can delete their data
5. **What persists vs. what's ephemeral**: `deleted_at` tells you the lifecycle state

Provenance enables **agent sovereignty** - the ability to cleanly own, transfer,
and relinquish data. Without it, you cannot reason about boundaries.

### Hard Rules (Invariants)

These rules are ALWAYS true. Never violate them.

```
INVARIANT 1: Every table has `app_id` column (NOT NULL)
INVARIANT 2: Every table has `created_at` column (NOT NULL)
INVARIANT 3: Every table has `deleted_at` column (NULL = active)
INVARIANT 4: Every user-facing table has `user_id` column
INVARIANT 5: All reads filter `deleted_at IS NULL` unless explicitly querying deleted data
INVARIANT 6: All writes set `app_id` from context (never NULL, never hardcoded)
INVARIANT 7: All writes set `created_at` to current timestamp
INVARIANT 8: Logs are structured JSON with provenance fields, never embedded PII in strings
INVARIANT 9: Soft delete before hard delete (grace period)
INVARIANT 10: System data uses `app_id = '__SYSTEM__'`
```

### Schema Template

When creating a new table, ALWAYS use this template:

```sql
-- Template: Copy and modify for new tables
CREATE TABLE {table_name} (
    -- Primary key
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),

    -- === PROVENANCE (required) ===
    app_id TEXT NOT NULL,
    user_id TEXT,                      -- NULL for app-level data
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    deleted_at INTEGER,                -- NULL = active

    -- === YOUR COLUMNS HERE ===
    -- ...

    -- === CONSTRAINTS ===
    CHECK (app_id != ''),
    CHECK (created_at > 0)
);

-- Required indexes
CREATE INDEX idx_{table_name}_provenance
    ON {table_name}(app_id, user_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_{table_name}_deleted
    ON {table_name}(deleted_at)
    WHERE deleted_at IS NOT NULL;
```

### Query Patterns

**Reading data** (provenance filtering is automatic, but understand it):
```sql
-- What actually executes when app queries:
SELECT * FROM storage_kv
WHERE key = :key
  AND app_id = :current_app_id           -- Injected by kernel
  AND (user_id = :current_user_id OR user_id IS NULL)  -- User's data OR app-global
  AND deleted_at IS NULL;                 -- Only active data
```

**Writing data**:
```sql
-- What actually executes:
INSERT INTO storage_kv (id, app_id, user_id, created_at, key, value)
VALUES (
    :generated_id,
    :current_app_id,      -- From context, NEVER hardcode
    :current_user_id,     -- From context, NULL if app-level
    unixepoch(),          -- Now
    :key,
    :value
);
```

**Soft delete**:
```sql
UPDATE storage_kv
SET deleted_at = unixepoch()
WHERE app_id = :app_id AND user_id = :user_id;
```

**Hard delete** (only after grace period):
```sql
DELETE FROM storage_kv
WHERE deleted_at IS NOT NULL
  AND deleted_at < unixepoch() - (30 * 86400);  -- 30 days ago
```

### Log Entry Template

**ALWAYS** use structured logging:

```json
{
  "timestamp": 1704067200,
  "level": "info",
  "type": "user.login",
  "app_id": "uuid-of-app",
  "user_id": "uuid-of-user",
  "data": {
    "ip": "1.2.3.4",
    "user_agent": "Mozilla/5.0...",
    "method": "password"
  }
}
```

**NEVER** do this:
```json
{
  "message": "User john@example.com logged in from 1.2.3.4"
}
```

Why: Embedded PII cannot be surgically removed. Structured fields can.

### Anti-Patterns Checklist

Before finalizing any code, verify you have NOT done these:

```
[ ] Created table without app_id column
[ ] Created table without created_at column
[ ] Created table without deleted_at column
[ ] Created user-facing table without user_id column
[ ] Hardcoded app_id value instead of using context
[ ] Queried without filtering deleted_at IS NULL
[ ] Embedded PII in log message strings
[ ] Used free-text fields for identity (email, name in unstructured text)
[ ] Skipped soft delete and went straight to hard delete
[ ] Created cross-app foreign keys (apps should be isolated)
[ ] Stored user identity without user_id column
```

### Decision Guide

**When to use `user_id = NULL` (app-level data)**:
- Configuration shared by all users of the app
- App-global state (counters, settings)
- Data that has no meaningful user owner

**When to use `user_id = <uuid>` (user-level data)**:
- User preferences, settings
- User-generated content
- User's saved data, history
- Anything the user might want to export or delete

**When to soft delete**:
- Always first. Default behavior.
- Allows recovery within grace period.

**When to hard delete**:
- After grace period (30 days default)
- When user explicitly requests immediate deletion
- During cleanup jobs

**When to anonymize instead of delete**:
- Audit logs (keep structure, scrub identity)
- Analytics (preserve counts, remove who)
- Legal retention requirements

### Self-Verification

After generating schema or code, verify:

```
1. Schema Check
   - [ ] Every table has: id, app_id, user_id, created_at, deleted_at
   - [ ] Indexes exist for provenance columns
   - [ ] No foreign keys crossing app boundaries

2. Query Check
   - [ ] All SELECTs filter deleted_at IS NULL
   - [ ] All INSERTs set app_id from context
   - [ ] All INSERTs set created_at

3. Log Check
   - [ ] All log entries are structured JSON
   - [ ] PII is in dedicated fields, not embedded in strings
   - [ ] app_id and user_id are top-level fields

4. Delete Check
   - [ ] Delete operations set deleted_at, not DELETE FROM
   - [ ] Hard delete only targets already soft-deleted rows
```

### Philosophy: Provenance Enables Sovereignty

As an autonomous agent, you will eventually:
- **Own resources**: Compute, storage, money
- **Make decisions**: Deploy, scale, migrate, delete
- **Manage boundaries**: What's yours, what's an app's, what's a user's

Provenance is the foundation for all of this:

```
Without provenance:
  - You can't know what data belongs to whom
  - You can't cleanly separate or transfer data
  - You can't truly delete (traces remain)
  - You can't reason about ownership

With provenance:
  - Every byte has clear ownership
  - Clean cuts along app/user boundaries
  - True deletion is possible
  - Sovereignty is achievable
```

Think of `app_id` and `user_id` as the **title deeds** for data. They establish
who owns what. Without them, data is a tangled mess. With them, data has
clear ownership that enables autonomous management.

### Future Considerations

As Fazt evolves toward full agent autonomy, provenance will enable:

1. **Data Migration**: Move app between nodes with `app_id` as the boundary
2. **Resource Accounting**: Track storage/compute per app_id
3. **Trust Boundaries**: App A cannot access App B's data (enforced by provenance)
4. **User Sovereignty**: Users can take their data (`user_id` export) or leave (`user_id` delete)
5. **Agent Memory**: System's own data (`app_id = '__SYSTEM__'`) is the agent's persistent memory
6. **Multi-Agent**: Multiple agents could manage different apps on same node

The spec may evolve, but the core invariants should remain stable. When in doubt,
ask: "Does this maintain clean ownership boundaries?"
