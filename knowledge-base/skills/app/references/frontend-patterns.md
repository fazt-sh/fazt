# Frontend Patterns

Common patterns for fazt app frontends using Vue 3.

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
