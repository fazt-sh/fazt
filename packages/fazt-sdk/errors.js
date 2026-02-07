/**
 * fazt-sdk Error Types
 * Structured API errors with status helpers
 */

/**
 * API error with status code and structured details
 * @extends Error
 */
export class ApiError extends Error {
  /**
   * @param {number} status - HTTP status code (0 for network errors)
   * @param {string} code - Machine-readable error code
   * @param {string} message - Human-readable message
   * @param {*} [details] - Additional error details
   */
  constructor(status, code, message, details) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }

  /** @returns {boolean} 401 Unauthorized */
  get isAuth() { return this.status === 401 }

  /** @returns {boolean} 403 Forbidden */
  get isForbidden() { return this.status === 403 }

  /** @returns {boolean} 404 Not Found */
  get isNotFound() { return this.status === 404 }

  /** @returns {boolean} 429 Too Many Requests */
  get isRateLimit() { return this.status === 429 }

  /** @returns {boolean} 5xx Server Error */
  get isServer() { return this.status >= 500 }
}
