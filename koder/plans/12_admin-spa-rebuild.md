# Admin SPA Rebuild - Comprehensive Implementation Plan

**Date**: December 9, 2025
**Status**: Planning Complete - Ready for Implementation
**Tech Stack**: React + TypeScript + Vite + Tailwind CSS

---

## 1. Executive Summary

This plan details the complete rebuild of the Fazt admin interface as a modern, delightful React-based SPA. The admin will interface with all ~30 API endpoints, providing a comprehensive, native-like experience for managing sites, analytics, redirects, webhooks, and system configuration.

**Core Principles:**
- **Delight First**: This is the user's first impression of Fazt
- **Comprehensive**: Interface for all meaningful API endpoints
- **Extensible**: Architecture supports future API expansion
- **Performance**: Smart caching, no redundant fetches, instant feedback
- **Professional**: Elegant, minimal, functional design

---

## 2. Technology Stack

### Core
- **React 18+** - UI framework
- **TypeScript** - Type safety and developer experience
- **Vite** - Build tool (fast, simple)
- **Tailwind CSS** - Utility-first styling
- **React Router v6** - Hash-based routing

### State & Data
- **TanStack Query** (React Query) - API state management, caching
- **React Context** - Global app state (auth, theme, mock mode)
- **React Hook Form** - Complex form handling

### UI Components
- **Headless UI** - Accessible primitives (modals, dropdowns, tabs)
- **Lucide React** - Icon library (~1,500 icons)
- **Custom Components** - Built with Tailwind, fully controlled

### PWA
- **vite-plugin-pwa** - Service worker, manifest generation
- **Workbox** - Caching strategies

---

## 3. Project Structure

```
/admin                          # Development folder (repo root)
├── src/
│   ├── components/             # Reusable UI components
│   │   ├── ui/                 # Design system primitives
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Card.tsx
│   │   │   ├── Badge.tsx
│   │   │   ├── Skeleton.tsx
│   │   │   ├── Modal.tsx
│   │   │   ├── Dropdown.tsx
│   │   │   ├── Table.tsx
│   │   │   └── ...
│   │   ├── layout/             # Layout components
│   │   │   ├── AppShell.tsx    # Main layout container
│   │   │   ├── Navbar.tsx      # Top navigation
│   │   │   ├── Sidebar.tsx     # Side navigation
│   │   │   └── PageHeader.tsx
│   │   └── domain/             # Domain-specific components
│   │       ├── SiteCard.tsx
│   │       ├── AnalyticsChart.tsx
│   │       ├── RedirectForm.tsx
│   │       └── ...
│   │
│   ├── pages/                  # Page components (routes)
│   │   ├── Dashboard.tsx
│   │   ├── Sites.tsx
│   │   ├── SiteDetail.tsx
│   │   ├── Analytics.tsx
│   │   ├── Redirects.tsx
│   │   ├── Webhooks.tsx
│   │   ├── Settings.tsx
│   │   ├── Logs.tsx
│   │   ├── DesignSystem.tsx    # Component showcase
│   │   └── Login.tsx
│   │
│   ├── hooks/                  # Custom React hooks
│   │   ├── useAuth.ts
│   │   ├── useSites.ts
│   │   ├── useAnalytics.ts
│   │   ├── useTheme.ts
│   │   ├── useMockMode.ts
│   │   └── ...
│   │
│   ├── lib/                    # Libraries and utilities
│   │   ├── api.ts              # API client wrapper
│   │   ├── queryClient.ts      # TanStack Query config
│   │   ├── mockData.ts         # Mock data generator
│   │   ├── utils.ts            # Utility functions
│   │   └── constants.ts        # App constants
│   │
│   ├── context/                # React Context providers
│   │   ├── AuthContext.tsx
│   │   ├── ThemeContext.tsx
│   │   └── MockContext.tsx
│   │
│   ├── types/                  # TypeScript type definitions
│   │   ├── api.ts              # API response types
│   │   ├── models.ts           # Domain models
│   │   └── index.ts
│   │
│   ├── styles/                 # Global styles
│   │   └── globals.css         # Tailwind imports + customs
│   │
│   ├── App.tsx                 # Root component
│   ├── main.tsx                # Entry point
│   └── vite-env.d.ts
│
├── public/                     # Static assets
│   ├── favicon.ico
│   ├── logo.svg
│   ├── manifest.json           # PWA manifest
│   └── robots.txt
│
├── index.html                  # HTML template
├── vite.config.ts              # Vite configuration
├── tailwind.config.js          # Tailwind configuration
├── tsconfig.json               # TypeScript configuration
├── postcss.config.js           # PostCSS configuration
├── package.json
└── README.md
```

**Symlink Setup:**
```bash
# Create symlink from Go assets to admin build output
ln -s ../../../admin/dist internal/assets/system/admin
```

---

## 4. Design System

### 4.1 Color Palette

**Theme Color (Orange):**
```css
--color-primary: rgb(255, 149, 0);      /* #FF9500 */
--color-primary-dark: rgb(230, 134, 0);
--color-primary-light: rgb(255, 176, 51);
```

**Light Mode:**
```css
--bg-primary: rgb(255, 255, 255);
--bg-secondary: rgb(249, 250, 251);
--bg-tertiary: rgb(243, 244, 246);
--text-primary: rgb(17, 24, 39);
--text-secondary: rgb(107, 114, 128);
--border: rgb(229, 231, 235);
```

**Dark Mode:**
```css
--bg-primary: rgb(17, 24, 39);
--bg-secondary: rgb(31, 41, 55);
--bg-tertiary: rgb(55, 65, 81);
--text-primary: rgb(243, 244, 246);
--text-secondary: rgb(156, 163, 175);
--border: rgb(55, 65, 81);
```

**Semantic Colors:**
```css
--success: rgb(34, 197, 94);
--error: rgb(239, 68, 68);
--warning: rgb(251, 191, 36);
--info: rgb(59, 130, 246);
```

### 4.2 Typography

**Font:** Inter (Google Fonts)

**Scale:**
```css
text-xs:   0.75rem   (12px)
text-sm:   0.875rem  (14px)
text-base: 1rem      (16px)
text-lg:   1.125rem  (18px)
text-xl:   1.25rem   (20px)
text-2xl:  1.5rem    (24px)
text-3xl:  1.875rem  (30px)
text-4xl:  2.25rem   (36px)
```

### 4.3 Spacing

Tailwind default scale (0.25rem increments):
```
1: 0.25rem (4px)
2: 0.5rem  (8px)
3: 0.75rem (12px)
4: 1rem    (16px)
6: 1.5rem  (24px)
8: 2rem    (32px)
12: 3rem   (48px)
```

### 4.4 Component Library

**Primitives (src/components/ui/):**

1. **Button**
   - Variants: primary, secondary, ghost, danger
   - Sizes: sm, md, lg
   - States: default, hover, active, disabled, loading

2. **Input**
   - Types: text, email, password, number, url
   - States: default, focus, error, disabled
   - Addons: prefix, suffix, icon

3. **Card**
   - Variants: default, bordered, elevated
   - Sections: header, body, footer

4. **Badge**
   - Variants: default, success, error, warning, info
   - Sizes: sm, md, lg

5. **Modal**
   - Sizes: sm, md, lg, xl, full
   - Sections: header, body, footer
   - Backdrop: blur, dark

6. **Dropdown**
   - Trigger: button, custom
   - Position: auto, top, bottom, left, right
   - Items: clickable, with icons, dividers

7. **Table**
   - Features: sorting, pagination, selection
   - Responsive: scroll on mobile
   - Empty state, loading state

8. **Skeleton**
   - Variants: text, circle, rect
   - Animated pulse

9. **Tabs**
   - Variants: underline, pills, bordered
   - Orientation: horizontal

10. **Toast/Alert**
    - Variants: success, error, warning, info
    - Auto-dismiss, manual dismiss

11. **Spinner/Loader**
    - Sizes: sm, md, lg
    - Variants: spinner, dots, bars

---

## 5. Application Architecture

### 5.1 Layout Structure

```
┌─────────────────────────────────────┐
│         Navbar (fixed)              │
├────────┬────────────────────────────┤
│        │                            │
│ Side   │   Page Content             │
│ bar    │   (scrollable)             │
│ (fix)  │                            │
│        │                            │
└────────┴────────────────────────────┘
```

**Key Constraints:**
- Page never scrolls (html/body overflow: hidden)
- Only inner content areas scroll
- Fixed navbar + sidebar
- No layout janks (skeleton states)

### 5.2 Routing

**Hash-based routes:**
```
/#/                    → Dashboard
/#/sites               → Sites list
/#/sites/:id           → Site detail
/#/analytics           → Analytics overview
/#/redirects           → Redirects management
/#/webhooks            → Webhooks management
/#/settings            → Settings
/#/logs                → Logs viewer
/#/design-system       → Component showcase (dev mode)
```

**Implementation:**
```tsx
import { HashRouter, Routes, Route } from 'react-router-dom';

function App() {
  return (
    <HashRouter>
      <Routes>
        <Route path="/" element={<AppShell />}>
          <Route index element={<Dashboard />} />
          <Route path="sites" element={<Sites />} />
          <Route path="sites/:id" element={<SiteDetail />} />
          {/* ... */}
        </Route>
        <Route path="/login" element={<Login />} />
      </Routes>
    </HashRouter>
  );
}
```

### 5.3 State Management

**Three-tier approach:**

1. **Server State (TanStack Query)**
   - API data, caching, background refetch
   - Automatic loading/error states
   - Optimistic updates

2. **Global Client State (React Context)**
   - Auth state (user, session)
   - Theme (light/dark)
   - Mock mode (enabled/disabled)
   - Sidebar collapsed state

3. **Local Component State (useState)**
   - Form inputs
   - UI toggles
   - Temporary state

**Example: Auth Context**
```tsx
interface AuthContextValue {
  user: User | null;
  isAuthenticated: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue>(null!);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  // ... implementation
  return (
    <AuthContext.Provider value={{ user, isAuthenticated, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
```

### 5.4 API Integration

**Smart caching with TanStack Query:**

```tsx
// lib/queryClient.ts
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      cacheTime: 1000 * 60 * 30, // 30 minutes
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

// hooks/useSites.ts
export function useSites() {
  return useQuery({
    queryKey: ['sites'],
    queryFn: () => api.get('/api/sites'),
  });
}

export function useSite(id: string) {
  return useQuery({
    queryKey: ['sites', id],
    queryFn: () => api.get(`/api/sites/${id}`),
    enabled: !!id,
  });
}

export function useCreateSite() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSiteInput) => api.post('/api/sites', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sites'] });
    },
  });
}
```

**API Client (lib/api.ts):**
```tsx
import { useMockMode } from './useMockMode';
import { mockData } from './mockData';

class APIClient {
  private baseURL = '/api';

  async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    // Check mock mode
    if (useMockMode.getState().enabled) {
      return this.mockRequest<T>(endpoint);
    }

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      credentials: 'include', // Send cookies
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    });

    const json = await response.json();

    if (!response.ok) {
      throw new APIError(json.error);
    }

    return json.data;
  }

  get<T>(endpoint: string) {
    return this.request<T>(endpoint);
  }

  post<T>(endpoint: string, data: unknown) {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  // ... put, delete, patch
}

export const api = new APIClient();
```

### 5.5 Mock Data System

**Activation:**
```
URL: /#/?mock-data          → Enable mock mode
Storage: localStorage.mockMode = 'true'
Disable: window.mockMode.disable() or localStorage.removeItem('mockMode')
```

**Implementation:**
```tsx
// context/MockContext.tsx
export function MockProvider({ children }: { children: ReactNode }) {
  const [enabled, setEnabled] = useState(() => {
    const params = new URLSearchParams(window.location.search);
    const stored = localStorage.getItem('mockMode');
    return params.has('mock-data') || stored === 'true';
  });

  useEffect(() => {
    localStorage.setItem('mockMode', String(enabled));
    // Expose to window for debugging
    (window as any).mockMode = {
      enabled,
      enable: () => setEnabled(true),
      disable: () => setEnabled(false),
    };
  }, [enabled]);

  return (
    <MockContext.Provider value={{ enabled, setEnabled }}>
      {children}
    </MockContext.Provider>
  );
}
```

**Mock Data Structure (lib/mockData.ts):**
```tsx
export const mockData = {
  sites: [
    {
      id: 'site_1',
      name: 'my-blog',
      domain: 'blog.example.com',
      status: 'active',
      created_at: '2024-01-15T10:30:00Z',
      file_count: 42,
      total_size: 2048000,
    },
    // ... more sites
  ],
  analytics: {
    total_views: 12543,
    total_events: 8932,
    unique_visitors: 3421,
    // ... more stats
  },
  // ... other endpoints
};
```

---

## 6. Page Specifications

### 6.1 Dashboard (/#/)

**Purpose:** Overview of system health and key metrics

**Layout:**
- Top stats cards (4 cards)
  - Total Sites
  - Total Views (24h)
  - Total Events (24h)
  - Storage Used
- Recent activity timeline
- Quick actions (Create Site, View Analytics)
- System status

**API Endpoints:**
- `GET /api/system/health`
- `GET /api/system/config`
- `GET /api/sites` (count)
- `GET /api/analytics/stats`

**Components:**
- StatCard
- ActivityTimeline
- QuickActionButton

---

### 6.2 Sites (/#/sites)

**Purpose:** Manage all hosted sites

**Layout:**
- Page header with "Create Site" button
- Search/filter bar
- Sites grid/list (toggle view)
- Pagination

**Site Card:**
- Site name + domain
- Status badge
- File count, size
- Last deployed
- Actions: View, Edit, Deploy, Delete

**API Endpoints:**
- `GET /api/sites`
- `POST /api/sites` (create)
- `DELETE /api/sites/:id`
- `POST /api/deploy` (deploy)

**Components:**
- SiteCard
- CreateSiteModal
- ConfirmDeleteModal

---

### 6.3 Site Detail (/#/sites/:id)

**Purpose:** Detailed site management

**Tabs:**
1. **Overview**
   - Site info
   - Environment variables
   - API keys
   - Custom domain

2. **Files** (optional, or CLI-only)
   - File browser
   - Upload/delete files

3. **Analytics**
   - Site-specific analytics
   - Charts, tables

4. **Logs**
   - Console logs from serverless
   - Deployment logs

5. **Settings**
   - Edit name
   - Delete site
   - Advanced settings

**API Endpoints:**
- `GET /api/sites/:id`
- `PUT /api/sites/:id`
- `GET /api/env/:site`
- `POST /api/env/:site`
- `GET /api/keys`
- `POST /api/keys`
- `GET /api/sites/:site/logs`

---

### 6.4 Analytics (/#/analytics)

**Purpose:** System-wide analytics dashboard

**Layout:**
- Date range picker
- Stats overview (cards)
- Charts:
  - Views over time (line chart)
  - Top sites (bar chart)
  - Event types (pie chart)
  - Geographic distribution (if available)
- Events table (paginated)

**API Endpoints:**
- `GET /api/analytics/stats`
- `GET /api/analytics/events`
- `GET /api/analytics/domains`
- `GET /api/analytics/tags`

**Components:**
- DateRangePicker
- LineChart
- BarChart
- PieChart
- EventsTable

---

### 6.5 Redirects (/#/redirects)

**Purpose:** URL shortener / redirect management

**Layout:**
- Page header with "Create Redirect" button
- Search/filter
- Redirects table
  - Short code
  - Target URL
  - Click count
  - Created date
  - Actions: Edit, Delete, Copy Link

**API Endpoints:**
- `GET /api/redirects`
- `POST /api/redirects`
- `PUT /api/redirects/:id`
- `DELETE /api/redirects/:id`
- `GET /api/track/redirect/:id`

**Components:**
- RedirectForm
- RedirectsTable

---

### 6.6 Webhooks (/#/webhooks)

**Purpose:** Webhook endpoint management

**Layout:**
- Page header with "Create Webhook" button
- Webhooks list
  - Endpoint path
  - HTTP method
  - Created date
  - Recent requests
  - Actions: Edit, Delete, View Logs

**API Endpoints:**
- `GET /api/webhooks`
- `POST /api/webhooks`
- `PUT /api/webhooks/:id`
- `DELETE /api/webhooks/:id`

**Components:**
- WebhookForm
- WebhookCard
- WebhookLogsModal

---

### 6.7 Settings (/#/settings)

**Purpose:** System configuration

**Sections:**
1. **General**
   - Server domain
   - Port (display only)
   - Rate limits

2. **Authentication**
   - Change password
   - Session timeout

3. **Appearance**
   - Theme toggle (light/dark)
   - Mock mode toggle (dev)

4. **System**
   - Version info
   - System health
   - Database size

**API Endpoints:**
- `GET /api/system/config`
- `PUT /api/system/config` (if editable)
- `GET /api/system/health`
- `POST /api/auth/change-password` (if exists)

---

### 6.8 Logs (/#/logs)

**Purpose:** View system and site logs

**Layout:**
- Filter: Site selector, date range, log level
- Logs viewer (virtualized list)
- Search
- Export button

**API Endpoints:**
- `GET /api/sites/:site/logs`
- `GET /api/deployments` (deployment history)

**Components:**
- LogViewer
- LogFilter

---

### 6.9 Design System (/#/design-system)

**Purpose:** Component showcase (development aid)

**Visible during development, hidden in production**

**Sections:**
- Colors
- Typography
- Buttons
- Inputs
- Cards
- Badges
- Modals
- Dropdowns
- Tables
- Skeletons
- Icons
- Toasts/Alerts

**Navigation:**
- Show in sidebar during development
- Hide in production (or only accessible via direct URL)

---

## 7. Implementation Phases

### Phase 1: Foundation & Shell (PRIORITY)

**Goal:** Establish visual design direction and technical patterns

**Deliverables:**
1. ✅ Project setup
   - Vite + React + TypeScript
   - Tailwind CSS configured
   - Folder structure
   - Dependencies installed

2. ✅ Design system primitives
   - Color palette
   - Typography
   - UI components (Button, Input, Card, Badge, Skeleton, Modal)

3. ✅ App shell
   - Navbar (user menu, theme toggle)
   - Sidebar (navigation)
   - Layout (AppShell with fixed navbar/sidebar, scrollable content)

4. ✅ Core infrastructure
   - API client
   - TanStack Query setup
   - Auth context
   - Theme context
   - Mock context
   - Router setup

5. ✅ Sample pages (demonstrate patterns)
   - Login page
   - Dashboard (with mock data)
   - Sites page (list view)
   - Design System page

6. ✅ PWA basics
   - Manifest.json
   - Service worker shell
   - Icons/favicon

**Success Criteria:**
- Visually complete (looks like finished product)
- 3 pages fully functional with mock data
- All primitives in design system page
- Theme switching works
- Mock mode works
- Routing works
- Builds successfully

**Estimated Effort:** 1-2 weeks

---

### Phase 2: Core Features

**Goal:** Complete essential pages

**Deliverables:**
1. Site Detail page (full tabs)
2. Analytics page (with charts)
3. Settings page
4. Real API integration (replace mock)
5. Form validation
6. Error handling
7. Loading states everywhere

**Estimated Effort:** 2-3 weeks

---

### Phase 3: Advanced Features

**Goal:** Complete remaining pages

**Deliverables:**
1. Redirects page
2. Webhooks page
3. Logs viewer
4. Search functionality
5. Advanced filters
6. Export features

**Estimated Effort:** 1-2 weeks

---

### Phase 4: Polish & PWA

**Goal:** Production-ready

**Deliverables:**
1. Animations & transitions
2. Service worker (full caching)
3. Offline support
4. Performance optimization
5. Accessibility audit
6. Responsive refinements
7. Toast notifications system
8. Keyboard shortcuts
9. Hide Design System in production

**Estimated Effort:** 1 week

---

## 8. Development Workflow

### 8.1 Setup

```bash
# 1. Create admin folder at repo root
cd /home/testman/workspace
mkdir admin

# 2. Initialize Vite project
cd admin
npm create vite@latest . -- --template react-ts

# 3. Install dependencies
npm install
npm install -D tailwindcss postcss autoprefixer
npm install react-router-dom @tanstack/react-query
npm install @headlessui/react lucide-react
npm install react-hook-form @hookform/resolvers zod
npm install workbox-window
npm install -D vite-plugin-pwa

# 4. Initialize Tailwind
npx tailwindcss init -p

# 5. Move old admin files
cd /home/testman/workspace
mv internal/assets/system/admin admin-old

# 6. Create symlink (after first build)
# ln -s ../../../admin/dist internal/assets/system/admin
```

### 8.2 Development

```bash
# Start dev server
cd admin
npm run dev
# → http://localhost:5173

# Enable mock mode
# → http://localhost:5173/#/?mock-data
```

### 8.3 Build & Deploy

```bash
# Build for production
cd admin
npm run build
# → Output to admin/dist/

# Create symlink (first time)
cd /home/testman/workspace
ln -s ../../../admin/dist internal/assets/system/admin

# Rebuild Go binary (embeds admin/dist/)
go build -o fazt ./cmd/server

# Go will seed admin site to VFS on startup
./fazt server start
```

### 8.4 Go Integration

**Update `internal/database/seed.go` (or wherever system sites are seeded):**

```go
// Seed admin site from embedded filesystem
func seedAdminSite(db *sql.DB) error {
    // Read all files from assets.SystemFS at "system/admin"
    // Insert into files table with site_id = "admin"
    // This runs on server startup if admin site is missing
}
```

**Routing:** Admin already served at `admin.{domain}` by existing host router.

---

## 9. Configuration Files

### 9.1 vite.config.ts

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { VitePWA } from 'vite-plugin-pwa';

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'logo.svg'],
      manifest: {
        name: 'Fazt Admin',
        short_name: 'Fazt',
        description: 'Fazt Platform Administration',
        theme_color: '#ff9500',
        background_color: '#ffffff',
        display: 'standalone',
        icons: [
          {
            src: 'icon-192.png',
            sizes: '192x192',
            type: 'image/png',
          },
          {
            src: 'icon-512.png',
            sizes: '512x512',
            type: 'image/png',
          },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
        runtimeCaching: [
          {
            urlPattern: /^https:\/\/fonts\.googleapis\.com\/.*/i,
            handler: 'CacheFirst',
            options: {
              cacheName: 'google-fonts-cache',
              expiration: {
                maxEntries: 10,
                maxAgeSeconds: 60 * 60 * 24 * 365, // 1 year
              },
            },
          },
        ],
      },
    }),
  ],
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
  },
});
```

### 9.2 tailwind.config.js

```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class', // Enable class-based dark mode
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: 'rgb(255, 149, 0)',
          dark: 'rgb(230, 134, 0)',
          light: 'rgb(255, 176, 51)',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
};
```

### 9.3 tsconfig.json

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

---

## 10. Testing Strategy

### 10.1 Component Tests

```bash
npm install -D vitest @testing-library/react @testing-library/jest-dom
```

**Test primitives:**
- Button variants, states
- Input validation
- Modal open/close
- Theme switching

### 10.2 Integration Tests

- Page rendering
- API integration with mock
- Form submission flows
- Navigation

### 10.3 E2E Tests (Optional)

- Playwright or Cypress
- Critical user flows

---

## 11. Key Technical Decisions

### 11.1 Why React over Vanilla?
- **Velocity**: 3-5x faster development at this scale
- **Maintainability**: Declarative UI, automatic updates
- **Ecosystem**: TanStack Query, Headless UI, proven patterns
- **Familiarity**: User already knows React

### 11.2 Why TypeScript?
- Type safety prevents bugs
- Better IDE support
- Self-documenting code
- Worth the slight overhead at this scale

### 11.3 Why TanStack Query?
- Solves "smart caching" requirement perfectly
- Background refetch, deduplication
- Loading/error states built-in
- Industry standard

### 11.4 Why Headless UI?
- Accessibility (ARIA, keyboard nav) handled
- Style with Tailwind (full control)
- No opinionated design
- Perfect for custom design systems

### 11.5 Why Hash Routing?
- SPA has no server-side routes
- Simpler than history API for embedded context
- No server config needed

---

## 12. Success Metrics

**Phase 1 Complete When:**
- [ ] Design system page shows all primitives
- [ ] 3 sample pages functional with mock data
- [ ] Theme switching works (light/dark, persists)
- [ ] Mock mode works (URL param, localStorage, window.mockMode)
- [ ] Layout complete (navbar, sidebar, no scroll)
- [ ] Builds without errors
- [ ] Looks visually complete

**Full Project Complete When:**
- [ ] All ~12 pages implemented
- [ ] All API endpoints integrated
- [ ] Forms validated, errors handled
- [ ] Loading states everywhere (skeletons)
- [ ] PWA installable, works offline
- [ ] Responsive (mobile, tablet, desktop)
- [ ] Accessible (WCAG AA)
- [ ] Performance: Lighthouse score > 90

---

## 13. Future Enhancements

**Post-v1:**
- Real-time updates (WebSocket integration)
- Drag-and-drop file uploads
- Visual analytics builder
- Advanced search with filters
- Keyboard shortcuts overlay
- Multi-language support (i18n)
- Export data (CSV, JSON)
- Audit log viewer
- User management (if multi-user added)

---

## 14. References

### API Documentation
- `koder/plans/11_api-standardization.md` - Complete API spec
- `koder/docs/admin-api/` - API examples

### Design Resources
- Theme color: `rgb(255, 149, 0)`
- Icons: Lucide React (https://lucide.dev)
- Font: Inter (https://fonts.google.com/specimen/Inter)

### Technical Stack
- React: https://react.dev
- Vite: https://vitejs.dev
- Tailwind CSS: https://tailwindcss.com
- TanStack Query: https://tanstack.com/query
- Headless UI: https://headlessui.com
- React Router: https://reactrouter.com

---

## 15. Next Steps

### Immediate (This Session)
1. ✅ Plan complete
2. Move old admin files to `/admin-old`
3. Update `NEXT_SESSION.md` for implementation

### Next Session (Implementation)
1. Setup project (Vite + deps)
2. Configure Tailwind
3. Create folder structure
4. Build design system primitives
5. Implement app shell
6. Create 3 sample pages
7. Test mock mode
8. Test build & embed

---

**Plan Status:** ✅ COMPLETE
**Ready for:** Implementation (Phase 1)
**Estimated Timeline:** 4-6 weeks total (Phase 1: 1-2 weeks)

---

**Remember:** Build to delight. This is the first impression of Fazt.
