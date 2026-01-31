/**
 * fazt-sdk HTTP Client
 * Core HTTP client for the Fazt API
 */

/**
 * Default fetch adapter
 * @param {string} url
 * @param {RequestInit} options
 * @returns {Promise<Response>}
 */
const defaultAdapter = (url, options) => fetch(url, options)

/**
 * Create HTTP client
 * @param {import('./types.js').ClientOptions} options
 */
export function createHttpClient(options = {}) {
  const {
    baseUrl = '',
    adapter = defaultAdapter,
    onError,
    onRequest,
    onResponse
  } = options

  /**
   * Make HTTP request
   * @param {string} path
   * @param {import('./types.js').RequestOptions} requestOptions
   * @returns {Promise<any>}
   */
  async function request(path, requestOptions = {}) {
    const url = baseUrl + path
    const config = {
      method: requestOptions.method || 'GET',
      headers: {
        'Content-Type': 'application/json',
        ...requestOptions.headers
      },
      credentials: 'same-origin', // Include cookies for auth
      signal: requestOptions.signal
    }

    if (requestOptions.body) {
      config.body = typeof requestOptions.body === 'string'
        ? requestOptions.body
        : JSON.stringify(requestOptions.body)
    }

    // Request interceptor
    if (onRequest) {
      const modified = await onRequest(url, config)
      if (modified) {
        Object.assign(config, modified)
      }
    }

    let response
    try {
      response = await adapter(url, config)
    } catch (err) {
      const error = { code: 'NETWORK_ERROR', message: err.message }
      if (onError) onError(error)
      throw error
    }

    // Response interceptor
    if (onResponse) {
      await onResponse(response)
    }

    // Parse response
    let data
    const contentType = response.headers.get('content-type')
    if (contentType && contentType.includes('application/json')) {
      data = await response.json()
    } else {
      data = await response.text()
    }

    if (!response.ok) {
      const error = {
        code: data?.error?.code || 'HTTP_ERROR',
        message: data?.error?.message || `HTTP ${response.status}`,
        status: response.status,
        details: data?.error?.details
      }
      if (onError) onError(error)
      throw error
    }

    // Fazt API wraps responses in { success: true, data: ... }
    return data?.data !== undefined ? data.data : data
  }

  return {
    /**
     * GET request
     * @param {string} path
     * @param {import('./types.js').RequestOptions} [options]
     */
    get: (path, options) => request(path, { ...options, method: 'GET' }),

    /**
     * POST request
     * @param {string} path
     * @param {Object} body
     * @param {import('./types.js').RequestOptions} [options]
     */
    post: (path, body, options) => request(path, { ...options, method: 'POST', body }),

    /**
     * PUT request
     * @param {string} path
     * @param {Object} body
     * @param {import('./types.js').RequestOptions} [options]
     */
    put: (path, body, options) => request(path, { ...options, method: 'PUT', body }),

    /**
     * DELETE request
     * @param {string} path
     * @param {import('./types.js').RequestOptions} [options]
     */
    delete: (path, options) => request(path, { ...options, method: 'DELETE' }),

    /**
     * PATCH request
     * @param {string} path
     * @param {Object} body
     * @param {import('./types.js').RequestOptions} [options]
     */
    patch: (path, body, options) => request(path, { ...options, method: 'PATCH', body })
  }
}
