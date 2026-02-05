import { ref, computed, onMounted, onUpdated, watch, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUIStore } from './stores/ui.js'
import { useAuthStore } from './stores/auth.js'
import { useAppsStore } from './stores/apps.js'
import { useAliasesStore } from './stores/aliases.js'
import { useHealthStore } from './stores/health.js'
import { useLogsStore } from './stores/logs.js'
import { client } from './client.js'
import { refreshIcons } from './lib/icons.js'

export default {
  name: 'App',
  setup() {
    const router = useRouter()
    const route = useRoute()

    // Stores
    const ui = useUIStore()
    const auth = useAuthStore()
    const apps = useAppsStore()
    const aliases = useAliasesStore()
    const health = useHealthStore()
    const logs = useLogsStore()

    // Local UI state
    const userDropdownOpen = ref(false)
    const notificationsOpen = ref(false)
    const mobileMenuOpen = ref(false)
    const commandInput = ref('')
    const selectedCommandIndex = ref(0)
    const settingsPanelPosition = ref({ x: 20, y: window.innerHeight - 420 })
    const draggingSettings = ref(false)
    const dragOffset = ref({ x: 0, y: 0 })

    // New app form
    const newAppName = ref('')
    const newAppTemplate = ref('minimal')

    // Alias forms
    const aliasSubdomain = ref('')
    const aliasType = ref('proxy')
    const aliasTarget = ref('')
    const editAliasData = ref({})

    // Computed
    const breadcrumbs = computed(() => {
      const meta = route.meta || {}
      const crumbs = [{ name: 'Dashboard', route: '/' }]

      if (meta.parent) {
        const parentRoute = router.options.routes.find(r => r.name === meta.parent)
        if (parentRoute) {
          crumbs.push({ name: parentRoute.meta?.title || meta.parent, route: parentRoute.path })
        }
      }

      if (route.name !== 'dashboard') {
        crumbs.push({ name: meta.title || route.name, route: route.path })
      }

      return crumbs
    })

    const appCount = computed(() => apps.items.length)
    const aliasCount = computed(() => aliases.items.length)
    const logCount = computed(() => logs.total)
    const healthStatus = computed(() => health.data.status)

    const isCurrentRoute = (routeName) => {
      return route.name === routeName
    }

    // Command palette commands
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
      { id: 'sign-out', name: 'Sign Out', icon: 'log-out', action: () => handleSignOut() }
    ])

    const filteredCommands = computed(() => {
      const query = commandInput.value.toLowerCase()
      if (!query) return commands.value
      return commands.value.filter(cmd => cmd.name.toLowerCase().includes(query))
    })

    // Methods
    function navigateTo(path) {
      mobileMenuOpen.value = false
      router.push(path)
    }

    function toggleMobileMenu() {
      mobileMenuOpen.value = !mobileMenuOpen.value
    }

    function toggleUserDropdown() {
      userDropdownOpen.value = !userDropdownOpen.value
      notificationsOpen.value = false
    }

    function toggleNotifications() {
      notificationsOpen.value = !notificationsOpen.value
      userDropdownOpen.value = false
    }

    function closeDropdowns() {
      userDropdownOpen.value = false
      notificationsOpen.value = false
    }

    async function handleSignOut() {
      userDropdownOpen.value = false
      await auth.signOut(client)
    }

    // Command palette
    function openCommandPalette() {
      ui.commandPaletteOpen = true
      commandInput.value = ''
      selectedCommandIndex.value = 0
      nextTick(() => {
        document.getElementById('command-input')?.focus()
      })
    }

    function closeCommandPalette() {
      ui.commandPaletteOpen = false
      commandInput.value = ''
      selectedCommandIndex.value = 0
    }

    function selectCommand(index) {
      selectedCommandIndex.value = index
    }

    function executeCommand(cmd) {
      if (cmd) {
        cmd.action()
        closeCommandPalette()
      }
    }

    function handleCommandKeydown(e) {
      const filtered = filteredCommands.value

      if (e.key === 'ArrowDown') {
        e.preventDefault()
        selectedCommandIndex.value = Math.min(selectedCommandIndex.value + 1, filtered.length - 1)
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        selectedCommandIndex.value = Math.max(selectedCommandIndex.value - 1, 0)
      } else if (e.key === 'Enter') {
        e.preventDefault()
        executeCommand(filtered[selectedCommandIndex.value])
      } else if (e.key === 'Escape') {
        e.preventDefault()
        closeCommandPalette()
      }
    }

    // Settings panel dragging
    function startDragSettings(e) {
      draggingSettings.value = true
      const panel = document.getElementById('settingsPanel')
      if (panel) {
        const rect = panel.getBoundingClientRect()
        dragOffset.value = {
          x: e.clientX - rect.left,
          y: e.clientY - rect.top
        }
      }
    }

    function onDragSettings(e) {
      if (!draggingSettings.value) return
      settingsPanelPosition.value = {
        x: e.clientX - dragOffset.value.x,
        y: e.clientY - dragOffset.value.y
      }
    }

    function stopDragSettings() {
      draggingSettings.value = false
    }

    // New app modal
    async function createApp() {
      if (!newAppName.value.trim()) return
      try {
        await apps.create(client, newAppName.value, newAppTemplate.value)
        ui.newAppModalOpen = false
        newAppName.value = ''
        newAppTemplate.value = 'minimal'
      } catch (err) {
        console.error('Failed to create app:', err)
      }
    }

    function cancelNewApp() {
      ui.newAppModalOpen = false
      newAppName.value = ''
      newAppTemplate.value = 'minimal'
    }

    // Create alias modal
    async function createAlias() {
      if (!aliasSubdomain.value.trim()) return
      try {
        const options = aliasType.value === 'app' ? { app_id: aliasTarget.value } : {}
        await aliases.create(client, aliasSubdomain.value, aliasType.value, options)
        ui.createAliasModalOpen = false
        aliasSubdomain.value = ''
        aliasType.value = 'app'
        aliasTarget.value = ''
      } catch (err) {
        console.error('Failed to create alias:', err)
      }
    }

    function cancelCreateAlias() {
      ui.createAliasModalOpen = false
      aliasSubdomain.value = ''
      aliasType.value = 'proxy'
      aliasTarget.value = ''
    }

    // Edit alias modal
    async function updateAlias() {
      if (!ui.editingAlias) return
      try {
        await aliases.update(client, ui.editingAlias.subdomain, editAliasData.value)
        ui.editAliasModalOpen = false
        ui.editingAlias = null
        editAliasData.value = {}
      } catch (err) {
        console.error('Failed to update alias:', err)
      }
    }

    function cancelEditAlias() {
      ui.editAliasModalOpen = false
      ui.editingAlias = null
      editAliasData.value = {}
    }

    // Keyboard shortcuts
    function handleGlobalKeydown(e) {
      // Cmd/Ctrl + K for command palette
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        openCommandPalette()
      }
      // Escape to close dropdowns
      if (e.key === 'Escape') {
        closeDropdowns()
        mobileMenuOpen.value = false
      }
    }

    // Palettes for settings panel swatches
    const palettes = [
      { id: 'stone', name: 'Stone', colors: ['#faf9f7', '#d97706'] },
      { id: 'slate', name: 'Slate', colors: ['#f8fafc', '#0284c7'] },
      { id: 'oxide', name: 'Oxide', colors: ['#faf8f8', '#e11d48'] },
      { id: 'forest', name: 'Forest', colors: ['#f7faf8', '#059669'] },
      { id: 'violet', name: 'Violet', colors: ['#faf9fc', '#7c3aed'] }
    ]

    // Initialize
    async function initialize() {
      // Init theme
      ui.initTheme()

      // Load session
      await auth.loadSession(client)

      // Check access (redirects if unauthorized)
      if (!ui.useMock && !auth.checkAccess()) {
        return
      }

      // Load all data in parallel
      await Promise.all([
        apps.load(client),
        aliases.load(client),
        health.load(client),
        logs.loadStats(client)
      ])
    }

    // Lifecycle
    onMounted(() => {
      initialize()
      refreshIcons()
      window.addEventListener('keydown', handleGlobalKeydown)
      window.addEventListener('mousemove', onDragSettings)
      window.addEventListener('mouseup', stopDragSettings)
    })

    onUpdated(() => {
      refreshIcons()
    })

    // Watch for editing alias changes
    watch(() => ui.editingAlias, (newVal) => {
      if (newVal) {
        editAliasData.value = { ...newVal }
      }
    })

    // Watch for command palette open
    watch(() => ui.commandPaletteOpen, (newVal) => {
      if (newVal) {
        nextTick(() => {
          document.getElementById('command-input')?.focus()
        })
      }
    })

    return {
      // Stores
      ui, auth, apps, aliases, health, logs,

      // Route
      route, router,

      // Local state
      userDropdownOpen, notificationsOpen, mobileMenuOpen,
      commandInput, selectedCommandIndex,
      settingsPanelPosition, draggingSettings,
      newAppName, newAppTemplate,
      aliasSubdomain, aliasType, aliasTarget,
      editAliasData,

      // Computed
      breadcrumbs, appCount, aliasCount, logCount, healthStatus,
      filteredCommands,

      // Data
      palettes,

      // Methods
      isCurrentRoute, navigateTo, toggleMobileMenu,
      toggleUserDropdown, toggleNotifications, closeDropdowns,
      handleSignOut,
      openCommandPalette, closeCommandPalette, selectCommand, executeCommand, handleCommandKeydown,
      startDragSettings,
      createApp, cancelNewApp,
      createAlias, cancelCreateAlias,
      updateAlias, cancelEditAlias
    }
  },

  template: `
    <!-- Settings Panel (floating, original markup) -->
    <div v-if="ui.settingsPanelOpen" id="settingsPanel" class="settings-panel fixed" style="z-index: 1100; bottom: 20px; left: 50%; transform: translateX(-50%);">
      <div class="settings-handle" @mousedown="startDragSettings">
        <div class="grip"></div>
      </div>
      <div class="p-4 flex items-center gap-6">
        <div class="flex items-center gap-3">
          <span class="text-micro" style="color:var(--text-3)">Theme</span>
          <div class="flex gap-1">
            <button class="pill" :class="{ active: ui.theme === 'light' }" @click="ui.setTheme('light')">
              <i data-lucide="sun" class="w-3.5 h-3.5 inline-block mr-1" style="vertical-align: -2px"></i>Light
            </button>
            <button class="pill" :class="{ active: ui.theme === 'dark' }" @click="ui.setTheme('dark')">
              <i data-lucide="moon" class="w-3.5 h-3.5 inline-block mr-1" style="vertical-align: -2px"></i>Dark
            </button>
          </div>
        </div>
        <div class="w-px h-6" style="background:var(--border)"></div>
        <div class="flex items-center gap-3">
          <span class="text-micro" style="color:var(--text-3)">Palette</span>
          <div class="flex gap-2">
            <div v-for="p in palettes" :key="p.id"
                 class="swatch" :class="{ active: ui.palette === p.id }"
                 :title="p.name"
                 :style="'background: linear-gradient(135deg, ' + p.colors[0] + ' 50%, ' + p.colors[1] + ' 50%)'"
                 @click="ui.setPalette(p.id)"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Command Palette Modal -->
    <div v-if="ui.commandPaletteOpen" id="command-backdrop" class="fixed inset-0 z-50 modal-backdrop" @click="closeCommandPalette"></div>
    <div v-if="ui.commandPaletteOpen" id="command-palette" class="fixed z-50" style="top: 20%; left: 50%; transform: translateX(-50%); width: 500px;">
      <div class="modal p-0 overflow-hidden">
        <div class="p-3 border-b" style="border-color: var(--border)">
          <input id="command-input" type="text" placeholder="Type a command..."
                 class="w-full bg-transparent outline-none text-body" style="color: var(--text-1)"
                 v-model="commandInput" @keydown="handleCommandKeydown">
        </div>
        <div id="command-results" class="max-h-80 overflow-auto scroll-panel">
          <div v-for="(cmd, idx) in filteredCommands" :key="cmd.id"
               class="px-4 py-2.5 flex items-center gap-3 cursor-pointer"
               :style="selectedCommandIndex === idx ? 'background:var(--accent-soft)' : ''"
               @click="executeCommand(cmd)"
               @mouseenter="selectCommand(idx)">
            <i :data-lucide="cmd.icon" class="w-4 h-4" :style="selectedCommandIndex === idx ? 'color:var(--accent)' : 'color:var(--text-3)'"></i>
            <span class="text-label" :style="selectedCommandIndex === idx ? 'color:var(--text-1)' : 'color:var(--text-2)'">{{ cmd.name }}</span>
          </div>
          <div v-if="filteredCommands.length === 0" class="px-4 py-8 text-center text-caption text-muted">
            No commands found
          </div>
        </div>
      </div>
    </div>

    <!-- New App Modal -->
    <div v-if="ui.newAppModalOpen" class="fixed inset-0 z-50 flex items-center justify-center modal-backdrop" @click="cancelNewApp">
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
              <input type="text" v-model="newAppName" placeholder="my-app" class="text-body" style="flex: 1">
            </div>
            <div class="text-caption text-muted mt-1 px-1">Lowercase letters, numbers, and dashes only</div>
          </div>
          <div>
            <label class="text-micro text-muted block mb-2">TEMPLATE</label>
            <div class="grid grid-cols-2 gap-2">
              <button type="button" class="p-3 text-left" @click="newAppTemplate = 'minimal'"
                      :style="'background:var(--bg-2);border:2px solid ' + (newAppTemplate === 'minimal' ? 'var(--accent)' : 'var(--border-subtle)') + ';border-radius:var(--radius-md);cursor:pointer'">
                <div class="flex items-center gap-2 mb-1">
                  <i data-lucide="file" class="w-4 h-4" :style="newAppTemplate === 'minimal' ? 'color:var(--accent)' : 'color:var(--text-3)'"></i>
                  <span class="text-label text-primary">Minimal</span>
                </div>
                <div class="text-caption text-muted">Static site</div>
              </button>
              <button type="button" class="p-3 text-left" @click="newAppTemplate = 'spa'"
                      :style="'background:var(--bg-2);border:2px solid ' + (newAppTemplate === 'spa' ? 'var(--accent)' : 'var(--border-subtle)') + ';border-radius:var(--radius-md);cursor:pointer'">
                <div class="flex items-center gap-2 mb-1">
                  <i data-lucide="layout" class="w-4 h-4" :style="newAppTemplate === 'spa' ? 'color:var(--accent)' : 'color:var(--text-3)'"></i>
                  <span class="text-label text-primary">SPA</span>
                </div>
                <div class="text-caption text-muted">With routing</div>
              </button>
              <button type="button" class="p-3 text-left" @click="newAppTemplate = 'api'"
                      :style="'background:var(--bg-2);border:2px solid ' + (newAppTemplate === 'api' ? 'var(--accent)' : 'var(--border-subtle)') + ';border-radius:var(--radius-md);cursor:pointer'">
                <div class="flex items-center gap-2 mb-1">
                  <i data-lucide="code" class="w-4 h-4" :style="newAppTemplate === 'api' ? 'color:var(--accent)' : 'color:var(--text-3)'"></i>
                  <span class="text-label text-primary">API</span>
                </div>
                <div class="text-caption text-muted">Serverless only</div>
              </button>
              <button type="button" class="p-3 text-left" @click="newAppTemplate = 'full'"
                      :style="'background:var(--bg-2);border:2px solid ' + (newAppTemplate === 'full' ? 'var(--accent)' : 'var(--border-subtle)') + ';border-radius:var(--radius-md);cursor:pointer'">
                <div class="flex items-center gap-2 mb-1">
                  <i data-lucide="layers" class="w-4 h-4" :style="newAppTemplate === 'full' ? 'color:var(--accent)' : 'color:var(--text-3)'"></i>
                  <span class="text-label text-primary">Full Stack</span>
                </div>
                <div class="text-caption text-muted">Static + API</div>
              </button>
            </div>
          </div>
        </div>
        <div class="flex gap-2">
          <button class="btn btn-secondary flex-1 text-label" style="padding:8px 16px" @click="cancelNewApp">Cancel</button>
          <button class="btn btn-primary flex-1 text-label" style="padding:8px 16px" @click="createApp">Create App</button>
        </div>
        <button class="absolute top-4 right-4 p-1 btn-ghost" style="color:var(--text-3);border-radius:var(--radius-sm)" @click="cancelNewApp">
          <i data-lucide="x" class="w-5 h-5"></i>
        </button>
      </div>
    </div>

    <!-- Create Alias Modal -->
    <div v-if="ui.createAliasModalOpen" class="fixed inset-0 z-50 flex items-center justify-center modal-backdrop" @click="cancelCreateAlias">
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
              <input type="text" v-model="aliasSubdomain" placeholder="my-alias" class="text-body" style="flex: 1">
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
          <button class="btn btn-secondary flex-1 text-label" style="padding:8px 16px" @click="cancelCreateAlias">Cancel</button>
          <button class="btn btn-primary flex-1 text-label" style="padding:8px 16px" @click="createAlias">Create Alias</button>
        </div>
        <button class="absolute top-4 right-4 p-1 btn-ghost" style="color:var(--text-3);border-radius:var(--radius-sm)" @click="cancelCreateAlias">
          <i data-lucide="x" class="w-5 h-5"></i>
        </button>
      </div>
    </div>

    <!-- Edit Alias Modal -->
    <div v-if="ui.editAliasModalOpen" class="fixed inset-0 z-50 flex items-center justify-center modal-backdrop" @click="cancelEditAlias">
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
            <div class="p-2 text-body" style="background:var(--bg-2);border-radius:var(--radius-sm);color:var(--text-1)">{{ editAliasData.type || 'proxy' }}</div>
          </div>
        </div>
        <div class="flex gap-2">
          <button class="btn btn-secondary text-label" style="padding:8px 16px;color:var(--error)">
            <i data-lucide="trash-2" class="w-4 h-4"></i>
          </button>
          <button class="btn btn-secondary flex-1 text-label" style="padding:8px 16px" @click="cancelEditAlias">Cancel</button>
          <button class="btn btn-primary flex-1 text-label" style="padding:8px 16px" @click="updateAlias">Save Changes</button>
        </div>
        <button class="absolute top-4 right-4 p-1 btn-ghost" style="color:var(--text-3);border-radius:var(--radius-sm)" @click="cancelEditAlias">
          <i data-lucide="x" class="w-5 h-5"></i>
        </button>
      </div>
    </div>

    <!-- Notifications Dropdown -->
    <div v-if="notificationsOpen" id="notificationsDropdown" class="dropdown fixed z-50 w-80" style="top:56px;right:60px">
      <div class="px-4 py-3 border-b flex items-center justify-between" style="border-color:var(--border)">
        <span class="text-heading" style="color:var(--text-1)">Notifications</span>
        <button class="text-caption" style="color:var(--accent)">Mark all read</button>
      </div>
      <div class="max-h-80 overflow-auto scroll-panel">
        <div v-if="ui.notifications.length === 0" class="px-4 py-8 text-center text-caption text-muted">No notifications</div>
        <div v-for="n in ui.notifications" :key="n.id" class="px-4 py-3 flex items-start gap-3 dropdown-item cursor-pointer">
          <div class="icon-box flex-shrink-0 mt-0.5"><i :data-lucide="n.type === 'error' ? 'alert-circle' : 'check-circle'" class="w-4 h-4"></i></div>
          <div class="flex-1 min-w-0">
            <div class="text-label" style="color:var(--text-1)">{{ n.message }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- User Dropdown -->
    <div v-if="userDropdownOpen" id="userDropdown" class="dropdown fixed z-50 w-64" style="top:56px;right:16px">
      <div class="px-4 py-3 border-b" style="border-color:var(--border)">
        <div class="text-label" style="color:var(--text-1)">{{ auth.user?.email || 'User' }}</div>
        <div class="text-caption mt-0.5" style="color:var(--text-3)">{{ auth.user?.role || 'user' }}</div>
      </div>
      <div class="p-1">
        <button class="w-full flex items-center gap-2.5 px-3 py-2 text-label dropdown-item" style="border-radius:var(--radius-sm)" @click="() => { userDropdownOpen = false; navigateTo('/settings') }">
          <i data-lucide="settings" class="w-4 h-4" style="color:var(--text-3)"></i>
          Settings
        </button>
        <button class="w-full flex items-center gap-2.5 px-3 py-2 text-label dropdown-item" style="border-radius:var(--radius-sm);color:var(--error)" @click="handleSignOut">
          <i data-lucide="log-out" class="w-4 h-4"></i>
          Sign Out
        </button>
      </div>
    </div>

    <!-- Sidebar Backdrop (mobile only) -->
    <div class="sidebar-backdrop" :class="{ active: mobileMenuOpen }" @click="mobileMenuOpen = false"></div>

    <!-- App Shell -->
    <div class="flex flex-col h-screen overflow-hidden">
      <div class="flex flex-1 overflow-hidden">
        <!-- Sidebar -->
        <aside id="sidebar" class="sidebar flex flex-col border-r overflow-hidden"
               :class="{ collapsed: ui.sidebarCollapsed, 'mobile-open': mobileMenuOpen }"
               style="background:var(--bg-1);border-color:var(--border)">
          <!-- Logo -->
          <div class="flex items-center gap-2.5 px-4 h-12 border-b flex-shrink-0" style="border-color:var(--border)">
            <button class="w-6 h-6 flex items-center justify-center btn-ghost" style="border-radius:var(--radius-sm)" @click="ui.toggleSidebar()">
              <i data-lucide="panel-left" class="w-4 h-4" style="color:var(--text-3)"></i>
            </button>
            <div class="sidebar-logo-text flex items-center gap-2 flex-1">
              <div class="w-6 h-6 flex items-center justify-center" style="background:var(--accent);border-radius:var(--radius-sm)">
                <i data-lucide="zap" class="w-3.5 h-3.5 text-white"></i>
              </div>
              <span class="text-heading" style="color:var(--text-1)">fazt</span>
              <span class="text-caption mono ml-auto" style="color:var(--text-4)">{{ health.data.version || '0.25' }}</span>
            </div>
          </div>

          <!-- Search -->
          <div class="sidebar-search px-3 py-2.5 border-b flex-shrink-0" style="border-color:var(--border)">
            <button class="w-full flex items-center gap-2 px-2.5 py-1.5 text-label border" @click="openCommandPalette"
                    style="background:var(--bg-2);border-color:var(--border-subtle);color:var(--text-3);border-radius:var(--radius-sm)">
              <i data-lucide="search" class="w-4 h-4"></i>
              <span class="sidebar-text">Search...</span>
              <span class="ml-auto kbd sidebar-text">&#x2318;K</span>
            </button>
          </div>

          <!-- Nav -->
          <nav class="flex-1 p-2 space-y-0.5 scroll-panel">
            <a href="#" class="nav-item flex items-center gap-2.5 px-3 py-2 text-label"
               :class="{ active: isCurrentRoute('dashboard') }"
               @click.prevent="navigateTo('/')">
              <i data-lucide="layout-grid" class="w-4 h-4"></i>
              <span class="sidebar-text">Dashboard</span>
            </a>

            <div class="menu-section pt-4 pb-1">
              <div class="menu-toggle open flex items-center gap-1 px-3">
                <i data-lucide="chevron-right" class="w-3 h-3 chevron" style="color:var(--text-4)"></i>
                <span class="text-micro" style="color:var(--text-4)">Resources</span>
              </div>
            </div>
            <div class="submenu open space-y-0.5">
              <a href="#" class="nav-item flex items-center gap-2.5 px-3 py-2 text-label"
                 :class="{ active: isCurrentRoute('apps') || isCurrentRoute('app-detail') }"
                 @click.prevent="navigateTo('/apps')">
                <i data-lucide="layers" class="w-4 h-4"></i>
                <span class="sidebar-text">Apps</span>
                <span v-if="appCount" class="nav-badge ml-auto text-caption mono px-1.5 py-0.5" style="background:var(--bg-2);color:var(--text-3);border-radius:var(--radius-sm)">{{ appCount }}</span>
              </a>
              <a href="#" class="nav-item flex items-center gap-2.5 px-3 py-2 text-label"
                 :class="{ active: isCurrentRoute('aliases') }"
                 @click.prevent="navigateTo('/aliases')">
                <i data-lucide="link" class="w-4 h-4"></i>
                <span class="sidebar-text">Aliases</span>
                <span v-if="aliasCount" class="nav-badge ml-auto text-caption mono px-1.5 py-0.5" style="background:var(--bg-2);color:var(--text-3);border-radius:var(--radius-sm)">{{ aliasCount }}</span>
              </a>
            </div>

            <div class="menu-section pt-4 pb-1">
              <div class="menu-toggle open flex items-center gap-1 px-3">
                <i data-lucide="chevron-right" class="w-3 h-3 chevron" style="color:var(--text-4)"></i>
                <span class="text-micro" style="color:var(--text-4)">Observability</span>
              </div>
            </div>
            <div class="submenu open space-y-0.5">
              <a href="#" class="nav-item flex items-center gap-2.5 px-3 py-2 text-label"
                 :class="{ active: isCurrentRoute('logs') }"
                 @click.prevent="navigateTo('/logs')">
                <i data-lucide="terminal" class="w-4 h-4"></i>
                <span class="sidebar-text">Logs</span>
                <span v-if="logCount" class="nav-badge ml-auto text-caption mono px-1.5 py-0.5" style="background:var(--bg-2);color:var(--text-3);border-radius:var(--radius-sm)">{{ logCount }}</span>
              </a>
            </div>

            <div class="menu-section pt-4 pb-1">
              <div class="menu-toggle open flex items-center gap-1 px-3">
                <i data-lucide="chevron-right" class="w-3 h-3 chevron" style="color:var(--text-4)"></i>
                <span class="text-micro" style="color:var(--text-4)">System</span>
              </div>
            </div>
            <div class="submenu open space-y-0.5">
              <a href="#" class="nav-item flex items-center gap-2.5 px-3 py-2 text-label"
                 :class="{ active: isCurrentRoute('system') }"
                 @click.prevent="navigateTo('/system')">
                <i data-lucide="heart-pulse" class="w-4 h-4"></i>
                <span class="sidebar-text">Health</span>
                <span v-if="healthStatus === 'healthy'" class="nav-badge ml-auto w-2 h-2 pulse" style="background:var(--success);border-radius:var(--radius-sm)"></span>
              </a>
              <a href="#" class="nav-item flex items-center gap-2.5 px-3 py-2 text-label"
                 :class="{ active: isCurrentRoute('settings') }"
                 @click.prevent="navigateTo('/settings')">
                <i data-lucide="settings" class="w-4 h-4"></i>
                <span class="sidebar-text">Settings</span>
              </a>
            </div>
          </nav>
        </aside>

        <!-- Main -->
        <main class="flex-1 flex flex-col overflow-hidden">
          <!-- Header -->
          <header class="flex items-center justify-between px-5 h-12 border-b flex-shrink-0" style="background:var(--bg-1);border-color:var(--border)">
            <div class="flex items-center gap-3">
              <button class="hamburger-btn" @click="toggleMobileMenu">
                <i data-lucide="menu" class="w-5 h-5" style="color:var(--text-2)"></i>
              </button>
              <div id="breadcrumb" class="flex items-center gap-2 text-label">
                <template v-for="(crumb, idx) in breadcrumbs" :key="idx">
                  <span v-if="idx < breadcrumbs.length - 1" style="color:var(--text-3)" class="cursor-pointer" @click="navigateTo(crumb.route)">{{ crumb.name }}</span>
                  <span v-else style="color:var(--text-1)">{{ crumb.name }}</span>
                  <i v-if="idx < breadcrumbs.length - 1" data-lucide="chevron-right" class="w-3 h-3" style="color:var(--text-4)"></i>
                </template>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <button class="btn btn-ghost relative p-2" style="color:var(--text-2)" @click="toggleNotifications">
                <i data-lucide="bell" class="w-4 h-4"></i>
                <span v-if="ui.notifications.length > 0" class="badge absolute -top-0.5 -right-0.5 flex items-center justify-center" style="background:var(--error);color:white">{{ ui.notifications.length }}</span>
              </button>
              <button id="userBtn" class="flex items-center gap-2 px-2 py-1 btn-ghost" style="border-radius:var(--radius-sm)" @click="toggleUserDropdown">
                <div class="avatar-icon">
                  <i data-lucide="user" class="w-3.5 h-3.5"></i>
                </div>
                <span class="text-label sidebar-text" style="color:var(--text-1)">{{ auth.user?.email?.split('@')[0] || 'User' }}</span>
                <span class="text-micro sidebar-text" style="color:var(--text-4)">{{ auth.user?.role || '' }}</span>
              </button>
            </div>
          </header>

          <!-- Content -->
          <div id="page-content" class="flex-1 overflow-hidden p-5">
            <router-view></router-view>
          </div>

          <!-- Footer -->
          <footer class="px-3 py-2 border-t flex items-center justify-between flex-shrink-0" style="background:var(--bg-1);border-color:var(--border);height:var(--footer-height)">
            <div class="flex items-center gap-2 text-caption" style="color:var(--text-4)">
              <span class="mono text-xs">v{{ health.data.version || '0.25.4' }}</span>
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

    <!-- Unauthorized overlay -->
    <div v-if="auth.unauthorized" class="fixed inset-0 z-50 flex items-center justify-center modal-backdrop">
      <div class="modal w-full max-w-sm p-6 text-center">
        <div class="w-12 h-12 mx-auto mb-4 flex items-center justify-center" style="background:var(--error-soft);border-radius:var(--radius-md)">
          <i data-lucide="shield-alert" class="w-6 h-6" style="color:var(--error)"></i>
        </div>
        <h2 class="text-title" style="color:var(--text-1)">Access Denied</h2>
        <p class="text-caption mt-1 mb-6" style="color:var(--text-3)">You don't have permission to access the admin panel.</p>
        <button class="btn btn-primary text-label" style="padding:8px 24px" @click="handleSignOut">Sign Out</button>
      </div>
    </div>
  `
}
