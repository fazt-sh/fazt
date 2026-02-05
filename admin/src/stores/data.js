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

// Analytics stats data (events-based)
export const stats = map({
  total_events_today: 0,
  total_events_week: 0,
  total_events_month: 0,
  total_events_all_time: 0,
  events_by_source_type: {},
  top_domains: [],
  top_tags: []
})

// Auth state
export const auth = map({
  authenticated: false,
  user: null,
  loading: true,
  unauthorized: false // true when user lacks required role
})

// Current app detail
export const currentApp = map({
  id: null,
  title: null,
  description: null,
  visibility: null,
  source: null,
  source_url: null,
  file_count: 0,
  size_bytes: 0,
  created_at: null,
  updated_at: null,
  files: []
})

// Events data
export const events = list([])

// Logs data
export const activityLogs = map({
  entries: [],
  total: 0,
  showing: 0,
  offset: 0,
  limit: 20
})

// Logs stats
export const activityStats = map({
  total_count: 0,
  count_by_weight: {},
  oldest_entry: null,
  newest_entry: null,
  size_estimate_bytes: 0
})

// LEGACY_CODE: Old logs data (application logs)
export const logs = list([])

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
 * Load analytics stats from API (events-based)
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 */
export async function loadStats(client) {
  try {
    const data = await client.stats.analytics()
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
    apps.remove(app => app.id === id || app.title === id)
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
 * Load events from API
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 * @param {Object} options - Filter options
 */
export async function loadEvents(client, options = {}) {
  loading.setKey('events', true)
  try {
    const data = await client.events.list(options)
    events.set(data || [])
    error.set(null)
  } catch (err) {
    error.set(err.message)
    notify({ type: 'error', message: 'Failed to load events' })
  } finally {
    loading.setKey('events', false)
  }
}

/**
 * Load logs from API
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 * @param {Object} filters - Filter options (min_weight, max_weight, type, resource, app, alias, user, actor_type, action, result, since, until, limit, offset)
 */
export async function loadActivityLogs(client, filters = {}) {
  loading.setKey('activity-logs', true)
  try {
    const data = await client.logs.list(filters)
    activityLogs.set(data || { entries: [], total: 0, showing: 0, offset: 0, limit: 20 })
    error.set(null)
  } catch (err) {
    error.set(err.message)
    notify({ type: 'error', message: 'Failed to load logs' })
  } finally {
    loading.setKey('activity-logs', false)
  }
}

/**
 * Load logs stats from API
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 * @param {Object} filters - Same filters as loadActivityLogs
 */
export async function loadActivityStats(client, filters = {}) {
  try {
    const data = await client.logs.stats(filters)
    activityStats.set(data || {})
  } catch (err) {
    console.warn('Failed to load activity stats:', err.message)
  }
}

// LEGACY_CODE: Old logs function for application logs
export async function loadLogs(client, siteId, options = {}) {
  loading.setKey('logs', true)
  try {
    const data = await client.logs.list(siteId, options)
    logs.set(data?.logs || [])
    error.set(null)
  } catch (err) {
    error.set(err.message)
    notify({ type: 'error', message: 'Failed to load logs' })
  } finally {
    loading.setKey('logs', false)
  }
}

/**
 * Create app from template
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 * @param {string} name - App name
 * @param {string} template - Template ID
 */
export async function createApp(client, name, template = 'minimal') {
  try {
    const data = await client.apps.create(name, template)
    notify({ type: 'success', message: `App "${name}" created` })
    // Reload apps to include new one
    await loadApps(client)
    return data
  } catch (err) {
    notify({ type: 'error', message: err.message || 'Failed to create app' })
    throw err
  }
}

/**
 * Create alias
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 * @param {string} subdomain
 * @param {string} type
 * @param {Object} options
 */
export async function createAlias(client, subdomain, type, options = {}) {
  try {
    await client.aliases.create(subdomain, type, options)
    notify({ type: 'success', message: `Alias "${subdomain}" created` })
    await loadAliases(client)
    return true
  } catch (err) {
    notify({ type: 'error', message: err.message || 'Failed to create alias' })
    throw err
  }
}

/**
 * Update alias
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 * @param {string} subdomain
 * @param {Object} data
 */
export async function updateAlias(client, subdomain, data) {
  try {
    await client.aliases.update(subdomain, data)
    notify({ type: 'success', message: `Alias "${subdomain}" updated` })
    await loadAliases(client)
    return true
  } catch (err) {
    notify({ type: 'error', message: err.message || 'Failed to update alias' })
    throw err
  }
}

/**
 * Delete alias
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 * @param {string} subdomain
 */
export async function deleteAlias(client, subdomain) {
  try {
    await client.aliases.delete(subdomain)
    notify({ type: 'success', message: `Alias "${subdomain}" deleted` })
    await loadAliases(client)
    return true
  } catch (err) {
    notify({ type: 'error', message: err.message || 'Failed to delete alias' })
    throw err
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

    // Capture full current URL to return after re-login
    const currentUrl = window.location.href
    const redirectParam = encodeURIComponent(currentUrl)

    // Compute root domain for auth (dev auth lives on root domain)
    const parts = window.location.hostname.split('.')
    const rootDomain = parts.length > 2 && parts[0] !== 'www'
      ? parts.slice(1).join('.')
      : window.location.hostname

    // Redirect to auth login on root domain with redirect parameter
    // Use /auth/login (not /auth/dev/login) - it shows appropriate providers (Google in prod, dev locally)
    const loginUrl = `${window.location.protocol}//${rootDomain}${window.location.port ? ':' + window.location.port : ''}/auth/login?redirect=${redirectParam}`
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
