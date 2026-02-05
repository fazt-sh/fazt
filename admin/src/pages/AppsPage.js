import { ref, computed, onMounted, onUpdated } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppsStore } from '../stores/apps.js'
import { useUIStore } from '../stores/ui.js'
import { client } from '../client.js'
import { refreshIcons } from '../lib/icons.js'
import { formatBytes, formatRelativeTime } from '../lib/format.js'

export default {
  name: 'AppsPage',
  setup() {
    const route = useRoute()
    const router = useRouter()
    const store = useAppsStore()
    const uiStore = useUIStore()
    const searchQuery = ref('')
    const panelCollapsed = ref(uiStore.getUIState('apps.list.collapsed', false))

    const isDetailMode = computed(() => !!route.params.id)
    const currentAppId = computed(() => route.params.id)

    const currentApp = computed(() => {
      if (!currentAppId.value) return null
      return store.items.find(app => app.id === currentAppId.value)
    })

    const filteredApps = computed(() => {
      if (!searchQuery.value) return store.items
      const query = searchQuery.value.toLowerCase()
      return store.items.filter(app => {
        const title = (app.title || app.name || '').toLowerCase()
        const id = (app.id || '').toLowerCase()
        return title.includes(query) || id.includes(query)
      })
    })

    const togglePanel = () => {
      panelCollapsed.value = !panelCollapsed.value
      uiStore.setUIState('apps.list.collapsed', panelCollapsed.value)
    }

    const navigateToApp = (appId) => {
      router.push(`/apps/${appId}`)
    }

    const navigateToList = () => {
      router.push('/apps')
    }

    const openNewAppModal = () => {
      uiStore.newAppModalOpen = true
    }

    const deleteApp = async (appId) => {
      if (!confirm('Are you sure you want to delete this app? This action cannot be undone.')) {
        return
      }
      try {
        await store.deleteApp(client, appId)
        if (isDetailMode.value) {
          navigateToList()
        }
      } catch (error) {
        console.error('Failed to delete app:', error)
      }
    }

    onMounted(() => {
      store.load(client)
      refreshIcons()
    })

    onUpdated(() => {
      refreshIcons()
    })

    return {
      store,
      searchQuery,
      panelCollapsed,
      isDetailMode,
      currentApp,
      filteredApps,
      togglePanel,
      navigateToApp,
      navigateToList,
      openNewAppModal,
      deleteApp,
      formatBytes,
      formatRelativeTime
    }
  },
  template: `
    <div class="design-system-page">
      <div class="content-container">
        <div class="content-scroll">

          <!-- LIST MODE -->
          <template v-if="!isDetailMode">
            <div class="panel-group" :class="{ collapsed: panelCollapsed }">
              <div class="panel-group-card card" style="flex: 1; display: flex; flex-direction: column; min-height: 0">
                <header class="panel-group-header">
                  <button class="collapse-toggle" @click="togglePanel">
                    <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                    <span class="text-heading text-primary">Apps</span>
                    <span class="text-caption mono px-1.5 py-0.5 ml-2 badge-muted" style="border-radius: var(--radius-sm)">{{ store.items.length }}</span>
                  </button>
                </header>
                <div class="panel-group-body" style="padding: 0; flex: 1; display: flex; flex-direction: column; min-height: 0">
                  <div style="border: none; border-radius: 0; flex: 1; display: flex; flex-direction: column; min-height: 0">
                    <!-- Toolbar -->
                    <div class="card-header flex items-center justify-between" style="flex-shrink: 0">
                      <div class="input toolbar-search">
                        <i data-lucide="search" class="w-4 h-4 text-faint"></i>
                        <input type="text" placeholder="Filter..." v-model="searchQuery">
                      </div>
                      <div class="flex items-center gap-2" style="flex-shrink: 0">
                        <button class="btn btn-sm btn-primary toolbar-btn" title="New App" @click="openNewAppModal">
                          <i data-lucide="plus" class="w-4 h-4"></i>
                        </button>
                      </div>
                    </div>
                    <!-- Table -->
                    <div class="panel-scroll-area scroll-panel" style="flex: 1; overflow: auto; min-height: 0">
                      <!-- Empty state -->
                      <div v-if="filteredApps.length === 0" class="flex flex-col items-center justify-center p-8 text-center" style="min-height: 200px">
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
                            <tr v-for="app in filteredApps" :key="app.id"
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
                    <!-- Footer -->
                    <div style="flex-shrink: 0">
                      <div class="card-footer flex items-center justify-between" style="border-radius: 0">
                        <span class="text-caption text-muted">{{ store.items.length }} app{{ store.items.length === 1 ? '' : 's' }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </template>

          <!-- DETAIL MODE -->
          <template v-else>
            <div v-if="!currentApp" class="flex items-center justify-center h-full">
              <div class="text-caption text-muted">Loading...</div>
            </div>
            <div v-else>
              <!-- Back button -->
              <div class="mb-4">
                <button class="flex items-center gap-2 text-caption text-muted cursor-pointer" @click="navigateToList" style="background:none;border:none">
                  <i data-lucide="arrow-left" class="w-4 h-4"></i>
                  Back to Apps
                </button>
              </div>

              <!-- App header -->
              <div class="flex items-center gap-4 mb-6">
                <div class="w-12 h-12 flex items-center justify-center" style="background:var(--accent-soft);border-radius:var(--radius-lg)">
                  <i data-lucide="box" class="w-6 h-6" style="color:var(--accent)"></i>
                </div>
                <div class="flex-1">
                  <h1 class="text-title text-primary">{{ currentApp.title || currentApp.name }}</h1>
                  <div class="text-caption mono text-muted">{{ currentApp.id }}</div>
                </div>
                <div class="flex items-center gap-2">
                  <button class="btn btn-secondary btn-sm" title="Refresh">
                    <i data-lucide="refresh-cw" class="w-4 h-4"></i>
                  </button>
                  <button class="btn btn-secondary btn-sm" style="color:var(--error)" title="Delete" @click="deleteApp(currentApp.id)">
                    <i data-lucide="trash-2" class="w-4 h-4"></i>
                  </button>
                </div>
              </div>

              <!-- Stat cards -->
              <div class="panel-grid grid-4 mb-6">
                <div class="stat-card card">
                  <div class="stat-card-header">
                    <span class="text-micro text-muted">Files</span>
                    <i data-lucide="file" class="w-4 h-4 text-faint"></i>
                  </div>
                  <div class="stat-card-value text-display mono text-primary">{{ currentApp.file_count || 0 }}</div>
                  <div class="stat-card-subtitle text-caption text-muted">files deployed</div>
                </div>
                <div class="stat-card card">
                  <div class="stat-card-header">
                    <span class="text-micro text-muted">Size</span>
                    <i data-lucide="hard-drive" class="w-4 h-4 text-faint"></i>
                  </div>
                  <div class="stat-card-value text-display mono text-primary">{{ formatBytes(currentApp.size_bytes) }}</div>
                  <div class="stat-card-subtitle text-caption text-muted">total storage</div>
                </div>
                <div class="stat-card card">
                  <div class="stat-card-header">
                    <span class="text-micro text-muted">Updated</span>
                    <i data-lucide="clock" class="w-4 h-4 text-faint"></i>
                  </div>
                  <div class="stat-card-value text-display mono text-primary">{{ formatRelativeTime(currentApp.updated_at) }}</div>
                  <div class="stat-card-subtitle text-caption text-muted">last deploy</div>
                </div>
                <div class="stat-card card">
                  <div class="stat-card-header">
                    <span class="text-micro text-muted">Status</span>
                    <i data-lucide="activity" class="w-4 h-4 text-faint"></i>
                  </div>
                  <div class="stat-card-value text-display mono text-success">
                    <span class="flex items-center gap-2">
                      <span class="status-dot status-dot-success pulse"></span>
                      Live
                    </span>
                  </div>
                  <div class="stat-card-subtitle text-caption text-muted">serving traffic</div>
                </div>
              </div>

              <!-- Details card -->
              <div class="card mb-4">
                <div class="card-header">
                  <span class="text-heading text-primary">Details</span>
                </div>
                <div class="card-body">
                  <div class="details-list">
                    <div class="detail-item">
                      <span class="detail-label">App ID</span>
                      <span class="detail-value mono">{{ currentApp.id }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">Name</span>
                      <span class="detail-value">{{ currentApp.name }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">Title</span>
                      <span class="detail-value">{{ currentApp.title || '-' }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">Created</span>
                      <span class="detail-value">{{ formatRelativeTime(currentApp.created_at) }}</span>
                    </div>
                    <div class="detail-item">
                      <span class="detail-label">Updated</span>
                      <span class="detail-value">{{ formatRelativeTime(currentApp.updated_at) }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </template>

        </div>
      </div>
    </div>
  `
}
