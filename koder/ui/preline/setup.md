# Preline UI Integration Guide for Fazt Admin

**Objective**: Convert the current Fazt admin UI from custom components to Preline UI components while maintaining theme flexibility and functionality.

**Last Updated**: December 9, 2025

---

## 1. Overview

### Current State
- Custom React components in `/admin/src/components/ui/`
- Custom CSS variables for theming in `/admin/src/styles/globals.css`
- Tailwind CSS with custom color scheme
- Components: Button, Input, Card, Badge, Skeleton, Modal, Dropdown, Spinner, Terminal

### Target State
- Preline UI components with React integration
- Maintain Fazt brand colors via CSS variables
- Component wrappers for easy migration
- Same API with Preline internals

---

## 2. Preline UI Installation

### Prerequisites
- React 19.0.0+ (current: matches)
- Vite 6.3.1+ (current: matches)
- Tailwind CSS already configured

### Installation Commands

```bash
cd /home/testman/workspace/admin

# Install Preline UI
npm install preline

# Install required Tailwind CSS Forms plugin
npm install -D @tailwindcss/forms

# Optional: Install third-party libraries for advanced components
npm install jquery lodash datatables.net dropzone vanilla-calendar-pro nouislider
```

### Configuration Updates

#### Update `src/index.css`
```css
@import "tailwindcss";

@import "preline/variants.css";
@source "../node_modules/preline/dist/*.js";

/* Optional Preline UI Datepicker Plugin */
/* @import "preline/src/plugins/datepicker/styles.css"; */
```

#### Create `src/global.d.ts` (if not exists)
```typescript
import type { DataTable } from "datatables.net";
import type { Dropzone } from "dropzone";
import type { VanillaCalendarPro } from "vanilla-calendar-pro";
import type { noUiSlider } from "nouislider";
import type { IStaticMethods } from "preline/dist";

declare global {
  interface Window {
    _: typeof import("lodash");
    $: typeof import("jquery");
    jQuery: typeof import("jquery");
    DataTable: typeof DataTable;
    Dropzone: typeof Dropzone;
    VanillaCalendarPro: typeof VanillaCalendarPro;
    noUiSlider: typeof noUiSlider;
    HSStaticMethods: IStaticMethods;
  }
}

export {};
```

#### Update `src/main.tsx`
```typescript
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App.tsx';
import './index.css';

// Import external libraries for Preline
import $ from 'jquery';
import _ from 'lodash';
import noUiSlider from 'nouislider';
import 'datatables.net';
import 'dropzone/dist/dropzone-min.js';
import * as VanillaCalendarPro from 'vanilla-calendar-pro';

// Attach to window for Preline
window._ = _;
window.$ = $;
window.jQuery = $;
window.DataTable = $.fn.dataTable;
window.noUiSlider = noUiSlider;
window.VanillaCalendarPro = VanillaCalendarPro;

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
```

#### Update `src/App.tsx`
```typescript
import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';

async function loadPreline() {
  return import('preline/dist/index.js');
}

function App() {
  const location = useLocation();

  useEffect(() => {
    const initPreline = async () => {
      await loadPreline();

      if (
        window.HSStaticMethods &&
        typeof window.HSStaticMethods.autoInit === 'function'
      ) {
        window.HSStaticMethods.autoInit();
      }
    };

    initPreline();
  }, [location.pathname]);

  // ... rest of App component
}

export default App;
```

#### Update `tailwind.config.js`
```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
    "./node_modules/preline/dist/*.js", // Add Preline content
  ],
  plugins: [
    require('@tailwindcss/forms'), // Add forms plugin
    require('preline/plugin'), // Add Preline plugin
  ],
  // ... rest of config
}
```

---

## 3. Component Mapping Strategy

### Current Components → Preline Equivalents

| Current Component | Preline Component | File Location | Priority |
|------------------|-------------------|--------------|----------|
| Button | Button | Preline docs: Buttons | **High** |
| Input | Input | Preline docs: Inputs | **High** |
| Card | Card | Preline docs: Cards | **High** |
| Badge | Badge | Preline docs: Badges | **High** |
| Skeleton | Skeleton | Preline docs: Skeletons | **Medium** |
| Spinner | Spinner | Preline docs: Spinners | **Medium** |
| Modal | Modal | Preline docs: Modal | **Medium** |
| Dropdown | Dropdown | Preline docs: Dropdown | **Medium** |
| Terminal | (Custom) | Build from Preline Card | **Low** |

### Wrapper Component Pattern

Create wrapper components in `/admin/src/components/ui-preline/`:

```tsx
// Example: Button.tsx
import React from 'react';
import type { ButtonHTMLAttributes } from 'react';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger';
  size?: 'sm' | 'md' | 'lg';
}

export function Button({
  variant = 'primary',
  size = 'md',
  className = '',
  children,
  ...props
}: ButtonProps) {
  const baseClasses = 'inline-flex items-center gap-x-2 font-medium rounded-lg border transition-all duration-150 focus:outline-none focus:ring-2 focus:ring-[rgb(var(--accent-mid))]/50 disabled:opacity-50 disabled:pointer-events-none';

  const sizeClasses = {
    sm: 'py-2 px-3 text-sm',
    md: 'py-3 px-4 text-sm',
    lg: 'p-4 sm:p-5 text-sm',
  };

  const variantClasses = {
    primary: 'border-transparent bg-gradient-to-r from-[rgb(var(--accent-start))] to-[rgb(var(--accent-mid))] text-white hover:from-[rgb(var(--accent-start))]/90 hover:to-[rgb(var(--accent-mid))]/90 shadow-sm hover:shadow-md hover-glow active:scale-[0.98]',
    secondary: 'border-[rgb(var(--border-primary))] bg-[rgb(var(--bg-elevated))] text-[rgb(var(--text-primary))] hover:bg-[rgb(var(--bg-hover))]',
    ghost: 'border-transparent text-[rgb(var(--text-secondary))] hover:text-[rgb(var(--text-primary))] hover:bg-[rgb(var(--bg-hover))]',
    danger: 'border-transparent bg-[rgb(var(--accent-start))] text-white hover:bg-[rgb(var(--accent-start))]/90',
  };

  return (
    <button
      className={`${baseClasses} ${sizeClasses[size]} ${variantClasses[variant]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}
```

---

## 4. Theme Integration

### Maintain Current CSS Variables

Keep the current theme system in `globals.css`:

```css
:root {
  /* Docs-inspired Accent Palette */
  --accent-start: 255 59 48;    /* #ff3b30 - Bright red-orange */
  --accent-mid: 255 149 0;      /* #ff9500 - Orange/yellow */
  --accent-end: 255 204 0;      /* #ffcc00 - Yellow */
  --accent: var(--accent-mid);
  --accent-glow: 255 59 48 / 0.1;

  /* Dark Mode - Docs-inspired */
  --bg-base: 5 5 5;
  --bg-elevated: 20 20 20;
  --bg-subtle: 26 26 26;
  --bg-hover: 35 35 35;
  /* ... rest of variables */
}
```

### Preline Overrides

Create `src/styles/preline-overrides.css`:

```css
/* Override Preline defaults with Fazt theme */

/* Buttons */
.hs-button-primary {
  background-color: rgb(var(--accent-mid));
  border-color: rgb(var(--accent-mid));
}

.hs-button-primary:hover {
  background-color: rgb(var(--accent-start));
  border-color: rgb(var(--accent-start));
}

/* Inputs */
.hs-input:focus {
  border-color: rgb(var(--accent-mid));
  --tw-ring-color: rgb(var(--accent-mid));
}

/* Dropdowns */
.hs-dropdown-menu {
  background-color: rgb(var(--bg-elevated));
  border-color: rgb(var(--border-primary));
}

.hs-dropdown-item:hover {
  background-color: rgb(var(--bg-hover));
}

/* Modals */
.hs-modal-backdrop {
  background-color: rgba(0, 0, 0, 0.5);
}

.hs-modal-content {
  background-color: rgb(var(--bg-elevated));
  border-color: rgb(var(--border-primary));
}

/* Cards */
.hs-card {
  background-color: rgb(var(--bg-elevated));
  border-color: rgb(var(--border-primary));
}

/* Dark mode overrides */
.dark .hs-dropdown-menu {
  background-color: rgb(var(--bg-elevated));
  border-color: rgb(var(--border-primary));
}
```

Import in `index.css`:
```css
@import "./preline-overrides.css";
```

---

## 5. Migration Steps

### Step 1: Setup Phase
1. Install Preline and dependencies
2. Update configuration files
3. Create wrapper component structure
4. Test basic Preline component

### Step 2: Component Migration (High Priority)
1. **Button**
   - Create `/admin/src/components/ui-preline/Button.tsx`
   - Replace imports in files one by one
   - Test all variants and sizes

2. **Input**
   - Create `/admin/src/components/ui-preline/Input.tsx`
   - Maintain validation and error states
   - Test all input types

3. **Card**
   - Create `/admin/src/components/ui-preline/Card.tsx`
   - Support header, body, footer variants
   - Maintain glassmorphism effects

4. **Badge**
   - Create `/admin/src/components/ui-preline/Badge.tsx`
   - Support all color variants
   - Add soft/hard variants

### Step 3: Component Migration (Medium Priority)
1. **Modal**
   - Create `/admin/src/components/ui-preline/Modal.tsx`
   - Maintain open/close state management
   - Support all sizes

2. **Dropdown**
   - Create `/admin/src/components/ui-preline/Dropdown.tsx`
   - Support items, dividers, icons
   - Test with React Router

3. **Skeleton**
   - Create `/admin/src/components/ui-preline/Skeleton.tsx`
   - Support text, circle, rect variants
   - Add animations

4. **Spinner**
   - Create `/admin/src/components/ui-preline/Spinner.tsx`
   - Support all sizes and colors
   - Add loading states

### Step 4: Component Migration (Low Priority)
1. **Terminal** (Custom)
   - Build from Preline Card component
   - Add window controls styling
   - Maintain copy functionality

### Step 5: Cleanup
1. Remove old `/admin/src/components/ui/` directory
2. Update all imports
3. Run full test suite
4. Optimize bundle size

---

## 6. Testing Strategy

### Unit Tests
```tsx
// Example: Button.test.tsx
import { render, screen } from '@testing-library/react';
import { Button } from './Button';

test('Button renders with primary variant', () => {
  render(<Button variant="primary">Click me</Button>);
  expect(screen.getByRole('button')).toHaveClass('bg-gradient-to-r');
});

test('Button applies theme colors', () => {
  render(<Button>Click me</Button>);
  const button = screen.getByRole('button');
  expect(button).toHaveStyle({
    '--accent-mid': '255 149 0'
  });
});
```

### Integration Tests
- Test components in actual pages
- Verify theme switching works
- Test all interactive states
- Check responsiveness

---

## 7. Advanced Preline Components

### Available for Future Use
Based on `more.txt`, these advanced components are available:

#### Form Components
- Datepicker
- TimePicker
- Advanced Select
- ComboBox
- File Upload
- Range Slider

#### Data Display
- DataTables
- Charts
- Tree View
- Timeline

#### Navigation
- Mega Menu
- Breadcrumb
- Pagination
- Stepper
- Tabs

#### Feedback
- Toasts
- Alerts
- Progress
- Ratings

#### Layout
- Sidebar
- Offcanvas (Drawer)
- Accordion
- Collapse

### Example: Adding DataTable
```tsx
// When needed, add advanced components
npm install datatables.net

// In component
useEffect(() => {
  if (window.DataTable && tableRef.current) {
    new window.DataTable(tableRef.current, {
      // options
    });
  }
}, []);
```

---

## 8. Troubleshooting

### Common Issues

1. **Preline components not initializing**
   - Ensure `window.HSStaticMethods.autoInit()` is called
   - Check if `preline/dist/index.js` is imported
   - Verify `main.tsx` has external libraries attached

2. **Theme colors not applying**
   - Check CSS variables are defined
   - Ensure overrides CSS is imported
   - Verify `dark:` prefixes are working

3. **TypeScript errors**
   - Ensure `global.d.ts` includes all Preline types
   - Check external libraries are properly typed

4. **Bundle size increase**
   - Only import components you need
   - Use dynamic imports for advanced features
   - Consider tree shaking with `preline/dist/*`

### Debug Commands
```bash
# Check Preline version
npm list preline

# Check TypeScript
npx tsc --noEmit

# Check build
npm run build

# Check bundle size
npm run build && npx vite-bundle-analyzer dist
```

---

## 9. Rollback Plan

If migration fails:
1. Keep `/admin/src/components/ui/` until migration is complete
2. Use feature flags to toggle between old/new components
3. Git branch strategy: `feature/preline-migration`
4. Quick rollback: Revert imports in components

---

## 10. Success Criteria

### Functional Requirements
- [ ] All current components have Preline equivalents
- [ ] Theme colors apply correctly in both light/dark modes
- [ ] All interactive states work (hover, focus, active, disabled)
- [ ] TypeScript types are properly defined
- [ ] Build completes without errors
- [ ] Bundle size increase is reasonable (< 500KB)

### Design Requirements
- [ ] Visual consistency with current design
- [ ] Smooth animations and transitions
- [ ] Proper spacing and typography
- [ ] Accessibility features maintained

### Performance Requirements
- [ ] No performance regression
- [ ] Components render efficiently
- [ ] Initial load time not significantly impacted

---

## 11. References

- [Preline UI Documentation](https://preline.co/docs/index.html)
- [Preline GitHub](https://github.com/htmlstreamofficial/preline)
- [Current Component Directory](/admin/src/components/ui/)
- [Selected Components](./selected-few.txt)
- [All Components List](./more.txt)
- [Current Theme CSS](/admin/src/styles/globals.css)

---

**Author**: Claude AI Assistant
**Review Date**: December 9, 2025
**Next Review**: After initial migration complete