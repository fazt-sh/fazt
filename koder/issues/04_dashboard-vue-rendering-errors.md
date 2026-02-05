# Dashboard Vue Rendering Errors

**Created**: 2026-02-06
**Status**: BLOCKING
**Priority**: HIGH
**Affected**: Admin UI Dashboard

## Problem

The Dashboard page is throwing Vue rendering errors after attempting to simplify the layout by moving Apps/Aliases/Logs into the System panel as stat cards.

## Errors (Browser Console)

```
TypeError: Cannot read properties of null (reading 'insertBefore')
    at insert (index-*.js:3:46305)

TypeError: Cannot set properties of null (setting 'textContent')
    at setElementText (index-*.js:3:46761)

TypeError: Cannot read properties of null (reading 'emitsOptions')
    at Jh (index-*.js:3:19275)
```

These errors occur during Vue's mounting and rendering process, specifically when stores are being updated and Vue's reactivity system tries to update DOM nodes that are null/undefined.

## What We Were Trying to Do

User requested to:
1. Move three "new cards" (Apps, Aliases, Logs) **inside** the System panel
2. Fix the Logs card which was showing empty

## What Changed

### Original Dashboard Structure
- **System Panel**: 4 stat cards (Status, Uptime, Memory, Storage)
- **Apps Panel**: Full collapsible panel with table showing top 5 apps (complex: loading states, empty states, navigation)
- **Aliases Panel**: Full collapsible panel with table showing top 5 aliases (complex: similar structure)

### Attempted New Structure
- **System Panel**: 7 stat cards (Status, Uptime, Memory, Storage, Apps, Aliases, Logs)
- Removed the full Apps and Aliases panels entirely

## Changes Made

**File**: `admin/src/pages/DashboardPage.js`

1. Added `useLogsStore` import
2. Added `logsStore.loadStats(client)` to `onMounted`
3. Removed `appsPanelCollapsed`, `aliasesPanelCollapsed` refs
4. Removed `toggleAppsPanel`, `toggleAliasesPanel` functions
5. Removed `topApps`, `topAliases` computed properties
6. Removed `navigateToApp` function
7. Added 3 new stat cards inside System panel:
   ```javascript
   <div class="stat-card card row-clickable" @click="navigateTo('/apps')">
     // Apps card
   </div>

   <div class="stat-card card row-clickable" @click="navigateTo('/aliases')">
     // Aliases card
   </div>

   <div class="stat-card card row-clickable" @click="navigateTo('/logs')">
     // Logs card - uses logsStore.stats.total_count || 0
   </div>
   ```
8. Removed entire Apps panel template section
9. Removed entire Aliases panel template section

## What We Tried

1. **First attempt**: Added inline `style="cursor:pointer"` - errors occurred
2. **Second attempt**: Replaced with `clickable` CSS class - errors persisted
3. **Third attempt**: Used `row-clickable` class (existing design system class) - errors still occur
4. **Reverted and redid carefully**: Same errors

## Current State

**Git Status**: `M admin/src/pages/DashboardPage.js` (uncommitted changes)

**Current Code**: Modified version with simplified structure
**Deployed**: Yes, to `@local` - errors in browser console
**Build**: Succeeds without errors - issues only appear at runtime in browser

## Errors Analysis

The stack trace shows:
- Errors happen during app initialization when stores load data
- `set value` triggers in stores → Vue reactivity triggers → Vue tries to update DOM → nodes are null
- Pattern: `store value set` → `trigger` → `insertBefore/setElementText` → null reference

This suggests:
- Template syntax issue (Vue can't compile/mount properly)
- DOM mount point not ready
- Component lifecycle issue (rendering before mount complete)
- Reactivity timing issue (stores updating before DOM ready)

## Attempts That Failed

- Changing inline styles to CSS classes
- Adding fallback values (`|| 0`)
- Careful removal of old code
- Revert and redo from scratch

## Debugging Questions

1. Why does removing the Apps/Aliases panels cause Vue mounting errors?
2. Is there something special about the panel-group structure that Vue depends on?
3. Could the grid CSS class (`grid-4` with 7 items) cause issues?
4. Is `row-clickable` class compatible with `stat-card card`?
5. Are there template syntax errors not caught by Vite build?

## Next Steps for Investigation

1. **Check if issue is specific to the cards or broader**
   - Try reverting just the panel removal, keep stat cards separate
   - Try adding just one new card at a time

2. **Verify template syntax**
   - Look for unclosed tags
   - Check for Vue directive issues
   - Validate HTML structure in browser DevTools

3. **Check store initialization**
   - Ensure logsStore.stats is properly initialized before rendering
   - Add v-if guards on stat cards until data loads

4. **Test with simpler changes**
   - Just add Logs card to System panel, don't remove other panels
   - Keep Apps/Aliases panels, add duplicate stat cards to test

5. **Review design system**
   - Is `row-clickable` meant for stat cards?
   - Should stat cards be in `panel-grid grid-4`?

## Files to Review

- `admin/src/pages/DashboardPage.js` - The problematic component
- `admin/src/stores/logs.js` - Logs store structure
- `admin/src/stores/apps.js` - Apps store structure
- `admin/src/stores/aliases.js` - Aliases store structure
- Design system CSS - Panel grid, stat card, clickable classes

## Workaround Options

If fix proves complex:
1. Keep original 3-panel structure, just add Logs stats to existing panels
2. Add Logs panel as 4th panel instead of integrating
3. Use different layout approach (not stat cards)

## Resolution (2026-02-06)

**Status**: CLOSED

### Root Cause (actual)

**`lucide.createIcons()` — external DOM mutation breaking Vue's VDOM.**

The original `refreshIcons()` called `window.lucide.createIcons()` which **replaces** `<i data-lucide>` elements with `<svg>` elements in the DOM. Vue's virtual DOM keeps references to those `<i>` elements. When a Pinia store update triggered a reactive re-render, Vue's patcher tried to operate on the original `<i>` nodes which no longer existed → `insertBefore` null, `setElementText` null.

This was NOT caused by:
- Template compilation issues (red herring)
- Fragment patching (we fixed this too but it wasn't the crash cause)
- Store timing (the stores were fine)
- The BFBB pattern itself (pattern is sound)

The error sequence:
```
1. Vue renders <i data-lucide="heart-pulse">
2. onUpdated → refreshIcons() → lucide.createIcons()
3. createIcons() REPLACES <i> with <svg> in DOM
4. Store update triggers re-render
5. Vue's patcher references the old <i> → gone → crash
```

### Fixes Applied (3 layers)

**Layer 1 — Icon rendering** (the actual fix):
Rewrote `admin/src/lib/icons.js` to inject SVGs **inside** `<i>` elements instead of replacing them. Vue still owns the `<i>`, its references stay valid, patches work.

**Layer 2 — Router timing**:
Changed `main.js` to `router.isReady().then(() => app.mount('#app'))` to prevent router's `install()` from triggering navigation before DOM exists.

**Layer 3 — Component granularity** (defense in depth):
Extracted App.js from 775-line monolith into 7 granular components (Sidebar, HeaderBar, CommandPalette, SettingsPanel, NewAppModal, CreateAliasModal, EditAliasModal). Single root element per component.

### Lesson Learned

Never use libraries that **replace** DOM elements Vue owns. This includes `lucide.createIcons()`, `highlight.js` auto-mode, and any jQuery-style `.replaceWith()`. Always inject content **inside** Vue-owned elements instead.

Documented in: `knowledge-base/skills/app/references/frontend-patterns.md` → "External DOM Mutation (CRITICAL)"
