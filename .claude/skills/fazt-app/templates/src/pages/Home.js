// Home Page Template
import { defineComponent, onMounted } from 'vue'
import { useAppStore } from '../stores/index.js'

export const HomePage = defineComponent({
  name: 'HomePage',

  setup() {
    const store = useAppStore()

    onMounted(() => {
      store.init()
    })

    return { store }
  },

  template: `
    <div class="h-full flex flex-col">
      <!-- Header -->
      <header class="flex-none bg-white dark:bg-neutral-900 border-b border-neutral-200 dark:border-neutral-800 px-4 py-4 safe-top">
        <div class="max-w-4xl mx-auto flex items-center justify-between">
          <h1 class="text-xl font-semibold">App Name</h1>
          <router-link
            to="/settings"
            class="p-2 rounded-lg hover:bg-neutral-100 dark:hover:bg-neutral-800"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
            </svg>
          </router-link>
        </div>
      </header>

      <!-- Content -->
      <main class="flex-1 overflow-y-auto">
        <div class="max-w-4xl mx-auto px-4 py-6 pb-24">
          <!-- Loading State -->
          <div v-if="store.initializing" class="flex items-center justify-center py-12">
            <div class="spinner w-8 h-8 text-blue-500"></div>
          </div>

          <!-- Error State -->
          <div v-else-if="store.error" class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
            <p class="text-red-600 dark:text-red-400">{{ store.error }}</p>
            <button
              @click="store.init()"
              class="mt-2 text-sm text-red-600 dark:text-red-400 underline"
            >
              Retry
            </button>
          </div>

          <!-- Empty State -->
          <div v-else-if="store.items.length === 0" class="text-center py-12">
            <p class="text-neutral-500 dark:text-neutral-400">No items yet</p>
            <p class="text-sm text-neutral-400 dark:text-neutral-500 mt-1">
              Click the + button to add your first item
            </p>
          </div>

          <!-- Items List -->
          <div v-else class="space-y-4">
            <div
              v-for="item in store.sortedItems"
              :key="item.id"
              class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg p-4"
            >
              <h3 class="font-medium">{{ item.name }}</h3>
              <p v-if="item.description" class="text-sm text-neutral-500 mt-1">
                {{ item.description }}
              </p>
            </div>
          </div>
        </div>
      </main>

      <!-- FAB -->
      <button
        class="fixed bottom-6 right-6 w-14 h-14 bg-blue-500 hover:bg-blue-600 text-white rounded-full shadow-lg flex items-center justify-center touch-feedback safe-bottom"
        @click="showAddModal = true"
      >
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
        </svg>
      </button>
    </div>
  `
})
