/**
 * Fazt Admin - Main Entry Point
 */

import { createRouter, createCommands, createAgentContext } from '../packages/zap/index.js'
import { createClient, mockAdapter } from '../packages/fazt-sdk/index.js'
import { routes } from './routes.js'
import { initTheme, sidebarCollapsed, toggleSidebar, theme, palette, setTheme, setPalette } from './stores/app.js'
import { loadApps, loadAliases, loadHealth, loadStats } from './stores/data.js'

// Pages
import * as dashboardPage from './pages/dashboard.js'
import * as appsPage from './pages/apps.js'
import * as aliasesPage from './pages/aliases.js'
import * as systemPage from './pages/system.js'
import * as settingsPage from './pages/settings.js'

// Page map
const pages = {
  dashboard: dashboardPage,
  apps: appsPage,
  'app-detail': appsPage, // Reuse apps page for now
  aliases: aliasesPage,
  system: systemPage,
  settings: settingsPage
}

// Use mock adapter for development, real API in production
const useMock = new URLSearchParams(window.location.search).get('mock') === 'true'
const client = createClient(useMock ? { adapter: mockAdapter } : {})

// Create router
const router = createRouter({
  mode: 'hash',
  routes,
  onNavigate: (match) => {
    renderPage(match)
    updateBreadcrumb(match)
    updateNav(match)
  }
})

// Create command palette
const commands = createCommands()

// Register commands
commands.registerAll([
  { id: 'dashboard', title: 'Go to Dashboard', shortcut: 'G H', icon: 'layout-grid', action: () => router.push('/') },
  { id: 'apps', title: 'Go to Apps', shortcut: 'G A', icon: 'layers', action: () => router.push('/apps') },
  { id: 'aliases', title: 'Go to Aliases', shortcut: 'G L', icon: 'link', action: () => router.push('/aliases') },
  { id: 'system', title: 'Go to System', shortcut: 'G S', icon: 'heart-pulse', action: () => router.push('/system') },
  { id: 'settings', title: 'Go to Settings', shortcut: 'G ,', icon: 'settings', action: () => router.push('/settings') },
  { id: 'theme-light', title: 'Switch to Light Theme', icon: 'sun', action: () => setTheme('light') },
  { id: 'theme-dark', title: 'Switch to Dark Theme', icon: 'moon', action: () => setTheme('dark') },
  { id: 'refresh', title: 'Refresh Data', shortcut: 'R', icon: 'refresh-cw', action: loadData }
])

// Create agent context
createAgentContext(router, commands, {
  theme,
  palette,
  sidebarCollapsed
})

// DOM elements
let pageContainer
let breadcrumbContainer
let navItems
let currentCleanup = null

/**
 * Initialize app
 */
function init() {
  // Initialize theme
  initTheme()

  // Get DOM elements
  pageContainer = document.getElementById('page-content')
  breadcrumbContainer = document.getElementById('breadcrumb')

  // Setup sidebar toggle
  document.getElementById('sidebar-toggle')?.addEventListener('click', () => {
    toggleSidebar()
    document.getElementById('sidebar')?.classList.toggle('collapsed', sidebarCollapsed.get())
  })

  // Setup nav items
  setupNav()

  // Setup command palette UI
  setupCommandPalette()

  // Setup theme/palette pickers
  setupThemePicker()

  // Load initial data
  loadData()

  // Render current route
  const match = router.current.get()
  if (match) {
    renderPage(match)
    updateBreadcrumb(match)
    updateNav(match)
  }

  console.log('[Fazt Admin] Initialized', useMock ? '(mock mode)' : '')
}

/**
 * Load data from API
 */
async function loadData() {
  await Promise.all([
    loadApps(client),
    loadAliases(client),
    loadHealth(client),
    loadStats(client)
  ])
}

/**
 * Render page
 * @param {import('../packages/zap/router.js').RouteMatch} match
 */
function renderPage(match) {
  if (!pageContainer) return

  // Cleanup previous page
  if (currentCleanup) {
    currentCleanup()
    currentCleanup = null
  }

  const page = pages[match.route.name]
  if (page?.render) {
    currentCleanup = page.render(pageContainer, {
      router,
      client,
      commands,
      params: match.params,
      refresh: loadData
    })
  } else {
    pageContainer.innerHTML = `
      <div class="flex items-center justify-center h-full">
        <div class="text-center">
          <i data-lucide="alert-circle" class="w-12 h-12 text-faint mx-auto mb-2"></i>
          <div class="text-heading text-primary mb-1">Page Not Found</div>
          <div class="text-caption text-muted">The page you're looking for doesn't exist.</div>
          <button class="btn btn-primary mt-4" onclick="window.location.hash='/'">
            Go to Dashboard
          </button>
        </div>
      </div>
    `
    if (window.lucide) window.lucide.createIcons()
  }
}

/**
 * Update breadcrumb
 * @param {import('../packages/zap/router.js').RouteMatch} match
 */
function updateBreadcrumb(match) {
  if (!breadcrumbContainer) return

  const route = match.route
  const parent = route.meta?.parent ? routes.find(r => r.name === route.meta.parent) : null

  breadcrumbContainer.innerHTML = `
    <span class="text-muted">Home</span>
    <i data-lucide="chevron-right" class="w-3 h-3 text-faint"></i>
    ${parent ? `
      <a href="#${parent.path}" class="text-muted hover:underline">${parent.meta?.title || parent.name}</a>
      <i data-lucide="chevron-right" class="w-3 h-3 text-faint"></i>
    ` : ''}
    <span class="text-primary">${route.meta?.title || route.name}</span>
  `

  if (window.lucide) window.lucide.createIcons()
}

/**
 * Update active nav item
 * @param {import('../packages/zap/router.js').RouteMatch} match
 */
function updateNav(match) {
  document.querySelectorAll('.nav-item[data-route]').forEach(item => {
    const isActive = item.dataset.route === match.route.name ||
      (match.route.meta?.parent && item.dataset.route === match.route.meta.parent)
    item.classList.toggle('active', isActive)
  })
}

/**
 * Setup navigation click handlers
 */
function setupNav() {
  document.querySelectorAll('.nav-item[data-route]').forEach(item => {
    item.addEventListener('click', (e) => {
      e.preventDefault()
      const route = routes.find(r => r.name === item.dataset.route)
      if (route) {
        router.push(route.path)
      }
    })
  })

  // Menu toggles
  document.querySelectorAll('.menu-toggle').forEach(toggle => {
    toggle.addEventListener('click', () => {
      toggle.classList.toggle('open')
      toggle.parentElement?.nextElementSibling?.classList.toggle('open')
    })
  })
}

/**
 * Setup command palette UI
 */
function setupCommandPalette() {
  const paletteEl = document.getElementById('command-palette')
  const backdropEl = document.getElementById('command-backdrop')
  const inputEl = document.getElementById('command-input')
  const resultsEl = document.getElementById('command-results')

  if (!paletteEl || !inputEl || !resultsEl) return

  // Open/close handlers
  commands.isOpen.subscribe(isOpen => {
    paletteEl.classList.toggle('hidden', !isOpen)
    backdropEl?.classList.toggle('hidden', !isOpen)
    if (isOpen) {
      inputEl.focus()
      inputEl.value = ''
      renderCommandResults()
    }
  })

  // Close on backdrop click
  backdropEl?.addEventListener('click', () => commands.close())

  // Input handler
  inputEl.addEventListener('input', (e) => {
    commands.query.set(e.target.value)
    renderCommandResults()
  })

  // Keyboard navigation
  inputEl.addEventListener('keydown', (e) => {
    commands.handleKeyDown(e)
  })

  function renderCommandResults() {
    const cmds = commands.filteredCommands.get()
    const selected = commands.selectedIndex.get()

    resultsEl.innerHTML = cmds.map((cmd, i) => `
      <div class="dropdown-item ${i === selected ? 'active' : ''}" data-command="${cmd.id}" style="${i === selected ? 'background: var(--bg-2)' : ''}">
        <i data-lucide="${cmd.icon || 'terminal'}" class="w-4 h-4 text-muted"></i>
        <span class="text-label text-primary flex-1">${cmd.title}</span>
        ${cmd.shortcut ? `<kbd class="kbd">${cmd.shortcut}</kbd>` : ''}
      </div>
    `).join('')

    if (window.lucide) window.lucide.createIcons()

    // Click handlers
    resultsEl.querySelectorAll('[data-command]').forEach(item => {
      item.addEventListener('click', () => {
        commands.execute(item.dataset.command)
      })
    })
  }
}

/**
 * Setup theme/palette pickers
 */
function setupThemePicker() {
  // Theme buttons
  document.querySelectorAll('.theme-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      setTheme(btn.dataset.theme)
      document.querySelectorAll('.theme-btn').forEach(b => b.classList.remove('active'))
      btn.classList.add('active')
    })
  })

  // Palette swatches
  document.querySelectorAll('.swatch[data-palette]').forEach(swatch => {
    swatch.addEventListener('click', () => {
      setPalette(swatch.dataset.palette)
      document.querySelectorAll('.swatch').forEach(s => s.classList.remove('active'))
      swatch.classList.add('active')
    })
  })
}

// Initialize on DOM ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init)
} else {
  init()
}

// Export for debugging
window.__fazt_admin = {
  router,
  commands,
  client,
  loadData
}
