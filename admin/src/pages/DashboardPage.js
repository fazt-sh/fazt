import { ref, computed, onMounted, onUpdated } from 'vue'
import { useRouter } from 'vue-router'
import { useHealthStore } from '../stores/health.js'
import { useAppsStore } from '../stores/apps.js'
import { useAliasesStore } from '../stores/aliases.js'
import { useUIStore } from '../stores/ui.js'
import { client } from '../client.js'
import { refreshIcons } from '../lib/icons.js'
import { formatBytes, formatUptime, formatRelativeTime } from '../lib/format.js'

export default {
  name: 'DashboardPage',
  setup() {
    const router = useRouter()
    const healthStore = useHealthStore()
    const appsStore = useAppsStore()
    const aliasesStore = useAliasesStore()
    const uiStore = useUIStore()

    const systemPanelCollapsed = ref(uiStore.getUIState('dashboard.system.collapsed', false))
    const appsPanelCollapsed = ref(uiStore.getUIState('dashboard.apps.collapsed', false))
    const aliasesPanelCollapsed = ref(uiStore.getUIState('dashboard.aliases.collapsed', false))

    const toggleSystemPanel = () => {
      systemPanelCollapsed.value = !systemPanelCollapsed.value
      uiStore.setUIState('dashboard.system.collapsed', systemPanelCollapsed.value)
    }

    const toggleAppsPanel = () => {
      appsPanelCollapsed.value = !appsPanelCollapsed.value
      uiStore.setUIState('dashboard.apps.collapsed', appsPanelCollapsed.value)
    }

    const toggleAliasesPanel = () => {
      aliasesPanelCollapsed.value = !aliasesPanelCollapsed.value
      uiStore.setUIState('dashboard.aliases.collapsed', aliasesPanelCollapsed.value)
    }

    const topApps = computed(() => {
      return appsStore.items.slice(0, 5)
    })

    const topAliases = computed(() => {
      return aliasesStore.items.slice(0, 5)
    })

    const statusClass = computed(() => {
      const status = healthStore.data?.status || 'unknown'
      if (status === 'healthy') return 'text-success'
      if (status === 'warning') return 'text-warning'
      return 'text-error'
    })

    const statusText = computed(() => {
      return healthStore.data?.status || '-'
    })

    const uptimeText = computed(() => {
      const uptime = healthStore.data?.uptime_seconds || 0
      return formatUptime(uptime)
    })

    const memoryText = computed(() => {
      const memory = healthStore.data?.memory?.used_mb || healthStore.data?.memory_used_mb || 0
      return memory.toFixed(1)
    })

    const storageText = computed(() => {
      const totalStorage = appsStore.items.reduce((sum, app) => sum + (app.size_bytes || 0), 0)
      return formatBytes(totalStorage)
    })

    const versionText = computed(() => {
      return healthStore.data?.version ? `v${healthStore.data.version}` : '-'
    })

    const modeText = computed(() => {
      return healthStore.data?.mode || '-'
    })

    const goroutinesText = computed(() => {
      return `${healthStore.data?.runtime?.goroutines || 0} goroutines`
    })

    const navigateToApp = (appId) => {
      router.push(`/apps/${appId}`)
    }

    const navigateTo = (path) => {
      router.push(path)
    }

    onMounted(() => {
      healthStore.load(client)
      appsStore.load(client)
      aliasesStore.load(client)
      refreshIcons()
    })

    onUpdated(() => {
      refreshIcons()
    })

    return {
      router,
      healthStore,
      appsStore,
      aliasesStore,
      systemPanelCollapsed,
      appsPanelCollapsed,
      aliasesPanelCollapsed,
      toggleSystemPanel,
      toggleAppsPanel,
      toggleAliasesPanel,
      topApps,
      topAliases,
      statusClass,
      statusText,
      uptimeText,
      memoryText,
      storageText,
      versionText,
      modeText,
      goroutinesText,
      navigateToApp,
      navigateTo,
      formatBytes,
      formatRelativeTime
    }
  },
  template: `
    <div class="design-system-page">
      <div class="content-container">
        <div class="content-scroll">

          <!-- Panel Group: System Status -->
          <div class="panel-group" :class="{ collapsed: systemPanelCollapsed }">
            <div class="panel-group-card card">
              <header class="panel-group-header">
                <button class="collapse-toggle" @click="toggleSystemPanel">
                  <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                  <span class="text-heading text-primary">System</span>
                  <span class="text-caption text-faint ml-auto hide-mobile">4 metrics</span>
                </button>
              </header>
              <div class="panel-group-body">
                <div class="panel-grid grid-4">
                  <div class="stat-card card">
                    <div class="stat-card-header">
                      <span class="text-micro text-muted">Status</span>
                      <i data-lucide="heart-pulse" class="w-4 h-4 text-faint"></i>
                    </div>
                    <div class="stat-card-value text-display mono" :class="statusClass">{{ statusText }}</div>
                    <div class="stat-card-subtitle text-caption text-muted">{{ versionText }}</div>
                  </div>

                  <div class="stat-card card">
                    <div class="stat-card-header">
                      <span class="text-micro text-muted">Uptime</span>
                      <i data-lucide="clock" class="w-4 h-4 text-faint"></i>
                    </div>
                    <div class="stat-card-value text-display mono text-primary">{{ uptimeText }}</div>
                    <div class="stat-card-subtitle text-caption text-muted">{{ modeText }}</div>
                  </div>

                  <div class="stat-card card">
                    <div class="stat-card-header">
                      <span class="text-micro text-muted">Memory</span>
                      <i data-lucide="cpu" class="w-4 h-4 text-faint"></i>
                    </div>
                    <div class="stat-card-value text-display mono text-primary">{{ memoryText }}<span class="text-caption">MB</span></div>
                    <div class="stat-card-subtitle text-caption text-muted">{{ goroutinesText }}</div>
                  </div>

                  <div class="stat-card card">
                    <div class="stat-card-header">
                      <span class="text-micro text-muted">Storage</span>
                      <i data-lucide="database" class="w-4 h-4 text-faint"></i>
                    </div>
                    <div class="stat-card-value text-display mono text-primary">{{ storageText }}</div>
                    <div class="stat-card-subtitle text-caption text-muted">{{ appsStore.items.length }} apps</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Panel Group: Apps -->
          <div class="panel-group" :class="{ collapsed: appsPanelCollapsed }">
            <div class="panel-group-card card">
              <header class="panel-group-header">
                <button class="collapse-toggle" @click="toggleAppsPanel">
                  <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                  <span class="text-heading text-primary">Apps</span>
                  <span class="nav-badge text-caption" v-if="appsStore.items.length">{{ appsStore.items.length }}</span>
                </button>
              </header>
              <div class="panel-group-body" style="min-height: 200px; max-height: 300px; overflow-y: auto">
                <!-- Loading state -->
                <div v-if="appsStore.loading" class="flex flex-col items-center justify-center p-8 text-center" style="min-height: 200px">
                  <i data-lucide="loader" class="w-5 h-5 text-faint spin"></i>
                  <div class="text-caption text-muted mt-2">Loading apps...</div>
                </div>

                <!-- Empty state -->
                <div v-else-if="topApps.length === 0" class="flex flex-col items-center justify-center p-8 text-center" style="min-height: 200px">
                  <div class="icon-box mb-3" style="width:48px;height:48px;opacity:0.5">
                    <i data-lucide="layers" class="w-6 h-6"></i>
                  </div>
                  <div class="text-heading text-primary mb-1">No apps yet</div>
                  <div class="text-caption text-muted">Deploy your first app via CLI</div>
                </div>

                <!-- Table -->
                <div v-else class="table-container" style="overflow-x: visible">
                  <table style="width: 100%; min-width: 0">
                    <thead class="sticky" style="top: 0; background: var(--bg-1)">
                      <tr style="border-bottom: 1px solid var(--border-subtle)">
                        <th class="px-3 py-2 text-left text-micro text-muted">App</th>
                        <th class="px-3 py-2 text-left text-micro text-muted hide-mobile">Updated</th>
                        <th class="px-3 py-2 text-left text-micro text-muted hide-mobile">Size</th>
                        <th class="px-3 py-2 text-left text-micro text-muted hide-mobile">Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="app in topApps"
                          :key="app.id"
                          class="row row-clickable"
                          style="border-bottom: 1px solid var(--border-subtle)"
                          @click="navigateToApp(app.id)">
                        <td class="px-3 py-2">
                          <div class="flex items-center gap-2" style="min-width: 0">
                            <span class="status-dot status-dot-success pulse show-mobile" style="flex-shrink: 0"></span>
                            <div class="icon-box icon-box-sm" style="flex-shrink: 0">
                              <i data-lucide="box" class="w-3.5 h-3.5"></i>
                            </div>
                            <div style="min-width: 0; overflow: hidden">
                              <div class="text-label text-primary" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ app.title || app.name }}</div>
                              <div class="text-caption mono text-faint show-mobile">{{ formatBytes(app.size_bytes) }}</div>
                            </div>
                          </div>
                        </td>
                        <td class="px-3 py-2 hide-mobile">
                          <span class="text-caption text-muted">{{ formatRelativeTime(app.updated_at) }}</span>
                        </td>
                        <td class="px-3 py-2 hide-mobile">
                          <span class="text-caption mono text-muted">{{ formatBytes(app.size_bytes) }}</span>
                        </td>
                        <td class="px-3 py-2 hide-mobile">
                          <span class="flex items-center gap-1 text-caption text-success">
                            <span class="status-dot status-dot-success pulse"></span>
                            Live
                          </span>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
              <div class="card-footer flex items-center justify-between" style="border-radius: 0">
                <span class="text-caption text-muted">{{ appsStore.items.length }} app{{ appsStore.items.length === 1 ? '' : 's' }}</span>
                <button class="text-caption font-medium text-accent" @click="navigateTo('/apps')">View all &rarr;</button>
              </div>
            </div>
          </div>

          <!-- Panel Group: Aliases -->
          <div class="panel-group" :class="{ collapsed: aliasesPanelCollapsed }">
            <div class="panel-group-card card">
              <header class="panel-group-header">
                <button class="collapse-toggle" @click="toggleAliasesPanel">
                  <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                  <span class="text-heading text-primary">Aliases</span>
                  <span class="nav-badge text-caption" v-if="aliasesStore.items.length">{{ aliasesStore.items.length }}</span>
                </button>
              </header>
              <div class="panel-group-body" style="min-height: 200px; max-height: 300px; overflow-y: auto">
                <!-- Loading state -->
                <div v-if="aliasesStore.loading" class="flex flex-col items-center justify-center p-8 text-center" style="min-height: 200px">
                  <i data-lucide="loader" class="w-5 h-5 text-faint spin"></i>
                  <div class="text-caption text-muted mt-2">Loading aliases...</div>
                </div>

                <!-- Empty state -->
                <div v-else-if="topAliases.length === 0" class="flex flex-col items-center justify-center p-8 text-center" style="min-height: 200px">
                  <div class="icon-box mb-3" style="width:48px;height:48px;opacity:0.5">
                    <i data-lucide="link" class="w-6 h-6"></i>
                  </div>
                  <div class="text-heading text-primary mb-1">No aliases yet</div>
                  <div class="text-caption text-muted">Create aliases via CLI</div>
                </div>

                <!-- Table -->
                <div v-else class="table-container" style="overflow-x: visible">
                  <table style="width: 100%; min-width: 0">
                    <thead class="sticky" style="top: 0; background: var(--bg-1)">
                      <tr style="border-bottom: 1px solid var(--border-subtle)">
                        <th class="px-3 py-2 text-left text-micro text-muted">Alias</th>
                        <th class="px-3 py-2 text-left text-micro text-muted hide-mobile">Type</th>
                        <th class="px-3 py-2 text-left text-micro text-muted hide-mobile">Target</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="alias in topAliases"
                          :key="alias.subdomain"
                          class="row row-clickable"
                          style="border-bottom: 1px solid var(--border-subtle)"
                          @click="navigateTo('/aliases')">
                        <td class="px-3 py-2">
                          <div class="flex items-center gap-2" style="min-width: 0">
                            <div class="icon-box icon-box-sm" style="flex-shrink: 0">
                              <i data-lucide="link" class="w-3.5 h-3.5"></i>
                            </div>
                            <div style="min-width: 0; overflow: hidden">
                              <div class="text-label text-primary" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ alias.subdomain }}</div>
                              <div class="text-caption text-faint show-mobile">{{ alias.type }}</div>
                            </div>
                          </div>
                        </td>
                        <td class="px-3 py-2 hide-mobile">
                          <span class="text-caption text-muted">{{ alias.type }}</span>
                        </td>
                        <td class="px-3 py-2 hide-mobile">
                          <span class="text-caption mono text-muted">{{ alias.targets?.app_id || alias.targets?.url || '-' }}</span>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
              <div class="card-footer flex items-center justify-between" style="border-radius: 0">
                <span class="text-caption text-muted">{{ aliasesStore.items.length }} alias{{ aliasesStore.items.length === 1 ? '' : 'es' }}</span>
                <button class="text-caption font-medium text-accent" @click="navigateTo('/aliases')">View all &rarr;</button>
              </div>
            </div>
          </div>

        </div>
      </div>
    </div>
  `
}
