/**
 * Apps Page & Detail
 */

import { apps, aliases, loadAppDetail, currentApp, deleteApp } from '../stores/data.js'
import { loading } from '../stores/app.js'
import {
  renderPanel, setupPanel, getUIState, setUIState,
  renderToolbar, setupToolbar,
  renderTable, setupTableClicks, renderTableFooter,
  renderIconCell, renderStatusCell
} from '../components/index.js'

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
 * Render apps page or app detail
 */
export function render(container, ctx) {
  const { router, params, client, refresh } = ctx

  if (params?.id) {
    return renderDetail(container, { router, client, appId: params.id, refresh })
  }

  return renderList(container, { router, client, refresh })
}

/**
 * Render apps list page
 */
function renderList(container, ctx) {
  const { router } = ctx
  let filter = ''

  function update() {
    const appList = apps.get()
    const aliasList = aliases.get()
    const isLoading = loading.getKey('apps')

    // Map aliases to apps
    const appAliases = {}
    aliasList.forEach(alias => {
      if (alias.type === 'proxy' && alias.targets?.app_id) {
        const appId = alias.targets.app_id
        if (!appAliases[appId]) appAliases[appId] = []
        appAliases[appId].push(alias.subdomain)
      }
    })

    // Filter apps
    const filteredApps = filter
      ? appList.filter(app =>
          app.name.toLowerCase().includes(filter.toLowerCase()) ||
          app.id.toLowerCase().includes(filter.toLowerCase()) ||
          (appAliases[app.id] || []).some(a => a.toLowerCase().includes(filter.toLowerCase()))
        )
      : appList

    // Table columns - mobile shows only Name (with status dot)
    const columns = [
      {
        key: 'name',
        label: 'Name',
        render: (name, app) => `
          <div class="flex items-center gap-2" style="min-width: 0">
            <span class="status-dot status-dot-success pulse show-mobile" style="flex-shrink: 0"></span>
            <div class="icon-box icon-box-sm" style="flex-shrink: 0">
              <i data-lucide="box" class="w-3.5 h-3.5"></i>
            </div>
            <div style="min-width: 0; overflow: hidden">
              <div class="text-label text-primary" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">${name}</div>
              <div class="text-caption mono text-faint show-mobile">${app.file_count} files · ${formatBytes(app.size_bytes)}</div>
            </div>
          </div>
        `
      },
      {
        key: 'aliases',
        label: 'Aliases',
        hideOnMobile: true,
        render: (_, app) => {
          const list = appAliases[app.id] || []
          if (list.length === 0) return '<span class="text-caption text-faint">-</span>'
          return `<span class="text-caption mono text-muted">${list[0]}${list.length > 1 ? ` +${list.length - 1}` : ''}</span>`
        }
      },
      {
        key: 'file_count',
        label: 'Files',
        hideOnMobile: true,
        render: (v) => `<span class="text-caption text-muted">${v}</span>`
      },
      {
        key: 'size_bytes',
        label: 'Size',
        hideOnMobile: true,
        render: (v) => `<span class="text-caption text-muted">${formatBytes(v)}</span>`
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
        render: () => renderStatusCell('Live', true)
      }
    ]

    // Render table content
    const tableContent = isLoading
      ? `<div class="flex items-center justify-center p-8"><div class="text-caption text-muted">Loading...</div></div>`
      : renderTable({
          columns,
          data: filteredApps,
          rowKey: 'id',
          rowDataAttr: 'app-id',
          clickable: true,
          emptyIcon: filter ? 'search-x' : 'layers',
          emptyTitle: filter ? 'No apps found' : 'No apps yet',
          emptyMessage: filter ? 'Try a different search term' : 'Deploy your first app via CLI'
        })

    // Render toolbar
    const toolbar = renderToolbar({
      searchId: 'filter-input',
      searchValue: filter,
      searchPlaceholder: 'Filter...',
      buttons: [] // No buttons - deploy via CLI
    })

    // Render panel
    const footer = filteredApps.length > 0 ? renderTableFooter(filteredApps.length, 'app') : ''

    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">
            ${renderPanel({
              id: 'apps.list',
              title: 'Apps',
              count: appList.length,
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
      }
    })

    setupTableClicks(container, 'app-id', (id) => {
      router.push(`/apps/${id}`)
    })
  }

  /**
   * Re-render just the content area (for filtering)
   */
  function renderContent() {
    const appList = apps.get()
    const aliasList = aliases.get()
    const isLoading = loading.getKey('apps')

    const appAliases = {}
    aliasList.forEach(alias => {
      if (alias.type === 'proxy' && alias.targets?.app_id) {
        const appId = alias.targets.app_id
        if (!appAliases[appId]) appAliases[appId] = []
        appAliases[appId].push(alias.subdomain)
      }
    })

    const filteredApps = filter
      ? appList.filter(app =>
          app.name.toLowerCase().includes(filter.toLowerCase()) ||
          app.id.toLowerCase().includes(filter.toLowerCase()) ||
          (appAliases[app.id] || []).some(a => a.toLowerCase().includes(filter.toLowerCase()))
        )
      : appList

    const columns = [
      {
        key: 'name',
        label: 'Name',
        render: (name, app) => `
          <div class="flex items-center gap-2" style="min-width: 0">
            <span class="status-dot status-dot-success pulse show-mobile" style="flex-shrink: 0"></span>
            <div class="icon-box icon-box-sm" style="flex-shrink: 0">
              <i data-lucide="box" class="w-3.5 h-3.5"></i>
            </div>
            <div style="min-width: 0; overflow: hidden">
              <div class="text-label text-primary" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">${name}</div>
              <div class="text-caption mono text-faint show-mobile">${app.file_count} files · ${formatBytes(app.size_bytes)}</div>
            </div>
          </div>
        `
      },
      {
        key: 'aliases',
        label: 'Aliases',
        hideOnMobile: true,
        render: (_, app) => {
          const list = appAliases[app.id] || []
          if (list.length === 0) return '<span class="text-caption text-faint">-</span>'
          return `<span class="text-caption mono text-muted">${list[0]}${list.length > 1 ? ` +${list.length - 1}` : ''}</span>`
        }
      },
      {
        key: 'file_count',
        label: 'Files',
        hideOnMobile: true,
        render: (v) => `<span class="text-caption text-muted">${v}</span>`
      },
      {
        key: 'size_bytes',
        label: 'Size',
        hideOnMobile: true,
        render: (v) => `<span class="text-caption text-muted">${formatBytes(v)}</span>`
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
        render: () => renderStatusCell('Live', true)
      }
    ]

    const scrollArea = container.querySelector('.panel-scroll-area')
    if (!scrollArea) return

    if (isLoading) {
      scrollArea.innerHTML = `<div class="flex items-center justify-center p-8"><div class="text-caption text-muted">Loading...</div></div>`
      return
    }

    scrollArea.innerHTML = renderTable({
      columns,
      data: filteredApps,
      rowKey: 'id',
      rowDataAttr: 'app-id',
      clickable: true,
      emptyIcon: filter ? 'search-x' : 'layers',
      emptyTitle: filter ? 'No apps found' : 'No apps yet',
      emptyMessage: filter ? 'Try a different search term' : 'Deploy your first app via CLI'
    })

    if (window.lucide) window.lucide.createIcons()

    setupTableClicks(scrollArea, 'app-id', (id) => {
      router.push(`/apps/${id}`)
    })

    // Update footer
    const panelInner = container.querySelector('.panel-inner')
    let footer = panelInner?.querySelector('.card-footer')
    if (filteredApps.length > 0) {
      if (!footer) {
        panelInner?.insertAdjacentHTML('beforeend', renderTableFooter(filteredApps.length, 'app'))
      } else {
        footer.querySelector('span').textContent = `${filteredApps.length} app${filteredApps.length === 1 ? '' : 's'}`
      }
    } else if (footer) {
      footer.remove()
    }
  }

  // Subscribe to data changes
  const unsubApps = apps.subscribe(update)
  const unsubAliases = aliases.subscribe(update)
  const unsubLoading = loading.subscribeKey('apps', update)

  // Initial render
  update()

  return () => {
    unsubApps()
    unsubAliases()
    unsubLoading()
  }
}

/**
 * Render app detail page
 */
function renderDetail(container, ctx) {
  const { router, client, appId, refresh } = ctx

  function renderContent() {
    const app = currentApp.get()
    const aliasList = aliases.get()
    const isLoading = loading.getKey('apps')

    const appAliasesList = aliasList
      .filter(a => a.type === 'proxy' && a.targets?.app_id === appId)
      .map(a => a.subdomain)

    const detailsCollapsed = getUIState(`apps.detail.${appId}.details.collapsed`, false)
    const aliasesCollapsed = getUIState(`apps.detail.${appId}.aliases.collapsed`, false)
    const filesCollapsed = getUIState(`apps.detail.${appId}.files.collapsed`, false)

    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">
            ${isLoading ? `
              <div class="flex items-center justify-center h-full">
                <div class="text-caption text-muted">Loading...</div>
              </div>
            ` : !app.id ? `
              <div class="flex items-center justify-center h-full">
                <div class="text-center">
                  <i data-lucide="alert-circle" class="w-12 h-12 text-faint mx-auto mb-2"></i>
                  <div class="text-heading text-primary mb-1">App not found</div>
                  <div class="text-caption text-muted mb-4">The app you're looking for doesn't exist</div>
                  <button id="back-btn" class="btn btn-primary text-label" style="padding: 6px 12px">
                    <i data-lucide="arrow-left" class="w-4 h-4 mr-1.5" style="display:inline-block;vertical-align:-2px"></i>
                    Back to Apps
                  </button>
                </div>
              </div>
            ` : `
              <!-- Page Header -->
              <div class="flex items-center justify-between mb-4">
                <div class="flex items-center gap-3">
                  <button id="back-btn" class="btn-icon btn-ghost" style="color:var(--text-3)">
                    <i data-lucide="arrow-left" class="w-4 h-4"></i>
                  </button>
                  <div>
                    <h1 class="text-title text-primary">${app.name}</h1>
                    <p class="text-caption mono text-muted">${app.id}</p>
                  </div>
                </div>
                <div class="flex items-center gap-2">
                  <button id="refresh-btn" class="btn btn-secondary text-label" style="padding: 6px 12px">
                    <i data-lucide="refresh-cw" class="w-4 h-4"></i>
                  </button>
                  <button id="delete-btn" class="btn btn-secondary text-label" style="padding: 6px 12px; color: var(--error)">
                    <i data-lucide="trash-2" class="w-4 h-4"></i>
                  </button>
                </div>
              </div>

              <!-- Stats Grid -->
              <div class="panel-grid grid-3 mb-4">
                <div class="stat-card card">
                  <div class="stat-card-header">
                    <span class="text-micro text-muted">Files</span>
                    <i data-lucide="file" class="w-4 h-4 text-faint"></i>
                  </div>
                  <div class="stat-card-value text-display mono text-primary">${app.file_count}</div>
                </div>
                <div class="stat-card card">
                  <div class="stat-card-header">
                    <span class="text-micro text-muted">Size</span>
                    <i data-lucide="hard-drive" class="w-4 h-4 text-faint"></i>
                  </div>
                  <div class="stat-card-value text-display mono text-primary">${formatBytes(app.size_bytes)}</div>
                </div>
                <div class="stat-card card">
                  <div class="stat-card-header">
                    <span class="text-micro text-muted">Updated</span>
                    <i data-lucide="clock" class="w-4 h-4 text-faint"></i>
                  </div>
                  <div class="stat-card-value text-heading text-primary">${formatRelativeTime(app.updated_at)}</div>
                </div>
              </div>

              <!-- Panel Group: Details -->
              <div class="panel-group ${detailsCollapsed ? 'collapsed' : ''}">
                <div class="panel-group-card card">
                  <header class="panel-group-header" data-panel="apps.detail.${appId}.details">
                    <button class="collapse-toggle">
                      <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                      <span class="text-heading text-primary">Details</span>
                    </button>
                  </header>
                  <div class="panel-group-body">
                    <div class="space-y-3">
                      <div>
                        <div class="text-micro text-muted mb-1">SOURCE</div>
                        <div class="text-label text-primary">${app.source}</div>
                      </div>
                      <div>
                        <div class="text-micro text-muted mb-1">CREATED</div>
                        <div class="text-label text-primary">${formatRelativeTime(app.created_at)}</div>
                      </div>
                      ${app.manifest?.version ? `
                        <div>
                          <div class="text-micro text-muted mb-1">VERSION</div>
                          <div class="text-label text-primary mono">${app.manifest.version}</div>
                        </div>
                      ` : ''}
                    </div>
                  </div>
                </div>
              </div>

              <!-- Panel Group: Aliases -->
              <div class="panel-group ${aliasesCollapsed ? 'collapsed' : ''}">
                <div class="panel-group-card card">
                  <header class="panel-group-header" data-panel="apps.detail.${appId}.aliases">
                    <button class="collapse-toggle">
                      <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                      <span class="text-heading text-primary">Aliases</span>
                      <span class="text-caption text-faint ml-auto hide-mobile">${appAliasesList.length} configured</span>
                    </button>
                  </header>
                  <div class="panel-group-body">
                    ${appAliasesList.length === 0 ? `
                      <div class="text-center py-4">
                        <div class="text-caption text-muted">No aliases configured</div>
                      </div>
                    ` : `
                      <div class="flex flex-wrap gap-2">
                        ${appAliasesList.map(alias => `
                          <button class="btn btn-secondary text-label alias-btn" data-alias="${alias}" style="padding: 6px 12px; display: flex; align-items: center; gap: 6px">
                            <span class="mono">${alias}</span>
                            <i data-lucide="external-link" class="w-3 h-3"></i>
                          </button>
                        `).join('')}
                      </div>
                    `}
                  </div>
                </div>
              </div>

              <!-- Panel Group: Files -->
              <div class="panel-group ${filesCollapsed ? 'collapsed' : ''}">
                <div class="panel-group-card card">
                  <header class="panel-group-header" data-panel="apps.detail.${appId}.files">
                    <button class="collapse-toggle">
                      <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                      <span class="text-heading text-primary">Files</span>
                      <span class="text-caption text-faint ml-auto hide-mobile">${app.files?.length || 0} files</span>
                    </button>
                  </header>
                  <div class="panel-group-body" style="padding: 0">
                    ${app.files && app.files.length > 0 ? `
                      <div class="table-container">
                        <table style="table-layout: fixed">
                        <colgroup>
                          <col style="width: 55%">
                          <col style="width: 15%">
                          <col style="width: 30%">
                        </colgroup>
                        <thead>
                          <tr style="border-bottom: 1px solid var(--border-subtle)">
                            <th class="px-4 py-3 text-label text-primary">Path</th>
                            <th class="px-4 py-3 text-label text-primary hide-mobile">Size</th>
                            <th class="px-4 py-3 text-label text-primary hide-mobile">Type</th>
                          </tr>
                        </thead>
                        <tbody>
                          ${app.files.map(file => `
                            <tr class="row" style="border-bottom: 1px solid var(--border-subtle)">
                              <td class="px-4 py-2 text-label mono text-primary" style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis">${file.path}</td>
                              <td class="px-4 py-2 text-caption text-muted hide-mobile">${formatBytes(file.size)}</td>
                              <td class="px-4 py-2 text-caption text-muted hide-mobile" style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis">${file.mime_type}</td>
                            </tr>
                          `).join('')}
                        </tbody>
                      </table>
                      </div>
                    ` : `
                      <div class="p-4 text-center text-caption text-muted">No files</div>
                    `}
                  </div>
                </div>
              </div>
            `}
          </div>
        </div>
      </div>
    `

    if (window.lucide) window.lucide.createIcons()

    // Setup collapse handlers
    setupPanel(container)

    // Back button
    container.querySelector('#back-btn')?.addEventListener('click', () => {
      router.push('/apps')
    })

    // Refresh button
    container.querySelector('#refresh-btn')?.addEventListener('click', async () => {
      await loadAppDetail(client, appId)
      if (refresh) await refresh()
    })

    // Delete button
    container.querySelector('#delete-btn')?.addEventListener('click', async () => {
      if (confirm(`Delete app "${app.name}"? This cannot be undone.`)) {
        const success = await deleteApp(client, appId)
        if (success) router.push('/apps')
      }
    })

    // Alias buttons
    container.querySelectorAll('.alias-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        const alias = btn.dataset.alias
        const protocol = window.location.protocol
        const hostname = window.location.hostname
        const port = window.location.port
        const url = `${protocol}//${alias}.${hostname}${port ? ':' + port : ''}`
        window.open(url, '_blank')
      })
    })
  }

  loadAppDetail(client, appId)

  const unsubApp = currentApp.subscribe(renderContent)
  const unsubAliases = aliases.subscribe(renderContent)
  const unsubLoading = loading.subscribeKey('apps', renderContent)

  renderContent()

  return () => {
    unsubApp()
    unsubAliases()
    unsubLoading()
  }
}
