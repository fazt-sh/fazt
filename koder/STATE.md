# Fazt Implementation State

**Last Updated**: 2026-01-31
**Current Version**: v0.17.0

## Status

State: **CLEAN** - Admin UI mobile responsiveness complete

---

## Current Session (2026-01-31)

**Admin UI Mobile Responsiveness**

Implemented comprehensive mobile and tablet responsive design for the Admin UI.

### Mobile Navigation

**Hamburger Menu:**
- Added mobile menu button in header (hidden on desktop ≥1024px)
- Sidebar slides in from left with backdrop overlay
- Touch-friendly 44px minimum tap targets
- Auto-closes on: backdrop click, nav link click, resize to desktop

**Breakpoints:**
- Mobile: < 768px (phones)
- Tablet: 768px - 1023px
- Desktop: ≥ 1024px

### Responsive Tables

**Horizontal Scroll:**
- All tables wrapped in `.table-container` for mobile scrolling
- Minimum 600px width, smooth touch scrolling
- Updated pages: apps list, app detail, aliases, dashboard

### Responsive Grids

**Dashboard:**
- Stats cards: 2 cols (mobile) → 3 cols (tablet) → 5 cols (desktop)
- Main grid: 1 col (mobile/tablet) → 3 cols (desktop)

**Apps Page:**
- Cards view: 1 col (mobile) → 2 cols (tablet) → 3 cols (desktop)

### Mobile UI Adjustments

**Modals & Overlays:**
- Full-screen modals on mobile (< 768px)
- Command palette: full-screen with better scrolling
- Settings panel: full-width at bottom
- Dropdowns: repositioned to bottom of screen

**Layout:**
- Header: reduced padding, compact breadcrumb
- Content: 16px padding on mobile
- Footer: wraps on small screens, auto height
- User button: avatar-only on mobile

**Touch Targets:**
- All interactive elements: 44px minimum
- Larger icon action areas
- Better spacing for fingers

### Utility Classes

- `.hide-mobile`: Hide on mobile (< 768px)
- `.show-mobile`: Show only on mobile

---

## Previous Session (2026-01-31 - Earlier)

**Admin UI Polish + Workflows Documentation**

Completed major UI refinements and created comprehensive workflows documentation.

### 1. Apps Page Refinements

**Filter Fix:**
- Fixed input focus loss on every keystroke
- Now only updates content area, keeps input focused
- Smooth filtering experience

**Layout Improvements:**
- New App button proper spacing (`padding: 6px 12px`)
- Fixed table layout jank with `table-layout: fixed`
- Column widths: Name 18%, Aliases 26%, ID 16%, etc.
- Text truncation with ellipsis for long content

**View Modes:**
- Added list/cards toggle (persists in localStorage)
- Cards view: 3-column grid with alias tags
- List view: Table with all fields + fixed columns

**Aliases Display:**
- Shows up to 3 aliases + "+N" badge for more
- Tag-style design with borders
- Searches across app name, ID, and aliases
- Mock data includes multi-alias apps for testing

**Empty States:**
- Professional empty state with icon
- "No apps yet" vs "No results" messages
- Call-to-action button when truly empty

### 2. App Detail Page

**Overview:**
- Stats cards: Files, Size, Updated
- Details card: Source, Created, Version
- Aliases card: All aliases as clickable buttons
- Files card: Complete file listing table

**Clickable Aliases:**
- Each alias is a button (not static text)
- Click opens `http://alias.hostname:port` in new tab
- Flexbox layout wraps naturally
- External link icon on each button

**Actions:**
- Back button to apps list
- Refresh button
- Delete with confirmation
- Breadcrumb navigation

### 3. Breadcrumb Fixes

**Navigation:**
- Clicking breadcrumb parent now navigates properly
- Uses `router.push()` instead of hash links
- Works seamlessly with SPA routing

**Dynamic Titles:**
- App detail shows actual app name (not "App Detail")
- Updates when app data loads
- Subscribes to `currentApp` changes

### 4. Sidebar Refinements

**Highlighting Fix:**
- Only exact route match highlights (not parent)
- Removed parent highlight logic
- Clean, predictable selection state

**Group Rename:**
- "Apps" → "Resources"
- "All Apps" → "Apps"
- More intuitive hierarchy

### 5. New App Modal

**Professional "Coming Soon":**
- Rocket icon + clean design
- Explains feature in development
- Shows CLI alternative: `fazt app deploy ./my-app`
- Multiple close options (button, X, backdrop)
- No broken functionality or alerts

### 6. SPA Routing Fix (Critical!)

**Problem:** Direct URLs like `/apps` returned 404

**Root Cause:** `manifest.json` had `"spa": true` but deploy wasn't reading it

**Solution:** Modified `cmd/server/app.go` to auto-detect SPA from manifest:
```go
if !*spaFlag {
    manifestPath := filepath.Join(dir, "manifest.json")
    if manifestData, err := os.ReadFile(manifestPath); err == nil {
        var manifest struct { SPA bool `json:"spa"` }
        if json.Unmarshal(manifestData, &manifest) == nil && manifest.SPA {
            *spaFlag = true
        }
    }
}
```

**Result:** Deploy shows `SPA: enabled (clean URLs)`, all routes work

### 7. Workflows Documentation

Created `knowledge-base/workflows/` structure:

**Files Created:**
- `workflows/README.md` - Index + navigation
- `workflows/admin-ui/architecture.md` - State management deep dive
- `workflows/admin-ui/adding-features.md` - Backend-first workflow
- `workflows/admin-ui/checklist.md` - Pre-implementation validation
- `workflows/admin-ui/testing.md` - Mock vs real mode testing

**Frontmatter Format:**
```yaml
---
title: Document Title
description: Brief description
updated: 2026-01-31
category: workflows
tags: [relevant, tags]
---
```

**Key Principle:** Backend-first development
- Never build UI without API support
- Validate endpoint exists before implementing
- Push back if backend missing
- Show "Coming Soon" as professional alternative

**Updated CLAUDE.md:**
- Added workflows section
- Rule: Check `updated:` date (if >2 days old, verify accuracy)
- Links to specific workflow guides

---

## Next Up

### Immediate: Admin UI Features

With mobile responsiveness complete, focus on completing missing pages:

- **Aliases Page** (high priority - currently placeholder)
- **System Page** (health dashboard)
- **Settings Page** (config UI)
- **Real-time updates** (WebSocket/SSE)

---

## Technical Debt / LEGACY_CODE

```bash
grep -rn "LEGACY_CODE" internal/
```

- `internal/storage/bindings.go` - `fazt.storage.*` namespace
- `internal/appid/appid.go` - old `app_*` format
- `internal/auth/service.go` - `generateUUID()`

See `koder/LEGACY.md` for removal guide.

---

## Quick Reference

```bash
# Admin UI
cd admin && npm run build
fazt app deploy admin --to local --name admin-ui

# Test modes
open http://admin-ui.192.168.64.3.nip.io:8080        # Real
open http://admin-ui.192.168.64.3.nip.io:8080?mock=true  # Mock

# Binary rebuild (if internal/ changes)
go build -o ~/.local/bin/fazt ./cmd/server

# Local server
systemctl --user restart fazt-local
journalctl --user -u fazt-local -f

# Check workflows
cat knowledge-base/workflows/README.md
```

---

## Architecture Notes

**Admin UI Data Flow:**
```
Backend API → fazt-sdk → Data Stores → UI Components
```

**State Management:**
- Reactive stores (apps, aliases, currentApp, auth)
- Subscribe/update pattern
- No direct API calls in UI
- Mock adapter for development

**Key Files:**
- `admin/packages/fazt-sdk/index.js` - API client
- `admin/src/stores/data.js` - State management
- `admin/src/pages/*.js` - Page components
- `admin/packages/fazt-sdk/fixtures/*.json` - Mock data

**Rule:** Mock data must exactly match real API structure
