/**
 * zap/router
 * Hash and history mode routing
 */

import { atom } from './atom.js'

/**
 * @typedef {Object} Route
 * @property {string} path - Route path pattern
 * @property {string} name - Route name
 * @property {Function} [component] - Component function
 * @property {Object} [meta] - Route metadata
 */

/**
 * @typedef {Object} RouterOptions
 * @property {'hash' | 'history'} [mode] - Routing mode (default: hash)
 * @property {Route[]} routes - Route definitions
 * @property {Function} [onNavigate] - Navigation callback
 */

/**
 * @typedef {Object} RouteMatch
 * @property {Route} route - Matched route
 * @property {Object<string, string>} params - Route params
 * @property {Object<string, string>} query - Query params
 * @property {string} path - Current path
 */

/**
 * Parse query string into object
 * @param {string} search
 * @returns {Object<string, string>}
 */
function parseQuery(search) {
  const params = new URLSearchParams(search)
  const query = {}
  for (const [key, value] of params) {
    query[key] = value
  }
  return query
}

/**
 * Match path against route pattern
 * @param {string} pattern
 * @param {string} path
 * @returns {{ match: boolean, params: Object<string, string> }}
 */
function matchPath(pattern, path) {
  // Exact match
  if (pattern === path) {
    return { match: true, params: {} }
  }

  const patternParts = pattern.split('/').filter(Boolean)
  const pathParts = path.split('/').filter(Boolean)

  // Length mismatch (unless pattern has catch-all)
  if (patternParts.length !== pathParts.length) {
    // Check for catch-all
    const lastPattern = patternParts[patternParts.length - 1]
    if (!lastPattern?.startsWith('*')) {
      return { match: false, params: {} }
    }
  }

  const params = {}

  for (let i = 0; i < patternParts.length; i++) {
    const patternPart = patternParts[i]
    const pathPart = pathParts[i]

    if (patternPart.startsWith(':')) {
      // Dynamic segment
      params[patternPart.slice(1)] = pathPart
    } else if (patternPart.startsWith('*')) {
      // Catch-all
      params[patternPart.slice(1) || 'rest'] = pathParts.slice(i).join('/')
      break
    } else if (patternPart !== pathPart) {
      // Mismatch
      return { match: false, params: {} }
    }
  }

  return { match: true, params }
}

/**
 * Create router
 * @param {RouterOptions} options
 */
export function createRouter(options) {
  const { mode = 'hash', routes, onNavigate } = options

  // Current route state
  const currentRoute = atom(/** @type {RouteMatch | null} */ (null))

  /**
   * Get current path from URL
   */
  function getCurrentPath() {
    if (mode === 'hash') {
      return window.location.hash.slice(1) || '/'
    }
    return window.location.pathname
  }

  /**
   * Get current query from URL
   */
  function getCurrentQuery() {
    return parseQuery(window.location.search)
  }

  /**
   * Match current URL to routes
   */
  function matchRoute() {
    const path = getCurrentPath()
    const query = getCurrentQuery()

    for (const route of routes) {
      const { match, params } = matchPath(route.path, path)
      if (match) {
        return { route, params, query, path }
      }
    }

    // No match - try to find 404 route
    const notFound = routes.find(r => r.path === '*' || r.name === '404')
    if (notFound) {
      return { route: notFound, params: {}, query, path }
    }

    return null
  }

  /**
   * Update current route
   */
  function updateRoute() {
    const match = matchRoute()
    currentRoute.set(match)
    if (onNavigate && match) {
      onNavigate(match)
    }
  }

  /**
   * Navigate to path
   * @param {string} path
   * @param {Object} [options]
   * @param {boolean} [options.replace]
   * @param {Object<string, string>} [options.query]
   */
  function push(path, options = {}) {
    const { replace = false, query } = options

    let url = path
    if (query && Object.keys(query).length > 0) {
      const search = new URLSearchParams(query).toString()
      url = `${path}?${search}`
    }

    if (mode === 'hash') {
      if (replace) {
        window.location.replace(`#${url}`)
      } else {
        window.location.hash = url
      }
    } else {
      if (replace) {
        window.history.replaceState(null, '', url)
      } else {
        window.history.pushState(null, '', url)
      }
      updateRoute()
    }
  }

  /**
   * Replace current history entry
   * @param {string} path
   * @param {Object} [options]
   */
  function replace(path, options = {}) {
    push(path, { ...options, replace: true })
  }

  /**
   * Go back in history
   */
  function back() {
    window.history.back()
  }

  /**
   * Go forward in history
   */
  function forward() {
    window.history.forward()
  }

  /**
   * Go to specific history entry
   * @param {number} delta
   */
  function go(delta) {
    window.history.go(delta)
  }

  // Listen to navigation events
  if (mode === 'hash') {
    window.addEventListener('hashchange', updateRoute)
  } else {
    window.addEventListener('popstate', updateRoute)
  }

  // Initial route match
  updateRoute()

  return {
    /** Current route (reactive) */
    current: currentRoute,

    /** Navigation methods */
    push,
    replace,
    back,
    forward,
    go,

    /** Route definitions */
    routes,

    /** Routing mode */
    mode,

    /**
     * Get route by name
     * @param {string} name
     */
    getRoute(name) {
      return routes.find(r => r.name === name)
    },

    /**
     * Generate path for named route
     * @param {string} name
     * @param {Object<string, string>} [params]
     */
    resolve(name, params = {}) {
      const route = routes.find(r => r.name === name)
      if (!route) return '/'

      let path = route.path
      for (const [key, value] of Object.entries(params)) {
        path = path.replace(`:${key}`, value)
      }
      return path
    },

    /**
     * Check if current route matches name
     * @param {string} name
     */
    isActive(name) {
      return currentRoute.get()?.route?.name === name
    },

    /**
     * Cleanup router
     */
    destroy() {
      if (mode === 'hash') {
        window.removeEventListener('hashchange', updateRoute)
      } else {
        window.removeEventListener('popstate', updateRoute)
      }
    }
  }
}
