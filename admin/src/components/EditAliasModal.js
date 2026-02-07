import { ref, watch } from 'vue'
import { useUIStore } from '../stores/ui.js'
import { useAliasesStore } from '../stores/aliases.js'
import { client } from '../client.js'
import { useIcons } from '../lib/useIcons.js'
import FModal from './FModal.js'

export default {
  name: 'EditAliasModal',
  components: { FModal },
  setup() {
    useIcons()
    const ui = useUIStore()
    const aliases = useAliasesStore()

    const editData = ref({})

    watch(() => ui.editingAlias, (newVal) => {
      if (newVal) {
        editData.value = { ...newVal }
      }
    })

    async function update() {
      if (!ui.editingAlias) return
      try {
        await aliases.update(client, ui.editingAlias.subdomain, editData.value)
        ui.editAliasModalOpen = false
        ui.editingAlias = null
        editData.value = {}
      } catch (err) {
        console.error('Failed to update alias:', err)
      }
    }

    function cancel() {
      ui.editAliasModalOpen = false
      ui.editingAlias = null
      editData.value = {}
    }

    return { ui, editData, update, cancel }
  },
  template: `
    <FModal :open="ui.editAliasModalOpen" title="Edit Alias" subtitle="Modify alias settings" icon="edit-3"
            @update:open="cancel">
      <div class="space-y-4">
        <div>
          <label class="text-micro text-muted block mb-2">TYPE</label>
          <div class="p-2 text-body" style="background:var(--bg-2);border-radius:var(--radius-sm);color:var(--text-1)">{{ editData.type || 'proxy' }}</div>
        </div>
      </div>
      <template #footer>
        <button class="btn btn-secondary text-label" style="padding:8px 16px;color:var(--error)">
          <i data-lucide="trash-2" class="w-4 h-4"></i>
        </button>
        <button class="btn btn-secondary flex-1 text-label" style="padding:8px 16px" @click="cancel">Cancel</button>
        <button class="btn btn-primary flex-1 text-label" style="padding:8px 16px" @click="update">Save Changes</button>
      </template>
    </FModal>
  `
}
