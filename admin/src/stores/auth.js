import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useUIStore } from './ui.js'

export const useAuthStore = defineStore('auth', () => {
  const authenticated = ref(false)
  const user = ref(null)
  const loading = ref(true)
  const unauthorized = ref(false)

  const isAdmin = computed(() => {
    return user.value?.role === 'owner' || user.value?.role === 'admin'
  })

  async function loadSession(client) {
    loading.value = true
    try {
      const data = await client.auth.session()
      authenticated.value = data.authenticated || false
      user.value = data.user || null
    } catch (err) {
      console.log('[Auth] Session load failed:', err.message)
      authenticated.value = false
      user.value = null
    } finally {
      loading.value = false
    }
  }

  async function signOut(client) {
    try {
      await client.auth.signOut()
      authenticated.value = false
      user.value = null

      const ui = useUIStore()
      ui.notify({ type: 'success', message: 'Signed out successfully' })

      // Redirect to login
      const currentUrl = window.location.href
      const redirectParam = encodeURIComponent(currentUrl)
      const parts = window.location.hostname.split('.')
      const rootDomain = parts.length > 2 && parts[0] !== 'www'
        ? parts.slice(1).join('.')
        : window.location.hostname
      const loginUrl = `${window.location.protocol}//${rootDomain}${window.location.port ? ':' + window.location.port : ''}/auth/login?redirect=${redirectParam}`

      setTimeout(() => { window.location.href = loginUrl }, 500)
    } catch (err) {
      console.error('[Auth] Sign out failed:', err)
      const ui = useUIStore()
      ui.notify({ type: 'error', message: 'Failed to sign out: ' + err.message })
      throw err
    }
  }

  function checkAccess() {
    if (loading.value) return true
    if (!authenticated.value || !user.value) {
      // Redirect to login
      const authDomain = getAuthDomain()
      const currentUrl = window.location.href
      const authUrl = `${window.location.protocol}//${authDomain}${window.location.port ? ':' + window.location.port : ''}/auth/login?redirect=${encodeURIComponent(currentUrl)}`
      window.location.href = authUrl
      return false
    }
    if (!isAdmin.value) {
      unauthorized.value = true
      return false
    }
    return true
  }

  function getAuthDomain() {
    const parts = window.location.hostname.split('.')
    if (parts.length > 2 && parts[0] !== 'www') {
      return parts.slice(1).join('.')
    }
    return window.location.hostname
  }

  return {
    authenticated, user, loading, unauthorized, isAdmin,
    loadSession, signOut, checkAccess, getAuthDomain
  }
})
