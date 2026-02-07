# Fazt Implementation State

**Last Updated**: 2026-02-07
**Current Version**: v0.27.0

## Status

State: CLEAN
Plan 45 (Admin UI Component Rebuild) implemented.

---

## Last Session (2026-02-07) — Plan 45: Admin UI Component Rebuild

### What Changed

Full rebuild of admin component layer — same visual output, better architecture.

#### New Composables (`lib/`)
- **`useIcons.js`** — replaces `onMounted/onUpdated + refreshIcons()` boilerplate in 15 files
- **`usePanel.js`** — replaces panel collapse state + localStorage persistence boilerplate
- **`palettes.js`** — shared palette data (was copy-pasted between SettingsPanel + SettingsPage)

#### New Components (`components/`)
- **`FPanel.js`** — foundational collapsible panel (header/body/footer, 3 height modes: fill/content/fixed)
- **`FTable.js`** — data table with columns config, empty state, loading state, row clicks, scoped slots
- **`FToolbar.js`** — search input + filter/action slots
- **`FModal.js`** — modal wrapper (backdrop, close, header/body/footer slots)
- **`FEmpty.js`** — empty state (icon, title, message)
- **`StatCard.js`** — stat display card with click support
- **`FilterDropdown.js`** — dropdown filter with position:fixed placement
- **`FPagination.js`** — page navigation controls

#### Pages Rebuilt
- **DashboardPage** — 199→93 lines, uses FPanel + StatCard, stats as computed array
- **AppsPage** — 301→234 lines, uses FPanel + FTable + FToolbar, **fixed `deleteApp` bug** (`store.deleteApp()` → `store.remove()`)
- **AliasesPage** — 198→152 lines, uses FPanel + FTable + FToolbar
- **LogsPage** — 377→218 lines (biggest reduction), uses FPanel + FTable + FToolbar + FilterDropdown + FPagination
- **SystemPage** — 244→190 lines, adopted `design-system-page` shell + `grid-2` responsive
- **SettingsPage** — 147→123 lines, adopted `design-system-page` shell + shared `palettes.js`

#### Modals Refactored
- **NewAppModal** — uses FModal wrapper, dropped ~15 lines boilerplate
- **CreateAliasModal** — uses FModal wrapper
- **EditAliasModal** — uses FModal wrapper

#### Existing Components Updated
- **Sidebar, HeaderBar, CommandPalette, SettingsPanel** — all use `useIcons()` composable

#### CSS Cleanup
- **Deleted ~250 lines** of hand-rolled Tailwind utilities (lines 649-897) — Tailwind CDN already provides these
- **Added Tailwind CDN config** to index.html (font-family extension)
- **Added `.panel-grid.grid-2`** responsive CSS for SystemPage
- **Fixed panel collapse** — replaced `max-height: 3000px` hack with `display: none` (FPanel uses `v-if`)
- **admin.css**: 897→664 lines

#### Deleted
- **`AppPanel.js`** — replaced by FPanel
- **`DataTable.js`** — replaced by FTable
- **`PanelToolbar.js`** — replaced by FToolbar

### Metrics

| Metric | Before | After |
|--------|--------|-------|
| Page JS lines (6 pages) | ~1,366 | ~1,010 |
| New component lines | 0 | 353 |
| CSS lines | 897 | 664 |
| refreshIcons boilerplate | ~30 lines (15 files) | ~15 lines (useIcons in each) |
| Panel collapse | max-height: 3000px hack | v-if (reliable) |
| Page shell consistency | 4/7 pages | 7/7 pages |
| deleteApp bug | broken | fixed |
| Bundle size | 325 KB | 313 KB |

### Key Files
```bash
admin/src/lib/useIcons.js        # Icon composable
admin/src/lib/usePanel.js        # Panel state composable
admin/src/lib/palettes.js        # Shared palette data
admin/src/components/FPanel.js   # Foundational panel
admin/src/components/FTable.js   # Data table
admin/src/components/FToolbar.js # Toolbar
admin/src/components/FModal.js   # Modal wrapper
admin/src/components/StatCard.js # Stat card
admin/src/components/FilterDropdown.js  # Filter dropdown
admin/src/components/FPagination.js     # Pagination
```

---

## Next Session

### Priority
- **Test admin against real server** — Auth, apps, aliases all work after rebuild
- **Test admin with `?mock=true`** — All pages load, data renders
- **Migrate Preview app** — Use `createAppClient()` instead of hand-rolled `api.js`
- **Document media APIs in KB** — `fazt.app.media.{probe,transcode,serve,resize}`

### Direction
- **Plan 44: Drop app** — File/folder hosting via fazt (idea stage)
- **SDK external consumption** — Relative imports for now; Drop will host the bundle later

### Known Issues
- **`fazt @local app list`** — Returns empty error (pre-existing bug)

---

## Quick Reference

```bash
# Test admin build
cd admin && npm run build

# Test all Go
go test ./... -short -count=1

# Deploy admin
fazt @local app deploy ./admin

# Deploy preview
fazt @local app deploy ./servers/local/preview
```
