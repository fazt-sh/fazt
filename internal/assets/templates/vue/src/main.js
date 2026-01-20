import { createApp, ref } from 'vue'

const App = {
  setup() {
    const count = ref(0)
    return { count }
  },
  template: `
    <div class="flex items-center justify-center min-h-screen">
      <div class="text-center">
        <h1 class="text-4xl font-bold text-gray-800 mb-4">{{.Name}}</h1>
        <p class="text-gray-600 mb-8">Your Vue app is ready.</p>
        <div class="space-x-2">
          <button
            @click="count++"
            class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            Count: {{ "{{" }} count {{ "}}" }}
          </button>
        </div>
      </div>
    </div>
  `
}

createApp(App).mount('#app')
