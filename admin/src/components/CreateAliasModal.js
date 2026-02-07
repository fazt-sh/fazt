import { ref } from 'vue'
import { useUIStore } from '../stores/ui.js'
import { useAliasesStore } from '../stores/aliases.js'
import { client } from '../client.js'
import { useIcons } from '../lib/useIcons.js'
import FModal from './FModal.js'

export default {
  name: 'CreateAliasModal',
  components: { FModal },
  setup() {
    useIcons()
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

    return { ui, subdomain, aliasType, target, create, cancel }
  },
  template: `
    <FModal :open="ui.createAliasModalOpen" title="Create Alias" subtitle="Point a subdomain to an app or URL" icon="link"
            @update:open="cancel">
      <div class="space-y-4">
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
      <template #footer>
        <button class="btn btn-secondary flex-1 text-label" style="padding:8px 16px" @click="cancel">Cancel</button>
        <button class="btn btn-primary flex-1 text-label" style="padding:8px 16px" @click="create">Create Alias</button>
      </template>
    </FModal>
  `
}
