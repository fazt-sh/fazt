import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useUIStore } from './ui.js'

export const useAliasesStore = defineStore('aliases', () => {
  const items = ref([])
  const loading = ref(false)

  async function load(client) {
    loading.value = true
    try {
      items.value = await client.aliases.list() || []
    } catch (err) {
      const ui = useUIStore()
      ui.notify({ type: 'error', message: 'Failed to load aliases' })
    } finally {
      loading.value = false
    }
  }

  async function create(client, subdomain, type, options = {}) {
    try {
      await client.aliases.create(subdomain, type, options)
      const ui = useUIStore()
      ui.notify({ type: 'success', message: `Alias "${subdomain}" created` })
      await load(client)
      return true
    } catch (err) {
      const ui = useUIStore()
      ui.notify({ type: 'error', message: err.message || 'Failed to create alias' })
      throw err
    }
  }

  async function update(client, subdomain, data) {
    try {
      await client.aliases.update(subdomain, data)
      const ui = useUIStore()
      ui.notify({ type: 'success', message: `Alias "${subdomain}" updated` })
      await load(client)
      return true
    } catch (err) {
      const ui = useUIStore()
      ui.notify({ type: 'error', message: err.message || 'Failed to update alias' })
      throw err
    }
  }

  async function remove(client, subdomain) {
    try {
      await client.aliases.delete(subdomain)
      const ui = useUIStore()
      ui.notify({ type: 'success', message: `Alias "${subdomain}" deleted` })
      await load(client)
      return true
    } catch (err) {
      const ui = useUIStore()
      ui.notify({ type: 'error', message: err.message || 'Failed to delete alias' })
      throw err
    }
  }

  return { items, loading, load, create, update, remove }
})
