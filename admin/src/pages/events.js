/**
 * Events Page
 * Activity timeline for the fazt instance
 */

import { events, loadEvents } from '../stores/data.js'
import { loading } from '../stores/app.js'

/**
 * Format relative time
 * @param {string} timestamp
 */
function formatRelativeTime(timestamp) {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now - date
  const minutes = Math.floor(diff / (1000 * 60))
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) return `${days}d ago`
  if (hours > 0) return `${hours}h ago`
  if (minutes > 0) return `${minutes}m ago`
  return 'Just now'
}

/**
 * Format full timestamp
 * @param {string} timestamp
 */
function formatTimestamp(timestamp) {
  const date = new Date(timestamp)
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit'
  })
}

/**
 * Get event type icon
 * @param {string} eventType
 */
function getEventIcon(eventType) {
  const icons = {
    deploy: 'rocket',
    pageview: 'eye',
    api_call: 'code',
    config_update: 'settings',
    error: 'alert-circle'
  }
  return icons[eventType] || 'activity'
}

/**
 * Get event type color
 * @param {string} eventType
 */
function getEventColor(eventType) {
  const colors = {
    deploy: 'var(--success)',
    error: 'var(--error)',
    config_update: 'var(--warning)',
    pageview: 'var(--accent)',
    api_call: 'var(--text-3)'
  }
  return colors[eventType] || 'var(--text-3)'
}

/**
 * Get source type badge
 * @param {string} sourceType
 */
function getSourceBadge(sourceType) {
  const badges = {
    cli: { bg: 'var(--accent-soft)', color: 'var(--accent)' },
    web: { bg: 'var(--success-soft)', color: 'var(--success)' },
    api: { bg: 'var(--warning-soft)', color: 'var(--warning)' }
  }
  return badges[sourceType] || { bg: 'var(--bg-3)', color: 'var(--text-3)' }
}

/**
 * Render events page
 * @param {HTMLElement} container
 * @param {Object} ctx
 */
export function render(container, ctx) {
  const { client } = ctx
  let filterType = ''
  let filterSource = ''

  // Load events on mount
  loadEvents(client)

  function update() {
    const eventList = events.get()
    const isLoading = loading.getKey('events')

    // Apply filters
    let filteredEvents = [...eventList]
    if (filterType) {
      filteredEvents = filteredEvents.filter(e => e.event_type === filterType)
    }
    if (filterSource) {
      filteredEvents = filteredEvents.filter(e => e.source_type === filterSource)
    }

    // Get unique event types and sources for filters
    const eventTypes = [...new Set(eventList.map(e => e.event_type))]
    const sourceTypes = [...new Set(eventList.map(e => e.source_type))]

    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">

            <!-- Page Header -->
            <div class="flex items-center justify-between mb-4">
              <div>
                <h1 class="text-title text-primary">Events</h1>
                <p class="text-caption text-muted">${eventList.length} total events</p>
              </div>
              <div class="flex items-center gap-2">
                <select id="filter-type" class="btn btn-secondary text-label" style="padding: 6px 12px; cursor: pointer">
                  <option value="">All Types</option>
                  ${eventTypes.map(type => `
                    <option value="${type}" ${filterType === type ? 'selected' : ''}>${type}</option>
                  `).join('')}
                </select>
                <select id="filter-source" class="btn btn-secondary text-label" style="padding: 6px 12px; cursor: pointer">
                  <option value="">All Sources</option>
                  ${sourceTypes.map(source => `
                    <option value="${source}" ${filterSource === source ? 'selected' : ''}>${source}</option>
                  `).join('')}
                </select>
                <button id="refresh-btn" class="btn btn-secondary text-label" style="padding: 6px 12px" title="Refresh">
                  <i data-lucide="refresh-cw" class="w-4 h-4"></i>
                </button>
              </div>
            </div>

            <!-- Events List -->
            <div class="card">
              ${isLoading ? `
                <div class="flex items-center justify-center p-8">
                  <div class="text-caption text-muted">Loading events...</div>
                </div>
              ` : filteredEvents.length === 0 ? `
                <div class="flex items-center justify-center p-8">
                  <div class="text-center">
                    <i data-lucide="activity" class="w-12 h-12 text-faint mx-auto mb-2"></i>
                    <div class="text-heading text-primary mb-1">No events</div>
                    <div class="text-caption text-muted">${filterType || filterSource ? 'No events match your filters' : 'Activity will appear here'}</div>
                  </div>
                </div>
              ` : `
                <div class="activity-list">
                  ${filteredEvents.map(event => {
                    const icon = getEventIcon(event.event_type)
                    const color = getEventColor(event.event_type)
                    const sourceBadge = getSourceBadge(event.source_type)
                    return `
                      <div class="activity-item">
                        <div class="activity-icon">
                          <div class="icon-box" style="background: ${color}15">
                            <i data-lucide="${icon}" class="w-4 h-4" style="color: ${color}"></i>
                          </div>
                        </div>
                        <div class="activity-content">
                          <div class="flex items-center gap-2 mb-1">
                            <span class="text-label text-primary">${event.event_type}</span>
                            <span class="text-caption px-2 py-0.5" style="background: ${sourceBadge.bg}; color: ${sourceBadge.color}; border-radius: var(--radius-sm)">${event.source_type}</span>
                            ${event.tags?.length ? event.tags.map(tag => `
                              <span class="text-caption px-1.5 py-0.5" style="background: var(--bg-2); color: var(--text-3); border-radius: var(--radius-sm)">${tag}</span>
                            `).join('') : ''}
                          </div>
                          <div class="text-caption text-muted">
                            <span class="mono">${event.domain}</span>
                            ${event.path && event.path !== '/' ? `<span class="text-faint">${event.path}</span>` : ''}
                          </div>
                        </div>
                        <div class="activity-time text-right">
                          <div class="text-caption text-muted">${formatRelativeTime(event.created_at)}</div>
                          <div class="text-caption text-faint">${formatTimestamp(event.created_at)}</div>
                        </div>
                      </div>
                    `
                  }).join('')}
                </div>
              `}
            </div>

          </div>
        </div>
      </div>
    `

    // Re-render Lucide icons
    if (window.lucide) window.lucide.createIcons()

    // Filter handlers
    container.querySelector('#filter-type')?.addEventListener('change', (e) => {
      filterType = e.target.value
      update()
    })

    container.querySelector('#filter-source')?.addEventListener('change', (e) => {
      filterSource = e.target.value
      update()
    })

    // Refresh handler
    container.querySelector('#refresh-btn')?.addEventListener('click', () => {
      loadEvents(client, { event_type: filterType || undefined, source_type: filterSource || undefined })
    })
  }

  // Subscribe to data changes
  const unsubEvents = events.subscribe(update)
  const unsubLoading = loading.subscribeKey('events', update)

  // Initial render
  update()

  // Return cleanup function
  return () => {
    unsubEvents()
    unsubLoading()
  }
}
