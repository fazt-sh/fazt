/**
 * Dashboard Page
 */

import { apps, stats, health } from '../stores/data.js'

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

    // Get collapse state from localStorage
    const statsCollapsed = localStorage.getItem('dashboard-stats-collapsed') === 'true'
    const appsCollapsed = localStorage.getItem('dashboard-apps-collapsed') === 'true'
    const activityCollapsed = localStorage.getItem('dashboard-activity-collapsed') === 'true'

    container.innerHTML = `
      <div class="flex flex-col h-full overflow-y-auto lg:overflow-hidden">
        <!-- Stats Section -->
        <div class="mb-4 flex-shrink-0">
          <div class="flex items-center gap-2 mb-3 cursor-pointer" data-accordion="stats">
            <i data-lucide="chevron-right" class="w-4 h-4 chevron ${!statsCollapsed ? 'open' : ''}" style="color:var(--text-3);transition:transform 0.15s ease"></i>
            <span class="text-heading text-primary">Overview</span>
            <span class="text-caption text-faint ml-auto hide-mobile">5 metrics</span>
          </div>
          <div class="accordion-content ${!statsCollapsed ? 'open' : ''}" data-accordion-content="stats">
            <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3 mb-4">
          <div class="card p-4">
            <div class="flex items-center justify-between mb-2">
              <span class="text-micro text-muted">Apps</span>
              <i data-lucide="layers" class="w-4 h-4 text-faint"></i>
            </div>
            <div class="text-display mono text-primary">${appList.length || statsData.apps || 0}</div>
            <div class="flex items-center gap-1 text-caption mt-1 text-success">
              <i data-lucide="trending-up" class="w-3 h-3"></i>
              Active
            </div>
          </div>

          <div class="card p-4">
            <div class="flex items-center justify-between mb-2">
              <span class="text-micro text-muted">Requests</span>
              <i data-lucide="activity" class="w-4 h-4 text-faint"></i>
            </div>
            <div class="text-display mono text-primary">${(statsData.requests_24h || 0).toLocaleString()}</div>
            <div class="spark mt-2">
              ${[40, 65, 45, 80, 55, 90, 70].map(h => `<div class="spark-bar" style="height:${h}%"></div>`).join('')}
            </div>
          </div>

          <div class="card p-4">
            <div class="flex items-center justify-between mb-2">
              <span class="text-micro text-muted">Storage</span>
              <i data-lucide="database" class="w-4 h-4 text-faint"></i>
            </div>
            <div class="text-display mono text-primary">${formatBytes(statsData.storage_bytes || 0)}</div>
            <div class="progress mt-2">
              <div class="progress-bar" style="width: 24%"></div>
            </div>
          </div>

          <div class="card p-4">
            <div class="flex items-center justify-between mb-2">
              <span class="text-micro text-muted">Uptime</span>
              <i data-lucide="clock" class="w-4 h-4 text-faint"></i>
            </div>
            <div class="text-display mono text-success">${statsData.uptime_percent || 99.97}<span class="text-caption">%</span></div>
            <div class="text-caption mt-1 text-muted">30d average</div>
          </div>

          <div class="card p-4">
            <div class="flex items-center justify-between mb-2">
              <span class="text-micro text-muted">Status</span>
              <i data-lucide="heart-pulse" class="w-4 h-4 text-faint"></i>
            </div>
            <div class="text-display mono ${healthData.status === 'healthy' ? 'text-success' : 'text-warning'}">
              ${healthData.status || 'Unknown'}
            </div>
            <div class="text-caption mt-1 text-muted">v${healthData.version || '0.17.0'}</div>
          </div>
            </div>
          </div>
        </div>

        <!-- Main Grid -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-4 flex-1 lg:min-h-0 lg:overflow-hidden">
          <!-- Left Column: Apps -->
          <div class="lg:col-span-2 flex flex-col lg:min-h-0 lg:overflow-hidden">
            <!-- Apps Section Header -->
            <div class="flex items-center gap-2 mb-3 cursor-pointer flex-shrink-0" data-accordion="apps">
              <i data-lucide="chevron-right" class="w-4 h-4 chevron ${!appsCollapsed ? 'open' : ''}" style="color:var(--text-3);transition:transform 0.15s ease"></i>
              <span class="text-heading text-primary">Apps</span>
              <span class="text-caption mono px-1.5 py-0.5 ml-2 badge-muted" style="border-radius: var(--radius-sm)">${appList.length}</span>
            </div>
            <!-- Apps Content -->
            <div class="accordion-content ${!appsCollapsed ? 'open' : ''} flex flex-col lg:min-h-0 lg:overflow-hidden" data-accordion-content="apps">
              <div class="card flex flex-col overflow-hidden flex-1">
                <div class="card-header">
                  <div class="input" style="padding: 4px 8px">
                    <i data-lucide="search" class="w-3.5 h-3.5 text-faint"></i>
                    <input type="text" placeholder="Filter..." class="text-caption" style="width: 120px">
                  </div>
              </div>
              <div class="flex-1 overflow-auto scroll-panel">
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

          <!-- Right Column -->
          <div class="flex flex-col lg:min-h-0 gap-4 lg:overflow-hidden">
            <!-- Quick Actions -->
            <div class="flex-shrink-0">
              <div class="flex items-center gap-2 mb-3 cursor-pointer" data-accordion="actions">
                <i data-lucide="chevron-right" class="w-4 h-4 chevron open" style="color:var(--text-3);transition:transform 0.15s ease"></i>
                <span class="text-heading text-primary">Quick Actions</span>
              </div>
              <div class="accordion-content open" data-accordion-content="actions">
                <div class="card p-4">
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

            <!-- Recent Activity -->
            <div class="flex-shrink-0 flex-1 flex flex-col lg:min-h-0">
              <div class="flex items-center gap-2 mb-3 cursor-pointer" data-accordion="activity">
                <i data-lucide="chevron-right" class="w-4 h-4 chevron ${!activityCollapsed ? 'open' : ''}" style="color:var(--text-3);transition:transform 0.15s ease"></i>
                <span class="text-heading text-primary">Recent Activity</span>
              </div>
              <div class="accordion-content ${!activityCollapsed ? 'open' : ''} flex flex-col flex-1 lg:min-h-0" data-accordion-content="activity">
                <div class="card flex flex-col overflow-hidden flex-1">
              <div class="flex-1 overflow-auto scroll-panel">
                ${[
                  { icon: 'check-circle', title: 'momentum deployed', time: '2h ago' },
                  { icon: 'settings', title: 'Config updated', time: '5h ago' },
                  { icon: 'alert-triangle', title: 'reflex error', time: '1d ago' },
                  { icon: 'key', title: 'API token created', time: '2d ago' },
                  { icon: 'upload', title: 'nexus deployed', time: '2d ago' },
                  { icon: 'shield-check', title: 'SSL renewed', time: '4d ago' }
                ].map(item => `
                  <div class="px-4 py-3 flex items-center gap-3 row cursor-pointer">
                    <div class="icon-box"><i data-lucide="${item.icon}" class="w-4 h-4"></i></div>
                    <div class="flex-1 min-w-0">
                      <div class="text-label text-primary">${item.title}</div>
                      <div class="text-caption text-faint">${item.time}</div>
                    </div>
                  </div>
                `).join('')}
              </div>
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

    // Accordion handlers
    container.querySelectorAll('[data-accordion]').forEach(trigger => {
      trigger.addEventListener('click', () => {
        const section = trigger.dataset.accordion
        const content = container.querySelector(`[data-accordion-content="${section}"]`)
        const chevron = trigger.querySelector('.chevron')

        if (content && chevron) {
          const isOpen = content.classList.contains('open')

          if (isOpen) {
            content.classList.remove('open')
            chevron.classList.remove('open')
            localStorage.setItem(`dashboard-${section}-collapsed`, 'true')
          } else {
            content.classList.add('open')
            chevron.classList.add('open')
            localStorage.setItem(`dashboard-${section}-collapsed`, 'false')
          }

          // Update grid column collapse state
          updateGridColumnState(trigger)
        }
      })
    })

    // Helper function to update grid column collapse state
    function updateGridColumnState(trigger) {
      const gridColumn = trigger.closest('.grid > div')
      if (gridColumn) {
        const hasOpenContent = gridColumn.querySelector('.accordion-content.open')
        if (hasOpenContent) {
          gridColumn.classList.remove('all-collapsed')
        } else {
          gridColumn.classList.add('all-collapsed')
        }
      }
    }

    // Set initial grid column states based on accordion states
    container.querySelectorAll('.grid > div').forEach(column => {
      const hasOpenContent = column.querySelector('.accordion-content.open')
      if (!hasOpenContent) {
        column.classList.add('all-collapsed')
        console.log('[Dashboard] Added all-collapsed to column:', column)
      } else {
        console.log('[Dashboard] Column has open content:', column)
      }
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
