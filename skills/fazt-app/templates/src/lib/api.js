// API Client
// Session-aware HTTP client for fazt APIs

import { getSession } from './session.js'

const BASE_URL = '' // Same origin

async function request(path, options = {}) {
  const url = BASE_URL + path
  const config = {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    },
    ...options
  }

  const response = await fetch(url, config)

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: response.statusText }))
    throw new Error(error.error || `HTTP ${response.status}`)
  }

  return response.json()
}

export const api = {
  session: null,

  // Initialize with session
  init() {
    this.session = getSession()
    return this.session
  },

  // GET with session in query
  async get(path) {
    if (!this.session) this.init()
    const sep = path.includes('?') ? '&' : '?'
    return request(`${path}${sep}session=${encodeURIComponent(this.session)}`)
  },

  // POST with session in body
  async post(path, data = {}) {
    if (!this.session) this.init()
    return request(path, {
      method: 'POST',
      body: JSON.stringify({ ...data, session: this.session })
    })
  },

  // PUT with session in body
  async put(path, data = {}) {
    if (!this.session) this.init()
    return request(path, {
      method: 'PUT',
      body: JSON.stringify({ ...data, session: this.session })
    })
  },

  // DELETE with session in query
  async delete(path) {
    if (!this.session) this.init()
    const sep = path.includes('?') ? '&' : '?'
    return request(`${path}${sep}session=${encodeURIComponent(this.session)}`, {
      method: 'DELETE'
    })
  }
}
