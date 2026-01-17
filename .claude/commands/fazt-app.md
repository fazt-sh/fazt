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
├── manifest.json      # Required: {"name": "my-app"}
├── index.html         # Entry point with PWA meta tags
├── main.js            # ES6 module entry
├── lib/
│   ├── session.js     # Session management
│   ├── settings.js    # Settings/preferences
│   ├── sounds.js      # Sound effects
│   └── theme.js       # Theme management
├── components/        # Vue components (plain JS)
├── api/               # Serverless functions
│   └── data.js        # → GET/POST /api/data
└── assets/            # Copied from defaults or user-provided
    ├── favicon.png
    ├── logo.png
    ├── logo.svg
    ├── apple-touch-icon.png
    └── social_preview.png
```

**Default Assets**: Copy from `.claude/fazt-assets/` unless user provides custom ones.
Only request custom assets if the user wants app-specific branding.

---

## Frontend Stack (Zero Build)

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
          },
          colors: {
            // Theme colors injected by theme.js
          }
        }
      }
    }
  </script>

  <!-- Lucide Icons -->
  <script src="https://unpkg.com/lucide@latest/dist/umd/lucide.min.js"></script>

  <!-- Import Map -->
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
  setup() {
    const toggle = (key) => {
      updateSetting(key, !settings.value[key])
      playSound('toggle')
    }

    return { settings, updateSetting, toggle }
  },
  template: `
    <div class="p-6 space-y-6">
      <h2 class="text-title">Settings</h2>

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
  `
}
```

---

## Main App Template

```javascript
// main.js
import { createApp, ref } from 'vue'
import { getSession, getSessionUrl } from './lib/session.js'
import { settings } from './lib/settings.js'
import { initTheme } from './lib/theme.js'
import { playSound, haptic } from './lib/sounds.js'
import SettingsPanel from './components/SettingsPanel.js'

// Initialize
initTheme()
const sessionId = getSession()

const App = {
  components: { SettingsPanel },
  setup() {
    const showSettings = ref(false)

    const openSettings = () => {
      showSettings.value = true
      playSound('tap')
      haptic('light')
    }

    return {
      sessionId,
      settings,
      showSettings,
      openSettings,
      playSound,
      haptic
    }
  },
  template: `
    <div class="min-h-screen flex flex-col">
      <!-- Header -->
      <header class="sticky top-0 z-40 backdrop-blur-xl bg-white/80 dark:bg-neutral-950/80
                     border-b border-neutral-200 dark:border-neutral-800 no-select">
        <div class="px-4 h-14 flex items-center justify-between">
          <h1 class="text-headline">App Name</h1>
          <button
            @click="openSettings"
            class="p-2 rounded-full hover:bg-neutral-100 dark:hover:bg-neutral-800 touch-feedback"
          >
            <i data-lucide="settings" class="w-5 h-5"></i>
          </button>
        </div>
      </header>

      <!-- Main Content -->
      <main class="flex-1 p-4">
        <!-- App content here -->
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
              @click="showSettings = false"
            ></div>
            <div class="relative w-full max-w-md bg-white dark:bg-neutral-900
                        rounded-t-3xl sm:rounded-3xl max-h-[80vh] overflow-auto
                        safe-area-bottom">
              <div class="sticky top-0 p-4 border-b border-neutral-200 dark:border-neutral-800
                          flex items-center justify-between bg-white dark:bg-neutral-900">
                <span class="text-headline">Settings</span>
                <button
                  @click="showSettings = false; playSound('tap')"
                  class="p-2 rounded-full hover:bg-neutral-100 dark:hover:bg-neutral-800"
                >
                  <i data-lucide="x" class="w-5 h-5"></i>
                </button>
              </div>
              <SettingsPanel />
            </div>
          </div>
        </Transition>
      </Teleport>
    </div>
  `,
  mounted() {
    // Initialize Lucide icons
    lucide.createIcons()
  },
  updated() {
    lucide.createIcons()
  }
}

createApp(App).mount('#app')
```

---

## Serverless Functions

```javascript
// api/data.js
function handler(req) {
  const db = fazt.storage.kv
  const params = new URLSearchParams(req.url.split('?')[1] || '')
  const session = params.get('session')

  if (!session) {
    return { status: 400, body: JSON.stringify({ error: 'session required' }) }
  }

  const key = `data:${session}`

  if (req.method === 'GET') {
    const data = db.get(key) || {}
    return {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    }
  }

  if (req.method === 'POST') {
    const data = JSON.parse(req.body)
    db.set(key, data)
    return {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ success: true })
    }
  }

  return { status: 405, body: 'Method not allowed' }
}
```

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
| `fazt app create <name> --template vite` | Scaffold Vite app |
| `fazt app list [peer]` | See deployed apps |
| `fazt app deploy <dir> --to <peer>` | Deploy app |
| `fazt app deploy <dir> --to <peer> --no-build` | Deploy without building |

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
3. Create the app with:
   - `manifest.json` with app name
   - `index.html` with full PWA meta tags, Inter font, Lucide, Tailwind
   - `lib/` folder with session, settings, sounds, theme modules
   - `main.js` with Vue app, settings panel, proper structure
   - `components/` for any needed components
   - `api/` endpoints if persistence needed
   - `assets/` folder with defaults copied from `.claude/fazt-assets/`
4. Copy default assets: `cp .claude/fazt-assets/* <app>/assets/`
5. Deploy: `fazt app deploy <folder> --to zyt`
6. Report URL with session: `https://{name}.zyt.app?s=cat-apple-tree`
