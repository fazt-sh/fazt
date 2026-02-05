import { ref, computed, onMounted, onUpdated } from 'vue'
import { useRouter } from 'vue-router'
import { useHealthStore } from '../stores/health.js'
import { useAppsStore } from '../stores/apps.js'
import { useAliasesStore } from '../stores/aliases.js'
import { useLogsStore } from '../stores/logs.js'
import { useUIStore } from '../stores/ui.js'
import { refreshIcons } from '../lib/icons.js'
import { formatBytes, formatUptime } from '../lib/format.js'

export default {
  name: 'DashboardPage',
  setup() {
    const router = useRouter()
    const healthStore = useHealthStore()
    const appsStore = useAppsStore()
    const aliasesStore = useAliasesStore()
    const logsStore = useLogsStore()
    const uiStore = useUIStore()

    const systemPanelCollapsed = ref(uiStore.getUIState('dashboard.system.collapsed', false))

    const toggleSystemPanel = () => {
      systemPanelCollapsed.value = !systemPanelCollapsed.value
      uiStore.setUIState('dashboard.system.collapsed', systemPanelCollapsed.value)
    }

    // System metrics
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
      return healthStore.data?.version ? 'v' + healthStore.data.version : '-'
    })

    const modeText = computed(() => {
      return healthStore.data?.mode || '-'
    })

    const goroutinesText = computed(() => {
      const count = healthStore.data?.runtime?.goroutines || 0
      return count + ' goroutines'
    })

    // Overview metrics
    const appsCount = computed(() => {
      return appsStore.items.length
    })

    const aliasesCount = computed(() => {
      return aliasesStore.items.length
    })

    const activityCount = computed(() => {
      return logsStore.stats?.total_count || 0
    })

    const navigateTo = (path) => {
      router.push(path)
    }

    onMounted(() => {
      refreshIcons()
    })

    onUpdated(() => {
      refreshIcons()
    })

    return {
      systemPanelCollapsed,
      toggleSystemPanel,
      statusClass,
      statusText,
      uptimeText,
      memoryText,
      storageText,
      versionText,
      modeText,
      goroutinesText,
      appsCount,
      aliasesCount,
      activityCount,
      navigateTo
    }
  },
  template: `
    <div class="design-system-page">
      <div class="content-container">
        <div class="content-scroll">

          <div class="panel-group" :class="{ collapsed: systemPanelCollapsed }">
            <div class="panel-group-card card">
              <header class="panel-group-header">
                <button class="collapse-toggle" @click="toggleSystemPanel">
                  <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                  <span class="text-heading text-primary">Dashboard</span>
                  <span class="text-caption text-faint ml-auto hide-mobile">7 metrics</span>
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
                    <div class="stat-card-subtitle text-caption text-muted">{{ appsCount }} apps</div>
                  </div>

                  <div class="stat-card card row-clickable" @click="navigateTo('/apps')">
                    <div class="stat-card-header">
                      <span class="text-micro text-muted">Apps</span>
                      <i data-lucide="layers" class="w-4 h-4 text-faint"></i>
                    </div>
                    <div class="stat-card-value text-display mono text-primary">{{ appsCount }}</div>
                    <div class="stat-card-subtitle text-caption text-muted">deployed</div>
                  </div>

                  <div class="stat-card card row-clickable" @click="navigateTo('/aliases')">
                    <div class="stat-card-header">
                      <span class="text-micro text-muted">Aliases</span>
                      <i data-lucide="link" class="w-4 h-4 text-faint"></i>
                    </div>
                    <div class="stat-card-value text-display mono text-primary">{{ aliasesCount }}</div>
                    <div class="stat-card-subtitle text-caption text-muted">configured</div>
                  </div>

                  <div class="stat-card card row-clickable" @click="navigateTo('/logs')">
                    <div class="stat-card-header">
                      <span class="text-micro text-muted">Activity</span>
                      <i data-lucide="activity" class="w-4 h-4 text-faint"></i>
                    </div>
                    <div class="stat-card-value text-display mono text-primary">{{ activityCount }}</div>
                    <div class="stat-card-subtitle text-caption text-muted">events logged</div>
                  </div>

                </div>
              </div>
            </div>
          </div>

        </div>
      </div>
    </div>
  `
}
