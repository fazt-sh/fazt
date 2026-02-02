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
 * Get UI state from localStorage
 */
function getUIState(key, defaultValue = false) {
  try {
    const state = JSON.parse(localStorage.getItem('fazt.web.ui.state') || '{}')
    return state[key] !== undefined ? state[key] : defaultValue
  } catch {
    return defaultValue
  }
}

/**
 * Set UI state to localStorage
 */
function setUIState(key, value) {
  try {
    const state = JSON.parse(localStorage.getItem('fazt.web.ui.state') || '{}')
    state[key] = value
    localStorage.setItem('fazt.web.ui.state', JSON.stringify(state))
  } catch (e) {
    console.error('Failed to save UI state:', e)
  }
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

    const listCollapsed = getUIState('events.list.collapsed', false)

    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">

            <!-- Panel Group: Events List -->
            <div class="panel-group ${listCollapsed ? 'collapsed' : ''}">
              <div class="panel-group-card card">
                <header class="panel-group-header" data-group="list">
                  <button class="collapse-toggle">
                    <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                    <span class="text-heading text-primary">Events</span>
                    <span class="text-caption mono px-1.5 py-0.5 ml-2 badge-muted" style="border-radius: var(--radius-sm)">${eventList.length}</span>
                  </button>
                  <div class="flex items-center gap-2 ml-auto">
                    <select id="filter-type" class="btn btn-secondary btn-sm hide-mobile" style="padding: 4px 8px; cursor: pointer">
                      <option value="">All Types</option>
                      ${eventTypes.map(type => `
                        <option value="${type}" ${filterType === type ? 'selected' : ''}>${type}</option>
                      `).join('')}
                    </select>
                    <select id="filter-source" class="btn btn-secondary btn-sm hide-mobile" style="padding: 4px 8px; cursor: pointer">
                      <option value="">All Sources</option>
                      ${sourceTypes.map(source => `
                        <option value="${source}" ${filterSource === source ? 'selected' : ''}>${source}</option>
                      `).join('')}
                    </select>
                    <button id="refresh-btn" class="btn btn-secondary btn-sm" style="padding: 4px 8px" title="Refresh">
                      <i data-lucide="refresh-cw" class="w-3.5 h-3.5"></i>
                    </button>
                  </div>
                </header>
                <div class="panel-group-body" style="padding: 0">
                  <div class="flex-1 overflow-auto scroll-panel" style="max-height: 600px">
                    ${isLoading ? `
                      <div class="flex items-center justify-center p-8">
                        <div class="text-caption text-muted">Loading events...</div>
                      </div>
                    ` : filteredEvents.length === 0 ? `
                      <div class="flex flex-col items-center justify-center p-8 text-center">
                        <div class="icon-box mb-3" style="width:48px;height:48px;opacity:0.5">
                          <i data-lucide="activity" class="w-6 h-6"></i>
                        </div>
                        <div class="text-heading text-primary mb-1">No events</div>
                        <div class="text-caption text-muted">${filterType || filterSource ? 'No events match your filters' : 'Activity will appear here'}</div>
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
                                <div class="flex items-center gap-2 mb-1 flex-wrap">
                                  <span class="text-label text-primary">${event.event_type}</span>
                                  <span class="text-caption px-2 py-0.5" style="background: ${sourceBadge.bg}; color: ${sourceBadge.color}; border-radius: var(--radius-sm)">${event.source_type}</span>
                                  <span class="text-caption text-muted show-mobile">${formatRelativeTime(event.created_at)}</span>
                                </div>
                                <div class="text-caption text-muted">
                                  <span class="mono">${event.domain}</span>
                                  ${event.path && event.path !== '/' ? `<span class="text-faint hide-mobile">${event.path}</span>` : ''}
                                </div>
                              </div>
                              <div class="activity-time text-right hide-mobile">
                                <div class="text-caption text-muted">${formatRelativeTime(event.created_at)}</div>
                                <div class="text-caption text-faint">${formatTimestamp(event.created_at)}</div>
                              </div>
                            </div>
                          `
                        }).join('')}
                      </div>
                    `}
                  </div>
                  ${filteredEvents.length > 0 ? `
                    <div class="card-footer flex items-center justify-between" style="border-radius: 0">
                      <span class="text-caption text-muted">${filteredEvents.length} event${filteredEvents.length === 1 ? '' : 's'}</span>
                    </div>
                  ` : ''}
                </div>
              </div>
            </div>

          </div>
        </div>
      </div>
    `

    // Re-render Lucide icons
    if (window.lucide) window.lucide.createIcons()

    // Setup collapse handlers
    container.querySelectorAll('.collapse-toggle').forEach(toggle => {
      toggle.addEventListener('click', () => {
        const header = toggle.closest('.panel-group-header')
        const group = header.dataset.group
        const panelGroup = header.closest('.panel-group')
        const isCollapsed = panelGroup.classList.toggle('collapsed')
        setUIState(`events.${group}.collapsed`, isCollapsed)
      })
    })

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
