# Micro-Document Storage Pattern (Shards)

## Summary

A storage pattern using SQLite as a document database. It replaces giant JSON
blobs with **sharded rows** and uses **functional indexes** to query JSON keys
at B-tree speeds.

## Rationale

### The Problem

High-volume data (analytics, logs, events) can overwhelm SQLite:
- Single table with millions of rows
- JSON blobs require full-table scans
- Index on entire column is wasteful

### The Solution

1. **Prefix Sharding**: Split keys by time/category
2. **Functional Indexes**: Index specific JSON fields
3. **Batch Buffering**: RAM → Disk flushing
4. **JSONB Storage**: Binary JSON for faster parsing

## Core Components

### 1. Prefix Sharding

Keys are structured to enable efficient range queries:

```
events:2025:01:shard_001
events:2025:01:shard_002
events:2025:02:shard_001
```

Benefits:
- Query single month: `WHERE key LIKE 'events:2025:01:%'`
- Archival: Move old shards to cold storage
- Cleanup: Drop entire shards without vacuuming

### 2. Functional Indexes

Index specific JSON paths, not entire documents:

```sql
CREATE TABLE events (
    key TEXT PRIMARY KEY,
    value TEXT  -- JSON blob
);

-- Index specific fields for fast lookup
CREATE INDEX idx_events_type ON events(
    json_extract(value, '$.type')
);

CREATE INDEX idx_events_user ON events(
    json_extract(value, '$.user_id')
);
```

Query performance:
- Without index: Full table scan (slow)
- With functional index: B-tree lookup (fast)

### 3. Batch Buffering

Analytics events buffer in RAM before flushing:

```go
type EventBuffer struct {
    events []Event
    mu     sync.Mutex
    ticker *time.Ticker
}

func (b *EventBuffer) Add(e Event) {
    b.mu.Lock()
    b.events = append(b.events, e)
    b.mu.Unlock()

    if len(b.events) >= 1000 {
        b.Flush()
    }
}

func (b *EventBuffer) Flush() {
    // Batch insert 1000 events in single transaction
}
```

Benefits:
- Reduces SQLite write contention
- Groups events for efficient sharding
- Survives brief spikes

### 4. JSONB Storage

SQLite 3.45+ supports binary JSON:

```sql
-- Store as JSONB (2-3x faster parsing)
INSERT INTO events (key, value) VALUES (?, jsonb(?));

-- Query syntax unchanged
SELECT * FROM events WHERE json_extract(value, '$.type') = 'pageview';
```

## Schema Design

```sql
-- Sharded event storage
CREATE TABLE event_shards (
    shard_key TEXT,           -- "events:2025:01:001"
    data TEXT,                -- JSON array of events
    count INTEGER,            -- Number of events in shard
    created_at INTEGER,
    PRIMARY KEY (shard_key)
);

-- Functional indexes for common queries
CREATE INDEX idx_shard_month ON event_shards(
    substr(shard_key, 1, 14)  -- "events:2025:01"
);
```

## Implementation Rules

### Index the Intent

Only index keys used for `WHERE` or `JOIN`:

```sql
-- Good: Index fields used in queries
CREATE INDEX idx_type ON events(json_extract(value, '$.type'));

-- Bad: Index everything (wasteful)
CREATE INDEX idx_all ON events(value);
```

### Shard by Pattern

Auto-split when a key prefix exceeds threshold:

```go
func (s *Storage) Write(key string, value interface{}) {
    shardKey := s.getShardKey(key)

    count := s.shardCount(shardKey)
    if count >= MAX_SHARD_SIZE {
        shardKey = s.nextShard(key)
    }

    s.insert(shardKey, value)
}
```

### Transparent API

Apps don't see shards. The `syscall` layer handles it:

```javascript
// App code (simple)
await fazt.storage.ds.insert('events', { type: 'pageview', url: '/' });

// Kernel handles sharding internally
```

## Performance Targets

| Metric                     | Target         | Notes            |
| -------------------------- | -------------- | ---------------- |
| Write throughput           | 10k events/sec | With buffering   |
| Query by indexed field     | <10ms          | Functional index |
| Query by non-indexed field | <1s            | Full shard scan  |
| Storage per 1M events      | ~500MB         | With compression |

## Architect's Advice

> "Treat the DB as a searchable log."
> "Use the kernel to hide sharding logic from the app."
> "Store everything as JSON; index only what matters for speed."
> "Sanity is found in batching; uptime is found in indexes."
