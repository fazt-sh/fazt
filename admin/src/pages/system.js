/**
 * System Page
 */

import { health } from '../stores/data.js'

/**
 * Format uptime
 * @param {number} seconds
 */
function formatUptime(seconds) {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const mins = Math.floor((seconds % 3600) / 60)

  if (days > 0) return `${days}d ${hours}h ${mins}m`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

/**
 * Render system page
 * @param {HTMLElement} container
 * @param {Object} ctx
 */
export function render(container, ctx) {
  function update() {
    const data = health.get()

    container.innerHTML = `
      <div class="flex flex-col h-full overflow-hidden">
        <!-- Header -->
        <div class="flex items-center justify-between mb-4 flex-shrink-0">
          <div>
            <h1 class="text-title text-primary">System</h1>
            <p class="text-caption text-muted">Health and configuration</p>
          </div>
          <button class="btn btn-secondary" id="refresh-btn">
            <i data-lucide="refresh-cw" class="w-4 h-4"></i>
            Refresh
          </button>
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-auto scroll-panel">
          <div class="grid grid-cols-2 gap-4">
            <!-- Health Status -->
            <div class="card">
              <div class="card-header">
                <span class="text-heading text-primary">Health Status</span>
                <span class="flex items-center gap-1 text-caption ${data.status === 'healthy' ? 'text-success' : 'text-warning'}">
                  <span class="status-dot ${data.status === 'healthy' ? 'status-dot-success' : 'status-dot-warning'} pulse"></span>
                  ${data.status || 'Unknown'}
                </span>
              </div>
              <div class="card-body">
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <div class="text-micro text-muted mb-1">Uptime</div>
                    <div class="text-heading mono text-primary">${formatUptime(data.uptime_seconds || 0)}</div>
                  </div>
                  <div>
                    <div class="text-micro text-muted mb-1">Version</div>
                    <div class="text-heading mono text-primary">${data.version || 'Unknown'}</div>
                  </div>
                  <div>
                    <div class="text-micro text-muted mb-1">Mode</div>
                    <div class="text-heading text-primary">${data.mode || 'Unknown'}</div>
                  </div>
                  <div>
                    <div class="text-micro text-muted mb-1">Goroutines</div>
                    <div class="text-heading mono text-primary">${data.runtime?.goroutines || 0}</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Memory -->
            <div class="card">
              <div class="card-header">
                <span class="text-heading text-primary">Memory</span>
              </div>
              <div class="card-body">
                <div class="mb-4">
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-caption text-muted">Used</span>
                    <span class="text-caption mono text-primary">${(data.memory?.used_mb || 0).toFixed(1)} MB</span>
                  </div>
                  <div class="progress">
                    <div class="progress-bar" style="width: ${((data.memory?.used_mb || 0) / (data.memory?.limit_mb || 512)) * 100}%"></div>
                  </div>
                </div>
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <div class="text-micro text-muted mb-1">Limit</div>
                    <div class="text-heading mono text-primary">${(data.memory?.limit_mb || 0).toFixed(0)} MB</div>
                  </div>
                  <div>
                    <div class="text-micro text-muted mb-1">VFS Cache</div>
                    <div class="text-heading mono text-primary">${(data.memory?.vfs_cache_mb || 0).toFixed(1)} MB</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Database -->
            <div class="card">
              <div class="card-header">
                <span class="text-heading text-primary">Database</span>
              </div>
              <div class="card-body">
                <div class="mb-3">
                  <div class="text-micro text-muted mb-1">Path</div>
                  <div class="text-caption mono text-secondary truncate">${data.database?.path || 'Unknown'}</div>
                </div>
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <div class="text-micro text-muted mb-1">Open Connections</div>
                    <div class="text-heading mono text-primary">${data.database?.open_connections || 0}</div>
                  </div>
                  <div>
                    <div class="text-micro text-muted mb-1">In Use</div>
                    <div class="text-heading mono text-primary">${data.database?.in_use || 0}</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Runtime -->
            <div class="card">
              <div class="card-header">
                <span class="text-heading text-primary">Runtime</span>
              </div>
              <div class="card-body">
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <div class="text-micro text-muted mb-1">Queued Events</div>
                    <div class="text-heading mono text-primary">${data.runtime?.queued_events || 0}</div>
                  </div>
                  <div>
                    <div class="text-micro text-muted mb-1">Goroutines</div>
                    <div class="text-heading mono text-primary">${data.runtime?.goroutines || 0}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    `

    // Re-render Lucide icons
    if (window.lucide) {
      window.lucide.createIcons()
    }

    // Refresh button handler
    const refreshBtn = container.querySelector('#refresh-btn')
    if (refreshBtn && ctx.refresh) {
      refreshBtn.addEventListener('click', ctx.refresh)
    }
  }

  // Subscribe to data changes
  const unsub = health.subscribe(update)

  // Initial render
  update()

  // Return cleanup function
  return () => {
    unsub()
  }
}
