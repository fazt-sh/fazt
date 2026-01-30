# Plan 28: SPA Routing Support

## Problem

BFBB apps need clean URLs (`/dashboard`) instead of hash URLs (`/#/dashboard`),
while preserving static-hostability of source code.

**Current state:**
- Apps use hash routing for universal compatibility
- Hash URLs work everywhere but are visually cluttered
- No way to get clean URLs without breaking static-hostability

## Solution

Build-time switch that defaults to hash routing (static-hostable) but can
produce history routing (clean URLs) when deploying to fazt.

## Design Principles

1. **Source always static-hostable** - Default hash routing works anywhere
2. **Clean URLs on fazt** - `--spa` flag enables history routing
3. **No config burden** - LLM agents (/fazt-app) handle the flag automatically
4. **Explicit over magic** - `spa: true` in manifest is readable

---

## Part 1: Server-Side (Fazt)

### 1.1 Manifest Schema

```json
{
  "name": "my-app",
  "spa": true
}
```

### 1.2 Handler Changes

**File:** `internal/hosting/handler.go`

Add SPA fallback logic after existing file resolution:

```go
// After existing 404 logic:
// If spa mode enabled and path has no extension, serve index.html
if notFound && app.SPA && !hasExtension(path) {
    file, err := fs.ReadFile(siteID, "index.html")
    if err == nil {
        // serve index.html
    }
}
```

**Logic:**
1. Try exact path → serve if found
2. Try path/index.html → serve if found
3. If not found AND `spa: true` AND path has no extension → serve /index.html
4. Else → 404

### 1.3 App Metadata

**File:** `internal/hosting/manager.go` or app loading

Read `spa` field from manifest when loading app metadata.

```go
type AppManifest struct {
    Name string `json:"name"`
    SPA  bool   `json:"spa"`
}
```

---

## Part 2: Deploy Command

### 2.1 New Flag

**File:** `cmd/server/app.go` (or wherever deploy is)

```bash
fazt app deploy ./my-app --to zyt --spa
```

### 2.2 Deploy Flow with --spa

```
fazt app deploy ./my-app --to zyt --spa
    │
    ├── 1. Set VITE_SPA_ROUTING=true in environment
    │
    ├── 2. Run build (npm run build)
    │      └── Vite replaces import.meta.env.VITE_SPA_ROUTING with true
    │
    ├── 3. Read/create manifest.json in dist/
    │      └── Add "spa": true
    │
    └── 4. Upload to peer
```

### 2.3 Manifest Injection

After build, before upload:

```go
// Read existing manifest or create new
manifest := readManifest(distDir) // or create default

// Add spa flag
if spaFlag {
    manifest.SPA = true
}

// Write back
writeManifest(distDir, manifest)
```

---

## Part 3: /fazt-app Skill Updates

**Location:** `~/.claude/skills/fazt-app/`

### 3.1 SKILL.md

Add routing section and update deploy commands:

```markdown
## Routing Mode (Clean URLs)

BFBB apps use hash routing by default (source remains static-hostable). When
deploying to fazt, the `--spa` flag enables clean URLs.

| Mode | URLs | Use Case |
|------|------|----------|
| Default | `/#/dashboard` | Source served directly, any static server |
| SPA | `/dashboard` | Deployed to fazt with `--spa` flag |

**Always use `--spa` when deploying to fazt.**
```

Update Quick Reference section:
```markdown
### Essential Commands

```bash
fazt remote list                              # List peers
fazt @<peer> auth providers                   # Check OAuth status
fazt app deploy ./my-app --to local --spa     # Deploy to local
fazt app deploy ./my-app --to <peer> --spa    # Deploy to prod
```
```

### 3.2 references/frontend-patterns.md

Replace router setup with env-aware version:

```javascript
// src/main.js
import { createApp } from 'vue'
import { createRouter, createWebHistory, createWebHashHistory } from 'vue-router'
import App from './App.vue'

const router = createRouter({
  // Hash routing for static hosting, history for fazt --spa deploy
  history: import.meta.env.VITE_SPA_ROUTING
    ? createWebHistory()
    : createWebHashHistory(),
  routes: [
    { path: '/', component: () => import('./pages/Home.vue') },
    { path: '/dashboard', component: () => import('./pages/Dashboard.vue') },
  ]
})

createApp(App).use(router).mount('#app')
```

Add note:
```markdown
**Routing:** Uses hash routing by default for static-hostability. Deploy with
`fazt app deploy --spa` for clean URLs.
```

### 3.3 fazt/hosting-quirks.md

Replace "Hash Routing Requirement" section:

```markdown
## Routing Modes

### Hash Routing (Default)

Source code uses hash routing for universal static-hostability:

```javascript
history: createWebHashHistory()  // URLs: /#/dashboard
```

- Works on ANY static server (nginx, S3, GitHub Pages, etc.)
- No server configuration needed
- Source can be served directly without build

### Clean URLs (--spa flag)

When deploying to fazt with `--spa`, clean URLs are enabled:

```javascript
// Build-time switch via environment variable
history: import.meta.env.VITE_SPA_ROUTING
  ? createWebHistory()    // URLs: /dashboard
  : createWebHashHistory()
```

```bash
fazt app deploy ./my-app --to zyt --spa
```

This:
1. Sets `VITE_SPA_ROUTING=true` during build
2. Adds `spa: true` to manifest.json
3. Fazt serves index.html for all non-file routes

**BFBB preserved:** Source still works on any static server (hash routing).
Built output for fazt gets clean URLs.
```

### 3.4 fazt/deployment.md

Update deploy examples:

```bash
# Local deployment with clean URLs
fazt app deploy ./my-app --to local --spa

# Production with clean URLs
fazt app deploy ./my-app --to <remote-peer> --spa

# Skip clean URLs (hash routing in output)
fazt app deploy ./my-app --to local
```

Add to "What Happens" section:

```markdown
### With --spa flag

Additional steps:
1. Sets `VITE_SPA_ROUTING=true` environment variable
2. Build produces history-routing code
3. Injects `"spa": true` into manifest.json
4. Server returns index.html for non-file routes
```

### 3.5 references/auth-integration.md

Update OAuth redirect gotcha - simpler with clean URLs:

```markdown
### 1. OAuth Redirect (Simplified with --spa)

With clean URLs (`--spa` deploy), redirects are straightforward:

```javascript
// Works correctly with --spa
login(window.location.href)  // e.g., https://myapp.zyt.app/dashboard
```

**Without --spa (hash routing)**, include the hash:
```javascript
const redirect = location.origin + '/#/dashboard'
```
```

Update checklist:
```markdown
### Quick Checklist

- [ ] Deploy with `--spa` flag for clean URLs
- [ ] OAuth redirect uses `location.href` (absolute URL)
- [ ] Logout uses `fetch('/auth/logout', { method: 'POST' })`
- [ ] Router uses env-aware history mode (see frontend-patterns.md)
```

### 3.6 fazt/cli-app.md

Add --spa flag documentation:

```markdown
### fazt app deploy

Deploy an app to a peer.

```bash
fazt app deploy <directory> --to <peer> [flags]
```

**Flags:**
| Flag | Description |
|------|-------------|
| `--to` | Target peer (required) |
| `--spa` | Enable clean URLs (history routing) |
| `--no-build` | Skip build step, deploy as-is |

**Examples:**
```bash
# Standard deploy with clean URLs
fazt app deploy ./my-app --to zyt --spa

# Deploy pre-built directory
fazt app deploy ./my-app/dist --to zyt --no-build
```
```

---

## Part 4: Project Documentation

### 4.1 CLAUDE.md

Add to App Structure section:

```markdown
### manifest.json Fields

| Field | Required | Description |
|-------|----------|-------------|
| name | Yes | App identifier (becomes subdomain) |
| spa | No | Enable SPA fallback routing for clean URLs |
```

### 4.2 koder/scratch/04_bfbb.md

Add routing section to capture the pattern:

```markdown
## Routing (Clean URLs)

BFBB apps default to hash routing (static-hostable). Clean URLs enabled via:

```bash
fazt app deploy ./my-app --to zyt --spa
```

Build-time env var `VITE_SPA_ROUTING` switches router mode.
```

---

## Implementation Order

1. **Server: Read spa from manifest** - Load spa field when resolving app
2. **Server: SPA fallback logic** - Handler serves index.html for clean URLs
3. **Deploy: --spa flag** - Accept flag, set env var during build
4. **Deploy: Manifest injection** - Add spa:true to manifest when flag used
5. **Skill: Router template** - Update with env var check
6. **Skill: Deploy commands** - Add --spa to all examples
7. **Docs: Update** - hosting-quirks.md, auth-integration.md

---

## Testing

### Server-side

```bash
# Deploy with spa
fazt app deploy ./test-spa --to local --spa

# Test clean URL access
curl http://test-spa.192.168.64.3.nip.io:8080/dashboard
# Should return index.html content, not 404

# Test file access still works
curl http://test-spa.192.168.64.3.nip.io:8080/static/app.js
# Should return JS file
```

### Build-time

```bash
# Without flag - hash routing
npm run build
# Check dist/assets/*.js for createWebHashHistory

# With flag - history routing
VITE_SPA_ROUTING=true npm run build
# Check dist/assets/*.js for createWebHistory
```

---

## Edge Cases

| Case | Behavior |
|------|----------|
| `/api/something` | API routes handled before static, unaffected |
| `/favicon.ico` | Has extension, normal file lookup |
| `/dashboard` | No extension, spa fallback to index.html |
| `/assets/app.js` | Has extension, normal file lookup |
| Missing index.html | 404 (can't fallback to nothing) |

---

## Rollback

If issues arise:
- Remove `--spa` flag from deploy commands
- Apps revert to hash routing
- No server changes needed (spa:false is default)
