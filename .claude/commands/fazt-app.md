# /fazt-app - Build Fazt Apps

Build and deploy polished, PWA-ready apps to fazt instances with Claude.

## Usage

```
/fazt-app <description>
/fazt-app "build a pomodoro tracker with task persistence"
```

## Context

You are building an app for fazt.sh - a single-binary personal cloud.
Apps should feel native, polished, and delightful - like Apple apps.

---

## App Structure

```
my-app/
├── manifest.json        # Required: {"name": "my-app"}
├── index.html           # Entry point with PWA meta tags + import maps
├── main.js              # ES6 module entry
├── package.json         # Dev dependencies (Vite only)
├── vite.config.js       # Vite configuration
├── lib/
│   ├── session.js       # Session management
│   ├── settings.js      # Settings/preferences
│   ├── sounds.js        # Sound effects
│   └── theme.js         # Theme management
├── components/          # Vue components (plain JS, one per file)
│   ├── App.js           # Root component
│   ├── Header.js        # Navigation/header
│   └── SettingsPanel.js # Settings modal
├── api/                 # Serverless functions
│   └── main.js          # → /api/* routes
└── assets/              # Copied from defaults or user-provided
    ├── favicon.png
    ├── logo.png
    ├── logo.svg
    ├── apple-touch-icon.png
    └── social_preview.png
```

**Default Assets**: Copy from `.claude/fazt-assets/` unless user provides custom ones.
Only request custom assets if the user wants app-specific branding.

---

## Frontend Stack (Vue + Vite)

**Mandatory**: All fazt apps use Vue 3 with Vite. Apps must work in two modes:

1. **Static hosting** (import maps) - Direct deploy to fazt, no build step
2. **Vite dev/build** - Fast HMR in dev, optimized production builds

This dual-mode approach lets you iterate with Vite locally while deploying
the same source files directly to fazt static hosting.

### package.json (Development Only)

```json
{
  "name": "my-app",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "devDependencies": {
    "vite": "^5.0.0"
  }
}
```

**Note**: `vue` is NOT a dependency - it's loaded via import maps in production.

### vite.config.js

```javascript
import { defineConfig } from 'vite'

export default defineConfig({
  // Resolve bare 'vue' import to CDN version for consistency
  resolve: {
    alias: {
      'vue': 'https://unpkg.com/vue@3/dist/vue.esm-browser.js'
    }
  },
  // Don't bundle Vue - use CDN in both dev and prod
  build: {
    rollupOptions: {
      external: ['vue'],
      output: {
        paths: {
          vue: 'https://unpkg.com/vue@3/dist/vue.esm-browser.js'
        }
      }
    }
  },
  server: {
    port: 5173,
    // Proxy API calls to local fazt server
    proxy: {
      '/api': {
        target: 'http://my-app.192.168.64.3.nip.io:8080',
        changeOrigin: true
      }
    }
  }
})
```

### index.html

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

  <!-- SEO & Social -->
  <title>App Name</title>
  <meta name="description" content="App description here">
  <meta property="og:title" content="App Name">
  <meta property="og:description" content="App description here">
  <meta property="og:image" content="/assets/social_preview.png">
  <meta property="og:type" content="website">

  <!-- Icons -->
  <link rel="icon" type="image/png" href="/assets/favicon.png">
  <link rel="apple-touch-icon" href="/assets/apple-touch-icon.png">

  <!-- Fonts -->
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">

  <!-- Tailwind -->
  <script src="https://cdn.tailwindcss.com"></script>
  <script>
    tailwind.config = {
      darkMode: 'class',
      theme: {
        extend: {
          fontFamily: {
            sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
          }
        }
      }
    }
  </script>

  <!-- Lucide Icons -->
  <script src="https://unpkg.com/lucide@latest/dist/umd/lucide.min.js"></script>

  <!-- Import Map: Makes 'vue' resolve to CDN in static hosting mode -->
  <!-- Vite ignores this in dev/build and uses its own resolution -->
  <script type="importmap">
  {
    "imports": {
      "vue": "https://unpkg.com/vue@3/dist/vue.esm-browser.js"
    }
  }
  </script>

  <style>
    /* iOS safe areas */
    body {
      padding: env(safe-area-inset-top) env(safe-area-inset-right) env(safe-area-inset-bottom) env(safe-area-inset-left);
    }

    /* Prevent text selection on UI elements */
    .no-select {
      -webkit-user-select: none;
      user-select: none;
      -webkit-touch-callout: none;
    }

    /* Smooth touch feedback */
    .touch-feedback {
      -webkit-tap-highlight-color: transparent;
    }
    .touch-feedback:active {
      opacity: 0.7;
      transform: scale(0.98);
    }

    /* Typography scale */
    .text-display { font-size: 2.5rem; font-weight: 700; letter-spacing: -0.02em; }
    .text-title { font-size: 1.5rem; font-weight: 600; letter-spacing: -0.01em; }
    .text-headline { font-size: 1.125rem; font-weight: 600; }
    .text-body { font-size: 1rem; font-weight: 400; }
    .text-caption { font-size: 0.875rem; font-weight: 500; color: rgb(115 115 115); }
  </style>
</head>
<body class="font-sans antialiased bg-neutral-50 dark:bg-neutral-950 text-neutral-900 dark:text-neutral-100 min-h-screen">
  <div id="app"></div>
  <script type="module" src="main.js"></script>
</body>
</html>
```

### How It Works

| Mode | Vue Resolution | Tailwind | Lucide |
|------|----------------|----------|--------|
| Static (fazt) | Import map → CDN | CDN | CDN |
| Vite dev | Alias → CDN | CDN | CDN |
| Vite build | External → CDN | CDN | CDN |

The same `index.html` and `main.js` work in all three modes because:
- `import { createApp } from 'vue'` resolves to CDN everywhere
- No build step required for deployment
- Vite just provides HMR during development

---

## Session Management (3-Word IDs)

```javascript
// lib/session.js
const WORDS = [
  'cat', 'dog', 'fox', 'owl', 'bee', 'ant', 'elk', 'emu',
  'red', 'blue', 'gold', 'jade', 'mint', 'rose', 'sage', 'teal',
  'apple', 'berry', 'grape', 'lemon', 'mango', 'peach', 'plum',
  'cloud', 'river', 'stone', 'leaf', 'wave', 'star', 'moon', 'sun'
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
```

---

## Settings System

```javascript
// lib/settings.js
import { ref, watch } from 'vue'

const DEFAULTS = {
  displayName: '',
  theme: 'system',      // 'light', 'dark', 'system'
  soundEnabled: true,
  hapticEnabled: true,
  reducedMotion: false,
}

const settings = ref(loadSettings())

function loadSettings() {
  const stored = localStorage.getItem('app-settings')
  return stored ? { ...DEFAULTS, ...JSON.parse(stored) } : { ...DEFAULTS }
}

function saveSettings() {
  localStorage.setItem('app-settings', JSON.stringify(settings.value))
}

watch(settings, saveSettings, { deep: true })

export { settings }

export function updateSetting(key, value) {
  settings.value[key] = value
}
```

---

## Theme System

```javascript
// lib/theme.js
import { settings } from './settings.js'
import { watch } from 'vue'

// Theme presets (Apple-inspired)
const themes = {
  light: {
    bg: 'bg-neutral-50',
    bgCard: 'bg-white',
    text: 'text-neutral-900',
    textMuted: 'text-neutral-500',
    border: 'border-neutral-200',
    accent: 'bg-blue-500 text-white',
  },
  dark: {
    bg: 'bg-neutral-950',
    bgCard: 'bg-neutral-900',
    text: 'text-neutral-100',
    textMuted: 'text-neutral-400',
    border: 'border-neutral-800',
    accent: 'bg-blue-500 text-white',
  }
}

export function initTheme() {
  const applyTheme = () => {
    const pref = settings.value.theme
    const isDark = pref === 'dark' ||
      (pref === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)

    document.documentElement.classList.toggle('dark', isDark)

    // Update theme-color meta tag
    const themeColor = isDark ? '#0a0a0a' : '#ffffff'
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', themeColor)
  }

  applyTheme()
  watch(() => settings.value.theme, applyTheme)
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', applyTheme)
}

export { themes }
```

---

## Sound Effects

```javascript
// lib/sounds.js
import { settings } from './settings.js'

// Soft, pleasant sounds (base64 encoded tiny audio)
const sounds = {
  tap: createTone(800, 0.05, 'sine'),
  success: createTone(880, 0.1, 'sine'),
  error: createTone(220, 0.15, 'triangle'),
  toggle: createTone(600, 0.03, 'sine'),
}

function createTone(freq, duration, type) {
  return () => {
    if (!settings.value.soundEnabled) return

    const ctx = new (window.AudioContext || window.webkitAudioContext)()
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()

    osc.type = type
    osc.frequency.value = freq
    gain.gain.setValueAtTime(0.1, ctx.currentTime)
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + duration)

    osc.connect(gain)
    gain.connect(ctx.destination)
    osc.start()
    osc.stop(ctx.currentTime + duration)
  }
}

export function playSound(name) {
  sounds[name]?.()
}

// Haptic feedback for iOS
export function haptic(style = 'light') {
  if (!settings.value.hapticEnabled) return
  if ('vibrate' in navigator) {
    navigator.vibrate(style === 'light' ? 10 : style === 'medium' ? 20 : 30)
  }
}
```

---

## Settings Panel Component

```javascript
// components/SettingsPanel.js
import { settings, updateSetting } from '../lib/settings.js'
import { playSound } from '../lib/sounds.js'

export default {
  name: 'SettingsPanel',
  emits: ['close'],
  setup(props, { emit }) {
    const toggle = (key) => {
      updateSetting(key, !settings.value[key])
      playSound('toggle')
    }

    const close = () => {
      emit('close')
      playSound('tap')
    }

    return { settings, updateSetting, toggle, close, playSound }
  },
  template: `
    <div>
      <!-- Header -->
      <div class="sticky top-0 p-4 border-b border-neutral-200 dark:border-neutral-800
                  flex items-center justify-between bg-white dark:bg-neutral-900">
        <span class="text-headline">Settings</span>
        <button
          @click="close"
          class="p-2 rounded-full hover:bg-neutral-100 dark:hover:bg-neutral-800"
        >
          <i data-lucide="x" class="w-5 h-5"></i>
        </button>
      </div>

      <!-- Content -->
      <div class="p-6 space-y-6">
        <!-- Display Name -->
        <div class="space-y-2">
          <label class="text-caption">Display Name</label>
          <input
            type="text"
            v-model="settings.displayName"
            placeholder="Enter your name"
            class="w-full px-4 py-3 rounded-xl bg-neutral-100 dark:bg-neutral-800
                   border-0 focus:ring-2 focus:ring-blue-500 outline-none"
          >
        </div>

        <!-- Theme -->
        <div class="space-y-2">
          <label class="text-caption">Theme</label>
          <div class="grid grid-cols-3 gap-2">
            <button
              v-for="t in ['light', 'dark', 'system']"
              :key="t"
              @click="updateSetting('theme', t); playSound('tap')"
              class="px-4 py-3 rounded-xl capitalize transition-all touch-feedback"
              :class="settings.theme === t
                ? 'bg-blue-500 text-white'
                : 'bg-neutral-100 dark:bg-neutral-800'"
            >
              {{ t }}
            </button>
          </div>
        </div>

        <!-- Toggles -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span>Sound Effects</span>
            <button
              @click="toggle('soundEnabled')"
              class="w-12 h-7 rounded-full transition-colors touch-feedback"
              :class="settings.soundEnabled ? 'bg-blue-500' : 'bg-neutral-300 dark:bg-neutral-700'"
            >
              <div
                class="w-5 h-5 bg-white rounded-full shadow transition-transform mx-1"
                :class="settings.soundEnabled ? 'translate-x-5' : 'translate-x-0'"
              ></div>
            </button>
          </div>

          <div class="flex items-center justify-between">
            <span>Haptic Feedback</span>
            <button
              @click="toggle('hapticEnabled')"
              class="w-12 h-7 rounded-full transition-colors touch-feedback"
              :class="settings.hapticEnabled ? 'bg-blue-500' : 'bg-neutral-300 dark:bg-neutral-700'"
            >
              <div
                class="w-5 h-5 bg-white rounded-full shadow transition-transform mx-1"
                :class="settings.hapticEnabled ? 'translate-x-5' : 'translate-x-0'"
              ></div>
            </button>
          </div>
        </div>
      </div>
    </div>
  `
}
```

---

## Component Architecture

**Mandatory**: Components must be in separate `.js` files for maintainability.

### main.js (Entry Point)

```javascript
// main.js - Minimal entry point
import { createApp } from 'vue'
import { initTheme } from './lib/theme.js'
import App from './components/App.js'

// Initialize theme before mounting
initTheme()

// Mount the app
createApp(App).mount('#app')
```

### components/App.js (Root Component)

```javascript
// components/App.js
import { ref } from 'vue'
import { getSession } from '../lib/session.js'
import { settings } from '../lib/settings.js'
import { playSound, haptic } from '../lib/sounds.js'
import Header from './Header.js'
import SettingsPanel from './SettingsPanel.js'

export default {
  name: 'App',
  components: { Header, SettingsPanel },
  setup() {
    const sessionId = getSession()
    const showSettings = ref(false)

    const openSettings = () => {
      showSettings.value = true
      playSound('tap')
      haptic('light')
    }

    const closeSettings = () => {
      showSettings.value = false
      playSound('tap')
    }

    return {
      sessionId,
      settings,
      showSettings,
      openSettings,
      closeSettings,
      playSound,
      haptic
    }
  },
  template: `
    <div class="min-h-screen flex flex-col">
      <Header @open-settings="openSettings" />

      <!-- Main Content -->
      <main class="flex-1 p-4">
        <!-- App-specific content here -->
      </main>

      <!-- Settings Modal -->
      <Teleport to="body">
        <Transition name="modal">
          <div
            v-if="showSettings"
            class="fixed inset-0 z-50 flex items-end sm:items-center justify-center"
          >
            <div
              class="absolute inset-0 bg-black/50 backdrop-blur-sm"
              @click="closeSettings"
            ></div>
            <div class="relative w-full max-w-md bg-white dark:bg-neutral-900
                        rounded-t-3xl sm:rounded-3xl max-h-[80vh] overflow-auto">
              <SettingsPanel @close="closeSettings" />
            </div>
          </div>
        </Transition>
      </Teleport>
    </div>
  `,
  mounted() {
    lucide.createIcons()
  },
  updated() {
    lucide.createIcons()
  }
}
```

### components/Header.js

```javascript
// components/Header.js
export default {
  name: 'Header',
  emits: ['open-settings'],
  template: `
    <header class="sticky top-0 z-40 backdrop-blur-xl bg-white/80 dark:bg-neutral-950/80
                   border-b border-neutral-200 dark:border-neutral-800 no-select">
      <div class="px-4 h-14 flex items-center justify-between">
        <h1 class="text-headline">App Name</h1>
        <button
          @click="$emit('open-settings')"
          class="p-2 rounded-full hover:bg-neutral-100 dark:hover:bg-neutral-800 touch-feedback"
        >
          <i data-lucide="settings" class="w-5 h-5"></i>
        </button>
      </div>
    </header>
  `
}
```

---

## Data Persistence Protocol

**Critical**: Follow this protocol for ALL fazt apps with persistence.

### Core Principles

1. **Session = Identity**: The `?s=...` URL parameter is the user's identity
2. **Document per Session**: Store user data as a document in a `sessions` collection
3. **Server is Truth**: `fazt.storage` is the source of truth, not localStorage
4. **Queryable by Design**: Use document store so you can query across users

### Storage Primitives

Fazt provides three storage APIs, all auto-namespaced by `app_id`:

#### Document Store (`ds`) - Primary storage

Use for: session state, user data, app content, anything queryable.

```javascript
var ds = fazt.storage.ds;

ds.insert(collection, doc)           // Returns generated id
ds.find(collection, query)           // Returns array of docs
ds.findOne(collection, id)           // Returns single doc or null
ds.update(collection, query, changes)// Returns count updated
ds.delete(collection, query)         // Returns count deleted
```

Query operators: `$eq`, `$ne`, `$gt`, `$lt`, `$in`, `$contains`

#### Key-Value (`kv`) - Simple lookups

Use for: global config, counters, feature flags, caches.

```javascript
var kv = fazt.storage.kv;

kv.set(key, value)      // Store any JSON value
kv.set(key, value, ttl) // With TTL in milliseconds
kv.get(key)             // Returns value or undefined
kv.delete(key)          // Remove key
kv.list(prefix)         // Returns [{key, value, expiresAt?}, ...]
```

#### Blob Storage (`s3`) - Files

Use for: images, uploads, any binary content.

```javascript
var s3 = fazt.storage.s3;

s3.put(path, data, mimeType)  // data as string or base64
s3.get(path)                  // Returns {data (base64), mime, size, hash}
s3.delete(path)               // Remove file
s3.list(prefix)               // Returns [{path, mime, size, updatedAt}, ...]
```

### Session Document Structure

Store each user's state as a document in `sessions` collection:

```javascript
{
  sessionId: "fox-gold-river",    // The ?s= value

  // Identity
  displayName: "Alex",

  // Preferences
  settings: {
    theme: "dark",
    soundEnabled: true
  },

  // App-specific state
  current: { /* work in progress */ },
  history: [ /* past activity */ ],
  stats: { /* aggregates */ },

  // Metadata (auto-managed)
  createdAt: 1705708800000,
  updatedAt: 1705795200000
}
```

### API Pattern

```javascript
// api/main.js
function handler(req) {
  var ds = fazt.storage.ds;
  var path = req.path;

  // GET /api/state?session=xxx
  if (path === '/api/state' && req.method === 'GET') {
    var sessionId = req.query.session;
    if (!sessionId) return error(400, 'session required');

    var docs = ds.find('sessions', { sessionId: sessionId });
    var state = docs.length > 0 ? docs[0] : null;
    return json(state);
  }

  // POST /api/state {session, ...fields}
  if (path === '/api/state' && req.method === 'POST') {
    var body = req.body || {};
    var sessionId = body.session;
    if (!sessionId) return error(400, 'session required');

    var existing = ds.find('sessions', { sessionId: sessionId });

    if (existing.length > 0) {
      ds.update('sessions', { sessionId: sessionId }, {
        $set: { ...body.state, updatedAt: Date.now() }
      });
    } else {
      ds.insert('sessions', {
        sessionId: sessionId,
        ...body.state,
        createdAt: Date.now(),
        updatedAt: Date.now()
      });
    }
    return json({ ok: true });
  }

  // GET /api/leaderboard - Query across sessions
  if (path === '/api/leaderboard' && req.method === 'GET') {
    var all = ds.find('sessions', {});
    var sorted = all
      .filter(function(s) { return s.stats && s.stats.highScore; })
      .sort(function(a, b) { return b.stats.highScore - a.stats.highScore; })
      .slice(0, 10)
      .map(function(s) {
        return { displayName: s.displayName, score: s.stats.highScore };
      });
    return json(sorted);
  }

  return error(404, 'not found');
}

function json(data) {
  return { status: 200, headers: {'Content-Type':'application/json'}, body: JSON.stringify(data) };
}

function error(status, msg) {
  return { status: status, headers: {'Content-Type':'application/json'}, body: JSON.stringify({error: msg}) };
}

handler(request);
```

### Which Storage to Use

| Data | Primitive | Why |
|------|-----------|-----|
| Session/user state | `ds` | Queryable, structured |
| Global config | `kv` | Simple key lookup |
| Leaderboards | `ds` | Query from sessions |
| User uploads | `s3` | Binary content |
| Counters, flags | `kv` | Atomic, simple |

### Storage Boundaries

| Use | For |
|-----|-----|
| `fazt.storage.ds` | Session state, user data, app content |
| `fazt.storage.kv` | Global config, counters, caches |
| `fazt.storage.s3` | Files, images, binary uploads |
| `localStorage` | Device-only prefs (e.g., "don't show again") |
| URL `?s=` | Session identity only |

### Future: Relational (`rd`)

A 4th primitive for raw SQL (JOINs, complex aggregations) is spec'd but not built.
If you hit a wall with `ds` - needing multi-table joins or complex queries - ask
the user: "Is it time to build `fazt.storage.rd`?"

---

## Design Guidelines

### Apple-Esque UI Principles

- **Clarity**: Clean layouts, generous whitespace, readable typography
- **Deference**: UI recedes, content takes center stage
- **Depth**: Subtle shadows, layering, translucent materials

### Visual Style

- **Colors**: Neutral grays, single accent color (blue by default)
- **Corners**: Large radius (xl/2xl/3xl) for cards and buttons
- **Shadows**: Soft, diffused shadows
- **Blur**: Backdrop blur for overlays and headers

### Typography (Inter)

| Class | Use |
|-------|-----|
| `text-display` | Hero numbers, main stats |
| `text-title` | Section headings |
| `text-headline` | Card titles, nav items |
| `text-body` | Body text |
| `text-caption` | Labels, hints |

### Touch Interactions

- Large touch targets (min 44x44px)
- Visual feedback on tap (opacity/scale)
- Haptic feedback for important actions
- Smooth 60fps animations

### PWA Requirements

- Proper meta tags for iOS standalone mode
- Safe area padding for notch/home indicator
- Theme color matching light/dark mode
- Apple touch icon (180x180)

---

## Default Assets

Fazt default assets are located in `.claude/fazt-assets/`:

| File | Size | Purpose |
|------|------|---------|
| `favicon.png` | 32x32 | Browser tab icon |
| `apple-touch-icon.png` | 180x180 | iOS home screen |
| `logo.png` | 256x256 | In-app branding |
| `logo.svg` | vector | Scalable logo |
| `logo-transparent.png` | 256x256 | Logo without background |
| `social_preview.png` | 1200x630 | Open Graph sharing |

**Behavior**: Copy these defaults to `assets/` for every new app.
Only ask for custom assets if the user explicitly wants app-specific branding.

---

## CLI Reference

| Command | Purpose |
|---------|---------|
| `fazt app create <name>` | Scaffold from template |
| `fazt app list [peer]` | See deployed apps |
| `fazt app deploy <dir> --to local` | Deploy to local server |
| `fazt app deploy <dir> --to zyt` | Deploy to production |

---

## Local Development Workflow

**Always test locally first, then deploy to production.**

### Development Modes

| Mode | When to Use | HMR | API |
|------|-------------|-----|-----|
| Vite dev | Iterating on UI | Yes | Proxied to fazt |
| Direct deploy | Testing API / final check | No | Real fazt server |

### Option A: Vite Dev Server (Recommended for UI work)

```bash
cd servers/zyt/my-app

# Install Vite (first time only)
npm install

# Start Vite dev server
npm run dev
```

Access at: `http://localhost:5173`

Vite provides:
- Hot Module Replacement (instant UI updates)
- Better error messages
- Source maps for debugging

API calls are proxied to your local fazt server via `vite.config.js`.

### Option B: Direct Deploy (For API testing)

```bash
# Ensure local fazt server is running
fazt remote status local

# If not running, start it
fazt server start --port 8080 --domain 192.168.64.3 --db /tmp/fazt-local.db

# Deploy
fazt app deploy ./my-app --to local
```

Access at: `http://my-app.192.168.64.3.nip.io:8080`

### Debug with /_fazt/* Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/_fazt/info` | GET | App metadata, file count |
| `/_fazt/storage` | GET | List all storage (KV, DS) |
| `/_fazt/storage/:key` | GET | Get specific KV value |
| `/_fazt/logs` | GET | Recent serverless logs |
| `/_fazt/errors` | GET | Recent errors only |

```bash
# Check app info
curl http://my-app.192.168.64.3.nip.io:8080/_fazt/info

# Check storage state
curl http://my-app.192.168.64.3.nip.io:8080/_fazt/storage

# Check for errors
curl http://my-app.192.168.64.3.nip.io:8080/_fazt/errors
```

### Deploy to Production

Once working locally:

```bash
fazt app deploy ./my-app --to zyt
```

---

## Location Behavior

| Scenario | Location |
|----------|----------|
| In fazt repo | `servers/zyt/{app}/` |
| Not in fazt repo | `/tmp/fazt-{app}-{hash}/` |
| `--in <dir>` flag | `<dir>/{app}/` |

---

## Instructions

When the user invokes `/fazt-app`:

1. Parse the description to understand what app to build
2. Determine target location (fazt repo → `servers/zyt/`, else `/tmp/`)
3. Create the app with **mandatory Vue + Vite structure**:
   - `manifest.json` with app name
   - `package.json` with Vite dev dependency
   - `vite.config.js` with API proxy and Vue CDN alias
   - `index.html` with import maps, PWA meta tags, Inter font, Lucide, Tailwind
   - `main.js` minimal entry point (just imports and mounts App)
   - `components/` folder with granular Vue components:
     - `App.js` - root component
     - `Header.js` - navigation
     - `SettingsPanel.js` - settings modal
     - Additional app-specific components
   - `lib/` folder with session, settings, sounds, theme modules
   - `api/main.js` for serverless routes (if persistence needed)
   - `assets/` folder with defaults copied from `.claude/fazt-assets/`
4. Copy default assets: `cp .claude/fazt-assets/* <app>/assets/`
5. **Local-first workflow:**
   - Option A (UI iteration): `cd <app> && npm install && npm run dev`
   - Option B (API testing): `fazt app deploy <folder> --to local`
   - Debug with `/_fazt/info`, `/_fazt/storage`, `/_fazt/errors`
   - Fix any issues found
6. Deploy to production: `fazt app deploy <folder> --to zyt`
7. Report production URL: `https://{name}.zyt.app?s=cat-apple-tree`

### Component File Naming

- One component per file
- PascalCase names: `Header.js`, `GameBoard.js`, `SettingsPanel.js`
- Export default object with `name`, `template`, and optionally `emits`, `props`, `setup`
- Keep templates readable - break large templates into child components
