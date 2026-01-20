// Settings helper - persists user settings in localStorage

const SETTINGS_KEY = '{{.Name}}_settings'

const defaults = {
  theme: 'light',
  notifications: true
}

export function getSettings() {
  try {
    const stored = localStorage.getItem(SETTINGS_KEY)
    return stored ? { ...defaults, ...JSON.parse(stored) } : defaults
  } catch {
    return defaults
  }
}

export function saveSettings(settings) {
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings))
}

export function getSetting(key) {
  return getSettings()[key]
}

export function setSetting(key, value) {
  const settings = getSettings()
  settings[key] = value
  saveSettings(settings)
}
