/**
 * Fazt Admin - Main Entry Point
 */

import { createRouter, createCommands, createAgentContext } from '../packages/zap/index.js'
import { createClient, mockAdapter } from '../packages/fazt-sdk/index.js'
import { routes } from './routes.js'
import { initTheme, sidebarCollapsed, toggleSidebar, theme, palette, setTheme, setPalette } from './stores/app.js'
import { loadApps, loadAliases, loadHealth, loadAuth, auth, signOut } from './stores/data.js'

// Pages
import * as dashboardPage from './pages/dashboard.js'
import * as appsPage from './pages/apps.js'
import * as aliasesPage from './pages/aliases.js'
import * as systemPage from './pages/system.js'
import * as settingsPage from './pages/settings.js'
import * as designSystemPage from './pages/design-system.js'

// Page map
const pages = {
  dashboard: dashboardPage,
  apps: appsPage,
  'app-detail': appsPage, // Reuse apps page for now
  aliases: aliasesPage,
  system: systemPage,
  settings: settingsPage,
  'design-system': designSystemPage
}

// Use mock adapter for development, real API in production
const useMock = new URLSearchParams(window.location.search).get('mock') === 'true'

// Compute API base URL - admin UI may be served from different subdomain than API
// API always lives on admin.* subdomain
function getApiBaseUrl() {
  const hostname = window.location.hostname
  const port = window.location.port ? ':' + window.location.port : ''
  const protocol = window.location.protocol

  // Replace any subdomain with 'admin' to get API URL
  const parts = hostname.split('.')
  if (parts.length >= 2) {
    parts[0] = 'admin'
    return `${protocol}//${parts.join('.')}${port}`
  }
  // Fallback: same origin
  return ''
}

const apiBaseUrl = useMock ? '' : getApiBaseUrl()
console.log('[Admin] Mode:', useMock ? 'MOCK' : 'REAL')
console.log('[Admin] API Base:', apiBaseUrl || '(same origin)')

const client = createClient(useMock ? { adapter: mockAdapter } : { baseUrl: apiBaseUrl })

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
 * Show unauthorized access modal
 */
function showUnauthorizedModal() {
  const authState = auth.get()
  const userName = authState.user?.name || authState.user?.email || 'User'
  const userRole = authState.user?.role || 'user'

  // Create modal overlay
  const modalHTML = `
    <div id="unauthorized-modal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
      <div class="bg-[var(--bg-1)] border border-[var(--border)] rounded-lg shadow-2xl max-w-md w-full mx-4 p-6">
        <div class="flex items-start gap-4 mb-4">
          <div class="flex-shrink-0 w-12 h-12 rounded-full bg-[var(--error-soft)] flex items-center justify-center">
            <svg class="w-6 h-6 text-[var(--error)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <div class="flex-1">
            <h3 class="text-lg font-semibold text-[var(--text-1)] mb-1">Access Denied</h3>
            <p class="text-sm text-[var(--text-2)] mb-3">
              Sorry, <strong>${userName}</strong> (<code>${userRole}</code> role), you don't have permission to access the admin dashboard.
            </p>
            <p class="text-sm text-[var(--text-3)]">
              Admin access requires <strong>admin</strong> or <strong>owner</strong> role.
            </p>
          </div>
        </div>

        <div class="flex gap-3 mt-6">
          <button
            id="unauthorized-signout"
            class="flex-1 px-4 py-2 bg-[var(--accent)] hover:bg-[var(--accent-hover)] text-white font-medium rounded-md transition-colors"
          >
            Sign Out & Try Again
          </button>
        </div>
      </div>
    </div>
  `

  // Insert modal into DOM
  document.body.insertAdjacentHTML('beforeend', modalHTML)

  // Setup sign out button
  document.getElementById('unauthorized-signout')?.addEventListener('click', async () => {
    try {
      await signOut(client)
      // signOut handles redirect to login
    } catch (err) {
      console.error('[Unauthorized] Sign out failed:', err)
      // Still redirect to login even if signout fails - include redirect parameter
      const currentUrl = window.location.href
      const redirectParam = encodeURIComponent(currentUrl)

      const parts = window.location.hostname.split('.')
      const rootDomain = parts.length > 2 && parts[0] !== 'www'
        ? parts.slice(1).join('.')
        : window.location.hostname

      window.location.href = `${window.location.protocol}//${rootDomain}${window.location.port ? ':' + window.location.port : ''}/auth/dev/login?redirect=${redirectParam}`
    }
  })
}

/**
 * Initialize app
 */
function init() {
  console.log('[init] Starting initialization')
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
  console.log('[loadData] STARTING')
  // Load auth first to check permissions
  await loadAuth(client)
  console.log('[loadData] Auth loaded')

  // Check if user has required role
  const authState = auth.get()
  console.log('[loadData] Auth state:', authState)

  if (authState.authenticated && authState.user) {
    const user = authState.user
    console.log('[loadData] User role:', user.role)

    if (user.role !== 'owner' && user.role !== 'admin') {
      // User authenticated but lacks required role
      console.log('[loadData] UNAUTHORIZED - showing modal')
      auth.setKey('unauthorized', true)
      showUnauthorizedModal()
      return // Don't load other data
    }

    console.log('[loadData] User authorized, loading data')
  }

  // User has required role, load data
  await Promise.all([
    loadApps(client),
    loadAliases(client),
    loadHealth(client)
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
