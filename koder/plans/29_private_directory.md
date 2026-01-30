# Plan 29: Private Directory for Serverless Data

## Problem

Apps often need data files that:
- Should NOT be publicly accessible via HTTP
- Should be readable by serverless functions
- Can be versioned with the app (committed to git)
- Enable rapid iteration without database setup

**Current state:**
- All deployed files are publicly accessible
- Data must go in database (storage APIs) or be exposed
- No way to bundle private config/data with app

## Solution

Reserved `private/` directory:
- Uploaded during deploy (stored in VFS)
- Blocked from HTTP access
- Readable by serverless via `fazt.private.*` API

## Use Cases

| Use Case | Example | Why Private? |
|----------|---------|--------------|
| Seed data | `private/users.json` | Bootstrap data, not API response |
| Config | `private/config.json` | Runtime settings, may have secrets |
| Mock data | `private/products.json` | Rapid iteration, versioned with code |
| Fixtures | `private/scenarios.json` | Test data, not for users |
| Static lookup | `private/countries.json` | Reference data, filtered by API |

---

## Part 1: Directory Structure

```
my-app/
├── index.html
├── manifest.json
├── api/
│   └── main.js           # Serverless handler
├── private/              # NEW: Not served via HTTP
│   ├── data.json
│   ├── users.json
│   └── config.json
└── public/               # Served at root
    └── version.json
```

### Reserved Paths

| Path | HTTP Access | Serverless Access |
|------|-------------|-------------------|
| `/api/*` | Executes serverless | N/A |
| `/private/*` | 403 Forbidden | `fazt.private.read()` |
| `/*` | Served as static | N/A |

---

## Part 2: HTTP Access (Auth-Gated Streaming)

### 2.1 Design Rationale

Private files have **two access modes**:

| Access Method | Use Case | Behavior |
|---------------|----------|----------|
| HTTP `GET /private/*` | Serve file to user | Auth check → stream directly |
| Serverless `fazt.private.*` | Process data in code | Read into JS runtime |

This enables:
- 100MB video served to authenticated users (HTTP streaming, no serverless)
- 8MB JSON processed by serverless logic (loaded into JS)
- Same directory, complementary access patterns

### 2.2 Request Handling

**File:** `cmd/server/main.go` (siteHandler)

Auth-gated serving (not blocking):

```go
// Auth-gated private directory access
if strings.HasPrefix(r.URL.Path, "/private/") || r.URL.Path == "/private" {
    // Check authentication
    user, err := authProvider.GetSessionFromRequest(r)
    if err != nil || user == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    // Stream file directly to authenticated user
    servePrivateFile(w, r, siteID)
    return
}
```

### 2.3 File Type Handling

| File Type | HTTP | Serverless |
|-----------|------|------------|
| Large binary (video, audio) | ✅ Stream | ❌ Too large |
| Images | ✅ Stream | ⚠️ Possible but no use case |
| JSON < 10MB | ✅ Stream | ✅ `readJSON()` |
| YAML < 10MB | ✅ Stream | ⚠️ Needs YAML lib |
| SQLite < 10MB | ✅ Download | ⚠️ Needs sql.js lib |

**Key insight**: HTTP = delivery, Serverless = processing.

---

## Part 3: Serverless API

### 3.1 New Namespace

**File:** `internal/runtime/private_bindings.go` (new)

```go
func InjectPrivateNamespace(vm *goja.Runtime, appID string, db *sql.DB) error {
    private := vm.NewObject()

    // fazt.private.read(path) -> string
    private.Set("read", func(path string) string {
        content, err := loadPrivateFile(db, appID, path)
        if err != nil {
            return ""
        }
        return content
    })

    // fazt.private.readJSON(path) -> object
    private.Set("readJSON", func(path string) interface{} {
        content, err := loadPrivateFile(db, appID, path)
        if err != nil {
            return nil
        }
        var data interface{}
        json.Unmarshal([]byte(content), &data)
        return data
    })

    // fazt.private.exists(path) -> bool
    private.Set("exists", func(path string) bool {
        _, err := loadPrivateFile(db, appID, path)
        return err == nil
    })

    // fazt.private.list() -> []string
    private.Set("list", func() []string {
        return listPrivateFiles(db, appID)
    })

    fazt := vm.Get("fazt").(*goja.Object)
    fazt.Set("private", private)
    return nil
}

func loadPrivateFile(db *sql.DB, appID, path string) (string, error) {
    fullPath := "private/" + strings.TrimPrefix(path, "/")
    var content string
    err := db.QueryRow(`
        SELECT content FROM files
        WHERE site_id = ? AND path = ?
    `, appID, fullPath).Scan(&content)
    return content, err
}
```

### 3.2 API Surface

```javascript
// api/main.js

// Read raw file content
var configStr = fazt.private.read('config.json')

// Read and parse JSON
var config = fazt.private.readJSON('config.json')
var users = fazt.private.readJSON('users.json')

// Check if file exists
if (fazt.private.exists('feature-flags.json')) {
    var flags = fazt.private.readJSON('feature-flags.json')
}

// List all private files
var files = fazt.private.list()
// ['config.json', 'users.json', 'data/products.json']
```

---

## Part 4: Deploy Handling

### 4.1 Gitignore-Aware Deployment

Private files may be gitignored (not in repo) but still need deployment.

**Deploy behavior:**

| `private/` state | `--include-private` | Behavior |
|------------------|---------------------|----------|
| Not gitignored | - | Deploy normally |
| Gitignored | No | Warn + skip |
| Gitignored | Yes | Info + include |

```bash
# Warning when private/ is gitignored
$ fazt app deploy ./my-app --to zyt
Warning: private/ is gitignored but exists
  Use --include-private to deploy private files
  Skipping private/...

# Explicit include
$ fazt app deploy ./my-app --to zyt --include-private
Including gitignored private/ (5 files)
```

### 4.2 Implementation

**File:** `cmd/server/main.go` - `createDeployZipWithOptions()`

```go
type DeployZipOptions struct {
    IncludePrivate bool
}

type DeployZipResult struct {
    Buffer            *bytes.Buffer
    FileCount         int
    PrivateExists     bool
    PrivateGitignored bool
    PrivateIncluded   bool
    PrivateFileCount  int
}
```

### 4.3 Storage

Private files stored in VFS like any other file:
- Path: `private/data.json`
- Stored normally in `files` table
- HTTP access is auth-gated (not blocked)

---

## Part 5: Mock REST API (Future Enhancement)

### 5.1 Concept

Auto-generate CRUD endpoints from JSON files in `private/`.

If `private/users.json` contains:
```json
[
  {"id": 1, "name": "Alice", "email": "alice@example.com"},
  {"id": 2, "name": "Bob", "email": "bob@example.com"}
]
```

Fazt auto-generates:

| Endpoint | Method | Behavior |
|----------|--------|----------|
| `/api/users` | GET | List all users |
| `/api/users/1` | GET | Get user by id |
| `/api/users` | POST | Add user (returns new with id) |
| `/api/users/1` | PUT | Update user |
| `/api/users/1` | DELETE | Remove user |

### 5.2 Enabling

Via manifest:
```json
{
  "name": "my-app",
  "mockApi": {
    "users": "private/users.json",
    "products": "private/products.json"
  }
}
```

Or convention: any `private/*.json` array gets endpoints.

### 5.3 Persistence Options

| Mode | Behavior |
|------|----------|
| Ephemeral (default) | Changes lost on restart |
| Session-scoped | Changes persist per session |
| Persistent | Updates written back to VFS |

**Recommendation:** Start with ephemeral - safest for PoCs.

### 5.4 Implementation Complexity

This is a significant feature. Suggest:
1. Implement basic `private/` directory first (this plan)
2. Mock REST API as separate plan (Plan 30?)

---

## Implementation Order

1. **Auth-gated HTTP** - Check auth, stream file for `/private/*`
2. **Private bindings** - `fazt.private.read/readJSON/exists/list`
3. **Inject in handler** - Add to serverless execution
4. **Documentation** - Update /fazt-app skill
5. **Testing** - Verify auth-gated serving, serverless works

---

## /fazt-app Skill Updates

### serverless-api.md

Add new section:

```markdown
## Private Files (fazt.private)

Read files from the `private/` directory. These files are:
- Uploaded with your app
- NOT accessible via HTTP
- Only readable by serverless functions

### API

```javascript
// Read as string
var content = fazt.private.read('config.json')

// Read and parse JSON
var users = fazt.private.readJSON('users.json')

// Check existence
if (fazt.private.exists('feature-flags.json')) { ... }

// List all private files
var files = fazt.private.list()
```

### Use Cases

- **Seed data**: Bundle initial data with app
- **Config**: Runtime settings without exposing
- **Mock data**: Rapid iteration, versioned with code
- **Lookup tables**: Countries, categories, etc.
```

### frontend-patterns.md

Add note about project structure:

```markdown
## Project Structure

```
my-app/
├── src/                 # Vue source
├── api/
│   └── main.js          # Serverless
├── private/             # Server-only data (not HTTP accessible)
│   └── seed-data.json
└── public/              # Static files (HTTP accessible)
    └── version.json
```
```

---

## Testing

### HTTP Blocking

```bash
# Should return 403
curl -I http://my-app.192.168.64.3.nip.io:8080/private/data.json
# HTTP/1.1 403 Forbidden

# Should return 403 for nested paths too
curl -I http://my-app.192.168.64.3.nip.io:8080/private/nested/file.json
# HTTP/1.1 403 Forbidden
```

### Serverless Access

```javascript
// api/main.js - test endpoint
if (request.path === '/api/test-private') {
    var data = fazt.private.readJSON('test.json')
    respond({ success: true, data: data })
}
```

```bash
curl http://my-app.192.168.64.3.nip.io:8080/api/test-private
# {"success": true, "data": {...}}
```

---

## Security Considerations

| Concern | Mitigation |
|---------|------------|
| Path traversal | Normalize path, block `..` |
| Large files | Size limit (e.g., 1MB per file) |
| Binary files | Allow but document as text-focused |
| Secrets in private/ | Warn in docs - use env vars for real secrets |

### Path Sanitization

```go
func loadPrivateFile(db *sql.DB, appID, path string) (string, error) {
    // Prevent traversal
    path = filepath.Clean(path)
    if strings.Contains(path, "..") {
        return "", errors.New("invalid path")
    }

    fullPath := "private/" + strings.TrimPrefix(path, "/")
    // ... load from DB
}
```

---

## Future Enhancements

1. **Mock REST API** (Plan 30) - Auto-generate CRUD from JSON
2. **Binary support** - `fazt.private.readBinary()` returning base64
3. **Write support** - `fazt.private.write()` for persistent changes
4. **Encryption** - Encrypt private files at rest

### JS Library Expansion

Serverless capabilities for private files can be extended with bundled JS libraries:

| Library | Enables | Use Case |
|---------|---------|----------|
| YAML parser | `fazt.private.readYAML()` | Config files in YAML format |
| sql.js | Query SQLite files | `private/app.db` as read-only data source |
| CSV parser | `fazt.private.readCSV()` | Spreadsheet data processing |

These would work with files < 10MB loaded into JS runtime. Implementation would:
1. Bundle the library into Goja runtime
2. Add corresponding `fazt.private.*` method
3. Parse binary/text data using the library

**Note**: HTTP streaming works for ANY file regardless of library support.
Serverless processing is the optional enhancement for data manipulation.
