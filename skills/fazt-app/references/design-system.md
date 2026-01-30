# Design System

Visual guidelines for fazt apps. Apps should feel native, polished, and
delightful—like Apple apps.

## Colors

Use neutral palette with a single accent color.

### Light Mode

| Element | Class | Hex |
|---------|-------|-----|
| Background | `bg-neutral-50` | #fafafa |
| Surface/Card | `bg-white` | #ffffff |
| Text | `text-neutral-900` | #171717 |
| Muted text | `text-neutral-500` | #737373 |
| Border | `border-neutral-200` | #e5e5e5 |
| Accent | `bg-blue-500` | #3b82f6 |

### Dark Mode

| Element | Class | Hex |
|---------|-------|-----|
| Background | `bg-neutral-950` | #0a0a0a |
| Surface/Card | `bg-neutral-900` | #171717 |
| Text | `text-neutral-100` | #f5f5f5 |
| Muted text | `text-neutral-400` | #a3a3a3 |
| Border | `border-neutral-800` | #262626 |
| Accent | `bg-blue-500` | #3b82f6 |

### Semantic Colors

| State | Light | Dark |
|-------|-------|------|
| Success | `text-green-600` | `text-green-400` |
| Warning | `text-amber-600` | `text-amber-400` |
| Error | `text-red-600` | `text-red-400` |

## Typography

Use Inter font family. Fall back to system-ui.

```html
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
```

### Scale

| Style | Class | Use |
|-------|-------|-----|
| Display | `text-4xl font-bold` | Hero numbers, main metrics |
| Title | `text-2xl font-semibold` | Page titles |
| Headline | `text-xl font-semibold` | Section headers |
| Subhead | `text-lg font-medium` | Card titles |
| Body | `text-base` | Main content |
| Caption | `text-sm text-neutral-500` | Labels, hints, metadata |
| Small | `text-xs` | Timestamps, badges |

## Layout

### Full Height App

Prevent scroll jumps with fixed-height container:

```html
<div class="h-screen flex flex-col overflow-hidden">
  <!-- Header: fixed at top -->
  <header class="flex-none px-4 py-3 border-b">
    ...
  </header>

  <!-- Content: scrollable -->
  <main class="flex-1 overflow-y-auto">
    ...
  </main>

  <!-- Footer: fixed at bottom (optional) -->
  <footer class="flex-none px-4 py-3 border-t">
    ...
  </footer>
</div>
```

### Scrollbar Stability

Prevent layout shift when scrollbar appears:

```css
.overflow-y-auto {
  scrollbar-gutter: stable;
}
```

### Safe Areas (iOS)

```css
body {
  padding: env(safe-area-inset-top) env(safe-area-inset-right)
           env(safe-area-inset-bottom) env(safe-area-inset-left);
}
```

## Spacing

Use consistent spacing scale.

| Size | Value | Use |
|------|-------|-----|
| `gap-1` | 4px | Icon-text gap |
| `gap-2` | 8px | Tight grouping |
| `gap-3` | 12px | List items |
| `gap-4` | 16px | Card content |
| `p-4` | 16px | Card padding |
| `p-6` | 24px | Page padding |
| `space-y-6` | 24px | Section gaps |

## Corners

| Element | Class |
|---------|-------|
| Buttons, inputs | `rounded-lg` |
| Cards | `rounded-xl` |
| Modals | `rounded-2xl` |
| Pills, badges | `rounded-full` |

## Shadows

Use sparingly. Cards usually don't need shadows with proper borders.

| Level | Class | Use |
|-------|-------|-----|
| Subtle | `shadow-sm` | Dropdowns |
| Medium | `shadow-md` | Floating cards |
| Large | `shadow-xl` | Modals |

## Touch Targets

Minimum 44x44px for all interactive elements.

```html
<button class="min-h-[44px] min-w-[44px] px-4 py-2">
  Click me
</button>
```

### Touch Feedback

```css
.touch-active:active {
  opacity: 0.7;
  transform: scale(0.98);
}
```

## Buttons

### Primary

```html
<button class="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 active:bg-blue-700 transition-colors">
  Primary
</button>
```

### Secondary

```html
<button class="px-4 py-2 bg-neutral-100 dark:bg-neutral-800 rounded-lg hover:bg-neutral-200 dark:hover:bg-neutral-700 transition-colors">
  Secondary
</button>
```

### Ghost

```html
<button class="px-4 py-2 text-neutral-600 dark:text-neutral-400 hover:bg-neutral-100 dark:hover:bg-neutral-800 rounded-lg transition-colors">
  Ghost
</button>
```

### Danger

```html
<button class="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors">
  Delete
</button>
```

## Inputs

```html
<input
  type="text"
  class="w-full px-3 py-2 bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
  placeholder="Enter text..."
/>
```

## Cards

```html
<div class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-xl p-4">
  <h3 class="text-lg font-medium">Card Title</h3>
  <p class="mt-2 text-neutral-600 dark:text-neutral-400">Card content goes here.</p>
</div>
```

## Loading States

### Spinner

```html
<div class="animate-spin rounded-full h-6 w-6 border-2 border-blue-500 border-t-transparent"></div>
```

### Skeleton

```html
<div class="animate-pulse bg-neutral-200 dark:bg-neutral-700 rounded h-4 w-3/4"></div>
```

## Transitions

Use 150ms for micro-interactions, 300ms for larger changes.

```html
<div class="transition-all duration-150">
  ...
</div>
```

## PWA Meta Tags

```html
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="theme-color" content="#ffffff" media="(prefers-color-scheme: light)">
<meta name="theme-color" content="#0a0a0a" media="(prefers-color-scheme: dark)">
```
