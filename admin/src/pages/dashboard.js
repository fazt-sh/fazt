/**
 * Dashboard Page - Panel-Based Layout
 */

import { apps, aliases, health } from '../stores/data.js'
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
 * Format uptime from seconds
 * @param {number} seconds
 */
function formatUptime(seconds) {
  if (!seconds) return '-'
  const hours = Math.floor(seconds / 3600)
  const days = Math.floor(hours / 24)
  const remainingHours = hours % 24

  if (days > 0) return `${days}d ${remainingHours}h`
  if (hours > 0) return `${hours}h`
  return `${Math.floor(seconds / 60)}m`
}

/**
 * Calculate total storage from apps
 * @param {Array} appList
 */
function calcTotalStorage(appList) {
  return appList.reduce((sum, app) => sum + (app.size_bytes || 0), 0)
}

/**
 * Render apps panel for dashboard
 */
function renderAppsPanel(appList, isCollapsed) {
  const columns = [
    {
      key: 'title',
      label: 'App',
      render: (title, app) => `
        <div class="flex items-center gap-2" style="min-width: 0">
          <span class="status-dot status-dot-success pulse show-mobile" style="flex-shrink: 0"></span>
          <div class="icon-box icon-box-sm" style="flex-shrink: 0">
            <i data-lucide="box" class="w-3.5 h-3.5"></i>
          </div>
          <div style="min-width: 0; overflow: hidden">
            <div class="text-label text-primary" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">${title}</div>
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
 * Render aliases panel for dashboard
 */
function renderAliasesPanel(aliasList, isCollapsed) {
  const columns = [
    {
      key: 'subdomain',
      label: 'Alias',
      render: (subdomain, alias) => `
        <div class="flex items-center gap-2" style="min-width: 0">
          <div class="icon-box icon-box-sm" style="flex-shrink: 0">
            <i data-lucide="link" class="w-3.5 h-3.5"></i>
          </div>
          <div style="min-width: 0; overflow: hidden">
            <div class="text-label text-primary" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">${subdomain}</div>
            <div class="text-caption text-faint show-mobile">${alias.type}</div>
          </div>
        </div>
      `
    },
    {
      key: 'type',
      label: 'Type',
      hideOnMobile: true,
      render: (type) => `<span class="text-caption text-muted">${type}</span>`
    },
    {
      key: 'targets',
      label: 'Target',
      hideOnMobile: true,
      render: (targets) => {
        if (!targets) return '<span class="text-caption text-faint">-</span>'
        const appId = targets.app_id || targets.url || '-'
        return `<span class="text-caption mono text-muted">${appId}</span>`
      }
    }
  ]

  const displayAliases = aliasList.slice(0, 5)

  const tableContent = renderTable({
    columns,
    data: displayAliases,
    rowKey: 'subdomain',
    rowDataAttr: 'alias-subdomain',
    clickable: true,
    emptyIcon: 'link',
    emptyTitle: 'No aliases yet',
    emptyMessage: 'Create aliases via CLI'
  })

  const footer = `
    <div class="card-footer flex items-center justify-between" style="border-radius: 0">
      <span class="text-caption text-muted">${aliasList.length} alias${aliasList.length === 1 ? '' : 'es'}</span>
      <button class="text-caption font-medium text-accent" data-navigate="/aliases">View all &rarr;</button>
    </div>
  `

  return renderPanel({
    id: 'dashboard.aliases',
    title: 'Aliases',
    count: aliasList.length,
    toolbar: renderToolbar({ searchId: 'dashboard-aliases-filter', searchPlaceholder: 'Filter...' }),
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
    const aliasList = aliases.get()
    const healthData = health.get()

    // Compute derived data
    const totalStorage = calcTotalStorage(appList)
    const memoryUsed = healthData.memory?.used_mb || 0

    // Get collapse states
    const statusCollapsed = getUIState('dashboard.status.collapsed', false)
    const appsCollapsed = getUIState('dashboard.apps.collapsed', false)
    const aliasesCollapsed = getUIState('dashboard.aliases.collapsed', false)

    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">

            <!-- Panel Group: System Status -->
            <div class="panel-group ${statusCollapsed ? 'collapsed' : ''}">
              <div class="panel-group-card card">
                <header class="panel-group-header" data-group="status">
                  <button class="collapse-toggle">
                    <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                    <span class="text-heading text-primary">System</span>
                    <span class="text-caption text-faint ml-auto hide-mobile">4 metrics</span>
                  </button>
                </header>
                <div class="panel-group-body">
                    <div class="panel-grid grid-4">
                    <div class="stat-card card">
                      <div class="stat-card-header">
                        <span class="text-micro text-muted">Status</span>
                        <i data-lucide="heart-pulse" class="w-4 h-4 text-faint"></i>
                      </div>
                      <div class="stat-card-value text-display mono ${healthData.status === 'healthy' ? 'text-success' : 'text-warning'}">
                        ${healthData.status || '-'}
                      </div>
                      <div class="stat-card-subtitle text-caption text-muted">v${healthData.version || '-'}</div>
                    </div>

                    <div class="stat-card card">
                      <div class="stat-card-header">
                        <span class="text-micro text-muted">Uptime</span>
                        <i data-lucide="clock" class="w-4 h-4 text-faint"></i>
                      </div>
                      <div class="stat-card-value text-display mono text-primary">${formatUptime(healthData.uptime_seconds)}</div>
                      <div class="stat-card-subtitle text-caption text-muted">${healthData.mode || '-'}</div>
                    </div>

                    <div class="stat-card card">
                      <div class="stat-card-header">
                        <span class="text-micro text-muted">Memory</span>
                        <i data-lucide="cpu" class="w-4 h-4 text-faint"></i>
                      </div>
                      <div class="stat-card-value text-display mono text-primary">${memoryUsed.toFixed(1)}<span class="text-caption">MB</span></div>
                      <div class="stat-card-subtitle text-caption text-muted">${healthData.runtime?.goroutines || 0} goroutines</div>
                    </div>

                    <div class="stat-card card">
                      <div class="stat-card-header">
                        <span class="text-micro text-muted">Storage</span>
                        <i data-lucide="database" class="w-4 h-4 text-faint"></i>
                      </div>
                      <div class="stat-card-value text-display mono text-primary">${formatBytes(totalStorage)}</div>
                      <div class="stat-card-subtitle text-caption text-muted">${appList.length} apps</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Panel Group: Apps -->
            ${renderAppsPanel(appList, appsCollapsed)}

            <!-- Panel Group: Aliases -->
            ${renderAliasesPanel(aliasList, aliasesCollapsed)}

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
  const unsubAliases = aliases.subscribe(update)
  const unsubHealth = health.subscribe(update)

  // Initial render
  update()

  // Return cleanup function
  return () => {
    unsubApps()
    unsubAliases()
    unsubHealth()
  }
}
