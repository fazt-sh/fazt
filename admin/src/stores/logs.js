import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useUIStore } from './ui.js'

export const useLogsStore = defineStore('logs', () => {
  const entries = ref([])
  const total = ref(0)
  const showing = ref(0)
  const loading = ref(false)

  // Stats
  const stats = ref({
    total_count: 0,
    count_by_weight: {},
    oldest_entry: null,
    newest_entry: null,
    size_estimate_bytes: 0
  })

  // Filters
  const filterWeight = ref('')
  const filterAction = ref('')
  const filterActor = ref('')
  const filterType = ref('')
  const searchQuery = ref('')
  const currentPage = ref(1)
  const pageSize = 50

  async function load(client) {
    loading.value = true
    const params = {
      limit: pageSize,
      offset: (currentPage.value - 1) * pageSize
    }
    if (filterWeight.value) params.min_weight = filterWeight.value
    if (filterAction.value) params.action = filterAction.value
    if (filterActor.value) params.actor_type = filterActor.value
    if (filterType.value) params.type = filterType.value
    if (searchQuery.value) params.action = searchQuery.value

    try {
      const data = await client.logs.list(params) || { entries: [], total: 0, showing: 0 }
      entries.value = data.entries || []
      total.value = data.total || 0
      showing.value = data.showing || 0
    } catch (err) {
      const ui = useUIStore()
      ui.notify({ type: 'error', message: 'Failed to load logs' })
    } finally {
      loading.value = false
    }
  }

  async function loadStats(client) {
    try {
      const data = await client.logs.stats({}) || {}
      stats.value = data
    } catch (err) {
      console.warn('Failed to load activity stats:', err.message)
    }
  }

  function setFilter(key, value, client) {
    if (key === 'weight') filterWeight.value = value
    else if (key === 'action') filterAction.value = value
    else if (key === 'actor') filterActor.value = value
    else if (key === 'type') filterType.value = value
    else if (key === 'search') searchQuery.value = value
    currentPage.value = 1
    load(client)
  }

  function setPage(page, client) {
    currentPage.value = page
    load(client)
  }

  return {
    entries, total, showing, loading, stats,
    filterWeight, filterAction, filterActor, filterType, searchQuery,
    currentPage, pageSize,
    load, loadStats, setFilter, setPage
  }
})
