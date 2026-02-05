import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useUIStore } from './ui.js'

export const useAppsStore = defineStore('apps', () => {
  const items = ref([])
  const loading = ref(false)
  const currentApp = ref({ id: null, title: null, files: [] })
  const currentAppLoading = ref(false)

  async function load(client) {
    loading.value = true
    try {
      items.value = await client.apps.list() || []
    } catch (err) {
      const ui = useUIStore()
      ui.notify({ type: 'error', message: 'Failed to load apps' })
    } finally {
      loading.value = false
    }
  }

  async function loadDetail(client, id) {
    currentAppLoading.value = true
    try {
      const [appData, filesData] = await Promise.all([
        client.apps.get(id),
        client.apps.files(id).catch(() => [])
      ])
      currentApp.value = { ...appData, files: filesData || [] }
    } catch (err) {
      const ui = useUIStore()
      ui.notify({ type: 'error', message: 'Failed to load app detail' })
    } finally {
      currentAppLoading.value = false
    }
  }

  async function create(client, name, template = 'minimal') {
    try {
      const data = await client.apps.create(name, template)
      const ui = useUIStore()
      ui.notify({ type: 'success', message: `App "${name}" created` })
      await load(client)
      return data
    } catch (err) {
      const ui = useUIStore()
      ui.notify({ type: 'error', message: err.message || 'Failed to create app' })
      throw err
    }
  }

  async function remove(client, id) {
    try {
      await client.apps.delete(id)
      items.value = items.value.filter(app => app.id !== id && app.title !== id)
      const ui = useUIStore()
      ui.notify({ type: 'success', message: 'App deleted' })
      return true
    } catch (err) {
      const ui = useUIStore()
      ui.notify({ type: 'error', message: err.message })
      return false
    }
  }

  return { items, loading, currentApp, currentAppLoading, load, loadDetail, create, remove }
})
