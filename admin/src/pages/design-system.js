/**
 * Design System Test Page
 * Fresh slate implementation of panel-based layout
 */

/**
 * Get UI state from localStorage (web-specific key)
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
 * Set UI state to localStorage (web-specific key)
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
 * Render design system test page
 */
export function render(container, ctx) {
  const { router } = ctx

  // Get collapse states
  const statsCollapsed = getUIState('ds.stats.collapsed', false)
  const metricsCollapsed = getUIState('ds.metrics.collapsed', false)
  const activityCollapsed = getUIState('ds.activity.collapsed', false)

  container.innerHTML = `
    <div class="design-system-page">
      <!-- Content Container (fixed-width, centered) -->
      <div class="content-container">
        <div class="content-scroll">

          <!-- Page Header -->
          <div class="page-header">
            <div>
              <h1 class="text-title text-primary">Design System Test</h1>
              <p class="text-caption text-muted">Panel-based layout with collapsible sections</p>
            </div>
            <button class="btn btn-primary" onclick="window.location.hash='/'">
              <i data-lucide="arrow-left" class="w-4 h-4"></i>
              Back to Dashboard
            </button>
          </div>

          <!-- Panel Group: Stats Cards -->
          <div class="panel-group ${statsCollapsed ? 'collapsed' : ''}">
            <div class="panel-group-card card">
              <header class="panel-group-header" data-group="stats">
                <button class="collapse-toggle">
                  <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                  <span class="text-heading text-primary">Overview Stats</span>
                  <span class="text-caption text-faint ml-auto">5 metrics</span>
                </button>
              </header>
              <div class="panel-group-body">
              <div class="panel-grid grid-5">
                ${renderStatCard('Apps', '12', 'layers', 'success', '+3 this week')}
                ${renderStatCard('Requests', '45.2K', 'activity', 'primary', '24h total')}
                ${renderStatCard('Storage', '2.4 GB', 'database', 'primary', '68% used')}
                ${renderStatCard('Uptime', '99.98%', 'clock', 'success', '30d average')}
                ${renderStatCard('Status', 'Healthy', 'heart-pulse', 'success', 'v0.17.0')}
              </div>
              </div>
            </div>
          </div>

          <!-- Panel Group: Metrics Grid -->
          <div class="panel-group ${metricsCollapsed ? 'collapsed' : ''}">
            <div class="panel-group-card card">
              <header class="panel-group-header" data-group="metrics">
                <button class="collapse-toggle">
                  <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                  <span class="text-heading text-primary">Detailed Metrics</span>
                </button>
              </header>
              <div class="panel-group-body">
              <div class="panel-grid grid-3">
                ${renderMetricPanel('Response Time', '124ms', 'avg', 'zap')}
                ${renderMetricPanel('Error Rate', '0.02%', 'errors/total', 'alert-triangle')}
                ${renderMetricPanel('Throughput', '1.2K/min', 'requests', 'trending-up')}
              </div>
              </div>
            </div>
          </div>

          <!-- Panel Group: Activity Feed -->
          <div class="panel-group ${activityCollapsed ? 'collapsed' : ''}">
            <div class="panel-group-card card">
              <header class="panel-group-header" data-group="activity">
                <button class="collapse-toggle">
                  <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                  <span class="text-heading text-primary">Recent Activity</span>
                  <span class="text-caption text-accent ml-auto">View all</span>
                </button>
              </header>
              <div class="panel-group-body">
                <div class="activity-list">
                  ${renderActivityItem('check-circle', 'App deployed successfully', 'momentum', '2h ago', 'success')}
                  ${renderActivityItem('settings', 'Configuration updated', 'SSL certificates', '5h ago', 'primary')}
                  ${renderActivityItem('alert-triangle', 'High memory usage detected', 'nexus app', '1d ago', 'warning')}
                  ${renderActivityItem('upload', 'New version released', 'v0.17.0', '2d ago', 'primary')}
                  ${renderActivityItem('key', 'API token generated', 'Admin user', '3d ago', 'primary')}
                  ${renderActivityItem('shield-check', 'Security scan passed', 'All apps', '4d ago', 'success')}
                </div>
              </div>
            </div>
          </div>

          <!-- Extraneous Panel (like Twitter sidebar) -->
          <aside class="side-panel">
            <div class="card p-4">
              <div class="text-heading text-primary mb-3">Quick Links</div>
              <div class="space-y-2">
                <a href="#" class="flex items-center gap-2 text-label text-secondary hover:text-primary">
                  <i data-lucide="book-open" class="w-4 h-4"></i>
                  Documentation
                </a>
                <a href="#" class="flex items-center gap-2 text-label text-secondary hover:text-primary">
                  <i data-lucide="code" class="w-4 h-4"></i>
                  API Reference
                </a>
                <a href="#" class="flex items-center gap-2 text-label text-secondary hover:text-primary">
                  <i data-lucide="github" class="w-4 h-4"></i>
                  GitHub
                </a>
              </div>
            </div>
          </aside>

        </div>
      </div>
    </div>
  `

  // Render icons
  if (window.lucide) window.lucide.createIcons()

  // Setup collapse handlers
  setupCollapseHandlers(container)

  // Cleanup
  return () => {}
}

/**
 * Render a stat card
 */
function renderStatCard(label, value, icon, color = 'primary', subtitle = '') {
  return `
    <div class="stat-card card">
      <div class="stat-card-header">
        <span class="text-micro text-muted">${label}</span>
        <i data-lucide="${icon}" class="w-4 h-4 text-${color}"></i>
      </div>
      <div class="stat-card-value text-display mono text-primary">${value}</div>
      ${subtitle ? `<div class="stat-card-subtitle text-caption text-muted">${subtitle}</div>` : ''}
    </div>
  `
}

/**
 * Render a metric panel
 */
function renderMetricPanel(title, value, label, icon) {
  return `
    <div class="metric-panel card p-4">
      <div class="flex items-center justify-between mb-3">
        <span class="text-label text-primary">${title}</span>
        <i data-lucide="${icon}" class="w-5 h-5 text-accent"></i>
      </div>
      <div class="text-display text-primary mb-1">${value}</div>
      <div class="text-caption text-muted">${label}</div>
    </div>
  `
}

/**
 * Render an activity item
 */
function renderActivityItem(icon, title, subtitle, time, color = 'primary') {
  return `
    <div class="activity-item">
      <div class="activity-icon icon-box-sm">
        <i data-lucide="${icon}" class="w-4 h-4"></i>
      </div>
      <div class="activity-content">
        <div class="text-label text-primary">${title}</div>
        <div class="text-caption text-muted">${subtitle}</div>
      </div>
      <div class="activity-time text-caption text-faint">${time}</div>
    </div>
  `
}

/**
 * Setup collapse toggle handlers
 */
function setupCollapseHandlers(container) {
  container.querySelectorAll('.collapse-toggle').forEach(toggle => {
    toggle.addEventListener('click', () => {
      const header = toggle.closest('.panel-group-header')
      const group = header.dataset.group
      const section = header.parentElement
      const isCollapsed = section.classList.contains('collapsed')

      if (isCollapsed) {
        section.classList.remove('collapsed')
        setUIState(`ds.${group}.collapsed`, false)
      } else {
        section.classList.add('collapsed')
        setUIState(`ds.${group}.collapsed`, true)
      }
    })
  })
}
