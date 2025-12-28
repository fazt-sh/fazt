# Kernel-Level Row Security (RLS)

## Summary

The kernel automatically filters all storage queries by `app_uuid` and
`user_id`. Apps get multi-tenancy without writing security checks.

## How It Works

```javascript
// App code (simple)
const posts = await fazt.storage.ds.find('posts', {});

// Kernel rewrites to:
// SELECT * FROM posts WHERE app_uuid = ? AND user_id = ?
```

## Automatic Scoping

### App Isolation

Every query is scoped to the current app:

```sql
-- What app writes
INSERT INTO posts (title) VALUES ('Hello');

-- What kernel executes
INSERT INTO posts (app_uuid, title) VALUES ('app_x9z2k', 'Hello');
```

### User Isolation (with auth)

If the app has authenticated users:

```sql
-- What app writes
SELECT * FROM notes;

-- What kernel executes (if user logged in)
SELECT * FROM notes WHERE app_uuid = ? AND user_id = ?;
```

## Benefits

1. **Zero Security Code**: Apps can't accidentally leak data
2. **Simple Mental Model**: "My queries only see my data"
3. **Performance**: Indexes on (app_uuid, user_id) are automatic

## Bypass (Privileged)

System apps can bypass RLS with permission:

```json
{
  "permissions": ["storage:bypass-rls"]
}
```

```javascript
const allPosts = await fazt.storage.ds.find('posts', {}, {
    bypassRLS: true
});
```
