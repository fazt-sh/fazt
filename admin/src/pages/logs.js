/**
 * Activity Logs Page
 * View and filter activity logs
 */

import { activityLogs, activityStats, loadActivityLogs, loadActivityStats } from '../stores/data.js'
import { loading } from '../stores/app.js'

/**
 * Weight labels and colors
 */
const WEIGHT_INFO = {
  9: { label: 'Security', color: 'var(--error)' },
  8: { label: 'Auth', color: 'var(--error)' },
  7: { label: 'Config', color: 'var(--warning)' },
  6: { label: 'Deploy', color: 'var(--accent)' },
  5: { label: 'Data', color: 'var(--accent)' },
  4: { label: 'Action', color: 'var(--text-2)' },
  3: { label: 'Nav', color: 'var(--text-3)' },
  2: { label: 'Analytics', color: 'var(--text-3)' },
  1: { label: 'System', color: 'var(--text-4)' },
  0: { label: 'Debug', color: 'var(--text-4)' }
}

/**
 * Actor type labels and icons
 */
const ACTOR_INFO = {
  user: { label: 'User', icon: 'user' },
  system: { label: 'System', icon: 'server' },
  api_key: { label: 'API Key', icon: 'key' },
  anonymous: { label: 'Anonymous', icon: 'user-x' }
}

/**
 * Format relative time
 */
function formatRelativeTime(timestamp) {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now - date
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) return `${days}d ago`
  if (hours > 0) return `${hours}h ago`
  if (minutes > 0) return `${minutes}m ago`
  return 'Just now'
}

/**
 * Format bytes to human readable
 */
function formatBytes(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
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
 * Render activity logs page
 */
export function render(container, ctx) {
  const { client } = ctx
  let filterWeight = ''
  let filterAction = ''
  let filterActor = ''
  let filterType = ''
  let currentPage = 1
  const pageSize = 50

  function applyFilters() {
    const params = {
      limit: pageSize,
      offset: (currentPage - 1) * pageSize
    }
    if (filterWeight) params.min_weight = filterWeight
    if (filterAction) params.action = filterAction
    if (filterActor) params.actor_type = filterActor
    if (filterType) params.type = filterType

    loadActivityLogs(client, params)
    loadActivityStats(client, params)
  }

  function update() {
    const logsData = activityLogs.get()
    const statsData = activityStats.get()
    const isLoading = loading.getKey('activity-logs')

    const totalPages = Math.ceil((logsData.total || 0) / pageSize)
    const listCollapsed = getUIState('logs.list.collapsed', false)

    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">

            <!-- Panel Group: Activity Logs -->
            <div class="panel-group ${listCollapsed ? 'collapsed' : ''}">
              <div class="panel-group-card card">
                <header class="panel-group-header" data-group="list">
                  <button class="collapse-toggle">
                    <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                    <span class="text-heading text-primary">Activity Logs</span>
                    <span class="text-caption mono px-1.5 py-0.5 ml-2 badge-muted" style="border-radius: var(--radius-sm)">${logsData.total || 0}</span>
                  </button>
                  <div class="flex items-center gap-2 ml-auto">
                    <select id="filter-weight" class="btn btn-secondary btn-sm hide-mobile" style="padding: 4px 8px; cursor: pointer">
                      <option value="">All Priorities</option>
                      <option value="5" ${filterWeight === '5' ? 'selected' : ''}>Important (5+)</option>
                      <option value="7" ${filterWeight === '7' ? 'selected' : ''}>Critical (7+)</option>
                      <option value="9" ${filterWeight === '9' ? 'selected' : ''}>Security (9)</option>
                    </select>
                    <select id="filter-action" class="btn btn-secondary btn-sm hide-mobile" style="padding: 4px 8px; cursor: pointer">
                      <option value="">All Actions</option>
                      <option value="pageview" ${filterAction === 'pageview' ? 'selected' : ''}>Pageview</option>
                      <option value="deploy" ${filterAction === 'deploy' ? 'selected' : ''}>Deploy</option>
                      <option value="login" ${filterAction === 'login' ? 'selected' : ''}>Login</option>
                      <option value="create" ${filterAction === 'create' ? 'selected' : ''}>Create</option>
                      <option value="delete" ${filterAction === 'delete' ? 'selected' : ''}>Delete</option>
                    </select>
                    <select id="filter-actor" class="btn btn-secondary btn-sm hide-mobile" style="padding: 4px 8px; cursor: pointer">
                      <option value="">All Actors</option>
                      <option value="user" ${filterActor === 'user' ? 'selected' : ''}>User</option>
                      <option value="system" ${filterActor === 'system' ? 'selected' : ''}>System</option>
                      <option value="api_key" ${filterActor === 'api_key' ? 'selected' : ''}>API Key</option>
                      <option value="anonymous" ${filterActor === 'anonymous' ? 'selected' : ''}>Anonymous</option>
                    </select>
                    <select id="filter-type" class="btn btn-secondary btn-sm hide-mobile" style="padding: 4px 8px; cursor: pointer">
                      <option value="">All Types</option>
                      <option value="page" ${filterType === 'page' ? 'selected' : ''}>Page</option>
                      <option value="app" ${filterType === 'app' ? 'selected' : ''}>App</option>
                      <option value="alias" ${filterType === 'alias' ? 'selected' : ''}>Alias</option>
                      <option value="kv" ${filterType === 'kv' ? 'selected' : ''}>KV</option>
                      <option value="session" ${filterType === 'session' ? 'selected' : ''}>Session</option>
                    </select>
                    <button id="refresh-btn" class="btn btn-secondary btn-sm" style="padding: 4px 8px" title="Refresh">
                      <i data-lucide="refresh-cw" class="w-3.5 h-3.5"></i>
                    </button>
                  </div>
                </header>
                <div class="panel-group-body" style="padding: 0">
                  <div class="flex-1 overflow-auto scroll-panel" style="max-height: 600px">
                    ${isLoading ? `
                      <div class="flex items-center justify-center p-8">
                        <div class="text-caption text-muted">Loading logs...</div>
                      </div>
                    ` : (logsData.entries || []).length === 0 ? `
                      <div class="flex flex-col items-center justify-center p-8 text-center">
                        <div class="icon-box mb-3" style="width:48px;height:48px;opacity:0.5">
                          <i data-lucide="activity" class="w-6 h-6"></i>
                        </div>
                        <div class="text-heading text-primary mb-1">No activity logs</div>
                        <div class="text-caption text-muted">${filterWeight || filterAction || filterActor || filterType ? 'No logs match your filters' : 'Activity will appear here as it happens'}</div>
                      </div>
                    ` : `
                      <div class="table-container">
                        <table>
                          <thead class="sticky" style="top: 0; background: var(--bg-1)">
                            <tr class="border-b">
                              <th class="px-4 py-2 text-left text-micro text-muted" style="width: 120px">Activity</th>
                              <th class="px-4 py-2 text-left text-micro text-muted hide-mobile" style="width: 140px">Resource</th>
                              <th class="px-4 py-2 text-left text-micro text-muted hide-mobile" style="width: 100px">Actor</th>
                              <th class="px-4 py-2 text-left text-micro text-muted hide-mobile" style="width: 80px">Priority</th>
                              <th class="px-4 py-2 text-left text-micro text-muted hide-mobile" style="width: 80px">Result</th>
                              <th class="px-4 py-2 text-left text-micro text-muted" style="width: 100px">Time</th>
                            </tr>
                          </thead>
                          <tbody>
                            ${(logsData.entries || []).map(entry => {
                              const weightInfo = WEIGHT_INFO[entry.weight] || {}
                              const actorInfo = ACTOR_INFO[entry.actor_type] || {}
                              const isSuccess = entry.result === 'success'
                              return `
                                <tr class="row" style="border-bottom: 1px solid var(--border-subtle)">
                                  <td class="px-4 py-2">
                                    <div class="flex items-center gap-2">
                                      <div class="icon-box icon-box-sm" style="border-color: ${weightInfo.color}; flex-shrink: 0">
                                        <i data-lucide="${actorInfo.icon || 'activity'}" class="w-3.5 h-3.5" style="color: ${weightInfo.color}"></i>
                                      </div>
                                      <div style="min-width: 0">
                                        <div class="text-label text-primary">${entry.action}</div>
                                        <div class="text-caption text-faint show-mobile">${entry.resource_type}</div>
                                      </div>
                                    </div>
                                  </td>
                                  <td class="px-4 py-2 hide-mobile">
                                    <span class="badge badge-muted">${entry.resource_type}</span>
                                    ${entry.resource_id ? `<div class="text-caption mono text-muted mt-0.5" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 120px">${entry.resource_id}</div>` : ''}
                                  </td>
                                  <td class="px-4 py-2 hide-mobile">
                                    <span class="text-caption text-muted">${entry.actor_id || actorInfo.label}</span>
                                  </td>
                                  <td class="px-4 py-2 hide-mobile">
                                    <span class="text-caption" style="color: ${weightInfo.color}">${weightInfo.label}</span>
                                  </td>
                                  <td class="px-4 py-2 hide-mobile">
                                    <span class="flex items-center gap-1 text-caption ${isSuccess ? 'text-success' : 'text-error'}">
                                      <span class="status-dot ${isSuccess ? 'status-dot-success' : 'status-dot-error'}"></span>
                                      ${entry.result}
                                    </span>
                                  </td>
                                  <td class="px-4 py-2">
                                    <span class="text-caption text-muted">${formatRelativeTime(entry.timestamp)}</span>
                                  </td>
                                </tr>
                              `
                            }).join('')}
                          </tbody>
                        </table>
                      </div>
                    `}
                  </div>
                  ${logsData.total > 0 ? `
                    <div class="card-footer flex items-center justify-between" style="border-radius: 0">
                      <span class="text-caption text-muted">
                        Showing ${logsData.showing} of ${logsData.total}
                        ${statsData.size_estimate_bytes > 0 ? ` · ${formatBytes(statsData.size_estimate_bytes)}` : ''}
                      </span>
                      ${totalPages > 1 ? `
                        <div class="flex items-center gap-2">
                          <button id="first-page-btn" class="btn btn-secondary btn-sm" ${currentPage === 1 ? 'disabled' : ''} title="First page">
                            <i data-lucide="chevrons-left" class="w-3.5 h-3.5"></i>
                          </button>
                          <button id="prev-page-btn" class="btn btn-secondary btn-sm" ${currentPage === 1 ? 'disabled' : ''} title="Previous">
                            <i data-lucide="chevron-left" class="w-3.5 h-3.5"></i>
                          </button>
                          <span class="text-caption text-muted px-2">Page ${currentPage} of ${totalPages}</span>
                          <button id="next-page-btn" class="btn btn-secondary btn-sm" ${currentPage === totalPages ? 'disabled' : ''} title="Next">
                            <i data-lucide="chevron-right" class="w-3.5 h-3.5"></i>
                          </button>
                          <button id="last-page-btn" class="btn btn-secondary btn-sm" ${currentPage === totalPages ? 'disabled' : ''} title="Last page">
                            <i data-lucide="chevrons-right" class="w-3.5 h-3.5"></i>
                          </button>
                        </div>
                      ` : ''}
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

    // Filter handlers
    container.querySelector('#filter-weight')?.addEventListener('change', (e) => {
      filterWeight = e.target.value
      currentPage = 1
      applyFilters()
    })

    container.querySelector('#filter-action')?.addEventListener('change', (e) => {
      filterAction = e.target.value
      currentPage = 1
      applyFilters()
    })

    container.querySelector('#filter-actor')?.addEventListener('change', (e) => {
      filterActor = e.target.value
      currentPage = 1
      applyFilters()
    })

    container.querySelector('#filter-type')?.addEventListener('change', (e) => {
      filterType = e.target.value
      currentPage = 1
      applyFilters()
    })

    // Refresh handler
    container.querySelector('#refresh-btn')?.addEventListener('click', () => {
      applyFilters()
    })

    // Pagination handlers
    container.querySelector('#first-page-btn')?.addEventListener('click', () => {
      currentPage = 1
      applyFilters()
    })

    container.querySelector('#prev-page-btn')?.addEventListener('click', () => {
      if (currentPage > 1) {
        currentPage--
        applyFilters()
      }
    })

    container.querySelector('#next-page-btn')?.addEventListener('click', () => {
      if (currentPage < totalPages) {
        currentPage++
        applyFilters()
      }
    })

    container.querySelector('#last-page-btn')?.addEventListener('click', () => {
      currentPage = totalPages
      applyFilters()
    })
  }

  // Subscribe to data changes
  const unsubLogs = activityLogs.subscribe(update)
  const unsubStats = activityStats.subscribe(update)
  const unsubLoading = loading.subscribeKey('activity-logs', update)

  // Initial load
  applyFilters()
  update()

  // Return cleanup function
  return () => {
    unsubLogs()
    unsubStats()
    unsubLoading()
  }
}
