/**
 * PanelToolbar Component
 * Search input + filter/action buttons for panels
 */

/**
 * Toolbar configuration
 * @typedef {Object} ToolbarConfig
 * @property {string} [searchId] - ID for search input
 * @property {string} [searchValue=''] - Current search value
 * @property {string} [searchPlaceholder='Filter...'] - Search placeholder
 * @property {Array<ToolbarButton>} [buttons] - Action buttons
 */

/**
 * @typedef {Object} ToolbarButton
 * @property {string} id - Button ID
 * @property {string} icon - Lucide icon name
 * @property {string} [title] - Button title/tooltip
 * @property {boolean} [primary] - Use primary style
 * @property {boolean} [hideOnMobile] - Hide on mobile
 */

/**
 * Render toolbar HTML
 * @param {ToolbarConfig} config
 * @returns {string} HTML string
 */
export function renderToolbar(config) {
  const {
    searchId = 'filter-input',
    searchValue = '',
    searchPlaceholder = 'Filter...',
    buttons = []
  } = config

  const buttonsHtml = buttons.map(btn => {
    const classes = [
      'btn btn-sm',
      btn.primary ? 'btn-primary' : 'btn-secondary',
      btn.hideOnMobile ? 'hide-mobile' : ''
    ].filter(Boolean).join(' ')

    return `
      <button id="${btn.id}" class="${classes} toolbar-btn" title="${btn.title || ''}">
        <i data-lucide="${btn.icon}" class="w-4 h-4"></i>
      </button>
    `
  }).join('')

  return `
    <div class="input toolbar-search">
      <i data-lucide="search" class="w-4 h-4 text-faint"></i>
      <input type="text" id="${searchId}" placeholder="${searchPlaceholder}" value="">
    </div>
    ${buttons.length > 0 ? `
      <div class="flex items-center gap-2" style="flex-shrink: 0">
        ${buttonsHtml}
      </div>
    ` : ''}
  `
}

/**
 * Setup toolbar event handlers
 * @param {HTMLElement} container
 * @param {Object} callbacks
 * @param {Function} [callbacks.onSearch] - Called on search input (value) => void
 * @param {Function} [callbacks.onClear] - Called when search cleared () => void
 * @param {Object} [callbacks.buttons] - Button click handlers { buttonId: () => void }
 * @param {string} [searchId='filter-input']
 */
export function setupToolbar(container, callbacks = {}, searchId = 'filter-input') {
  const input = container.querySelector(`#${searchId}`)

  if (input) {
    input.addEventListener('input', (e) => {
      callbacks.onSearch?.(e.target.value)
    })

    input.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') {
        input.value = ''
        callbacks.onSearch?.('')
        callbacks.onClear?.()
      }
    })
  }

  // Setup button handlers
  if (callbacks.buttons) {
    Object.entries(callbacks.buttons).forEach(([id, handler]) => {
      container.querySelector(`#${id}`)?.addEventListener('click', handler)
    })
  }
}
