# Fazt Implementation State

**Last Updated**: 2026-01-25
**Current Version**: v0.10.13

## Status

State: CLEAN - NEXUS dashboard system with multi-layout support

---

## Last Session

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

### 5. Shell Prompt Enhancement

Updated `~/dotfiles/prompts/lino.zsh` to show repo name in prompt:
```
(fazt/master: *)  →  (zyt/main: +)
```

### 6. zyt Apps Version Control

Initialized git repo in `servers/zyt/`:
- 228 files, 16 apps committed
- `config.json` (with token) properly gitignored
- Ready for remote: `gh repo create zyt-apps --private`

### 7. WebSocket Hijack Fix

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

servers/zyt/.gitignore            # New - protects config.json
~/dotfiles/prompts/lino.zsh       # Repo name in prompt
cmd/server/main.go                # WebSocket Hijack() fix
```

## Next Up

1. **Refine NEXUS dashboard**
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

# zyt apps repo
cd servers/zyt
git status
gh repo create zyt-apps --private --source=. --push

# Reload shell prompt
source ~/dotfiles/prompts/lino.zsh
```
