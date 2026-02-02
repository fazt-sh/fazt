/**
 * Aliases Page - Component-Based Layout
 */

import { aliases, apps } from '../stores/data.js'
import { loading } from '../stores/app.js'
import {
  renderPanel, setupPanel,
  renderToolbar, setupToolbar,
  renderTable, setupTableClicks, renderTableFooter
} from '../components/index.js'

/**
 * Format relative time
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
 * Get type badge class
 */
function getTypeBadge(type) {
  switch (type) {
    case 'proxy': return 'badge-success'
    case 'redirect': return 'badge-warning'
    case 'split': return 'badge'
    case 'reserved': return 'badge-muted'
    default: return 'badge-muted'
  }
}

/**
 * Get app name by ID
 */
function getAppName(appId) {
  const appList = apps.get()
  const app = appList.find(a => a.id === appId)
  return app?.name || appId
}

/**
 * Get icon for alias type
 */
function getAliasIcon(type) {
  switch (type) {
    case 'redirect': return 'external-link'
    case 'reserved': return 'lock'
    default: return 'link'
  }
}

/**
 * Build columns config for aliases table
 */
function getColumns() {
  return [
    {
      key: 'subdomain',
      label: 'Subdomain',
      render: (subdomain, alias) => `
        <div class="flex items-center gap-2" style="min-width: 0">
          <span class="status-dot ${alias.type === 'reserved' ? '' : 'status-dot-success pulse'} show-mobile" style="flex-shrink: 0"></span>
          <div class="icon-box icon-box-sm" style="flex-shrink: 0">
            <i data-lucide="${getAliasIcon(alias.type)}" class="w-3.5 h-3.5"></i>
          </div>
          <div style="min-width: 0; overflow: hidden">
            <div class="text-label text-primary" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">${subdomain}</div>
            <div class="text-caption mono text-faint show-mobile">${alias.type}</div>
          </div>
        </div>
      `
    },
    {
      key: 'type',
      label: 'Type',
      hideOnMobile: true,
      render: (type) => `<span class="badge ${getTypeBadge(type)}">${type}</span>`
    },
    {
      key: 'target',
      label: 'Target',
      hideOnMobile: true,
      render: (_, alias) => {
        const target = alias.type === 'proxy' ? getAppName(alias.targets?.app_id) :
                       alias.type === 'redirect' ? alias.targets?.url || '-' : '-'
        return `<span class="text-caption mono text-muted" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: block; max-width: 180px">${target}</span>`
      }
    },
    {
      key: 'updated_at',
      label: 'Updated',
      hideOnMobile: true,
      render: (v) => `<span class="text-caption text-muted">${formatRelativeTime(v)}</span>`
    },
    {
      key: 'status',
      label: 'Status',
      hideOnMobile: true,
      render: (_, alias) => `
        <span class="flex items-center gap-1 text-caption ${alias.type === 'reserved' ? 'text-muted' : 'text-success'}">
          <span class="status-dot ${alias.type === 'reserved' ? '' : 'status-dot-success pulse'}"></span>
          ${alias.type === 'reserved' ? 'Reserved' : 'Active'}
        </span>
      `
    }
  ]
}

/**
 * Render aliases page
 */
export function render(container, ctx) {
  let filter = ''

  function update() {
    const aliasList = aliases.get()
    const isLoading = loading.getKey('aliases')

    const filteredAliases = filter
      ? aliasList.filter(alias =>
          alias.subdomain.toLowerCase().includes(filter.toLowerCase()) ||
          alias.type.toLowerCase().includes(filter.toLowerCase())
        )
      : aliasList

    // Render table content
    const tableContent = isLoading
      ? `<div class="flex items-center justify-center p-8"><div class="text-caption text-muted">Loading...</div></div>`
      : renderTable({
          columns: getColumns(),
          data: filteredAliases,
          rowKey: 'subdomain',
          rowDataAttr: 'alias',
          clickable: true,
          emptyIcon: filter ? 'search-x' : 'link',
          emptyTitle: filter ? 'No aliases found' : 'No aliases yet',
          emptyMessage: filter ? 'Try a different search term' : 'Create your first alias to get started'
        })

    // Render toolbar
    const toolbar = renderToolbar({
      searchId: 'filter-input',
      searchValue: filter,
      searchPlaceholder: 'Filter...',
      buttons: [
        { id: 'new-alias-btn', icon: 'plus', title: 'New Alias', primary: true }
      ]
    })

    // Render footer
    const footer = filteredAliases.length > 0 ? renderTableFooter(filteredAliases.length, 'alias', 'aliases') : ''

    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">
            ${renderPanel({
              id: 'aliases.list',
              title: 'Aliases',
              count: aliasList.length,
              toolbar,
              content: tableContent,
              footer,
              minHeight: 400,
              maxHeight: 600
            })}
          </div>
        </div>
      </div>
    `

    // Setup handlers
    setupPanel(container)

    setupToolbar(container, {
      onSearch: (value) => {
        filter = value
        renderContent()
      },
      buttons: {
        'new-alias-btn': () => window.__fazt_openCreateAliasModal?.()
      }
    })

    setupTableClicks(container, 'alias', (subdomain) => {
      const alias = aliasList.find(a => a.subdomain === subdomain)
      if (alias) {
        window.__fazt_openEditAliasModal?.(alias)
      }
    })
  }

  /**
   * Re-render just the content area (for filtering)
   */
  function renderContent() {
    const aliasList = aliases.get()
    const isLoading = loading.getKey('aliases')

    const filteredAliases = filter
      ? aliasList.filter(alias =>
          alias.subdomain.toLowerCase().includes(filter.toLowerCase()) ||
          alias.type.toLowerCase().includes(filter.toLowerCase())
        )
      : aliasList

    const scrollArea = container.querySelector('.panel-scroll-area')
    if (!scrollArea) return

    if (isLoading) {
      scrollArea.innerHTML = `<div class="flex items-center justify-center p-8"><div class="text-caption text-muted">Loading...</div></div>`
      return
    }

    scrollArea.innerHTML = renderTable({
      columns: getColumns(),
      data: filteredAliases,
      rowKey: 'subdomain',
      rowDataAttr: 'alias',
      clickable: true,
      emptyIcon: filter ? 'search-x' : 'link',
      emptyTitle: filter ? 'No aliases found' : 'No aliases yet',
      emptyMessage: filter ? 'Try a different search term' : 'Create your first alias to get started'
    })

    if (window.lucide) window.lucide.createIcons()

    setupTableClicks(scrollArea, 'alias', (subdomain) => {
      const alias = aliasList.find(a => a.subdomain === subdomain)
      if (alias) {
        window.__fazt_openEditAliasModal?.(alias)
      }
    })

    // Update footer
    const panelInner = container.querySelector('.panel-inner')
    let footer = panelInner?.querySelector('.card-footer')
    if (filteredAliases.length > 0) {
      if (!footer) {
        panelInner?.insertAdjacentHTML('beforeend', renderTableFooter(filteredAliases.length, 'alias', 'aliases'))
      } else {
        footer.querySelector('span').textContent = `${filteredAliases.length} alias${filteredAliases.length === 1 ? '' : 'es'}`
      }
    } else if (footer) {
      footer.remove()
    }
  }

  // Subscribe to data changes
  const unsubAliases = aliases.subscribe(update)
  const unsubApps = apps.subscribe(update)
  const unsubLoading = loading.subscribeKey('aliases', update)

  // Initial render
  update()

  // Return cleanup function
  return () => {
    unsubAliases()
    unsubApps()
    unsubLoading()
  }
}
