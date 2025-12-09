# Next Session for Fazt Admin Project

## Current State
- **Project**: Fazt.sh Admin Dashboard (React + TypeScript + Vite)
- **Location**: `/home/testman/workspace/admin`
- **Branch**: `glm/ui-refinements`
- **Status**: Phase 2A - Route Hierarchy Implementation
- **Dev Server**: Running on http://localhost:37180
- **Active Plan**: `koder/plans/14_admin-spa-complete.md`

## Most Recent Work (December 9, 2025)

### Plan Consolidation
- Created **Plan 14**: Unified Admin SPA plan combining Plans 12 and 13
- Marked Plans 12 and 13 as superseded
- Plan 14 is now the **single source of truth** for all Admin SPA work

### Completed Features (Phase 1)
1. **Foundation & Shell** - 100% complete
   - Vite + React + TypeScript setup
   - Tailwind CSS with custom design system
   - App shell (Navbar, Sidebar with sub-menus)
   - Core infrastructure (API client, contexts, router)
   - Login, Dashboard, Sites, DesignSystem pages
   - PWA basics

2. **Design System**
   - "Technical Precision" aesthetic (Linear/Vercel-inspired)
   - Inter + Consolas fonts
   - CSS variables for theming
   - All UI primitives built

3. **Dashboard**
   - Stats cards
   - World map with Datamaps (theme-aware)
   - Visitor traffic visualization

4. **Sidebar**
   - Sub-menu system with expand/collapse
   - Proper state management
   - Mobile responsive

## Current Task: Phase 2A - Route Hierarchy

### What Needs to Be Done

Implement the UX-first route structure with placeholders for all pages.

**Files to Create (28 total)**:
- 1 Breadcrumbs component
- 1 PlaceholderPage component
- 5 Layout components (Sites, System, Apps, Security, External)
- 1 NotFound page
- 18 Placeholder pages (see Plan 14 section 6.5)

**Files to Modify (2 total)**:
- `App.tsx` - New routing structure
- `Sidebar.tsx` - New menu items

### Route Structure

```
/#/                          Dashboard
/#/sites/*                   Sites Management
/#/system/*                  System Administration
/#/apps/*                    Application Management
/#/security/*                Security Management
/#/external/*                External Integrations
/#/*                         404 Not Found
```

See Plan 14, Section 5.2 for complete route map.

## Bootstrap Chain

To continue work, run:
```bash
read and execute koder/start.md
```

Or jump directly to the plan:
```bash
read koder/plans/14_admin-spa-complete.md
```

## Context Payload (Read These)

1. **Active Plan**: `koder/plans/14_admin-spa-complete.md` (PRIMARY)
2. **Architecture**: `koder/analysis/04_comprehensive_technical_overview.md`
3. **API Reference**: `koder/plans/11_api-standardization.md`

### Superseded Plans (for reference only)
- `koder/plans/12_admin-spa-rebuild.md` - Original plan (Phase 1 complete)
- `koder/plans/13_route-hierarchy-implementation.md` - Route design (merged into 14)

## Development Commands

```bash
# Start dev server
cd /home/testman/workspace/admin
npm run dev -- --port 37180 --host 0.0.0.0

# Build for production
npm run build

# Copy to Go embed location
rm -rf internal/assets/system/admin
cp -r admin/dist internal/assets/system/admin

# Build Go binary
go build -o fazt ./cmd/server
```

## Implementation Order (Phase 2A)

Execute in this exact sequence:

1. Create directories: `pages/sites/`, `pages/system/`, `pages/apps/`, `pages/security/`, `pages/external/`
2. Create `components/ui/Breadcrumbs.tsx`
3. Create `components/PlaceholderPage.tsx`
4. Create 5 layout components
5. Create `pages/NotFound.tsx`
6. Create all 18 placeholder pages
7. Update `App.tsx`
8. Update `Sidebar.tsx` menuItems
9. Test all routes
10. Verify sidebar highlighting

## Verification Checklist

After implementation, test each route works:
- [ ] All 25+ routes accessible
- [ ] Breadcrumbs show on nested pages
- [ ] Sidebar highlights correctly
- [ ] 404 page shows for invalid routes
- [ ] No console errors
- [ ] Build succeeds

## Notes

- Server is running on port 37180
- Hash-based routing (#/) is working
- No backwards compatibility needed (no users yet)
- Design System stays in production at `/system/design-system`

---

## Your Mission (This Session)

### Implement Phase 2A: Route Hierarchy

**Goal**: Create all routes with pretty placeholders so we can visualize the final admin structure.

**Steps**:
1. Read Plan 14 section 6 for exact implementation details
2. Create all files in the specified order
3. Test all routes work
4. Verify sidebar and breadcrumbs

**Success Criteria**:
- All 25 routes accessible
- Pretty placeholders on unimplemented pages
- Navigation works correctly
- Build succeeds

---

**Session Goal**: Complete Phase 2A route hierarchy implementation.
