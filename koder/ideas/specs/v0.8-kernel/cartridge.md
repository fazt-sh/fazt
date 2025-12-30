# Cartridge (App Data Portability)

## Summary

A cartridge is a portable, self-contained SQLite file containing all data
for a single app. Export an app, copy the file, import elsewhere.

Built on [Provenance](./provenance.md) - the `app_id` column makes clean cuts possible.

## What's in a Cartridge

```
my-app.cart (SQLite database)
├── _meta              # Cartridge metadata
├── files              # VFS files
├── storage_kv         # Key-value data
├── storage_ds         # Document store
├── storage_s3         # Blob references
├── events             # App events
├── analytics          # App analytics
├── redirects          # Redirects
├── webhooks           # Webhooks
├── workers            # Job history
├── email              # Received emails
├── comments           # Comments
├── short_urls         # Short URLs
├── forms              # Form submissions
└── search_indexes     # Search data
```

Every table has the same schema as main.db, filtered to one `app_id`.

## Metadata Table

```sql
CREATE TABLE _meta (
    key TEXT PRIMARY KEY,
    value TEXT
);

-- Required entries:
-- fazt_version: "0.8.0"
-- app_id: "original-uuid"
-- app_name: "my-app"
-- exported_at: 1704067200
-- schema_version: 1
```

## Export

```bash
fazt app export my-app --output my-app.cart
```

What happens:
1. Create new SQLite file
2. Create `_meta` table with metadata
3. For each app-scoped table:
   ```sql
   INSERT INTO cart.{table}
   SELECT * FROM main.{table}
   WHERE app_id = :app_id AND deleted_at IS NULL;
   ```
4. Return file path

## Import

```bash
fazt app import my-app.cart [--mode overwrite|skip|merge] [--name new-name]
```

### Modes

| Mode | Behavior |
|------|----------|
| `skip` (default) | Fail if app exists |
| `overwrite` | Delete existing app, import fresh |
| `merge` | Add new rows, skip existing (by id) |

### App ID Handling

| Scenario | Behavior |
|----------|----------|
| App doesn't exist | Create with original UUID |
| App exists, different node | Create with original UUID |
| App exists, same node, `--name` given | Create with new UUID and name |
| App exists, same node, no `--name` | Fail (or use mode) |

```bash
# Import as new app with different name
fazt app import my-app.cart --name my-app-copy
# → Creates new UUID, name = "my-app-copy"
```

## File Format Details

- **Extension**: `.cart`
- **Format**: SQLite 3 database
- **Compression**: None (SQLite handles it)
- **Encryption**: None (use OS-level encryption if needed)
- **Max size**: 1GB default (configurable)

## Integrity

On import, kernel verifies:
1. `_meta` table exists with required keys
2. `schema_version` is compatible
3. All rows have same `app_id`
4. No rows have `deleted_at` set (export filters these)

## CLI

```bash
# Export
fazt app export <app> [--output path.cart]

# Import
fazt app import <file> [--mode overwrite|skip|merge] [--name new-name]

# Inspect without importing
fazt app cartridge info <file>
# → Shows: app name, size, row counts, export date, schema version
```

## JS API

None. Cartridge operations are admin-only, not available to apps.

## For LLM Agents

Cartridge is how you move app data between nodes:

```bash
# Backup before risky operation
fazt app export my-app --output backup.cart

# Migrate to new node
scp my-app.cart newnode:
ssh newnode fazt app import my-app.cart

# Clone for testing
fazt app export prod-app -o test.cart
fazt app import test.cart --name test-app
```

The key insight: Provenance (`app_id` on every row) makes this trivial.
Without provenance, you'd need complex logic to find all app data.
With provenance, it's just `WHERE app_id = ?`.
