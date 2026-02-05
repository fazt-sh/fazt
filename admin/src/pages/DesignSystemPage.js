import { ref, onMounted, onUpdated } from 'vue'
import { useUIStore } from '../stores/ui.js'
import { useRouter } from 'vue-router'
import { refreshIcons } from '../lib/icons.js'

export default {
  name: 'DesignSystemPage',
  setup() {
    const uiStore = useUIStore()
    const router = useRouter()

    const statsCollapsed = ref(uiStore.getUIState('ds.stats.collapsed', false))
    const metricsCollapsed = ref(uiStore.getUIState('ds.metrics.collapsed', false))
    const activityCollapsed = ref(uiStore.getUIState('ds.activity.collapsed', false))

    const toggleStats = () => {
      statsCollapsed.value = !statsCollapsed.value
      uiStore.setUIState('ds.stats.collapsed', statsCollapsed.value)
    }

    const toggleMetrics = () => {
      metricsCollapsed.value = !metricsCollapsed.value
      uiStore.setUIState('ds.metrics.collapsed', metricsCollapsed.value)
    }

    const toggleActivity = () => {
      activityCollapsed.value = !activityCollapsed.value
      uiStore.setUIState('ds.activity.collapsed', activityCollapsed.value)
    }

    const goBack = () => {
      router.push('/')
    }

    onMounted(() => {
      refreshIcons()
    })

    onUpdated(() => {
      refreshIcons()
    })

    return {
      statsCollapsed,
      metricsCollapsed,
      activityCollapsed,
      toggleStats,
      toggleMetrics,
      toggleActivity,
      goBack
    }
  },
  template: `
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
            <button class="btn btn-primary" @click="goBack">
              <i data-lucide="arrow-left" class="w-4 h-4"></i>
              Back to Dashboard
            </button>
          </div>

          <!-- Panel Group: Stats Cards -->
          <div class="panel-group" :class="{ collapsed: statsCollapsed }">
            <div class="panel-group-card card">
              <header class="panel-group-header" data-group="stats">
                <button class="collapse-toggle" @click="toggleStats">
                  <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                  <span class="text-heading text-primary">Overview Stats</span>
                  <span class="text-caption text-faint ml-auto">5 metrics</span>
                </button>
              </header>
              <div class="panel-group-body">
                <div class="panel-grid grid-5">
                  <div class="stat-card card">
                    <div class="stat-card-header">
                      <span class="text-micro text-muted">Apps</span>
                      <i data-lucide="layers" class="w-4 h-4 text-success"></i>
                    </div>
                    <div class="stat-card-value text-display mono text-primary">12</div>
                    <div class="stat-card-subtitle text-caption text-muted">+3 this week</div>
                  </div>
                  <div class="stat-card card">
                    <div class="stat-card-header">
                      <span class="text-micro text-muted">Requests</span>
                      <i data-lucide="activity" class="w-4 h-4 text-primary"></i>
                    </div>
                    <div class="stat-card-value text-display mono text-primary">45.2K</div>
                    <div class="stat-card-subtitle text-caption text-muted">24h total</div>
                  </div>
                  <div class="stat-card card">
                    <div class="stat-card-header">
                      <span class="text-micro text-muted">Storage</span>
                      <i data-lucide="database" class="w-4 h-4 text-primary"></i>
                    </div>
                    <div class="stat-card-value text-display mono text-primary">2.4 GB</div>
                    <div class="stat-card-subtitle text-caption text-muted">68% used</div>
                  </div>
                  <div class="stat-card card">
                    <div class="stat-card-header">
                      <span class="text-micro text-muted">Uptime</span>
                      <i data-lucide="clock" class="w-4 h-4 text-success"></i>
                    </div>
                    <div class="stat-card-value text-display mono text-primary">99.98%</div>
                    <div class="stat-card-subtitle text-caption text-muted">30d average</div>
                  </div>
                  <div class="stat-card card">
                    <div class="stat-card-header">
                      <span class="text-micro text-muted">Status</span>
                      <i data-lucide="heart-pulse" class="w-4 h-4 text-success"></i>
                    </div>
                    <div class="stat-card-value text-display mono text-primary">Healthy</div>
                    <div class="stat-card-subtitle text-caption text-muted">v0.17.0</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Panel Group: Metrics Grid -->
          <div class="panel-group" :class="{ collapsed: metricsCollapsed }">
            <div class="panel-group-card card">
              <header class="panel-group-header" data-group="metrics">
                <button class="collapse-toggle" @click="toggleMetrics">
                  <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                  <span class="text-heading text-primary">Detailed Metrics</span>
                </button>
              </header>
              <div class="panel-group-body">
                <div class="panel-grid grid-3">
                  <div class="metric-panel card p-4">
                    <div class="flex items-center justify-between mb-3">
                      <span class="text-label text-primary">Response Time</span>
                      <i data-lucide="zap" class="w-5 h-5 text-accent"></i>
                    </div>
                    <div class="text-display text-primary mb-1">124ms</div>
                    <div class="text-caption text-muted">avg</div>
                  </div>
                  <div class="metric-panel card p-4">
                    <div class="flex items-center justify-between mb-3">
                      <span class="text-label text-primary">Error Rate</span>
                      <i data-lucide="alert-triangle" class="w-5 h-5 text-accent"></i>
                    </div>
                    <div class="text-display text-primary mb-1">0.02%</div>
                    <div class="text-caption text-muted">errors/total</div>
                  </div>
                  <div class="metric-panel card p-4">
                    <div class="flex items-center justify-between mb-3">
                      <span class="text-label text-primary">Throughput</span>
                      <i data-lucide="trending-up" class="w-5 h-5 text-accent"></i>
                    </div>
                    <div class="text-display text-primary mb-1">1.2K/min</div>
                    <div class="text-caption text-muted">requests</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Panel Group: Activity Feed -->
          <div class="panel-group" :class="{ collapsed: activityCollapsed }">
            <div class="panel-group-card card">
              <header class="panel-group-header" data-group="activity">
                <button class="collapse-toggle" @click="toggleActivity">
                  <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                  <span class="text-heading text-primary">Recent Activity</span>
                  <span class="text-caption text-accent ml-auto">View all</span>
                </button>
              </header>
              <div class="panel-group-body">
                <div class="activity-list">
                  <div class="activity-item">
                    <div class="activity-icon icon-box-sm">
                      <i data-lucide="check-circle" class="w-4 h-4"></i>
                    </div>
                    <div class="activity-content">
                      <div class="text-label text-primary">App deployed successfully</div>
                      <div class="text-caption text-muted">momentum</div>
                    </div>
                    <div class="activity-time text-caption text-faint">2h ago</div>
                  </div>
                  <div class="activity-item">
                    <div class="activity-icon icon-box-sm">
                      <i data-lucide="settings" class="w-4 h-4"></i>
                    </div>
                    <div class="activity-content">
                      <div class="text-label text-primary">Configuration updated</div>
                      <div class="text-caption text-muted">SSL certificates</div>
                    </div>
                    <div class="activity-time text-caption text-faint">5h ago</div>
                  </div>
                  <div class="activity-item">
                    <div class="activity-icon icon-box-sm">
                      <i data-lucide="alert-triangle" class="w-4 h-4"></i>
                    </div>
                    <div class="activity-content">
                      <div class="text-label text-primary">High memory usage detected</div>
                      <div class="text-caption text-muted">nexus app</div>
                    </div>
                    <div class="activity-time text-caption text-faint">1d ago</div>
                  </div>
                  <div class="activity-item">
                    <div class="activity-icon icon-box-sm">
                      <i data-lucide="upload" class="w-4 h-4"></i>
                    </div>
                    <div class="activity-content">
                      <div class="text-label text-primary">New version released</div>
                      <div class="text-caption text-muted">v0.17.0</div>
                    </div>
                    <div class="activity-time text-caption text-faint">2d ago</div>
                  </div>
                  <div class="activity-item">
                    <div class="activity-icon icon-box-sm">
                      <i data-lucide="key" class="w-4 h-4"></i>
                    </div>
                    <div class="activity-content">
                      <div class="text-label text-primary">API token generated</div>
                      <div class="text-caption text-muted">Admin user</div>
                    </div>
                    <div class="activity-time text-caption text-faint">3d ago</div>
                  </div>
                  <div class="activity-item">
                    <div class="activity-icon icon-box-sm">
                      <i data-lucide="shield-check" class="w-4 h-4"></i>
                    </div>
                    <div class="activity-content">
                      <div class="text-label text-primary">Security scan passed</div>
                      <div class="text-caption text-muted">All apps</div>
                    </div>
                    <div class="activity-time text-caption text-faint">4d ago</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Side Panel -->
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
}
