import { ref, computed, onMounted, onUpdated } from 'vue'
import { useAliasesStore } from '../stores/aliases.js'
import { useAppsStore } from '../stores/apps.js'
import { useUIStore } from '../stores/ui.js'
import { client } from '../client.js'
import { refreshIcons } from '../lib/icons.js'
import { formatRelativeTime } from '../lib/format.js'

export default {
  name: 'AliasesPage',
  setup() {
    const store = useAliasesStore()
    const appsStore = useAppsStore()
    const uiStore = useUIStore()
    const searchQuery = ref('')
    const panelCollapsed = ref(uiStore.getUIState('aliases.list.collapsed', false))

    const filteredAliases = computed(() => {
      if (!searchQuery.value) return store.items
      const query = searchQuery.value.toLowerCase()
      return store.items.filter(alias => {
        const subdomain = (alias.subdomain || '').toLowerCase()
        const type = (alias.type || '').toLowerCase()
        return subdomain.includes(query) || type.includes(query)
      })
    })

    const togglePanel = () => {
      panelCollapsed.value = !panelCollapsed.value
      uiStore.setUIState('aliases.list.collapsed', panelCollapsed.value)
    }

    const getTypeBadge = (type) => {
      switch (type) {
        case 'proxy': return 'badge-success'
        case 'redirect': return 'badge-warning'
        case 'split': return 'badge'
        case 'reserved': return 'badge-muted'
        default: return 'badge-muted'
      }
    }

    const getAliasIcon = (type) => {
      switch (type) {
        case 'redirect': return 'external-link'
        case 'reserved': return 'lock'
        default: return 'link'
      }
    }

    const getAppName = (appId) => {
      if (!appId) return '-'
      const app = appsStore.items.find(a => a.id === appId)
      return app?.name || appId
    }

    const getTarget = (alias) => {
      if (alias.type === 'proxy') return getAppName(alias.targets?.app_id)
      if (alias.type === 'redirect') return alias.targets?.url || '-'
      return '-'
    }

    const openNewAliasModal = () => {
      uiStore.createAliasModalOpen = true
    }

    const openEditAliasModal = (alias) => {
      uiStore.editingAlias = alias
      uiStore.editAliasModalOpen = true
    }

    onMounted(() => {
      store.load(client)
      appsStore.load(client)
      refreshIcons()
    })

    onUpdated(() => {
      refreshIcons()
    })

    return {
      store,
      searchQuery,
      panelCollapsed,
      filteredAliases,
      togglePanel,
      getTypeBadge,
      getAliasIcon,
      getTarget,
      openNewAliasModal,
      openEditAliasModal,
      formatRelativeTime
    }
  },
  template: `
    <div class="design-system-page">
      <div class="content-container">
        <div class="content-scroll">

          <div class="panel-group" :class="{ collapsed: panelCollapsed }">
            <div class="panel-group-card card" style="flex: 1; display: flex; flex-direction: column; min-height: 0">
              <header class="panel-group-header">
                <button class="collapse-toggle" @click="togglePanel">
                  <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                  <span class="text-heading text-primary">Aliases</span>
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
                      <button class="btn btn-sm btn-primary toolbar-btn" title="New Alias" @click="openNewAliasModal">
                        <i data-lucide="plus" class="w-4 h-4"></i>
                      </button>
                    </div>
                  </div>
                  <!-- Table -->
                  <div class="panel-scroll-area scroll-panel" style="flex: 1; overflow: auto; min-height: 0">
                    <!-- Empty state -->
                    <div v-if="filteredAliases.length === 0" class="flex flex-col items-center justify-center p-8 text-center" style="min-height: 200px">
                      <div class="icon-box mb-3" style="width:48px;height:48px;opacity:0.5">
                        <i :data-lucide="searchQuery ? 'search-x' : 'link'" class="w-6 h-6"></i>
                      </div>
                      <div class="text-heading text-primary mb-1">{{ searchQuery ? 'No aliases found' : 'No aliases yet' }}</div>
                      <div class="text-caption text-muted">{{ searchQuery ? 'Try a different search term' : 'Create your first alias to get started' }}</div>
                    </div>
                    <!-- Table -->
                    <div v-else class="table-container" style="overflow-x: visible">
                      <table style="width: 100%; min-width: 0">
                        <thead class="sticky" style="top: 0; background: var(--bg-1)">
                          <tr style="border-bottom: 1px solid var(--border-subtle)">
                            <th class="px-3 py-2 text-left text-micro text-muted">Subdomain</th>
                            <th class="px-3 py-2 text-left text-micro text-muted hide-mobile">Type</th>
                            <th class="px-3 py-2 text-left text-micro text-muted hide-mobile">Target</th>
                            <th class="px-3 py-2 text-left text-micro text-muted hide-mobile">Updated</th>
                            <th class="px-3 py-2 text-left text-micro text-muted hide-mobile">Status</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr v-for="alias in filteredAliases" :key="alias.subdomain"
                              class="row row-clickable"
                              style="border-bottom: 1px solid var(--border-subtle)"
                              @click="openEditAliasModal(alias)">
                            <td class="px-3 py-2">
                              <div class="flex items-center gap-2" style="min-width: 0">
                                <span class="status-dot show-mobile" :class="alias.type === 'reserved' ? '' : 'status-dot-success pulse'" style="flex-shrink: 0"></span>
                                <div class="icon-box icon-box-sm" style="flex-shrink: 0">
                                  <i :data-lucide="getAliasIcon(alias.type)" class="w-3.5 h-3.5"></i>
                                </div>
                                <div style="min-width: 0; overflow: hidden">
                                  <div class="text-label text-primary" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ alias.subdomain }}</div>
                                  <div class="text-caption mono text-faint show-mobile">{{ alias.type }}</div>
                                </div>
                              </div>
                            </td>
                            <td class="px-3 py-2 hide-mobile">
                              <span class="badge" :class="getTypeBadge(alias.type)">{{ alias.type }}</span>
                            </td>
                            <td class="px-3 py-2 hide-mobile">
                              <span class="text-caption mono text-muted" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap; display: block; max-width: 180px">{{ getTarget(alias) }}</span>
                            </td>
                            <td class="px-3 py-2 hide-mobile">
                              <span class="text-caption text-muted">{{ formatRelativeTime(alias.updated_at) }}</span>
                            </td>
                            <td class="px-3 py-2 hide-mobile">
                              <span class="flex items-center gap-1 text-caption" :class="alias.type === 'reserved' ? 'text-muted' : 'text-success'">
                                <span class="status-dot" :class="alias.type === 'reserved' ? '' : 'status-dot-success pulse'"></span>
                                {{ alias.type === 'reserved' ? 'Reserved' : 'Active' }}
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
                      <span class="text-caption text-muted">{{ store.items.length }} alias{{ store.items.length === 1 ? '' : 'es' }}</span>
                    </div>
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
