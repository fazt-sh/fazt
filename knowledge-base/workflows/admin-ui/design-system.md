---
title: Admin UI Design System
description: Panel-based layout, responsive patterns, and CSS architecture for Fazt Admin
updated: 2026-02-01
category: workflows
tags: [admin-ui, design-system, css, responsive, layout]
---

# Admin UI Design System

The Fazt Admin UI uses a **panel-based layout system** designed for responsive, maintainable interfaces. This is the foundation for all pages.

## Core Principles

1. **Panel-based architecture** - Self-contained sections that collapse/expand independently
2. **Edge-to-edge mobile** - Maximize real estate on small screens
3. **Progressive enhancement** - Mobile-first, then tablet, then desktop
4. **Single source of truth** - CSS variables for colors, spacing, radii

## Layout Architecture

### Page Structure

```
┌─────────────────────────────────────────────────────────┐
│ main.flex-1.overflow-hidden                             │
├─────────────────────────────────────────────────────────┤
│ header (fixed height: 48px)                             │
├─────────────────────────────────────────────────────────┤
│ #page-content.flex-1.overflow-hidden                    │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ .design-system-page                                 │ │
│ │ ┌─────────────────────────────────────────────────┐ │ │
│ │ │ .content-container (centers content)            │ │ │
│ │ │ ┌─────────────────────────────────────────────┐ │ │ │
│ │ │ │ .content-scroll (scrollable, max-width)     │ │ │ │
│ │ │ │                                             │ │ │ │
│ │ │ │   .panel-group                              │ │ │ │
│ │ │ │   .panel-group                              │ │ │ │
│ │ │ │   .panel-group                              │ │ │ │
│ │ │ │                                             │ │ │ │
│ │ │ └─────────────────────────────────────────────┘ │ │ │
│ │ └─────────────────────────────────────────────────┘ │ │
│ └─────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────┤
│ footer (fixed height: 40px)                             │
└─────────────────────────────────────────────────────────┘
```

### Panel Group Anatomy

```html
<div class="panel-group">                    <!-- Collapse state container -->
  <div class="panel-group-card card">        <!-- Visual card wrapper -->
    <header class="panel-group-header">      <!-- Clickable header -->
      <button class="collapse-toggle">
        <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
        <span class="text-heading text-primary">Section Title</span>
      </button>
    </header>
    <div class="panel-group-body">           <!-- Collapsible content -->
      <!-- Content here -->
    </div>
  </div>
</div>
```

### Collapse States

```css
/* Expanded (default) */
.panel-group { }
.panel-group .chevron { transform: rotate(90deg); }
.panel-group-body { max-height: 3000px; opacity: 1; }

/* Collapsed */
.panel-group.collapsed .chevron { transform: rotate(0deg); }
.panel-group.collapsed .panel-group-body { max-height: 0; opacity: 0; }
```

## Responsive Breakpoints

| Breakpoint | Width | Behavior |
|------------|-------|----------|
| **Mobile** | < 768px | Edge-to-edge panels, no border-radius, hamburger menu |
| **Tablet** | 768-1023px | 16px padding, rounded corners, sidebar toggle |
| **Desktop** | ≥ 1024px | 20px padding, full sidebar, max-width content |

### Responsive CSS Pattern

```css
/* Base styles (desktop) */
.content-scroll {
  padding: 20px;
  max-width: 1280px;
}

/* Tablet */
@media (max-width: 1023px) {
  .content-scroll {
    padding: 16px;
  }
}

/* Mobile - MUST come after tablet rule */
@media (max-width: 767px) {
  .content-scroll {
    padding: 0;  /* Edge-to-edge */
  }
}
```

**Important:** Mobile rules must come AFTER tablet rules in the CSS file, otherwise tablet rules will override them (both media queries match when width < 768px).

## Edge-to-Edge Mobile Design

On mobile (< 768px), panels extend to screen edges:

```css
@media (max-width: 767px) {
  /* Page content: no horizontal padding */
  #page-content {
    padding: 8px 0 !important;
  }

  /* Content scroll: no padding */
  .content-scroll {
    padding: 0;
  }

  /* Cards: flat edges, no side borders */
  .card {
    border-radius: 0;
  }

  .panel-group-card.card {
    border-left: none;
    border-right: none;
  }
}
```

**Result:** Full-width panels with internal padding only.

## CSS Variables

### Colors

```css
:root {
  /* Background layers */
  --bg-0: #f5f5f5;    /* Page background */
  --bg-1: #ffffff;    /* Card background */
  --bg-2: #f9fafb;    /* Subtle highlight */
  --bg-3: #f3f4f6;    /* Darker highlight */

  /* Text hierarchy */
  --text-1: #111827;  /* Primary text */
  --text-2: #374151;  /* Secondary text */
  --text-3: #6b7280;  /* Muted text */
  --text-4: #9ca3af;  /* Faint text */

  /* Borders */
  --border: #e5e7eb;
  --border-subtle: #f3f4f6;

  /* Accent */
  --accent: #f97316;
  --accent-soft: rgba(249, 115, 22, 0.1);

  /* Status */
  --success: #22c55e;
  --warning: #f59e0b;
  --error: #ef4444;
}
```

### Spacing & Radii

```css
:root {
  --radius-sm: 4px;
  --radius-md: 6px;
  --radius-lg: 8px;
  --radius-xl: 12px;
}
```

## Component Patterns

### Stat Card

```html
<div class="stat-card card">
  <div class="stat-card-header">
    <span class="text-micro text-muted">Label</span>
    <i data-lucide="icon" class="w-4 h-4 text-faint"></i>
  </div>
  <div class="stat-card-value text-display mono text-primary">Value</div>
  <div class="stat-card-subtitle text-caption text-muted">Subtitle</div>
</div>
```

### Activity List

```html
<div class="activity-list">
  <div class="activity-item">
    <i data-lucide="icon" class="w-4 h-4 text-muted"></i>
    <span class="text-label text-primary">Event description</span>
    <span class="text-caption text-muted">2h ago</span>
  </div>
</div>
```

### Panel Grid

```html
<div class="panel-grid grid-5">  <!-- 5-column grid -->
  <div class="stat-card card">...</div>
  <div class="stat-card card">...</div>
  <!-- ... -->
</div>
```

Responsive behavior:
- Desktop: 5 columns
- Tablet: 2 columns
- Mobile: 1 column (stacked)

## UI State Management

### Pattern: localStorage for Collapse States

```javascript
// Unified key for all UI state
const UI_STATE_KEY = 'fazt.web.ui.state'

function getUIState(key, defaultValue = false) {
  try {
    const state = JSON.parse(localStorage.getItem(UI_STATE_KEY) || '{}')
    return state[key] !== undefined ? state[key] : defaultValue
  } catch {
    return defaultValue
  }
}

function setUIState(key, value) {
  try {
    const state = JSON.parse(localStorage.getItem(UI_STATE_KEY) || '{}')
    state[key] = value
    localStorage.setItem(UI_STATE_KEY, JSON.stringify(state))
  } catch (e) {
    console.error('Failed to save UI state:', e)
  }
}
```

### Usage in Pages

```javascript
// Read initial state
const statsCollapsed = getUIState('dashboard.stats.collapsed', false)

// Render with state
container.innerHTML = `
  <div class="panel-group ${statsCollapsed ? 'collapsed' : ''}">
    ...
  </div>
`

// Toggle handler
toggle.addEventListener('click', () => {
  const isCollapsed = panelGroup.classList.toggle('collapsed')
  setUIState('dashboard.stats.collapsed', isCollapsed)
})
```

### Key Naming Convention

```
fazt.web.ui.state = {
  "dashboard.stats.collapsed": false,
  "dashboard.apps.collapsed": false,
  "dashboard.activity.collapsed": true,
  "apps.view": "list",  // or "grid"
  "sidebar.collapsed": false
}
```

## Building New Pages

### Step 1: Page Shell

```javascript
export function render(container, ctx) {
  const { router, client } = ctx

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
    // Setup handlers, icons
  }

  // Subscribe to data stores
  const unsub = someStore.subscribe(update)
  update()

  return () => unsub()
}
```

### Step 2: Add Panel Groups

```javascript
container.innerHTML = `
  <div class="design-system-page">
    <div class="content-container">
      <div class="content-scroll">

        <!-- Section 1 -->
        <div class="panel-group ${getUIState('page.section1.collapsed') ? 'collapsed' : ''}">
          <div class="panel-group-card card">
            <header class="panel-group-header" data-group="section1">
              <button class="collapse-toggle">
                <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
                <span class="text-heading text-primary">Section Title</span>
              </button>
            </header>
            <div class="panel-group-body">
              <!-- Content -->
            </div>
          </div>
        </div>

        <!-- Section 2 -->
        <div class="panel-group">
          <!-- ... -->
        </div>

      </div>
    </div>
  </div>
`
```

### Step 3: Collapse Handlers

```javascript
function setupCollapseHandlers(container) {
  container.querySelectorAll('.collapse-toggle').forEach(toggle => {
    toggle.addEventListener('click', () => {
      const header = toggle.closest('.panel-group-header')
      const group = header.dataset.group
      const panelGroup = header.closest('.panel-group')
      const isCollapsed = panelGroup.classList.toggle('collapsed')
      setUIState(`page.${group}.collapsed`, isCollapsed)
    })
  })
}
```

## Grid Layouts

### Panel Grid (Stats)

```css
.panel-grid {
  display: grid;
  gap: 16px;
}

.panel-grid.grid-5 {
  grid-template-columns: repeat(5, 1fr);
}

@media (max-width: 1023px) {
  .panel-grid.grid-5 {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 767px) {
  .panel-grid.grid-5 {
    grid-template-columns: 1fr;
  }
}
```

### General Responsive Grid

```css
.grid-cols-3 {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

@media (max-width: 767px) {
  .grid-cols-3 {
    grid-template-columns: repeat(2, 1fr);
  }
}
```

## Typography Classes

| Class | Usage |
|-------|-------|
| `.text-display` | Large numbers (32px) |
| `.text-heading` | Section titles (14px, 600) |
| `.text-label` | Item names (13px, 500) |
| `.text-caption` | Secondary info (12px) |
| `.text-micro` | Tiny labels (10px) |

Color modifiers: `.text-primary`, `.text-secondary`, `.text-muted`, `.text-faint`

## Icons (Lucide)

```html
<!-- Basic icon -->
<i data-lucide="icon-name" class="w-4 h-4"></i>

<!-- Icon with color -->
<i data-lucide="check" class="w-4 h-4 text-success"></i>

<!-- Re-render after DOM update -->
<script>
  if (window.lucide) window.lucide.createIcons()
</script>
```

## Checklist: New Page

- [ ] Uses `.design-system-page > .content-container > .content-scroll` structure
- [ ] Sections use `.panel-group` with `.panel-group-card.card`
- [ ] Collapse states persisted with `getUIState()` / `setUIState()`
- [ ] Collapse handlers set up on render
- [ ] Icons re-rendered with `lucide.createIcons()`
- [ ] Responsive at all breakpoints (mobile, tablet, desktop)
- [ ] Edge-to-edge on mobile

## Anti-Patterns

### Don't: Use CSS Grid for collapsible layouts

CSS Grid with flex-1 causes gap issues when sections collapse. Use panel-groups instead.

### Don't: Hard-code padding

Use CSS variables or utility classes. Override only in media queries.

### Don't: Put mobile rules before tablet rules

Both `max-width: 767px` and `max-width: 1023px` match on mobile. Later rules win.

### Don't: Store collapse state in component variables

Use localStorage via `getUIState()`/`setUIState()` so state persists across navigation.

## Related Files

| File | Purpose |
|------|---------|
| `admin/index.html` | CSS design system, layout classes |
| `admin/src/pages/dashboard.js` | Reference implementation |
| `admin/src/pages/design-system.js` | Layout testing page |
