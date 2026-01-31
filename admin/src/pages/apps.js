/**
 * Apps Page
 */

import { apps } from '../stores/data.js'
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
 * Render apps page
 * @param {HTMLElement} container
 * @param {Object} ctx
 */
export function render(container, ctx) {
  const { router } = ctx
  let filter = ''

  function update() {
    const appList = apps.get()
    const isLoading = loading.getKey('apps')

    const filteredApps = filter
      ? appList.filter(app =>
          app.name.toLowerCase().includes(filter.toLowerCase())
        )
      : appList

    container.innerHTML = `
      <div class="flex flex-col h-full overflow-hidden">
        <!-- Header -->
        <div class="flex items-center justify-between mb-4 flex-shrink-0">
          <div>
            <h1 class="text-title text-primary">Apps</h1>
            <p class="text-caption text-muted">${appList.length} apps deployed</p>
          </div>
          <div class="flex items-center gap-2">
            <div class="input">
              <i data-lucide="search" class="w-4 h-4 text-faint"></i>
              <input type="text" id="filter-input" placeholder="Filter apps..." class="text-body" value="${filter}" style="width: 200px">
            </div>
            <button class="btn btn-primary">
              <i data-lucide="plus" class="w-4 h-4"></i>
              New App
            </button>
          </div>
        </div>

        <!-- Apps Grid -->
        <div class="flex-1 overflow-auto scroll-panel">
          ${isLoading ? `
            <div class="flex items-center justify-center h-full">
              <div class="text-caption text-muted">Loading...</div>
            </div>
          ` : filteredApps.length === 0 ? `
            <div class="flex items-center justify-center h-full">
              <div class="text-center">
                <i data-lucide="inbox" class="w-12 h-12 text-faint mx-auto mb-2"></i>
                <div class="text-caption text-muted">${filter ? 'No apps match your filter' : 'No apps yet'}</div>
              </div>
            </div>
          ` : `
            <div class="grid grid-cols-3 gap-4">
              ${filteredApps.map(app => `
                <div class="card cursor-pointer" data-app-id="${app.id}">
                  <div class="p-4">
                    <div class="flex items-start justify-between mb-3">
                      <div class="flex items-center gap-3">
                        <div class="icon-box">
                          <i data-lucide="box" class="w-4 h-4"></i>
                        </div>
                        <div>
                          <div class="text-heading text-primary">${app.name}</div>
                          <div class="text-caption mono text-faint">${app.name}.zyt.app</div>
                        </div>
                      </div>
                      <span class="flex items-center gap-1 text-caption text-success">
                        <span class="status-dot status-dot-success pulse"></span>
                        Live
                      </span>
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
                  <div class="px-4 py-2 border-t flex items-center justify-between" style="background: var(--bg-2)">
                    <span class="text-caption text-muted">${formatRelativeTime(app.updated_at)}</span>
                    <span class="text-caption text-faint">${app.source}</span>
                  </div>
                </div>
              `).join('')}
            </div>
          `}
        </div>
      </div>
    `

    // Re-render Lucide icons
    if (window.lucide) {
      window.lucide.createIcons()
    }

    // Add filter input handler
    const filterInput = container.querySelector('#filter-input')
    if (filterInput) {
      filterInput.addEventListener('input', (e) => {
        filter = e.target.value
        update()
      })
    }

    // Add click handlers for app cards
    container.querySelectorAll('[data-app-id]').forEach(card => {
      card.addEventListener('click', () => {
        router.push(`/apps/${card.dataset.appId}`)
      })
    })
  }

  // Subscribe to data changes
  const unsubApps = apps.subscribe(update)
  const unsubLoading = loading.subscribeKey('apps', update)

  // Initial render
  update()

  // Return cleanup function
  return () => {
    unsubApps()
    unsubLoading()
  }
}
