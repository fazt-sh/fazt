/**
 * Data Store
 * API data state
 */

import { list, map } from '../../packages/zap/index.js'
import { loading, error, notify } from './app.js'

// Apps data
export const apps = list([])

// Aliases data
export const aliases = list([])

// System health data
export const health = map({
  status: null,
  uptime_seconds: 0,
  version: null,
  mode: null,
  memory: null,
  database: null,
  runtime: null
})

// Stats data
export const stats = map({
  apps: 0,
  requests_24h: 0,
  storage_bytes: 0,
  uptime_percent: 0
})

// Auth state
export const auth = map({
  authenticated: false,
  user: null,
  loading: true
})

// Current app detail
export const currentApp = map({
  id: null,
  name: null,
  source: null,
  manifest: null,
  file_count: 0,
  size_bytes: 0,
  created_at: null,
  updated_at: null,
  files: []
})

/**
 * Load apps from API
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 */
export async function loadApps(client) {
  loading.setKey('apps', true)
  try {
    const data = await client.apps.list()
    apps.set(data || [])
    error.set(null)
  } catch (err) {
    error.set(err.message)
    notify({ type: 'error', message: 'Failed to load apps' })
  } finally {
    loading.setKey('apps', false)
  }
}

/**
 * Load aliases from API
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 */
export async function loadAliases(client) {
  loading.setKey('aliases', true)
  try {
    const data = await client.aliases.list()
    aliases.set(data || [])
    error.set(null)
  } catch (err) {
    error.set(err.message)
    notify({ type: 'error', message: 'Failed to load aliases' })
  } finally {
    loading.setKey('aliases', false)
  }
}

/**
 * Load system health from API
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 */
export async function loadHealth(client) {
  loading.setKey('system', true)
  try {
    const data = await client.system.health()
    health.set(data || {})
    error.set(null)
  } catch (err) {
    error.set(err.message)
    notify({ type: 'error', message: 'Failed to load system health' })
  } finally {
    loading.setKey('system', false)
  }
}

/**
 * Load stats from API
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 */
export async function loadStats(client) {
  try {
    const data = await client.stats.overview()
    stats.set(data || {})
  } catch (err) {
    // Stats are optional, don't show error
    console.warn('Failed to load stats:', err.message)
  }
}

/**
 * Load app detail
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 * @param {string} id
 */
export async function loadAppDetail(client, id) {
  loading.setKey('apps', true)
  try {
    const [appData, filesData] = await Promise.all([
      client.apps.get(id),
      client.apps.files(id).catch(() => [])
    ])
    currentApp.set({
      ...appData,
      files: filesData || []
    })
    error.set(null)
  } catch (err) {
    error.set(err.message)
    notify({ type: 'error', message: 'Failed to load app detail' })
  } finally {
    loading.setKey('apps', false)
  }
}

/**
 * Delete app
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 * @param {string} id
 */
export async function deleteApp(client, id) {
  try {
    await client.apps.delete(id)
    apps.remove(app => app.id === id || app.name === id)
    notify({ type: 'success', message: 'App deleted' })
    return true
  } catch (err) {
    notify({ type: 'error', message: err.message })
    return false
  }
}

/**
 * Load auth session
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 */
export async function loadAuth(client) {
  auth.setKey('loading', true)
  try {
    const data = await client.auth.session()
    console.log('[Auth] Session loaded:', data.authenticated ? data.user?.email : 'not authenticated')
    auth.set({
      authenticated: data.authenticated || false,
      user: data.user || null,
      loading: false
    })
  } catch (err) {
    console.log('[Auth] Session load failed:', err.message)
    auth.set({
      authenticated: false,
      user: null,
      loading: false
    })
  }
}

/**
 * Sign out
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 */
export async function signOut(client) {
  console.log('[Auth] Starting sign out...')

  try {
    const result = await client.auth.signOut()
    console.log('[Auth] Sign out response:', result)

    auth.set({
      authenticated: false,
      user: null,
      loading: false
    })

    notify({ type: 'success', message: 'Signed out successfully' })

    // Get auth domain for redirect
    const parts = window.location.hostname.split('.')
    let authDomain = window.location.hostname
    if (parts.length > 2 && parts[0] !== 'www') {
      authDomain = parts.slice(1).join('.')
    }

    // Redirect to auth login on root domain
    const loginUrl = `${window.location.protocol}//${authDomain}${window.location.port ? ':' + window.location.port : ''}/auth/dev/login`
    console.log('[Auth] Redirecting to:', loginUrl)

    setTimeout(() => {
      window.location.href = loginUrl
    }, 500)
  } catch (err) {
    console.error('[Auth] Sign out failed:', err)
    notify({ type: 'error', message: 'Failed to sign out: ' + err.message })
    throw err
  }
}
