/**
 * Dashboard Page - Panel-Based Layout
 */

import { apps, aliases, stats, health } from '../stores/data.js'
import {
  renderPanel, setupPanel, getUIState, setUIState,
  renderToolbar, setupToolbar,
  renderTable, setupTableClicks, renderTableFooter
} from '../components/index.js'

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
 * Render apps panel for dashboard
 */
function renderAppsPanel(appList, isCollapsed) {
  const columns = [
    {
      key: 'name',
      label: 'App',
      render: (name, app) => `
        <div class="flex items-center gap-2" style="min-width: 0">
          <span class="status-dot status-dot-success pulse show-mobile" style="flex-shrink: 0"></span>
          <div class="icon-box icon-box-sm" style="flex-shrink: 0">
            <i data-lucide="box" class="w-3.5 h-3.5"></i>
          </div>
          <div style="min-width: 0; overflow: hidden">
            <div class="text-label text-primary" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">${name}</div>
            <div class="text-caption mono text-faint show-mobile">${formatBytes(app.size_bytes)}</div>
          </div>
        </div>
      `
    },
    {
      key: 'updated_at',
      label: 'Updated',
      hideOnMobile: true,
      render: (v) => `<span class="text-caption text-muted">${formatRelativeTime(v)}</span>`
    },
    {
      key: 'size_bytes',
      label: 'Size',
      hideOnMobile: true,
      render: (v) => `<span class="text-caption mono text-muted">${formatBytes(v)}</span>`
    },
    {
      key: 'status',
      label: 'Status',
      hideOnMobile: true,
      render: () => `
        <span class="flex items-center gap-1 text-caption text-success">
          <span class="status-dot status-dot-success pulse"></span>
          Live
        </span>
      `
    }
  ]

  const displayApps = appList.slice(0, 5)

  const tableContent = renderTable({
    columns,
    data: displayApps,
    rowKey: 'id',
    rowDataAttr: 'app-id',
    clickable: true,
    emptyIcon: 'layers',
    emptyTitle: 'No apps yet',
    emptyMessage: 'Deploy your first app via CLI'
  })

  const footer = `
    <div class="card-footer flex items-center justify-between" style="border-radius: 0">
      <span class="text-caption text-muted">${appList.length} app${appList.length === 1 ? '' : 's'}</span>
      <button class="text-caption font-medium text-accent" data-navigate="/apps">View all &rarr;</button>
    </div>
  `

  return renderPanel({
    id: 'dashboard.apps',
    title: 'Apps',
    count: appList.length,
    toolbar: renderToolbar({ searchId: 'dashboard-apps-filter', searchPlaceholder: 'Filter...' }),
    content: tableContent,
    footer,
    minHeight: 200,
    maxHeight: 300
  })
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
            ${renderAppsPanel(appList, appsCollapsed)}

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

    // Setup panel collapse handlers
    setupPanel(container)

    // Setup legacy collapse handlers for non-component panels
    setupLegacyCollapseHandlers(container)

    // Setup apps table clicks
    setupTableClicks(container, 'app-id', (id) => {
      router.push(`/apps/${id}`)
    })

    // Setup navigation buttons
    container.querySelectorAll('[data-navigate]').forEach(btn => {
      btn.addEventListener('click', () => {
        router.push(btn.dataset.navigate)
      })
    })
  }

  /**
   * Setup collapse handlers for legacy panels (stats, activity)
   */
  function setupLegacyCollapseHandlers(container) {
    container.querySelectorAll('.panel-group-header[data-group]').forEach(header => {
      const toggle = header.querySelector('.collapse-toggle')
      if (!toggle) return

      toggle.addEventListener('click', () => {
        const group = header.dataset.group
        const panelGroup = header.closest('.panel-group')
        const isCollapsed = panelGroup.classList.toggle('collapsed')
        setUIState(`dashboard.${group}.collapsed`, isCollapsed)
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
