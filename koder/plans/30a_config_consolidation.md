# Plan 30a: Config Consolidation

**Status**: Ready to implement
**Created**: 2026-01-30
**Target**: v0.14.0
**Depends on**: None
**Unlocks**: Plan 30b, 30c

## Overview

Restore fazt's core philosophy: **single binary + single SQLite database**.

Move all configuration into the database. Eliminate external config files.

## Problem

Current state violates single-DB philosophy:

| File | Contains | Problem |
|------|----------|---------|
| `~/.config/fazt/config.json` | port, domain, auth, ntfy, HTTPS | Outside DB |
| `~/.fazt/config.json` | peer list, tokens | Outside DB |
| `data.db` | apps, storage, analytics | Correct |

The fazt binary IS the CLI. Peers are uniform (local = remote, just different domains).
Having separate "client config" makes no sense.

## Solution

**Eliminate all external config files.** Everything goes in the database.

**Bootstrap (only exception):**
- DB path via: `--db` flag, `FAZT_DB_PATH` env, or default `./data.db`
- Nothing else needed outside

---

## Config Table

```sql
CREATE TABLE config (
  key TEXT PRIMARY KEY,
  value JSON NOT NULL,
  updated_at INTEGER DEFAULT (unixepoch())
);
```

### Key Hierarchy

```
config
├── instance.*           # Server settings (port, domain, env)
├── auth.*               # Instance auth (username, password_hash)
├── ntfy.*               # Notifications
├── https.*              # TLS settings
├── peers.*              # Known peers
│   ├── <name>.url
│   ├── <name>.token
│   └── default
└── app.<name>.*         # Per-app config (used by 30c)
```

### Example Data

```sql
-- Instance config
INSERT INTO config (key, value) VALUES
  ('instance.port', '4698'),
  ('instance.domain', '"https://zyt.app"'),
  ('instance.env', '"production"'),
  ('auth.username', '"admin"'),
  ('auth.password_hash', '"$2a$..."'),
  ('ntfy.topic', '"fazt-alerts"'),
  ('ntfy.url', '"https://ntfy.sh"'),
  ('https.enabled', 'true'),
  ('https.email', '"admin@zyt.app"');

-- Peer config (replaces ~/.fazt/config.json)
INSERT INTO config (key, value) VALUES
  ('peers.zyt.url', '"https://admin.zyt.app"'),
  ('peers.zyt.token', '"fazt_tok_..."'),
  ('peers.local.url', '"http://localhost:4698"'),
  ('peers.local.token', '"fazt_tok_..."'),
  ('peers.default', '"zyt"');
```

---

## Migration Strategy

On startup, if external config files exist:

1. Read `~/.config/fazt/config.json` (instance config)
2. Read `~/.fazt/config.json` (peer config)
3. Import into DB config table
4. Rename originals to `.bak` (safety, don't delete)
5. Log: "Config migrated to database"

---

## CLI Changes

```bash
# Before (reads ~/.fazt/config.json)
fazt remote list

# After (reads from DB)
fazt remote list      # Same command, reads from DB

# Peer management now updates DB
fazt remote add zyt --url https://admin.zyt.app --token fazt_tok_...
# → INSERT INTO config (key, value) VALUES ('peers.zyt.url', ...)

fazt remote remove zyt
# → DELETE FROM config WHERE key LIKE 'peers.zyt.%'

fazt remote default zyt
# → UPDATE config SET value = '"zyt"' WHERE key = 'peers.default'
```

---

## Server Startup Changes

```go
// Before
func main() {
    cfg, _ := config.LoadFromFile("~/.config/fazt/config.json")
    // ...
}

// After
func main() {
    dbPath := getDBPath()  // --db flag, env, or default
    db := openDB(dbPath)
    cfg := config.LoadFromDB(db)
    // ...
}

func getDBPath() string {
    // 1. --db flag (highest priority)
    if flagDB != "" {
        return flagDB
    }
    // 2. FAZT_DB_PATH env
    if envDB := os.Getenv("FAZT_DB_PATH"); envDB != "" {
        return envDB
    }
    // 3. Default
    return "./data.db"
}
```

---

## Systemd Service

The service file specifies DB location (only external thing needed):

```ini
# ~/.config/systemd/user/fazt-local.service
[Unit]
Description=Fazt Local Server

[Service]
Environment=FAZT_DB_PATH=/home/user/.local/share/fazt/data.db
ExecStart=/home/user/.local/bin/fazt serve
Restart=always

[Install]
WantedBy=default.target
```

Or via flag:

```ini
ExecStart=/home/user/.local/bin/fazt serve --db /path/to/data.db
```

---

## What This Fixes

| Before | After |
|--------|-------|
| 3 places to look for config | 1 place (DB) |
| Can't query config | `SELECT * FROM config WHERE key LIKE 'peers.%'` |
| Config doesn't travel with DB | Copy DB = copy everything |
| CLI vs server have different views | Unified |
| "Client config" concept | Gone - fazt IS the client |

---

## Implementation Tasks

### Database

- [ ] Create `config` table with schema above
- [ ] Add migration to create table on startup

### Config Package

- [ ] Create `internal/config/db.go` for DB-based config
- [ ] Implement `config.Get(key string) (any, error)`
- [ ] Implement `config.Set(key string, value any) error`
- [ ] Implement `config.GetString(key)`, `config.GetInt(key)`, `config.GetBool(key)`
- [ ] Implement `config.List(prefix string) map[string]any`
- [ ] Implement `config.Delete(key string) error`

### Migration

- [ ] Detect existing `~/.config/fazt/config.json`
- [ ] Detect existing `~/.fazt/config.json`
- [ ] Import instance config to DB
- [ ] Import peer config to DB
- [ ] Rename originals to `.bak`
- [ ] Log migration status

### Server Startup

- [ ] Update `cmd/server/main.go` to use DB config
- [ ] Implement `getDBPath()` (flag > env > default)
- [ ] Remove dependency on `~/.config/fazt/config.json`

### CLI Commands

- [ ] Update `fazt remote list` to read from DB
- [ ] Update `fazt remote add` to write to DB
- [ ] Update `fazt remote remove` to delete from DB
- [ ] Update `fazt remote default` to update DB
- [ ] Update `fazt remote status` to read from DB
- [ ] Remove `internal/clientconfig/` package

### Testing

- [ ] Config read/write tests
- [ ] Migration tests (both files present, one present, none present)
- [ ] Bootstrap tests (flag, env, default)
- [ ] CLI remote commands with DB backend

### Documentation

- [ ] Update setup docs (no more config.json)
- [ ] Update remote/peer docs
- [ ] Migration guide for existing users

---

## Success Criteria

1. **Zero external config files** after migration
2. All config stored in SQLite `config` table
3. Existing JSON configs auto-migrated on first startup
4. `fazt remote *` commands work with DB-stored peers
5. Server starts with only DB path specified
6. Systemd service works with `FAZT_DB_PATH` env

---

## Files to Modify

| File | Change |
|------|--------|
| `internal/config/config.go` | Add DB-based loading |
| `internal/config/db.go` | New: DB config functions |
| `internal/clientconfig/` | Delete entire package |
| `cmd/server/main.go` | Use DB config, bootstrap logic |
| `cmd/server/remote.go` | Use DB for peer management |
| `internal/database/migrations/` | Add config table migration |

---

## Rollback Plan

If issues arise:
1. Config table remains in DB
2. Can restore `.bak` files to original names
3. Revert code to read from JSON files

Low risk - config is simple key-value data.
