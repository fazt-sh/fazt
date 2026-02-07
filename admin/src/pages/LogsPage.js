import { computed, onMounted } from 'vue'
import { useLogsStore } from '../stores/logs.js'
import { client } from '../client.js'
import { useIcons } from '../lib/useIcons.js'
import { usePanel } from '../lib/usePanel.js'
import { formatBytes, formatRelativeTime } from '../lib/format.js'
import FPanel from '../components/FPanel.js'
import FTable from '../components/FTable.js'
import FToolbar from '../components/FToolbar.js'
import FilterDropdown from '../components/FilterDropdown.js'
import FPagination from '../components/FPagination.js'

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
  components: { FPanel, FTable, FToolbar, FilterDropdown, FPagination },
  setup() {
    useIcons()
    const store = useLogsStore()
    const panel = usePanel('logs.list.collapsed', false)

    const totalPages = computed(() => Math.ceil((store.total || 0) / store.pageSize))

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

    const weightOptions = [
      { value: '', label: 'All Priorities' },
      { value: '5', label: 'Important (5+)' },
      { value: '7', label: 'Critical (7+)' },
      { value: '9', label: 'Security (9)' },
    ]

    const actionOptions = [
      { value: '', label: 'All Actions' },
      { value: 'pageview', label: 'Pageview' },
      { value: 'deploy', label: 'Deploy' },
      { value: 'login', label: 'Login' },
      { value: 'create', label: 'Create' },
      { value: 'delete', label: 'Delete' },
    ]

    const actorOptions = [
      { value: '', label: 'All Actors' },
      { value: 'user', label: 'User' },
      { value: 'system', label: 'System' },
      { value: 'api_key', label: 'API Key' },
      { value: 'anonymous', label: 'Anonymous' },
    ]

    const typeOptions = [
      { value: '', label: 'All Types' },
      { value: 'page', label: 'Page' },
      { value: 'app', label: 'App' },
      { value: 'alias', label: 'Alias' },
      { value: 'kv', label: 'KV' },
      { value: 'session', label: 'Session' },
    ]

    const columns = [
      { key: 'activity', label: 'Activity' },
      { key: 'resource', label: 'Resource', hideOnMobile: true },
      { key: 'actor', label: 'Actor', hideOnMobile: true },
      { key: 'priority', label: 'Priority', hideOnMobile: true },
      { key: 'result', label: '', hideOnMobile: true },
      { key: 'time', label: 'Time', hideOnMobile: true },
    ]

    const onSearch = (val) => { store.setFilter('search', val, client) }
    const selectFilter = (key, value) => { store.setFilter(key, value, client) }
    const refresh = () => { store.load(client); store.loadStats(client) }
    const onPageChange = (page) => { store.setPage(page, client) }

    const getWeightInfo = (weight) => WEIGHT_INFO[weight] || {}
    const getActorInfo = (type) => ACTOR_INFO[type] || {}

    onMounted(() => {
      store.load(client)
      store.loadStats(client)
    })

    return {
      store, panel, totalPages,
      weightLabel, actionLabel, actorLabel, typeLabel,
      weightOptions, actionOptions, actorOptions, typeOptions,
      columns, onSearch, selectFilter, refresh, onPageChange,
      getWeightInfo, getActorInfo, formatBytes, formatRelativeTime
    }
  },
  template: `
    <div class="design-system-page">
      <div class="content-container">
        <div class="content-scroll">

          <FPanel title="Logs" :count="store.total || 0" mode="fill"
                  :collapsed="panel.collapsed" @update:collapsed="panel.toggle">
            <template #toolbar>
              <FToolbar :model-value="store.searchQuery" search-placeholder="Search logs..."
                        @update:model-value="onSearch">
                <template #filters>
                  <FilterDropdown :label="weightLabel" :options="weightOptions" :model-value="store.filterWeight"
                                  @update:model-value="v => selectFilter('weight', v)" />
                  <FilterDropdown :label="actionLabel" :options="actionOptions" :model-value="store.filterAction"
                                  @update:model-value="v => selectFilter('action', v)" />
                  <FilterDropdown :label="actorLabel" :options="actorOptions" :model-value="store.filterActor"
                                  @update:model-value="v => selectFilter('actor', v)" />
                  <FilterDropdown :label="typeLabel" :options="typeOptions" :model-value="store.filterType"
                                  @update:model-value="v => selectFilter('type', v)" />
                </template>
                <template #actions>
                  <button class="btn btn-secondary btn-sm" style="padding: 4px 8px" title="Refresh" @click="refresh">
                    <i data-lucide="refresh-cw" class="w-3.5 h-3.5"></i>
                  </button>
                </template>
              </FToolbar>
            </template>

            <FTable :columns="columns" :rows="store.entries" row-key="id" :clickable="false"
                    :loading="store.loading"
                    empty-icon="activity"
                    :empty-title="store.searchQuery ? 'No logs found' : 'No logs'"
                    :empty-message="store.searchQuery ? 'Try a different search' : 'Activity will appear here as it happens'">
              <template #row="{ row }">
                <!-- Activity -->
                <td class="px-3 py-2">
                  <div class="flex items-center gap-2" style="min-width: 0">
                    <div class="icon-box icon-box-sm" style="flex-shrink: 0" :style="{ borderColor: getWeightInfo(row.weight).color }">
                      <i :data-lucide="getActorInfo(row.actor_type).icon || 'activity'" class="w-3.5 h-3.5" :style="{ color: getWeightInfo(row.weight).color }"></i>
                    </div>
                    <div style="min-width: 0; overflow: hidden">
                      <div class="text-label text-primary truncate">{{ row.action }}</div>
                      <div class="text-caption text-faint show-mobile">{{ row.resource_type }} &middot; {{ formatRelativeTime(row.timestamp) }}</div>
                    </div>
                  </div>
                </td>
                <!-- Resource -->
                <td class="px-3 py-2 hide-mobile">
                  <div style="min-width: 0; overflow: hidden">
                    <span class="badge badge-muted">{{ row.resource_type }}</span>
                    <div v-if="row.resource_id" class="text-caption mono text-muted mt-0.5 truncate">{{ row.resource_id }}</div>
                  </div>
                </td>
                <!-- Actor -->
                <td class="px-3 py-2 hide-mobile">
                  <span class="text-caption text-muted">{{ row.actor_id || getActorInfo(row.actor_type).label }}</span>
                </td>
                <!-- Priority -->
                <td class="px-3 py-2 hide-mobile">
                  <span class="text-caption" :style="{ color: getWeightInfo(row.weight).color }">{{ getWeightInfo(row.weight).label }}</span>
                </td>
                <!-- Result dot -->
                <td class="px-3 py-2 hide-mobile">
                  <span class="flex items-center justify-center" :title="row.result">
                    <span class="status-dot" :class="row.result === 'success' ? 'status-dot-success' : 'status-dot-error'"></span>
                  </span>
                </td>
                <!-- Time -->
                <td class="px-3 py-2 hide-mobile">
                  <span class="text-caption text-muted">{{ formatRelativeTime(row.timestamp) }}</span>
                </td>
              </template>
            </FTable>

            <template #footer>
              <div class="card-footer flex items-center justify-between" style="border-radius: 0">
                <span class="text-caption text-muted">
                  Showing {{ store.showing }} of {{ store.total }}
                  <template v-if="store.stats.size_estimate_bytes > 0"> &middot; {{ formatBytes(store.stats.size_estimate_bytes) }}</template>
                </span>
                <FPagination :current-page="store.currentPage" :total-pages="totalPages"
                             @page-change="onPageChange" />
              </div>
            </template>
          </FPanel>

        </div>
      </div>
    </div>
  `
}
