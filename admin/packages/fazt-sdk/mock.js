/**
 * fazt-sdk Mock Adapter
 * Mock adapter for testing and development
 */

import apps from './fixtures/apps.json' with { type: 'json' }
import aliases from './fixtures/aliases.json' with { type: 'json' }
import health from './fixtures/health.json' with { type: 'json' }
import user from './fixtures/user.json' with { type: 'json' }

// Simulated delay
const delay = (ms = 100) => new Promise(resolve => setTimeout(resolve, ms))

// Route handlers
const routes = {
  'GET /api/apps': () => apps,
  'GET /api/apps/:id': (params) => {
    const app = apps.find(a => a.id === params.id || a.name === params.id)
    if (!app) throw { code: 'APP_NOT_FOUND', message: 'App not found', status: 404 }
    return app
  },
  'GET /api/apps/:id/files': (params) => {
    const app = apps.find(a => a.id === params.id || a.name === params.id)
    if (!app) throw { code: 'APP_NOT_FOUND', message: 'App not found', status: 404 }
    return [
      { path: 'index.html', size: 1234, mime_type: 'text/html' },
      { path: 'manifest.json', size: 89, mime_type: 'application/json' },
      { path: 'api/main.js', size: 567, mime_type: 'application/javascript' }
    ]
  },
  'DELETE /api/apps/:id': (params) => {
    const app = apps.find(a => a.id === params.id || a.name === params.id)
    if (!app) throw { code: 'APP_NOT_FOUND', message: 'App not found', status: 404 }
    return { message: 'App deleted', name: app.name }
  },
  'GET /api/aliases': () => aliases,
  'GET /api/aliases/:subdomain': (params) => {
    const alias = aliases.find(a => a.subdomain === params.subdomain)
    if (!alias) throw { code: 'ALIAS_NOT_FOUND', message: 'Alias not found', status: 404 }
    return alias
  },
  'POST /api/aliases': (_, body) => ({
    subdomain: body.subdomain,
    type: body.type || 'proxy',
    message: 'Alias created'
  }),
  'DELETE /api/aliases/:subdomain': (params) => {
    const alias = aliases.find(a => a.subdomain === params.subdomain)
    if (!alias) throw { code: 'ALIAS_NOT_FOUND', message: 'Alias not found', status: 404 }
    return { subdomain: params.subdomain, message: 'Alias deleted' }
  },
  'GET /api/system/health': () => health,
  'GET /api/system/config': () => ({
    version: '0.17.0',
    domain: 'zyt.app',
    env: 'dev',
    https: false,
    ntfy: false
  }),
  'GET /auth/session': () => ({ authenticated: true, user }),
  'POST /auth/logout': () => ({ message: 'Logged out' }),
  'GET /api/stats': () => ({
    apps: 24,
    requests_24h: 12847,
    storage_bytes: 2410000000,
    uptime_percent: 99.97
  })
}

/**
 * Parse route pattern and extract params
 * @param {string} pattern
 * @param {string} path
 * @returns {{ match: boolean, params: Object }}
 */
function matchRoute(pattern, path) {
  const patternParts = pattern.split('/')
  const pathParts = path.split('?')[0].split('/')

  if (patternParts.length !== pathParts.length) {
    return { match: false, params: {} }
  }

  const params = {}
  for (let i = 0; i < patternParts.length; i++) {
    if (patternParts[i].startsWith(':')) {
      params[patternParts[i].slice(1)] = pathParts[i]
    } else if (patternParts[i] !== pathParts[i]) {
      return { match: false, params: {} }
    }
  }

  return { match: true, params }
}

/**
 * Create mock adapter
 * @param {Object} [options]
 * @param {number} [options.delay] - Simulated delay in ms
 * @param {Object} [options.overrides] - Route overrides
 * @returns {Function}
 */
export function createMockAdapter(options = {}) {
  const { delay: delayMs = 100, overrides = {} } = options

  return async function mockAdapter(url, config) {
    await delay(delayMs)

    const method = (config.method || 'GET').toUpperCase()
    const path = new URL(url, 'http://localhost').pathname
    let body = null

    if (config.body) {
      try {
        body = JSON.parse(config.body)
      } catch {
        body = config.body
      }
    }

    // Find matching route
    const allRoutes = { ...routes, ...overrides }
    for (const [routeKey, handler] of Object.entries(allRoutes)) {
      const [routeMethod, routePattern] = routeKey.split(' ')
      if (routeMethod !== method) continue

      const { match, params } = matchRoute(routePattern, path)
      if (match) {
        try {
          const data = await handler(params, body)
          return new Response(JSON.stringify({ success: true, data }), {
            status: 200,
            headers: { 'Content-Type': 'application/json' }
          })
        } catch (err) {
          return new Response(JSON.stringify({
            success: false,
            error: {
              code: err.code || 'ERROR',
              message: err.message || 'Unknown error'
            }
          }), {
            status: err.status || 500,
            headers: { 'Content-Type': 'application/json' }
          })
        }
      }
    }

    // No route found
    return new Response(JSON.stringify({
      success: false,
      error: { code: 'NOT_FOUND', message: `No mock for ${method} ${path}` }
    }), {
      status: 404,
      headers: { 'Content-Type': 'application/json' }
    })
  }
}

/**
 * Default mock adapter instance
 */
export const mockAdapter = createMockAdapter()
