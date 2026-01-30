// Settings Management
// Theme and preferences with localStorage persistence

const STORAGE_KEY = 'app-settings'

const DEFAULTS = {
  theme: 'system',        // 'light', 'dark', 'system'
  soundEnabled: true,
  animationsEnabled: true
}

// Load settings from localStorage
export function loadSettings() {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      return { ...DEFAULTS, ...JSON.parse(stored) }
    }
  } catch (e) {
    console.warn('[settings] Failed to load:', e)
  }
  return { ...DEFAULTS }
}

// Save settings to localStorage
export function saveSettings(settings) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings))
  } catch (e) {
    console.warn('[settings] Failed to save:', e)
  }
}

// Apply theme to document
export function applyTheme(theme) {
  const isDark = theme === 'dark' ||
    (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)

  document.documentElement.classList.toggle('dark', isDark)
}

// Update a single setting
export function updateSetting(key, value) {
  const settings = loadSettings()
  settings[key] = value
  saveSettings(settings)

  if (key === 'theme') {
    applyTheme(value)
  }

  return settings
}
