import { computed, onMounted } from 'vue'
import { useHealthStore } from '../stores/health.js'
import { client } from '../client.js'
import { useIcons } from '../lib/useIcons.js'
import { formatBytes, formatUptime } from '../lib/format.js'

export default {
  name: 'SystemPage',
  setup() {
    useIcons()
    const store = useHealthStore()

    const statusClass = computed(() => {
      const status = store.data?.status || 'unknown'
      if (status === 'healthy') return 'status-healthy'
      if (status === 'warning') return 'status-warning'
      return 'status-error'
    })

    const statusText = computed(() => {
      const status = store.data?.status || 'unknown'
      return status.charAt(0).toUpperCase() + status.slice(1)
    })

    const statusDotClass = computed(() => {
      const status = store.data?.status || 'unknown'
      if (status === 'healthy') return 'status-dot status-dot-success pulse'
      return 'status-dot status-dot-warning pulse'
    })

    const statusColorClass = computed(() => {
      const status = store.data?.status || 'unknown'
      if (status === 'healthy') return 'text-success'
      return 'text-warning'
    })

    const uptimeText = computed(() => formatUptime(store.data?.uptime_seconds || 0))
    const memoryUsed = computed(() => store.data?.memory?.used_mb || 0)
    const memoryLimit = computed(() => store.data?.memory?.limit_mb || 0)
    const memoryPercent = computed(() => {
      const limit = memoryLimit.value || 512
      if (limit === 0) return 0
      return Math.round((memoryUsed.value / limit) * 100)
    })
    const vfsCacheMb = computed(() => store.data?.memory?.vfs_cache_mb || 0)
    const dbPath = computed(() => store.data?.database?.path || 'Unknown')
    const dbConnections = computed(() => store.data?.database?.open_connections || 0)
    const dbInUse = computed(() => store.data?.database?.in_use || 0)
    const queuedEvents = computed(() => store.data?.runtime?.queued_events || 0)
    const goroutines = computed(() => store.data?.runtime?.goroutines || 0)
    const version = computed(() => store.data?.version || 'Unknown')
    const mode = computed(() => store.data?.mode || 'Unknown')

    const refresh = () => { store.load(client) }

    onMounted(() => { store.load(client) })

    return {
      store, statusClass, statusText, statusDotClass, statusColorClass,
      uptimeText, memoryUsed, memoryLimit, memoryPercent, vfsCacheMb,
      dbPath, dbConnections, dbInUse, queuedEvents, goroutines,
      version, mode, refresh, formatBytes, formatUptime
    }
  },
  template: `
    <div class="design-system-page">
      <div class="content-container">
        <div class="content-scroll">

          <!-- Header -->
          <div class="flex items-center justify-between mb-4 flex-shrink-0">
            <div>
              <h1 class="text-title text-primary">System</h1>
              <p class="text-caption text-muted">Health and configuration</p>
            </div>
            <button class="btn btn-secondary" @click="refresh">
              <i data-lucide="refresh-cw" class="w-4 h-4"></i>
              Refresh
            </button>
          </div>

          <!-- Content -->
          <div class="panel-grid grid-2">
            <!-- Health Status -->
            <div class="card">
              <div class="card-header">
                <span class="text-heading text-primary">Health Status</span>
                <span class="flex items-center gap-1 text-caption" :class="statusColorClass">
                  <span :class="statusDotClass"></span>
                  {{ store.data?.status || 'Unknown' }}
                </span>
              </div>
              <div class="card-body">
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <div class="text-micro text-muted mb-1">Uptime</div>
                    <div class="text-heading mono text-primary">{{ uptimeText }}</div>
                  </div>
                  <div>
                    <div class="text-micro text-muted mb-1">Version</div>
                    <div class="text-heading mono text-primary">{{ version }}</div>
                  </div>
                  <div>
                    <div class="text-micro text-muted mb-1">Mode</div>
                    <div class="text-heading text-primary">{{ mode }}</div>
                  </div>
                  <div>
                    <div class="text-micro text-muted mb-1">Goroutines</div>
                    <div class="text-heading mono text-primary">{{ goroutines }}</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Memory -->
            <div class="card">
              <div class="card-header">
                <span class="text-heading text-primary">Memory</span>
              </div>
              <div class="card-body">
                <div class="mb-4">
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-caption text-muted">Used</span>
                    <span class="text-caption mono text-primary">{{ memoryUsed.toFixed(1) }} MB</span>
                  </div>
                  <div class="progress">
                    <div class="progress-bar" :style="{ width: memoryPercent + '%' }"></div>
                  </div>
                </div>
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <div class="text-micro text-muted mb-1">Limit</div>
                    <div class="text-heading mono text-primary">{{ memoryLimit.toFixed(0) }} MB</div>
                  </div>
                  <div>
                    <div class="text-micro text-muted mb-1">VFS Cache</div>
                    <div class="text-heading mono text-primary">{{ vfsCacheMb.toFixed(1) }} MB</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Database -->
            <div class="card">
              <div class="card-header">
                <span class="text-heading text-primary">Database</span>
              </div>
              <div class="card-body">
                <div class="mb-3">
                  <div class="text-micro text-muted mb-1">Path</div>
                  <div class="text-caption mono text-secondary truncate">{{ dbPath }}</div>
                </div>
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <div class="text-micro text-muted mb-1">Open Connections</div>
                    <div class="text-heading mono text-primary">{{ dbConnections }}</div>
                  </div>
                  <div>
                    <div class="text-micro text-muted mb-1">In Use</div>
                    <div class="text-heading mono text-primary">{{ dbInUse }}</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Runtime -->
            <div class="card">
              <div class="card-header">
                <span class="text-heading text-primary">Runtime</span>
              </div>
              <div class="card-body">
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <div class="text-micro text-muted mb-1">Queued Events</div>
                    <div class="text-heading mono text-primary">{{ queuedEvents }}</div>
                  </div>
                  <div>
                    <div class="text-micro text-muted mb-1">Goroutines</div>
                    <div class="text-heading mono text-primary">{{ goroutines }}</div>
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
