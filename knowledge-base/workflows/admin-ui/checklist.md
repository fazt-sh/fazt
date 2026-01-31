---
title: Admin UI Feature Checklist
description: Quick validation checklist before implementing UI features
updated: 2026-01-31
category: workflows
tags: [admin-ui, checklist, validation, backend-first]
---

# Admin UI Feature Checklist

Use this checklist **before** implementing any UI feature to ensure backend support exists.

## Pre-Implementation Checklist

### ✅ Backend Validation

- [ ] **API endpoint exists** in `internal/handlers/`
  ```bash
  grep -rn "endpoint_name" internal/handlers/
  ```

- [ ] **Response structure is documented** or code is readable
  ```bash
  # Check handler code for struct definitions
  ```

- [ ] **Error cases are handled** (404, 500, validation errors)

- [ ] **Authentication is required** (if applicable)

- [ ] **Rate limiting exists** (if applicable)

### ✅ SDK Validation

- [ ] **SDK method exists** in `admin/packages/fazt-sdk/index.js`
  ```javascript
  client.resource.method()
  ```

- [ ] **SDK method signature matches** backend expectations

- [ ] **Error handling is implemented** in SDK

### ✅ Mock Data Validation

- [ ] **Mock fixture exists** in `admin/packages/fazt-sdk/fixtures/`

- [ ] **Mock data structure matches** real API response exactly

- [ ] **Mock adapter route defined** in `mock.js`

### ✅ State Management

- [ ] **Store exists** for this data (or will be created)

- [ ] **Load function exists** (or will be created)

- [ ] **Update logic is defined** (set, remove, merge)

### ✅ UI Implementation

- [ ] **Component location determined** (page, modal, card)

- [ ] **Loading state will be shown**

- [ ] **Error messages will be displayed**

- [ ] **Success feedback will be provided**

- [ ] **Destructive actions have confirmation**

## Quick Validation Commands

### Check Backend Endpoint

```bash
# Search for endpoint
grep -rn "POST.*apps" internal/handlers/

# View handler
cat internal/handlers/apps_handler.go | grep -A 30 "HandleAppCreate"
```

### Check SDK Method

```bash
# View SDK
cat admin/packages/fazt-sdk/index.js | grep -A 5 "apps:"
```

### Check Mock Data

```bash
# List fixtures
ls admin/packages/fazt-sdk/fixtures/

# View fixture
cat admin/packages/fazt-sdk/fixtures/apps.json
```

### Test in Browser Console

```javascript
// Access admin internals
const { client, apps } = window.__fazt_admin

// Test SDK method
await client.apps.list()

// Check store
apps.get()
```

## Decision Tree

```
┌─────────────────────────────────────┐
│ Does backend endpoint exist?        │
└────────┬───────────────┬────────────┘
         │ YES           │ NO
         ↓               ↓
    Implement UI    ┌─────────────────┐
                    │ Choose:          │
                    │ 1. Backend first │
                    │ 2. Coming Soon   │
                    │ 3. Backlog       │
                    └─────────────────┘
```

## Red Flags 🚩

Stop and validate if you're about to:

- ❌ Add a field that doesn't exist in backend response
- ❌ Call an API endpoint that isn't documented
- ❌ Use mock data that differs from real structure
- ❌ Implement a feature without error handling
- ❌ Skip loading states
- ❌ Bypass the SDK with direct fetch calls

## Push-Back Template

When backend support is missing:

```
⚠️ Backend Validation Failed

Feature: [Feature Name]
Required: [Endpoint or API]
Status: Not found in internal/handlers/

Options:
1. Implement backend first (recommended)
   - Add handler in internal/handlers/
   - Test with real database
   - Then return to UI

2. Show "Coming Soon" (quick)
   - Professional placeholder
   - No broken functionality
   - Add to backlog

3. Defer to backlog
   - Document requirement
   - Prioritize later

Which approach do you prefer?
```

## Testing Checklist

### Before Deployment

- [ ] **Tested in mock mode** (`?mock=true`)
- [ ] **Tested in real mode** (without mock param)
- [ ] **Loading states work**
- [ ] **Error states work**
- [ ] **Success states work**
- [ ] **Edge cases handled** (empty, single item, many items)
- [ ] **Mobile responsive** (if applicable)
- [ ] **Keyboard accessible** (if applicable)

### After Deployment

- [ ] **Real API calls succeed**
- [ ] **Data displays correctly**
- [ ] **Interactions work**
- [ ] **No console errors**
- [ ] **Performance acceptable**

## Common Scenarios

### ✅ Scenario: Field Already Exists

```
User: "Show deployment date in app list"
Check: Does API return `deployed_at`?
→ YES: Just use it in UI
```

### ⚠️ Scenario: Endpoint Missing

```
User: "Add restart button"
Check: Does POST /api/apps/:id/restart exist?
→ NO: Push back, suggest backend first
```

### ✅ Scenario: Read-Only Display

```
User: "Show system stats on dashboard"
Check: Does GET /api/stats exist?
→ YES: Load and display
```

### ⚠️ Scenario: Complex Operation

```
User: "Add blue-green deployment UI"
Check: Does backend support this workflow?
→ UNCLEAR: Clarify backend capabilities first
```

## Quick Reference

| Task | Command |
|------|---------|
| Find endpoint | `grep -rn "endpoint" internal/handlers/` |
| Check SDK | `cat admin/packages/fazt-sdk/index.js` |
| View fixture | `cat admin/packages/fazt-sdk/fixtures/*.json` |
| Test API | Browser console: `await client.resource.method()` |
| Build & deploy | `cd admin && npm run build && fazt app deploy admin --to local --name admin-ui` |

