/**
 * Apps Page & Detail
 */

import { apps, aliases, loadAppDetail, currentApp, deleteApp } from '../stores/data.js'
import { loading } from '../stores/app.js'

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
 * Render alias tags (max 3 + "..." badge)
 * @param {string[]} aliasList
 */
function renderAliasTags(aliasList) {
  if (!aliasList || aliasList.length === 0) {
    return '<span class="text-caption text-faint">No aliases</span>'
  }

  const visible = aliasList.slice(0, 3)
  const hasMore = aliasList.length > 3
  const moreCount = aliasList.length - 3

  return `
    <div class="flex items-center gap-1 flex-wrap">
      ${visible.map(alias => `
        <span class="text-caption px-2 py-0.5 mono" style="background:var(--bg-2);color:var(--text-2);border-radius:var(--radius-sm);border:1px solid var(--border-subtle)">${alias}</span>
      `).join('')}
      ${hasMore ? `<span class="text-caption px-2 py-0.5" style="background:var(--bg-3);color:var(--text-3);border-radius:var(--radius-sm);font-weight:500">+${moreCount}</span>` : ''}
    </div>
  `
}

/**
 * Render apps page or app detail
 * @param {HTMLElement} container
 * @param {Object} ctx
 */
export function render(container, ctx) {
  const { router, params, client, refresh } = ctx

  // If we have an ID param, render detail view
  if (params?.id) {
    return renderDetail(container, { router, client, appId: params.id, refresh })
  }

  // Otherwise render list view
  return renderList(container, { router, client, refresh })
}

/**
 * Render apps list page
 * @param {HTMLElement} container
 * @param {Object} ctx
 */
function renderList(container, ctx) {
  const { router } = ctx
  let filter = ''
  let viewMode = localStorage.getItem('apps-view-mode') || 'cards' // 'cards' or 'list'

  function renderAppsContent() {
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

    const filteredApps = filter
      ? appList.filter(app =>
          app.name.toLowerCase().includes(filter.toLowerCase()) ||
          app.id.toLowerCase().includes(filter.toLowerCase()) ||
          (appAliases[app.id] || []).some(a => a.toLowerCase().includes(filter.toLowerCase()))
        )
      : appList

    const appsContainer = container.querySelector('#apps-content')
    if (!appsContainer) return

    if (isLoading) {
      appsContainer.innerHTML = `
        <div class="flex items-center justify-center h-full">
          <div class="text-caption text-muted">Loading...</div>
        </div>
      `
      return
    }

    if (filteredApps.length === 0) {
      appsContainer.innerHTML = `
        <div class="flex items-center justify-center h-full">
          <div class="text-center">
            <div class="icon-box mx-auto mb-3" style="width:48px;height:48px">
              <i data-lucide="${filter ? 'search-x' : 'layers'}" class="w-6 h-6"></i>
            </div>
            <div class="text-heading text-primary mb-1">${filter ? 'No apps found' : 'No apps yet'}</div>
            <div class="text-caption text-muted mb-4">${filter ? 'Try a different search term' : 'Deploy your first app to get started'}</div>
            ${!filter ? `
              <button class="btn btn-primary text-label" style="padding: 8px 16px">
                <i data-lucide="plus" class="w-4 h-4 mr-1.5" style="display:inline-block;vertical-align:-2px"></i>
                Deploy App
              </button>
            ` : ''}
          </div>
        </div>
      `
      if (window.lucide) window.lucide.createIcons()
      return
    }

    // Cards view
    if (viewMode === 'cards') {
      appsContainer.innerHTML = `
        <div class="grid grid-cols-3 gap-4">
          ${filteredApps.map(app => {
            const appAliasesList = appAliases[app.id] || []
            return `
              <div class="card cursor-pointer" data-app-id="${app.id}">
                <div class="p-4">
                  <div class="flex items-start justify-between mb-3">
                    <div class="flex items-center gap-3">
                      <div class="icon-box">
                        <i data-lucide="box" class="w-4 h-4"></i>
                      </div>
                      <div>
                        <div class="text-heading text-primary">${app.name}</div>
                        <div class="text-caption mono text-faint">${app.id}</div>
                      </div>
                    </div>
                    <span class="flex items-center gap-1 text-caption text-success">
                      <span class="status-dot status-dot-success pulse"></span>
                      Live
                    </span>
                  </div>
                  <div class="mb-3">
                    ${renderAliasTags(appAliasesList)}
                  </div>
                  <div class="flex items-center gap-4 text-caption text-muted">
                    <span class="flex items-center gap-1">
                      <i data-lucide="file" class="w-3 h-3"></i>
                      ${app.file_count} files
                    </span>
                    <span class="flex items-center gap-1">
                      <i data-lucide="hard-drive" class="w-3 h-3"></i>
                      ${formatBytes(app.size_bytes)}
                    </span>
                  </div>
                </div>
                <div class="card-footer flex items-center justify-between">
                  <span class="text-caption text-muted">${formatRelativeTime(app.updated_at)}</span>
                  <span class="text-caption text-faint">${app.source}</span>
                </div>
              </div>
            `
          }).join('')}
        </div>
      `
    } else {
      // List view
      appsContainer.innerHTML = `
        <div class="card">
          <table style="table-layout: fixed">
            <colgroup>
              <col style="width: 18%">
              <col style="width: 26%">
              <col style="width: 16%">
              <col style="width: 8%">
              <col style="width: 10%">
              <col style="width: 12%">
              <col style="width: 10%">
            </colgroup>
            <thead>
              <tr style="border-bottom: 1px solid var(--border-subtle)">
                <th class="px-4 py-3 text-label text-primary">Name</th>
                <th class="px-4 py-3 text-label text-primary">Aliases</th>
                <th class="px-4 py-3 text-label text-primary">ID</th>
                <th class="px-4 py-3 text-label text-primary">Files</th>
                <th class="px-4 py-3 text-label text-primary">Size</th>
                <th class="px-4 py-3 text-label text-primary">Updated</th>
                <th class="px-4 py-3 text-label text-primary">Status</th>
              </tr>
            </thead>
            <tbody>
              ${filteredApps.map(app => `
                <tr class="row row-clickable" data-app-id="${app.id}" style="border-bottom: 1px solid var(--border-subtle)">
                  <td class="px-4 py-3">
                    <div class="flex items-center gap-2">
                      <div class="icon-box icon-box-sm">
                        <i data-lucide="box" class="w-3.5 h-3.5"></i>
                      </div>
                      <span class="text-label text-primary" style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis">${app.name}</span>
                    </div>
                  </td>
                  <td class="px-4 py-3">${renderAliasTags(appAliases[app.id] || [])}</td>
                  <td class="px-4 py-3 text-caption mono text-muted" style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis">${app.id}</td>
                  <td class="px-4 py-3 text-caption text-muted">${app.file_count}</td>
                  <td class="px-4 py-3 text-caption text-muted">${formatBytes(app.size_bytes)}</td>
                  <td class="px-4 py-3 text-caption text-muted">${formatRelativeTime(app.updated_at)}</td>
                  <td class="px-4 py-3">
                    <span class="flex items-center gap-1 text-caption text-success">
                      <span class="status-dot status-dot-success pulse"></span>
                      Live
                    </span>
                  </td>
                </tr>
              `).join('')}
            </tbody>
          </table>
        </div>
      `
    }

    // Re-render Lucide icons
    if (window.lucide) window.lucide.createIcons()

    // Add click handlers for app cards/rows
    appsContainer.querySelectorAll('[data-app-id]').forEach(el => {
      el.addEventListener('click', () => {
        router.push(`/apps/${el.dataset.appId}`)
      })
    })
  }

  function updateHeader() {
    const appList = apps.get()
    const count = container.querySelector('#apps-count')
    if (count) count.textContent = `${appList.length} app${appList.length === 1 ? '' : 's'} deployed`
  }

  function init() {
    container.innerHTML = `
      <div class="flex flex-col h-full overflow-hidden">
        <!-- Header -->
        <div class="flex items-center justify-between mb-4 flex-shrink-0">
          <div>
            <h1 class="text-title text-primary">Apps</h1>
            <p class="text-caption text-muted" id="apps-count">0 apps deployed</p>
          </div>
          <div class="flex items-center gap-2">
            <div class="input">
              <i data-lucide="search" class="w-4 h-4 text-faint"></i>
              <input type="text" id="filter-input" placeholder="Filter apps..." class="text-body" style="width: 200px">
            </div>
            <div class="flex items-center border" style="border-radius: var(--radius-sm); border-color: var(--border-subtle); background: var(--bg-1)">
              <button id="view-cards" class="btn-icon btn-ghost" title="Cards view" style="border-radius: var(--radius-sm) 0 0 var(--radius-sm)">
                <i data-lucide="layout-grid" class="w-4 h-4"></i>
              </button>
              <button id="view-list" class="btn-icon btn-ghost" title="List view" style="border-radius: 0 var(--radius-sm) var(--radius-sm) 0">
                <i data-lucide="list" class="w-4 h-4"></i>
              </button>
            </div>
            <button id="new-app-btn" class="btn btn-primary text-label" style="padding: 6px 12px; display: flex; align-items: center; gap: 6px">
              <i data-lucide="plus" class="w-4 h-4"></i>
              <span>New App</span>
            </button>
          </div>
        </div>

        <!-- Apps Content -->
        <div id="apps-content" class="flex-1 overflow-auto scroll-panel"></div>
      </div>
    `

    // Re-render icons
    if (window.lucide) window.lucide.createIcons()

    // Update view mode buttons
    function updateViewButtons() {
      const cardsBtn = container.querySelector('#view-cards')
      const listBtn = container.querySelector('#view-list')
      if (cardsBtn && listBtn) {
        if (viewMode === 'cards') {
          cardsBtn.style.color = 'var(--accent)'
          cardsBtn.style.background = 'var(--accent-soft)'
          listBtn.style.color = 'var(--text-3)'
          listBtn.style.background = 'transparent'
        } else {
          listBtn.style.color = 'var(--accent)'
          listBtn.style.background = 'var(--accent-soft)'
          cardsBtn.style.color = 'var(--text-3)'
          cardsBtn.style.background = 'transparent'
        }
      }
    }

    updateViewButtons()

    // Filter input handler - don't re-render entire container
    const filterInput = container.querySelector('#filter-input')
    if (filterInput) {
      filterInput.addEventListener('input', (e) => {
        filter = e.target.value
        renderAppsContent()
      })
    }

    // View mode handlers
    const cardsBtn = container.querySelector('#view-cards')
    const listBtn = container.querySelector('#view-list')

    cardsBtn?.addEventListener('click', () => {
      viewMode = 'cards'
      localStorage.setItem('apps-view-mode', 'cards')
      updateViewButtons()
      renderAppsContent()
    })

    listBtn?.addEventListener('click', () => {
      viewMode = 'list'
      localStorage.setItem('apps-view-mode', 'list')
      updateViewButtons()
      renderAppsContent()
    })

    // New app button
    container.querySelector('#new-app-btn')?.addEventListener('click', () => {
      const modal = document.getElementById('newAppModal')
      if (modal) {
        modal.classList.remove('hidden')
        modal.classList.add('flex')
      }
    })

    // Initial render
    updateHeader()
    renderAppsContent()
  }

  // Subscribe to data changes
  const unsubApps = apps.subscribe(() => {
    updateHeader()
    renderAppsContent()
  })
  const unsubAliases = aliases.subscribe(renderAppsContent)
  const unsubLoading = loading.subscribeKey('apps', renderAppsContent)

  // Initial render
  init()

  // Return cleanup function
  return () => {
    unsubApps()
    unsubAliases()
    unsubLoading()
  }
}

/**
 * Render app detail page
 * @param {HTMLElement} container
 * @param {Object} ctx
 */
function renderDetail(container, ctx) {
  const { router, client, appId, refresh } = ctx

  function renderContent() {
    const app = currentApp.get()
    const aliasList = aliases.get()
    const isLoading = loading.getKey('apps')

    // Get aliases for this app
    const appAliasesList = aliasList
      .filter(a => a.type === 'proxy' && a.targets?.app_id === appId)
      .map(a => a.subdomain)

    container.innerHTML = `
      <div class="flex flex-col h-full overflow-hidden">
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
          <!-- Header -->
          <div class="flex items-center justify-between mb-4 flex-shrink-0">
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

          <!-- Content -->
          <div class="flex-1 overflow-auto scroll-panel">
            <div class="grid grid-cols-3 gap-4 mb-4">
              <!-- Stats Cards -->
              <div class="card">
                <div class="card-body">
                  <div class="flex items-center gap-3">
                    <div class="icon-box">
                      <i data-lucide="file" class="w-4 h-4"></i>
                    </div>
                    <div>
                      <div class="text-caption text-muted">Files</div>
                      <div class="text-heading text-primary">${app.file_count}</div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="card">
                <div class="card-body">
                  <div class="flex items-center gap-3">
                    <div class="icon-box">
                      <i data-lucide="hard-drive" class="w-4 h-4"></i>
                    </div>
                    <div>
                      <div class="text-caption text-muted">Size</div>
                      <div class="text-heading text-primary">${formatBytes(app.size_bytes)}</div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="card">
                <div class="card-body">
                  <div class="flex items-center gap-3">
                    <div class="icon-box">
                      <i data-lucide="clock" class="w-4 h-4"></i>
                    </div>
                    <div>
                      <div class="text-caption text-muted">Updated</div>
                      <div class="text-heading text-primary">${formatRelativeTime(app.updated_at)}</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <!-- Details Card -->
              <div class="card">
                <div class="card-header">
                  <span class="text-heading text-primary">Details</span>
                </div>
                <div class="card-body space-y-3">
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

              <!-- Aliases Card -->
              <div class="card">
                <div class="card-header">
                  <span class="text-heading text-primary">Aliases</span>
                  <button class="btn btn-sm btn-secondary">
                    <i data-lucide="plus" class="w-3 h-3 mr-1" style="display:inline-block;vertical-align:-1px"></i>
                    Add
                  </button>
                </div>
                <div class="card-body">
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

            <!-- Files Card -->
            <div class="card mt-4">
              <div class="card-header">
                <span class="text-heading text-primary">Files</span>
              </div>
              <div class="card-body p-0">
                ${app.files && app.files.length > 0 ? `
                  <table style="table-layout: fixed">
                    <colgroup>
                      <col style="width: 55%">
                      <col style="width: 15%">
                      <col style="width: 30%">
                    </colgroup>
                    <thead>
                      <tr style="border-bottom: 1px solid var(--border-subtle)">
                        <th class="px-4 py-3 text-label text-primary">Path</th>
                        <th class="px-4 py-3 text-label text-primary">Size</th>
                        <th class="px-4 py-3 text-label text-primary">Type</th>
                      </tr>
                    </thead>
                    <tbody>
                      ${app.files.map(file => `
                        <tr class="row" style="border-bottom: 1px solid var(--border-subtle)">
                          <td class="px-4 py-2 text-label mono text-primary" style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis">${file.path}</td>
                          <td class="px-4 py-2 text-caption text-muted">${formatBytes(file.size)}</td>
                          <td class="px-4 py-2 text-caption text-muted" style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis">${file.mime_type}</td>
                        </tr>
                      `).join('')}
                    </tbody>
                  </table>
                ` : `
                  <div class="p-4 text-center text-caption text-muted">No files</div>
                `}
              </div>
            </div>
          </div>
        `}
      </div>
    `

    // Re-render Lucide icons
    if (window.lucide) window.lucide.createIcons()

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

    // Alias buttons - open in new tab
    container.querySelectorAll('.alias-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        const alias = btn.dataset.alias
        // Construct URL based on current location
        const protocol = window.location.protocol
        const hostname = window.location.hostname
        const port = window.location.port
        const url = `${protocol}//${alias}.${hostname}${port ? ':' + port : ''}`
        window.open(url, '_blank')
      })
    })
  }

  // Load app detail
  loadAppDetail(client, appId)

  // Subscribe to changes
  const unsubApp = currentApp.subscribe(renderContent)
  const unsubAliases = aliases.subscribe(renderContent)
  const unsubLoading = loading.subscribeKey('apps', renderContent)

  // Initial render
  renderContent()

  // Return cleanup
  return () => {
    unsubApp()
    unsubAliases()
    unsubLoading()
  }
}
