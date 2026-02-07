/**
 * fazt-sdk
 * Universal API client for Fazt — serves both Admin UI and fazt-apps
 *
 * Admin usage:
 *   import { createClient } from 'fazt-sdk'
 *   const client = createClient()
 *   const apps = await client.apps.list()
 *
 * App usage:
 *   import { createAppClient } from 'fazt-sdk'
 *   const app = createAppClient()
 *   const user = await app.auth.me()
 *   const photos = await app.http.get('/api/photos')
 *
 * Mock mode:
 *   import { createClient, mockAdapter } from 'fazt-sdk'
 *   const client = createClient({ adapter: mockAdapter })
 */

import { createHttpClient } from './client.js'
import { createAdminNamespace } from './admin.js'
import { createAppNamespace } from './app.js'

export { ApiError } from './errors.js'
export { createMockAdapter, mockAdapter } from './mock.js'

/**
 * Create Fazt Admin API client
 * @param {import('./types.js').ClientOptions} [options]
 */
export function createClient(options = {}) {
  const http = createHttpClient(options)
  return { http, ...createAdminNamespace(http) }
}

/**
 * Create Fazt App API client
 * @param {import('./types.js').ClientOptions} [options]
 */
export function createAppClient(options = {}) {
  const http = createHttpClient(options)
  return createAppNamespace(http)
}

// Default client instance
let defaultClient = null

/**
 * Get or create default admin client
 * @param {import('./types.js').ClientOptions} [options]
 */
export function getClient(options) {
  if (!defaultClient || options) {
    defaultClient = createClient(options)
  }
  return defaultClient
}
