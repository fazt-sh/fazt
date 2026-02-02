/**
 * Table Component
 * Responsive table with configurable columns
 */

/**
 * Column configuration
 * @typedef {Object} ColumnConfig
 * @property {string} key - Data key
 * @property {string} label - Column header label
 * @property {boolean} [hideOnMobile] - Hide column on mobile
 * @property {boolean} [showOnMobile] - Only show on mobile (secondary info)
 * @property {Function} [render] - Custom render function (value, row) => HTML
 * @property {string} [align='left'] - Text alignment
 * @property {string} [width] - Column width (e.g., '120px', '20%')
 */

/**
 * Table configuration
 * @typedef {Object} TableConfig
 * @property {Array<ColumnConfig>} columns
 * @property {Array<Object>} data
 * @property {string} [rowKey='id'] - Key for row identifier
 * @property {string} [rowDataAttr] - Data attribute name for row clicks
 * @property {boolean} [clickable=false] - Make rows clickable
 * @property {string} [emptyIcon='inbox'] - Icon when no data
 * @property {string} [emptyTitle='No data'] - Title when no data
 * @property {string} [emptyMessage=''] - Message when no data
 */

/**
 * Render table HTML
 * @param {TableConfig} config
 * @returns {string} HTML string
 */
export function renderTable(config) {
  const {
    columns,
    data,
    rowKey = 'id',
    rowDataAttr,
    clickable = false,
    emptyIcon = 'inbox',
    emptyTitle = 'No data',
    emptyMessage = ''
  } = config

  if (data.length === 0) {
    return `
      <div class="flex flex-col items-center justify-center p-8 text-center" style="min-height: 200px">
        <div class="icon-box mb-3" style="width:48px;height:48px;opacity:0.5">
          <i data-lucide="${emptyIcon}" class="w-6 h-6"></i>
        </div>
        <div class="text-heading text-primary mb-1">${emptyTitle}</div>
        ${emptyMessage ? `<div class="text-caption text-muted">${emptyMessage}</div>` : ''}
      </div>
    `
  }

  const visibleColumns = columns.filter(col => !col.showOnMobile)

  return `
    <div class="table-container" style="overflow-x: visible">
      <table style="width: 100%; min-width: 0">
        <thead class="sticky" style="top: 0; background: var(--bg-1)">
          <tr style="border-bottom: 1px solid var(--border-subtle)">
            ${visibleColumns.map(col => `
              <th class="px-3 py-2 text-left text-micro text-muted ${col.hideOnMobile ? 'hide-mobile' : ''}" ${col.width ? `style="width: ${col.width}"` : ''}>
                ${col.label}
              </th>
            `).join('')}
          </tr>
        </thead>
        <tbody>
          ${data.map(row => {
            const rowId = row[rowKey]
            const dataAttr = rowDataAttr ? `data-${rowDataAttr}="${rowId}"` : ''
            const rowClass = clickable ? 'row row-clickable' : 'row'

            return `
              <tr class="${rowClass}" ${dataAttr} style="border-bottom: 1px solid var(--border-subtle)">
                ${visibleColumns.map((col, colIndex) => {
                  const value = row[col.key]
                  const cellContent = col.render ? col.render(value, row, colIndex) : escapeHtml(String(value ?? ''))
                  const mobileClass = col.hideOnMobile ? 'hide-mobile' : ''

                  return `
                    <td class="px-3 py-2 ${mobileClass}">
                      ${cellContent}
                    </td>
                  `
                }).join('')}
              </tr>
            `
          }).join('')}
        </tbody>
      </table>
    </div>
  `
}

/**
 * Setup table row click handlers
 * @param {HTMLElement} container
 * @param {string} dataAttr - Data attribute name (without 'data-' prefix)
 * @param {Function} onClick - Click handler (id, row) => void
 */
export function setupTableClicks(container, dataAttr, onClick) {
  container.querySelectorAll(`[data-${dataAttr}]`).forEach(row => {
    row.addEventListener('click', () => {
      onClick(row.dataset[toCamelCase(dataAttr)])
    })
  })
}

/**
 * Render a footer with count
 * @param {number} count
 * @param {string} singular
 * @param {string} [plural]
 */
export function renderTableFooter(count, singular, plural) {
  const label = count === 1 ? singular : (plural || `${singular}s`)
  return `
    <div class="card-footer flex items-center justify-between" style="border-radius: 0">
      <span class="text-caption text-muted">${count} ${label}</span>
    </div>
  `
}

// Helpers

function escapeHtml(str) {
  const div = document.createElement('div')
  div.textContent = str
  return div.innerHTML
}

function toCamelCase(str) {
  return str.replace(/-([a-z])/g, (g) => g[1].toUpperCase())
}

// Common cell renderers

/**
 * Render a cell with icon and text
 */
export function renderIconCell(icon, primary, secondary) {
  return `
    <div class="flex items-center gap-2">
      <div class="icon-box icon-box-sm">
        <i data-lucide="${icon}" class="w-3.5 h-3.5"></i>
      </div>
      <div>
        <div class="text-label text-primary">${escapeHtml(primary)}</div>
        ${secondary ? `<div class="text-caption mono text-faint show-mobile">${escapeHtml(secondary)}</div>` : ''}
      </div>
    </div>
  `
}

/**
 * Render a status cell with dot
 */
export function renderStatusCell(label, isActive = true) {
  const colorClass = isActive ? 'text-success' : 'text-muted'
  const dotClass = isActive ? 'status-dot-success pulse' : ''

  return `
    <span class="flex items-center gap-1 text-caption ${colorClass}">
      <span class="status-dot ${dotClass}"></span>
      <span class="hide-mobile">${escapeHtml(label)}</span>
    </span>
  `
}

/**
 * Render a mono text cell
 */
export function renderMonoCell(text, maxWidth) {
  const style = maxWidth ? `max-width: ${maxWidth}; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;` : ''
  return `<span class="text-caption mono text-muted" style="${style}">${escapeHtml(text)}</span>`
}
