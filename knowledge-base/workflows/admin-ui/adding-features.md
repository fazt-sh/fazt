---
title: Adding Admin UI Features
description: Step-by-step workflow for implementing new features in the Admin UI
updated: 2026-02-01
category: workflows
tags: [admin-ui, development, features, backend-first, design-system]
---

# Adding Admin UI Features

This guide walks through the **backend-first** workflow for adding features to the Admin UI.

## The Golden Rules

> **1. Never build UI for non-existent backend APIs.**
>
> **2. Always use the design system patterns.**

Always validate backend support before implementing UI features, and use the panel-based layout system for all pages. See [design-system.md](design-system.md) for the foundational patterns.

## Implementation Flow

```
┌─────────────────────────────────────────────────────┐
│  1. Validate Backend API Exists                     │
│     Check handlers in internal/handlers/            │
└──────────────────┬──────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────────────────┐
│  2. Add/Verify SDK Method                           │
│     Define in admin/packages/fazt-sdk/              │
└──────────────────┬──────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────────────────┐
│  3. Add Data Store Function                         │
│     Load/update in admin/src/stores/data.js         │
└──────────────────┬──────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────────────────┐
│  4. Implement UI Component                          │
│     Page/modal in admin/src/pages/                  │
└──────────────────┬──────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────────────────┐
│  5. Test in Mock Mode                               │
│     Add fixtures, verify UI                         │
└──────────────────┬──────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────────────────┐
│  6. Test in Real Mode                               │
│     Remove ?mock=true, verify against real API      │
└─────────────────────────────────────────────────────┘
```

## Step-by-Step Example: "Add Restart App Button"

Let's walk through adding a feature to restart apps.

### Step 1: Validate Backend API

**Check if the endpoint exists:**

```bash
# Search for restart handler
grep -rn "restart" internal/handlers/

# Check API routes
grep -rn "POST.*apps.*restart" internal/
```

**Options:**

**A) API Exists ✅**
```bash
# Found in internal/handlers/apps_handler.go:
# POST /api/apps/:id/restart
```
→ Continue to Step 2

**B) API Doesn't Exist ❌**
```
No matches found
```
→ **STOP**: Implement backend first, or show "Coming Soon"

**If API doesn't exist, you have two options:**

1. **Implement backend first** (preferred):
   - Add handler in `internal/handlers/`
   - Wire to app lifecycle
   - Test thoroughly
   - Then return to Step 2

2. **Show "Coming Soon" in UI**:
   - Add button with disabled state
   - Show tooltip: "Feature coming soon"
   - Professional placeholder
   - No broken functionality

### Step 2: Add/Verify SDK Method

**File**: `admin/packages/fazt-sdk/index.js`

```javascript
apps: {
  // ... existing methods

  /** Restart app */
  restart: (id) => http.post(`/api/apps/${id}/restart`, {})
}
```

**Test the SDK method in browser console:**

```javascript
const client = window.__fazt_admin.client
await client.apps.restart('app_01hw3xyz')
```

### Step 3: Add Data Store Function

**File**: `admin/src/stores/data.js`

```javascript
/**
 * Restart app
 * @param {import('../../packages/fazt-sdk/index.js').createClient} client
 * @param {string} id
 */
export async function restartApp(client, id) {
  try {
    await client.apps.restart(id)

    // Optionally reload app detail to show new status
    if (currentApp.get().id === id) {
      await loadAppDetail(client, id)
    }

    notify({ type: 'success', message: 'App restarted' })
    return true
  } catch (err) {
    notify({ type: 'error', message: err.message })
    return false
  }
}
```

**Export the function:**

```javascript
export { restartApp }
```

### Step 4: Implement UI Component

**File**: `admin/src/pages/apps.js` (in detail view)

**Use the design system patterns** from [design-system.md](design-system.md):

```javascript
// Import the new function
import { restartApp } from '../stores/data.js'

// Use panel-based layout structure
container.innerHTML = `
  <div class="design-system-page">
    <div class="content-container">
      <div class="content-scroll">
        <div class="panel-group">
          <div class="panel-group-card card">
            <header class="panel-group-header">
              <button class="collapse-toggle">
                <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                <span class="text-heading text-primary">Actions</span>
              </button>
            </header>
            <div class="panel-group-body">
              <button id="restart-btn" class="btn btn-secondary text-label">
                <i data-lucide="refresh-cw" class="w-4 h-4"></i>
                Restart
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
`

// Add event handler:
container.querySelector('#restart-btn')?.addEventListener('click', async () => {
  if (confirm(`Restart "${app.name}"? The app will be briefly unavailable.`)) {
    const btn = container.querySelector('#restart-btn')
    btn.disabled = true
    btn.innerHTML = 'Restarting...'

    await restartApp(client, appId)

    btn.disabled = false
    btn.innerHTML = '<i data-lucide="refresh-cw" class="w-4 h-4"></i> Restart'
    if (window.lucide) window.lucide.createIcons()
  }
})
```

### Step 5: Add Mock Fixture (If Needed)

**File**: `admin/packages/fazt-sdk/mock.js`

```javascript
const routes = {
  // ... existing routes

  'POST /api/apps/:id/restart': (params) => {
    const app = apps.find(a => a.id === params.id || a.name === params.id)
    if (!app) throw { code: 'APP_NOT_FOUND', message: 'App not found', status: 404 }
    return { message: 'App restarted', name: app.name, id: app.id }
  }
}
```

### Step 6: Test Both Modes

**Mock Mode:**
```
http://admin-ui.192.168.64.3.nip.io:8080?mock=true
```
- Click restart button
- Verify success message
- Check console for API call

**Real Mode:**
```
http://admin-ui.192.168.64.3.nip.io:8080
```
- Click restart button
- Verify actual app restarts
- Check for errors

## Common Scenarios

### Scenario A: Adding a New Page

**Example: Add "Logs" page**

1. **Validate**: Does `GET /api/logs` exist?
2. **SDK**: Add `client.logs.list()`
3. **Store**: Add `loadLogs(client)` and `logs` store
4. **Route**: Add to `src/routes.js`:
   ```javascript
   { name: 'logs', path: '/logs', meta: { title: 'Logs' } }
   ```
5. **Page**: Create `src/pages/logs.js` using **design system structure**:
   ```javascript
   export function render(container, ctx) {
     function update() {
       container.innerHTML = `
         <div class="design-system-page">
           <div class="content-container">
             <div class="content-scroll">
               <!-- Panel groups here -->
             </div>
           </div>
         </div>
       `
       setupCollapseHandlers(container)
       if (window.lucide) window.lucide.createIcons()
     }
     // Subscribe, render, return cleanup
   }
   ```
   See [design-system.md](design-system.md) for full patterns.
6. **Import**: Add to `index.html`:
   ```javascript
   import * as logsPage from './src/pages/logs.js'
   const pages = { ..., logs: logsPage }
   ```
7. **Nav**: Add to sidebar in `index.html`

### Scenario B: Adding a Modal Dialog

**Example: Add "Create Alias" modal**

1. **Validate**: Does `POST /api/aliases` exist?
2. **SDK**: Verify `client.aliases.create(subdomain, type, options)`
3. **Store**: Add `createAlias(client, data)` function
4. **Modal HTML**: Add to `index.html`:
   ```html
   <div id="createAliasModal" class="fixed inset-0 z-50 hidden...">
     <!-- Modal content -->
   </div>
   ```
5. **Open Handler**: In aliases page:
   ```javascript
   document.getElementById('createAliasModal').classList.remove('hidden')
   ```
6. **Form Submit**: Call `createAlias()`, reload aliases

### Scenario C: Adding a Field to Existing Data

**Example: Show "last deployed by" in app list**

1. **Validate**: Does API return `last_deploy_user` field?
   ```bash
   # Check handler response
   grep -A 20 "func.*GetApp" internal/handlers/apps_handler.go
   ```
2. **If NO**: Backend needs to add this field first
3. **If YES**: Just use it in UI:
   ```javascript
   <span>${app.last_deploy_user || 'Unknown'}</span>
   ```
4. **Update Mock**: Add field to `fixtures/apps.json`

### Scenario D: Backend Doesn't Exist Yet

**Example: "Deploy from Git URL" feature**

**Option 1: Show "Coming Soon"**
```javascript
<button class="btn btn-primary" disabled title="Coming soon">
  <i data-lucide="git-branch" class="w-4 h-4"></i>
  Deploy from Git
</button>
```

**Option 2: Show Modal with Alternative**
```javascript
<div class="modal">
  <h2>Deploy from Git</h2>
  <p>Web-based Git deployment is coming soon.</p>
  <div class="mt-4">
    <strong>Use CLI instead:</strong>
    <code>fazt app deploy git@github.com:user/repo.git</code>
  </div>
</div>
```

**Option 3: Implement Backend First**
→ See [../fazt-binary/adding-apis.md](../fazt-binary/adding-apis.md)

## Pre-Implementation Checklist

Before writing any UI code, verify:

- [ ] Backend endpoint exists in `internal/handlers/`
- [ ] Endpoint is documented or code is readable
- [ ] Response structure is known (check handler or API call)
- [ ] SDK method exists or can be added
- [ ] Mock adapter can simulate the behavior
- [ ] Error cases are handled in backend
- [ ] Authentication/authorization is considered

**If any checkbox is unchecked**: Clarify with backend first.

## Push-Back Examples

### ❌ Bad: Implement Without Validation

```
User: "Add a button to scale apps up/down"
Agent: "Sure! Adding scale buttons..."
→ Implements UI → Doesn't work in real mode
```

### ✅ Good: Validate First

```
User: "Add a button to scale apps up/down"
Agent: "Let me check if the backend supports this..."

[Checks internal/handlers/]

Agent: "⚠️ The backend doesn't have a scaling API yet.

       Options:
       1. Implement backend scaling API first (recommended)
       2. Show 'Coming Soon' button as placeholder
       3. Add to feature backlog

       Which approach do you prefer?"
```

## Testing Workflow

### 1. Build & Deploy

```bash
cd admin && npm run build
fazt app deploy admin --to local --name admin-ui
```

### 2. Test Mock Mode

```bash
open http://admin-ui.192.168.64.3.nip.io:8080?mock=true
```

- Verify UI renders correctly
- Test interactions
- Check console for errors
- Verify mock data structure

### 3. Test Real Mode

```bash
open http://admin-ui.192.168.64.3.nip.io:8080
```

- Verify API calls succeed
- Check real data displays correctly
- Test error handling
- Verify loading states

### 4. Browser Console Debugging

```javascript
// Access app internals
const { router, client, apps, aliases } = window.__fazt_admin

// Test SDK method
await client.apps.list()

// Check state
apps.get()

// Trigger navigation
router.push('/apps/app_01hw3xyz')
```

## Common Pitfalls

### ❌ Pitfall 1: Mock-Only Features

```javascript
// Bad: Feature only works in mock mode
const mockData = { deployment_count: 5 }  // Field doesn't exist in real API
```

### ❌ Pitfall 2: Bypassing SDK

```javascript
// Bad: Direct API call
const response = await fetch('/api/apps')
```

### ❌ Pitfall 3: Assumed API Structure

```javascript
// Bad: Assuming pagination exists
const { apps, total, page } = await client.apps.list()
// Real API might just return: apps[]
```

### ❌ Pitfall 4: No Error Handling

```javascript
// Bad: No try-catch
const result = await client.apps.delete(id)
```

## Best Practices

✅ **Always use SDK methods**
✅ **Update stores, let UI react**
✅ **Handle loading states**
✅ **Show user-friendly errors**
✅ **Confirm destructive actions**
✅ **Disable buttons during operations**
✅ **Match mock data to real structure**
✅ **Test both modes before deploying**

## Next Steps

- Read [architecture.md](architecture.md) to understand state management
- Read [testing.md](testing.md) for testing strategies
- Check [checklist.md](checklist.md) before each feature
- See [../fazt-sdk/extending.md](../fazt-sdk/extending.md) for SDK details
