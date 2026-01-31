/**
 * Dashboard Page - Panel-Based Layout
 */

import { apps, stats, health } from '../stores/data.js'

/**
 * Get UI state from localStorage (web-specific key)
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
 * Set UI state to localStorage (web-specific key)
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
 * Format bytes to human readable
 * @param {number} bytes
 */
function formatBytes(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * Format relative time
 * @param {string} timestamp
 */
function formatRelativeTime(timestamp) {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now - date
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const days = Math.floor(hours / 24)

  if (days > 0) return `${days}d ago`
  if (hours > 0) return `${hours}h ago`
  return 'Just now'
}

/**
 * Render dashboard page
 * @param {HTMLElement} container
 * @param {Object} ctx
 */
export function render(container, ctx) {
  const { router } = ctx

  function update() {
    const appList = apps.get()
    const statsData = stats.get()
    const healthData = health.get()

    // Get collapse states
    const statsCollapsed = getUIState('dashboard.stats.collapsed', false)
    const appsCollapsed = getUIState('dashboard.apps.collapsed', false)
    const activityCollapsed = getUIState('dashboard.activity.collapsed', false)

    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">

            <!-- Panel Group: Overview Stats -->
            <div class="panel-group ${statsCollapsed ? 'collapsed' : ''}">
              <div class="panel-group-card card">
                <header class="panel-group-header" data-group="stats">
                  <button class="collapse-toggle">
                    <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                    <span class="text-heading text-primary">Overview</span>
                    <span class="text-caption text-faint ml-auto hide-mobile">5 metrics</span>
                  </button>
                </header>
                <div class="panel-group-body">
                    <div class="panel-grid grid-5">
                    <div class="stat-card card">
                      <div class="stat-card-header">
                        <span class="text-micro text-muted">Apps</span>
                        <i data-lucide="layers" class="w-4 h-4 text-faint"></i>
                      </div>
                      <div class="stat-card-value text-display mono text-primary">${appList.length || statsData.apps || 0}</div>
                      <div class="stat-card-subtitle flex items-center gap-1 text-caption text-success">
                        <i data-lucide="trending-up" class="w-3 h-3"></i>
                        Active
                      </div>
                    </div>

                    <div class="stat-card card">
                      <div class="stat-card-header">
                        <span class="text-micro text-muted">Requests</span>
                        <i data-lucide="activity" class="w-4 h-4 text-faint"></i>
                      </div>
                      <div class="stat-card-value text-display mono text-primary">${(statsData.requests_24h || 0).toLocaleString()}</div>
                      <div class="stat-card-subtitle">
                        <div class="spark">
                          ${[40, 65, 45, 80, 55, 90, 70].map(h => `<div class="spark-bar" style="height:${h}%"></div>`).join('')}
                        </div>
                      </div>
                    </div>

                    <div class="stat-card card">
                      <div class="stat-card-header">
                        <span class="text-micro text-muted">Storage</span>
                        <i data-lucide="database" class="w-4 h-4 text-faint"></i>
                      </div>
                      <div class="stat-card-value text-display mono text-primary">${formatBytes(statsData.storage_bytes || 0)}</div>
                      <div class="stat-card-subtitle">
                        <div class="progress">
                          <div class="progress-bar" style="width: 24%"></div>
                        </div>
                      </div>
                    </div>

                    <div class="stat-card card">
                      <div class="stat-card-header">
                        <span class="text-micro text-muted">Uptime</span>
                        <i data-lucide="clock" class="w-4 h-4 text-faint"></i>
                      </div>
                      <div class="stat-card-value text-display mono text-success">${statsData.uptime_percent || 99.97}<span class="text-caption">%</span></div>
                      <div class="stat-card-subtitle text-caption text-muted">30d average</div>
                    </div>

                    <div class="stat-card card">
                      <div class="stat-card-header">
                        <span class="text-micro text-muted">Status</span>
                        <i data-lucide="heart-pulse" class="w-4 h-4 text-faint"></i>
                      </div>
                      <div class="stat-card-value text-display mono ${healthData.status === 'healthy' ? 'text-success' : 'text-warning'}">
                        ${healthData.status || 'Unknown'}
                      </div>
                      <div class="stat-card-subtitle text-caption text-muted">v${healthData.version || '0.17.0'}</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Panel Group: Apps -->
            <div class="panel-group ${appsCollapsed ? 'collapsed' : ''}">
              <div class="panel-group-card card">
                <header class="panel-group-header" data-group="apps">
                  <button class="collapse-toggle">
                    <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                    <span class="text-heading text-primary">Apps</span>
                    <span class="text-caption mono px-1.5 py-0.5 ml-2 badge-muted" style="border-radius: var(--radius-sm)">${appList.length}</span>
                  </button>
                </header>
                <div class="panel-group-body" style="padding: 0">
                  <div class="card" style="border: none; border-radius: 0">
                    <div class="card-header">
                      <div class="input" style="padding: 4px 8px">
                        <i data-lucide="search" class="w-3.5 h-3.5 text-faint"></i>
                        <input type="text" placeholder="Filter..." class="text-caption" style="width: 120px">
                      </div>
                    </div>
                    <div class="flex-1 overflow-auto scroll-panel" style="max-height: 400px">
                ${appList.length === 0 ? `
                  <div class="flex flex-col items-center justify-center h-full p-8 text-center">
                    <div class="icon-box mb-3" style="width:48px;height:48px;opacity:0.5">
                      <i data-lucide="inbox" class="w-6 h-6"></i>
                    </div>
                    <div class="text-heading text-primary mb-1">No apps yet</div>
                    <div class="text-caption text-muted mb-4">Deploy your first app to get started</div>
                    <button class="btn btn-primary px-4 py-2 text-label">
                      <i data-lucide="plus" class="w-4 h-4 inline-block mr-1" style="vertical-align:-2px"></i>
                      Deploy App
                    </button>
                  </div>
                ` : `
                  <div class="table-container">
                    <table>
                    <thead class="sticky" style="top: 0; background: var(--bg-1)">
                      <tr class="border-b">
                        <th class="px-4 py-2 text-left text-micro text-muted">App</th>
                        <th class="px-4 py-2 text-left text-micro text-muted">Status</th>
                        <th class="px-4 py-2 text-left text-micro text-muted">Updated</th>
                        <th class="px-4 py-2 text-left text-micro text-muted">Size</th>
                        <th class="px-4 py-2"></th>
                      </tr>
                    </thead>
                    <tbody>
                      ${appList.slice(0, 7).map(app => `
                        <tr class="row row-clickable border-b" data-app-id="${app.id}" style="border-color: var(--border-subtle)">
                          <td class="px-4 py-2">
                            <div class="flex items-center gap-2">
                              <div class="icon-box"><i data-lucide="box" class="w-4 h-4"></i></div>
                              <div>
                                <div class="text-label text-primary">${app.name}</div>
                                <div class="text-caption mono text-faint">${app.name}.zyt.app</div>
                              </div>
                            </div>
                          </td>
                          <td class="px-4 py-2">
                            <span class="flex items-center gap-1 text-caption text-success">
                              <span class="status-dot status-dot-success pulse"></span>
                              Live
                            </span>
                          </td>
                          <td class="px-4 py-2 text-caption text-muted">${formatRelativeTime(app.updated_at)}</td>
                          <td class="px-4 py-2 text-caption mono text-muted">${formatBytes(app.size_bytes)}</td>
                          <td class="px-4 py-2">
                            <button class="btn btn-ghost btn-icon btn-sm" style="color: var(--text-4)">
                              <i data-lucide="more-horizontal" class="w-4 h-4"></i>
                            </button>
                          </td>
                        </tr>
                      `).join('')}
                    </tbody>
                  </table>
                  </div>
                      `}
                    </div>
                    <div class="card-footer flex items-center justify-between">
                      <span class="text-caption text-muted">Showing ${Math.min(7, appList.length)} of ${appList.length}</span>
                      <button class="text-caption font-medium text-accent" data-navigate="/apps">View all &rarr;</button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Panel Group: Quick Actions -->
            <div class="panel-group">
              <div class="panel-group-card card">
                <header class="panel-group-header" data-group="actions">
                  <button class="collapse-toggle">
                    <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                    <span class="text-heading text-primary">Quick Actions</span>
                  </button>
                </header>
                <div class="panel-group-body">
                  <div class="grid grid-cols-3 gap-2">
                    <button class="flex flex-col items-center gap-2 p-3" style="background:var(--bg-2);border-radius:var(--radius-md)">
                      <div class="icon-action"><i data-lucide="rocket" class="w-5 h-5"></i></div>
                      <span class="text-caption text-secondary">Deploy</span>
                    </button>
                    <button class="flex flex-col items-center gap-2 p-3" style="background:var(--bg-2);border-radius:var(--radius-md)">
                      <div class="icon-action"><i data-lucide="download" class="w-5 h-5"></i></div>
                      <span class="text-caption text-secondary">Backup</span>
                    </button>
                    <button class="flex flex-col items-center gap-2 p-3" style="background:var(--bg-2);border-radius:var(--radius-md)">
                      <div class="icon-action"><i data-lucide="terminal" class="w-5 h-5"></i></div>
                      <span class="text-caption text-secondary">CLI</span>
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Panel Group: Notifications -->
            <div class="panel-group ${activityCollapsed ? 'collapsed' : ''}">
              <div class="panel-group-card card">
                <header class="panel-group-header" data-group="activity">
                  <button class="collapse-toggle">
                    <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                    <span class="text-heading text-primary">Notifications</span>
                  </button>
                </header>
                <div class="panel-group-body">
                  <div class="activity-list">
                    ${[
                      { icon: 'check-circle', title: 'momentum deployed', time: '2h ago' },
                      { icon: 'settings', title: 'Config updated', time: '5h ago' },
                      { icon: 'alert-triangle', title: 'reflex error', time: '1d ago' },
                      { icon: 'key', title: 'API token created', time: '2d ago' },
                      { icon: 'upload', title: 'nexus deployed', time: '2d ago' },
                      { icon: 'shield-check', title: 'SSL renewed', time: '4d ago' }
                    ].map(item => `
                      <div class="activity-item">
                        <i data-lucide="${item.icon}" class="w-4 h-4 text-muted" style="flex-shrink: 0;"></i>
                        <span class="text-label text-primary" style="flex: 1; min-width: 0;">${item.title}</span>
                        <span class="text-caption text-muted" style="flex-shrink: 0;">${item.time}</span>
                      </div>
                    `).join('')}
                  </div>
                </div>
              </div>
            </div>

          </div>
        </div>
      </div>
    `

    // Re-render Lucide icons
    if (window.lucide) {
      window.lucide.createIcons()
    }

    // Setup collapse handlers
    setupCollapseHandlers(container)

    // Add click handlers
    container.querySelectorAll('[data-app-id]').forEach(row => {
      row.addEventListener('click', () => {
        router.push(`/apps/${row.dataset.appId}`)
      })
    })

    container.querySelectorAll('[data-navigate]').forEach(btn => {
      btn.addEventListener('click', () => {
        router.push(btn.dataset.navigate)
      })
    })
  }

  /**
   * Setup collapse toggle handlers
   */
  function setupCollapseHandlers(container) {
    container.querySelectorAll('.collapse-toggle').forEach(toggle => {
      toggle.addEventListener('click', () => {
        const header = toggle.closest('.panel-group-header')
        const group = header.dataset.group
        const panelGroup = header.closest('.panel-group')
        const isCollapsed = panelGroup.classList.contains('collapsed')

        if (isCollapsed) {
          panelGroup.classList.remove('collapsed')
          setUIState(`dashboard.${group}.collapsed`, false)
        } else {
          panelGroup.classList.add('collapsed')
          setUIState(`dashboard.${group}.collapsed`, true)
        }
      })
    })
  }

  // Subscribe to data changes
  const unsubApps = apps.subscribe(update)
  const unsubStats = stats.subscribe(update)
  const unsubHealth = health.subscribe(update)

  // Initial render
  update()

  // Return cleanup function
  return () => {
    unsubApps()
    unsubStats()
    unsubHealth()
  }
}
