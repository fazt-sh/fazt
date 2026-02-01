---
title: Admin UI Architecture
description: State management, data flow, and component model for the Fazt Admin UI
updated: 2026-02-01
category: workflows
tags: [admin-ui, architecture, state-management, data-flow, design-system]
---

# Admin UI Architecture

The Fazt Admin UI is a single-page application built with vanilla JavaScript, using reactive state stores and a clean data flow model.

> **UI Foundation**: All pages use the panel-based layout system. See [design-system.md](design-system.md) for CSS patterns, responsive breakpoints, and component structures.

## Overview

```
┌─────────────────────────────────────────────────────┐
│                   Fazt Backend                       │
│              (Go + SQLite/PostgreSQL)                │
└───────────────────┬─────────────────────────────────┘
                    │ HTTP/JSON API
                    ↓
┌─────────────────────────────────────────────────────┐
│                   fazt-sdk                           │
│         (JavaScript API Client)                      │
│  • Real adapter (fetch)                              │
│  • Mock adapter (fixtures)                           │
└───────────────────┬─────────────────────────────────┘
                    │ Async calls
                    ↓
┌─────────────────────────────────────────────────────┐
│                 Data Stores                          │
│           (Reactive State)                           │
│  • apps                                              │
│  • aliases                                           │
│  • currentApp                                        │
│  • auth                                              │
│  • health, stats                                     │
└───────────────────┬─────────────────────────────────┘
                    │ Subscribe
                    ↓
┌─────────────────────────────────────────────────────┐
│               UI Components                          │
│         (Pages + Render Functions)                   │
│  • Read from stores                                  │
│  • Call SDK methods                                  │
│  • Auto-update on changes                            │
└─────────────────────────────────────────────────────┘
```

## Core Concepts

### 1. Reactive State Stores

**Location**: `admin/src/stores/data.js`

State stores are reactive containers that notify subscribers when data changes:

```javascript
// Define store
export const apps = list([])

// Subscribe to changes
apps.subscribe(() => {
  console.log('Apps updated:', apps.get())
  renderAppsContent()
})

// Update state (triggers all subscribers)
apps.set([...newApps])
```

**Available Stores**:

| Store | Type | Purpose |
|-------|------|---------|
| `apps` | list | All deployed apps |
| `aliases` | list | All subdomain aliases |
| `currentApp` | map | Currently viewed app detail |
| `auth` | map | Authentication state |
| `health` | map | System health metrics |
| `stats` | map | Overview statistics |
| `loading` | map | Loading states by key |
| `error` | string | Global error state |

### 2. Data Flow (Unidirectional)

```
User Action → SDK Method → Update Store → UI Re-renders
```

**Example: Loading Apps**

```javascript
// 1. User navigates to /apps
// 2. Page calls load function
await loadApps(client)

// 3. Load function calls SDK
export async function loadApps(client) {
  loading.setKey('apps', true)
  const data = await client.apps.list()  // ← SDK
  apps.set(data || [])                   // ← Store
  loading.setKey('apps', false)
}

// 4. Store update triggers subscribers
apps.subscribe(renderAppsContent)  // ← UI updates
```

**Example: Deleting an App**

```javascript
// 1. User clicks delete button
// 2. Component calls SDK method
const success = await deleteApp(client, appId)

// 3. SDK method updates store
export async function deleteApp(client, id) {
  await client.apps.delete(id)     // ← API call
  apps.remove(app => app.id === id) // ← Store update
  return true
}

// 4. Store update triggers re-render (app removed from list)
```

### 3. Component Model

**Location**: `admin/src/pages/`

Pages export a `render()` function that:
- Receives container element and context
- Sets up subscriptions
- Returns cleanup function

```javascript
export function render(container, ctx) {
  const { router, client, params, refresh } = ctx

  // Render function
  function renderContent() {
    const data = apps.get()  // Read from store
    container.innerHTML = `...`
    // Attach event handlers
  }

  // Subscribe to data changes
  const unsub = apps.subscribe(renderContent)

  // Initial render
  renderContent()

  // Return cleanup (unsubscribe)
  return () => unsub()
}
```

### 4. SDK as Single API Layer

**Location**: `admin/packages/fazt-sdk/`

All backend communication goes through the SDK:

```javascript
// ✅ Correct: Use SDK
const apps = await client.apps.list()

// ❌ Wrong: Direct fetch
const response = await fetch('/api/apps')
```

**SDK Methods Match Backend 1:1**:

| SDK Method | Backend Endpoint | Handler |
|------------|------------------|---------|
| `client.apps.list()` | `GET /api/apps` | `internal/handlers/apps_handler.go` |
| `client.apps.get(id)` | `GET /api/apps/:id` | `internal/handlers/apps_handler.go` |
| `client.aliases.list()` | `GET /api/aliases` | `internal/handlers/aliases_handler.go` |

## State Management Patterns

### Pattern 1: List Data (Apps, Aliases)

```javascript
// Store
export const apps = list([])

// Load
export async function loadApps(client) {
  const data = await client.apps.list()
  apps.set(data || [])
}

// Use in UI
const appList = apps.get()
const filteredApps = appList.filter(app => ...)
```

### Pattern 2: Detail Data (Current App)

```javascript
// Store
export const currentApp = map({
  id: null,
  name: null,
  files: []
})

// Load
export async function loadAppDetail(client, id) {
  const [appData, filesData] = await Promise.all([
    client.apps.get(id),
    client.apps.files(id)
  ])
  currentApp.set({ ...appData, files: filesData })
}

// Use in UI
const app = currentApp.get()
const fileName = app.files[0].path
```

### Pattern 3: Loading States

```javascript
// Store (keyed map)
export const loading = map({})

// Set loading
loading.setKey('apps', true)

// Check loading
const isLoading = loading.getKey('apps')

// Subscribe to specific key
loading.subscribeKey('apps', (isLoading) => {
  console.log('Apps loading:', isLoading)
})
```

### Pattern 4: Derived State

```javascript
// Compute on read, don't store
const appList = apps.get()
const aliasList = aliases.get()

// Map aliases to apps
const appAliases = {}
aliasList.forEach(alias => {
  if (alias.type === 'proxy' && alias.targets?.app_id) {
    const appId = alias.targets.app_id
    if (!appAliases[appId]) appAliases[appId] = []
    appAliases[appId].push(alias.subdomain)
  }
})

// Use derived data
const myAppAliases = appAliases[myApp.id] || []
```

## Routing

**Location**: `admin/src/routes.js`

Uses history mode (clean URLs with SPA support):

```javascript
export const routes = [
  {
    name: 'apps',
    path: '/apps',
    meta: { title: 'Apps' }
  },
  {
    name: 'app-detail',
    path: '/apps/:id',
    meta: { title: 'App Detail', parent: 'apps' }
  }
]
```

**Router API**:

```javascript
// Navigate
router.push('/apps')
router.push('/apps/app_01hw3xyz')

// Get current route
const match = router.current.get()
console.log(match.route.name)    // 'app-detail'
console.log(match.params.id)     // 'app_01hw3xyz'

// Subscribe to route changes
router.current.subscribe((match) => {
  console.log('Route changed:', match.route.name)
})
```

## Modal System

Modals use standard HTML with backdrop overlay:

```javascript
// Open modal
const modal = document.getElementById('myModal')
modal.classList.remove('hidden')
modal.classList.add('flex')

// Close modal
modal.classList.add('hidden')
modal.classList.remove('flex')

// Close on backdrop click
modal.addEventListener('click', (e) => {
  if (e.target === modal) closeModal()
})
```

## Mock Mode

Toggle between real and mock data:

```javascript
// In index.html
const useMock = new URLSearchParams(window.location.search).get('mock') === 'true'
const client = createClient(useMock ? { adapter: mockAdapter } : {})
```

**Mock Fixtures**: `admin/packages/fazt-sdk/fixtures/`
- `apps.json` - Sample app data
- `aliases.json` - Sample alias data
- `health.json` - System health data
- `user.json` - Mock user session

**Rule**: Mock data must exactly match real API response structure.

## File Structure

```
admin/
├── index.html                  # Main entry, shell, modals
├── src/
│   ├── routes.js              # Route definitions
│   ├── stores/
│   │   ├── app.js             # UI state (theme, sidebar)
│   │   └── data.js            # API data state
│   └── pages/
│       ├── dashboard.js       # Dashboard page
│       ├── apps.js            # Apps list + detail
│       ├── aliases.js         # Aliases page
│       ├── system.js          # System health
│       └── settings.js        # Settings page
└── packages/
    ├── zap/                   # Micro framework (router, stores)
    └── fazt-sdk/              # API client
        ├── index.js           # Client + methods
        ├── client.js          # HTTP adapter
        ├── mock.js            # Mock adapter
        └── fixtures/          # Mock data
```

## Key Rules

1. **Never bypass the SDK**: All API calls go through `client.*`
2. **Never mutate stores directly**: Use `.set()`, `.setKey()`, `.remove()`
3. **Always unsubscribe**: Return cleanup functions from render()
4. **Derive, don't duplicate**: Compute from stores, don't cache separately
5. **Mock must match real**: Fixtures = real API structure

## Performance Considerations

- **Selective updates**: Use `subscribeKey()` for specific state slices
- **Debounce input**: Filter inputs update on every keystroke (by design)
- **Pagination**: Not implemented yet (load all, filter client-side)
- **Caching**: No client-side cache, reload on mount

## Next Steps

- Read [design-system.md](design-system.md) for layout patterns and CSS architecture
- Read [adding-features.md](adding-features.md) to implement new features
- Read [testing.md](testing.md) to understand mock vs real mode
- Check [checklist.md](checklist.md) before starting work
