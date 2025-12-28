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

| Operator | Description | Example |
|----------|-------------|---------|
| `$eq` | Equals | `{ status: { $eq: 'active' } }` |
| `$ne` | Not equals | `{ status: { $ne: 'deleted' } }` |
| `$gt` | Greater than | `{ age: { $gt: 18 } }` |
| `$lt` | Less than | `{ age: { $lt: 65 } }` |
| `$in` | In array | `{ role: { $in: ['admin', 'mod'] } }` |
| `$contains` | Array contains | `{ tags: { $contains: 'featured' } }` |
