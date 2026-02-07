# Plan 45: Admin UI Component Rebuild

**Status**: Draft
**Priority**: High — foundational for all future admin work
**Scope**: Full rebuild of admin component layer, CSS cleanup, Tailwind CDN alignment

## Problem

The admin UI has structural debt that blocks reliable iteration:

1. **Three components exist but nobody uses them** (AppPanel, DataTable, PanelToolbar)
2. **~380 lines of duplicated patterns** across pages (panels, tables, toolbars, modals, empty states)
3. **250 lines of hand-rolled Tailwind utilities** in admin.css, while Tailwind CDN is already loaded
4. **207 inline `style=` attributes** scattered across templates
5. **Inconsistent page shells** (Dashboard/Apps/Aliases/Logs use `design-system-page` wrapper; System/Settings use raw flex)
6. **Panel collapse uses `max-height: 3000px` hack** instead of proper CSS
7. **`refreshIcons()` boilerplate** in all 15 files
8. **`palettes` array** copy-pasted between SettingsPanel and SettingsPage
9. **AppsPage bug**: calls `store.deleteApp()` which doesn't exist (method is `remove()`)
10. **Unused parallel CSS** in `packages/fazt-ui/` (1,005 lines never imported)

## Goals

- **Pixel-identical visual output** — same look, better architecture
- **Granular, composable components** — build pages from primitives
- **Tailwind CDN native** — delete hand-rolled utility CSS, use what's already loaded
- **Consistent page shell** — every page uses the same layout wrapper
- **Reliable panel system** — CSS flexbox height management, no hacks
- **BFBB compliant** — plain .js files, CDN Tailwind, no build required

## Non-Goals

- No visual redesign (colors, typography, spacing stay identical)
- No new features (scope is pure refactoring)
- No .vue SFC migration (stays .js with template strings for BFBB)
- No Tailwind installation (stays CDN for BFBB)

---

## Architecture

### Component Hierarchy

```
lib/
  useIcons.js          — composable: replaces 15x onMounted/onUpdated boilerplate
  usePanel.js          — composable: panel collapse state with localStorage persistence
  palettes.js          — shared palette data (eliminates SettingsPanel/SettingsPage duplication)
  format.js            — (keep as-is)
  icons.js             — (keep as-is)

components/
  FPanel.js            — foundational collapsible panel (header/body/footer, 3 height modes)
  FTable.js            — data table with columns config, empty state, row clicks
  FToolbar.js          — search input + action buttons slot
  FModal.js            — modal wrapper (backdrop, close, header/body/footer slots)
  FEmpty.js            — empty state (icon, title, message)
  StatCard.js          — stat display card
  FilterDropdown.js    — dropdown filter (used by LogsPage)
  Pagination.js        — page navigation controls
  Sidebar.js           — (keep, minor cleanup: useIcons composable)
  HeaderBar.js         — (keep, minor cleanup: useIcons composable)
  CommandPalette.js    — (keep, minor cleanup: useIcons composable)
  SettingsPanel.js     — (refactor: use shared palettes.js)

pages/
  DashboardPage.js     — rebuilt with FPanel + StatCard
  AppsPage.js          — rebuilt with FPanel + FTable + FToolbar (fix deleteApp bug)
  AliasesPage.js       — rebuilt with FPanel + FTable + FToolbar
  LogsPage.js          — rebuilt with FPanel + FTable + FToolbar + FilterDropdown + Pagination
  SystemPage.js        — rebuilt with consistent page shell + responsive grid
  SettingsPage.js      — rebuilt with consistent page shell + shared palettes.js

DELETE:
  components/AppPanel.js      — replaced by FPanel
  components/DataTable.js     — replaced by FTable
  components/PanelToolbar.js  — replaced by FToolbar
  packages/fazt-ui/           — unused parallel CSS system
```

### CSS Strategy

**Keep** (design tokens + component classes, ~450 lines):
- Lines 1-157: Fonts, CSS variables, palettes (light/dark × 5 colors), typography, text colors
- Lines 158-308: Component styles (nav-item, card, btn, icon-box, badge, status-dot, input, toolbar, pill, swatch, menu, accordion, scrollbar, sidebar, modal, dropdown, avatar, settings-panel, table)
- Lines 309-422: Mobile responsiveness (sidebar overlay, hamburger, breakpoints, touch targets, edge-to-edge cards)
- Lines 424-648: Panel layout system (design-system-page, content-container, content-scroll, panel-group, panel-grid, stat-card, details-list)

**Delete** (~250 lines):
- Lines 649-897: Hand-rolled Tailwind utilities (.flex, .gap-2, .p-4, etc.) — Tailwind CDN already provides these

**Refactor** (panel-group styles):
- Replace `max-height: 3000px` / `max-height: 0` collapse hack with `display: none` on `.panel-group-body`
- The FPanel component will manage this via v-show or conditional rendering

**Delete entirely**:
- `packages/fazt-ui/` directory — unused, duplicates admin.css

### Tailwind CDN Configuration

The index.html already loads `<script src="https://cdn.tailwindcss.com"></script>`.

Add Tailwind config to index.html to extend with our design tokens:

```html
<script>
  tailwind.config = {
    theme: {
      extend: {
        fontFamily: {
          sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', 'sans-serif'],
          mono: ['JetBrains Mono', 'SF Mono', 'Consolas', 'monospace'],
        },
      }
    }
  }
</script>
```

This lets us use `font-sans` and `font-mono` from Tailwind while keeping our CSS variable system for colors (which Tailwind CDN can't dynamically consume for palette switching).

---

## Component Specifications

### FPanel — Foundational Panel

The most critical component. Every data page uses it.

**Props:**
```javascript
{
  title: String,           // Panel heading text
  count: [Number, String], // Badge count shown next to title
  icon: String,            // Lucide icon name for chevron (default: 'chevron-right')
  collapsed: Boolean,      // v-model:collapsed — external collapse state
  mode: String,            // 'fill' (default) | 'content' | 'fixed'
  height: String,          // Only used when mode='fixed', e.g. '400px'
}
```

**Slots:**
- `toolbar` — rendered between header and body (search bar, filter buttons)
- `default` — body content
- `footer` — rendered at bottom, always visible when expanded

**Emits:**
- `update:collapsed` — for v-model:collapsed

**Height modes (CSS-only, no JS):**

The page layout is a flex column. Each FPanel is a flex child.

```
mode="fill"     → flex: 1 1 0; min-height: 0;  (takes remaining space)
mode="content"  → flex: 0 0 auto;               (fits content, scrolls internally if needed)
mode="fixed"    → flex: 0 0 <height>;           (exact height)
```

When collapsed, ALL modes become `flex: 0 0 auto` (just header height).

**Collapse behavior:**
- Collapsed: body + toolbar + footer hidden via `display: none` (not max-height hack)
- Chevron rotates 90deg when expanded
- Smooth? No — `display: none` is instant, which is actually better for reliability. The old `max-height: 3000px` created scroll/layout issues.

**Template structure:**
```html
<div class="panel-group" :class="panelClasses">
  <div class="panel-group-card card" :style="panelStyle">
    <header class="panel-group-header">
      <button class="collapse-toggle" @click="toggle">
        <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
        <span class="text-heading text-primary">{{ title }}</span>
        <span v-if="count != null" class="text-caption mono px-1.5 py-0.5 ml-2 badge-muted" style="border-radius: var(--radius-sm)">{{ count }}</span>
        <slot name="header-actions"></slot>
      </button>
    </header>
    <template v-if="!isCollapsed">
      <slot name="toolbar"></slot>
      <div class="panel-group-body" :style="bodyStyle">
        <slot></slot>
      </div>
      <slot name="footer"></slot>
    </template>
  </div>
</div>
```

### FTable — Data Table

**Props:**
```javascript
{
  columns: Array,          // [{ key, label, class?, hideOnMobile? }]
  rows: Array,             // Data rows
  rowKey: String,          // Key field for :key binding (default: 'id')
  clickable: Boolean,      // Rows are clickable (default: true)
  emptyIcon: String,       // Lucide icon for empty state
  emptyTitle: String,      // Empty state heading
  emptyMessage: String,    // Empty state message
  loading: Boolean,        // Show loading state
}
```

**Slots:**
- `row` — scoped slot: `{ row, index }` — custom row rendering
- `cell-<key>` — scoped slot per column: `{ row, value }` — custom cell rendering
- `empty` — override empty state entirely

**Emits:**
- `row-click` — `(row)` when a row is clicked

**Structure:**
- Sticky thead with `background: var(--bg-1)`
- `hide-mobile` class on columns where `hideOnMobile: true`
- `scroll-panel` wrapper for scrollable body
- Built-in empty state (icon-box + title + message)
- Built-in loading state

### FToolbar — Toolbar

**Props:**
```javascript
{
  searchPlaceholder: String,  // Default: 'Filter...'
  modelValue: String,         // v-model for search
}
```

**Slots:**
- `actions` — right side buttons (new, refresh, etc.)
- `filters` — between search and actions (for filter dropdowns)

**Emits:**
- `update:modelValue` — search input

### FModal — Modal Wrapper

**Props:**
```javascript
{
  open: Boolean,      // v-model:open
  title: String,      // Modal heading
  subtitle: String,   // Optional subtitle
  icon: String,       // Lucide icon name
  maxWidth: String,   // Default: 'max-w-md'
}
```

**Slots:**
- `default` — modal body content
- `footer` — action buttons (cancel + submit)

**Emits:**
- `update:open` — when closed (backdrop click or X button)
- `close` — when closed

### StatCard

**Props:**
```javascript
{
  label: String,      // Top-left label (text-micro)
  icon: String,       // Top-right Lucide icon
  value: String,      // Large display value
  subtitle: String,   // Bottom text
  valueClass: String, // Optional class for value (e.g. 'text-success')
  clickable: Boolean, // Adds hover effect + cursor
}
```

**Emits:**
- `click` — when clicked (if clickable)

### FEmpty — Empty State

**Props:**
```javascript
{
  icon: String,       // Lucide icon name
  title: String,      // Heading
  message: String,    // Subtitle text
}
```

### FilterDropdown

**Props:**
```javascript
{
  label: String,         // Button text
  options: Array,        // [{ value, label }]
  modelValue: String,    // v-model — selected value
}
```

**Emits:**
- `update:modelValue` — when option selected

**Behavior:**
- Click button → show dropdown below it
- Click option → emit update, close dropdown
- Click outside → close dropdown
- Uses `position: fixed` + `getBoundingClientRect` (current approach, works on desktop)
- Has `hide-mobile` class — hidden on phones (same as current behavior)

### Pagination

**Props:**
```javascript
{
  currentPage: Number,
  totalPages: Number,
  showing: String,      // e.g. "1-25"
  total: Number,        // Total items
}
```

**Emits:**
- `page-change` — `(pageNumber)`

---

## Composables

### useIcons()

```javascript
// lib/useIcons.js
import { onMounted, onUpdated } from 'vue'
import { refreshIcons } from './icons.js'

export function useIcons() {
  onMounted(() => refreshIcons())
  onUpdated(() => refreshIcons())
}
```

Replaces 30 lines of boilerplate across 15 files.

### usePanel(key, defaultCollapsed = false)

```javascript
// lib/usePanel.js
import { ref } from 'vue'
import { useUIStore } from '../stores/ui.js'

export function usePanel(key, defaultCollapsed = false) {
  const ui = useUIStore()
  const collapsed = ref(ui.getUIState(key, defaultCollapsed))

  function toggle() {
    collapsed.value = !collapsed.value
    ui.setUIState(key, collapsed.value)
  }

  return { collapsed, toggle }
}
```

Replaces ~8 lines of panel state boilerplate per page.

### palettes.js

```javascript
// lib/palettes.js
export const palettes = [
  { id: 'stone', name: 'Stone', colors: ['#faf9f7', '#d97706'] },
  { id: 'slate', name: 'Slate', colors: ['#f8fafc', '#0284c7'] },
  { id: 'oxide', name: 'Oxide', colors: ['#faf8f8', '#e11d48'] },
  { id: 'forest', name: 'Forest', colors: ['#f7faf8', '#059669'] },
  { id: 'violet', name: 'Violet', colors: ['#faf9fc', '#7c3aed'] },
]
```

Eliminates copy-paste between SettingsPanel.js and SettingsPage.js.

---

## Page Rebuilds

### Consistent Page Shell

ALL pages must use this wrapper:

```html
<div class="design-system-page">
  <div class="content-container">
    <div class="content-scroll">
      <!-- page content -->
    </div>
  </div>
</div>
```

Currently System and Settings use raw flex containers. They must adopt this shell.

### DashboardPage (currently 199 lines → ~80 lines)

```html
<FPanel title="Dashboard" count="7 metrics" mode="content"
        v-model:collapsed="panel.collapsed">
  <div class="panel-grid grid-4">
    <StatCard v-for="stat in stats" :key="stat.label" v-bind="stat"
              :clickable="!!stat.route" @click="stat.route && navigateTo(stat.route)" />
  </div>
</FPanel>
```

Stats computed as array instead of 7 separate computeds.

### AppsPage — List Mode (currently 188 lines of list template → ~40 lines)

```html
<FPanel title="Apps" :count="store.items.length" mode="fill"
        v-model:collapsed="panel.collapsed">
  <template #toolbar>
    <FToolbar v-model="searchQuery" @update:modelValue="onSearch">
      <template #actions>
        <button class="btn btn-sm btn-primary toolbar-btn" @click="openNewAppModal">
          <i data-lucide="plus" class="w-4 h-4"></i>
        </button>
      </template>
    </FToolbar>
  </template>

  <FTable :columns="columns" :rows="filteredApps"
          empty-icon="layers" empty-title="No apps yet" empty-message="Deploy your first app via CLI"
          @row-click="navigateToApp($event.id)">
    <template #cell-title="{ row }">
      <div class="flex items-center gap-2" style="min-width: 0">
        <span class="status-dot status-dot-success pulse show-mobile" style="flex-shrink: 0"></span>
        <div class="icon-box icon-box-sm" style="flex-shrink: 0">
          <i data-lucide="box" class="w-3.5 h-3.5"></i>
        </div>
        <div style="min-width: 0; overflow: hidden">
          <div class="text-label text-primary truncate">{{ row.title || row.name }}</div>
          <div class="text-caption mono text-faint show-mobile">{{ formatBytes(row.size_bytes) }}</div>
        </div>
      </div>
    </template>
  </FTable>

  <template #footer>
    <div class="card-footer flex items-center justify-between" style="border-radius: 0">
      <span class="text-caption text-muted">{{ store.items.length }} app{{ store.items.length === 1 ? '' : 's' }}</span>
    </div>
  </template>
</FPanel>
```

**Bug fix**: Change `store.deleteApp(client, appId)` → `store.remove(client, appId)`.

### AliasesPage (currently 198 lines → ~50 lines)

Nearly identical structure to AppsPage list mode. Uses FPanel + FTable + FToolbar.

### LogsPage (currently 377 lines → ~120 lines)

Biggest reduction. The 4 duplicate FilterDropdown structures (~120 lines) become:

```html
<template #filters>
  <FilterDropdown label="Priority" :options="weightOptions" v-model="store.filterWeight"
                  @update:modelValue="v => selectFilter('weight', v)" />
  <FilterDropdown label="Action" :options="actionOptions" v-model="store.filterAction"
                  @update:modelValue="v => selectFilter('action', v)" />
  <FilterDropdown label="Actor" :options="actorOptions" v-model="store.filterActor"
                  @update:modelValue="v => selectFilter('actor', v)" />
  <FilterDropdown label="Type" :options="typeOptions" v-model="store.filterType"
                  @update:modelValue="v => selectFilter('type', v)" />
</template>
```

Footer uses `<Pagination>` component.

### SystemPage (currently 244 lines → ~120 lines)

Adopt `design-system-page` shell. Add responsive grid handling:

```html
<div class="design-system-page">
  <div class="content-container">
    <div class="content-scroll">
      <div class="panel-grid grid-2">
        <!-- 4 system cards (Health, Memory, Database, Runtime) -->
      </div>
    </div>
  </div>
</div>
```

Add CSS for `grid-2`:
```css
.panel-grid.grid-2 { grid-template-columns: repeat(2, 1fr); }
@media (max-width: 767px) { .panel-grid.grid-2 { grid-template-columns: 1fr; } }
```

### SettingsPage (currently 147 lines → ~130 lines)

Adopt `design-system-page` shell. Import `palettes` from shared `lib/palettes.js`.

### Modals: NewAppModal, CreateAliasModal, EditAliasModal

Wrap with `<FModal>`:

```html
<FModal v-model:open="ui.newAppModalOpen" title="Create New App"
        subtitle="Create an app from a template" icon="rocket">
  <!-- form content only — no wrapper boilerplate -->
  <template #footer>
    <button class="btn btn-secondary flex-1 text-label" @click="cancel">Cancel</button>
    <button class="btn btn-primary flex-1 text-label" @click="create">Create App</button>
  </template>
</FModal>
```

Each modal drops ~15 lines of wrapper boilerplate.

---

## Implementation Order

### Phase 1: Foundation (no visual changes)

1. Create `lib/useIcons.js`
2. Create `lib/usePanel.js`
3. Create `lib/palettes.js`
4. Create all component files: FPanel, FTable, FToolbar, FModal, FEmpty, StatCard, FilterDropdown, Pagination

### Phase 2: CSS Cleanup

5. Delete lines 649-897 from admin.css (hand-rolled Tailwind utilities)
6. Add Tailwind CDN config to index.html (font-family extension)
7. Add `.panel-grid.grid-2` responsive CSS
8. Refactor panel-group collapse CSS: replace max-height hack
9. Verify build passes, visual output unchanged

### Phase 3: Page Rebuilds (one at a time)

Each page: rebuild → deploy to local → visually verify → next page.

10. **DashboardPage** — simplest, good first test of FPanel + StatCard
11. **AppsPage** — FPanel + FTable + FToolbar + fix deleteApp bug
12. **AliasesPage** — nearly identical to AppsPage structure
13. **LogsPage** — most complex: FPanel + FTable + FToolbar + FilterDropdown + Pagination
14. **SystemPage** — adopt design-system-page shell, add grid-2 responsive
15. **SettingsPage** — adopt design-system-page shell, shared palettes

### Phase 4: Modals + Cleanup

16. **FModal** integration into NewAppModal, CreateAliasModal, EditAliasModal
17. Adopt `useIcons()` across all remaining components (Sidebar, HeaderBar, CommandPalette, SettingsPanel)
18. Delete old unused components (AppPanel.js, DataTable.js, PanelToolbar.js)
19. Delete `packages/fazt-ui/` directory
20. Final build verification + deploy to local + visual QA

### Phase 5: Deploy + Verify

21. Deploy to local, test all pages and interactions
22. Deploy to zyt, verify production

---

## Inline Style Elimination Strategy

Current: `style="background:var(--bg-1);border-color:var(--border)"` scattered everywhere.

These inline styles exist because Tailwind CDN can't consume CSS variables for background/border colors. The solution is **keep component CSS classes** for theme-aware colors:

- `style="background:var(--bg-1)"` → keep in component CSS (`.card` already handles this)
- `style="color:var(--text-3)"` → use `.text-muted` class (already exists)
- `style="border-color:var(--border)"` → use `.border` class (already exists in admin.css)

For the remaining inline styles that genuinely need CSS variables, we accept them as necessary for the palette-switching system. The goal is to eliminate unnecessary ones, not all of them.

---

## Testing Strategy

After each phase:
1. `cd admin && npm run build` — verify build passes
2. `fazt @local app deploy ./admin` — deploy to local
3. Visually verify each page at desktop + mobile widths
4. Test panel collapse/expand behavior
5. Test dark mode + palette switching
6. Test mock mode (`?mock=true`)

---

## Risk Mitigation

- **Visual regression**: Every page rebuild is immediately deployed and visually compared
- **BFBB compliance**: All components use .js with template strings, Tailwind CDN only
- **Incremental**: Each page is independently migratable. If one breaks, others unaffected
- **Rollback**: Git history preserves every state. Each phase is a commit.

---

## Expected Outcomes

| Metric | Before | After |
|--------|--------|-------|
| Total JS lines (pages + components) | ~2,820 | ~1,800 |
| Duplicated patterns | ~380 lines | ~0 |
| Inline style attributes | 207 | ~30 (necessary CSS var ones) |
| CSS file lines | 897 + 1,005 (unused) | ~500 |
| Components used by pages | 0 of 3 | 8+ of 8 |
| Panel collapse mechanism | max-height: 3000px hack | display: none (reliable) |
| Page shell consistency | 4 of 7 pages | 7 of 7 pages |
| refreshIcons boilerplate | 30 lines across 15 files | 15 lines (one per file, 1 line each) |
