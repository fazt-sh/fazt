# Authentication Integration

How to integrate fazt authentication into apps.

## OAuth Requirement: Remote Only

**OAuth authentication requires deployment to a remote peer with a real domain.**

| Environment | OAuth Works? | Why |
|-------------|--------------|-----|
| Local (IP address) | No | No HTTPS, invalid callback URL |
| Remote (real domain) | Yes | HTTPS + valid callback URL |

OAuth providers (Google, GitHub) require:
- HTTPS callback URL
- Domain matches registered application
- No IP addresses or localhost

## Check If OAuth Is Already Configured

Before building an auth-enabled app, check if the target peer has OAuth set up:

```bash
fazt @<remote-peer> auth providers
```

If you see `google enabled`, you can use auth immediately. No setup needed.

## Mock OAuth Provider (Local Development)

**Status**: Proposed feature for fazt

The ideal solution is a `dev` OAuth provider built into fazt that:
- Only available on local/non-HTTPS instances
- Shows a simple login form (enter any email/name)
- Creates a real auth session (same cookies, same flow)
- `fazt.auth.getUser()` returns the mock user exactly like real OAuth
- **Zero code changes** when deploying to remote with real OAuth

This would allow full auth flow testing locally with the same code path as
production.

**How it would work:**
```
1. User clicks "Sign in" → /auth/login
2. Fazt detects local mode → shows dev provider option
3. User enters email/name in simple form
4. Fazt creates session, sets cookie
5. fazt.auth.getUser() returns { id, email, name, role }
6. Deploy to remote → same code uses real Google OAuth
```

Until this feature exists, use the workarounds below.

---

## Current Local Testing Strategies

Since real OAuth doesn't work locally (requires HTTPS), use these approaches:

### Strategy 1: Feature Toggle

Build app to work with or without auth. Test non-auth features locally.

```javascript
// api/main.js
var user = fazt.auth.getUser()  // null locally

if (resource === 'public-data') {
  // Works without auth
  respond({ items: ds.find('public', {}) })
}

if (resource === 'my-data') {
  if (!user) {
    respond(401, { error: 'Login required' })
    return
  }
  respond({ items: ds.find('items', { userId: user.id }) })
}
```

```javascript
// Frontend - graceful degradation
const { user, loading } = useAuth()

// Show login prompt instead of breaking
if (!user) {
  return <LoginPrompt />
}
```

### Strategy 2: Mock User Header

Add mock user support for local development only.

```javascript
// api/main.js
var user = fazt.auth.getUser()

// Mock user for local testing (NEVER in production)
if (!user && request.headers['x-mock-user']) {
  try {
    user = JSON.parse(request.headers['x-mock-user'])
  } catch (e) {}
}

if (!user) {
  respond(401, { error: 'Unauthorized' })
  return
}
```

```javascript
// Frontend - src/lib/api.js (development only)
const MOCK_USER = {
  id: 'mock_user_1',
  email: 'dev@example.com',
  name: 'Dev User',
  role: 'user'
}

async function request(url, options = {}) {
  const headers = { 'Content-Type': 'application/json' }

  // Add mock user in development
  if (window.location.hostname.includes('192.168')) {
    headers['X-Mock-User'] = JSON.stringify(MOCK_USER)
  }

  return fetch(url, { headers, ...options })
}
```

### Strategy 3: Test Auth on Remote, Everything Else Local

```bash
# 1. Test UI and API logic locally (builds automatically)
fazt app deploy ./my-app --to local
# Test at http://my-app.<local-ip>.nip.io:<port>

# 2. Test auth flow on remote (builds automatically)
fazt app deploy ./my-app --to <remote-peer>
# Test at https://my-app.<domain>
```

## Frontend Auth Integration

### Critical: Use Absolute URLs for Redirects

**OAuth flows through the root domain** (e.g., `example.com`), not the subdomain
(`myapp.example.com`). After authentication, fazt redirects to the `redirect`
parameter you provide.

| Redirect Value | After OAuth Goes To | Result |
|----------------|---------------------|--------|
| `/dashboard` | `<domain>/dashboard` | WRONG - loses subdomain |
| `https://myapp.<domain>/dashboard` | `myapp.<domain>/dashboard` | Correct |

**Always use absolute URLs** (including origin) for the redirect parameter.

### Vue Composable

```javascript
// src/lib/auth.js
import { ref, readonly } from 'vue'

const user = ref(null)
const loading = ref(true)

export function useAuth() {
  async function init() {
    try {
      const res = await fetch('/api/me')
      if (res.ok) {
        const data = await res.json()
        user.value = data.user
      }
    } finally {
      loading.value = false
    }
  }

  // IMPORTANT: Use location.href (absolute) not location.pathname (relative)
  // OAuth goes through root domain - relative paths lose the subdomain
  function login(redirect = location.href) {
    location.href = '/auth/login?redirect=' + encodeURIComponent(redirect)
  }

  // IMPORTANT: Logout requires POST, not GET (returns 405 on GET)
  async function logout() {
    await fetch('/auth/logout', { method: 'POST' })
    user.value = null
    location.href = '/'  // Or location.reload()
  }

  init()

  return {
    user: readonly(user),
    loading: readonly(loading),
    isLoggedIn: () => !!user.value,
    login,
    logout
  }
}
```

### API Endpoint for User Info

```javascript
// api/main.js
if (resource === 'me') {
  var user = fazt.auth.getUser()
  respond({ user: user })  // null if not logged in
}
```

### Protected Route Pattern

```javascript
// src/pages/Dashboard.js
export default {
  setup() {
    const { user, loading, login } = useAuth()

    return { user, loading, login }
  },

  template: `
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="!user" class="login-prompt">
      <p>Please sign in to continue</p>
      <button @click="login">Sign In</button>
    </div>
    <div v-else>
      <h1>Welcome, {{ user.name }}</h1>
      <!-- Protected content -->
    </div>
  `
}
```

## Auth URLs

| URL | Method | Purpose |
|-----|--------|---------|
| `/auth/login` | GET | Initiate login flow |
| `/auth/login?redirect=<absolute-url>` | GET | Login then redirect (use full URL!) |
| `/auth/logout` | **POST** | Log out user (GET returns 405) |
| `/auth/callback/<provider>` | GET | OAuth callback (internal) |

**Remember**: The redirect parameter must be an absolute URL including the
origin (e.g., `https://myapp.<domain>/dashboard`), not a relative path.

## Session + Auth Combined

For apps needing both user identity AND shareable workspaces:

```javascript
// api/main.js
var user = fazt.auth.getUser()  // May be null
var session = request.query.session

// User-specific data (requires auth)
if (resource === 'settings') {
  fazt.auth.requireLogin()
  respond(ds.findOne('settings', { userId: user.id }))
}

// Session data (shareable, no auth required)
if (resource === 'workspace') {
  respond(ds.find('workspace_items', { session: session }))
}

// User owns workspace items but workspace is shareable
if (resource === 'items' && request.method === 'POST') {
  var item = {
    id: genId(),
    session: session,
    createdBy: user ? user.id : 'anonymous',
    ...request.body
  }
  ds.insert('items', item)
  respond(201, item)
}
```

---

## Common Gotchas

Practical lessons learned integrating OAuth with fazt apps.

### 1. OAuth Redirect Loses Subdomain

**Symptom**: After Google login, user lands on `example.com` instead of `myapp.example.com`.

**Cause**: OAuth flows through the root domain. Relative paths like `/dashboard` redirect to `example.com/dashboard`, not your subdomain.

**Fix**: Always use absolute URLs with `location.origin`:

```javascript
// WRONG - loses subdomain
location.href = '/auth/login?redirect=/dashboard'

// CORRECT - preserves full URL
const redirect = location.origin + '/#/dashboard'
location.href = '/auth/login?redirect=' + encodeURIComponent(redirect)
```

### 2. Logout Returns 405 Method Not Allowed

**Symptom**: `location.href = '/auth/logout'` returns 405 error.

**Cause**: Logout endpoint only accepts POST (CSRF protection).

**Fix**: Use fetch with POST:

```javascript
async function signOut() {
  await fetch('/auth/logout', { method: 'POST' })
  user.value = null
  location.href = '/'
}
```

### 3. Use Hash Routing for Static Deployment

**Symptom**: Direct URL access or page refresh returns 404.

**Cause**: History mode (`/dashboard`) requires server-side routing config. Fazt static servers don't have this.

**Fix**: Use hash routing - works on any static server:

```javascript
import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(),  // URLs become /#/dashboard
  routes: [...]
})
```

### 4. Add Retry Logic for Auth Check

**Symptom**: `/api/me` occasionally returns 500, works on retry.

**Cause**: Auth API can timeout intermittently on cold starts or under load.

**Fix**: Add retry logic with timeout:

```javascript
async function checkAuth() {
  for (let i = 0; i < 3; i++) {
    try {
      const controller = new AbortController()
      const timeout = setTimeout(() => controller.abort(), 3000)

      const res = await fetch('/api/me', { signal: controller.signal })
      clearTimeout(timeout)

      if (res.ok) {
        const data = await res.json()
        if (!data.error) {
          user.value = data.user
          break
        }
      }
      await new Promise(r => setTimeout(r, 500))
    } catch (e) {
      if (i === 2) user.value = null
    }
  }
  loading.value = false
}
```

### 5. Service Worker Blocks External Images

**Symptom**: Console flooded with CSP violations for CDN images, Google profile pictures.

**Cause**: Default Service Worker tries to cache ALL fetch requests. External URLs violate Content Security Policy.

**Fix**: Either skip external URLs in SW or remove it entirely:

```javascript
// Option A: Skip external URLs in SW
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url)
  if (url.origin !== self.location.origin) return  // Let external pass through
  // ... handle same-origin only
})

// Option B: Remove sw.js and registration code from index.html
```

### 6. Z-Index Stacking Context with Dropdowns

**Symptom**: User menu dropdown appears behind other panels (filter panel, etc.).

**Cause**: `backdrop-blur` on header creates new stacking context. Dropdown's `z-50` only works within header, not against siblings outside it.

**Fix**: Add z-index to header container, not just dropdown:

```html
<!-- Header needs z-index too -->
<header class="... relative z-50">
  <!-- Dropdown inside -->
  <div class="absolute ... z-[100]">...</div>
</header>

<!-- Other panels have lower z-index -->
<div class="relative z-40">FilterPanel</div>
```

### Quick Checklist

- [ ] OAuth redirect uses `location.origin + '/#/path'` (absolute URL)
- [ ] Logout uses `fetch('/auth/logout', { method: 'POST' })`
- [ ] Router uses `createWebHashHistory()` for static deploy
- [ ] Frontend has retry logic for `/api/me`
- [ ] Header has z-index higher than other panels
- [ ] Service worker skips external URLs (or removed entirely)
