# Layout Patterns

## Fixed-Height Layout (Prevents Scroll Jumps)

**Problem**: When content grows, scrollbar appears/disappears causing layout shift.

**Solution**: Full-height container with inner scrolling.

### HTML Structure
```html
<style>
  /* Full height, no scrollbars on outer container */
  html, body {
    height: 100%;
    overflow: hidden;
    margin: 0;
    padding: 0;
  }

  /* Full height app container */
  #app { height: 100%; }

  /* Prevent scrollbar layout shift */
  .overflow-y-auto {
    scrollbar-gutter: stable;
  }
</style>

<body class="font-sans antialiased bg-neutral-50 dark:bg-neutral-950">
  <div id="app"></div>
</body>
```

### Vue Template
```vue
<div class="h-screen flex flex-col overflow-hidden">
  <!-- Fixed Header -->
  <div class="flex-none bg-white border-b px-4 py-4">
    <h1>App Title</h1>
  </div>

  <!-- Scrollable Content -->
  <div class="flex-1 overflow-y-auto">
    <div class="max-w-4xl mx-auto px-4 py-6 pb-24">
      <!-- Your content here -->
    </div>
  </div>

  <!-- Fixed FAB (outside scroll container) -->
  <button class="fixed bottom-6 right-6 ...">+</button>
</div>
```

### Key Points
- App container: `h-screen flex flex-col overflow-hidden`
- Header: `flex-none` (fixed at top, doesn't scroll)
- Content: `flex-1 overflow-y-auto` (grows to fill space, scrollable)
- Bottom padding: `pb-24` on inner content (prevents FAB overlap)
- FAB: `fixed` positioning (outside scroll container)
- Scrollbar-gutter: `stable` (reserves space even when scrollbar hidden)

## Responsive Layout

### Mobile-First
```vue
<div class="max-w-4xl mx-auto px-4">
  <!-- Single column on mobile -->
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
    <div>Card 1</div>
    <div>Card 2</div>
    <div>Card 3</div>
  </div>
</div>
```

### Safe Areas (iOS)
```css
body {
  padding: env(safe-area-inset-top)
           env(safe-area-inset-right)
           env(safe-area-inset-bottom)
           env(safe-area-inset-left);
}
```

## Multi-View Layout

### Tab-Based Views
```vue
<div class="h-screen flex flex-col overflow-hidden">
  <!-- Header -->
  <div class="flex-none ...">...</div>

  <!-- Fixed Tab Bar -->
  <div class="flex-none bg-white border-b px-4">
    <div class="flex gap-2 p-1">
      <button
        @click="view = 'list'"
        :class="view === 'list' ? 'bg-blue-500 text-white' : 'text-neutral-600'"
        class="flex-1 py-2 rounded-lg"
      >
        List
      </button>
      <button
        @click="view = 'grid'"
        :class="view === 'grid' ? 'bg-blue-500 text-white' : 'text-neutral-600'"
        class="flex-1 py-2 rounded-lg"
      >
        Grid
      </button>
    </div>
  </div>

  <!-- Scrollable Content (per view) -->
  <div class="flex-1 overflow-y-auto">
    <div v-if="view === 'list'" class="max-w-4xl mx-auto px-4 py-6">
      <!-- List view content -->
    </div>
    <div v-if="view === 'grid'" class="max-w-4xl mx-auto px-4 py-6">
      <!-- Grid view content -->
    </div>
  </div>
</div>
```

## Custom Scrollbar

```css
/* Webkit browsers (Chrome, Safari, Edge) */
::-webkit-scrollbar {
  width: 6px;
}

::-webkit-scrollbar-thumb {
  background: rgba(128, 128, 128, 0.3);
  border-radius: 3px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

/* Firefox */
* {
  scrollbar-width: thin;
  scrollbar-color: rgba(128, 128, 128, 0.3) transparent;
}
```

## Summary

✓ Use `h-screen flex flex-col overflow-hidden` for app container
✓ Use `flex-none` for fixed headers/footers
✓ Use `flex-1 overflow-y-auto` for scrollable content
✓ Use `scrollbar-gutter: stable` to prevent layout shift
✓ Use `fixed` positioning for FABs
✓ Add bottom padding (`pb-24`) to content for FAB clearance
