/**
 * Logs Page
 * View and filter logs
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
 * Build columns config for logs table
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
            ${id !== '-' ? `<div class="text-caption mono text-muted mt-0.5" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">${id}</div>` : ''}
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
      label: '',
      hideOnMobile: true,
      render: (result) => {
        const isSuccess = result === 'success'
        return `
          <span class="flex items-center justify-center" title="${result}">
            <span class="status-dot ${isSuccess ? 'status-dot-success' : 'status-dot-error'}"></span>
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
 * Render logs page
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
  let isInitialized = false

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

  function updateTableContent() {
    const logsData = activityLogs.get()
    const statsData = activityStats.get()
    const isLoading = loading.getKey('activity-logs')
    const totalPages = Math.ceil((logsData.total || 0) / pageSize)

    // Find the scroll area and footer to update
    const scrollArea = container.querySelector('.panel-scroll-area')
    const footerArea = container.querySelector('.card-footer')

    if (!scrollArea) return

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
          emptyTitle: searchQuery ? 'No logs found' : 'No logs',
          emptyMessage: searchQuery ? 'Try a different search' : 'Activity will appear here as it happens'
        })

    // Update scroll area
    scrollArea.innerHTML = tableContent

    // Update footer
    if (logsData.total > 0) {
      const footerHTML = `
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
      `

      if (footerArea) {
        footerArea.innerHTML = footerHTML
      }
    } else if (footerArea) {
      footerArea.innerHTML = ''
    }

    // Re-render Lucide icons
    if (window.lucide) window.lucide.createIcons()

    // Setup pagination handlers
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

  function initialRender() {
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
          emptyTitle: searchQuery ? 'No logs found' : 'No logs',
          emptyMessage: searchQuery ? 'Try a different search' : 'Activity will appear here as it happens'
        })

    // Get filter labels
    const getWeightLabel = () => {
      if (!filterWeight) return 'All Priorities'
      const labels = { '5': 'Important (5+)', '7': 'Critical (7+)', '9': 'Security (9)' }
      return labels[filterWeight] || 'All Priorities'
    }

    const getActionLabel = () => {
      if (!filterAction) return 'All Actions'
      return filterAction.charAt(0).toUpperCase() + filterAction.slice(1)
    }

    const getActorLabel = () => {
      if (!filterActor) return 'All Actors'
      const labels = { 'user': 'User', 'system': 'System', 'api_key': 'API Key', 'anonymous': 'Anonymous' }
      return labels[filterActor] || 'All Actors'
    }

    const getTypeLabel = () => {
      if (!filterType) return 'All Types'
      return filterType.charAt(0).toUpperCase() + filterType.slice(1)
    }

    // Build toolbar (search + filter dropdowns)
    const toolbar = `
      ${renderToolbar({
        searchId: 'logs-search',
        searchValue: searchQuery,
        searchPlaceholder: 'Search logs...',
        buttons: []
      })}
      <div class="flex items-center justify-end gap-2">
        <!-- Priority Filter -->
        <div class="relative hide-mobile">
          <button id="filter-weight-btn" class="btn btn-secondary btn-sm flex items-center gap-1.5" style="padding: 4px 8px">
            <span class="text-caption">${getWeightLabel()}</span>
            <i data-lucide="chevron-down" class="w-3.5 h-3.5" style="color:var(--text-3)"></i>
          </button>
          <div id="filter-weight-menu" class="dropdown hidden fixed z-50" style="min-width: 160px">
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="">All Priorities</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="5">Important (5+)</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="7">Critical (7+)</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="9">Security (9)</div>
          </div>
        </div>

        <!-- Action Filter -->
        <div class="relative hide-mobile">
          <button id="filter-action-btn" class="btn btn-secondary btn-sm flex items-center gap-1.5" style="padding: 4px 8px">
            <span class="text-caption">${getActionLabel()}</span>
            <i data-lucide="chevron-down" class="w-3.5 h-3.5" style="color:var(--text-3)"></i>
          </button>
          <div id="filter-action-menu" class="dropdown hidden fixed z-50" style="min-width: 140px">
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="">All Actions</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="pageview">Pageview</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="deploy">Deploy</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="login">Login</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="create">Create</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="delete">Delete</div>
          </div>
        </div>

        <!-- Actor Filter -->
        <div class="relative hide-mobile">
          <button id="filter-actor-btn" class="btn btn-secondary btn-sm flex items-center gap-1.5" style="padding: 4px 8px">
            <span class="text-caption">${getActorLabel()}</span>
            <i data-lucide="chevron-down" class="w-3.5 h-3.5" style="color:var(--text-3)"></i>
          </button>
          <div id="filter-actor-menu" class="dropdown hidden fixed z-50" style="min-width: 140px">
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="">All Actors</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="user">User</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="system">System</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="api_key">API Key</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="anonymous">Anonymous</div>
          </div>
        </div>

        <!-- Type Filter -->
        <div class="relative hide-mobile">
          <button id="filter-type-btn" class="btn btn-secondary btn-sm flex items-center gap-1.5" style="padding: 4px 8px">
            <span class="text-caption">${getTypeLabel()}</span>
            <i data-lucide="chevron-down" class="w-3.5 h-3.5" style="color:var(--text-3)"></i>
          </button>
          <div id="filter-type-menu" class="dropdown hidden fixed z-50" style="min-width: 120px">
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="">All Types</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="page">Page</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="app">App</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="alias">Alias</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="kv">KV</div>
            <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" data-value="session">Session</div>
          </div>
        </div>

        <button id="refresh-btn" class="btn btn-secondary btn-sm" style="padding: 4px 8px" title="Refresh">
          <i data-lucide="refresh-cw" class="w-3.5 h-3.5"></i>
        </button>
      </div>
    `

    // Render footer with pagination
    const footer = logsData.total > 0 ? `
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
    ` : ''

    // Render panel
    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">
            ${renderPanel({
              id: 'logs.list',
              title: 'Logs',
              count: logsData.total || 0,
              toolbar,
              content: tableContent,
              footer,
              fillHeight: true
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

    // Setup custom dropdown handlers
    const setupDropdown = (btnId, menuId, onSelect) => {
      const btn = container.querySelector(`#${btnId}`)
      const menu = container.querySelector(`#${menuId}`)

      if (!btn || !menu) return

      // Toggle dropdown
      btn.addEventListener('click', (e) => {
        e.stopPropagation()

        // Close other dropdowns
        container.querySelectorAll('.dropdown').forEach(d => {
          if (d !== menu) d.classList.add('hidden')
        })

        // Toggle this dropdown
        menu.classList.toggle('hidden')

        // Position menu below button
        if (!menu.classList.contains('hidden')) {
          const rect = btn.getBoundingClientRect()
          menu.style.top = `${rect.bottom + 4}px`
          menu.style.left = `${rect.left}px`
        }
      })

      // Handle item selection
      menu.querySelectorAll('.dropdown-item').forEach(item => {
        item.addEventListener('click', (e) => {
          e.stopPropagation()
          const value = item.dataset.value
          onSelect(value)
          menu.classList.add('hidden')
        })
      })
    }

    setupDropdown('filter-weight-btn', 'filter-weight-menu', (value) => {
      filterWeight = value
      currentPage = 1
      applyFilters()
    })

    setupDropdown('filter-action-btn', 'filter-action-menu', (value) => {
      filterAction = value
      currentPage = 1
      applyFilters()
    })

    setupDropdown('filter-actor-btn', 'filter-actor-menu', (value) => {
      filterActor = value
      currentPage = 1
      applyFilters()
    })

    setupDropdown('filter-type-btn', 'filter-type-menu', (value) => {
      filterType = value
      currentPage = 1
      applyFilters()
    })

    // Close dropdowns when clicking outside
    const closeDropdowns = () => {
      container.querySelectorAll('.dropdown').forEach(d => d.classList.add('hidden'))
    }

    document.addEventListener('click', closeDropdowns)

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

    isInitialized = true
  }

  function update() {
    if (!isInitialized) {
      initialRender()
    } else {
      updateTableContent()
    }
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
