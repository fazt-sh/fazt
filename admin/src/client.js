/**
 * Shared API client instance
 * All stores import this to make API calls
 */
import { createClient, mockAdapter } from '../packages/fazt-sdk/index.js'

const useMock = new URLSearchParams(window.location.search).get('mock') === 'true'

function getApiBaseUrl() {
  const hostname = window.location.hostname
  const port = window.location.port ? ':' + window.location.port : ''
  const protocol = window.location.protocol
  const parts = hostname.split('.')
  if (parts.length >= 2) {
    parts[0] = 'admin'
    return `${protocol}//${parts.join('.')}${port}`
  }
  return ''
}

const apiBaseUrl = useMock ? '' : getApiBaseUrl()
console.log('[Admin] Mode:', useMock ? 'MOCK' : 'REAL')
console.log('[Admin] API Base:', apiBaseUrl || '(same origin)')

export const client = createClient(useMock ? { adapter: mockAdapter } : { baseUrl: apiBaseUrl })
