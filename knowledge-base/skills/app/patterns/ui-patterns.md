# UI Patterns

Common UI patterns for polished fazt apps.

## Prevent Layout Shift

### Always-Visible Buttons

**Problem**: Buttons that appear/disappear (via `v-if`) cause layout shift.

**Solution**: Always render, but disable when inactive.

```vue
<!-- BAD - causes layout shift -->
<button v-if="hasFilters" @click="clear">Clear</button>

<!-- GOOD - always visible, disabled when inactive -->
<button
  @click="clear"
  :disabled="!hasFilters"
  class="transition-opacity"
  :class="hasFilters
    ? 'text-blue-500 hover:underline cursor-pointer'
    : 'text-gray-300 cursor-default'"
>
  Clear
</button>
```

### Badge Overlays vs Inline Counts

**Problem**: Inline count badges resize their parent when count changes.

**Solution**: Use absolute positioning for badge overlays.

```vue
<!-- BAD - inline badge shifts button width -->
<button class="flex items-center gap-2">
  <FilterIcon />
  Filters
  <span v-if="count" class="bg-blue-500 px-1 rounded">{{ count }}</span>
</button>

<!-- GOOD - overlay badge, no shift -->
<div class="relative">
  <button class="flex items-center gap-2">
    <FilterIcon />
    Filters
  </button>
  <span
    v-if="count"
    class="absolute -top-1.5 -right-1.5 min-w-[18px] h-[18px]
           flex items-center justify-center text-[10px] font-semibold
           bg-yellow-400 text-gray-900 rounded-full"
  >
    {{ count }}
  </span>
</div>
```

---

## Click-Outside-to-Close

Pattern for closing dropdowns/modals when clicking outside.

### Overlay Method (Recommended)

```vue
<template>
  <div class="relative">
    <!-- Trigger -->
    <button @click="isOpen = !isOpen">Menu</button>

    <!-- Dropdown -->
    <div
      v-if="isOpen"
      class="absolute right-0 mt-2 w-48 bg-white rounded-xl shadow-lg z-[100]"
    >
      <button @click="doSomething">Action</button>
    </div>

    <!-- Invisible overlay catches outside clicks -->
    <div
      v-if="isOpen"
      class="fixed inset-0 z-[99]"
      @click="isOpen = false"
    />
  </div>
</template>
```

**Key points**:
- Overlay: `fixed inset-0 z-[99]` (covers entire screen)
- Dropdown: `z-[100]` (above overlay)
- Overlay is invisible but captures clicks

### Z-Index Layering

Consistent z-index scale for stacking:

| Layer | Z-Index | Use |
|-------|---------|-----|
| Base content | 0 | Normal page content |
| Sticky headers | `z-40` | Fixed headers, filter panels |
| Header | `z-50` | Main app header |
| Dropdown overlay | `z-[99]` | Click-outside catcher |
| Dropdown content | `z-[100]` | Menus, popovers |
| Modal backdrop | `z-[200]` | Modal overlay |
| Modal content | `z-[201]` | Modal dialog |
| Toast/notifications | `z-[300]` | Top-level alerts |

**Important**: Parent elements with `backdrop-blur` or `transform` create new stacking contexts. Set z-index on the parent container, not just children.

```vue
<!-- Header needs z-index for dropdown to escape -->
<header class="relative z-50 backdrop-blur-xl">
  <DropdownMenu />  <!-- z-[100] works relative to page, not just header -->
</header>

<FilterPanel class="relative z-40" />  <!-- Lower than header -->
```

---

## Mobile Navbar Optimization

Space-efficient patterns for mobile headers.

### Hide Logo, Expand Search

```vue
<header class="h-12 flex items-center gap-3 px-4">
  <!-- Logo: hidden on mobile -->
  <button @click="goHome" class="hidden sm:flex items-center gap-2">
    <Logo class="w-6 h-6" />
    <span class="font-semibold">AppName</span>
  </button>

  <!-- Search: fills space on mobile, fixed width on desktop -->
  <div class="flex-1 sm:flex-initial sm:w-80">
    <input type="text" placeholder="Search..." class="w-full ..." />
  </div>

  <!-- Spacer: only on desktop (search doesn't fill) -->
  <div class="hidden sm:block flex-1" />

  <!-- Actions -->
  <button>...</button>
</header>
```

### Move Controls to Avatar Menu

On mobile, move secondary controls (theme toggle, view switcher) into the user menu:

```vue
<div class="relative">
  <button @click="showMenu = !showMenu">
    <Avatar />
  </button>

  <div v-if="showMenu" class="absolute right-0 mt-2 w-52 bg-white rounded-xl shadow-lg">
    <!-- Logo row (visible on mobile, acts as home link) -->
    <button @click="goHome" class="w-full px-4 py-3 flex items-center gap-3 border-b">
      <Logo class="w-6 h-6" />
      <span class="font-semibold">AppName</span>
    </button>

    <!-- User info -->
    <div class="px-4 py-3 border-b">
      <p class="font-medium">{{ user.name }}</p>
      <p class="text-xs text-gray-500">{{ user.email }}</p>
    </div>

    <!-- View toggle -->
    <div class="px-4 py-2.5 border-b flex items-center justify-between">
      <span class="text-xs text-gray-500">View</span>
      <ViewToggle v-model="viewMode" />
    </div>

    <!-- Theme toggle -->
    <div class="px-4 py-2.5 border-b flex items-center justify-between">
      <span class="text-xs text-gray-500">Theme</span>
      <ThemeToggle />
    </div>

    <!-- Sign out -->
    <button @click="signOut" class="w-full px-4 py-2.5 text-left">
      Sign out
    </button>
  </div>
</div>
```

---

## Responsive Filter Panel

Desktop: horizontal row with flex-wrap. Mobile: stacked grid.

```vue
<template>
  <div class="px-4 py-3">
    <!-- Desktop: horizontal wrap -->
    <div class="hidden sm:flex flex-wrap items-end gap-3">
      <FilterSelect v-for="filter in filters" :key="filter.id" ... />
      <ClearButton />
    </div>

    <!-- Mobile: stacked rows -->
    <div class="sm:hidden space-y-3">
      <!-- Full-width filter (e.g., range slider) -->
      <RangeSlider v-model="yearRange" />

      <!-- 2-column grid for related filters -->
      <div class="grid grid-cols-2 gap-3">
        <FilterSelect label="Fund" ... />
        <FilterSelect label="Country" ... />
      </div>

      <div class="grid grid-cols-2 gap-3">
        <FilterSelect label="Stage" ... />
        <FilterSelect label="Revenue" ... />
      </div>

      <!-- Last row: filter + clear button -->
      <div class="flex items-end gap-3">
        <div class="flex-1">
          <FilterSelect label="Industry" ... />
        </div>
        <ClearButton class="flex-shrink-0" />
      </div>
    </div>
  </div>
</template>
```

---

---

## Edge-to-Edge Mobile (Optional)

**When to use**: Apps that need maximum screen real estate on mobile (data-heavy UIs, dashboards, lists with dense information).

**When to skip**: Landing pages, marketing sites, simple tools where breathing room looks better.

**Ask the user**: "Should content extend edge-to-edge on mobile for maximum space, or keep padding for a more relaxed feel?"

### Implementation

```css
/* Responsive breakpoints */
@media (max-width: 1023px) {
  .main-content { padding: 16px; }
}

/* Edge-to-edge mobile - MUST come after tablet rule */
@media (max-width: 767px) {
  .main-content { padding: 8px 0; }

  /* Cards: flat edges, no side borders */
  .card {
    border-radius: 0;
    border-left: none;
    border-right: none;
  }
}
```

**Important**: Mobile rules must come AFTER tablet rules in CSS. Both `max-width: 767px` and `max-width: 1023px` match on mobile - later rules win.

### Responsive Breakpoints Reference

| Breakpoint | Width | Typical Behavior |
|------------|-------|------------------|
| Mobile | < 768px | Edge-to-edge (if enabled), stacked layouts, hamburger menu |
| Tablet | 768-1023px | 16px padding, 2-column grids |
| Desktop | ≥ 1024px | 20px+ padding, multi-column, sidebars |

---

## UI State Persistence (Optional)

**When to use**: Apps where users benefit from remembered preferences (collapsed sections, view modes, sort orders, filter states).

**When to skip**: Simple single-purpose tools, content sites where fresh state is preferred.

**Ask the user**: "Should the app remember user preferences like view mode or collapsed sections?"

### Implementation

```javascript
// Unified localStorage key for your app
const UI_STATE_KEY = 'myapp.ui.state'

function getUIState(key, defaultValue = null) {
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

### Usage Examples

```javascript
// View mode (list vs grid)
const viewMode = getUIState('view', 'list')
setUIState('view', 'grid')

// Sort preference
const sortBy = getUIState('sort', 'date')

// Collapsed sections (if using collapsible UI)
const isCollapsed = getUIState('section.filters.collapsed', false)
```

### Key Naming Convention

Use dot notation for namespacing:
```
myapp.ui.state = {
  "view": "grid",
  "sort": "name",
  "filters.collapsed": false,
  "sidebar.width": 280
}
```

---

## Summary

**Always apply:**
- Render UI elements always, disable instead of hide (prevents layout shift)
- Use absolute positioning for badges/counts
- Click-outside: invisible overlay at `z-[99]`, content at `z-[100]`
- Set z-index on parent containers with backdrop-blur
- Mobile: hide logo, expand search, move controls to menu

**Ask the user first:**
- Edge-to-edge mobile (space efficiency vs visual breathing room)
- UI state persistence (remember preferences vs fresh state)
