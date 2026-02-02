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
 * Get UI state from localStorage
 */
function getUIState(key, defaultValue = false) {
  try {
    const state = JSON.parse(localStorage.getItem('fazt.web.ui.state') || '{}')
    return state[key] !== undefined ? state[key] : defaultValue
  } catch {
    return defaultValue
  }
}

/**
 * Set UI state to localStorage
 */
function setUIState(key, value) {
  try {
    const state = JSON.parse(localStorage.getItem('fazt.web.ui.state') || '{}')
    state[key] = value
    localStorage.setItem('fazt.web.ui.state', JSON.stringify(state))
  } catch (e) {
    console.error('Failed to save UI state:', e)
  }
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

    const listCollapsed = getUIState('logs.list.collapsed', false)

    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">

            <!-- Panel Group: Logs List -->
            <div class="panel-group ${listCollapsed ? 'collapsed' : ''}">
              <div class="panel-group-card card">
                <header class="panel-group-header" data-group="list">
                  <button class="collapse-toggle">
                    <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                    <span class="text-heading text-primary">Logs</span>
                    ${selectedAppId ? `<span class="text-caption mono px-1.5 py-0.5 ml-2 badge-muted" style="border-radius: var(--radius-sm)">${filteredLogs.length}</span>` : ''}
                  </button>
                  <div class="flex items-center gap-2 ml-auto">
                    <select id="app-select" class="btn btn-secondary btn-sm" style="padding: 4px 8px; cursor: pointer; max-width: 140px">
                      <option value="">Select App</option>
                      ${appList.map(app => `
                        <option value="${app.id}" ${selectedAppId === app.id ? 'selected' : ''}>${app.name}</option>
                      `).join('')}
                    </select>
                    <select id="filter-level" class="btn btn-secondary btn-sm hide-mobile" style="padding: 4px 8px; cursor: pointer" ${!selectedAppId ? 'disabled' : ''}>
                      <option value="">All Levels</option>
                      ${levels.map(level => `
                        <option value="${level}" ${filterLevel === level ? 'selected' : ''}>${level}</option>
                      `).join('')}
                    </select>
                    <button id="refresh-btn" class="btn btn-secondary btn-sm" style="padding: 4px 8px" title="Refresh" ${!selectedAppId ? 'disabled' : ''}>
                      <i data-lucide="refresh-cw" class="w-3.5 h-3.5"></i>
                    </button>
                  </div>
                </header>
                <div class="panel-group-body" style="padding: 0">
                  <div class="flex-1 overflow-auto scroll-panel" style="max-height: 600px">
                    ${!selectedAppId ? `
                      <div class="flex flex-col items-center justify-center p-8 text-center">
                        <div class="icon-box mb-3" style="width:48px;height:48px;opacity:0.5">
                          <i data-lucide="terminal" class="w-6 h-6"></i>
                        </div>
                        <div class="text-heading text-primary mb-1">Select an app</div>
                        <div class="text-caption text-muted">Choose an app from the dropdown to view its logs</div>
                      </div>
                    ` : isLoading ? `
                      <div class="flex items-center justify-center p-8">
                        <div class="text-caption text-muted">Loading logs...</div>
                      </div>
                    ` : filteredLogs.length === 0 ? `
                      <div class="flex flex-col items-center justify-center p-8 text-center">
                        <div class="icon-box mb-3" style="width:48px;height:48px;opacity:0.5">
                          <i data-lucide="file-text" class="w-6 h-6"></i>
                        </div>
                        <div class="text-heading text-primary mb-1">No logs</div>
                        <div class="text-caption text-muted">${filterLevel ? 'No logs match your filter' : 'No logs for this app yet'}</div>
                      </div>
                    ` : `
                      <div class="table-container">
                        <table>
                          <thead class="sticky" style="top: 0; background: var(--bg-1)">
                            <tr class="border-b">
                              <th class="px-4 py-2 text-left text-micro text-muted" style="width: 80px">Level</th>
                              <th class="px-4 py-2 text-left text-micro text-muted hide-mobile" style="width: 140px">Time</th>
                              <th class="px-4 py-2 text-left text-micro text-muted">Message</th>
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
                                  <td class="px-4 py-2 text-caption text-muted mono hide-mobile">${formatTimestamp(log.created_at)}</td>
                                  <td class="px-4 py-2">
                                    <div class="text-caption mono text-primary" style="word-break: break-all">${escapeHtml(log.message)}</div>
                                    <div class="text-caption text-faint show-mobile">${formatTimestamp(log.created_at)}</div>
                                  </td>
                                </tr>
                              `
                            }).join('')}
                          </tbody>
                        </table>
                      </div>
                    `}
                  </div>
                  ${selectedAppId && filteredLogs.length > 0 ? `
                    <div class="card-footer flex items-center justify-between" style="border-radius: 0">
                      <span class="text-caption text-muted">${filteredLogs.length} log${filteredLogs.length === 1 ? '' : 's'}</span>
                    </div>
                  ` : ''}
                </div>
              </div>
            </div>

          </div>
        </div>
      </div>
    `

    // Re-render Lucide icons
    if (window.lucide) window.lucide.createIcons()

    // Setup collapse handlers
    container.querySelectorAll('.collapse-toggle').forEach(toggle => {
      toggle.addEventListener('click', () => {
        const header = toggle.closest('.panel-group-header')
        const group = header.dataset.group
        const panelGroup = header.closest('.panel-group')
        const isCollapsed = panelGroup.classList.toggle('collapsed')
        setUIState(`logs.${group}.collapsed`, isCollapsed)
      })
    })

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
