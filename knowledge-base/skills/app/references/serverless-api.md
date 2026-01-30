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
request.body        // Parsed JSON body (POST/PUT)
request.headers     // Request headers (lowercase keys)
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

## Storage APIs

### Document Store (fazt.storage.ds)

Primary storage for structured data.

```javascript
var ds = fazt.storage.ds

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

### Key-Value Store (fazt.storage.kv)

Simple key-value lookups, caches, counters.

```javascript
var kv = fazt.storage.kv

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

### Blob Storage (fazt.storage.s3)

File/binary storage.

```javascript
var s3 = fazt.storage.s3

// Store file
s3.put('uploads/image.png', binaryData, 'image/png')

// Retrieve
var file = s3.get('uploads/image.png')

// Delete
s3.delete('uploads/image.png')

// List by prefix
var files = s3.list('uploads/')
```

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
- **No network calls** - Can't fetch external URLs
- **ES5 syntax** - Use `var`, not `let`/`const`
