# Hosting Quirks

Important considerations when deploying to fazt's static hosting.

## Static Files vs API Endpoints

### When to Use Static Files

| Use Case | Static File | API Endpoint |
|----------|-------------|--------------|
| Version info | ✅ `/version.json` | ❌ Cold start delays |
| Config/settings | ✅ `config.json` | ❌ Unnecessary overhead |
| Infrequently changing data | ✅ Pre-generated JSON | ❌ |
| User-specific data | ❌ | ✅ Needs auth context |
| Real-time data | ❌ | ✅ Needs fresh fetch |
| Form submissions | ❌ | ✅ Needs processing |

**Rule of thumb**: If data is the same for all users and changes only at deploy time, use a static file.

### Static File Advantages

- **No cold start**: Served immediately from CDN/disk
- **Cacheable**: Browser and CDN can cache
- **Works offline**: Available without network (with service worker)
- **Simpler**: No serverless runtime overhead

### Example: Version Endpoint

```javascript
// BAD - API endpoint with cold start issues
// api/version.js
respond({ version: '1.0.0', build: 'abc123' })

// GOOD - Static file, generated at build time
// Build script creates /version.json
{
  "version": "1.0.0",
  "build": "abc123",
  "timestamp": "2026-01-29T12:00:00Z"
}

// Frontend fetches static file
const res = await fetch('/version.json')
const { version, build } = await res.json()
```

---

## Client-Side Routing

Fazt is static hosting by default - it serves files as-is. For SPAs with
client-side routing, you have two options.

### Option 1: Hash Routing (Default)

Hash routing works everywhere without server config:

```javascript
import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(),  // /#/dashboard instead of /dashboard
  routes: [...]
})
```

URLs become `https://myapp.example.com/#/dashboard` - the `#` portion is
handled client-side and never sent to the server.

**OAuth + Hash Routing**: Include the hash in redirects:

```javascript
const redirect = location.origin + '/#/dashboard'
location.href = '/auth/login?redirect=' + encodeURIComponent(redirect)
```

### Option 2: SPA Mode (Clean URLs)

Deploy with `--spa` for clean URLs without hashes:

```bash
fazt app deploy ./my-app --to <peer> --spa
```

With SPA enabled, fazt serves `index.html` for any URL that doesn't match a
real file. Routes like `/dashboard` work directly.

**Trailing slashes**: Fazt normalizes `/about/` → `/about` (301 redirect) for
SEO consistency.

### BFBB Pattern: Build-Free But Buildable

The recommended approach for apps that need both local dev and SPA deployment:

**Source code** uses hash routing (works with simple HTTP servers):
```javascript
// main.js - default to hash routing
const useSPA = import.meta.env.VITE_SPA_ROUTING === 'true'

const router = createRouter({
  history: useSPA ? createWebHistory() : createWebHashHistory(),
  routes: [...]
})
```

**Development**: Use `npm run dev` or any static server - hash routing works.

**Production**: Deploy with `--spa` - build gets `VITE_SPA_ROUTING=true`:
```bash
fazt app deploy ./my-app --to zyt --spa
```

This gives you:
- Zero-config local development (no server routing needed)
- Clean URLs in production (server handles SPA fallback)
- Same codebase, different routing based on environment

---

## Serverless API Limitations

Fazt's serverless runtime has constraints compared to full Node.js.

### No Node.js Built-ins

```javascript
// These DON'T work in api/main.js
const fs = require('fs')        // ❌ No filesystem
const path = require('path')    // ❌ No path module
import { readFile } from 'fs'   // ❌ No ES module imports

// Use fazt's built-in storage instead
fazt.storage.kv.get('key')      // ✅ Key-value store
fazt.storage.ds.find('col', {}) // ✅ Document store
```

### Cold Start Timeouts

First request after idle period may timeout.

```javascript
// Frontend: Add retry logic
async function fetchWithRetry(url, options = {}, retries = 3) {
  for (let i = 0; i < retries; i++) {
    try {
      const controller = new AbortController()
      const timeout = setTimeout(() => controller.abort(), 3000)

      const res = await fetch(url, { ...options, signal: controller.signal })
      clearTimeout(timeout)

      if (res.ok) return res
    } catch (e) {
      if (i === retries - 1) throw e
      await new Promise(r => setTimeout(r, 500 * (i + 1)))
    }
  }
}
```

### Execution Time Limit

Long-running operations will timeout. Keep API handlers fast.

```javascript
// BAD - too much work in single request
if (action === 'process-all') {
  const items = ds.find('items', {})  // Could be thousands
  items.forEach(item => heavyProcessing(item))  // Will timeout
  respond({ done: true })
}

// GOOD - paginate or use background jobs
if (action === 'process-batch') {
  const items = ds.find('items', { processed: false }, { limit: 10 })
  items.forEach(item => {
    lightProcessing(item)
    ds.update('items', { id: item.id }, { processed: true })
  })
  respond({ processed: items.length, hasMore: items.length === 10 })
}
```

---

## Build Artifacts & Git

### What to Gitignore

```gitignore
# Build output
dist/
build/

# Generated at build time
version.json
public/version.json
*.build

# Dependencies
node_modules/

# Environment
.env
.env.local
```

### What Gets Deployed vs Committed

| File | Committed | Deployed |
|------|-----------|----------|
| `src/**` | ✅ | ❌ (built to dist) |
| `dist/**` | ❌ | ✅ |
| `version.json` | ❌ | ✅ (generated at build) |
| `api/main.js` | ✅ | ✅ (copied to dist) |
| `node_modules/` | ❌ | ❌ |

### Build Script Pattern

```json
{
  "scripts": {
    "build": "node scripts/version.js && vite build && cp -r api dist/ && cp version.json dist/"
  }
}
```

Order matters:
1. Generate `version.json` (captures git hash before build)
2. Run vite build (creates `dist/`)
3. Copy `api/` folder to `dist/api/`
4. Copy `version.json` to `dist/`

---

## Summary

- Prefer static files for data that doesn't change per-request
- Use hash routing (`createWebHashHistory`) by default for client-side routing
- Deploy with `--spa` for clean URLs (serves index.html for unknown routes)
- Use BFBB pattern: hash routing in dev, SPA mode in prod
- Serverless has no Node.js built-ins - use fazt.storage
- Add retry logic for cold start timeouts
- Gitignore generated files but ensure they're copied to dist
