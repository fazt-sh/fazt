# Frontend Patterns

Common patterns for fazt app frontends using Vue 3.

## Component Granularity (CRITICAL)

**Every component MUST have a single root element.** Vue 3 supports fragments (multiple root elements) but they cause rendering crashes when combined with reactive store updates. This is not theoretical — it's a proven production issue.

### Rules

1. **Single root `<div>` per component** — never use multi-root fragments
2. **One concern per component** — sidebar, header, modal, palette = separate files
3. **Pinia stores for shared state** — no prop drilling for widely-used state (theme, auth, counts)
4. **Each component refreshes its own icons** — `onMounted(() => refreshIcons())` and `onUpdated(() => refreshIcons())` (see [External DOM Mutation](#external-dom-mutation-critical) for how `refreshIcons` works safely)

### Why No Fragments

Vue's fragment patcher tracks start/end anchors for each root element. When multiple `v-if` roots toggle during concurrent reactive updates (common with Pinia stores loading data), the anchors desync → `insertBefore` null → crash. Single root = no fragment patching = no crash.

### App Shell Pattern

```javascript
// App.js — thin orchestrator, single root div
export default {
  components: { Sidebar, HeaderBar, CommandPalette, SettingsPanel },
  setup() {
    // Initialize stores once here
    // Global keyboard shortcuts here
    return { ui, auth }
  },
  template: `
    <div>
      <SettingsPanel />
      <CommandPalette />
      <div class="flex h-screen">
        <Sidebar />
        <main class="flex-1 flex flex-col">
          <HeaderBar />
          <router-view />
        </main>
      </div>
    </div>
  `
}
```

### Component File Pattern

```javascript
// components/Sidebar.js — owns its own store access and icon refresh
import { computed, onMounted, onUpdated } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppsStore } from '../stores/apps.js'
import { refreshIcons } from '../lib/icons.js'

export default {
  name: 'AppSidebar',
  setup() {
    const apps = useAppsStore()
    const appCount = computed(() => apps.items.length)  // resolve in computed, not template
    onMounted(() => refreshIcons())
    onUpdated(() => refreshIcons())
    return { appCount }
  },
  template: `
    <aside class="sidebar">
      <span>{{ appCount }} apps</span>
    </aside>
  `
}
```

### Anti-patterns

```javascript
// BAD: Multiple root elements with v-if (fragment)
template: `
  <div v-if="showA">...</div>
  <div v-if="showB">...</div>
  <div class="main">...</div>
`

// GOOD: Single root, children handle their own visibility
template: `
  <div>
    <ModalA />
    <ModalB />
    <div class="main">...</div>
  </div>
`

// BAD: Complex expressions in template strings
template: `<span>{{ store.data?.nested?.value || 'default' }}</span>`

// GOOD: Resolve in computed
const displayValue = computed(() => store.data?.nested?.value || 'default')
template: `<span>{{ displayValue }}</span>`
```

## External DOM Mutation (CRITICAL)

**Never use libraries that REPLACE DOM elements Vue owns.** This is the #1 source of `insertBefore` null / `setElementText` null crashes in Vue apps.

### The Rule

If a library's API replaces or removes DOM elements, it will break Vue's virtual DOM patcher. Vue keeps references to real DOM nodes; when they vanish, the next reactive update crashes.

### Common Offenders

- `lucide.createIcons()` — replaces `<i data-lucide>` with `<svg>`
- `highlight.js` auto-mode — replaces `<code>` elements
- Any jQuery-style `.replaceWith()` or `.html()` on Vue-managed elements

### The Fix: Inject Inside, Never Replace

Instead of letting a library replace an element, **inject content inside it** so Vue's reference to the outer element stays valid:

```javascript
// lib/icons.js — canonical pattern for external icon libraries
import { nextTick } from 'vue'

let pending = false

export function refreshIcons() {
  if (pending) return
  pending = true
  nextTick(() => {
    pending = false
    document.querySelectorAll('i[data-lucide]').forEach(el => {
      if (el.querySelector('svg')) return  // already rendered
      const name = el.getAttribute('data-lucide')
      if (!name) return
      const svg = createIconSvg(name)
      if (!svg) return
      // Inject INSIDE — Vue still owns the <i> element
      el.style.display = 'inline-flex'
      el.style.alignItems = 'center'
      el.style.justifyContent = 'center'
      el.appendChild(svg)
    })
  })
}
```

### Why This Works

1. Vue's VNode references the `<i>` element — it never gets removed
2. The `<i>` has no VNode children, so Vue won't touch the injected SVG
3. `nextTick` + debounce ensures one scan per render cycle
4. The `el.querySelector('svg')` guard prevents double-injection

### Why `lucide.createIcons()` Crashes Vue

```
Template: <i data-lucide="heart"> → Vue VNode references this <i>
createIcons() runs          → replaces <i> with <svg> in DOM
Store update triggers render → Vue tries insertBefore on <i>
<i> is gone                 → insertBefore(null) → crash
```

This is NOT a Vue bug or a BFBB pattern bug. It's a fundamental conflict between external DOM mutation and any VDOM framework (React, Vue, Svelte would all crash).

## Project Structure

```
my-app/
├── index.html           # Entry point, PWA meta, imports
├── vite.config.js       # Vite configuration
├── package.json
├── scripts/
│   └── version.js       # Version generator (run at build)
├── src/
│   ├── main.js          # Vue app initialization
│   ├── App.vue          # Root component
│   ├── pages/           # Page components
│   ├── components/      # Reusable components
│   ├── stores/          # Pinia stores
│   └── lib/             # Utilities
│       ├── api.js       # API client
│       ├── session.js   # URL session management
│       ├── settings.js  # Local settings
│       └── auth.js      # Authentication
├── api/
│   └── main.js          # Serverless API
├── private/             # Server-only files (auth-gated HTTP, serverless access)
│   ├── config.json      # App configuration
│   └── seed-data.json   # Initial/seed data
├── version.json         # Generated at build time
└── dist/                # Built output (deploy this)
```

## Session Management

URL-based sessions for shareable state.

```javascript
// src/lib/session.js
const WORDS = [
  'cat', 'dog', 'fox', 'owl', 'bee', 'ant', 'elk', 'bat',
  'red', 'blue', 'gold', 'jade', 'mint', 'rose', 'sage',
  'apple', 'berry', 'grape', 'lemon', 'mango', 'peach',
  'cloud', 'river', 'stone', 'leaf', 'wave', 'star'
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

export function newSession() {
  const id = generateSessionId()
  history.replaceState(null, '', `?s=${id}`)
  return id
}
```

## API Client

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

## Settings System

Local settings with theme support.

```javascript
// src/lib/settings.js
const SETTINGS_KEY = 'appname_settings'

const DEFAULTS = {
  theme: 'system',
  soundEnabled: true,
  animationsEnabled: true
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
    (theme === 'system' && matchMedia('(prefers-color-scheme: dark)').matches)
  document.documentElement.classList.toggle('dark', isDark)
}

// Initialize on load
applyTheme(getSettings().theme)

// React to system changes
matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
  if (getSettings().theme === 'system') applyTheme('system')
})
```

## Sound Effects

```javascript
// src/lib/sounds.js
export function playSound(type, enabled = true) {
  if (!enabled) return
  try {
    const ctx = new AudioContext()
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    osc.connect(gain)
    gain.connect(ctx.destination)

    const sounds = {
      success: () => {
        osc.frequency.setValueAtTime(523, ctx.currentTime)
        osc.frequency.setValueAtTime(659, ctx.currentTime + 0.1)
        osc.frequency.setValueAtTime(784, ctx.currentTime + 0.2)
        gain.gain.setValueAtTime(0.3, ctx.currentTime)
        gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.4)
        osc.start()
        osc.stop(ctx.currentTime + 0.4)
      },
      tap: () => {
        osc.frequency.setValueAtTime(440, ctx.currentTime)
        gain.gain.setValueAtTime(0.15, ctx.currentTime)
        gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.1)
        osc.start()
        osc.stop(ctx.currentTime + 0.1)
      },
      error: () => {
        osc.frequency.setValueAtTime(200, ctx.currentTime)
        gain.gain.setValueAtTime(0.2, ctx.currentTime)
        gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.2)
        osc.start()
        osc.stop(ctx.currentTime + 0.2)
      }
    }

    if (sounds[type]) sounds[type]()
  } catch (e) {}
}
```

## Modal Pattern

```vue
<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-50" @keydown.esc="close">
      <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="close" />
      <div class="absolute inset-4 sm:inset-auto sm:top-1/2 sm:left-1/2 sm:-translate-x-1/2 sm:-translate-y-1/2 sm:max-w-md sm:w-full">
        <div class="bg-white dark:bg-neutral-900 rounded-2xl shadow-xl p-6">
          <slot />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'

const props = defineProps(['show'])
const emit = defineEmits(['close'])

function close() {
  emit('close')
}

onMounted(() => {
  const handler = (e) => { if (e.key === 'Escape') close() }
  document.addEventListener('keydown', handler)
  onUnmounted(() => document.removeEventListener('keydown', handler))
})
</script>
```

## Keyboard Shortcuts

```javascript
// In setup()
onMounted(() => {
  const handler = (e) => {
    // Global shortcuts
    if (e.key === 'Escape') closeModal()
    if (e.key === '?' && !e.target.matches('input, textarea')) showHelp()

    // Ctrl/Cmd shortcuts
    if ((e.ctrlKey || e.metaKey) && e.key === 's') {
      e.preventDefault()
      save()
    }
  }

  document.addEventListener('keydown', handler)
  onUnmounted(() => document.removeEventListener('keydown', handler))
})
```

## Vite Configuration

```javascript
// vite.config.js
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: 'dist',
    assetsDir: 'static'
  }
})
```

## Package.json

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "private": true,
  "scripts": {
    "dev": "vite",
    "version": "node scripts/version.js",
    "build": "node scripts/version.js && vite build && cp version.json dist/",
    "preview": "vite preview"
  },
  "dependencies": {
    "vue": "^3.4"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0",
    "autoprefixer": "^10.4",
    "postcss": "^8.4",
    "tailwindcss": "^3.4",
    "vite": "^5.0"
  }
}
```

**Note**: The build script generates `version.json` with git hash and timestamp.
See [deployment.md](../fazt/deployment.md#app-versioning) for details.
