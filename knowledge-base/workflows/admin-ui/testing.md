---
title: Testing Admin UI
description: Mock vs real mode testing strategies and validation
updated: 2026-01-31
category: workflows
tags: [admin-ui, testing, mock-mode, validation]
---

# Testing Admin UI

The Admin UI supports both **mock mode** (development) and **real mode** (production). This guide explains how to test in both modes and ensure consistency.

## Mock vs Real Mode

### Mock Mode (`?mock=true`)

Uses fixtures for all API calls:

```
http://admin-ui.192.168.64.3.nip.io:8080?mock=true
```

**Purpose:**
- ✅ Develop UI without backend
- ✅ Test UI logic and interactions
- ✅ Demo and screenshots
- ✅ Rapid iteration

**Limitations:**
- ❌ Doesn't validate real API
- ❌ May drift from reality
- ❌ No real error scenarios

### Real Mode (default)

Uses actual backend API:

```
http://admin-ui.192.168.64.3.nip.io:8080
```

**Purpose:**
- ✅ Validate API integration
- ✅ Test with real data
- ✅ Catch API mismatches
- ✅ Production readiness

## Fixture Management

### Location

`admin/packages/fazt-sdk/fixtures/`

```
fixtures/
├── apps.json      # Sample apps
├── aliases.json   # Sample aliases
├── health.json    # System health
└── user.json      # Mock user
```

### Structure Rules

**✅ Match real API exactly:**

```javascript
// Real API response
{
  "success": true,
  "data": {
    "id": "app_01hw3xyz",
    "name": "momentum",
    "file_count": 24,
    "size_bytes": 1258291
  }
}

// Mock fixture (apps.json)
{
  "id": "app_01hw3xyz",
  "name": "momentum",
  "file_count": 24,
  "size_bytes": 1258291
}
```

**❌ Don't add mock-only fields:**

```javascript
// Bad: field doesn't exist in real API
{
  "id": "app_01hw3xyz",
  "name": "momentum",
  "deployment_count": 5  // ← Not in real API
}
```

### Updating Fixtures

When backend API changes:

1. **Capture real response:**
   ```bash
   curl http://localhost:8080/api/apps | jq
   ```

2. **Update fixture:**
   ```bash
   cat admin/packages/fazt-sdk/fixtures/apps.json
   # Copy real response structure
   ```

3. **Test mock mode:**
   ```bash
   open http://admin-ui.192.168.64.3.nip.io:8080?mock=true
   ```

## Testing Workflow

### 1. Development (Mock Mode)

```bash
# Start dev server
cd admin && npm run dev

# Open in mock mode
open http://localhost:5173?mock=true
```

**What to test:**
- UI renders correctly
- Interactions work
- Loading states show
- Error states show
- Success states show
- Modals open/close
- Navigation works
- Filtering works

### 2. Validation (Real Mode)

```bash
# Build and deploy
cd admin && npm run build
fazt app deploy admin --to local --name admin-ui

# Open in real mode
open http://admin-ui.192.168.64.3.nip.io:8080
```

**What to test:**
- API calls succeed
- Real data displays
- Mutations work (create, update, delete)
- Errors are handled
- Loading states accurate
- Auth flow works
- Permissions enforced

### 3. Comparison

Test the same flow in both modes:

| Action | Mock Mode | Real Mode |
|--------|-----------|-----------|
| Load apps list | ✅ Shows fixtures | ✅ Shows real apps |
| Click app detail | ✅ Shows mock data | ✅ Shows real data |
| Delete app | ✅ Simulated | ✅ Actually deletes |
| Create alias | ✅ Simulated | ✅ Actually creates |
| Error case | ✅ Mock error | ✅ Real error |

## Browser Console Testing

### Access Admin Internals

```javascript
// Available in browser console
const { router, client, apps, aliases, currentApp } = window.__fazt_admin

// Check mode
const useMock = new URLSearchParams(window.location.search).get('mock') === 'true'
console.log('Mode:', useMock ? 'MOCK' : 'REAL')
```

### Test SDK Methods

```javascript
// List apps
const appsList = await client.apps.list()
console.log('Apps:', appsList)

// Get specific app
const app = await client.apps.get('app_01hw3xyz')
console.log('App:', app)

// Error handling
try {
  await client.apps.get('nonexistent')
} catch (err) {
  console.error('Error:', err)
}
```

### Inspect State

```javascript
// Read current state
console.log('Apps store:', apps.get())
console.log('Current app:', currentApp.get())
console.log('Auth:', auth.get())

// Watch changes
apps.subscribe(data => console.log('Apps updated:', data))
```

### Trigger Actions

```javascript
// Navigate
router.push('/apps')
router.push('/apps/app_01hw3xyz')

// Load data
await loadApps(client)
await loadAppDetail(client, 'app_01hw3xyz')
```

## Common Testing Scenarios

### Scenario 1: New List Page

**Mock mode:**
1. Add fixture data
2. Verify list renders
3. Test filtering
4. Test empty state
5. Test loading state

**Real mode:**
1. Verify API call succeeds
2. Verify real data displays
3. Test with empty database
4. Test with large dataset
5. Test pagination (if applicable)

### Scenario 2: Detail Page

**Mock mode:**
1. Add detailed fixture
2. Verify all fields display
3. Test navigation
4. Test actions (mock)

**Real mode:**
1. Navigate to real entity
2. Verify all data loads
3. Test actions (real mutations)
4. Verify breadcrumbs update

### Scenario 3: Create/Update Form

**Mock mode:**
1. Fill form
2. Submit
3. Verify mock success
4. Check state updated

**Real mode:**
1. Fill form
2. Submit
3. Verify API call
4. Check entity created/updated
5. Verify list updates

### Scenario 4: Delete Operation

**Mock mode:**
1. Click delete
2. Confirm
3. Verify mock removal
4. Check state updated

**Real mode:**
1. Click delete
2. Confirm
3. Verify API call
4. Check entity removed
5. Verify list updates
6. Verify cascade (if applicable)

## Error Testing

### Simulated Errors (Mock Mode)

```javascript
// In mock.js
'DELETE /api/apps/:id': (params) => {
  if (params.id === 'protected') {
    throw { code: 'PROTECTED', message: 'Cannot delete protected app', status: 403 }
  }
  return { message: 'Deleted' }
}
```

### Real Errors (Real Mode)

Test actual error scenarios:
- Network failure (disconnect network)
- 404 Not Found (invalid ID)
- 403 Forbidden (wrong permissions)
- 500 Server Error (backend bug)
- Timeout (slow network)

## Validation Checklist

Before considering a feature complete:

- [ ] **Works in mock mode**
- [ ] **Works in real mode**
- [ ] **Loading states show**
- [ ] **Error messages clear**
- [ ] **Success feedback given**
- [ ] **Edge cases handled** (empty, single, many)
- [ ] **Permissions respected**
- [ ] **Navigation works**
- [ ] **State updates correctly**
- [ ] **No console errors**

## Debugging Tips

### Issue: UI doesn't update after mutation

**Check:**
1. Did store update? `apps.get()`
2. Is component subscribed? `apps.subscribe(...)`
3. Is cleanup returning unsubscribe?

### Issue: API call fails in real mode

**Check:**
1. Is endpoint correct? `client.apps.method`
2. Does backend handler exist? `grep -rn handler internal/`
3. Is auth working? Check cookies/session
4. Check network tab in DevTools

### Issue: Mock and real differ

**Check:**
1. Compare fixture to real API response
2. Update fixture structure
3. Check SDK response parsing

### Issue: State gets stale

**Check:**
1. Are you calling load functions?
2. Are subscriptions set up?
3. Is component re-rendering?

## Performance Testing

### Load Testing

```javascript
// Generate large dataset in mock
const manyApps = Array.from({ length: 100 }, (_, i) => ({
  id: `app_${i}`,
  name: `app-${i}`,
  file_count: 10,
  size_bytes: 1000000
}))
```

Test with:
- 10 items (normal)
- 100 items (busy instance)
- 1000 items (stress test)

### Network Testing

Simulate slow network:
- Chrome DevTools → Network → Throttling
- Test loading states
- Test timeout handling

## Integration Testing

### Manual Test Suite

**Dashboard:**
- [ ] Loads without errors
- [ ] Stats display correctly
- [ ] Apps table shows recent apps
- [ ] Activity feed shows events

**Apps List:**
- [ ] Loads all apps
- [ ] Filter works
- [ ] Cards/list toggle works
- [ ] Click navigates to detail

**App Detail:**
- [ ] Loads app data
- [ ] Shows aliases
- [ ] Shows files
- [ ] Actions work (delete, refresh)

**Aliases:**
- [ ] Lists aliases
- [ ] Shows types (proxy, redirect, reserved)
- [ ] Links to apps work

## Continuous Validation

### On Every Deployment

```bash
# 1. Build
cd admin && npm run build

# 2. Deploy
fazt app deploy admin --to local --name admin-ui

# 3. Test real mode
open http://admin-ui.192.168.64.3.nip.io:8080

# 4. Smoke test
# - Load dashboard
# - Navigate to apps
# - Open app detail
# - Check for errors
```

### Before Release

- [ ] All features work in real mode
- [ ] No console errors
- [ ] Permissions enforced
- [ ] Mobile responsive
- [ ] Performance acceptable
- [ ] Error handling graceful
