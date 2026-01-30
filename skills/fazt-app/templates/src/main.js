// Main App Entry Point
// Works raw (served directly) AND built (via Vite/Bun)

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { router } from './router.js'
import { App } from './App.js'
import { applyTheme, loadSettings } from './lib/settings.js'

// Initialize app
const app = createApp(App)

// State management
const pinia = createPinia()
app.use(pinia)

// Routing
app.use(router)

// Mount
app.mount('#app')

// Apply saved theme
const settings = loadSettings()
applyTheme(settings.theme)

// Listen for system theme changes
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
  const settings = loadSettings()
  if (settings.theme === 'system') {
    applyTheme('system')
  }
})

// Debug: expose app for devtools
if (import.meta.env?.DEV || window.location.hostname.includes('192.168')) {
  window.__fazt_app__ = app
  window.__fazt_pinia__ = pinia
}
