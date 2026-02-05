import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useUIStore } from './ui.js'

export const useHealthStore = defineStore('health', () => {
  const data = ref({
    status: null,
    uptime_seconds: 0,
    version: null,
    mode: null,
    memory: null,
    database: null,
    runtime: null
  })
  const loading = ref(false)

  async function load(client) {
    loading.value = true
    try {
      data.value = await client.system.health() || {}
    } catch (err) {
      const ui = useUIStore()
      ui.notify({ type: 'error', message: 'Failed to load system health' })
    } finally {
      loading.value = false
    }
  }

  return { data, loading, load }
})
