/**
 * fazt-sdk App Namespace
 * API surface for fazt-apps (user-facing applications)
 */

/**
 * Create app API namespace
 * @param {ReturnType<import('./client.js').createHttpClient>} http
 */
export function createAppNamespace(http) {
  return {
    /**
     * Auth — standard for every fazt-app
     */
    auth: {
      /** Get current user session */
      me: () => http.get('/api/me'),

      /** Redirect to login page */
      login: (redirect) => {
        const params = redirect ? `?redirect=${encodeURIComponent(redirect)}` : ''
        window.location.href = `/auth/login${params}`
      },

      /** Sign out */
      logout: () => http.post('/auth/logout', {})
    },

    /** Direct HTTP access for app-specific endpoints */
    http,

    /**
     * Upload a file with progress tracking
     * @param {File} file - File to upload
     * @param {string} [url='/api/upload'] - Upload endpoint
     * @param {import('./types.js').UploadOptions & { field?: string }} [options]
     * @returns {Promise<any>}
     */
    upload(file, url = '/api/upload', options = {}) {
      const { field = 'file', onProgress, signal } = options
      const formData = new FormData()
      formData.append(field, file)
      return http.upload(url, formData, { onProgress, signal })
    },

    /**
     * Paginate through results using offset-based pagination
     * @param {string} url - API endpoint
     * @param {import('./types.js').PaginateOptions} [options]
     * @returns {import('./types.js').PaginatedResult}
     */
    paginate(url, options = {}) {
      const { limit = 50, params = {} } = options
      let offset = 0
      let done = false

      return {
        /** Fetch next page of results */
        async next() {
          if (done) return { items: [], done: true }
          const data = await http.get(url, {
            params: { ...params, limit, offset }
          })
          const items = Array.isArray(data) ? data : data?.entries || data?.items || []
          if (items.length < limit) done = true
          offset += items.length
          return { items, done: items.length < limit }
        },

        /** Reset pagination to start */
        reset() {
          offset = 0
          done = false
        }
      }
    }
  }
}
