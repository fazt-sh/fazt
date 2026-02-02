/**
 * Panel Component
 * Collapsible panel with header, toolbar area, and content
 */

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
 * Panel configuration
 * @typedef {Object} PanelConfig
 * @property {string} id - Unique panel identifier for state persistence
 * @property {string} title - Panel title
 * @property {number} [count] - Optional count badge
 * @property {string} [toolbar] - HTML for toolbar content (search, filters, buttons)
 * @property {string} content - HTML for panel body content (scrollable)
 * @property {string} [footer] - HTML for footer (fixed, outside scroll area)
 * @property {number} [minHeight=400] - Minimum height in px
 * @property {number} [maxHeight=600] - Maximum height in px
 */

/**
 * Render a collapsible panel
 * @param {PanelConfig} config
 * @returns {string} HTML string
 */
export function renderPanel(config) {
  const {
    id,
    title,
    count,
    toolbar = '',
    content,
    footer = '',
    minHeight = 400,
    maxHeight = 600
  } = config

  const isCollapsed = getUIState(`panel.${id}.collapsed`, false)

  return `
    <div class="panel-group ${isCollapsed ? 'collapsed' : ''}" data-panel-id="${id}">
      <div class="panel-group-card card">
        <header class="panel-group-header" data-panel="${id}">
          <button class="collapse-toggle">
            <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
            <span class="text-heading text-primary">${title}</span>
            ${count !== undefined ? `
              <span class="text-caption mono px-1.5 py-0.5 ml-2 badge-muted" style="border-radius: var(--radius-sm)">${count}</span>
            ` : ''}
          </button>
        </header>
        <div class="panel-group-body" style="padding: 0">
          <div class="panel-inner" style="border: none; border-radius: 0">
            ${toolbar ? `
              <div class="card-header flex items-center justify-between">
                ${toolbar}
              </div>
            ` : ''}
            <div class="panel-scroll-area" style="min-height: ${minHeight}px; max-height: ${maxHeight}px; overflow: auto">
              ${content}
            </div>
            ${footer}
          </div>
        </div>
      </div>
    </div>
  `
}

/**
 * Setup panel event handlers
 * @param {HTMLElement} container - Container element
 * @param {Object} [callbacks] - Optional callbacks
 * @param {Function} [callbacks.onCollapse] - Called when panel collapses/expands
 */
export function setupPanel(container, callbacks = {}) {
  container.querySelectorAll('.collapse-toggle').forEach(toggle => {
    toggle.addEventListener('click', () => {
      const header = toggle.closest('.panel-group-header')
      const panelId = header.dataset.panel
      const panelGroup = header.closest('.panel-group')
      const isCollapsed = panelGroup.classList.toggle('collapsed')
      setUIState(`panel.${panelId}.collapsed`, isCollapsed)
      callbacks.onCollapse?.(panelId, isCollapsed)
    })
  })

  // Re-render icons
  if (window.lucide) window.lucide.createIcons()
}

export { getUIState, setUIState }
