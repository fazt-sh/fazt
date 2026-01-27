---
name: fazt-app
description: Build and deploy polished Vue+API apps to fazt instances. Creates PWA-ready apps with advanced storage, session management, authentication, and production-quality UX. Use when building new fazt apps.
context: fork
---

# /fazt-app - Build Fazt Apps

Build and deploy polished, PWA-ready apps to fazt instances with Claude.

## Supporting Documentation

This skill includes reference materials:
- **examples/cashflow.md** - Full reference app with advanced storage patterns
- **patterns/layout.md** - Fixed-height layouts, responsive design
- **patterns/modals.md** - Modal patterns with click-outside close
- **patterns/testing.md** - Testing framework for validating data flow and state
- **templates/** - Complete app template (mirrors final app structure)
- **koder/CAPACITY.md** - Performance limits and real-time capability guide

Reference these files when building apps to follow established patterns.

## Authentication Note

Fazt has built-in authentication with OAuth (Google, GitHub, Discord, Microsoft).
To use auth in apps, the instance must have at least one provider configured:

```bash
# Configure provider (run once per instance)
fazt auth provider google --client-id XXX --client-secret YYY --db <path>
fazt auth provider google --enable --db <path>
```

See the **Authentication** section below for integration patterns.

## Standard Assets

Use the fazt branding assets from `.claude/fazt-assets/` for all apps unless
the user specifies custom branding:

```
.claude/fazt-assets/
├── apple-touch-icon.png   # 180x180 iOS home screen icon
├── favicon.png            # Browser tab icon
├── logo.png               # Full logo with background
├── logo.svg               # Vector logo
├── logo-transparent.png   # Logo with transparency
└── social_preview.png     # OpenGraph/social sharing image
```

**Copy these to your app's root or `static/` directory:**
```bash
cp .claude/fazt-assets/favicon.png servers/zyt/<app>/
cp .claude/fazt-assets/apple-touch-icon.png servers/zyt/<app>/
```

**Reference in index.html:**
```html
<link rel="icon" type="image/png" href="/favicon.png">
<link rel="apple-touch-icon" href="/apple-touch-icon.png">
```

## Capacity Awareness

Before building, understand fazt's limits (see `koder/CAPACITY.md`):

| Capability | Limit ($6 VPS) | Notes |
|------------|----------------|-------|
| Read throughput | ~20,000/s | Static files, VFS |
| Write throughput | ~800/s | SQLite single-writer |
| Mixed workload | ~2,300/s | 30% writes typical |
| WebSocket connections | 5,000-10,000 | Real-time features |
| Broadcasts/sec | 10,000+ | Cursors, presence, chat |

**Key insight for real-time apps:** Broadcasts (cursors, typing, presence) are
unlimited - they never hit disk. Only persist what matters (messages, documents).

For collaborative features, batch writes and keep ephemeral data in-memory.

## Usage

```
/fazt-app <description>
/fazt-app "build a pomodoro tracker with sound effects"
```

## Context

You are building an app for fazt - a single-binary personal cloud.
Apps should feel native, polished, and delightful - like Apple apps.

---

## Workflow

### 0. Check Available Peers

First, check what fazt peers are configured:

```bash
fazt remote list
```

Common peers (from CLAUDE.md):
- `local` - Local dev server (http://192.168.64.3:8080)
- `zyt` - Production server (https://zyt.app)

### 1. Scaffold

**CRITICAL: Apps MUST be created in `servers/<peer>/` directory!**

```bash
# CORRECT - create in servers/<peer>/ (typically servers/zyt/ for production apps)
mkdir -p servers/zyt/<name>
cd servers/zyt/<name>

# Copy standard assets (favicon, icons)
cp .claude/fazt-assets/favicon.png servers/zyt/<name>/
cp .claude/fazt-assets/apple-touch-icon.png servers/zyt/<name>/

# WRONG - never create in repo root!
# This pollutes the fazt source code
```

The `servers/` directory is gitignored - apps are instance-specific, not
part of fazt source code.

### 2. Customize Fully

Copy from templates and customize:

- `index.html` - PWA meta tags, fonts, dark mode, import maps
- `src/main.js` - Vue app with Pinia stores, router
- `src/stores/` - Pinia stores for state management
- `src/pages/` - Page components
- `src/components/` - Reusable components
- `src/lib/` - Utilities (api, session, settings)
- `api/main.js` - Session-scoped serverless API

### 3. Test Locally First

```bash
fazt app deploy servers/zyt/<name> --to local
```

Access at: `http://<name>.192.168.64.3.nip.io:8080`

**Debug endpoints:**
- `/_fazt/info` - app metadata (includes app_id)
- `/_fazt/storage` - storage contents
- `/_fazt/logs` - recent execution logs
- `/_fazt/errors` - error logs

**STOP HERE** - Show URL to user and wait for approval before production deploy.

### 4. Deploy to Production (After Approval)

```bash
# Deploy to configured production peer (check with `fazt remote list`)
fazt app deploy servers/zyt/<name> --to <production-peer>
```

Example: `fazt app deploy servers/zyt/myapp --to zyt` → `https://myapp.zyt.app`

---

## Required Customizations

### index.html - Full PWA Setup

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover, user-scalable=no">

  <!-- PWA Meta Tags -->
  <meta name="mobile-web-app-capable" content="yes">
  <meta name="apple-mobile-web-app-capable" content="yes">
  <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
  <meta name="apple-mobile-web-app-title" content="App Name">
  <meta name="theme-color" content="#ffffff" media="(prefers-color-scheme: light)">
  <meta name="theme-color" content="#0a0a0a" media="(prefers-color-scheme: dark)">

  <!-- Favicon (copy from .claude/fazt-assets/) -->
  <link rel="icon" type="image/png" href="/favicon.png">
  <link rel="apple-touch-icon" href="/apple-touch-icon.png">

  <title>App Name</title>

  <!-- Fonts (choose one) -->
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">

  <!-- Vue + Tailwind -->
  <script type="importmap">
  { "imports": { "vue": "https://unpkg.com/vue@3/dist/vue.esm-browser.prod.js" } }
  </script>
  <script src="https://cdn.tailwindcss.com"></script>
  <script>
    tailwind.config = {
      darkMode: 'class',
      theme: { extend: { fontFamily: { sans: ['Inter', 'system-ui', 'sans-serif'] } } }
    }
  </script>

  <style>
    * { font-family: 'Inter', system-ui, sans-serif; }

    /* Full height, no scrollbars on outer container */
    html, body { height: 100%; overflow: hidden; margin: 0; padding: 0; }

    /* iOS safe areas */
    body { padding: env(safe-area-inset-top) env(safe-area-inset-right) env(safe-area-inset-bottom) env(safe-area-inset-left); }

    /* Full height app container */
    #app { height: 100%; }

    /* Touch feedback */
    .touch-active:active { opacity: 0.7; transform: scale(0.98); }

    /* Prevent text selection on buttons */
    button, .no-select { -webkit-user-select: none; user-select: none; }

    /* Custom scrollbar for inner containers */
    ::-webkit-scrollbar { width: 6px; }
    ::-webkit-scrollbar-thumb { background: rgba(128,128,128,0.3); border-radius: 3px; }
    ::-webkit-scrollbar-track { background: transparent; }

    /* Prevent scrollbar layout shift */
    .overflow-y-auto { scrollbar-gutter: stable; }

    /* Smooth transitions */
    .transition-all { transition-duration: 150ms; }
  </style>
</head>
<body class="font-sans antialiased bg-neutral-50 dark:bg-neutral-950 text-neutral-900 dark:text-neutral-100">
  <div id="app"></div>
  <script type="module" src="./src/main.js"></script>
</body>
</html>
```

### Session Management (URL-based, NOT localStorage)

```javascript
// src/lib/session.js
const WORDS = [
  'cat', 'dog', 'fox', 'owl', 'bee', 'ant', 'elk', 'bat', 'jay', 'hen',
  'red', 'blue', 'gold', 'jade', 'mint', 'rose', 'sage', 'teal', 'cyan', 'lime',
  'apple', 'berry', 'grape', 'lemon', 'mango', 'peach', 'plum', 'pear', 'kiwi',
  'cloud', 'river', 'stone', 'leaf', 'wave', 'star', 'moon', 'sun', 'sky', 'snow',
  'swift', 'bold', 'calm', 'keen', 'wild', 'free', 'warm', 'cool', 'soft', 'fast'
]

function generateSessionId() {
  const pick = () => WORDS[Math.floor(Math.random() * WORDS.length)]
  return `${pick()}-${pick()}-${pick()}`
}

export function getSession() {
  const params = new URLSearchParams(location.search)
  let id = params.get('s')
  if (!id) {
    id = generateSessionId()
    history.replaceState(null, '', `?s=${id}`)
  }
  return id
}

export function getSessionUrl() {
  return `${location.origin}${location.pathname}?s=${getSession()}`
}

export function generateNewSession() {
  const id = generateSessionId()
  history.replaceState(null, '', `?s=${id}`)
  return id
}
```

### API Helper (with session)

```javascript
// src/lib/api.js
import { getSession } from './session.js'

async function request(url, options = {}) {
  const res = await fetch(url, {
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(err.error || 'Request failed')
  }
  return res.json()
}

export const api = {
  session: null,

  init() {
    this.session = getSession()
    return this.session
  },

  async get(path) {
    const sep = path.includes('?') ? '&' : '?'
    return request(`${path}${sep}session=${encodeURIComponent(this.session)}`)
  },

  async post(path, data) {
    return request(path, {
      method: 'POST',
      body: JSON.stringify({ ...data, session: this.session })
    })
  },

  async put(path, data) {
    return request(path, {
      method: 'PUT',
      body: JSON.stringify({ ...data, session: this.session })
    })
  },

  async delete(path) {
    const sep = path.includes('?') ? '&' : '?'
    return request(`${path}${sep}session=${encodeURIComponent(this.session)}`, {
      method: 'DELETE'
    })
  }
}
```

### Settings System (with theme)

```javascript
// src/lib/settings.js
const SETTINGS_KEY = 'appname_settings'

const DEFAULTS = {
  theme: 'system',      // 'light', 'dark', 'system'
  soundEnabled: true,
  animationsEnabled: true,
  displayName: ''
}

export function getSettings() {
  try {
    const stored = localStorage.getItem(SETTINGS_KEY)
    return stored ? { ...DEFAULTS, ...JSON.parse(stored) } : { ...DEFAULTS }
  } catch {
    return { ...DEFAULTS }
  }
}

export function saveSettings(settings) {
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings))
  applyTheme(settings.theme)
}

export function applyTheme(theme) {
  const isDark = theme === 'dark' ||
    (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)
  document.documentElement.classList.toggle('dark', isDark)
}

// Initialize theme on load
applyTheme(getSettings().theme)

// Listen for system theme changes
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
  const settings = getSettings()
  if (settings.theme === 'system') applyTheme('system')
})
```

### Sound Effects

```javascript
// In main.js or separate lib/sounds.js
function playSound(type) {
  if (!settings.value.soundEnabled) return
  try {
    const ctx = new (window.AudioContext || window.webkitAudioContext)()
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    osc.connect(gain)
    gain.connect(ctx.destination)

    if (type === 'success') {
      osc.frequency.setValueAtTime(523, ctx.currentTime)
      osc.frequency.setValueAtTime(659, ctx.currentTime + 0.1)
      osc.frequency.setValueAtTime(784, ctx.currentTime + 0.2)
      gain.gain.setValueAtTime(0.3, ctx.currentTime)
      gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.4)
      osc.start(); osc.stop(ctx.currentTime + 0.4)
    } else if (type === 'tap') {
      osc.frequency.setValueAtTime(440, ctx.currentTime)
      gain.gain.setValueAtTime(0.15, ctx.currentTime)
      gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.1)
      osc.start(); osc.stop(ctx.currentTime + 0.1)
    } else if (type === 'error') {
      osc.frequency.setValueAtTime(200, ctx.currentTime)
      gain.gain.setValueAtTime(0.2, ctx.currentTime)
      gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.2)
      osc.start(); osc.stop(ctx.currentTime + 0.2)
    }
  } catch (e) {}
}
```

---

## API Pattern (Session-Scoped)

```javascript
// api/main.js
var ds = fazt.storage.ds

function genId() {
  return Date.now().toString(36) + Math.random().toString(36).slice(2, 8)
}

var parts = request.path.split('/').filter(Boolean)
var resource = parts[1]
var id = parts.length > 2 ? parts[2] : null
var session = request.query.session || (request.body && request.body.session)

if (!session) {
  respond(400, { error: 'session required' })
} else if (resource === 'items') {
  if (request.method === 'GET' && !id) {
    // List all for this session
    respond({ items: ds.find('items', { session: session }) })
  } else if (request.method === 'POST') {
    // Create new item for this session
    var item = { id: genId(), session: session, ...request.body, created: Date.now() }
    ds.insert('items', item)
    respond(201, item)
  } else if (request.method === 'PUT' && id) {
    // Update item (verify session ownership)
    ds.update('items', { id: id, session: session }, request.body)
    respond(ds.findOne('items', { id: id, session: session }))
  } else if (request.method === 'DELETE' && id) {
    // Delete item (verify session ownership)
    ds.delete('items', { id: id, session: session })
    respond({ ok: true })
  }
} else {
  respond(404, { error: 'Not found' })
}
```

---

## Required UI Features

### Settings Panel

Every app needs a settings panel with:

1. **Theme selector** - Light / Dark / System (3 buttons)
2. **Sound toggle** - Enable/disable sound effects
3. **Animations toggle** - Enable/disable motion (accessibility)
4. **Session info** - Show session ID, copy link button, new session button

### Keyboard Shortcuts

- **Escape** - Close any open modal/panel
- **Enter** - Submit forms
- App-specific shortcuts as needed

### Modal Pattern

```javascript
// In setup()
const showModal = ref(false)

function closeModal() {
  showModal.value = false
}

// Escape key handler
onMounted(() => {
  const handleEscape = (e) => {
    if (e.key === 'Escape') closeModal()
  }
  document.addEventListener('keydown', handleEscape)
  onUnmounted(() => document.removeEventListener('keydown', handleEscape))
})

// Template - click outside to close
`<div v-if="showModal" class="fixed inset-0 z-50" @click.self="closeModal">
  <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="closeModal"></div>
  <div class="relative ...">
    <!-- Modal content -->
  </div>
</div>`
```

---

## Design Guidelines

### Colors (Neutral + Accent)

| Element | Light Mode | Dark Mode |
|---------|------------|-----------|
| Background | `bg-neutral-50` | `bg-neutral-950` |
| Card | `bg-white` | `bg-neutral-900` |
| Text | `text-neutral-900` | `text-neutral-100` |
| Muted | `text-neutral-500` | `text-neutral-400` |
| Border | `border-neutral-200` | `border-neutral-800` |
| Accent | `bg-blue-500 text-white` | same |

### Layout Best Practices

**Fixed Height Container (Prevents Scroll Jumps)**:
- App container should be `h-screen flex flex-col overflow-hidden`
- Header should be `flex-none` (fixed at top)
- Content area should be `flex-1 overflow-y-auto` (scrollable)
- Use `scrollbar-gutter: stable` to prevent layout shift

### Typography

| Style | Class | Use |
|-------|-------|-----|
| Display | `text-3xl font-bold` | Hero numbers |
| Title | `text-xl font-semibold` | Page titles |
| Headline | `text-lg font-medium` | Section headers |
| Body | `text-base` | Main content |
| Caption | `text-sm text-neutral-500` | Labels, hints |

### Spacing

- Generous padding: `p-4` to `p-6`
- Card gaps: `gap-3` to `gap-4`
- Section gaps: `space-y-6` to `space-y-8`

### Corners

- Small elements: `rounded-lg`
- Cards: `rounded-xl` to `rounded-2xl`
- Modals: `rounded-2xl` to `rounded-3xl`
- Pills/badges: `rounded-full`

### Touch Targets

- Minimum 44x44px for tap targets
- Add `touch-active:active` for feedback
- Use `cursor-pointer` on clickable elements

---

## Authentication

Fazt provides built-in authentication with OAuth providers (Google, GitHub, etc.)
and cookie-based SSO across subdomains.

### When to Use Auth vs Sessions

| Use Case | Approach |
|----------|----------|
| Public app, shareable links | URL sessions (`?s=cat-blue-river`) |
| User accounts, personal data | Fazt auth (`fazt.auth.*`) |
| Multi-device sync | Fazt auth (user has account) |
| Anonymous collaboration | URL sessions |

You can combine both: auth for user identity, sessions for shareable workspaces.

### Auth API Reference (Serverless)

```javascript
// api/main.js - Server-side auth checks

// Get current user (null if not logged in)
var user = fazt.auth.getUser()
// Returns: { id, email, name, picture, role, provider }

// Check auth state
fazt.auth.isLoggedIn()   // boolean
fazt.auth.isOwner()      // boolean (role === 'owner')
fazt.auth.isAdmin()      // boolean (role === 'owner' or 'admin')
fazt.auth.hasRole('admin')  // boolean

// Require auth (redirects to login if not authenticated)
fazt.auth.requireLogin()     // Throws redirect to /auth/login
fazt.auth.requireAdmin()     // Throws 403 if not admin
fazt.auth.requireOwner()     // Throws 403 if not owner
fazt.auth.requireRole('admin')  // Throws 403 if missing role

// Get auth URLs
fazt.auth.getLoginURL('/dashboard')   // Login URL with redirect
fazt.auth.getLogoutURL()              // Logout URL
```

### Protected API Pattern

```javascript
// api/main.js - Protect entire API
fazt.auth.requireLogin()  // All requests require auth

var user = fazt.auth.getUser()
var ds = fazt.storage.ds

var parts = request.path.split('/').filter(Boolean)
var resource = parts[1]
var id = parts[2]

if (resource === 'me') {
  // Return current user
  respond({ user: user })

} else if (resource === 'items') {
  if (request.method === 'GET') {
    // User's items only
    respond({ items: ds.find('items', { userId: user.id }) })
  } else if (request.method === 'POST') {
    var item = {
      id: Date.now().toString(36),
      userId: user.id,
      ...request.body,
      created: Date.now()
    }
    ds.insert('items', item)
    respond(201, item)
  }
}
```

### Vue Auth Composable

```javascript
// src/lib/auth.js
import { ref, readonly } from 'vue'

const user = ref(null)
const loading = ref(true)
const error = ref(null)

let initialized = false

export function useAuth() {
  async function init() {
    if (initialized) return
    initialized = true

    try {
      const res = await fetch('/api/me')
      if (res.ok) {
        const data = await res.json()
        user.value = data.user
      }
    } catch (e) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  function login(redirect = window.location.pathname) {
    window.location.href = '/auth/login?redirect=' + encodeURIComponent(redirect)
  }

  function logout() {
    window.location.href = '/auth/logout'
  }

  // Auto-init on first use
  init()

  return {
    user: readonly(user),
    loading: readonly(loading),
    error: readonly(error),
    isLoggedIn: () => !!user.value,
    isAdmin: () => user.value?.role === 'owner' || user.value?.role === 'admin',
    isOwner: () => user.value?.role === 'owner',
    login,
    logout
  }
}
```

### Protected Page Pattern

```javascript
// src/pages/Dashboard.js
import { useAuth } from '../lib/auth.js'

export default {
  template: `
    <div v-if="loading" class="flex items-center justify-center h-full">
      <div class="animate-spin rounded-full h-8 w-8 border-2 border-blue-500 border-t-transparent"></div>
    </div>

    <div v-else-if="!user" class="flex flex-col items-center justify-center h-full gap-4">
      <p class="text-neutral-500">Please sign in to continue</p>
      <button @click="login" class="px-4 py-2 bg-blue-500 text-white rounded-lg">
        Sign In
      </button>
    </div>

    <div v-else class="p-6">
      <h1 class="text-2xl font-bold">Welcome, {{ user.name }}</h1>
      <button @click="logout" class="mt-4 text-sm text-neutral-500 hover:text-neutral-700">
        Sign Out
      </button>
    </div>
  `,

  setup() {
    const { user, loading, login, logout } = useAuth()
    return { user, loading, login, logout }
  }
}
```

### Auth + Sessions Combined

For apps that need both user identity AND shareable workspaces:

```javascript
// src/lib/api.js
import { getSession } from './session.js'
import { useAuth } from './auth.js'

export const api = {
  session: null,

  init() {
    this.session = getSession()
    return this.session
  },

  async get(path) {
    // Session in URL, auth in cookie (automatic)
    const sep = path.includes('?') ? '&' : '?'
    return fetch(`${path}${sep}session=${encodeURIComponent(this.session)}`)
      .then(r => r.json())
  }
}

// api/main.js - Server side
var user = fazt.auth.getUser()  // May be null
var session = request.query.session

// User-owned data requires auth
if (resource === 'user-settings') {
  fazt.auth.requireLogin()
  respond(ds.findOne('settings', { userId: user.id }))
}

// Session data is accessible to anyone with the session ID
if (resource === 'workspace') {
  respond(ds.find('items', { session: session }))
}
```

### Auth UI Components

```javascript
// src/components/AuthStatus.js
import { useAuth } from '../lib/auth.js'

export default {
  template: `
    <div class="flex items-center gap-3">
      <template v-if="loading">
        <div class="w-8 h-8 rounded-full bg-neutral-200 dark:bg-neutral-700 animate-pulse"></div>
      </template>

      <template v-else-if="user">
        <img v-if="user.picture" :src="user.picture" class="w-8 h-8 rounded-full">
        <div v-else class="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white text-sm font-medium">
          {{ user.name?.[0]?.toUpperCase() || '?' }}
        </div>
        <span class="text-sm hidden sm:inline">{{ user.name }}</span>
        <button @click="logout" class="text-sm text-neutral-500 hover:text-neutral-700 dark:hover:text-neutral-300">
          Sign Out
        </button>
      </template>

      <template v-else>
        <button @click="login" class="px-3 py-1.5 text-sm bg-blue-500 text-white rounded-lg hover:bg-blue-600">
          Sign In
        </button>
      </template>
    </div>
  `,

  setup() {
    const { user, loading, login, logout } = useAuth()
    return { user, loading, login, logout }
  }
}
```

---

## Storage Reference

```javascript
// Document Store (primary)
var ds = fazt.storage.ds
ds.insert('collection', { ...doc })
ds.find('collection', { field: value })
ds.findOne('collection', { id: '...' })
ds.update('collection', { id: '...' }, { ...changes })
ds.delete('collection', { id: '...' })

// Key-Value (simple lookups, caches)
var kv = fazt.storage.kv
kv.set('key', value)
kv.set('key', value, ttlMs)  // with expiry
kv.get('key')
kv.delete('key')
kv.list('prefix')

// Blob Storage (files)
var s3 = fazt.storage.s3
s3.put('path', data, mimeType)
s3.get('path')
s3.delete('path')
s3.list('prefix')
```

---

## Instructions

When user invokes `/fazt-app`:

1. **Check peers**: Run `fazt remote list` to see configured peers
2. **Parse** the description to understand what to build
3. **Determine auth needs**:
   - Personal/private data → Use fazt auth
   - Shareable links/collaboration → Use URL sessions
   - Both user accounts + sharing → Combine both
4. **Scaffold** in correct location:
   ```bash
   mkdir -p servers/zyt/<name>
   ```
   **NEVER create apps in repo root!** Always use `servers/<peer>/<name>` path.
5. **Customize fully** using templates from this skill:
   - `index.html` - PWA meta, fonts, dark mode, import maps, full-height layout
   - `src/main.js` - Vue app with Pinia, router
   - `src/stores/` - Pinia stores for state management
   - `src/pages/` - Page components (Home, Settings, etc.)
   - `src/components/ui/` - Reusable UI components
   - `src/lib/` - Utilities (api.js, session.js, settings.js, auth.js)
   - `api/main.js` - Session-scoped or auth-protected serverless API
6. **Deploy to local**: `fazt app deploy servers/zyt/<name> --to local`
7. **STOP** - Present local URL to user: `http://<name>.192.168.64.3.nip.io:8080`
8. **Wait for approval** before production deploy
9. After approval: `fazt app deploy servers/zyt/<name> --to <production-peer>`
10. Report production URL (e.g., `https://<name>.zyt.app`)

**Note**: The fazt system generates proper app_ids (like `app_7f3k9x2m`).
The app name (like "cashflow") becomes the **alias/subdomain**, not the ID.
