# Preline UI Design System for Fazt Admin

## Overview

This directory contains the complete design-system-preline/ implementation with wrapper components that use Preline UI styling while maintaining the Fazt brand identity.

## What's Built

### ✅ Components Created

1. **Button** (`src/components/ui-preline/Button.tsx`)
   - Variants: primary, secondary, ghost, danger
   - Sizes: sm, md, lg
   - Loading and disabled states
   - Gradient effects for primary buttons

2. **Input** (`src/components/ui-preline/Input.tsx`)
   - Label, helper text, and error states
   - Icon support (left/right positions)
   - Focus states with accent colors
   - Form validation styling

3. **Card** (`src/components/ui-preline/Card.tsx`)
   - Variants: default, bordered, elevated, glass
   - Hover effects
   - Header, Body, Footer sub-components
   - Glassmorphism support

4. **Badge** (`src/components/ui-preline/Badge.tsx`)
   - Variants: default, success, error, warning, info
   - Solid and soft styles
   - Size options
   - Dot indicators

5. **Modal** (`src/components/ui-preline/Modal.tsx`)
   - Multiple sizes: sm, md, lg, xl, full
   - Backdrop click handling
   - Keyboard (ESC) support
   - Header and footer support

6. **Dropdown** (`src/components/ui-preline/Dropdown.tsx`)
   - Placement options
   - Item icons
   - Disabled states
   - Dividers

7. **Skeleton** (`src/components/ui-preline/Skeleton.tsx`)
   - Text lines
   - Shapes: circle, rect, card
   - Animated pulses
   - Custom widths

8. **Spinner** (`src/components/ui-preline/Spinner.tsx`)
   - Multiple sizes
   - Color variants
   - Animation states

### ✅ Theme System

1. **CSS Variables** (`src/styles/globals.css`)
   - Docs-inspired color palette
   - Dark/light mode support
   - Accent colors: red-orange, orange, yellow

2. **Theme Overrides** (`src/styles/preline-overrides.css`)
   - Glassmorphism effects
   - Hover animations
   - Focus rings
   - Custom animations

3. **Tailwind Configuration**
   - Custom animations
   - Content sources for Preline
   - Forms plugin integration

### ✅ Demo Page

**DesignSystemPreline** (`src/pages/DesignSystemPreline.tsx`)
- Complete showcase of all components
- Interactive demos
- Usage examples
- Live preview

## Usage

### Import Components

```tsx
import { Button, Input, Card } from '../components/ui-preline';
```

### Example Usage

```tsx
<Button variant="primary" onClick={() => console.log('clicked')}>
  Click me
</Button>

<Input
  label="Email"
  type="email"
  icon={<Mail className="h-4 w-4" />}
  helperText="We'll never share your email"
/>
```

## Features Maintained

- ✅ Fazt brand colors
- ✅ Dark mode support
- ✅ TypeScript types
- ✅ Accessibility (ARIA attributes)
- ✅ Hover animations
- ✅ Focus states
- ✅ Loading states
- ✅ Error states

## Migration Strategy

The wrapper components maintain the same API as the original components, making migration simple:

1. Change import path: `ui/` → `ui-preline/`
2. Component names remain the same
3. Props are compatible
4. Theme colors are preserved

## Next Steps

1. Test all components in development
2. Verify theme switching works
3. Check accessibility compliance
4. Performance testing
5. Gradual migration of existing pages

## Build Status

- ✅ Types: All components typed
- ✅ Build: Compiles successfully
- ✅ Demo: Interactive page ready
- ⏳ Preline JS: Not yet integrated (waiting for plugin)

---

**Status**: Ready for review and testing