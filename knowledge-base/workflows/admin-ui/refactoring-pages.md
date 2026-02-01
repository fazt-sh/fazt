---
title: Refactoring Admin Pages to Design System
description: Step-by-step guide for converting existing pages to panel-based layout
updated: 2026-02-01
category: workflows
tags: [admin-ui, refactoring, design-system, migration]
---

# Refactoring Admin Pages to Design System

This guide documents the process of converting existing admin pages to use the panel-based design system.

## Overview

**Goal:** Convert pages to use `.design-system-page > .content-container > .content-scroll` structure with panel groups for collapsible sections.

**When to refactor:**
- Page uses old layout without design system classes
- Content would benefit from collapsible sections
- Page needs responsive edge-to-edge mobile layout
- Migrating from one-off layouts to consistent patterns

## Before You Start

1. **Read the current page** - Understand existing functionality and structure
2. **Check design-system.md** - Refresh on layout patterns and classes
3. **Identify sections** - What content should be in panels? What stays fixed?
4. **Plan collapse states** - Which sections should be collapsible?

## Refactoring Steps

### 1. Add UI State Helpers

Every page needs these functions for collapse state persistence:

```javascript
/**
 * Get UI state from localStorage
 */
function getUIState(key, defaultValue = false) {
  try {
    const state = JSON.parse(localStorage.getItem('fazt.web.ui.state') || '{}')
    return state[key] !== undefined ? state[key] : defaultValue
  } catch {
    return defaultValue
  }
}

/**
 * Set UI state to localStorage
 */
function setUIState(key, value) {
  try {
    const state = JSON.parse(localStorage.getItem('fazt.web.ui.state') || '{}')
    state[key] = value
    localStorage.setItem('fazt.web.ui.state', JSON.stringify(state))
  } catch (e) {
    console.error('Failed to save UI state:', e)
  }
}
```

**Location:** Add at the top of the page file, before the `render()` function.

### 2. Wrap Page in Design System Structure

Replace the old container with:

```javascript
container.innerHTML = `
  <div class="design-system-page">
    <div class="content-container">
      <div class="content-scroll">
        <!-- Page content here -->
      </div>
    </div>
  </div>
`
```

**Benefits:**
- Automatic responsive behavior
- Edge-to-edge on mobile
- Centered content on desktop
- Consistent max-width

### 3. Convert Sections to Panel Groups

Identify content sections and wrap them in panel groups:

```javascript
<!-- Panel Group: Section Name -->
<div class="panel-group ${collapsed ? 'collapsed' : ''}">
  <div class="panel-group-card card">
    <header class="panel-group-header" data-group="section-id">
      <button class="collapse-toggle">
        <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
        <span class="text-heading text-primary">Section Title</span>
        <span class="text-caption text-faint ml-auto hide-mobile">Optional meta</span>
      </button>
    </header>
    <div class="panel-group-body">
      <!-- Section content -->
    </div>
  </div>
</div>
```

**Key Points:**
- `data-group="section-id"` - Used to save/load collapse state
- `${collapsed ? 'collapsed' : ''}` - Initial state from `getUIState()`
- `.hide-mobile` - Hides optional metadata on small screens

### 4. Add Collapse Handlers

After rendering, set up collapse toggle handlers:

```javascript
container.querySelectorAll('.collapse-toggle').forEach(toggle => {
  toggle.addEventListener('click', () => {
    const header = toggle.closest('.panel-group-header')
    const group = header.dataset.group
    const panelGroup = header.closest('.panel-group')
    const isCollapsed = panelGroup.classList.toggle('collapsed')
    setUIState(`page.${group}.collapsed`, isCollapsed)
  })
})
```

**Important:** This must run after HTML is inserted into the DOM.

### 5. Update localStorage Usage

Replace direct localStorage calls with UI state helpers:

**Before:**
```javascript
let viewMode = localStorage.getItem('apps-view-mode') || 'cards'
// ...
localStorage.setItem('apps-view-mode', 'cards')
```

**After:**
```javascript
let viewMode = getUIState('apps.view', 'cards')
// ...
setUIState('apps.view', 'cards')
```

**Benefits:**
- Centralized state in `fazt.web.ui.state`
- Consistent naming convention
- Easier debugging and state inspection

### 6. Re-render Icons

After any HTML update:

```javascript
if (window.lucide) window.lucide.createIcons()
```

This re-renders all `data-lucide` icons in the updated DOM.

## Common Patterns

### Page Header (Non-Collapsible)

```javascript
<!-- Page Header -->
<div class="flex items-center justify-between mb-4">
  <div>
    <h1 class="text-title text-primary">Page Title</h1>
    <p class="text-caption text-muted">Subtitle or count</p>
  </div>
  <div class="flex items-center gap-2">
    <!-- Actions, filters, buttons -->
  </div>
</div>
```

Headers stay fixed at the top, outside panel groups.

### Stats Grid

Use `.panel-grid` with grid size classes:

```javascript
<div class="panel-grid grid-3">
  <div class="stat-card card">
    <div class="stat-card-header">
      <span class="text-micro text-muted">Label</span>
      <i data-lucide="icon" class="w-4 h-4 text-faint"></i>
    </div>
    <div class="stat-card-value text-display mono text-primary">123</div>
  </div>
  <!-- More stat cards -->
</div>
```

Responsive: Desktop = 3 cols, Tablet = 2 cols, Mobile = 1 col.

### Tables in Panels

Remove padding from panel-group-body:

```javascript
<div class="panel-group-body" style="padding: 0">
  <div class="table-container">
    <table>
      <!-- Table content -->
    </table>
  </div>
</div>
```

This prevents double padding around tables.

### Detail Pages with Dynamic IDs

Use page-specific keys for collapse state:

```javascript
const detailsCollapsed = getUIState(`apps.detail.${appId}.details.collapsed`, false)
```

This preserves collapse state per item.

## Example: Apps Page Refactor

### Before (Old Layout)

```javascript
function renderList(container, ctx) {
  function init() {
    container.innerHTML = `
      <div class="flex flex-col h-full overflow-hidden">
        <div class="flex items-center justify-between mb-4">
          <!-- Header -->
        </div>
        <div id="apps-content" class="flex-1 overflow-auto scroll-panel"></div>
      </div>
    `
  }
}
```

**Issues:**
- No design system structure
- Direct localStorage usage
- Not edge-to-edge on mobile

### After (Design System)

```javascript
function getUIState(key, defaultValue = false) { /* ... */ }
function setUIState(key, value) { /* ... */ }

function renderList(container, ctx) {
  let viewMode = getUIState('apps.view', 'cards')

  function init() {
    container.innerHTML = `
      <div class="design-system-page">
        <div class="content-container">
          <div class="content-scroll">

            <!-- Page Header -->
            <div class="flex items-center justify-between mb-4">
              <!-- Header content -->
            </div>

            <!-- Apps Content -->
            <div id="apps-content"></div>

          </div>
        </div>
      </div>
    `

    if (window.lucide) window.lucide.createIcons()
    // Event handlers...
  }
}
```

**Improvements:**
- Design system wrapper
- UI state helpers
- Consistent with other pages

### Detail View with Panels

```javascript
const detailsCollapsed = getUIState(`apps.detail.${appId}.details.collapsed`, false)

container.innerHTML = `
  <div class="design-system-page">
    <div class="content-container">
      <div class="content-scroll">

        <!-- Page Header -->
        <div class="flex items-center justify-between mb-4">
          <!-- Back button, title, actions -->
        </div>

        <!-- Stats Grid -->
        <div class="panel-grid grid-3 mb-4">
          <!-- Stat cards -->
        </div>

        <!-- Panel Group: Details -->
        <div class="panel-group ${detailsCollapsed ? 'collapsed' : ''}">
          <div class="panel-group-card card">
            <header class="panel-group-header" data-group="details">
              <button class="collapse-toggle">
                <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                <span class="text-heading text-primary">Details</span>
              </button>
            </header>
            <div class="panel-group-body">
              <!-- Details content -->
            </div>
          </div>
        </div>

      </div>
    </div>
  </div>
`

// Setup collapse handlers
container.querySelectorAll('.collapse-toggle').forEach(toggle => {
  toggle.addEventListener('click', () => {
    const header = toggle.closest('.panel-group-header')
    const group = header.dataset.group
    const panelGroup = header.closest('.panel-group')
    const isCollapsed = panelGroup.classList.toggle('collapsed')
    setUIState(`apps.detail.${appId}.${group}.collapsed`, isCollapsed)
  })
})
```

## UI State Key Naming

Use consistent naming for UI state keys:

```
fazt.web.ui.state = {
  // Page-level settings
  "apps.view": "cards",                    // View mode (cards vs list)
  "sidebar.collapsed": false,              // Sidebar state

  // Panel collapse states
  "dashboard.stats.collapsed": false,
  "dashboard.apps.collapsed": false,

  // Detail page states (per-item)
  "apps.detail.abc123.details.collapsed": false,
  "apps.detail.abc123.aliases.collapsed": false,
  "apps.detail.abc123.files.collapsed": true
}
```

**Convention:**
- `page.setting` - Page-level settings
- `page.section.collapsed` - Panel collapse states
- `page.detail.id.section.collapsed` - Detail page panels

## Checklist

After refactoring a page:

- [ ] Added `getUIState()` and `setUIState()` helpers
- [ ] Wrapped page in `.design-system-page > .content-container > .content-scroll`
- [ ] Converted sections to panel groups where appropriate
- [ ] Set up collapse handlers for all toggleable panels
- [ ] Replaced direct localStorage with `getUIState()`/`setUIState()`
- [ ] Icons re-rendered with `lucide.createIcons()`
- [ ] Tested on mobile (edge-to-edge layout)
- [ ] Tested collapse state persistence (refresh page)
- [ ] Tested on tablet and desktop breakpoints

## Common Mistakes

### ❌ Forgetting collapse handlers

Panel groups won't collapse without click handlers.

**Fix:** Add collapse handler setup after rendering.

### ❌ Wrong data-group attribute

```html
<header class="panel-group-header">  <!-- Missing data-group -->
```

**Fix:** Always include `data-group="section-id"` on headers.

### ❌ Icons not rendering

```javascript
container.innerHTML = `...`
// Missing: if (window.lucide) window.lucide.createIcons()
```

**Fix:** Call `createIcons()` after every HTML update.

### ❌ Table padding issues

Tables in panels have double padding if body has default padding.

**Fix:** Use `style="padding: 0"` on `.panel-group-body` for tables.

### ❌ Mobile layout not edge-to-edge

Page uses custom padding instead of design system structure.

**Fix:** Use `.design-system-page` wrapper - CSS handles responsive padding.

## Testing

After refactoring, test:

1. **Desktop view** - Panels render correctly, collapse/expand works
2. **Tablet view** - 16px padding, sidebar toggle
3. **Mobile view** - Edge-to-edge panels, no horizontal padding
4. **State persistence** - Collapse a panel, refresh page, state preserved
5. **View mode persistence** - Change view mode, refresh page, mode preserved
6. **Icons** - All lucide icons render after collapse/expand

## Related Files

| File | Purpose |
|------|---------|
| `knowledge-base/workflows/admin-ui/design-system.md` | Layout patterns reference |
| `admin/src/pages/dashboard.js` | Reference implementation |
| `admin/src/pages/apps.js` | Example refactor (completed) |
| `admin/index.html` | CSS design system |

## Next Pages to Refactor

1. **Aliases page** - High priority, currently placeholder
2. **System page** - Needs panel groups for metrics
3. **Settings page** - Needs panel groups for config sections

Use this guide as a template for each page refactor.
