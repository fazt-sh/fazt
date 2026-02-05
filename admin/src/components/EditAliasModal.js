import { ref, watch, onMounted, onUpdated } from 'vue'
import { useUIStore } from '../stores/ui.js'
import { useAliasesStore } from '../stores/aliases.js'
import { client } from '../client.js'
import { refreshIcons } from '../lib/icons.js'

export default {
  name: 'EditAliasModal',
  setup() {
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

    onMounted(() => refreshIcons())
    onUpdated(() => refreshIcons())

    return { ui, editData, update, cancel }
  },
  template: `
    <div v-if="ui.editAliasModalOpen" class="fixed inset-0 z-50 flex items-center justify-center modal-backdrop" @click="cancel">
      <div class="modal w-full max-w-md p-6 relative" @click.stop>
        <div class="text-center mb-6">
          <div class="w-12 h-12 mx-auto mb-4 flex items-center justify-center" style="background:var(--accent-soft);border-radius:var(--radius-md)">
            <i data-lucide="edit-3" class="w-6 h-6" style="color:var(--accent)"></i>
          </div>
          <h2 class="text-title" style="color:var(--text-1)">Edit Alias</h2>
          <p class="text-caption mt-1" style="color:var(--text-3)">Modify alias settings</p>
        </div>
        <div class="space-y-4 mb-6">
          <div>
            <label class="text-micro text-muted block mb-2">TYPE</label>
            <div class="p-2 text-body" style="background:var(--bg-2);border-radius:var(--radius-sm);color:var(--text-1)">{{ editData.type || 'proxy' }}</div>
          </div>
        </div>
        <div class="flex gap-2">
          <button class="btn btn-secondary text-label" style="padding:8px 16px;color:var(--error)">
            <i data-lucide="trash-2" class="w-4 h-4"></i>
          </button>
          <button class="btn btn-secondary flex-1 text-label" style="padding:8px 16px" @click="cancel">Cancel</button>
          <button class="btn btn-primary flex-1 text-label" style="padding:8px 16px" @click="update">Save Changes</button>
        </div>
        <button class="absolute top-4 right-4 p-1 btn-ghost" style="color:var(--text-3);border-radius:var(--radius-sm)" @click="cancel">
          <i data-lucide="x" class="w-5 h-5"></i>
        </button>
      </div>
    </div>
  `
}
