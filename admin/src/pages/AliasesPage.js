import { ref, computed, onMounted } from 'vue'
import { useAliasesStore } from '../stores/aliases.js'
import { useAppsStore } from '../stores/apps.js'
import { useUIStore } from '../stores/ui.js'
import { client } from '../client.js'
import { useIcons } from '../lib/useIcons.js'
import { usePanel } from '../lib/usePanel.js'
import { formatRelativeTime } from '../lib/format.js'
import FPanel from '../components/FPanel.js'
import FTable from '../components/FTable.js'
import FToolbar from '../components/FToolbar.js'

export default {
  name: 'AliasesPage',
  components: { FPanel, FTable, FToolbar },
  setup() {
    useIcons()
    const store = useAliasesStore()
    const appsStore = useAppsStore()
    const uiStore = useUIStore()
    const panel = usePanel('aliases.list.collapsed', false)
    const searchQuery = ref('')

    const filteredAliases = computed(() => {
      if (!searchQuery.value) return store.items
      const query = searchQuery.value.toLowerCase()
      return store.items.filter(alias => {
        const subdomain = (alias.subdomain || '').toLowerCase()
        const type = (alias.type || '').toLowerCase()
        return subdomain.includes(query) || type.includes(query)
      })
    })

    const columns = [
      { key: 'subdomain', label: 'Subdomain' },
      { key: 'type', label: 'Type', hideOnMobile: true },
      { key: 'target', label: 'Target', hideOnMobile: true },
      { key: 'updated_at', label: 'Updated', hideOnMobile: true },
      { key: 'status', label: 'Status', hideOnMobile: true },
    ]

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

    const openNewAliasModal = () => { uiStore.createAliasModalOpen = true }

    const openEditAliasModal = (alias) => {
      uiStore.editingAlias = alias
      uiStore.editAliasModalOpen = true
    }

    onMounted(() => {
      store.load(client)
      appsStore.load(client)
    })

    return {
      store, panel, searchQuery, filteredAliases, columns,
      getTypeBadge, getAliasIcon, getTarget,
      openNewAliasModal, openEditAliasModal, formatRelativeTime
    }
  },
  template: `
    <div class="design-system-page">
      <div class="content-container">
        <div class="content-scroll">

          <FPanel title="Aliases" :count="store.items.length" mode="fill"
                  :collapsed="panel.collapsed" @update:collapsed="panel.toggle">
            <template #toolbar>
              <FToolbar v-model="searchQuery">
                <template #actions>
                  <button class="btn btn-sm btn-primary toolbar-btn" title="New Alias" @click="openNewAliasModal">
                    <i data-lucide="plus" class="w-4 h-4"></i>
                  </button>
                </template>
              </FToolbar>
            </template>

            <FTable :columns="columns" :rows="filteredAliases" row-key="subdomain"
                    :empty-icon="searchQuery ? 'search-x' : 'link'"
                    :empty-title="searchQuery ? 'No aliases found' : 'No aliases yet'"
                    :empty-message="searchQuery ? 'Try a different search term' : 'Create your first alias to get started'"
                    @row-click="openEditAliasModal($event)">
              <template #cell-subdomain="{ row }">
                <div class="flex items-center gap-2" style="min-width: 0">
                  <span class="status-dot show-mobile" :class="row.type === 'reserved' ? '' : 'status-dot-success pulse'" style="flex-shrink: 0"></span>
                  <div class="icon-box icon-box-sm" style="flex-shrink: 0">
                    <i :data-lucide="getAliasIcon(row.type)" class="w-3.5 h-3.5"></i>
                  </div>
                  <div style="min-width: 0; overflow: hidden">
                    <div class="text-label text-primary truncate">{{ row.subdomain }}</div>
                    <div class="text-caption mono text-faint show-mobile">{{ row.type }}</div>
                  </div>
                </div>
              </template>
              <template #cell-type="{ row }">
                <span class="badge" :class="getTypeBadge(row.type)">{{ row.type }}</span>
              </template>
              <template #cell-target="{ row }">
                <span class="text-caption mono text-muted truncate" style="display: block; max-width: 180px">{{ getTarget(row) }}</span>
              </template>
              <template #cell-updated_at="{ row }">
                <span class="text-caption text-muted">{{ formatRelativeTime(row.updated_at) }}</span>
              </template>
              <template #cell-status="{ row }">
                <span class="flex items-center gap-1 text-caption" :class="row.type === 'reserved' ? 'text-muted' : 'text-success'">
                  <span class="status-dot" :class="row.type === 'reserved' ? '' : 'status-dot-success pulse'"></span>
                  {{ row.type === 'reserved' ? 'Reserved' : 'Active' }}
                </span>
              </template>
            </FTable>

            <template #footer>
              <div class="card-footer flex items-center justify-between" style="border-radius: 0">
                <span class="text-caption text-muted">{{ store.items.length }} alias{{ store.items.length === 1 ? '' : 'es' }}</span>
              </div>
            </template>
          </FPanel>

        </div>
      </div>
    </div>
  `
}
