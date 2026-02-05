/**
 * Lucide icon helper
 *
 * Injects SVG icons inside <i data-lucide> elements without replacing them.
 * This keeps Vue's DOM references intact — lucide.createIcons() replaces <i>
 * with <svg>, which breaks Vue's virtual DOM patcher (insertBefore null).
 */
import { nextTick } from 'vue'

let pending = false

export function refreshIcons() {
  if (pending) return
  pending = true
  nextTick(() => {
    pending = false
    processIcons()
  })
}

function processIcons() {
  if (!window.lucide?.createElement) return

  document.querySelectorAll('i[data-lucide]').forEach(el => {
    // Skip already-rendered icons
    if (el.querySelector('svg')) return

    const name = el.getAttribute('data-lucide')
    if (!name) return

    const svg = createIconSvg(name)
    if (!svg) return

    // Inject SVG inside <i> — Vue still owns the <i> element
    el.style.display = 'inline-flex'
    el.style.alignItems = 'center'
    el.style.justifyContent = 'center'
    el.appendChild(svg)
  })
}

function createIconSvg(name) {
  // Convert kebab-case to PascalCase for lucide lookup
  const pascal = name.replace(/(^|-)([a-z])/g, (_, __, c) => c.toUpperCase())
  const iconNode = window.lucide[pascal]

  // Verify it's an icon node (array), not a function like createElement
  if (!iconNode || !Array.isArray(iconNode)) return null

  try {
    const svg = window.lucide.createElement(iconNode)
    // Remove fixed dimensions — let the container (w-4 h-4 etc) control size
    svg.removeAttribute('width')
    svg.removeAttribute('height')
    svg.style.width = '100%'
    svg.style.height = '100%'
    return svg
  } catch {
    return null
  }
}
