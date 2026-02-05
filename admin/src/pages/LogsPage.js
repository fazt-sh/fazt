import { ref, computed, onMounted, onUpdated, onBeforeUnmount } from 'vue'
import { useLogsStore } from '../stores/logs.js'
import { useUIStore } from '../stores/ui.js'
import { client } from '../client.js'
import { refreshIcons } from '../lib/icons.js'
import { formatBytes, formatRelativeTime } from '../lib/format.js'

const WEIGHT_INFO = {
  9: { label: 'Security', color: 'var(--error)' },
  8: { label: 'Auth', color: 'var(--error)' },
  7: { label: 'Config', color: 'var(--warning)' },
  6: { label: 'Deploy', color: 'var(--accent)' },
  5: { label: 'Data', color: 'var(--accent)' },
  4: { label: 'Action', color: 'var(--text-2)' },
  3: { label: 'Nav', color: 'var(--text-3)' },
  2: { label: 'Analytics', color: 'var(--text-3)' },
  1: { label: 'System', color: 'var(--text-4)' },
  0: { label: 'Debug', color: 'var(--text-4)' }
}

const ACTOR_INFO = {
  user: { label: 'User', icon: 'user' },
  system: { label: 'System', icon: 'server' },
  api_key: { label: 'API Key', icon: 'key' },
  anonymous: { label: 'Anonymous', icon: 'user-x' }
}

export default {
  name: 'LogsPage',
  setup() {
    const store = useLogsStore()
    const uiStore = useUIStore()
    const panelCollapsed = ref(uiStore.getUIState('logs.list.collapsed', false))

    // Dropdown visibility
    const weightMenuOpen = ref(false)
    const actionMenuOpen = ref(false)
    const actorMenuOpen = ref(false)
    const typeMenuOpen = ref(false)

    // Dropdown button rects for positioning
    const weightBtnRect = ref(null)
    const actionBtnRect = ref(null)
    const actorBtnRect = ref(null)
    const typeBtnRect = ref(null)

    const totalPages = computed(() => {
      return Math.ceil((store.total || 0) / store.pageSize)
    })

    const weightLabel = computed(() => {
      if (!store.filterWeight) return 'All Priorities'
      const labels = { '5': 'Important (5+)', '7': 'Critical (7+)', '9': 'Security (9)' }
      return labels[store.filterWeight] || 'All Priorities'
    })

    const actionLabel = computed(() => {
      if (!store.filterAction) return 'All Actions'
      return store.filterAction.charAt(0).toUpperCase() + store.filterAction.slice(1)
    })

    const actorLabel = computed(() => {
      if (!store.filterActor) return 'All Actors'
      const labels = { 'user': 'User', 'system': 'System', 'api_key': 'API Key', 'anonymous': 'Anonymous' }
      return labels[store.filterActor] || 'All Actors'
    })

    const typeLabel = computed(() => {
      if (!store.filterType) return 'All Types'
      return store.filterType.charAt(0).toUpperCase() + store.filterType.slice(1)
    })

    const togglePanel = () => {
      panelCollapsed.value = !panelCollapsed.value
      uiStore.setUIState('logs.list.collapsed', panelCollapsed.value)
    }

    const onSearch = (e) => {
      store.setFilter('search', e.target.value, client)
    }

    const toggleDropdown = (which, event) => {
      // Close all others
      weightMenuOpen.value = false
      actionMenuOpen.value = false
      actorMenuOpen.value = false
      typeMenuOpen.value = false

      const rect = event.currentTarget.getBoundingClientRect()

      if (which === 'weight') {
        weightMenuOpen.value = true
        weightBtnRect.value = rect
      } else if (which === 'action') {
        actionMenuOpen.value = true
        actionBtnRect.value = rect
      } else if (which === 'actor') {
        actorMenuOpen.value = true
        actorBtnRect.value = rect
      } else if (which === 'type') {
        typeMenuOpen.value = true
        typeBtnRect.value = rect
      }
    }

    const selectFilter = (key, value) => {
      store.setFilter(key, value, client)
      weightMenuOpen.value = false
      actionMenuOpen.value = false
      actorMenuOpen.value = false
      typeMenuOpen.value = false
    }

    const refresh = () => {
      store.load(client)
      store.loadStats(client)
    }

    const goFirstPage = () => { store.setPage(1, client) }
    const goPrevPage = () => { if (store.currentPage > 1) store.setPage(store.currentPage - 1, client) }
    const goNextPage = () => { if (store.currentPage < totalPages.value) store.setPage(store.currentPage + 1, client) }
    const goLastPage = () => { store.setPage(totalPages.value, client) }

    const getWeightInfo = (weight) => WEIGHT_INFO[weight] || {}
    const getActorInfo = (type) => ACTOR_INFO[type] || {}

    const closeDropdowns = () => {
      weightMenuOpen.value = false
      actionMenuOpen.value = false
      actorMenuOpen.value = false
      typeMenuOpen.value = false
    }

    onMounted(() => {
      store.load(client)
      store.loadStats(client)
      refreshIcons()
      document.addEventListener('click', closeDropdowns)
    })

    onUpdated(() => {
      refreshIcons()
    })

    onBeforeUnmount(() => {
      document.removeEventListener('click', closeDropdowns)
    })

    return {
      store,
      panelCollapsed,
      totalPages,
      weightLabel,
      actionLabel,
      actorLabel,
      typeLabel,
      weightMenuOpen,
      actionMenuOpen,
      actorMenuOpen,
      typeMenuOpen,
      weightBtnRect,
      actionBtnRect,
      actorBtnRect,
      typeBtnRect,
      togglePanel,
      onSearch,
      toggleDropdown,
      selectFilter,
      refresh,
      goFirstPage,
      goPrevPage,
      goNextPage,
      goLastPage,
      getWeightInfo,
      getActorInfo,
      formatBytes,
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
                  <span class="text-heading text-primary">Logs</span>
                  <span class="text-caption mono px-1.5 py-0.5 ml-2 badge-muted" style="border-radius: var(--radius-sm)">{{ store.total || 0 }}</span>
                </button>
              </header>
              <div class="panel-group-body" style="padding: 0; flex: 1; display: flex; flex-direction: column; min-height: 0">
                <div style="border: none; border-radius: 0; flex: 1; display: flex; flex-direction: column; min-height: 0">
                  <!-- Toolbar -->
                  <div class="card-header flex items-center justify-between" style="flex-shrink: 0">
                    <div class="input toolbar-search">
                      <i data-lucide="search" class="w-4 h-4 text-faint"></i>
                      <input type="text" placeholder="Search logs..." :value="store.searchQuery" @input="onSearch">
                    </div>
                    <div class="flex items-center justify-end gap-2">
                      <!-- Priority Filter -->
                      <div class="relative hide-mobile">
                        <button class="btn btn-secondary btn-sm flex items-center gap-1.5" style="padding: 4px 8px" @click.stop="toggleDropdown('weight', $event)">
                          <span class="text-caption">{{ weightLabel }}</span>
                          <i data-lucide="chevron-down" class="w-3.5 h-3.5" style="color:var(--text-3)"></i>
                        </button>
                        <div v-if="weightMenuOpen" class="dropdown fixed z-50" :style="{ top: (weightBtnRect?.bottom + 4) + 'px', left: weightBtnRect?.left + 'px', minWidth: '160px' }">
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('weight', '')">All Priorities</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('weight', '5')">Important (5+)</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('weight', '7')">Critical (7+)</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('weight', '9')">Security (9)</div>
                        </div>
                      </div>

                      <!-- Action Filter -->
                      <div class="relative hide-mobile">
                        <button class="btn btn-secondary btn-sm flex items-center gap-1.5" style="padding: 4px 8px" @click.stop="toggleDropdown('action', $event)">
                          <span class="text-caption">{{ actionLabel }}</span>
                          <i data-lucide="chevron-down" class="w-3.5 h-3.5" style="color:var(--text-3)"></i>
                        </button>
                        <div v-if="actionMenuOpen" class="dropdown fixed z-50" :style="{ top: (actionBtnRect?.bottom + 4) + 'px', left: actionBtnRect?.left + 'px', minWidth: '140px' }">
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('action', '')">All Actions</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('action', 'pageview')">Pageview</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('action', 'deploy')">Deploy</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('action', 'login')">Login</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('action', 'create')">Create</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('action', 'delete')">Delete</div>
                        </div>
                      </div>

                      <!-- Actor Filter -->
                      <div class="relative hide-mobile">
                        <button class="btn btn-secondary btn-sm flex items-center gap-1.5" style="padding: 4px 8px" @click.stop="toggleDropdown('actor', $event)">
                          <span class="text-caption">{{ actorLabel }}</span>
                          <i data-lucide="chevron-down" class="w-3.5 h-3.5" style="color:var(--text-3)"></i>
                        </button>
                        <div v-if="actorMenuOpen" class="dropdown fixed z-50" :style="{ top: (actorBtnRect?.bottom + 4) + 'px', left: actorBtnRect?.left + 'px', minWidth: '140px' }">
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('actor', '')">All Actors</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('actor', 'user')">User</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('actor', 'system')">System</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('actor', 'api_key')">API Key</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('actor', 'anonymous')">Anonymous</div>
                        </div>
                      </div>

                      <!-- Type Filter -->
                      <div class="relative hide-mobile">
                        <button class="btn btn-secondary btn-sm flex items-center gap-1.5" style="padding: 4px 8px" @click.stop="toggleDropdown('type', $event)">
                          <span class="text-caption">{{ typeLabel }}</span>
                          <i data-lucide="chevron-down" class="w-3.5 h-3.5" style="color:var(--text-3)"></i>
                        </button>
                        <div v-if="typeMenuOpen" class="dropdown fixed z-50" :style="{ top: (typeBtnRect?.bottom + 4) + 'px', left: typeBtnRect?.left + 'px', minWidth: '120px' }">
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('type', '')">All Types</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('type', 'page')">Page</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('type', 'app')">App</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('type', 'alias')">Alias</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('type', 'kv')">KV</div>
                          <div class="dropdown-item px-4 py-2 cursor-pointer text-caption" @click.stop="selectFilter('type', 'session')">Session</div>
                        </div>
                      </div>

                      <button class="btn btn-secondary btn-sm" style="padding: 4px 8px" title="Refresh" @click="refresh">
                        <i data-lucide="refresh-cw" class="w-3.5 h-3.5"></i>
                      </button>
                    </div>
                  </div>

                  <!-- Table -->
                  <div class="panel-scroll-area scroll-panel" style="flex: 1; overflow: auto; min-height: 0">
                    <!-- Loading -->
                    <div v-if="store.loading" class="flex items-center justify-center p-8">
                      <div class="text-caption text-muted">Loading...</div>
                    </div>
                    <!-- Empty state -->
                    <div v-else-if="store.entries.length === 0" class="flex flex-col items-center justify-center p-8 text-center" style="min-height: 200px">
                      <div class="icon-box mb-3" style="width:48px;height:48px;opacity:0.5">
                        <i data-lucide="activity" class="w-6 h-6"></i>
                      </div>
                      <div class="text-heading text-primary mb-1">{{ store.searchQuery ? 'No logs found' : 'No logs' }}</div>
                      <div class="text-caption text-muted">{{ store.searchQuery ? 'Try a different search' : 'Activity will appear here as it happens' }}</div>
                    </div>
                    <!-- Table -->
                    <div v-else class="table-container" style="overflow-x: visible">
                      <table style="width: 100%; min-width: 0">
                        <thead class="sticky" style="top: 0; background: var(--bg-1)">
                          <tr style="border-bottom: 1px solid var(--border-subtle)">
                            <th class="px-3 py-2 text-left text-micro text-muted">Activity</th>
                            <th class="px-3 py-2 text-left text-micro text-muted hide-mobile">Resource</th>
                            <th class="px-3 py-2 text-left text-micro text-muted hide-mobile">Actor</th>
                            <th class="px-3 py-2 text-left text-micro text-muted hide-mobile">Priority</th>
                            <th class="px-3 py-2 text-left text-micro text-muted hide-mobile"></th>
                            <th class="px-3 py-2 text-left text-micro text-muted hide-mobile">Time</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr v-for="entry in store.entries" :key="entry.id"
                              class="row"
                              style="border-bottom: 1px solid var(--border-subtle)">
                            <!-- Activity -->
                            <td class="px-3 py-2">
                              <div class="flex items-center gap-2" style="min-width: 0">
                                <div class="icon-box icon-box-sm" style="flex-shrink: 0" :style="{ borderColor: getWeightInfo(entry.weight).color }">
                                  <i :data-lucide="getActorInfo(entry.actor_type).icon || 'activity'" class="w-3.5 h-3.5" :style="{ color: getWeightInfo(entry.weight).color }"></i>
                                </div>
                                <div style="min-width: 0; overflow: hidden">
                                  <div class="text-label text-primary" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ entry.action }}</div>
                                  <div class="text-caption text-faint show-mobile">{{ entry.resource_type }} &middot; {{ formatRelativeTime(entry.timestamp) }}</div>
                                </div>
                              </div>
                            </td>
                            <!-- Resource -->
                            <td class="px-3 py-2 hide-mobile">
                              <div style="min-width: 0; overflow: hidden">
                                <span class="badge badge-muted">{{ entry.resource_type }}</span>
                                <div v-if="entry.resource_id" class="text-caption mono text-muted mt-0.5" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ entry.resource_id }}</div>
                              </div>
                            </td>
                            <!-- Actor -->
                            <td class="px-3 py-2 hide-mobile">
                              <span class="text-caption text-muted">{{ entry.actor_id || getActorInfo(entry.actor_type).label }}</span>
                            </td>
                            <!-- Priority -->
                            <td class="px-3 py-2 hide-mobile">
                              <span class="text-caption" :style="{ color: getWeightInfo(entry.weight).color }">{{ getWeightInfo(entry.weight).label }}</span>
                            </td>
                            <!-- Result dot -->
                            <td class="px-3 py-2 hide-mobile">
                              <span class="flex items-center justify-center" :title="entry.result">
                                <span class="status-dot" :class="entry.result === 'success' ? 'status-dot-success' : 'status-dot-error'"></span>
                              </span>
                            </td>
                            <!-- Time -->
                            <td class="px-3 py-2 hide-mobile">
                              <span class="text-caption text-muted">{{ formatRelativeTime(entry.timestamp) }}</span>
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </div>

                  <!-- Footer -->
                  <div style="flex-shrink: 0">
                    <div class="card-footer flex items-center justify-between" style="border-radius: 0">
                      <span class="text-caption text-muted">
                        Showing {{ store.showing }} of {{ store.total }}
                        <template v-if="store.stats.size_estimate_bytes > 0"> &middot; {{ formatBytes(store.stats.size_estimate_bytes) }}</template>
                      </span>
                      <div v-if="totalPages > 1" class="flex items-center gap-2">
                        <button class="btn btn-secondary btn-sm" :disabled="store.currentPage === 1" title="First page" @click="goFirstPage">
                          <i data-lucide="chevrons-left" class="w-3.5 h-3.5"></i>
                        </button>
                        <button class="btn btn-secondary btn-sm" :disabled="store.currentPage === 1" title="Previous" @click="goPrevPage">
                          <i data-lucide="chevron-left" class="w-3.5 h-3.5"></i>
                        </button>
                        <span class="text-caption text-muted px-2">Page {{ store.currentPage }} of {{ totalPages }}</span>
                        <button class="btn btn-secondary btn-sm" :disabled="store.currentPage === totalPages" title="Next" @click="goNextPage">
                          <i data-lucide="chevron-right" class="w-3.5 h-3.5"></i>
                        </button>
                        <button class="btn btn-secondary btn-sm" :disabled="store.currentPage === totalPages" title="Last page" @click="goLastPage">
                          <i data-lucide="chevrons-right" class="w-3.5 h-3.5"></i>
                        </button>
                      </div>
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
