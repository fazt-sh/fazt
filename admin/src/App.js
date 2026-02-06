import { onMounted, onUnmounted } from 'vue'
import { useUIStore } from './stores/ui.js'
import { useAuthStore } from './stores/auth.js'
import { useAppsStore } from './stores/apps.js'
import { useAliasesStore } from './stores/aliases.js'
import { useHealthStore } from './stores/health.js'
import { useLogsStore } from './stores/logs.js'
import { client } from './client.js'

import Sidebar from './components/Sidebar.js'
import HeaderBar from './components/HeaderBar.js'
import CommandPalette from './components/CommandPalette.js'
import SettingsPanel from './components/SettingsPanel.js'
import NewAppModal from './components/NewAppModal.js'
import CreateAliasModal from './components/CreateAliasModal.js'
import EditAliasModal from './components/EditAliasModal.js'

export default {
  name: 'App',
  components: { Sidebar, HeaderBar, CommandPalette, SettingsPanel, NewAppModal, CreateAliasModal, EditAliasModal },
  setup() {
    const ui = useUIStore()
    const auth = useAuthStore()
    const health = useHealthStore()

    async function initialize() {
      ui.initTheme()
      await auth.loadSession(client)

      if (!ui.useMock && !auth.checkAccess()) {
        return
      }

      const apps = useAppsStore()
      const aliases = useAliasesStore()
      const logs = useLogsStore()

      await Promise.all([
        apps.load(client),
        aliases.load(client),
        health.load(client),
        logs.loadStats(client)
      ])
    }

    function handleGlobalKeydown(e) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        ui.commandPaletteOpen = true
      }
      if (e.key === 'Escape') {
        ui.closeDropdowns()
        ui.mobileMenuOpen = false
      }
    }

    onMounted(() => {
      initialize()
      window.addEventListener('keydown', handleGlobalKeydown)
    })

    onUnmounted(() => {
      window.removeEventListener('keydown', handleGlobalKeydown)
    })

    const version = () => health.data.version || '0.26.0'

    return { ui, auth, health, version }
  },
  template: `
    <div>
      <SettingsPanel />
      <CommandPalette />
      <NewAppModal />
      <CreateAliasModal />
      <EditAliasModal />

      <div class="sidebar-backdrop" :class="{ active: ui.mobileMenuOpen }" @click="ui.mobileMenuOpen = false"></div>

      <div class="flex flex-col h-screen overflow-hidden">
        <div class="flex flex-1 overflow-hidden">
          <Sidebar />
          <main class="flex-1 flex flex-col overflow-hidden">
            <HeaderBar />
            <div id="page-content" class="flex-1 overflow-hidden p-5">
              <router-view></router-view>
            </div>
            <footer class="px-3 py-2 border-t flex items-center justify-between flex-shrink-0" style="background:var(--bg-1);border-color:var(--border);height:var(--footer-height)">
              <div class="flex items-center gap-2 text-caption" style="color:var(--text-4)">
                <span class="mono text-xs">v{{ version() }}</span>
                <span class="w-1.5 h-1.5 hide-mobile" style="background:var(--success);border-radius:2px"></span>
              </div>
              <div class="flex items-center gap-1 text-caption" style="color:var(--text-4)">
                <button v-if="ui.useMock" class="btn-icon btn-ghost" title="Mock mode active" style="border-radius:var(--radius-sm);width:28px;height:28px;color:var(--warning)">
                  <i data-lucide="database" class="w-3.5 h-3.5"></i>
                </button>
                <button class="btn-icon btn-ghost" title="Toggle theme panel" style="border-radius:var(--radius-sm);width:28px;height:28px" @click="ui.settingsPanelOpen = !ui.settingsPanelOpen">
                  <i data-lucide="palette" class="w-3.5 h-3.5"></i>
                </button>
                <a href="#" class="btn-icon btn-ghost" title="Documentation" style="border-radius:var(--radius-sm);width:28px;height:28px;display:flex;align-items:center;justify-content:center">
                  <i data-lucide="book-open" class="w-3.5 h-3.5"></i>
                </a>
              </div>
            </footer>
          </main>
        </div>
      </div>

      <div v-if="auth.unauthorized" class="fixed inset-0 z-50 flex items-center justify-center modal-backdrop">
        <div class="modal w-full max-w-sm p-6 text-center">
          <div class="w-12 h-12 mx-auto mb-4 flex items-center justify-center" style="background:var(--error-soft);border-radius:var(--radius-md)">
            <i data-lucide="shield-alert" class="w-6 h-6" style="color:var(--error)"></i>
          </div>
          <h2 class="text-title" style="color:var(--text-1)">Access Denied</h2>
          <p class="text-caption mt-1 mb-6" style="color:var(--text-3)">You don't have permission to access the admin panel.</p>
          <button class="btn btn-primary text-label" style="padding:8px 24px" @click="auth.signOut(client)">Sign Out</button>
        </div>
      </div>
    </div>
  `
}
