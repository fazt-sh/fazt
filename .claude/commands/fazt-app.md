# /fazt-app - Build Fazt Apps

Build and deploy polished, PWA-ready apps to fazt instances with Claude.

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

### 1. Scaffold

**CRITICAL: Apps MUST be created in `servers/zyt/` directory!**

```bash
# CORRECT - always use this path
fazt app create servers/zyt/<name> --template vue-api

# WRONG - never create in repo root!
# fazt app create <name> --template vue-api
```

The `servers/zyt/` directory is gitignored - apps are instance-specific, not
part of fazt source code. Creating apps in the repo root pollutes the codebase.

### 2. Customize Fully

The template provides a minimal starting point. You MUST customize:

- `index.html` - Add PWA meta tags, fonts, dark mode, custom styles
- `src/main.js` - Build the full Vue app with components
- `src/lib/session.js` - Use URL-based sessions (not localStorage)
- `src/lib/api.js` - Include session in all API calls
- `api/main.js` - Scope all data by session ID

### 3. Test Locally First

```bash
fazt app deploy servers/zyt/<name> --to local
```

Access at: `http://<name>.192.168.64.3.nip.io:8080`

**Debug endpoints:**
- `/_fazt/info` - app metadata
- `/_fazt/storage` - storage contents
- `/_fazt/errors` - recent errors

**STOP HERE** - Show URL to user and wait for approval before production deploy.

### 4. Deploy to Production (After Approval)

```bash
fazt app deploy servers/zyt/<name> --to zyt
```

Access at: `https://<name>.zyt.app`

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

    /* iOS safe areas */
    body { padding: env(safe-area-inset-top) env(safe-area-inset-right) env(safe-area-inset-bottom) env(safe-area-inset-left); }

    /* Touch feedback */
    .touch-active:active { opacity: 0.7; transform: scale(0.98); }

    /* Prevent text selection on buttons */
    button, .no-select { -webkit-user-select: none; user-select: none; }

    /* Custom scrollbar */
    ::-webkit-scrollbar { width: 6px; }
    ::-webkit-scrollbar-thumb { background: rgba(128,128,128,0.3); border-radius: 3px; }
  </style>
</head>
<body class="font-sans antialiased bg-neutral-50 dark:bg-neutral-950 text-neutral-900 dark:text-neutral-100 min-h-screen">
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
  <div class="absolute inset-0 bg-black/50 backdrop-blur-sm"></div>
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

1. **Parse** the description to understand what to build
2. **Scaffold** in correct location:
   ```bash
   fazt app create servers/zyt/<name> --template vue-api
   ```
   **NEVER create apps in repo root!** Always use `servers/zyt/<name>` path.
3. **Customize fully**:
   - `index.html` with PWA meta, fonts, dark mode, styles
   - `src/lib/session.js` with URL-based 3-word sessions
   - `src/lib/api.js` with session in all requests
   - `src/lib/settings.js` with theme system
   - `src/main.js` with full Vue app including:
     - Settings panel (theme, sound, animations, session)
     - Escape key to close modals
     - Click outside to close modals
     - Sound effects
     - Proper light/dark mode
   - `api/main.js` with session-scoped data
4. **Validate**: `fazt app validate servers/zyt/<name>`
5. **Deploy to local**: `fazt app deploy servers/zyt/<name> --to local`
6. **STOP** - Present local URL to user: `http://<name>.192.168.64.3.nip.io:8080`
7. **Wait for approval** before production deploy
8. After approval: `fazt app deploy servers/zyt/<name> --to zyt`
9. Report production URL: `https://<name>.zyt.app`
