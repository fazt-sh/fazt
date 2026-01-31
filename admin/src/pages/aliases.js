/**
 * Aliases Page
 */

import { aliases } from '../stores/data.js'
import { loading } from '../stores/app.js'

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
 * Get type badge class
 * @param {string} type
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
 * Render aliases page
 * @param {HTMLElement} container
 * @param {Object} ctx
 */
export function render(container, ctx) {
  let filter = ''

  function update() {
    const aliasList = aliases.get()
    const isLoading = loading.getKey('aliases')

    const filteredAliases = filter
      ? aliasList.filter(alias =>
          alias.subdomain.toLowerCase().includes(filter.toLowerCase())
        )
      : aliasList

    container.innerHTML = `
      <div class="flex flex-col h-full overflow-hidden">
        <!-- Header -->
        <div class="flex items-center justify-between mb-4 flex-shrink-0">
          <div>
            <h1 class="text-title text-primary">Aliases</h1>
            <p class="text-caption text-muted">${aliasList.length} subdomains configured</p>
          </div>
          <div class="flex items-center gap-2">
            <div class="input">
              <i data-lucide="search" class="w-4 h-4 text-faint"></i>
              <input type="text" id="filter-input" placeholder="Filter aliases..." class="text-body" value="${filter}" style="width: 200px">
            </div>
            <button class="btn btn-primary">
              <i data-lucide="plus" class="w-4 h-4"></i>
              New Alias
            </button>
          </div>
        </div>

        <!-- Aliases Table -->
        <div class="card flex flex-col overflow-hidden flex-1">
          <div class="flex-1 overflow-auto scroll-panel">
            ${isLoading ? `
              <div class="flex items-center justify-center h-full">
                <div class="text-caption text-muted">Loading...</div>
              </div>
            ` : filteredAliases.length === 0 ? `
              <div class="flex items-center justify-center h-full">
                <div class="text-center">
                  <i data-lucide="link" class="w-12 h-12 text-faint mx-auto mb-2"></i>
                  <div class="text-caption text-muted">${filter ? 'No aliases match your filter' : 'No aliases configured'}</div>
                </div>
              </div>
            ` : `
              <table>
                <thead class="sticky" style="top: 0; background: var(--bg-1)">
                  <tr class="border-b">
                    <th class="px-4 py-3 text-left text-micro text-muted">Subdomain</th>
                    <th class="px-4 py-3 text-left text-micro text-muted">Type</th>
                    <th class="px-4 py-3 text-left text-micro text-muted">Target</th>
                    <th class="px-4 py-3 text-left text-micro text-muted">Updated</th>
                    <th class="px-4 py-3"></th>
                  </tr>
                </thead>
                <tbody>
                  ${filteredAliases.map(alias => `
                    <tr class="row row-clickable border-b" data-alias="${alias.subdomain}" style="border-color: var(--border-subtle)">
                      <td class="px-4 py-3">
                        <div class="flex items-center gap-2">
                          <div class="icon-box icon-box-sm">
                            <i data-lucide="${alias.type === 'redirect' ? 'external-link' : alias.type === 'reserved' ? 'lock' : 'link'}" class="w-3.5 h-3.5"></i>
                          </div>
                          <div>
                            <div class="text-label text-primary">${alias.subdomain}</div>
                            <div class="text-caption mono text-faint">${alias.subdomain}.zyt.app</div>
                          </div>
                        </div>
                      </td>
                      <td class="px-4 py-3">
                        <span class="badge ${getTypeBadge(alias.type)}">${alias.type}</span>
                      </td>
                      <td class="px-4 py-3">
                        <span class="text-caption mono text-muted">
                          ${alias.type === 'proxy' ? alias.targets?.app_id || '-' :
                            alias.type === 'redirect' ? alias.targets?.url || '-' :
                            '-'}
                        </span>
                      </td>
                      <td class="px-4 py-3 text-caption text-muted">${formatRelativeTime(alias.updated_at)}</td>
                      <td class="px-4 py-3">
                        <button class="btn btn-ghost btn-icon btn-sm" style="color: var(--text-4)">
                          <i data-lucide="more-horizontal" class="w-4 h-4"></i>
                        </button>
                      </td>
                    </tr>
                  `).join('')}
                </tbody>
              </table>
            `}
          </div>
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
  }

  // Subscribe to data changes
  const unsubAliases = aliases.subscribe(update)
  const unsubLoading = loading.subscribeKey('aliases', update)

  // Initial render
  update()

  // Return cleanup function
  return () => {
    unsubAliases()
    unsubLoading()
  }
}
