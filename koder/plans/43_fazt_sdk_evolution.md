# Plan 43: fazt-sdk Evolution — Universal Client for Admin + Apps

## Problem

fazt-sdk currently lives in `admin/packages/fazt-sdk/` and only wraps admin API
endpoints (apps, aliases, system, logs, etc.). It's a clean, framework-agnostic
HTTP client (737 lines, zero dependencies) — but it can't be used by fazt-apps.

Fazt-apps like Preview hand-roll their own `api.js` with raw `fetch()`, missing:
- Error handling patterns (retry, timeout, structured errors)
- File upload with progress
- Pagination helpers
- Auth flow integration
- Caching / request deduplication

Meanwhile, the admin already gets all this through the SDK. The gap widens as
more apps get built.

## Goal

One SDK that serves both admin UI and fazt-apps. Same core, different API
surfaces. Server-side credentials gate what's accessible — the SDK just
organizes the client.

## Architecture

```
admin/packages/fazt-sdk/
├── core/
│   ├── client.js          # HTTP client (exists, move here)
│   ├── types.js           # Shared types (exists, move here)
│   ├── errors.js          # Structured error classes
│   ├── upload.js          # File upload with progress + validation
│   └── paginator.js       # Offset/cursor pagination helper
├── admin.js               # Admin namespace factory (current index.js guts)
├── app.js                 # App namespace factory (NEW)
├── mock.js                # Mock adapter (exists, extend for app routes)
├── fixtures/              # Mock data (exists)
└── index.js               # Main export: createClient, createAppClient
```

### Two entry points, shared core

```javascript
// Admin UI (requires admin session)
import { createClient } from './packages/fazt-sdk/index.js'
const client = createClient()
await client.apps.list()       // admin namespace
await client.system.health()   // admin namespace

// Fazt-app (requires user session)
import { createAppClient } from './packages/fazt-sdk/index.js'
const app = createAppClient()
await app.auth.me()            // app namespace
await app.photos.list()        // custom — wait, this is app-specific...
```

### The key insight: what can fazt-sdk standardize for apps?

Apps define their own API routes in `api/main.js`. The SDK can't know about
`/api/photos` or `/api/todos` — those are app-specific. What it CAN standardize:

| Concern | SDK provides | App-specific? |
|---------|-------------|---------------|
| Auth flow | `app.auth.me()`, `app.auth.login()`, `app.auth.logout()` | No — same for all apps |
| HTTP client | `app.http.get('/api/photos')` with error handling | No — generic |
| File upload | `app.upload(file, '/api/upload')` with progress | No — generic |
| Pagination | `app.paginate('/api/photos', { limit: 50 })` | No — generic |
| Error handling | Structured errors, retry, timeout | No — generic |
| API routes | `/api/photos`, `/api/todos` | YES — app-specific |

So the app SDK is NOT about wrapping specific endpoints. It's about giving apps
a smart HTTP client with auth, uploads, pagination, and errors built in.

## Phase 1: Restructure (no new features)

Move existing code into `core/` directory. No behavior changes.

```
# Before
fazt-sdk/
├── index.js      # Factory + all admin namespaces
├── client.js     # HTTP client
├── mock.js       # Mock adapter
├── types.js      # Types
└── fixtures/

# After
fazt-sdk/
├── core/
│   ├── client.js
│   └── types.js
├── admin.js      # Extracted from index.js — admin namespace factory
├── mock.js       # Stays
├── fixtures/     # Stays
└── index.js      # Re-exports createClient (admin), createAppClient (app)
```

Admin singleton (`admin/src/client.js`) import path stays the same — index.js
re-exports everything.

## Phase 2: App client

Add `app.js` — the app-facing SDK:

```javascript
export function createAppNamespace(http) {
  return {
    auth: {
      me: () => http.get('/api/me'),
      login: () => { window.location.href = '/auth/login' },
      logout: () => http.post('/auth/logout'),
    },

    // Smart HTTP with built-in pagination
    http,

    // File upload with progress callback
    upload(file, url = '/api/upload', { field = 'photo', onProgress } = {}) {
      const form = new FormData()
      form.append(field, file)
      return http.upload(url, form, { onProgress })
    },

    // Paginated list helper
    paginate(url, { limit = 50, params = {} } = {}) {
      let offset = 0
      let done = false
      return {
        async next() {
          if (done) return { items: [], hasMore: false }
          const data = await http.get(url, {
            params: { ...params, limit, offset }
          })
          offset += limit
          done = !data.hasMore
          return data
        },
        reset() { offset = 0; done = false }
      }
    }
  }
}
```

## Phase 3: Upload with progress

Enhance `core/client.js` with XMLHttpRequest-based upload for progress:

```javascript
upload(url, formData, { onProgress } = {}) {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.open('POST', url)
    xhr.withCredentials = true
    if (onProgress) {
      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) onProgress(e.loaded / e.total)
      }
    }
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve(JSON.parse(xhr.responseText))
      } else {
        reject(new ApiError(xhr.status, xhr.responseText))
      }
    }
    xhr.onerror = () => reject(new Error('Upload failed'))
    xhr.send(formData)
  })
}
```

## Phase 4: Structured errors

```javascript
// core/errors.js
export class ApiError extends Error {
  constructor(status, body) {
    const parsed = typeof body === 'string' ? JSON.parse(body) : body
    super(parsed.error || `Request failed (${status})`)
    this.status = status
    this.code = parsed.code
  }

  get isAuth() { return this.status === 401 }
  get isNotFound() { return this.status === 404 }
  get isRateLimit() { return this.status === 429 }
}
```

## Phase 5: Preview app migration

Migrate Preview app from hand-rolled `api.js` to fazt-sdk:

```javascript
// Before (servers/local/preview/src/lib/api.js)
async function request(url, options = {}) { ... }
export const api = { getPhotos(filter) { ... } }

// After
import { createAppClient } from '../../packages/fazt-sdk/index.js'
// Wait — Preview is in servers/ which is gitignored...
```

**Problem**: Preview lives in `servers/` (gitignored). It can't import from
`admin/packages/fazt-sdk/`. Options:

1. **Move fazt-sdk to repo root** — `packages/fazt-sdk/` alongside `admin/`
2. **Publish to npm** — `@fazt/sdk` (overkill for now)
3. **Copy SDK into app** — defeats the purpose
4. **Symlink** — fragile

**Recommendation**: Move to repo root. fazt-sdk is a shared package, not an
admin-specific one. This matches the new vision.

```
fazt/
├── packages/
│   └── fazt-sdk/        # Shared SDK (tracked in git)
├── admin/               # Admin UI (tracked)
│   └── packages/ → gone, imports from ../../packages/fazt-sdk/
└── servers/             # Test apps (gitignored, but can import from ../packages/)
```

## Stale docs to fix

1. **CLAUDE.md** — References React hooks, @tanstack/react-query, "Framework
   adapters" section. Admin has always been Vue. Remove React references.
2. **CLAUDE.md** — fazt-sdk architecture diagram shows `adapters/` and
   `integrations/` directories that don't exist. Update to match reality.
3. **knowledge-base/workflows/admin-ui/architecture.md** — References old "zap"
   state library. Admin uses Pinia now.

## Implementation order

1. Fix stale CLAUDE.md docs (quick, high value)
2. Move fazt-sdk to `packages/fazt-sdk/` at repo root
3. Phase 1: Restructure into core/ + admin.js
4. Phase 2: Add app.js with auth + http + upload + paginate
5. Phase 3: Upload progress (XMLHttpRequest)
6. Phase 4: Structured errors
7. Phase 5: Migrate Preview to use fazt-sdk
8. Update mock.js with app-route handlers

## Non-goals

- No framework-specific adapters (Vue composables, React hooks) — stores wrap
  the SDK directly, which works fine
- No caching layer yet — Pinia/app state handles this at the store level
- No offline support
- No WebSocket integration (separate concern)
