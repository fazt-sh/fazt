/**
 * Activity Logs Page
 * View and filter activity logs
 */

import { activityLogs, activityStats, loadActivityLogs, loadActivityStats } from '../stores/data.js'
import { loading } from '../stores/app.js'
import {
  renderPanel, setupPanel,
  renderToolbar, setupToolbar,
  renderTable, renderTableFooter
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
            ${id !== '-' ? `<div class="text-caption mono text-muted mt-0.5" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 120px">${id}</div>` : ''}
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
    // Search query can filter by action
    if (searchQuery) params.action = searchQuery

    loadActivityLogs(client, params)
    loadActivityStats(client, params)
  }

  function update() {
    const logsData = activityLogs.get()
    const statsData = activityStats.get()
    const isLoading = loading.getKey('activity-logs')
    const totalPages = Math.ceil((logsData.total || 0) / pageSize)

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
          emptyTitle: searchQuery ? 'No logs found' : 'No activity logs',
          emptyMessage: searchQuery ? 'Try a different search' : 'Activity will appear here as it happens'
        })

    // Build toolbar (search + filter dropdowns)
    const toolbar = `
      ${renderToolbar({
        searchId: 'logs-search',
        searchValue: searchQuery,
        searchPlaceholder: 'Search logs...',
        buttons: []
      })}
      <div class="flex items-center justify-end gap-2">
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
    `

    // Render footer with pagination
    const footer = logsData.total > 0 ? `
      <div class="flex items-center justify-between">
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
    ` : ''

    // Render panel
    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">
            ${renderPanel({
              id: 'logs.list',
              title: 'Activity Logs',
              count: logsData.total || 0,
              toolbar,
              content: tableContent,
              footer: footer ? renderTableFooter(logsData.total, 'log').replace('>${logsData.total} log', `>${footer}`) : '',
              minHeight: 400,
              maxHeight: 600
            })}
          </div>
        </div>
      </div>
    `

    // Re-render Lucide icons
    if (window.lucide) window.lucide.createIcons()

    // Setup panel collapse
    setupPanel(container)

    // Setup toolbar (search)
    setupToolbar(container, {
      onSearch: (value) => {
        searchQuery = value
        currentPage = 1
        applyFilters()
      }
    }, 'logs-search')

    // Setup filter dropdowns
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
