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
 * Format timestamp
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
  let filters = {
    limit: 50,
    offset: 0
  }

  function applyFilters() {
    loadActivityLogs(client, filters)
    loadActivityStats(client, filters)
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

    // Build filter toolbar
    const toolbar = renderToolbar({
      title: 'Activity Logs',
      badge: logsData.total > 0 ? logsData.total : null,
      actions: [
        {
          id: 'refresh',
          label: 'Refresh',
          icon: 'refresh-cw',
          variant: 'secondary'
        }
      ],
      filters: [
        {
          id: 'weight',
          type: 'select',
          placeholder: 'All Priorities',
          options: [
            { value: '', label: 'All Priorities' },
            { value: '5', label: 'Important (5+)' },
            { value: '7', label: 'Critical (7+)' },
            { value: '9', label: 'Security (9)' }
          ]
        },
        {
          id: 'action',
          type: 'select',
          placeholder: 'All Actions',
          options: [
            { value: '', label: 'All Actions' },
            { value: 'pageview', label: 'Pageview' },
            { value: 'deploy', label: 'Deploy' },
            { value: 'login', label: 'Login' },
            { value: 'create', label: 'Create' },
            { value: 'delete', label: 'Delete' }
          ]
        },
        {
          id: 'actor_type',
          type: 'select',
          placeholder: 'All Actors',
          options: [
            { value: '', label: 'All Actors' },
            { value: 'user', label: 'User' },
            { value: 'system', label: 'System' },
            { value: 'api_key', label: 'API Key' },
            { value: 'anonymous', label: 'Anonymous' }
          ]
        },
        {
          id: 'type',
          type: 'select',
          placeholder: 'All Types',
          options: [
            { value: '', label: 'All Types' },
            { value: 'page', label: 'Page' },
            { value: 'app', label: 'App' },
            { value: 'alias', label: 'Alias' },
            { value: 'kv', label: 'KV Store' },
            { value: 'session', label: 'Session' }
          ]
        }
      ]
    })

    // Render footer with stats
    const footer = logsData.total > 0 ? `
      <div class="card-footer flex items-center justify-between gap-4" style="border-radius: 0">
        <div class="flex items-center gap-2 text-caption text-muted">
          <span>Showing ${logsData.showing} of ${logsData.total}</span>
          ${statsData.size_estimate_bytes > 0 ? `<span class="hide-mobile">·</span><span class="hide-mobile">${formatBytes(statsData.size_estimate_bytes)} used</span>` : ''}
        </div>
        <div class="flex items-center gap-2">
          ${logsData.offset > 0 ? `<button id="prev-btn" class="btn btn-secondary btn-sm">Previous</button>` : ''}
          ${logsData.offset + logsData.showing < logsData.total ? `<button id="next-btn" class="btn btn-secondary btn-sm">Next</button>` : ''}
        </div>
      </div>
    ` : ''

    // Render panel
    const panel = renderPanel({
      id: 'logs-list',
      title: toolbar,
      body: tableContent + footer,
      noPadding: true
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

    // Setup toolbar
    setupToolbar(container, (action) => {
      if (action === 'refresh') {
        applyFilters()
      }
    }, (filterId, value) => {
      // Handle filter changes
      if (filterId === 'weight') {
        filters.min_weight = value || undefined
      } else if (filterId === 'action') {
        filters.action = value || undefined
      } else if (filterId === 'actor_type') {
        filters.actor_type = value || undefined
      } else if (filterId === 'type') {
        filters.type = value || undefined
      }
      filters.offset = 0 // Reset pagination
      applyFilters()
    })

    // Setup pagination
    container.querySelector('#prev-btn')?.addEventListener('click', () => {
      filters.offset = Math.max(0, filters.offset - filters.limit)
      applyFilters()
    })

    container.querySelector('#next-btn')?.addEventListener('click', () => {
      filters.offset += filters.limit
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
