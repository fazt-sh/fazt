# Next Session Handoff: Admin SPA Planning & Design

**Date**: December 9, 2025
**Status**: 🟢 **READY FOR PLANNING** - API Complete, SPA Design Phase
**Current Phase**: Planning (Admin SPA Rebuild)
**Recommended Model**: Sonnet (design & architecture discussion)
**Branch**: gemini/api-reality

---

## 📋 Context Payload (Read These First)

When starting this session, read these files in order:

1. **Your Vision**: `koder/rough.md` ⭐ **START HERE - Your scattered thoughts**
2. **API Reality**: `koder/plans/11_api-standardization.md` (What you're building on)
3. **API Status**: Previous section below (What's complete)

---

## ✅ What's Complete (Previous Session)

### API Standardization (100% Done)
- ✅ All 11 handlers migrated to standardized format
- ✅ Response format: `{"data": ...}` for success, `{"error": {...}}` for errors
- ✅ Zero legacy patterns remaining
- ✅ All tests passing
- ✅ SPA-ready API with predictable structure

**Key for SPA Development:**
```javascript
// Single fetch wrapper works for ALL endpoints
async function apiFetch(url, options) {
  const resp = await fetch(url, options);
  const json = await resp.json();
  return resp.ok ? json.data : throw new Error(json.error.message);
}
```

**Endpoints Available (~30 total):**
- Authentication: login, logout, user/me, auth/status
- System: health, config, limits
- Hosting: sites CRUD, deploy, env vars, API keys
- Analytics: events, stats, domains, tags
- Redirects: CRUD operations
- Webhooks: CRUD operations
- Tracking: pixel, redirect tracking
- Logs: site logs, deployment history

---

## 🎯 Your Mission (This Session)

### Phase 1: Review & Discussion (Start Here)

**Read `koder/rough.md` and discuss with the user:**

1. **Scope Clarification**
   - Which features are MVP vs Nice-to-Have?
   - Which API endpoints to integrate first?
   - What's the page/route structure?

2. **Technical Decisions**
   - Component architecture pattern (Web Components? Custom elements? Plain modules?)
   - State management approach (simple object? event-driven? state machine?)
   - Routing strategy (hash-based? history API? single page with tabs?)
   - Mock data strategy details

3. **Design System**
   - Color palette beyond theme orange?
   - Typography scale (Inter font usage)?
   - Spacing/sizing system?
   - Icon library choice?

4. **PWA Requirements**
   - Offline-first? Or just installable?
   - Service worker strategy?
   - Cache strategy for API calls?

5. **Development Workflow**
   - Where to develop? (`admin/` folder in repo? Separate folder?)
   - How to embed in Go binary? (go:embed?)
   - Local dev server setup?

### Phase 2: Create Comprehensive Plan

After discussion, create: **`koder/plans/12_admin-spa-rebuild.md`**

**Plan should include:**
1. **Architecture Overview**
   - Folder structure
   - Component hierarchy
   - State management pattern
   - Routing strategy

2. **Design System**
   - Color palette
   - Typography
   - Spacing/layout
   - Component library (buttons, inputs, cards, etc.)

3. **Features & Pages**
   - Page breakdown with wireframes/descriptions
   - API integration per page
   - Mock data structure

4. **Technical Specifications**
   - ES6 module structure
   - State management implementation
   - API client abstraction
   - Mock data system

5. **Implementation Phases**
   - Phase 1: Foundation (layout, routing, state, API client)
   - Phase 2: Core pages (dashboard, sites, analytics)
   - Phase 3: Advanced features (webhooks, redirects, settings)
   - Phase 4: Polish (PWA, dark mode, animations)

6. **File Structure**
   ```
   admin/
   ├── index.html
   ├── css/
   │   └── app.css
   ├── js/
   │   ├── app.js (entry point)
   │   ├── router.js
   │   ├── state.js
   │   ├── api.js
   │   ├── mock.js
   │   ├── components/
   │   │   ├── Layout.js
   │   │   ├── Sidebar.js
   │   │   ├── ...
   │   └── pages/
   │       ├── Dashboard.js
   │       ├── Sites.js
   │       ├── ...
   └── assets/
       ├── favicon.ico
       └── ...
   ```

---

## 🎨 Key Requirements from `koder/rough.md`

### Core Constraints
- ✅ **Vanilla SPA** - No build tools (no webpack, vite, etc.)
- ✅ **PWA-capable** - Service worker, manifest
- ✅ **Independent folder** - Can develop separate from Go code
- ✅ **Tailwind CSS CDN** - No npm, just CDN link

### Features
- ✅ **Mock data system** - `?mock-data` query string, localStorage persistence
- ✅ **Light/Dark mode** - Theme switcher
- ✅ **Theme color**: `rgb(255, 149, 0)` (orange)
- ✅ **Inter font** - Typography
- ✅ **Icons** - Nice icon library (heroicons? lucide?)
- ✅ **No layout janks** - Fixed layout, only panels scroll
- ✅ **Memoized API** - Smart caching, avoid redundant fetches
- ✅ **Loading skeletons** - Graceful loading states
- ✅ **Native-like UX** - Super spiffy, fast, responsive

### Architecture
- ✅ **ES6 modules** - Native module system
- ✅ **Component-based** - Reusable component pattern
- ✅ **State management** - Extensible for future API changes
- ✅ **API-driven UI** - UI structure mirrors API design

### Philosophy
> "Build the UI to delight. This is the app that will introduce users to Fazt. Make it elegant, functional, and a delight to work with."

---

## 🤔 Questions to Ask User

Before creating the plan, clarify:

1. **Page Structure**
   - What pages/routes? (Dashboard, Sites, Analytics, Settings, ...?)
   - Navigation pattern? (Sidebar? Top nav? Tabs?)

2. **Priority Features**
   - What's the absolute MVP? (Just sites management? Or include analytics too?)
   - Which features can wait for v2?

3. **Mock Data**
   - Mock ALL endpoints? Or just specific ones?
   - Should mock data be realistic? (Random names/dates?)

4. **Component Architecture**
   - Preference for pattern? (Web Components? Class-based? Functional modules?)
   - How granular? (Many small components vs few large ones?)

5. **State Management**
   - Simple object? Event emitter? Pub/sub? State machine?
   - Global state vs local component state?

6. **Development Location**
   - Develop in `admin/` folder at repo root?
   - Or `internal/assets/system/admin/`?
   - Embed strategy? (go:embed on build?)

7. **Asset Sources**
   - Use assets from `docs/` site for now?
   - Specific logo/favicon preferences?

---

## 📚 Resources Available

### API Documentation
- `koder/plans/11_api-standardization.md` - Complete API spec
- `koder/docs/admin-api/request-response.md` - Example requests/responses
- All endpoints return: `{"data": ...}` or `{"error": {"code": "...", "message": "..."}}`

### Design Inspiration
- Theme color: `rgb(255, 149, 0)` (vibrant orange)
- Font: Inter (Google Fonts or CDN)
- Icons: Heroicons? Lucide? Feather?
- Style: Minimal, functional, elegant

### Technical Constraints
- No build process (vanilla JS/HTML/CSS)
- Tailwind CSS via CDN
- ES6 modules (native)
- Mock data system for development

---

## 🎯 Expected Outcome

By end of this session:

1. ✅ **Discussion complete** - All questions answered
2. ✅ **Plan created**: `koder/plans/12_admin-spa-rebuild.md`
3. ✅ **Plan includes**:
   - Complete architecture
   - File structure
   - Component breakdown
   - Page specifications
   - Implementation phases
   - Design system
4. ✅ **Ready to build** - Next session can start coding immediately

---

## 🚀 Quick Start (For Next Session)

**Command to start:**
```bash
read and execute koder/start.md
```

**What will happen:**
1. You (Sonnet) will read this file
2. You'll read `koder/rough.md`
3. You'll ask user clarifying questions
4. You'll discuss and refine the vision
5. You'll create comprehensive plan in `koder/plans/12_admin-spa-rebuild.md`
6. You'll update this file for the NEXT session (actual build)

---

## 📝 Notes for Next Session

- **Don't start building yet** - This session is for planning only
- **Ask questions** - Better to clarify now than rebuild later
- **Be thorough** - The plan should be detailed enough for implementation
- **Consider extensibility** - API will grow, UI should accommodate that

---

**Session Goal**: Exit with a crystal-clear plan that makes building the SPA straightforward.

**Remember**: The user has a clear vision in `rough.md`. Your job is to refine it into an actionable, comprehensive plan.
