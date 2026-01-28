# Plan 25: SQL Command

**Status**: Proposed
**Created**: 2026-01-29

## Summary

Add `fazt sql` command for running SQL queries against fazt databases, both
local and remote. Enables debugging and inspection without direct db access.

## Motivation

Currently debugging requires:
- SSH into server + sqlite3 CLI for remote
- Finding the db file + sqlite3 for local

With `fazt sql`:
```bash
fazt sql "SELECT * FROM apps"              # Local
fazt @zyt sql "SELECT * FROM auth_users"   # Remote
```

Same tool, same syntax, local or remote.

## Design

### CLI Interface

```bash
# Local (direct db access)
fazt sql "SELECT * FROM apps"
fazt sql --db ./data.db "SELECT ..."       # Explicit db path
fazt sql --write "UPDATE config SET ..."   # Mutations require flag

# Remote (via API)
fazt @zyt sql "SELECT * FROM apps"
fazt @zyt sql --write "DELETE FROM ..."
```

### Flags

| Flag | Description |
|------|-------------|
| `--db <path>` | Database path (local only, defaults to configured db) |
| `--write` | Allow mutations (INSERT/UPDATE/DELETE) |
| `--format` | Output format: `table` (default), `json`, `csv` |
| `--limit` | Max rows to return (default: 100) |

### API Endpoint

```
POST /api/sql
Authorization: Bearer <api-key>

{
  "query": "SELECT * FROM apps LIMIT 10",
  "write": false
}
```

Response:
```json
{
  "columns": ["id", "name", "created_at"],
  "rows": [
    ["app_abc", "tetris", 1706500000],
    ["app_def", "notes", 1706500100]
  ],
  "count": 2,
  "time_ms": 3
}
```

For mutations:
```json
{
  "affected": 1,
  "time_ms": 2
}
```

### Safety

1. **Read-only by default** - Mutations require explicit `--write` flag
2. **API key required** - Remote access already protected
3. **Query timeout** - 30 second max execution time
4. **Row limit** - Default 100 rows, configurable

### Optional Safety (Later)

- Block dangerous patterns without `--force`: `DROP TABLE`, `DELETE FROM x`
  (no WHERE)
- Audit log for remote SQL execution
- Role-based: owner only vs admin

## Implementation

### Phase 1: Local Command

```
cmd/server/sql.go
```

1. Parse query from args
2. Open database (configured or `--db` flag)
3. Execute query (check `--write` for mutations)
4. Format and print results

### Phase 2: Remote API

```
internal/handlers/sql.go
```

1. Add `POST /api/sql` endpoint
2. Require API key authentication
3. Parse request, execute query
4. Return JSON response

### Phase 3: Remote CLI

Update `cmd/server/sql.go`:
1. Detect `@peer` prefix
2. Send request to peer's `/api/sql` endpoint
3. Format and print response

## Files to Create/Modify

| File | Action |
|------|--------|
| `cmd/server/sql.go` | New - CLI command |
| `cmd/server/main.go` | Add sql command registration |
| `internal/handlers/sql.go` | New - API handler |
| `internal/handlers/routes.go` | Register endpoint |

## Output Examples

### Table Format (default)

```
$ fazt sql "SELECT id, name FROM apps LIMIT 3"
ID          NAME
-----------------------
app_abc123  tetris
app_def456  notes
app_ghi789  photos

3 rows (2ms)
```

### JSON Format

```
$ fazt sql --format json "SELECT id, name FROM apps LIMIT 2"
[
  {"id": "app_abc123", "name": "tetris"},
  {"id": "app_def456", "name": "notes"}
]
```

### Mutation

```
$ fazt sql --write "UPDATE config SET value = 'new' WHERE key = 'theme'"
1 row affected (1ms)
```

## Use Cases

1. **Debug auth issues**: `fazt @zyt sql "SELECT * FROM auth_sessions"`
2. **Check app storage**: `fazt sql "SELECT * FROM storage WHERE app_id = '...'"`
3. **Inspect config**: `fazt sql "SELECT * FROM config"`
4. **Count users**: `fazt @zyt sql "SELECT COUNT(*) FROM auth_users"`
5. **Check migrations**: `fazt sql "SELECT * FROM migrations"`

## Not in Scope

- Interactive SQL shell (just single queries)
- Schema management (use migrations)
- Backup/restore (separate commands)
