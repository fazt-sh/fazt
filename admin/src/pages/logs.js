/**
 * Activity Logs Page
 * View and filter activity logs
 */

import { activityLogs, activityStats, loadActivityLogs, loadActivityStats } from '../stores/data.js'
import { loading } from '../stores/app.js'
import {
  renderPanel, setupPanel,
  renderToolbar, setupToolbar,
  renderTable, setupTableClicks, renderTableFooter
} from '../components/index.js'

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
 * Build columns config for activity logs table
 */
function getColumns() {
  return [
    {
      key: 'action',
      label: 'Activity',
      render: (action, entry) => {
        const weightInfo = WEIGHT_INFO[entry.weight] || {}
        const actorInfo = ACTOR_INFO[entry.actor_type] || {}
        return `
          <div class="flex items-center gap-2" style="min-width: 0">
            <div class="icon-box icon-box-sm" style="flex-shrink: 0; border-color: ${weightInfo.color}">
              <i data-lucide="${actorInfo.icon || 'activity'}" class="w-3.5 h-3.5" style="color: ${weightInfo.color}"></i>
            </div>
            <div style="min-width: 0; overflow: hidden">
              <div class="text-label text-primary" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">${action}</div>
              <div class="text-caption text-faint show-mobile">${entry.resource_type} · ${formatRelativeTime(entry.timestamp)}</div>
            </div>
          </div>
        `
      }
    },
    {
      key: 'resource_type',
      label: 'Resource',
      hideOnMobile: true,
      render: (type, entry) => {
        const id = entry.resource_id || '-'
        return `
          <div style="min-width: 0; overflow: hidden">
            <span class="badge badge-muted">${type}</span>
            ${id !== '-' ? `<div class="text-caption mono text-muted mt-0.5" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 160px">${id}</div>` : ''}
          </div>
        `
      }
    },
    {
      key: 'actor_type',
      label: 'Actor',
      hideOnMobile: true,
      render: (type, entry) => {
        const actorInfo = ACTOR_INFO[type] || {}
        const actorId = entry.actor_id || actorInfo.label
        return `<span class="text-caption text-muted">${actorId}</span>`
      }
    },
    {
      key: 'weight',
      label: 'Priority',
      hideOnMobile: true,
      render: (weight) => {
        const info = WEIGHT_INFO[weight] || {}
        return `<span class="text-caption" style="color: ${info.color}">${info.label}</span>`
      }
    },
    {
      key: 'result',
      label: 'Result',
      hideOnMobile: true,
      render: (result) => {
        const isSuccess = result === 'success'
        return `
          <span class="flex items-center gap-1 text-caption ${isSuccess ? 'text-success' : 'text-error'}">
            <span class="status-dot ${isSuccess ? 'status-dot-success' : 'status-dot-error'}"></span>
            ${result}
          </span>
        `
      }
    },
    {
      key: 'timestamp',
      label: 'Time',
      hideOnMobile: true,
      render: (timestamp) => `<span class="text-caption text-muted">${formatRelativeTime(timestamp)}</span>`
    }
  ]
}

/**
 * Render activity logs page
 */
export function render(container, ctx) {
  const { client } = ctx
  let searchQuery = ''
  let filters = {
    limit: 50,
    offset: 0
  }

  function applyFilters() {
    // Combine search with filters
    const params = { ...filters }
    if (searchQuery) {
      // Search across action, resource_id, actor_id
      params.action = searchQuery
    }
    loadActivityLogs(client, params)
    loadActivityStats(client, params)
  }

  function update() {
    const logsData = activityLogs.get()
    const statsData = activityStats.get()
    const isLoading = loading.getKey('activity-logs')

    // Render table content
    const tableContent = isLoading
      ? `<div class="flex items-center justify-center p-8"><div class="text-caption text-muted">Loading...</div></div>`
      : renderTable({
          columns: getColumns(),
          data: logsData.entries || [],
          rowKey: 'id',
          rowDataAttr: 'log-id',
          clickable: false,
          emptyIcon: 'activity',
          emptyTitle: 'No activity logs',
          emptyMessage: 'Activity will appear here as it happens'
        })

    // Calculate pagination
    const totalPages = Math.ceil((logsData.total || 0) / filters.limit)
    const currentPage = Math.floor(filters.offset / filters.limit) + 1

    // Build toolbar (search bar)
    const toolbar = renderToolbar({
      searchId: 'logs-search',
      searchValue: searchQuery,
      searchPlaceholder: 'Search logs...',
      buttons: [
        { id: 'refresh-btn', icon: 'refresh-cw', title: 'Refresh' }
      ]
    })

    // Build filters row (separate row below toolbar)
    const filtersRow = `
      <div class="card-header flex items-center justify-end gap-2" style="border-top: 1px solid var(--border-subtle); padding-top: 8px; padding-bottom: 8px">
        <select id="filter-weight" class="btn btn-secondary btn-sm" style="padding: 4px 8px; cursor: pointer">
          <option value="">All Priorities</option>
          <option value="5" ${filters.min_weight === '5' ? 'selected' : ''}>Important (5+)</option>
          <option value="7" ${filters.min_weight === '7' ? 'selected' : ''}>Critical (7+)</option>
          <option value="9" ${filters.min_weight === '9' ? 'selected' : ''}>Security (9)</option>
        </select>
        <select id="filter-action" class="btn btn-secondary btn-sm" style="padding: 4px 8px; cursor: pointer">
          <option value="">All Actions</option>
          <option value="pageview" ${filters.action === 'pageview' ? 'selected' : ''}>Pageview</option>
          <option value="deploy" ${filters.action === 'deploy' ? 'selected' : ''}>Deploy</option>
          <option value="login" ${filters.action === 'login' ? 'selected' : ''}>Login</option>
          <option value="create" ${filters.action === 'create' ? 'selected' : ''}>Create</option>
          <option value="delete" ${filters.action === 'delete' ? 'selected' : ''}>Delete</option>
        </select>
        <select id="filter-actor" class="btn btn-secondary btn-sm" style="padding: 4px 8px; cursor: pointer">
          <option value="">All Actors</option>
          <option value="user" ${filters.actor_type === 'user' ? 'selected' : ''}>User</option>
          <option value="system" ${filters.actor_type === 'system' ? 'selected' : ''}>System</option>
          <option value="api_key" ${filters.actor_type === 'api_key' ? 'selected' : ''}>API Key</option>
          <option value="anonymous" ${filters.actor_type === 'anonymous' ? 'selected' : ''}>Anonymous</option>
        </select>
        <select id="filter-type" class="btn btn-secondary btn-sm" style="padding: 4px 8px; cursor: pointer">
          <option value="">All Types</option>
          <option value="page" ${filters.type === 'page' ? 'selected' : ''}>Page</option>
          <option value="app" ${filters.type === 'app' ? 'selected' : ''}>App</option>
          <option value="alias" ${filters.type === 'alias' ? 'selected' : ''}>Alias</option>
          <option value="kv" ${filters.type === 'kv' ? 'selected' : ''}>KV Store</option>
          <option value="session" ${filters.type === 'session' ? 'selected' : ''}>Session</option>
        </select>
      </div>
    `

    // Render footer with pagination
    const footer = logsData.total > 0 ? `
      <div class="card-footer flex items-center justify-between gap-4" style="border-radius: 0">
        <div class="flex items-center gap-2 text-caption text-muted">
          <span>Showing ${logsData.showing} of ${logsData.total}</span>
          ${statsData.size_estimate_bytes > 0 ? `<span class="hide-mobile">·</span><span class="hide-mobile">${formatBytes(statsData.size_estimate_bytes)} used</span>` : ''}
        </div>
        ${totalPages > 1 ? `
          <div class="flex items-center gap-2">
            <button id="first-page-btn" class="btn btn-secondary btn-sm" ${currentPage === 1 ? 'disabled' : ''} title="First page">
              <i data-lucide="chevrons-left" class="w-3.5 h-3.5"></i>
            </button>
            <button id="prev-page-btn" class="btn btn-secondary btn-sm" ${currentPage === 1 ? 'disabled' : ''} title="Previous page">
              <i data-lucide="chevron-left" class="w-3.5 h-3.5"></i>
            </button>
            <span class="text-caption text-muted px-2">Page ${currentPage} of ${totalPages}</span>
            <button id="next-page-btn" class="btn btn-secondary btn-sm" ${currentPage === totalPages ? 'disabled' : ''} title="Next page">
              <i data-lucide="chevron-right" class="w-3.5 h-3.5"></i>
            </button>
            <button id="last-page-btn" class="btn btn-secondary btn-sm" ${currentPage === totalPages ? 'disabled' : ''} title="Last page">
              <i data-lucide="chevrons-right" class="w-3.5 h-3.5"></i>
            </button>
          </div>
        ` : ''}
      </div>
    ` : ''

    // Render panel
    const panel = renderPanel({
      id: 'logs-list',
      title: 'Activity Logs',
      count: logsData.total,
      toolbar: toolbar + filtersRow,
      content: tableContent,
      footer,
      minHeight: 400,
      maxHeight: 600
    })

    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">
            ${panel}
          </div>
        </div>
      </div>
    `

    // Re-render Lucide icons
    if (window.lucide) window.lucide.createIcons()

    // Setup panel collapse
    setupPanel(container)

    // Setup toolbar (search + refresh)
    setupToolbar(container, {
      onSearch: (value) => {
        searchQuery = value
        filters.offset = 0 // Reset pagination
        applyFilters()
      },
      buttons: {
        'refresh-btn': () => applyFilters()
      }
    }, 'logs-search')

    // Setup filter dropdowns
    container.querySelector('#filter-weight')?.addEventListener('change', (e) => {
      filters.min_weight = e.target.value || undefined
      filters.offset = 0
      applyFilters()
    })

    container.querySelector('#filter-action')?.addEventListener('change', (e) => {
      filters.action = e.target.value || undefined
      filters.offset = 0
      applyFilters()
    })

    container.querySelector('#filter-actor')?.addEventListener('change', (e) => {
      filters.actor_type = e.target.value || undefined
      filters.offset = 0
      applyFilters()
    })

    container.querySelector('#filter-type')?.addEventListener('change', (e) => {
      filters.type = e.target.value || undefined
      filters.offset = 0
      applyFilters()
    })

    // Setup pagination
    container.querySelector('#first-page-btn')?.addEventListener('click', () => {
      filters.offset = 0
      applyFilters()
    })

    container.querySelector('#prev-page-btn')?.addEventListener('click', () => {
      filters.offset = Math.max(0, filters.offset - filters.limit)
      applyFilters()
    })

    container.querySelector('#next-page-btn')?.addEventListener('click', () => {
      filters.offset += filters.limit
      applyFilters()
    })

    container.querySelector('#last-page-btn')?.addEventListener('click', () => {
      const lastPageOffset = Math.floor((logsData.total - 1) / filters.limit) * filters.limit
      filters.offset = lastPageOffset
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
