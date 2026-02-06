# Serverless API Reference

Fazt runs JavaScript serverless functions via the Goja runtime.

## File Location

```
my-app/
└── api/
    └── main.js    # Handles all /api/* requests
```

## Request Object

```javascript
// Available in every request
request.method      // "GET", "POST", "PUT", "DELETE"
request.path        // "/api/items/123"
request.query       // { session: "cat-blue-river", limit: "10" }
request.body        // Parsed JSON or form fields (POST/PUT)
request.headers     // Request headers (lowercase keys)
request.files       // Uploaded files (multipart/form-data only)
```

## Response Function

```javascript
// Basic response
respond({ data: "value" })

// With status code
respond(201, { id: "new_item" })
respond(404, { error: "Not found" })

// Headers (optional third argument)
respond(200, data, { "X-Custom": "value" })
```

## File Uploads

Apps can receive file uploads via `multipart/form-data`. Uploaded files appear on `request.files`.

### HTML Form

```html
<form action="/api/upload" method="POST" enctype="multipart/form-data">
  <input type="file" name="photo" />
  <input type="text" name="caption" value="My photo" />
  <button>Upload</button>
</form>
```

### Handler

```javascript
// api/main.js
if (request.method === 'POST' && request.files) {
  fazt.auth.requireLogin()
  var file = request.files.photo  // { name, type, size, data }
  var caption = request.body.caption  // "My photo"

  // Store in user-scoped storage (isolated per user)
  fazt.app.user.s3.put('uploads/' + file.name, file.data, file.type)
  respond({ stored: file.name, size: file.size })
}
```

### File Object

Each entry in `request.files` has:

| Property | Type | Description |
|----------|------|-------------|
| `name` | string | Original filename (`"photo.jpg"`) |
| `type` | string | MIME type (`"image/jpeg"`) |
| `size` | number | Byte count |
| `data` | ArrayBuffer | Binary data (pass directly to `s3.put()`) |

Non-file form fields go into `request.body` as a key-value map.

### Storage Scoping

| API | Isolation | Use for |
|-----|-----------|---------|
| `fazt.app.user.s3` | Per-user | User uploads (profile photos, documents) |
| `fazt.app.s3` | Shared | App-wide assets (public uploads, admin content) |

Use `fazt.app.user.s3` for user uploads — files are automatically isolated so users can't access each other's data.

### Limits

Upload size is controlled by `system.Limits.Storage.MaxUpload` (default ~10MB, max 100MB depending on server RAM).

## Storage APIs

Two namespaces available:
- `fazt.app.*` - Shared app storage (all users see same data)
- `fazt.app.user.*` - User-scoped storage (automatic isolation per user)

### Document Store (fazt.app.ds)

Primary storage for structured data.

```javascript
var ds = fazt.app.ds

// Insert
ds.insert('items', { id: 'abc', name: 'Test', value: 42 })

// Find all matching
var items = ds.find('items', { userId: 'user_123' })

// Find one
var item = ds.findOne('items', { id: 'abc' })

// Update (query, changes)
ds.update('items', { id: 'abc' }, { value: 100 })

// Delete
ds.delete('items', { id: 'abc' })
```

### Key-Value Store (fazt.app.kv)

Simple key-value lookups, caches, counters.

```javascript
var kv = fazt.app.kv

// Set
kv.set('user:123:token', 'abc123')

// Set with TTL (milliseconds)
kv.set('cache:result', data, 60000)  // 1 minute

// Get
var value = kv.get('user:123:token')

// Delete
kv.delete('user:123:token')

// List by prefix
var keys = kv.list('user:123:')
```

### Blob Storage (fazt.app.s3)

File/binary storage.

```javascript
var s3 = fazt.app.s3

// Store file
s3.put('uploads/image.png', binaryData, 'image/png')

// Retrieve
var file = s3.get('uploads/image.png')

// Delete
s3.delete('uploads/image.png')

// List by prefix
var files = s3.list('uploads/')
```

### User-Scoped Storage (fazt.app.user.*)

Storage automatically isolated per authenticated user. Requires login.

```javascript
// User's private key-value store
fazt.app.user.kv.set('preferences', { theme: 'dark' })
var prefs = fazt.app.user.kv.get('preferences')

// User's private documents
fazt.app.user.ds.insert('notes', { title: 'My Note' })
var myNotes = fazt.app.user.ds.find('notes', {})

// User's private files
fazt.app.user.s3.put('avatar.png', imageData, 'image/png')
```

**Key difference**: With `fazt.app.user.*`, you don't need to include user ID in queries - isolation is automatic. Two users calling `fazt.app.user.kv.set('key', 'value')` will each have their own separate `key`.

```javascript
// Old pattern (manual scoping)
var userId = fazt.auth.getUser().id
fazt.app.ds.insert('notes', { userId: userId, title: 'Note' })
fazt.app.ds.find('notes', { userId: userId })

// New pattern (automatic scoping)
fazt.app.user.ds.insert('notes', { title: 'Note' })
fazt.app.user.ds.find('notes', {})  // Only returns this user's notes
```

### Legacy: fazt.storage.* (deprecated)

The old `fazt.storage.kv/ds/s3` namespace still works but is deprecated. Use `fazt.app.*` instead.

## Authentication APIs

### fazt.auth.getUser()

Get current authenticated user (null if not logged in).

```javascript
var user = fazt.auth.getUser()
// Returns: { id, email, name, picture, role, provider }
// Or null if not authenticated
```

### fazt.auth.isLoggedIn()

Check if user is authenticated.

```javascript
if (fazt.auth.isLoggedIn()) {
  // User is logged in
}
```

### fazt.auth.isOwner() / isAdmin()

Check user roles.

```javascript
fazt.auth.isOwner()     // role === 'owner'
fazt.auth.isAdmin()     // role === 'owner' OR 'admin'
fazt.auth.hasRole('x')  // role === 'x'
```

### fazt.auth.requireLogin()

Require authentication (redirects to login if not authenticated).

```javascript
fazt.auth.requireLogin()  // Throws redirect if not logged in
var user = fazt.auth.getUser()  // Guaranteed to exist now
```

### fazt.auth.requireAdmin() / requireOwner()

Require specific roles.

```javascript
fazt.auth.requireAdmin()  // 403 if not admin
fazt.auth.requireOwner()  // 403 if not owner
fazt.auth.requireRole('editor')  // 403 if not editor
```

### fazt.auth.getLoginURL() / getLogoutURL()

Get auth URLs for redirects.

```javascript
var loginUrl = fazt.auth.getLoginURL('/dashboard')
var logoutUrl = fazt.auth.getLogoutURL()
```

## Private Files (fazt.private)

Read files from the `private/` directory. These files have **two access modes**:

| Access | Use Case | Behavior |
|--------|----------|----------|
| HTTP `GET /private/*` | Serve to users | Auth required (401 if not logged in) |
| Serverless `fazt.private.*` | Process in code | Direct access for logic |

This enables:
- Large files (video, images) served to authenticated users via HTTP
- Small data files (JSON, config) processed by serverless logic

### File Structure

```
my-app/
├── api/main.js
├── private/           # Server-only files
│   ├── config.json
│   ├── seed-data.json
│   └── data/
│       └── users.json
└── index.html
```

### API

```javascript
// Read file as string
var content = fazt.private.read('config.json')

// Read and parse JSON
var config = fazt.private.readJSON('config.json')
var users = fazt.private.readJSON('data/users.json')

// Check if file exists
if (fazt.private.exists('feature-flags.json')) {
  var flags = fazt.private.readJSON('feature-flags.json')
}

// List all private files
var files = fazt.private.list()
// Returns: ['config.json', 'seed-data.json', 'data/users.json']
```

### Return Values

| Method | Found | Not Found |
|--------|-------|-----------|
| `read()` | string | undefined |
| `readJSON()` | object/array | null |
| `exists()` | true | false |
| `list()` | string[] | [] |

### Use Cases

| Use Case | Example |
|----------|---------|
| Seed data | `private/initial-users.json` |
| Config | `private/settings.json` |
| Mock data | `private/products.json` |
| Fixtures | `private/test-scenarios.json` |
| Lookup tables | `private/countries.json` |
| Protected media | `private/video.mp4` (via HTTP) |

### Deployment

If `private/` is gitignored, use `--include-private` to deploy:

```bash
# Warns and skips gitignored private/
fazt app deploy ./my-app --to zyt

# Explicitly includes gitignored private/
fazt app deploy ./my-app --to zyt --include-private
```

### Example: Data Seeding

```javascript
// api/main.js
var ds = fazt.storage.ds

if (request.path === '/api/seed' && request.method === 'POST') {
  var users = fazt.private.readJSON('seed-data.json')
  for (var i = 0; i < users.length; i++) {
    ds.insert('users', users[i])
  }
  respond({ seeded: users.length })
}
```

## Common Patterns

### Session-Scoped API

```javascript
var ds = fazt.storage.ds
var session = request.query.session || (request.body && request.body.session)

if (!session) {
  respond(400, { error: 'session required' })
}

// All queries scoped to session
var items = ds.find('items', { session: session })
```

### User-Scoped API

```javascript
fazt.auth.requireLogin()
var user = fazt.auth.getUser()
var ds = fazt.storage.ds

// All queries scoped to user
var items = ds.find('items', { userId: user.id })
```

### RESTful Routing

```javascript
var parts = request.path.split('/').filter(Boolean)
var resource = parts[1]  // e.g., "items"
var id = parts[2]        // e.g., "abc123"

if (resource === 'items') {
  if (request.method === 'GET' && !id) {
    respond({ items: ds.find('items', {}) })
  } else if (request.method === 'GET' && id) {
    respond(ds.findOne('items', { id: id }))
  } else if (request.method === 'POST') {
    var item = { id: genId(), ...request.body }
    ds.insert('items', item)
    respond(201, item)
  } else if (request.method === 'PUT' && id) {
    ds.update('items', { id: id }, request.body)
    respond(ds.findOne('items', { id: id }))
  } else if (request.method === 'DELETE' && id) {
    ds.delete('items', { id: id })
    respond({ ok: true })
  }
}
```

### ID Generation

```javascript
function genId() {
  return Date.now().toString(36) + Math.random().toString(36).slice(2, 8)
}
```

## Limitations

- **No async/await** - Goja is synchronous
- **No npm modules** - Built-in APIs only
- **ES5 syntax** - Use `var`, not `let`/`const`
- **Upload size** - Bounded by `system.Limits.Storage.MaxUpload` (default ~10MB)
