/**
 * App Store
 * Global application state
 */

import { atom, map } from '../../packages/zap/index.js'

// Theme state
export const theme = atom('light')
export const palette = atom('stone')

// Sidebar state
export const sidebarCollapsed = atom(false)

// User state
export const user = map({
  id: null,
  email: null,
  name: null,
  avatar: null,
  isLoggedIn: false
})

// Loading states
export const loading = map({
  apps: false,
  aliases: false,
  system: false,
  user: false
})

// Error state
export const error = atom(null)

// Notifications
export const notifications = atom([])

/**
 * Set theme
 * @param {'light' | 'dark'} newTheme
 */
export function setTheme(newTheme) {
  theme.set(newTheme)
  document.documentElement.classList.toggle('dark', newTheme === 'dark')
  localStorage.setItem('fazt-theme', newTheme)
}

/**
 * Set palette
 * @param {string} newPalette
 */
export function setPalette(newPalette) {
  const html = document.documentElement
  // Remove existing palette classes
  html.className = html.className.replace(/palette-\w+/g, '').trim()
  if (newPalette !== 'stone') {
    html.classList.add('palette-' + newPalette)
  }
  palette.set(newPalette)
  localStorage.setItem('fazt-palette', newPalette)
}

/**
 * Toggle sidebar
 */
export function toggleSidebar() {
  sidebarCollapsed.set(!sidebarCollapsed.get())
}

/**
 * Initialize theme from localStorage
 */
export function initTheme() {
  const savedTheme = localStorage.getItem('fazt-theme') || 'light'
  const savedPalette = localStorage.getItem('fazt-palette') || 'stone'
  setTheme(savedTheme)
  setPalette(savedPalette)
}

/**
 * Add notification
 * @param {{ type: 'success' | 'error' | 'warning' | 'info', message: string }} notification
 */
export function notify(notification) {
  const id = Date.now()
  notifications.set([...notifications.get(), { ...notification, id }])
  // Auto-remove after 5 seconds
  setTimeout(() => {
    notifications.set(notifications.get().filter(n => n.id !== id))
  }, 5000)
}
