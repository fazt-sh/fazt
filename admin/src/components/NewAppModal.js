import { ref, onMounted, onUpdated } from 'vue'
import { useUIStore } from '../stores/ui.js'
import { useAppsStore } from '../stores/apps.js'
import { client } from '../client.js'
import { refreshIcons } from '../lib/icons.js'

export default {
  name: 'NewAppModal',
  setup() {
    const ui = useUIStore()
    const apps = useAppsStore()

    const appName = ref('')
    const appTemplate = ref('minimal')

    async function create() {
      if (!appName.value.trim()) return
      try {
        await apps.create(client, appName.value, appTemplate.value)
        ui.newAppModalOpen = false
        appName.value = ''
        appTemplate.value = 'minimal'
      } catch (err) {
        console.error('Failed to create app:', err)
      }
    }

    function cancel() {
      ui.newAppModalOpen = false
      appName.value = ''
      appTemplate.value = 'minimal'
    }

    onMounted(() => refreshIcons())
    onUpdated(() => refreshIcons())

    return { ui, appName, appTemplate, create, cancel }
  },
  template: `
    <div v-if="ui.newAppModalOpen" class="fixed inset-0 z-50 flex items-center justify-center modal-backdrop" @click="cancel">
      <div class="modal w-full max-w-md p-6 relative" @click.stop>
        <div class="text-center mb-6">
          <div class="w-12 h-12 mx-auto mb-4 flex items-center justify-center" style="background:var(--accent-soft);border-radius:var(--radius-md)">
            <i data-lucide="rocket" class="w-6 h-6" style="color:var(--accent)"></i>
          </div>
          <h2 class="text-title" style="color:var(--text-1)">Create New App</h2>
          <p class="text-caption mt-1" style="color:var(--text-3)">Create an app from a template</p>
        </div>
        <div class="space-y-4 mb-6">
          <div>
            <label class="text-micro text-muted block mb-2">APP NAME</label>
            <div class="input">
              <i data-lucide="box" class="w-4 h-4 text-faint"></i>
              <input type="text" v-model="appName" placeholder="my-app" class="text-body" style="flex: 1">
            </div>
            <div class="text-caption text-muted mt-1 px-1">Lowercase letters, numbers, and dashes only</div>
          </div>
          <div>
            <label class="text-micro text-muted block mb-2">TEMPLATE</label>
            <div class="grid grid-cols-2 gap-2">
              <button type="button" class="p-3 text-left" @click="appTemplate = 'minimal'"
                      :style="'background:var(--bg-2);border:2px solid ' + (appTemplate === 'minimal' ? 'var(--accent)' : 'var(--border-subtle)') + ';border-radius:var(--radius-md);cursor:pointer'">
                <div class="flex items-center gap-2 mb-1">
                  <i data-lucide="file" class="w-4 h-4" :style="appTemplate === 'minimal' ? 'color:var(--accent)' : 'color:var(--text-3)'"></i>
                  <span class="text-label text-primary">Minimal</span>
                </div>
                <div class="text-caption text-muted">Static site</div>
              </button>
              <button type="button" class="p-3 text-left" @click="appTemplate = 'spa'"
                      :style="'background:var(--bg-2);border:2px solid ' + (appTemplate === 'spa' ? 'var(--accent)' : 'var(--border-subtle)') + ';border-radius:var(--radius-md);cursor:pointer'">
                <div class="flex items-center gap-2 mb-1">
                  <i data-lucide="layout" class="w-4 h-4" :style="appTemplate === 'spa' ? 'color:var(--accent)' : 'color:var(--text-3)'"></i>
                  <span class="text-label text-primary">SPA</span>
                </div>
                <div class="text-caption text-muted">With routing</div>
              </button>
              <button type="button" class="p-3 text-left" @click="appTemplate = 'api'"
                      :style="'background:var(--bg-2);border:2px solid ' + (appTemplate === 'api' ? 'var(--accent)' : 'var(--border-subtle)') + ';border-radius:var(--radius-md);cursor:pointer'">
                <div class="flex items-center gap-2 mb-1">
                  <i data-lucide="code" class="w-4 h-4" :style="appTemplate === 'api' ? 'color:var(--accent)' : 'color:var(--text-3)'"></i>
                  <span class="text-label text-primary">API</span>
                </div>
                <div class="text-caption text-muted">Serverless only</div>
              </button>
              <button type="button" class="p-3 text-left" @click="appTemplate = 'full'"
                      :style="'background:var(--bg-2);border:2px solid ' + (appTemplate === 'full' ? 'var(--accent)' : 'var(--border-subtle)') + ';border-radius:var(--radius-md);cursor:pointer'">
                <div class="flex items-center gap-2 mb-1">
                  <i data-lucide="layers" class="w-4 h-4" :style="appTemplate === 'full' ? 'color:var(--accent)' : 'color:var(--text-3)'"></i>
                  <span class="text-label text-primary">Full Stack</span>
                </div>
                <div class="text-caption text-muted">Static + API</div>
              </button>
            </div>
          </div>
        </div>
        <div class="flex gap-2">
          <button class="btn btn-secondary flex-1 text-label" style="padding:8px 16px" @click="cancel">Cancel</button>
          <button class="btn btn-primary flex-1 text-label" style="padding:8px 16px" @click="create">Create App</button>
        </div>
        <button class="absolute top-4 right-4 p-1 btn-ghost" style="color:var(--text-3);border-radius:var(--radius-sm)" @click="cancel">
          <i data-lucide="x" class="w-5 h-5"></i>
        </button>
      </div>
    </div>
  `
}
