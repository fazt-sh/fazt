# Next Session Handoff: Admin SPA Implementation (Phase 1)

**Date**: December 9, 2025
**Status**: 🟢 **READY FOR IMPLEMENTATION** - Plan Complete, Phase 1 Start
**Current Phase**: Implementation (Phase 1 - Foundation & Shell)
**Recommended Model**: Sonnet (hands-on coding)
**Branch**: master

---

## 📋 Context Payload (Read These First)

When starting this session, read these files in order:

1. **The Plan**: `koder/plans/12_admin-spa-rebuild.md` ⭐ **START HERE - Complete implementation plan**
2. **API Reference**: `koder/plans/11_api-standardization.md` (API endpoints to interface with)
3. **Architecture**: `koder/analysis/04_comprehensive_technical_overview.md` (System overview)

**Optional Reference:**
- `koder/rough.md` - Original vision and requirements
- `/admin-old/` - Previous admin implementation (moved from `internal/assets/system/admin/`)

---

## ✅ What's Complete (Previous Session)

### Planning Phase (100% Done)
- ✅ Requirements gathering and clarification
- ✅ Technology stack decided (React + TypeScript + Vite + Tailwind)
- ✅ Architecture designed (component structure, state management, routing)
- ✅ Design system specified (colors, typography, components)
- ✅ All 12 pages planned with specifications
- ✅ Implementation phases defined
- ✅ Comprehensive plan document created
- ✅ Old admin files moved to `/admin-old/`

### Tech Stack Finalized
- **Core**: React 18 + TypeScript + Vite
- **Styling**: Tailwind CSS (PostCSS, not CDN)
- **State**: TanStack Query + React Context
- **Forms**: React Hook Form
- **UI**: Headless UI + Custom components
- **Icons**: Lucide React
- **Routing**: React Router v6 (hash mode)
- **PWA**: vite-plugin-pwa

---

## 🎯 Your Mission (This Session)

### Phase 1: Foundation & Shell

**Goal:** Build the visual and technical foundation. Even with just 3 pages, it should look and feel like the finished product.

**Deliverables:**

1. **Project Setup**
   - [ ] Create `/admin` folder at repo root
   - [ ] Initialize Vite + React + TypeScript
   - [ ] Install and configure Tailwind CSS
   - [ ] Install dependencies (TanStack Query, React Router, Headless UI, Lucide React, etc.)
   - [ ] Set up folder structure (see plan section 3)
   - [ ] Configure vite.config.ts (see plan section 9.1)
   - [ ] Configure tailwind.config.js (see plan section 9.2)
   - [ ] Configure tsconfig.json (see plan section 9.3)

2. **Design System Primitives**
   - [ ] Create color palette (theme orange, light/dark modes)
   - [ ] Set up Inter font (Google Fonts)
   - [ ] Implement UI components in `src/components/ui/`:
     - [ ] Button (variants: primary, secondary, ghost, danger)
     - [ ] Input (with error states, icons)
     - [ ] Card (header, body, footer)
     - [ ] Badge (variants: success, error, warning, info)
     - [ ] Skeleton (text, circle, rect)
     - [ ] Modal (using Headless UI Dialog)
     - [ ] Dropdown (using Headless UI Menu)
     - [ ] Spinner/Loader

3. **Core Infrastructure**
   - [ ] API client (`src/lib/api.ts`) with standardized response handling
   - [ ] TanStack Query setup (`src/lib/queryClient.ts`)
   - [ ] Auth context (`src/context/AuthContext.tsx`)
   - [ ] Theme context (`src/context/ThemeContext.tsx`) - light/dark mode
   - [ ] Mock context (`src/context/MockContext.tsx`) - ?mock-data support
   - [ ] Mock data file (`src/lib/mockData.ts`) with sample data

4. **App Shell**
   - [ ] Layout components in `src/components/layout/`:
     - [ ] AppShell.tsx (fixed navbar + sidebar, scrollable content)
     - [ ] Navbar.tsx (logo, user menu, theme toggle)
     - [ ] Sidebar.tsx (navigation links, collapsible)
     - [ ] PageHeader.tsx (page title, actions)
   - [ ] React Router setup (hash mode)
   - [ ] Layout constraint: Page never scrolls, only content areas

5. **Sample Pages (3 pages to establish patterns)**
   - [ ] Login page (`src/pages/Login.tsx`)
     - Form with username/password
     - Error handling
     - Redirect on success
   - [ ] Dashboard (`src/pages/Dashboard.tsx`)
     - 4 stat cards (Sites, Views, Events, Storage)
     - Quick actions section
     - Uses mock data
   - [ ] Sites page (`src/pages/Sites.tsx`)
     - Site cards in grid
     - Create button
     - Search/filter
     - Uses mock data
   - [ ] Design System page (`src/pages/DesignSystem.tsx`)
     - Showcase all UI components
     - Color palette display
     - Typography samples
     - Visible in sidebar during development

6. **PWA Basics**
   - [ ] Configure vite-plugin-pwa
   - [ ] Create manifest.json (theme color: orange)
   - [ ] Service worker shell
   - [ ] Add icons/favicon

7. **Build & Integration Test**
   - [ ] Test `npm run dev` works
   - [ ] Test mock mode: `?mock-data` activates
   - [ ] Test theme switching persists
   - [ ] Test routing works
   - [ ] Build: `npm run build` creates `admin/dist/`
   - [ ] Create symlink: `ln -s ../../../admin/dist internal/assets/system/admin`
   - [ ] Test Go embed: Rebuild binary, verify admin loads

**Success Criteria:**
- ✅ Visually complete (looks like finished product)
- ✅ 3 pages fully functional with mock data
- ✅ All primitives showcased in design system page
- ✅ Theme switching works (persists to localStorage)
- ✅ Mock mode works (?mock-data, localStorage, window.mockMode)
- ✅ Layout perfect (no scroll jank, fixed navbar/sidebar)
- ✅ Builds successfully and embeds in Go binary

---

## 📚 Key References

### From the Plan (`koder/plans/12_admin-spa-rebuild.md`)

**File Structure:** Section 3
**Design System:** Section 4
**Architecture:** Section 5
**Configuration Files:** Section 9

### Setup Commands

```bash
# 1. Create admin folder
cd /home/testman/workspace
mkdir admin
cd admin

# 2. Initialize Vite
npm create vite@latest . -- --template react-ts

# 3. Install dependencies
npm install
npm install -D tailwindcss postcss autoprefixer
npm install react-router-dom @tanstack/react-query
npm install @headlessui/react lucide-react
npm install react-hook-form @hookform/resolvers zod
npm install -D vite-plugin-pwa

# 4. Initialize Tailwind
npx tailwindcss init -p

# 5. Start development
npm run dev
```

### Design Tokens

**Theme Color:** `rgb(255, 149, 0)` (#FF9500)

**Light Mode:**
- bg-primary: `rgb(255, 255, 255)`
- bg-secondary: `rgb(249, 250, 251)`
- text-primary: `rgb(17, 24, 39)`

**Dark Mode:**
- bg-primary: `rgb(17, 24, 39)`
- bg-secondary: `rgb(31, 41, 55)`
- text-primary: `rgb(243, 244, 246)`

**Font:** Inter (Google Fonts)

---

## 🎨 Development Approach

### Build Order (Recommended)

1. **Start with infrastructure** (boring but essential)
   - Vite setup
   - Tailwind config
   - Folder structure
   - Contexts (Auth, Theme, Mock)
   - API client

2. **Build design system** (fun, establishes patterns)
   - Colors in CSS variables
   - UI components (Button, Input, Card, etc.)
   - Design System page to showcase

3. **Build layout** (shell)
   - AppShell component
   - Navbar with theme toggle
   - Sidebar with navigation

4. **Build pages** (one by one)
   - Login (simple form)
   - Dashboard (stat cards, mock data)
   - Sites (grid of cards, mock data)

5. **Test everything**
   - Mock mode activation
   - Theme switching
   - Routing
   - Build & embed

### Key Patterns to Establish

**API Hook Pattern:**
```tsx
// hooks/useSites.ts
export function useSites() {
  const { enabled: mockMode } = useMockMode();
  return useQuery({
    queryKey: ['sites'],
    queryFn: () => (mockMode ? mockData.sites : api.get('/api/sites')),
  });
}
```

**Page Component Pattern:**
```tsx
// pages/Sites.tsx
export function Sites() {
  const { data: sites, isLoading } = useSites();

  if (isLoading) return <PageSkeleton />;

  return (
    <div className="p-6">
      <PageHeader
        title="Sites"
        action={<Button>Create Site</Button>}
      />
      <div className="grid grid-cols-3 gap-4 mt-6">
        {sites.map(site => <SiteCard key={site.id} site={site} />)}
      </div>
    </div>
  );
}
```

---

## 🛠️ Troubleshooting

### Common Issues

**Tailwind classes not working:**
- Check `tailwind.config.js` content paths
- Verify `globals.css` imports Tailwind directives
- Restart dev server

**React Router not loading pages:**
- Use `HashRouter`, not `BrowserRouter`
- Check route paths start with `/`

**Mock mode not activating:**
- Check `?mock-data` in URL
- Verify localStorage is being set
- Check `useMockMode()` hook implementation

**Build fails:**
- Check all imports are correct
- Verify TypeScript types are defined
- Run `npm install` again

**Symlink not working:**
- Use relative path: `../../../admin/dist`
- Verify `admin/dist/` exists (run build first)
- Check symlink with `ls -la internal/assets/system/admin`

---

## 🚀 Quick Start (For This Session)

**Command to start:**
```bash
read and execute koder/start.md
```

**What will happen:**
1. You'll read `koder/NEXT_SESSION.md` (this file)
2. You'll read `koder/plans/12_admin-spa-rebuild.md` (the plan)
3. You'll set up the project (Vite + React + TypeScript)
4. You'll build the foundation (primitives, layout, 3 pages)
5. You'll test everything works (mock mode, theme, build)

---

## 📝 Notes for Implementation

### Important Constraints

1. **No scroll on page** - Only inner panels scroll
2. **Mock data transparent** - Components shouldn't know if using mock
3. **Theme persists** - Save to localStorage
4. **Loading states everywhere** - Use skeletons, not spinners alone
5. **Design system first** - Establish patterns before building pages
6. **Visually complete** - Even with 3 pages, it should look finished

### What to Avoid

- ❌ Don't build all 12 pages yet - just 3 for Phase 1
- ❌ Don't integrate real API yet - mock data only for Phase 1
- ❌ Don't over-engineer - Keep it simple, establish patterns
- ❌ Don't skip Design System page - crucial for consistency

### What Success Looks Like

After Phase 1, you should be able to:
- Open `http://localhost:5173/#/?mock-data`
- See beautiful login page
- Log in (mock auth)
- See dashboard with stats
- Navigate to sites page
- See grid of site cards
- Click Design System in sidebar
- See all primitives
- Toggle theme (light/dark)
- Refresh - theme persists
- Build successfully
- Embed in Go binary

---

## 🎯 Expected Outcome

By end of this session:

1. ✅ **Project set up** - Vite + React + TypeScript running
2. ✅ **Design system complete** - All primitives built and showcased
3. ✅ **App shell complete** - Navbar, sidebar, layout
4. ✅ **3 pages working** - Login, Dashboard, Sites
5. ✅ **Core infrastructure** - API client, contexts, routing
6. ✅ **Mock mode working** - ?mock-data activates
7. ✅ **Theme working** - Light/dark toggle persists
8. ✅ **Builds successfully** - npm run build works
9. ✅ **Embeds in Go** - Symlink + rebuild works

**Visual Quality:** Should look like a finished product, even with just 3 pages.

---

## 📅 Next Session After This

**Phase 2: Core Features**
- Complete remaining pages (Analytics, Settings, Site Detail)
- Real API integration (replace mock)
- Form validation
- Error handling
- Advanced components (charts, tables)

---

## 🔗 Reference Links

**Plan:** `koder/plans/12_admin-spa-rebuild.md` (comprehensive spec)
**API:** `koder/plans/11_api-standardization.md` (endpoints)
**Architecture:** `koder/analysis/04_comprehensive_technical_overview.md`
**Old Admin:** `/admin-old/` (reference only, don't use)

---

**Session Goal:** Exit with a beautiful, functional foundation that demonstrates the final product's direction.

**Remember:** Build to delight. Quality over speed. Establish patterns that make Phase 2+ easy.
