import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUIStore } from '../stores/ui.js'
import { useAppsStore } from '../stores/apps.js'
import { useAliasesStore } from '../stores/aliases.js'
import { useHealthStore } from '../stores/health.js'
import { useLogsStore } from '../stores/logs.js'
import { useIcons } from '../lib/useIcons.js'

export default {
  name: 'AppSidebar',
  setup() {
    const router = useRouter()
    const route = useRoute()
    const ui = useUIStore()
    const apps = useAppsStore()
    const aliases = useAliasesStore()
    const health = useHealthStore()
    const logs = useLogsStore()

    const appCount = computed(() => apps.items.length)
    const aliasCount = computed(() => aliases.items.length)
    const logCount = computed(() => logs.stats.total_count || 0)
    const healthStatus = computed(() => health.data.status)
    const version = computed(() => health.data.version || '0.25')

    function isCurrentRoute(name) {
      return route.name === name
    }

    function navigateTo(path) {
      ui.mobileMenuOpen = false
      router.push(path)
    }

    useIcons()

    return { ui, appCount, aliasCount, logCount, healthStatus, version, isCurrentRoute, navigateTo }
  },
  template: `
    <aside id="sidebar" class="sidebar flex flex-col border-r overflow-hidden"
           :class="{ collapsed: ui.sidebarCollapsed, 'mobile-open': ui.mobileMenuOpen }"
           style="background:var(--bg-1);border-color:var(--border)">
      <div class="flex items-center gap-2.5 px-4 h-12 border-b flex-shrink-0" style="border-color:var(--border)">
        <button class="w-6 h-6 flex items-center justify-center btn-ghost" style="border-radius:var(--radius-sm)" @click="ui.toggleSidebar()">
          <i data-lucide="panel-left" class="w-4 h-4" style="color:var(--text-3)"></i>
        </button>
        <div class="sidebar-logo-text flex items-center gap-2 flex-1">
          <div class="w-6 h-6 flex items-center justify-center" style="background:var(--accent);border-radius:var(--radius-sm)">
            <i data-lucide="zap" class="w-3.5 h-3.5 text-white"></i>
          </div>
          <span class="text-heading" style="color:var(--text-1)">fazt</span>
          <span class="text-caption mono ml-auto" style="color:var(--text-4)">{{ version }}</span>
        </div>
      </div>

      <div class="sidebar-search px-3 py-2.5 border-b flex-shrink-0" style="border-color:var(--border)">
        <button class="w-full flex items-center gap-2 px-2.5 py-1.5 text-label border" @click="ui.commandPaletteOpen = true"
                style="background:var(--bg-2);border-color:var(--border-subtle);color:var(--text-3);border-radius:var(--radius-sm)">
          <i data-lucide="search" class="w-4 h-4"></i>
          <span class="sidebar-text">Search...</span>
          <span class="ml-auto kbd sidebar-text">&#x2318;K</span>
        </button>
      </div>

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
  `
}
