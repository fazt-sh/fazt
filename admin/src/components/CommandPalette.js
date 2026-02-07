import { ref, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useUIStore } from '../stores/ui.js'
import { useIcons } from '../lib/useIcons.js'

export default {
  name: 'CommandPalette',
  setup() {
    const router = useRouter()
    const ui = useUIStore()

    const commandInput = ref('')
    const selectedCommandIndex = ref(0)

    const commands = computed(() => [
      { id: 'dashboard', name: 'Dashboard', icon: 'layout-grid', action: () => router.push('/') },
      { id: 'apps', name: 'Apps', icon: 'layers', action: () => router.push('/apps') },
      { id: 'aliases', name: 'Aliases', icon: 'link', action: () => router.push('/aliases') },
      { id: 'logs', name: 'Logs', icon: 'terminal', action: () => router.push('/logs') },
      { id: 'system', name: 'System', icon: 'heart-pulse', action: () => router.push('/system') },
      { id: 'settings', name: 'Settings', icon: 'settings', action: () => router.push('/settings') },
      { id: 'new-app', name: 'New App', icon: 'plus', action: () => { ui.commandPaletteOpen = false; ui.newAppModalOpen = true } },
      { id: 'new-alias', name: 'New Alias', icon: 'link-2', action: () => { ui.commandPaletteOpen = false; ui.createAliasModalOpen = true } },
      { id: 'toggle-theme', name: 'Toggle Theme', icon: 'moon', action: () => ui.setTheme(ui.theme === 'dark' ? 'light' : 'dark') },
      { id: 'sign-out', name: 'Sign Out', icon: 'log-out', action: () => {} }
    ])

    const filteredCommands = computed(() => {
      const query = commandInput.value.toLowerCase()
      if (!query) return commands.value
      return commands.value.filter(cmd => cmd.name.toLowerCase().includes(query))
    })

    function close() {
      ui.commandPaletteOpen = false
      commandInput.value = ''
      selectedCommandIndex.value = 0
    }

    function execute(cmd) {
      if (cmd) {
        cmd.action()
        close()
      }
    }

    function handleKeydown(e) {
      const filtered = filteredCommands.value
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        selectedCommandIndex.value = Math.min(selectedCommandIndex.value + 1, filtered.length - 1)
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        selectedCommandIndex.value = Math.max(selectedCommandIndex.value - 1, 0)
      } else if (e.key === 'Enter') {
        e.preventDefault()
        execute(filtered[selectedCommandIndex.value])
      } else if (e.key === 'Escape') {
        e.preventDefault()
        close()
      }
    }

    watch(() => ui.commandPaletteOpen, (open) => {
      if (open) {
        commandInput.value = ''
        selectedCommandIndex.value = 0
        nextTick(() => {
          document.getElementById('command-input')?.focus()
        })
      }
    })

    useIcons()

    return { ui, commandInput, selectedCommandIndex, filteredCommands, close, execute, handleKeydown }
  },
  template: `
    <div v-if="ui.commandPaletteOpen">
      <div id="command-backdrop" class="fixed inset-0 z-50 modal-backdrop" @click="close"></div>
      <div id="command-palette" class="fixed z-50" style="top: 20%; left: 50%; transform: translateX(-50%); width: 500px;">
        <div class="modal p-0 overflow-hidden">
          <div class="p-3 border-b" style="border-color: var(--border)">
            <input id="command-input" type="text" placeholder="Type a command..."
                   class="w-full bg-transparent outline-none text-body" style="color: var(--text-1)"
                   v-model="commandInput" @keydown="handleKeydown">
          </div>
          <div id="command-results" class="max-h-80 overflow-auto scroll-panel">
            <div v-for="(cmd, idx) in filteredCommands" :key="cmd.id"
                 class="px-4 py-2.5 flex items-center gap-3 cursor-pointer"
                 :style="selectedCommandIndex === idx ? 'background:var(--accent-soft)' : ''"
                 @click="execute(cmd)"
                 @mouseenter="selectedCommandIndex = idx">
              <i :data-lucide="cmd.icon" class="w-4 h-4" :style="selectedCommandIndex === idx ? 'color:var(--accent)' : 'color:var(--text-3)'"></i>
              <span class="text-label" :style="selectedCommandIndex === idx ? 'color:var(--text-1)' : 'color:var(--text-2)'">{{ cmd.name }}</span>
            </div>
            <div v-if="filteredCommands.length === 0" class="px-4 py-8 text-center text-caption text-muted">
              No commands found
            </div>
          </div>
        </div>
      </div>
    </div>
  `
}
