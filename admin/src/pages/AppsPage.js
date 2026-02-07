import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppsStore } from '../stores/apps.js'
import { client } from '../client.js'
import { useIcons } from '../lib/useIcons.js'
import { usePanel } from '../lib/usePanel.js'
import { formatBytes, formatRelativeTime } from '../lib/format.js'
import { useUIStore } from '../stores/ui.js'
import FPanel from '../components/FPanel.js'
import FTable from '../components/FTable.js'
import FToolbar from '../components/FToolbar.js'

export default {
  name: 'AppsPage',
  components: { FPanel, FTable, FToolbar },
  setup() {
    useIcons()
    const route = useRoute()
    const router = useRouter()
    const store = useAppsStore()
    const uiStore = useUIStore()
    const panel = usePanel('apps.list.collapsed', false)
    const searchQuery = ref('')

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

    const columns = [
      { key: 'title', label: 'App' },
      { key: 'updated_at', label: 'Updated', hideOnMobile: true },
      { key: 'size_bytes', label: 'Size', hideOnMobile: true },
      { key: 'status', label: 'Status', hideOnMobile: true },
    ]

    const navigateToApp = (appId) => router.push(`/apps/${appId}`)
    const navigateToList = () => router.push('/apps')
    const openNewAppModal = () => { uiStore.newAppModalOpen = true }

    const deleteApp = async (appId) => {
      if (!confirm('Are you sure you want to delete this app? This action cannot be undone.')) return
      try {
        await store.remove(client, appId)
        if (isDetailMode.value) navigateToList()
      } catch (error) {
        console.error('Failed to delete app:', error)
      }
    }

    onMounted(() => { store.load(client) })

    return {
      store, panel, searchQuery, isDetailMode, currentApp, filteredApps, columns,
      navigateToApp, navigateToList, openNewAppModal, deleteApp,
      formatBytes, formatRelativeTime
    }
  },
  template: `
    <div class="design-system-page">
      <div class="content-container">
        <div class="content-scroll">

          <!-- LIST MODE -->
          <template v-if="!isDetailMode">
            <FPanel title="Apps" :count="store.items.length" mode="fill"
                    :collapsed="panel.collapsed" @update:collapsed="panel.toggle">
              <template #toolbar>
                <FToolbar v-model="searchQuery">
                  <template #actions>
                    <button class="btn btn-sm btn-primary toolbar-btn" title="New App" @click="openNewAppModal">
                      <i data-lucide="plus" class="w-4 h-4"></i>
                    </button>
                  </template>
                </FToolbar>
              </template>

              <FTable :columns="columns" :rows="filteredApps" row-key="id"
                      empty-icon="layers" empty-title="No apps yet" empty-message="Deploy your first app via CLI"
                      @row-click="navigateToApp($event.id)">
                <template #cell-title="{ row }">
                  <div class="flex items-center gap-2" style="min-width: 0">
                    <span class="status-dot status-dot-success pulse show-mobile" style="flex-shrink: 0"></span>
                    <div class="icon-box icon-box-sm" style="flex-shrink: 0">
                      <i data-lucide="box" class="w-3.5 h-3.5"></i>
                    </div>
                    <div style="min-width: 0; overflow: hidden">
                      <div class="text-label text-primary truncate">{{ row.title || row.name }}</div>
                      <div class="text-caption mono text-faint show-mobile">{{ formatBytes(row.size_bytes) }}</div>
                    </div>
                  </div>
                </template>
                <template #cell-updated_at="{ row }">
                  <span class="text-caption text-muted">{{ formatRelativeTime(row.updated_at) }}</span>
                </template>
                <template #cell-size_bytes="{ row }">
                  <span class="text-caption mono text-muted">{{ formatBytes(row.size_bytes) }}</span>
                </template>
                <template #cell-status>
                  <span class="flex items-center gap-1 text-caption text-success">
                    <span class="status-dot status-dot-success pulse"></span>
                    Live
                  </span>
                </template>
              </FTable>

              <template #footer>
                <div class="card-footer flex items-center justify-between" style="border-radius: 0">
                  <span class="text-caption text-muted">{{ store.items.length }} app{{ store.items.length === 1 ? '' : 's' }}</span>
                </div>
              </template>
            </FPanel>
          </template>

          <!-- DETAIL MODE -->
          <template v-else>
            <div v-if="!currentApp" class="flex items-center justify-center h-full">
              <div class="text-caption text-muted">Loading...</div>
            </div>
            <div v-else>
              <div class="mb-4">
                <button class="flex items-center gap-2 text-caption text-muted cursor-pointer" @click="navigateToList" style="background:none;border:none">
                  <i data-lucide="arrow-left" class="w-4 h-4"></i>
                  Back to Apps
                </button>
              </div>

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
