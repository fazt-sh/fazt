/**
 * fazt-sdk Type Definitions
 * JSDoc types for the Fazt API
 */

/**
 * @typedef {Object} App
 * @property {string} id - Unique app ID
 * @property {string} name - App name (subdomain)
 * @property {string} source - Source type (local, git, template)
 * @property {Object} [manifest] - App manifest.json contents
 * @property {number} file_count - Number of files
 * @property {number} size_bytes - Total size in bytes
 * @property {string} created_at - ISO timestamp
 * @property {string} updated_at - ISO timestamp
 */

/**
 * @typedef {Object} AppSource
 * @property {string} type - Source type
 * @property {string} [url] - Git URL if applicable
 * @property {string} [ref] - Git ref (branch/tag)
 * @property {string} [commit] - Git commit SHA
 */

/**
 * @typedef {Object} AppFile
 * @property {string} path - File path
 * @property {number} size - File size in bytes
 * @property {string} mime_type - MIME type
 * @property {string} modified - Last modified timestamp
 */

/**
 * @typedef {Object} Alias
 * @property {string} subdomain - Subdomain name
 * @property {string} type - Type: proxy, redirect, split, reserved
 * @property {Object} [targets] - Target configuration
 * @property {string} created_at - ISO timestamp
 * @property {string} updated_at - ISO timestamp
 */

/**
 * @typedef {Object} SystemHealth
 * @property {string} status - Health status
 * @property {number} uptime_seconds - Server uptime
 * @property {string} version - Fazt version
 * @property {string} mode - Server mode (dev, prod)
 * @property {Object} memory - Memory statistics
 * @property {Object} database - Database statistics
 * @property {Object} runtime - Runtime statistics
 */

/**
 * @typedef {Object} SystemConfig
 * @property {string} version - Fazt version
 * @property {string} domain - Server domain
 * @property {string} env - Environment (dev, prod)
 * @property {boolean} https - HTTPS enabled
 * @property {boolean} ntfy - Ntfy notifications enabled
 */

/**
 * @typedef {Object} User
 * @property {string} id - User ID
 * @property {string} email - User email
 * @property {string} name - Display name
 * @property {string} [avatar] - Avatar URL
 * @property {string} provider - Auth provider
 */

/**
 * @typedef {Object} ApiResponse
 * @template T
 * @property {boolean} success - Request success
 * @property {T} [data] - Response data
 * @property {ApiError} [error] - Error details
 */

/**
 * @typedef {Object} ApiError
 * @property {string} code - Error code
 * @property {string} message - Error message
 * @property {string} [details] - Additional details
 */

/**
 * @typedef {Object} ClientOptions
 * @property {string} [baseUrl] - Base URL (default: same origin)
 * @property {Function} [adapter] - Custom fetch adapter
 * @property {Function} [onError] - Error handler
 * @property {Function} [onRequest] - Request interceptor
 * @property {Function} [onResponse] - Response interceptor
 */

/**
 * @typedef {Object} RequestOptions
 * @property {string} [method] - HTTP method
 * @property {Object} [headers] - Request headers
 * @property {Object|string} [body] - Request body
 * @property {AbortSignal} [signal] - Abort signal
 */

export {}
