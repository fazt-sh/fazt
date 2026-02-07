/**
 * fazt-sdk HTTP Client
 * Core HTTP client for the Fazt API
 */

import { ApiError } from './errors.js'

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

  const hasCustomAdapter = adapter !== defaultAdapter

  /**
   * Make HTTP request
   * @param {string} path
   * @param {import('./types.js').RequestOptions} requestOptions
   * @returns {Promise<any>}
   */
  async function request(path, requestOptions = {}) {
    // Build URL with query params
    let url = baseUrl + path
    if (requestOptions.params) {
      const searchParams = new URLSearchParams()
      for (const [key, value] of Object.entries(requestOptions.params)) {
        if (value !== undefined && value !== null) {
          searchParams.append(key, String(value))
        }
      }
      const queryString = searchParams.toString()
      if (queryString) {
        url += (url.includes('?') ? '&' : '?') + queryString
      }
    }

    const isFormData = typeof FormData !== 'undefined' && requestOptions.body instanceof FormData

    const config = {
      method: requestOptions.method || 'GET',
      headers: {
        ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
        'Accept': 'application/json',
        ...requestOptions.headers
      },
      credentials: 'include',
      signal: requestOptions.signal
    }

    if (requestOptions.body) {
      config.body = isFormData
        ? requestOptions.body
        : typeof requestOptions.body === 'string'
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
      const error = new ApiError(0, 'NETWORK_ERROR', err.message)
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
      const error = new ApiError(
        response.status,
        data?.error?.code || 'HTTP_ERROR',
        data?.error?.message || `HTTP ${response.status}`,
        data?.error?.details
      )
      if (onError) onError(error)
      throw error
    }

    // Fazt API wraps responses in { success: true, data: ... }
    return data?.data !== undefined ? data.data : data
  }

  /**
   * Upload with progress tracking via XHR
   * Falls back to fetch adapter when custom adapter is set (e.g. mock mode)
   * @param {string} path
   * @param {FormData} formData
   * @param {import('./types.js').UploadOptions} [uploadOptions]
   * @returns {Promise<any>}
   */
  function upload(path, formData, uploadOptions = {}) {
    const { onProgress, signal } = uploadOptions
    const url = baseUrl + path

    // Fall back to adapter/fetch path when custom adapter is set (mock mode)
    if (hasCustomAdapter) {
      return request(path, { method: 'POST', body: formData })
    }

    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()
      xhr.open('POST', url)
      xhr.withCredentials = true
      xhr.setRequestHeader('Accept', 'application/json')

      if (onProgress) {
        xhr.upload.addEventListener('progress', (e) => {
          if (e.lengthComputable) {
            onProgress({ loaded: e.loaded, total: e.total, percent: Math.round((e.loaded / e.total) * 100) })
          }
        })
      }

      xhr.addEventListener('load', () => {
        let data
        try {
          data = JSON.parse(xhr.responseText)
        } catch {
          data = xhr.responseText
        }

        if (xhr.status >= 200 && xhr.status < 300) {
          resolve(data?.data !== undefined ? data.data : data)
        } else {
          const error = new ApiError(
            xhr.status,
            data?.error?.code || 'HTTP_ERROR',
            data?.error?.message || `HTTP ${xhr.status}`,
            data?.error?.details
          )
          if (onError) onError(error)
          reject(error)
        }
      })

      xhr.addEventListener('error', () => {
        const error = new ApiError(0, 'NETWORK_ERROR', 'Upload failed')
        if (onError) onError(error)
        reject(error)
      })

      if (signal) {
        signal.addEventListener('abort', () => xhr.abort())
        xhr.addEventListener('abort', () => {
          reject(new ApiError(0, 'ABORTED', 'Upload aborted'))
        })
      }

      xhr.send(formData)
    })
  }

  return {
    /** GET request */
    get: (path, options) => request(path, { ...options, method: 'GET' }),

    /** POST request */
    post: (path, body, options) => request(path, { ...options, method: 'POST', body }),

    /** PUT request */
    put: (path, body, options) => request(path, { ...options, method: 'PUT', body }),

    /** DELETE request */
    delete: (path, options) => request(path, { ...options, method: 'DELETE' }),

    /** PATCH request */
    patch: (path, body, options) => request(path, { ...options, method: 'PATCH', body }),

    /** Upload with progress */
    upload
  }
}
