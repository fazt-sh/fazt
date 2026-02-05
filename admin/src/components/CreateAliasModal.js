import { ref, onMounted, onUpdated } from 'vue'
import { useUIStore } from '../stores/ui.js'
import { useAliasesStore } from '../stores/aliases.js'
import { client } from '../client.js'
import { refreshIcons } from '../lib/icons.js'

export default {
  name: 'CreateAliasModal',
  setup() {
    const ui = useUIStore()
    const aliases = useAliasesStore()

    const subdomain = ref('')
    const aliasType = ref('proxy')
    const target = ref('')

    async function create() {
      if (!subdomain.value.trim()) return
      try {
        const options = aliasType.value === 'app' ? { app_id: target.value } : {}
        await aliases.create(client, subdomain.value, aliasType.value, options)
        ui.createAliasModalOpen = false
        subdomain.value = ''
        aliasType.value = 'proxy'
        target.value = ''
      } catch (err) {
        console.error('Failed to create alias:', err)
      }
    }

    function cancel() {
      ui.createAliasModalOpen = false
      subdomain.value = ''
      aliasType.value = 'proxy'
      target.value = ''
    }

    onMounted(() => refreshIcons())
    onUpdated(() => refreshIcons())

    return { ui, subdomain, aliasType, target, create, cancel }
  },
  template: `
    <div v-if="ui.createAliasModalOpen" class="fixed inset-0 z-50 flex items-center justify-center modal-backdrop" @click="cancel">
      <div class="modal w-full max-w-md p-6 relative" @click.stop>
        <div class="text-center mb-6">
          <div class="w-12 h-12 mx-auto mb-4 flex items-center justify-center" style="background:var(--accent-soft);border-radius:var(--radius-md)">
            <i data-lucide="link" class="w-6 h-6" style="color:var(--accent)"></i>
          </div>
          <h2 class="text-title" style="color:var(--text-1)">Create Alias</h2>
          <p class="text-caption mt-1" style="color:var(--text-3)">Point a subdomain to an app or URL</p>
        </div>
        <div class="space-y-4 mb-6">
          <div>
            <label class="text-micro text-muted block mb-2">SUBDOMAIN</label>
            <div class="input">
              <i data-lucide="at-sign" class="w-4 h-4 text-faint"></i>
              <input type="text" v-model="subdomain" placeholder="my-alias" class="text-body" style="flex: 1">
            </div>
          </div>
          <div>
            <label class="text-micro text-muted block mb-2">TYPE</label>
            <div class="flex gap-2">
              <button type="button" class="flex-1 p-2 text-center" @click="aliasType = 'proxy'"
                      :style="'border:2px solid ' + (aliasType === 'proxy' ? 'var(--accent)' : 'var(--border-subtle)') + ';border-radius:var(--radius-md);cursor:pointer;background:' + (aliasType === 'proxy' ? 'var(--accent-soft)' : 'var(--bg-2)')">
                <i data-lucide="arrow-right" class="w-4 h-4 mx-auto mb-1" :style="aliasType === 'proxy' ? 'color:var(--accent)' : 'color:var(--text-3)'"></i>
                <div class="text-caption text-primary">Proxy</div>
              </button>
              <button type="button" class="flex-1 p-2 text-center" @click="aliasType = 'redirect'"
                      :style="'border:2px solid ' + (aliasType === 'redirect' ? 'var(--accent)' : 'var(--border-subtle)') + ';border-radius:var(--radius-md);cursor:pointer;background:' + (aliasType === 'redirect' ? 'var(--accent-soft)' : 'var(--bg-2)')">
                <i data-lucide="external-link" class="w-4 h-4 mx-auto mb-1" :style="aliasType === 'redirect' ? 'color:var(--accent)' : 'color:var(--text-3)'"></i>
                <div class="text-caption text-primary">Redirect</div>
              </button>
              <button type="button" class="flex-1 p-2 text-center" @click="aliasType = 'reserved'"
                      :style="'border:2px solid ' + (aliasType === 'reserved' ? 'var(--accent)' : 'var(--border-subtle)') + ';border-radius:var(--radius-md);cursor:pointer;background:' + (aliasType === 'reserved' ? 'var(--accent-soft)' : 'var(--bg-2)')">
                <i data-lucide="lock" class="w-4 h-4 mx-auto mb-1" :style="aliasType === 'reserved' ? 'color:var(--accent)' : 'color:var(--text-3)'"></i>
                <div class="text-caption text-primary">Reserved</div>
              </button>
            </div>
          </div>
        </div>
        <div class="flex gap-2">
          <button class="btn btn-secondary flex-1 text-label" style="padding:8px 16px" @click="cancel">Cancel</button>
          <button class="btn btn-primary flex-1 text-label" style="padding:8px 16px" @click="create">Create Alias</button>
        </div>
        <button class="absolute top-4 right-4 p-1 btn-ghost" style="color:var(--text-3);border-radius:var(--radius-sm)" @click="cancel">
          <i data-lucide="x" class="w-5 h-5"></i>
        </button>
      </div>
    </div>
  `
}
