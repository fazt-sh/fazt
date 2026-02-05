import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUIStore = defineStore('ui', () => {
  const theme = ref(localStorage.getItem('fazt-theme') || 'light')
  const palette = ref(localStorage.getItem('fazt-palette') || 'stone')
  const sidebarCollapsed = ref(false)
  const settingsPanelOpen = ref(false)
  const commandPaletteOpen = ref(false)
  const notifications = ref([])

  // Modal states
  const newAppModalOpen = ref(false)
  const createAliasModalOpen = ref(false)
  const editAliasModalOpen = ref(false)
  const editingAlias = ref(null)

  // Transient UI states (shared across components)
  const mobileMenuOpen = ref(false)
  const userDropdownOpen = ref(false)
  const notificationsOpen = ref(false)

  function toggleUserDropdown() {
    userDropdownOpen.value = !userDropdownOpen.value
    notificationsOpen.value = false
  }

  function toggleNotifications() {
    notificationsOpen.value = !notificationsOpen.value
    userDropdownOpen.value = false
  }

  function closeDropdowns() {
    userDropdownOpen.value = false
    notificationsOpen.value = false
  }

  function setTheme(newTheme) {
    theme.value = newTheme
    document.documentElement.classList.toggle('dark', newTheme === 'dark')
    localStorage.setItem('fazt-theme', newTheme)
  }

  function setPalette(newPalette) {
    const html = document.documentElement
    html.className = html.className.replace(/palette-\w+/g, '').trim()
    if (newPalette !== 'stone') {
      html.classList.add('palette-' + newPalette)
    }
    palette.value = newPalette
    localStorage.setItem('fazt-palette', newPalette)
  }

  function initTheme() {
    setTheme(theme.value)
    setPalette(palette.value)
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function notify(notification) {
    const id = Date.now()
    notifications.value.push({ ...notification, id })
    setTimeout(() => {
      notifications.value = notifications.value.filter(n => n.id !== id)
    }, 5000)
  }

  function openEditAliasModal(alias) {
    editingAlias.value = alias
    editAliasModalOpen.value = true
  }

  // Persistent UI state (panel collapse, etc)
  function getUIState(key, defaultValue = false) {
    try {
      const state = JSON.parse(localStorage.getItem('fazt.web.ui.state') || '{}')
      return state[key] !== undefined ? state[key] : defaultValue
    } catch {
      return defaultValue
    }
  }

  function setUIState(key, value) {
    try {
      const state = JSON.parse(localStorage.getItem('fazt.web.ui.state') || '{}')
      state[key] = value
      localStorage.setItem('fazt.web.ui.state', JSON.stringify(state))
    } catch (e) {
      console.error('Failed to save UI state:', e)
    }
  }

  // Mock mode
  const useMock = new URLSearchParams(window.location.search).get('mock') === 'true'

  return {
    theme, palette, sidebarCollapsed, settingsPanelOpen, commandPaletteOpen, notifications,
    newAppModalOpen, createAliasModalOpen, editAliasModalOpen, editingAlias,
    mobileMenuOpen, userDropdownOpen, notificationsOpen,
    setTheme, setPalette, initTheme, toggleSidebar, notify,
    openEditAliasModal, toggleUserDropdown, toggleNotifications, closeDropdowns,
    getUIState, setUIState,
    useMock
  }
})
