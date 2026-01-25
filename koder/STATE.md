# Fazt Implementation State

**Last Updated**: 2026-01-26
**Current Version**: v0.10.13

## Status

State: CLEAN - Storage API improvements + NEXUS mall dashboard fixes

---

## Last Session

**Storage API Performance & NEXUS Mall Dashboard**

### 1. Storage API Enhancements (`internal/storage/`)

Added efficient query operations to prevent memory issues:

**New `ds.find()` options:**
```javascript
// Before: loaded ALL documents
fazt.storage.ds.find('collection', {})

// After: supports limit, offset, order
fazt.storage.ds.find('collection', {}, { limit: 100, offset: 0, order: 'desc' })
```

**New `ds.count()` method:**
```javascript
// Efficient count without loading documents
var count = fazt.storage.ds.count('collection', { type: 'active' })
```

**New `ds.deleteOldest()` method:**
```javascript
// Single SQL query, no memory overhead
fazt.storage.ds.deleteOldest('collection', 1000)  // Keep newest 1000
```

### 2. NEXUS API Fix (`servers/zyt/nexus/api/main.js`)

Fixed memory explosion caused by cleanup logic:

**Before (problematic):**
```javascript
// On EVERY ingest (~40 calls/second across 4 feeds):
var all = fazt.storage.ds.find('feed_x', {});  // Loads ALL into memory
if (all.length > 1000) {
    for (var i...) ds.delete(...);  // Individual deletes
}
```

**After (efficient):**
```javascript
// 1% chance per ingest, single efficient SQL query
if (Math.random() < 0.01) {
    fazt.storage.ds.deleteOldest('feed_x', 1000);
}
```

### 3. Mall Demo Mode (`servers/zyt/nexus/src/stores/mall.js`)

Added demo data generation for mall layout when offline:
- `initDemoData()` - Initialize realistic mall metrics
- `updateDemoData()` - Periodic updates with zone occupancy, sales, alerts
- Changed `stats` from `reactive()` to `ref()` for proper Vue reactivity

### 4. LayoutManager Fix (`servers/zyt/nexus/src/core/LayoutManager.js`)

**Root cause of widgets not updating**: `loadLayout()` was stripping `dataMap`
and `dataSource` properties from widgets, so they couldn't bind to data.

Fixed by adding these properties to the widget mapping in `loadLayout()` and
`exportLayout()`.

### 5. Widget Props Reactivity (`servers/zyt/nexus/src/App.js`)

- Added `widgetPropsMap` computed that explicitly accesses reactive values
- This ensures Vue tracks mall stats as dependencies for widget re-renders

### Files Modified

```
internal/storage/ds.go              # FindOptions, FindWithOptions, DeleteOldest
internal/storage/bindings.go        # JS bindings for count, deleteOldest, find opts
internal/storage/storage_test.go    # Tests for new methods
servers/zyt/nexus/api/main.js       # Efficient cleanup + count()
servers/zyt/nexus/src/stores/mall.js     # Demo data + ref-based stats
servers/zyt/nexus/src/App.js             # widgetPropsMap computed
servers/zyt/nexus/src/core/LayoutManager.js  # Preserve dataMap/dataSource
```

## Previous Session

**NEXUS Multi-Layout Dashboard System**

Built a comprehensive dashboard system in `servers/zyt/nexus/` for visualizing
data with customizable widgets and layouts.

### 1. Base Map Widget

Created `MapWidget.js` using Leaflet for flight/ship tracking:
- Supports markers, paths, circles, polygons
- Directional markers (aircraft triangles, ship diamonds) with heading rotation
- Themes: dark, light, satellite, voyager
- Rich tooltips with telemetry data

### 2. Multi-Layout System

Created layout registry (`layouts/index.js`) with:
- **Flight Tracker** (🛰️): Map, flights, ships, activity log
- **Web Analytics** (📊): Visitors, page views, sessions, bounce rate, top pages
- **Shopping Mall** (🏬): Foot traffic, occupancy, sales, zone activity

Each layout has metadata (name, icon, description, theme) and widget positions.

### 3. Layout Switcher

Header dropdown to switch between layouts:
- Shows current layout icon + name
- Lists all layouts with descriptions
- Persists selection to localStorage

### 4. DataManager (API-driven architecture)

`DataManager.js` for connecting widgets to API endpoints:
```javascript
dataSources: {
    traffic: { endpoint: '/api/analytics/traffic', refresh: 5000 }
},
widgets: [{
    dataSource: 'traffic',
    dataMap: { value: 'visitorsToday' }
}]
```

### 5. WebSocket Hijack Fix

Fixed `responseWriter` wrapper missing `http.Hijacker` interface:
- Logging middleware was breaking WebSocket upgrades
- Added `Hijack()` passthrough in `cmd/server/main.go`
- Found while building NEXUS real-time features

## Files Created/Modified

```
servers/zyt/nexus/
├── src/widgets/MapWidget.js      # Leaflet map widget
├── src/layouts/index.js          # Layout registry
├── src/core/DataManager.js       # API data fetching
├── src/core/LayoutSwitcher.js    # Layout dropdown
├── src/App.js                    # Multi-layout support
└── index.html                    # z-index fixes, styles

cmd/server/main.go                # WebSocket Hijack() fix
```

## Next Up

1. **Continue NEXUS dashboard refinement**
   - Better UI/UX polish
   - More widgets (heatmap, line chart, table, etc.)
   - Mobile responsiveness
   - Real API integration for fazt analytics

2. **Connect to fazt analytics**
   - Build `/api/analytics/*` endpoints
   - Wire up DataManager to real data

3. **Widget library expansion**
   - Table widget for data grids
   - Line chart widget
   - Heatmap widget
   - Progress/status widgets

---

## Quick Reference

```bash
# Deploy nexus to local
fazt app deploy servers/zyt/nexus --to local

# Access dashboard
http://nexus.192.168.64.3.nip.io:8080

# Run tests for storage
go test -v ./internal/storage/...
```
