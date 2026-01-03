# Evolutionary Database Design (EDD)

## Summary

EDD is a **non-destructive** schema philosophy. The database is treated as an
**append-only** ledger. Schemas evolve by adding columns and tables, never by
removing or renaming them.

## Rationale

### The Migration Trap

Traditional migrations break old binaries:

```sql
-- Migration 005: Rename column
ALTER TABLE sites RENAME COLUMN name TO title;
```

Problem: v0.7 binary crashes when it tries to SELECT `name`.

### The EDD Solution

Never remove, never rename. Only add.

```sql
-- Migration 005: Add new column
ALTER TABLE sites ADD COLUMN title TEXT;
-- Application logic: Read from title, fallback to name
```

## Rules

### 1. Nullable Only

Every new field must be optional or have a default:

```sql
-- Good
ALTER TABLE apps ADD COLUMN description TEXT;
ALTER TABLE apps ADD COLUMN priority INTEGER DEFAULT 0;

-- Bad (breaks existing rows)
ALTER TABLE apps ADD COLUMN required_field TEXT NOT NULL;
```

### 2. No Drops

Never use `DROP` or `RENAME`:

```sql
-- Forbidden
DROP TABLE old_sites;
ALTER TABLE apps DROP COLUMN legacy_field;
ALTER TABLE apps RENAME COLUMN old_name TO new_name;
```

### 3. Soft Relations

Use IDs without `FOREIGN KEY` constraints:

```sql
-- Good: Soft reference
CREATE TABLE files (
    app_id TEXT,  -- References apps.id, but no FK constraint
    ...
);

-- Bad: Hard constraint
CREATE TABLE files (
    app_id TEXT REFERENCES apps(id) ON DELETE CASCADE,
    ...
);
```

Integrity is enforced in Go, not SQL.

### 4. Version Mapping

Go structs map to active columns only:

```go
type App struct {
    ID          string  // Always present
    Name        string  // v0.7+
    Title       string  // v0.8+ (new)
    Description string  // v0.8+ (new)
    // LegacyField is NOT in struct even if column exists
}
```

## Advantages

| Benefit               | Description                              |
| --------------------- | ---------------------------------------- |
| **Perfect Rollbacks** | Old binaries always work with new DB     |
| **Zero Downtime**     | No table locks for complex alterations   |
| **Solo-Friendly**     | Reduces deployment risk                  |
| **Backup Safety**     | Any backup works with any binary version |

## Disadvantages

| Drawback           | Mitigation                             |
| ------------------ | -------------------------------------- |
| **Schema Bloat**   | Unused columns persist (disk is cheap) |
| **Data Integrity** | Must be managed in Go code             |
| **Messy Tables**   | Use views or documentation             |

## Implementation

### Migration File Convention

```sql
-- migrations/008_add_app_fields.sql
-- EDD: Adding new fields for v0.8 app model

ALTER TABLE sites ADD COLUMN uuid TEXT;
ALTER TABLE sites ADD COLUMN manifest TEXT;  -- JSON blob

CREATE INDEX IF NOT EXISTS idx_sites_uuid ON sites(uuid);
```

### Go Struct Pattern

```go
// AppFromRow handles both v0.7 and v0.8 schemas
func AppFromRow(row *sql.Row) (*App, error) {
    app := &App{}

    // Scan required fields
    err := row.Scan(&app.ID, &app.Name, ...)
    if err != nil {
        return nil, err
    }

    // Handle optional v0.8 fields
    if app.Title == "" {
        app.Title = app.Name  // Fallback
    }

    return app, nil
}
```

### Schema Documentation

Maintain a `SCHEMA.md` that documents:
- Active columns (used by current binary)
- Legacy columns (kept for compatibility)
- Column version history

## Architect's Advice

> "The DB is a bucket. It holds data; it doesn't judge validity."
> "Go is the brain. It knows what's valid; SQL does not."
> "Storage is cheap. Uptime and sanity are expensive."
> "Accept the mess. A messy working DB beats a clean dead one."
