/**
 * fazt-sdk Admin Namespace
 * API surface for the Fazt Admin UI
 */

/**
 * Create admin API namespace
 * @param {import('./client.js').createHttpClient} http
 */
export function createAdminNamespace(http) {
  return {
    /**
     * Auth / User endpoints
     */
    auth: {
      /** Get current session */
      session: () => http.get('/auth/session'),
      /** Sign out */
      signOut: () => http.post('/auth/logout', {})
    },

    /**
     * Apps endpoints
     */
    apps: {
      /** List all apps (includes unlisted apps in admin context) */
      list: () => http.get('/api/apps', { params: { all: 'true' } }),

      /** Get app by ID or name */
      get: (id) => http.get(`/api/apps/${id}`),

      /** Get app files */
      files: (id) => http.get(`/api/apps/${id}/files`),

      /** Get app source info */
      source: (id) => http.get(`/api/apps/${id}/source`),

      /** Get file content */
      file: (id, path) => http.get(`/api/apps/${id}/files/${path}`),

      /** Delete app */
      delete: (id) => http.delete(`/api/apps/${id}`),

      /** Create app from template */
      create: (name, template = 'minimal') =>
        http.post('/api/apps/create', { name, template }),

      /** Install app from git URL */
      install: (url, name) =>
        http.post('/api/apps/install', { url, name })
    },

    /**
     * Aliases endpoints
     */
    aliases: {
      /** List all aliases */
      list: () => http.get('/api/aliases'),

      /** Get alias by subdomain */
      get: (subdomain) => http.get(`/api/aliases/${subdomain}`),

      /** Create alias */
      create: (subdomain, type, options = {}) =>
        http.post('/api/aliases', { subdomain, type, ...options }),

      /** Update alias */
      update: (subdomain, data) =>
        http.put(`/api/aliases/${subdomain}`, data),

      /** Delete alias */
      delete: (subdomain) => http.delete(`/api/aliases/${subdomain}`),

      /** Reserve subdomain */
      reserve: (subdomain) =>
        http.post(`/api/aliases/${subdomain}/reserve`),

      /** Swap two aliases */
      swap: (alias1, alias2) =>
        http.post('/api/aliases/swap', { alias1, alias2 }),

      /** Configure traffic split */
      split: (subdomain, targets) =>
        http.post(`/api/aliases/${subdomain}/split`, { targets })
    },

    /**
     * System endpoints
     */
    system: {
      /** Get health status */
      health: () => http.get('/api/system/health'),

      /** Get server config */
      config: () => http.get('/api/system/config'),

      /** Get resource limits (nested: hardware, storage, runtime, capacity, net) */
      limits: () => http.get('/api/system/limits'),

      /** Get limits schema (labels, descriptions, ranges for admin UI) */
      limitsSchema: () => http.get('/api/system/limits/schema'),

      /** Get VFS cache stats */
      cache: () => http.get('/api/system/cache'),

      /** Get database stats */
      db: () => http.get('/api/system/db'),

      /** @deprecated Use limits() instead — returns same nested data */
      capacity: () => http.get('/api/system/capacity')
    },

    /**
     * Stats endpoints
     */
    stats: {
      /** Get analytics stats (events-based) */
      analytics: () => http.get('/api/stats'),

      /** Get app-specific stats */
      app: (id) => http.get(`/api/stats/apps/${id}`)
    },

    /**
     * Templates endpoints
     */
    templates: {
      /** List available templates */
      list: () => http.get('/api/templates')
    },

    /**
     * Events endpoints
     */
    events: {
      /** List events with optional filters */
      list: (options = {}) => http.get('/api/events', { params: options })
      // options: { domain, tags, source_type, event_type, limit, offset }
    },

    /**
     * Logs endpoints
     */
    logs: {
      /** Query logs with filters */
      list: (filters = {}) => http.get('/api/system/logs', { params: filters }),
      // filters: { min_weight, max_weight, type, resource, app, alias, user, actor_type, action, result, since, until, limit, offset }

      /** Get activity log statistics */
      stats: (filters = {}) => http.get('/api/system/logs/stats', { params: filters })
      // filters: same as list()
    }
  }
}
