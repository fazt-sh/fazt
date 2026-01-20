# /fazt-app - Build Fazt Apps

Build and deploy apps to fazt instances using CLI scaffolding.

## Usage

```
/fazt-app <description>
/fazt-app "build a pomodoro tracker"
```

---

## Workflow

### 1. Scaffold

```bash
fazt app create <name> --template vue-api
```

Creates a working app with Vue + Vite + Tailwind + serverless API.

### 2. Customize

**UI**: Edit `src/main.js` - the Vue app component with template.

**API**: Edit `api/main.js` - rename collection, add endpoints.

**Helpers available**:
- `src/lib/api.js` - fetch wrapper (get/post/put/delete)
- `src/lib/session.js` - session ID from localStorage
- `src/lib/settings.js` - user settings persistence

### 3. Validate

```bash
fazt app validate ./<name>
```

Checks manifest, JS syntax, and structure before deploy.

### 4. Test Locally

```bash
fazt app deploy ./<name> --to local
fazt app logs <name> --peer local -f
```

Access at: `http://<name>.192.168.64.3.nip.io:8080`

Debug endpoints:
- `/_fazt/info` - app metadata
- `/_fazt/storage` - storage contents
- `/_fazt/errors` - recent errors

### 5. Deploy to Production

```bash
fazt app deploy ./<name> --to zyt
```

Access at: `https://<name>.zyt.app`

---

## Storage Quick Reference

```javascript
// In api/*.js files
var ds = fazt.storage.ds

// Create
ds.insert('items', { name: 'x' })

// Read
ds.find('items', {})                    // all
ds.find('items', { status: 'active' })  // filtered
ds.findOne('items', { id: '...' })      // single

// Update
ds.update('items', { id: '...' }, { name: 'y' })

// Delete
ds.delete('items', { id: '...' })
```

**Other storage**:
- `fazt.storage.kv` - key-value (set/get/delete/list)
- `fazt.storage.s3` - blobs (put/get/delete/list)

**ID generation** (in template):
```javascript
genId()  // returns "mkme9b3m..." style IDs
```

---

## Design Notes

- **Clean**: Generous whitespace, neutral colors
- **Responsive**: Works on mobile and desktop
- **Simple**: One component per file when splitting

---

## Location Behavior

| Context | App Location |
|---------|--------------|
| In fazt repo | `servers/zyt/<name>/` |
| Elsewhere | `/tmp/fazt-<name>/` |

---

## Instructions

When user invokes `/fazt-app`:

1. Parse description to understand what to build
2. Scaffold: `fazt app create <name> --template vue-api`
3. Customize the scaffolded files for the specific use case
4. Validate: `fazt app validate ./<name>`
5. Deploy to local: `fazt app deploy ./<name> --to local`
6. Verify it works, check logs if issues
7. Deploy to production: `fazt app deploy ./<name> --to zyt`
8. Report URL: `https://<name>.zyt.app`
