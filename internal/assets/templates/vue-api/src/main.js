import { createApp, ref, onMounted } from 'vue'
import { api } from './lib/api.js'

const App = {
  setup() {
    const items = ref([])
    const newItem = ref('')
    const loading = ref(true)
    const error = ref(null)

    async function loadItems() {
      try {
        loading.value = true
        const data = await api.get('/api/items')
        items.value = data.items || []
      } catch (e) {
        error.value = e.message
      } finally {
        loading.value = false
      }
    }

    async function addItem() {
      if (!newItem.value.trim()) return
      try {
        await api.post('/api/items', { name: newItem.value })
        newItem.value = ''
        await loadItems()
      } catch (e) {
        error.value = e.message
      }
    }

    async function deleteItem(id) {
      try {
        await api.delete('/api/items/' + id)
        await loadItems()
      } catch (e) {
        error.value = e.message
      }
    }

    onMounted(loadItems)

    return { items, newItem, loading, error, addItem, deleteItem }
  },
  template: `
    <div class="max-w-md mx-auto p-6">
      <h1 class="text-2xl font-bold text-gray-800 mb-6">{{.Name}}</h1>

      <div v-if="error" class="bg-red-100 text-red-700 p-3 rounded mb-4">
        {{ "{{" }} error {{ "}}" }}
      </div>

      <div class="flex gap-2 mb-6">
        <input
          v-model="newItem"
          @keyup.enter="addItem"
          placeholder="Add item..."
          class="flex-1 px-3 py-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <button
          @click="addItem"
          class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          Add
        </button>
      </div>

      <div v-if="loading" class="text-gray-500">Loading...</div>

      <ul v-else class="space-y-2">
        <li
          v-for="item in items"
          :key="item.id"
          class="flex items-center justify-between p-3 bg-white rounded shadow"
        >
          <span>{{ "{{" }} item.name {{ "}}" }}</span>
          <button
            @click="deleteItem(item.id)"
            class="text-red-500 hover:text-red-700"
          >
            Delete
          </button>
        </li>
        <li v-if="items.length === 0" class="text-gray-500 text-center py-4">
          No items yet. Add one above!
        </li>
      </ul>
    </div>
  `
}

createApp(App).mount('#app')
