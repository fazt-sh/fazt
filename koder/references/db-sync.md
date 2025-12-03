# Fazt DB Sync — Technical Specification
# ============================================================
# One SQLite. Many Nodes. Zero Dependencies.
# ============================================================

## Overview

Bidirectional sync for Fazt's single-DB architecture.
Enables local-dev ↔ VPS synchronization with conflict resolution.

Goals:
  - No external services (S3, Redis, etc.)
  - No C extensions (pure Go + SQL)
  - Domain-specific merge strategies (not generic CRDTs)
  - Works offline, syncs when connected

Non-goals:
  - Real-time collaboration (use Yjs for that layer)
  - Multi-master with >2 nodes (keep simple for v1)


## Architecture

```
┌─────────────────┐                    ┌─────────────────┐
│  Local Fazt     │                    │  Remote Fazt    │
│                 │   POST /sync/push  │                 │
│  ┌───────────┐  │ ─────────────────► │  ┌───────────┐  │
│  │  data.db  │  │                    │  │  data.db  │  │
│  │           │  │   POST /sync/pull  │  │           │  │
│  │ _sync_log │  │ ◄───────────────── │  │ _sync_log │  │
│  └───────────┘  │                    │  └───────────┘  │
│                 │                    │                 │
│  node_id: A     │                    │  node_id: B     │
└─────────────────┘                    └─────────────────┘
```

Sync is pull-then-push. Remote-authoritative by default.


## Core Components

### 1. Hybrid Logical Clock (HLC)

Provides causal ordering without synchronized system clocks.

```go
// pkg/sync/hlc.go

type HLC struct {
    WallTime  int64  // unix millis
    Counter   uint16 // tie-breaker for same-ms events
    NodeID    string // 8-char unique identifier
}

// Key operations:
func (h *HLC) Now() Timestamp        // generate new timestamp
func (h *HLC) Update(remote Timestamp) // merge on receive
func Compare(a, b Timestamp) int     // -1, 0, 1

// Timestamp format (64-bit sortable):
// | wall_time (48 bits) | counter (16 bits) |
// String: "1705312847123-0042-nodeABCD"
```

Why HLC over vector clocks:
  - Single 64-bit value (fits in SQLite INTEGER)
  - Sortable without parsing
  - Good enough for 2-node sync

Reference: https://muratbuffalo.blogspot.com/2014/07/hybrid-logical-clocks.html


### 2. Change Tracking (_sync_log)

```sql
-- schema/sync.sql

CREATE TABLE _sync_meta (
    key   TEXT PRIMARY KEY,
    value TEXT
) STRICT;

-- Stores: node_id, last_sync_hlc, remote_url

CREATE TABLE _sync_log (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    hlc        INTEGER NOT NULL,           -- HLC timestamp
    node_id    TEXT NOT NULL,              -- origin node
    table_name TEXT NOT NULL,
    row_id     INTEGER NOT NULL,           -- PK of affected row
    op         TEXT NOT NULL,              -- INSERT | UPDATE | DELETE
    data       TEXT,                       -- JSON snapshot (NULL for DELETE)
    synced     INTEGER DEFAULT 0,          -- 0=pending, 1=synced
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
) STRICT;

CREATE INDEX idx_sync_pending ON _sync_log(synced, hlc);
CREATE INDEX idx_sync_table   ON _sync_log(table_name, row_id);
```


### 3. Trigger Generation

Auto-generate triggers for tracked tables.

```go
// pkg/sync/triggers.go

func GenerateTriggers(table string, pk string, cols []string) []string {
    // Returns SQL for INSERT, UPDATE, DELETE triggers
}

// Example output for 'sites' table:

/*
CREATE TRIGGER _sync_sites_insert AFTER INSERT ON sites
BEGIN
    INSERT INTO _sync_log (hlc, node_id, table_name, row_id, op, data)
    VALUES (
        fazt_hlc(),
        fazt_node_id(),
        'sites',
        NEW.id,
        'INSERT',
        json_object('id', NEW.id, 'name', NEW.name, 'domain', NEW.domain)
    );
END;

CREATE TRIGGER _sync_sites_update AFTER UPDATE ON sites
BEGIN
    INSERT INTO _sync_log (hlc, node_id, table_name, row_id, op, data)
    VALUES (
        fazt_hlc(),
        fazt_node_id(),
        'sites',
        NEW.id,
        'UPDATE',
        json_object('id', NEW.id, 'name', NEW.name, 'domain', NEW.domain)
    );
END;

CREATE TRIGGER _sync_sites_delete AFTER DELETE ON sites
BEGIN
    INSERT INTO _sync_log (hlc, node_id, table_name, row_id, op, data)
    VALUES (
        fazt_hlc(),
        fazt_node_id(),
        'sites',
        OLD.id,
        'DELETE',
        NULL
    );
END;
*/
```

Register SQLite functions via Go:

```go
// Register custom SQL functions
db.RegisterFunc("fazt_hlc", func() int64 {
    return hlc.Now().Int64()
})
db.RegisterFunc("fazt_node_id", func() string {
    return nodeID
})
```


### 4. Merge Strategies

```go
// pkg/sync/merge.go

type MergeStrategy interface {
    // Resolve conflict between local and remote versions
    Merge(ctx MergeContext) (Resolution, error)
}

type MergeContext struct {
    Table     string
    RowID     int64
    LocalHLC  int64
    RemoteHLC int64
    LocalData  map[string]any
    RemoteData map[string]any
}

type Resolution struct {
    Action string         // "use_local" | "use_remote" | "merge" | "skip"
    Data   map[string]any // merged result if Action="merge"
}
```

Built-in strategies:

```go
// Last-Write-Wins (default)
type LWW struct{}

func (LWW) Merge(ctx MergeContext) (Resolution, error) {
    if ctx.RemoteHLC > ctx.LocalHLC {
        return Resolution{Action: "use_remote"}, nil
    }
    return Resolution{Action: "use_local"}, nil
}

// Server-Authoritative (for certs, system config)
type ServerWins struct{}

func (ServerWins) Merge(ctx MergeContext) (Resolution, error) {
    return Resolution{Action: "use_remote"}, nil
}

// Field-Level Merge (for JSON documents)
type FieldMerge struct {
    Fields map[string]MergeStrategy // per-field strategy
}

// Append-Only (for analytics, logs)
type AppendOnly struct{}

func (AppendOnly) Merge(ctx MergeContext) (Resolution, error) {
    return Resolution{Action: "skip"}, nil // both versions kept
}
```

Table-to-strategy mapping:

```go
var DefaultStrategies = map[string]MergeStrategy{
    "sites":      LWW{},
    "files":      LWW{},            // or KeepBoth for safety
    "users":      LWW{},
    "certs":      ServerWins{},
    "analytics":  AppendOnly{},
    "config":     ServerWins{},
    "textdb_*":   FieldMerge{},     // wildcard for user collections
}
```


### 5. Sync Protocol

```go
// pkg/sync/protocol.go

// Push local changes to remote
type PushRequest struct {
    NodeID   string       `json:"node_id"`
    Since    int64        `json:"since_hlc"`  // last successful sync
    Changes  []ChangeEntry `json:"changes"`
}

type ChangeEntry struct {
    HLC       int64          `json:"hlc"`
    Table     string         `json:"table"`
    RowID     int64          `json:"row_id"`
    Op        string         `json:"op"`
    Data      map[string]any `json:"data,omitempty"`
}

type PushResponse struct {
    Accepted   int      `json:"accepted"`
    Conflicts  []Conflict `json:"conflicts"`
    RemoteHLC  int64    `json:"remote_hlc"`
}

// Pull remote changes to local
type PullRequest struct {
    NodeID  string `json:"node_id"`
    Since   int64  `json:"since_hlc"`
}

type PullResponse struct {
    Changes   []ChangeEntry `json:"changes"`
    RemoteHLC int64         `json:"remote_hlc"`
}

type Conflict struct {
    Table      string         `json:"table"`
    RowID      int64          `json:"row_id"`
    LocalHLC   int64          `json:"local_hlc"`
    RemoteHLC  int64          `json:"remote_hlc"`
    LocalData  map[string]any `json:"local_data"`
    RemoteData map[string]any `json:"remote_data"`
}
```

HTTP Endpoints:

```
POST /api/sync/push    — receive changes from another node
POST /api/sync/pull    — send changes to another node
GET  /api/sync/status  — current HLC, pending changes count
```


### 6. Sync Engine

```go
// pkg/sync/engine.go

type Engine struct {
    db         *sql.DB
    hlc        *HLC
    nodeID     string
    strategies map[string]MergeStrategy
    remote     string  // remote fazt URL
}

func NewEngine(db *sql.DB, nodeID string) *Engine

// Core operations
func (e *Engine) TrackTable(name string, strategy MergeStrategy) error
func (e *Engine) Push(ctx context.Context) (*PushResult, error)
func (e *Engine) Pull(ctx context.Context) (*PullResult, error)
func (e *Engine) Sync(ctx context.Context) error  // Pull then Push

// Conflict handling
func (e *Engine) PendingConflicts() ([]Conflict, error)
func (e *Engine) ResolveConflict(c Conflict, choice string) error
```

Sync flow:

```
Pull:
  1. GET /api/sync/pull?since=<last_hlc>
  2. For each remote change:
     a. Check if local row exists with higher HLC → conflict
     b. No conflict → apply change directly
     c. Conflict → use strategy to resolve
  3. Update last_sync_hlc

Push:
  1. SELECT * FROM _sync_log WHERE synced=0 ORDER BY hlc
  2. POST /api/sync/push with changes
  3. Handle response conflicts
  4. Mark pushed entries as synced=1
```


### 7. CLI Integration

```bash
# Set remote endpoint
fazt sync set-remote https://example.com

# Manual sync operations
fazt sync pull              # fetch remote changes
fazt sync push              # send local changes  
fazt sync                   # pull then push (default)

# Status and conflicts
fazt sync status            # show pending changes, last sync time
fazt sync conflicts         # list unresolved conflicts
fazt sync resolve <id>      # interactive conflict resolution

# Options
fazt sync --dry-run         # preview changes without applying
fazt sync --force-local     # resolve all conflicts with local version
fazt sync --force-remote    # resolve all conflicts with remote version
```

Config stored in `_sync_meta`:

```sql
INSERT INTO _sync_meta (key, value) VALUES
    ('node_id', 'abc12345'),
    ('remote_url', 'https://my.fazt.sh'),
    ('last_push_hlc', '1705312847123'),
    ('last_pull_hlc', '1705312841456');
```


## Yjs Integration (for app-pad, docs)

Yjs handles real-time CRDT sync. Fazt provides:
  1. Storage backend (Y.Doc state in `files` table)
  2. WebSocket relay (via serverless or dedicated handler)

```
┌──────────────┐     WebSocket      ┌──────────────┐
│  Browser A   │ ◄────────────────► │    Fazt      │
│  (Yjs)       │                    │   Server     │
└──────────────┘                    │              │
                                    │  ┌────────┐  │
┌──────────────┐     WebSocket      │  │ Y.Doc  │  │
│  Browser B   │ ◄────────────────► │  │ state  │  │
│  (Yjs)       │                    │  │ in DB  │  │
└──────────────┘                    │  └────────┘  │
                                    └──────────────┘
```

Storage schema:

```sql
CREATE TABLE _yjs_docs (
    doc_id     TEXT PRIMARY KEY,    -- e.g., "pad:site123:doc456"
    state      BLOB NOT NULL,       -- Y.encodeStateAsUpdate()
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
) STRICT;
```

Serverless handler (JS):

```javascript
// api/yjs-sync.js
import * as Y from 'yjs'

export default async function(req, ctx) {
    const { docId } = req.params
    
    if (req.method === 'GET') {
        // Load doc state
        const row = await ctx.db.get(
            'SELECT state FROM _yjs_docs WHERE doc_id = ?',
            [docId]
        )
        return new Response(row?.state || new Uint8Array())
    }
    
    if (req.method === 'POST') {
        // Merge incoming update
        const update = new Uint8Array(await req.arrayBuffer())
        const existing = await ctx.db.get(
            'SELECT state FROM _yjs_docs WHERE doc_id = ?',
            [docId]
        )
        
        const doc = new Y.Doc()
        if (existing) {
            Y.applyUpdate(doc, existing.state)
        }
        Y.applyUpdate(doc, update)
        
        const merged = Y.encodeStateAsUpdate(doc)
        await ctx.db.run(
            `INSERT OR REPLACE INTO _yjs_docs (doc_id, state, updated_at)
             VALUES (?, ?, datetime('now'))`,
            [docId, merged]
        )
        
        return new Response(merged)
    }
}
```

Fazt's DB sync handles _yjs_docs like any other table (LWW on the blob).
Yjs handles intra-document merge semantics.


## File Structure

```
pkg/
  sync/
    hlc.go           # Hybrid Logical Clock implementation
    hlc_test.go
    triggers.go      # SQL trigger generation
    triggers_test.go
    merge.go         # MergeStrategy interface + built-ins
    merge_test.go
    protocol.go      # Request/Response types
    engine.go        # Main sync orchestration
    engine_test.go
    handlers.go      # HTTP handlers for /api/sync/*

internal/
  cli/
    sync.go          # CLI commands: sync, sync pull, sync push

schema/
  sync.sql           # _sync_log, _sync_meta tables
```


## Testing Strategy

```go
// pkg/sync/engine_test.go

func TestSyncBasic(t *testing.T) {
    // Two in-memory DBs simulating local/remote
    local := setupTestDB(t, "node-A")
    remote := setupTestDB(t, "node-B")
    
    // Insert on local
    local.Exec(`INSERT INTO sites (name) VALUES ('test')`)
    
    // Sync
    localEngine := NewEngine(local.DB, "node-A")
    localEngine.SetRemote(mockServer(remote))
    localEngine.Sync(ctx)
    
    // Verify remote has the row
    var name string
    remote.QueryRow(`SELECT name FROM sites WHERE name='test'`).Scan(&name)
    assert.Equal(t, "test", name)
}

func TestSyncConflictLWW(t *testing.T) {
    local := setupTestDB(t, "node-A")
    remote := setupTestDB(t, "node-B")
    
    // Same row, different values, remote is newer
    local.Exec(`INSERT INTO sites (id, name) VALUES (1, 'local-name')`)
    remote.Exec(`INSERT INTO sites (id, name) VALUES (1, 'remote-name')`)
    // Manually set remote HLC higher
    
    localEngine := NewEngine(local.DB, "node-A")
    localEngine.SetRemote(mockServer(remote))
    localEngine.Pull(ctx)
    
    // Local should now have remote's value
    var name string
    local.QueryRow(`SELECT name FROM sites WHERE id=1`).Scan(&name)
    assert.Equal(t, "remote-name", name)
}

func TestHLCOrdering(t *testing.T) {
    hlc := NewHLC("nodeA")
    
    t1 := hlc.Now()
    t2 := hlc.Now()
    t3 := hlc.Now()
    
    assert.True(t, Compare(t1, t2) < 0)
    assert.True(t, Compare(t2, t3) < 0)
}

func TestHLCMerge(t *testing.T) {
    hlcA := NewHLC("nodeA")
    hlcB := NewHLC("nodeB")
    
    // B is "ahead" in wall time (simulated)
    remoteTime := hlcB.Now()
    
    // A receives remote timestamp, should update
    hlcA.Update(remoteTime)
    localTime := hlcA.Now()
    
    assert.True(t, Compare(remoteTime, localTime) < 0)
}

func TestTriggerGeneration(t *testing.T) {
    sql := GenerateTriggers("sites", "id", []string{"name", "domain"})
    
    assert.Contains(t, sql[0], "AFTER INSERT ON sites")
    assert.Contains(t, sql[1], "AFTER UPDATE ON sites")
    assert.Contains(t, sql[2], "AFTER DELETE ON sites")
}
```


## Edge Cases to Handle

1. **Clock drift**
   HLC's counter handles same-millisecond events.
   Large drift (>1min) should log warning, not fail.

2. **Partial sync failure**
   Track last successful HLC per direction (push/pull).
   Resume from last success, don't re-send everything.

3. **Schema mismatch**
   Version tables. Reject sync if schema versions differ.
   Force user to upgrade both nodes first.

4. **Large files in VFS**
   Don't put file content in _sync_log.data.
   Store hash only, separate blob transfer endpoint.

5. **Deleted then recreated**
   Same row_id, DELETE then INSERT.
   Sequence matters. Process _sync_log in HLC order.

6. **Network interruption mid-sync**
   Sync is not atomic across tables.
   Track per-table progress. Resume gracefully.


## Future Enhancements (Out of Scope for v1)

- [ ] Background sync daemon (fazt service with sync interval)
- [ ] Selective sync (sync only specific sites/tables)
- [ ] Sync over SSH tunnel (for firewalled VPS)
- [ ] Multi-node (>2) with crdt-style vector clocks
- [ ] Binary diff for large blobs (rsync-style)
- [ ] Encryption in transit (beyond HTTPS)


## Implementation Order

```
Phase 1: Foundation
  [x] Design complete (this doc)
  [ ] HLC implementation + tests
  [ ] _sync_log schema + triggers
  [ ] Trigger generation for existing tables

Phase 2: Protocol  
  [ ] Push/Pull request handlers
  [ ] Basic LWW merge
  [ ] CLI: fazt sync pull/push

Phase 3: Robustness
  [ ] Conflict detection + resolution UI
  [ ] All merge strategies
  [ ] Edge case handling

Phase 4: Integration
  [ ] Yjs storage backend
  [ ] WebSocket relay for real-time
  [ ] Background sync option
```


## References

- HLC Paper: https://cse.buffalo.edu/tech-reports/2014-04.pdf
- Yjs Docs: https://docs.yjs.dev/
- SQLite JSON: https://sqlite.org/json1.html
- mattn/go-sqlite3: https://github.com/mattn/go-sqlite3


## Notes for Implementer

1. Start with HLC. It's pure logic, easy to test, no DB needed.

2. Test triggers manually in sqlite3 CLI before generating.

3. Mock HTTP for engine tests. Don't need real server.

4. LWW is the 80% case. Get it solid before fancy merges.

5. The _sync_log will grow. Add periodic cleanup:
   `DELETE FROM _sync_log WHERE synced=1 AND hlc < ?`

6. Keep node_id short (8 chars). It appears in every log row.

7. Don't forget: triggers must be recreated after schema changes.