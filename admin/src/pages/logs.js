/**
 * Logs Page
 * Application logs viewer
 */

import { logs, loadLogs, apps } from '../stores/data.js'
import { loading } from '../stores/app.js'

/**
 * Format timestamp
 * @param {string} timestamp
 */
function formatTimestamp(timestamp) {
  const date = new Date(timestamp)
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    second: '2-digit'
  })
}

/**
 * Get log level styling
 * @param {string} level
 */
function getLevelStyle(level) {
  const styles = {
    info: { bg: 'var(--accent-soft)', color: 'var(--accent)' },
    warn: { bg: 'var(--warning-soft)', color: 'var(--warning)' },
    error: { bg: 'var(--error-soft)', color: 'var(--error)' },
    debug: { bg: 'var(--bg-3)', color: 'var(--text-3)' }
  }
  return styles[level] || styles.info
}

/**
 * Render logs page
 * @param {HTMLElement} container
 * @param {Object} ctx
 */
export function render(container, ctx) {
  const { client } = ctx
  let selectedAppId = ''
  let filterLevel = ''

  function update() {
    const logList = logs.get()
    const appList = apps.get()
    const isLoading = loading.getKey('logs')

    // Apply level filter
    let filteredLogs = [...logList]
    if (filterLevel) {
      filteredLogs = filteredLogs.filter(l => l.level === filterLevel)
    }

    // Get unique levels for filter
    const levels = ['info', 'warn', 'error', 'debug']

    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">

            <!-- Page Header -->
            <div class="flex items-center justify-between mb-4">
              <div>
                <h1 class="text-title text-primary">Logs</h1>
                <p class="text-caption text-muted">${selectedAppId ? `Showing logs for selected app` : 'Select an app to view logs'}</p>
              </div>
              <div class="flex items-center gap-2">
                <select id="app-select" class="btn btn-secondary text-label" style="padding: 6px 12px; cursor: pointer; min-width: 180px">
                  <option value="">Select App</option>
                  ${appList.map(app => `
                    <option value="${app.id}" ${selectedAppId === app.id ? 'selected' : ''}>${app.name}</option>
                  `).join('')}
                </select>
                <select id="filter-level" class="btn btn-secondary text-label" style="padding: 6px 12px; cursor: pointer" ${!selectedAppId ? 'disabled' : ''}>
                  <option value="">All Levels</option>
                  ${levels.map(level => `
                    <option value="${level}" ${filterLevel === level ? 'selected' : ''}>${level}</option>
                  `).join('')}
                </select>
                <button id="refresh-btn" class="btn btn-secondary text-label" style="padding: 6px 12px" title="Refresh" ${!selectedAppId ? 'disabled' : ''}>
                  <i data-lucide="refresh-cw" class="w-4 h-4"></i>
                </button>
              </div>
            </div>

            <!-- Logs List -->
            <div class="card">
              ${!selectedAppId ? `
                <div class="flex items-center justify-center p-8">
                  <div class="text-center">
                    <i data-lucide="terminal" class="w-12 h-12 text-faint mx-auto mb-2"></i>
                    <div class="text-heading text-primary mb-1">Select an app</div>
                    <div class="text-caption text-muted">Choose an app from the dropdown to view its logs</div>
                  </div>
                </div>
              ` : isLoading ? `
                <div class="flex items-center justify-center p-8">
                  <div class="text-caption text-muted">Loading logs...</div>
                </div>
              ` : filteredLogs.length === 0 ? `
                <div class="flex items-center justify-center p-8">
                  <div class="text-center">
                    <i data-lucide="file-text" class="w-12 h-12 text-faint mx-auto mb-2"></i>
                    <div class="text-heading text-primary mb-1">No logs</div>
                    <div class="text-caption text-muted">${filterLevel ? 'No logs match your filter' : 'No logs for this app yet'}</div>
                  </div>
                </div>
              ` : `
                <div class="scroll-panel" style="max-height: 600px">
                  <table>
                    <thead class="sticky" style="top: 0; background: var(--bg-1); z-index: 1">
                      <tr style="border-bottom: 1px solid var(--border)">
                        <th class="px-4 py-3 text-left text-micro text-muted" style="width: 100px">Level</th>
                        <th class="px-4 py-3 text-left text-micro text-muted" style="width: 160px">Time</th>
                        <th class="px-4 py-3 text-left text-micro text-muted">Message</th>
                      </tr>
                    </thead>
                    <tbody>
                      ${filteredLogs.map(log => {
                        const style = getLevelStyle(log.level)
                        return `
                          <tr class="row" style="border-bottom: 1px solid var(--border-subtle)">
                            <td class="px-4 py-2">
                              <span class="text-caption px-2 py-0.5 font-medium" style="background: ${style.bg}; color: ${style.color}; border-radius: var(--radius-sm); text-transform: uppercase">${log.level}</span>
                            </td>
                            <td class="px-4 py-2 text-caption text-muted mono">${formatTimestamp(log.created_at)}</td>
                            <td class="px-4 py-2 text-caption mono text-primary" style="word-break: break-all">${escapeHtml(log.message)}</td>
                          </tr>
                        `
                      }).join('')}
                    </tbody>
                  </table>
                </div>
              `}
            </div>

          </div>
        </div>
      </div>
    `

    // Re-render Lucide icons
    if (window.lucide) window.lucide.createIcons()

    // App select handler
    container.querySelector('#app-select')?.addEventListener('change', (e) => {
      selectedAppId = e.target.value
      filterLevel = ''
      if (selectedAppId) {
        loadLogs(client, selectedAppId)
      } else {
        logs.set([])
      }
      update()
    })

    // Level filter handler
    container.querySelector('#filter-level')?.addEventListener('change', (e) => {
      filterLevel = e.target.value
      update()
    })

    // Refresh handler
    container.querySelector('#refresh-btn')?.addEventListener('click', () => {
      if (selectedAppId) {
        loadLogs(client, selectedAppId)
      }
    })
  }

  // Subscribe to data changes
  const unsubLogs = logs.subscribe(update)
  const unsubApps = apps.subscribe(update)
  const unsubLoading = loading.subscribeKey('logs', update)

  // Initial render
  update()

  // Return cleanup function
  return () => {
    unsubLogs()
    unsubApps()
    unsubLoading()
  }
}

/**
 * Escape HTML entities
 * @param {string} str
 */
function escapeHtml(str) {
  const div = document.createElement('div')
  div.textContent = str
  return div.innerHTML
}
