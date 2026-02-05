# Plan 31: Admin UI Vue Port

**Created**: 2026-02-05
**Status**: IN PROGRESS
**Goal**: Port React admin UI to Vue using fazt-app patterns with proper state management

---

## Overview

### Why Port to Vue?
1. **Ecosystem consistency** - All fazt apps use Vue
2. **Better reactivity** - Vue's built-in reactivity is simpler than React hooks
3. **fazt-sdk integration** - Vue composables cleaner than React hooks
4. **State management** - Pinia is more elegant than React Context
5. **Skill reuse** - Patterns from /fazt-app skill transfer directly

### What Stays the Same
- Visual design (exact pixel match)
- Tailwind CSS + fazt-ui design tokens
- All component styling
- API endpoints (backend unchanged)

### What Changes
- React → Vue 3 (Composition API)
- react-router → vue-router
- React Context → Pinia stores
- @tanstack/react-query → @tanstack/vue-query
- lib/api.ts → fazt-sdk (single source of truth)

---

## Current Admin Inventory

### File Count
- **~60 TSX files** (components, pages, layouts)
- **~26,000 lines** total
- **4 context providers** (Auth, Mock, Theme, Toast)
- **~20 UI components**
- **~25 pages**

### Largest Files (Priority for careful porting)
| File | Lines | Notes |
|------|-------|-------|
| `SiteDetail.tsx` | 726 | Complex, data-heavy |
| `DesignSystemPreline.tsx` | 346 | Reference only |
| `Datamap.tsx` | 306 | D3/datamaps integration |
| `SitesAnalytics.tsx` | 305 | Charts, data fetching |
| `WorldMap.tsx` | 299 | D3 integration |
| `SystemInfo.tsx` | 294 | Multiple gauges |
| `Sidebar.tsx` | 266 | Navigation state |

### Route Structure (Preserve Exactly)
```
/login
/dashboard
/profile
/sites
  /sites/analytics
  /sites/domains
  /sites/create
  /sites/:id
/system
  /system/stats
  /system/limits
  /system/logs
  /system/backup
  /system/health
  /system/settings
  /system/design-system
/apps
  /apps (list)
  /apps/webhooks
  /apps/redirects
  /apps/tunnel
  /apps/proxy
  /apps/botdaddy
/security
  /security/ssh
  /security/tokens
/external
  /external/cloudflare
  /external/litestream
```

---

## New Vue Project Structure

```
admin-vue/
├── index.html
├── vite.config.js
├── package.json
├── tailwind.config.js
├── postcss.config.js
├── tsconfig.json
│
├── src/
│   ├── main.js                 # App entry, plugins
│   ├── App.vue                 # Root component
│   │
│   ├── router/
│   │   └── index.js            # Vue Router config
│   │
│   ├── stores/                 # Pinia stores
│   │   ├── auth.js             # Auth state
│   │   ├── theme.js            # Theme state
│   │   ├── toast.js            # Toast notifications
│   │   └── mock.js             # Mock mode toggle
│   │
│   ├── composables/            # Vue composables (hooks)
│   │   ├── useAuth.js
│   │   ├── useTheme.js
│   │   ├── useToast.js
│   │   └── useMock.js
│   │
│   ├── lib/
│   │   └── fazt-sdk/           # Copied from admin/packages/fazt-sdk
│   │       ├── index.js        # Main client
│   │       ├── client.js       # HTTP client
│   │       ├── mock.js         # Mock adapter
│   │       ├── types.js        # JSDoc types
│   │       ├── vue.js          # NEW: Vue composables
│   │       └── fixtures/       # Mock data
│   │
│   ├── pages/
│   │   ├── LoginPage.vue
│   │   ├── DashboardPage.vue
│   │   ├── ProfilePage.vue
│   │   ├── NotFoundPage.vue
│   │   │
│   │   ├── sites/
│   │   │   ├── SitesPage.vue
│   │   │   ├── SitesAnalyticsPage.vue
│   │   │   ├── SitesDomainsPage.vue
│   │   │   ├── CreateSitePage.vue
│   │   │   └── SiteDetailPage.vue
│   │   │
│   │   ├── system/
│   │   │   ├── SystemStatsPage.vue
│   │   │   ├── SystemLimitsPage.vue
│   │   │   ├── SystemLogsPage.vue
│   │   │   ├── SystemBackupPage.vue
│   │   │   ├── SystemHealthPage.vue
│   │   │   ├── SystemSettingsPage.vue
│   │   │   └── DesignSystemPage.vue
│   │   │
│   │   ├── apps/
│   │   │   ├── AppsListPage.vue
│   │   │   ├── WebhooksPage.vue
│   │   │   ├── RedirectsPage.vue
│   │   │   ├── TunnelPage.vue
│   │   │   ├── ProxyPage.vue
│   │   │   └── BotDaddyPage.vue
│   │   │
│   │   ├── security/
│   │   │   ├── SecuritySSHPage.vue
│   │   │   └── SecurityTokensPage.vue
│   │   │
│   │   └── external/
│   │       ├── CloudflarePage.vue
│   │       └── LitestreamPage.vue
│   │
│   ├── components/
│   │   ├── layout/
│   │   │   ├── AppShell.vue
│   │   │   ├── Navbar.vue
│   │   │   ├── Sidebar.vue
│   │   │   ├── PageHeader.vue
│   │   │   ├── SitesLayout.vue
│   │   │   ├── SystemLayout.vue
│   │   │   ├── AppsLayout.vue
│   │   │   ├── SecurityLayout.vue
│   │   │   └── ExternalLayout.vue
│   │   │
│   │   ├── ui/
│   │   │   ├── Button.vue
│   │   │   ├── Card.vue
│   │   │   ├── Badge.vue
│   │   │   ├── Input.vue
│   │   │   ├── Modal.vue
│   │   │   ├── Dropdown.vue
│   │   │   ├── Tabs.vue
│   │   │   ├── Spinner.vue
│   │   │   ├── Skeleton.vue
│   │   │   ├── Chart.vue
│   │   │   ├── Sparkline.vue
│   │   │   ├── SystemInfo.vue
│   │   │   ├── Breadcrumbs.vue
│   │   │   ├── SectionHeader.vue
│   │   │   ├── Terminal.vue
│   │   │   ├── Datamap.vue
│   │   │   ├── WorldMap.vue
│   │   │   ├── VisitorMap.vue
│   │   │   └── VisitorLocations.vue
│   │   │
│   │   ├── skeletons/
│   │   │   ├── DashboardSkeleton.vue
│   │   │   ├── SitesSkeleton.vue
│   │   │   └── LoginSkeleton.vue
│   │   │
│   │   └── PlaceholderPage.vue
│   │
│   ├── styles/
│   │   ├── globals.css         # Global styles
│   │   └── preline-overrides.css
│   │
│   ├── data/
│   │   └── mapData.js          # Map coordinates
│   │
│   └── types/
│       ├── models.ts           # Data models
│       └── api.ts              # API types
│
└── packages/
    └── fazt-ui/                # Design tokens (copy as-is)
        ├── index.css
        ├── tokens.css
        ├── base.css
        └── utilities.css
```

---

## fazt-sdk/vue.js - Vue Composables

The key improvement. Single source of truth for all API calls with proper caching.

```javascript
// src/lib/fazt-sdk/vue.js
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { computed } from 'vue'
import { createClient, mockAdapter } from './index.js'
import { useMockStore } from '../../stores/mock.js'

// Get client (mock or real based on store)
export function useClient() {
  const mockStore = useMockStore()
  return computed(() => {
    return mockStore.enabled
      ? createClient({ adapter: mockAdapter })
      : createClient()
  })
}

// Apps
export function useApps() {
  const client = useClient()
  return useQuery({
    queryKey: ['apps'],
    queryFn: () => client.value.apps.list(),
    staleTime: 5 * 60 * 1000,  // 5 min cache
  })
}

export function useApp(id) {
  const client = useClient()
  return useQuery({
    queryKey: ['apps', id],
    queryFn: () => client.value.apps.get(id),
    enabled: computed(() => !!id.value),
  })
}

// System
export function useHealth() {
  const client = useClient()
  return useQuery({
    queryKey: ['system', 'health'],
    queryFn: () => client.value.system.health(),
    refetchInterval: 30000,  // Refresh every 30s
  })
}

// Logs with filters
export function useLogs(filters) {
  const client = useClient()
  return useQuery({
    queryKey: ['logs', filters],
    queryFn: () => client.value.logs.list(filters.value),
    staleTime: 60 * 1000,  // 1 min cache
  })
}

export function useLogsStats(filters) {
  const client = useClient()
  return useQuery({
    queryKey: ['logs', 'stats', filters],
    queryFn: () => client.value.logs.stats(filters.value),
  })
}

// Auth
export function useSession() {
  const client = useClient()
  return useQuery({
    queryKey: ['auth', 'session'],
    queryFn: () => client.value.auth.session(),
    retry: false,
    staleTime: Infinity,  // Don't refetch automatically
  })
}

export function useLogout() {
  const client = useClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => client.value.auth.signOut(),
    onSuccess: () => {
      queryClient.clear()  // Clear all cache on logout
    },
  })
}

// Aliases
export function useAliases() {
  const client = useClient()
  return useQuery({
    queryKey: ['aliases'],
    queryFn: () => client.value.aliases.list(),
  })
}

// Stats
export function useStats() {
  const client = useClient()
  return useQuery({
    queryKey: ['stats'],
    queryFn: () => client.value.stats.analytics(),
    staleTime: 60 * 1000,
  })
}
```

---

## Porting Strategy

### Phase 1: Foundation (Files: 15)
1. Project setup (Vite + Vue + TypeScript)
2. Copy fazt-ui CSS tokens
3. Copy fazt-sdk, add vue.js composables
4. Set up Pinia stores (auth, theme, toast, mock)
5. Set up Vue Router with all routes
6. Port AppShell, Navbar, Sidebar

### Phase 2: Core UI Components (Files: 20)
Port all `components/ui/` - these are pure presentational:
- Button, Card, Badge, Input
- Modal, Dropdown, Tabs
- Spinner, Skeleton
- Chart, Sparkline
- SystemInfo, Breadcrumbs, etc.

### Phase 3: Layout Components (Files: 6)
- SitesLayout, SystemLayout, AppsLayout
- SecurityLayout, ExternalLayout
- PageHeader

### Phase 4: Main Pages (Files: 5)
- LoginPage (with auth flow)
- DashboardPage (with data fetching)
- ProfilePage
- NotFoundPage
- PlaceholderPage

### Phase 5: Sites Pages (Files: 5)
- SitesPage (list)
- SitesAnalyticsPage (charts)
- SitesDomainsPage
- CreateSitePage
- SiteDetailPage (largest, most complex)

### Phase 6: System Pages (Files: 7)
- SystemStatsPage
- SystemLimitsPage
- SystemLogsPage (proper data fetching!)
- SystemBackupPage
- SystemHealthPage
- SystemSettingsPage
- DesignSystemPage

### Phase 7: Remaining Pages (Files: 10)
- Apps pages (6)
- Security pages (2)
- External pages (2)

---

## Component Porting Pattern

### React TSX → Vue SFC

**React (Button.tsx):**
```tsx
interface ButtonProps {
  variant?: 'primary' | 'secondary' | 'ghost';
  children: React.ReactNode;
  onClick?: () => void;
}

export function Button({ variant = 'primary', children, onClick }: ButtonProps) {
  return (
    <button
      className={`btn btn-${variant}`}
      onClick={onClick}
    >
      {children}
    </button>
  );
}
```

**Vue (Button.vue):**
```vue
<script setup>
defineProps({
  variant: {
    type: String,
    default: 'primary',
    validator: (v) => ['primary', 'secondary', 'ghost'].includes(v)
  }
})

const emit = defineEmits(['click'])
</script>

<template>
  <button
    :class="`btn btn-${variant}`"
    @click="emit('click')"
  >
    <slot />
  </button>
</template>
```

### React Hooks → Vue Composables

**React:**
```tsx
const [data, setData] = useState(null);
const [loading, setLoading] = useState(true);

useEffect(() => {
  fetchData().then(setData).finally(() => setLoading(false));
}, []);
```

**Vue (with vue-query):**
```vue
<script setup>
import { useApps } from '@/lib/fazt-sdk/vue'

const { data, isLoading } = useApps()
</script>

<template>
  <div v-if="isLoading">Loading...</div>
  <div v-else>{{ data }}</div>
</template>
```

### React Context → Pinia Store

**React Context:**
```tsx
const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  return (
    <AuthContext.Provider value={{ user, setUser }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
```

**Pinia Store:**
```javascript
// stores/auth.js
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const isAuthenticated = computed(() => !!user.value)

  function setUser(newUser) {
    user.value = newUser
  }

  function logout() {
    user.value = null
  }

  return { user, isAuthenticated, setUser, logout }
})
```

---

## CSS Strategy

### Keep Unchanged
- `packages/fazt-ui/tokens.css` - All design tokens
- `packages/fazt-ui/base.css` - Base styles
- Tailwind config - Same classes
- All component Tailwind classes - Direct copy

### Minor Adjustments
- `index.css` - Clean up Vite defaults
- Remove React-specific styles if any

---

## Testing Strategy

1. **Visual comparison** - Side-by-side with React version
2. **Route-by-route** - Each page must match exactly
3. **Data fetching** - Verify no count=0 flicker
4. **Mock mode** - Toggle works correctly
5. **Auth flow** - Login/logout works
6. **Theme toggle** - Light/dark works

---

## Success Criteria

1. **Visual parity** - Exact same look as React version
2. **No flicker** - Data loads with proper loading states
3. **Caching works** - Navigate back doesn't re-fetch
4. **fazt-sdk only** - No duplicate API code
5. **Deployable** - `fazt @local app deploy admin-vue` works

---

## Execution Order

```
[ ] 1. Create admin-vue directory
[ ] 2. npm init + install deps (vue, vue-router, pinia, vue-query, tailwind)
[ ] 3. Copy fazt-ui CSS
[ ] 4. Copy fazt-sdk, create vue.js
[ ] 5. Set up main.js, App.vue
[ ] 6. Set up router with all routes
[ ] 7. Create Pinia stores
[ ] 8. Port UI components (20 files)
[ ] 9. Port layout components (6 files)
[ ] 10. Port LoginPage + auth flow
[ ] 11. Port DashboardPage
[ ] 12. Port Sites pages (5 files)
[ ] 13. Port System pages (7 files)
[ ] 14. Port remaining pages (10 files)
[ ] 15. Test all routes
[ ] 16. Test mock mode
[ ] 17. Deploy to local
[ ] 18. Replace admin/ with admin-vue/
```

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| D3/Datamaps integration breaks | Test early, may need vue-specific wrapper |
| ApexCharts issues | Use vue3-apexcharts wrapper |
| Complex state in SiteDetail | Break into smaller composables |
| Auth redirect issues | Test OAuth flow on remote early |
