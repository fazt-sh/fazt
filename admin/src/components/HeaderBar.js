import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUIStore } from '../stores/ui.js'
import { useAuthStore } from '../stores/auth.js'
import { client } from '../client.js'
import { useIcons } from '../lib/useIcons.js'

export default {
  name: 'HeaderBar',
  setup() {
    const router = useRouter()
    const route = useRoute()
    const ui = useUIStore()
    const auth = useAuthStore()

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

    const userName = computed(() => auth.user?.email?.split('@')[0] || 'User')
    const userRole = computed(() => auth.user?.role || '')
    const userEmail = computed(() => auth.user?.email || 'User')
    const userRoleFull = computed(() => auth.user?.role || 'user')

    function navigateTo(path) {
      router.push(path)
    }

    async function handleSignOut() {
      ui.userDropdownOpen = false
      await auth.signOut(client)
    }

    useIcons()

    return { ui, auth, breadcrumbs, userName, userRole, userEmail, userRoleFull, navigateTo, handleSignOut }
  },
  template: `
    <div>
      <header class="flex items-center justify-between px-5 h-12 border-b flex-shrink-0" style="background:var(--bg-1);border-color:var(--border)">
        <div class="flex items-center gap-3">
          <button class="hamburger-btn" @click="ui.mobileMenuOpen = !ui.mobileMenuOpen">
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
          <button class="btn btn-ghost relative p-2" style="color:var(--text-2)" @click="ui.toggleNotifications()">
            <i data-lucide="bell" class="w-4 h-4"></i>
            <span v-if="ui.notifications.length > 0" class="badge absolute -top-0.5 -right-0.5 flex items-center justify-center" style="background:var(--error);color:white">{{ ui.notifications.length }}</span>
          </button>
          <button id="userBtn" class="flex items-center gap-2 px-2 py-1 btn-ghost" style="border-radius:var(--radius-sm)" @click="ui.toggleUserDropdown()">
            <div class="avatar-icon">
              <i data-lucide="user" class="w-3.5 h-3.5"></i>
            </div>
            <span class="text-label sidebar-text" style="color:var(--text-1)">{{ userName }}</span>
            <span class="text-micro sidebar-text" style="color:var(--text-4)">{{ userRole }}</span>
          </button>
        </div>
      </header>

      <div v-if="ui.notificationsOpen" id="notificationsDropdown" class="dropdown fixed z-50 w-80" style="top:56px;right:60px">
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

      <div v-if="ui.userDropdownOpen" id="userDropdown" class="dropdown fixed z-50 w-64" style="top:56px;right:16px">
        <div class="px-4 py-3 border-b" style="border-color:var(--border)">
          <div class="text-label" style="color:var(--text-1)">{{ userEmail }}</div>
          <div class="text-caption mt-0.5" style="color:var(--text-3)">{{ userRoleFull }}</div>
        </div>
        <div class="p-1">
          <button class="w-full flex items-center gap-2.5 px-3 py-2 text-label dropdown-item" style="border-radius:var(--radius-sm)" @click="() => { ui.userDropdownOpen = false; navigateTo('/settings') }">
            <i data-lucide="settings" class="w-4 h-4" style="color:var(--text-3)"></i>
            Settings
          </button>
          <button class="w-full flex items-center gap-2.5 px-3 py-2 text-label dropdown-item" style="border-radius:var(--radius-sm);color:var(--error)" @click="handleSignOut">
            <i data-lucide="log-out" class="w-4 h-4"></i>
            Sign Out
          </button>
        </div>
      </div>
    </div>
  `
}
