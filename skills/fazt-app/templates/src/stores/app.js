// App Store - Global app state
// Example Pinia store pattern

import { defineStore } from 'pinia'
import { api } from '../lib/api.js'

export const useAppStore = defineStore('app', {
  // State: reactive data
  state: () => ({
    // Loading states
    initializing: true,
    loading: false,

    // Error handling
    error: null,

    // App data
    items: [],

    // UI state
    sidebarOpen: false
  }),

  // Getters: computed properties
  getters: {
    itemCount: (state) => state.items.length,

    hasError: (state) => state.error !== null,

    // Filtered/sorted data
    sortedItems: (state) => {
      return [...state.items].sort((a, b) => b.created - a.created)
    }
  },

  // Actions: methods that can be async
  actions: {
    // Initialize app data
    async init() {
      this.initializing = true
      this.error = null

      try {
        await this.loadItems()
      } catch (e) {
        this.error = e.message
        console.error('[store] init error:', e)
      } finally {
        this.initializing = false
      }
    },

    // Load items from API
    async loadItems() {
      this.loading = true
      try {
        const data = await api.get('/api/items')
        this.items = data.items || []
      } finally {
        this.loading = false
      }
    },

    // Create item
    async createItem(itemData) {
      const created = await api.post('/api/items', itemData)
      this.items.push(created)
      return created
    },

    // Update item
    async updateItem(id, updates) {
      const updated = await api.put(`/api/items/${id}`, updates)
      const index = this.items.findIndex(i => i.id === id)
      if (index !== -1) {
        this.items[index] = { ...this.items[index], ...updated }
      }
      return updated
    },

    // Delete item
    async deleteItem(id) {
      await api.delete(`/api/items/${id}`)
      this.items = this.items.filter(i => i.id !== id)
    },

    // Clear error
    clearError() {
      this.error = null
    },

    // Toggle sidebar
    toggleSidebar() {
      this.sidebarOpen = !this.sidebarOpen
    }
  }
})
