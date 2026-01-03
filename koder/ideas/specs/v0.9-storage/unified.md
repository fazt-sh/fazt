# Unified Storage Abstraction

## Summary

`fazt.storage` provides a single, provider-agnostic API for all application
state. Whether data lives in SQLite or AWS, the application code is identical.

## Rationale

### The Problem

Apps need different storage patterns:
- Settings → Key-Value
- User data → Documents
- Files → Blobs
- Analytics → Time-series

Without abstraction, apps import raw SQL or provider-specific SDKs.

### The Solution

One namespace, four pillars:

```javascript
fazt.storage.kv      // Key-Value
fazt.storage.ds      // Document Store
fazt.storage.rd      // Relational (future)
fazt.storage.s3      // Blob Storage
```

## The Four Pillars

### 1. Key-Value (`kv`)

High-speed, persistent key-value store.

```javascript
// Set with optional TTL (milliseconds)
await fazt.storage.kv.set('user:123:session', token, 3600000);

// Get
const session = await fazt.storage.kv.get('user:123:session');

// Delete
await fazt.storage.kv.delete('user:123:session');

// List by prefix
const sessions = await fazt.storage.kv.list('user:123:');
```

**Implementation**: SQLite table with key/value/expiry columns.

**State Prefix Convention** (pattern from Google ADK):
```
app:key    → shared across all users of an app
user:key   → shared across a user's sessions
temp:key   → in-memory only, auto-expires
key        → session-scoped (default)
```

Single API (`kv.get/set`), automatic scoping via prefix. Useful for harness
agents that need hierarchical state. See `koder/ideas/lite-extractions.md`
(go-adk PATTERN verdict).

### 2. Document Store (`ds`)

JSON document storage with query support.

```javascript
// Insert
const id = await fazt.storage.ds.insert('users', {
  email: 'alice@example.com',
  name: 'Alice',
  tags: ['admin', 'active']
});

// Find with query
const admins = await fazt.storage.ds.find('users', {
  'tags': { $contains: 'admin' }
});

// Update
await fazt.storage.ds.update('users',
  { email: 'alice@example.com' },
  { $set: { name: 'Alice Smith' } }
);

// Delete
await fazt.storage.ds.delete('users', { email: 'alice@example.com' });
```

**Implementation**: SQLite with JSON columns and functional indexes.

### 3. Blob Storage (`s3`)

File storage with S3-compatible semantics.

```javascript
// Put
await fazt.storage.s3.put('uploads/photo.jpg', imageData, 'image/jpeg');

// Get
const { data, mime } = await fazt.storage.s3.get('uploads/photo.jpg');

// Delete
await fazt.storage.s3.delete('uploads/photo.jpg');

// List
const files = await fazt.storage.s3.list('uploads/');
```

**Implementation**: SQLite `blobs` table or external S3 bucket.

### 4. Relational (`rd`) - Future

Namespaced virtual tables for complex queries.

```javascript
// Define schema
await fazt.storage.rd.define('orders', {
  id: 'TEXT PRIMARY KEY',
  user_id: 'TEXT',
  total: 'REAL',
  created_at: 'TIMESTAMP'
});

// Query
const orders = await fazt.storage.rd.query(
  'SELECT * FROM orders WHERE user_id = ? ORDER BY created_at DESC',
  [userId]
);
```

## Namespacing

All storage is automatically namespaced by `app_uuid`:

```sql
-- KV Table
CREATE TABLE app_kv (
    app_uuid TEXT,
    key TEXT,
    value TEXT,
    expires_at INTEGER,
    PRIMARY KEY (app_uuid, key)
);

-- Document Table
CREATE TABLE app_docs (
    app_uuid TEXT,
    collection TEXT,
    id TEXT,
    data TEXT,  -- JSON
    PRIMARY KEY (app_uuid, collection, id)
);
```

Apps cannot access other apps' data. The kernel enforces isolation.

## Provider Configuration

In `app.json`:

```json
{
  "storage": {
    "kv": "internal",
    "ds": "internal",
    "s3": "external:s3://my-bucket"
  }
}
```

- `internal`: Uses SQLite (default)
- `external:s3://...`: Routes to AWS S3
- `external:postgres://...`: Routes to Postgres (future)

## Migration Edge

The kernel provides migration tooling:

```bash
# Export internal storage to S3
fazt storage migrate app_x9z2k --from internal --to s3://bucket

# Zero application code changes needed
```

## Query Operators

For document store queries:

| Operator    | Description    | Example                               |
| ----------- | -------------- | ------------------------------------- |
| `$eq`       | Equals         | `{ status: { $eq: 'active' } }`       |
| `$ne`       | Not equals     | `{ status: { $ne: 'deleted' } }`      |
| `$gt`       | Greater than   | `{ age: { $gt: 18 } }`                |
| `$lt`       | Less than      | `{ age: { $lt: 65 } }`                |
| `$in`       | In array       | `{ role: { $in: ['admin', 'mod'] } }` |
| `$contains` | Array contains | `{ tags: { $contains: 'featured' } }` |

## Advanced Schema Patterns

Reference: [Redka](https://github.com/nalgeon/redka) (BSD-3, Redis-on-SQLite)
demonstrates patterns for SQL-backed data structures. Key patterns for future
extensions (see `koder/ideas/lite-extractions.md` PATTERN verdict):

1. **Unified key metadata table**: Single `rkey` table tracks key existence,
   type, TTL, length. Type-specific tables (string, list, set, hash) hold
   values with foreign key to `rkey`. `ON DELETE CASCADE` handles cleanup.

2. **Real-valued list positions**: Use `REAL` for position instead of `INTEGER`.
   Insert between pos=1.0 and pos=2.0 → use pos=1.5. O(1) insertions anywhere.

3. **View-based TTL filtering**: Create views that auto-filter expired keys:
   `WHERE etime IS NULL OR etime > unixepoch('subsec')`. Background GC cleans.

4. **Trigger-maintained consistency**: Database triggers update derived values
   (e.g., list length) so application code stays simple.

Adopt these patterns if extending KV to richer data structures (lists, sets,
sorted sets, hashes) while maintaining SQLite-first architecture.
