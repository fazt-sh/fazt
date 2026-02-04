# Plan: Replace Old Admin with Vue Admin-UI

**Date:** 2026-02-04
**Status:** Draft
**Priority:** High
**Complexity:** Medium

## Goal

Replace the old embedded React admin with the new Vue-based admin-ui, serving it directly on the `admin.*` subdomain with no redirects or dual systems.

## Current State

### What We Have

1. **Old Embedded React Admin**
   - Location: Unknown (need to find where it's stored)
   - Served at: `admin.*` subdomain (currently redirects to admin-ui)
   - Had its own auth system (password login)
   - Status: Deprecated, confusing, needs removal

2. **New Vue Admin-UI** ✅
   - Location: `admin/` (tracked in git)
   - Currently served at: `admin-ui.*` subdomain
   - Uses unified auth (database-backed OAuth/dev login)
   - Status: Working, modern, maintainable
   - Features: Dashboard, Apps, Aliases, Events, Logs, Health, Settings

### Current Routing (in `cmd/server/main.go`)

```go
// admin.* subdomain - currently redirects to admin-ui.*
dashboardMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    scheme := "https"
    if !cfg.IsProduction() && !cfg.HTTPS.Enabled {
        scheme = "http"
    }
    portSuffix := ""
    if cfg.Server.Port != "80" && cfg.Server.Port != "443" {
        portSuffix = ":" + cfg.Server.Port
    }
    redirectURL := fmt.Sprintf("%s://admin-ui.%s%s/", scheme, cfg.Server.Domain, portSuffix)
    http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
})
```

## Target State

### What We Want

```
admin.192.168.64.3.nip.io:8080  →  New Vue Admin-UI (direct, no redirect)
admin-ui.*                       →  REMOVED (no longer exists)
```

### Reserved Subdomain Handling

`admin` is a reserved subdomain that currently can't be deployed to via `fazt app deploy`. We need to:
1. Either: Special-case `admin` to serve the admin-ui app
2. Or: Allow deploying to reserved subdomains with special flag/permission

## Implementation Plan

### Phase 1: Discovery (15 min)

**Goal:** Find and understand what exists

- [ ] **Find old embedded admin files**
  ```bash
  find . -path "*/assets/system/admin*" -o -path "*/internal/admin*"
  grep -r "embedded.*admin" internal/
  ```

- [ ] **Find old admin routes**
  ```bash
  grep -n "dashboardMux" cmd/server/main.go | grep -i admin
  ```

- [ ] **Check for old admin-specific handlers**
  ```bash
  grep -r "AdminHandler\|admin.*Handler" internal/handlers/
  ```

- [ ] **Document what old admin had that new one doesn't**
  - Check endpoints, features, UI capabilities

### Phase 2: Prepare Admin-UI for Reserved Subdomain (30 min)

**Goal:** Make admin-ui deployable to `admin` subdomain

**Option A: Special-case in hosting system** (Recommended)

Update `internal/hosting/apps.go` or wherever app deployment is handled:

```go
// Allow deploying to 'admin' reserved subdomain with special handling
if appName == "admin" {
    // Special deployment logic for admin UI
    // Mark as system app
    // Store in apps table with special flag
}
```

**Option B: Update reserved subdomain routing**

Modify `cmd/server/main.go` to serve admin-ui files directly on `admin.*`:

```go
// admin.* subdomain - serve admin-ui
dashboardMux.Handle("/", http.FileServer(getAdminUIFS()))
```

Where `getAdminUIFS()` returns the admin-ui's VFS or file system.

**Decision needed:** Which approach? I recommend **Option A** for consistency.

### Phase 3: Remove Old Admin (30 min)

**Goal:** Clean removal of all old admin code

- [ ] **Remove old admin files**
  ```bash
  rm -rf internal/assets/system/admin/  # or wherever it is
  ```

- [ ] **Remove old admin routes from main.go**
  - Remove redirect handler on dashboardMux
  - Remove any old admin-specific routes

- [ ] **Remove old admin handlers**
  - Check `internal/handlers/` for admin-specific code that's no longer needed
  - Keep the unified auth handlers (we use those)

- [ ] **Update any imports/references**
  ```bash
  grep -r "assets/system/admin" .
  grep -r "embedded.*admin" .
  ```

- [ ] **Check for old admin middleware**
  - Remove if it's not used by new admin

### Phase 4: Deploy New Admin to `admin.*` (15 min)

**Goal:** Serve Vue admin-ui on main admin subdomain

**Step 1: Deploy admin-ui to `admin` subdomain**

```bash
# This currently fails with "reserved subdomain" error
# After Phase 2 changes, this should work:
fazt @local app deploy /home/kodeman/Projects/fazt/admin --name admin
```

**Step 2: Update admin-ui configuration (if needed)**

Check if admin-ui has any hardcoded references to `admin-ui.*`:
```bash
grep -r "admin-ui" admin/src/
```

If found, update to use `admin.*` or make it dynamic.

**Step 3: Test deployment**
```bash
curl -I http://admin.192.168.64.3.nip.io:8080/
# Should return 200, serve admin-ui
```

### Phase 5: Remove `admin-ui.*` Subdomain (10 min)

**Goal:** Complete cleanup

- [ ] **Delete admin-ui app from database**
  ```bash
  fazt @local app delete admin-ui
  ```

- [ ] **Verify admin-ui.* returns 404**
  ```bash
  curl -I http://admin-ui.192.168.64.3.nip.io:8080/
  # Should return 404 or redirect to admin.*
  ```

- [ ] **Update any documentation**
  - README, knowledge-base references to admin-ui
  - Change to just `admin.*`

### Phase 6: Testing & Verification (20 min)

**Goal:** Ensure everything works

- [ ] **Test authentication flow**
  ```bash
  agent-browser open http://admin.192.168.64.3.nip.io:8080/
  # Should redirect to dev login, then back to admin dashboard
  ```

- [ ] **Test all admin features**
  - [ ] Dashboard loads with metrics
  - [ ] Apps list shows all apps (with ?all=true)
  - [ ] Aliases list works
  - [ ] Events page works
  - [ ] Logs page works
  - [ ] Health page works
  - [ ] Settings page works
  - [ ] Logout works

- [ ] **Test role-based access**
  - [ ] Login as owner: should see everything
  - [ ] Login as admin: should see everything
  - [ ] Login as user: should get /unauthorized.html

- [ ] **Check server logs for errors**
  ```bash
  journalctl --user -u fazt-local -n 100 | grep -i error
  ```

## Files to Modify

Based on discovery, likely:

1. **cmd/server/main.go**
   - Remove redirect handler for admin.*
   - Add admin app serving logic or route to VFS

2. **internal/hosting/apps.go** (or similar)
   - Allow deployment to `admin` reserved subdomain
   - Special handling for system apps

3. **admin/src/** (if needed)
   - Update any hardcoded `admin-ui.*` references

4. **Remove files:**
   - Old embedded admin directory (TBD based on discovery)
   - Old admin-specific handlers (if any)

## Rollback Plan

If something goes wrong:

1. **Keep a backup of old admin files before deleting**
   ```bash
   cp -r internal/assets/system/admin /tmp/old-admin-backup
   ```

2. **Git branch for this work**
   ```bash
   git checkout -b replace-old-admin
   # Make all changes on this branch
   # Easy to revert if needed
   ```

3. **Re-enable redirect if new admin fails**
   - Revert main.go changes
   - Keep old admin accessible

## Post-Implementation

### Documentation Updates

- [ ] Update `CLAUDE.md` - Remove references to old admin
- [ ] Update `knowledge-base/agent-context/architecture.md`
- [ ] Update any admin-related docs to point to `admin.*`

### Future Enhancements

Track in separate issues:
- [ ] Add missing features from old admin (if any found)
- [ ] Improve admin-ui UX based on usage
- [ ] Add more role-based permissions
- [ ] Progressive enhancement towards full API parity

## Risk Assessment

**Low Risk:**
- New admin-ui is already working and tested
- Auth system is unified and stable
- Changes are mostly routing and cleanup

**Medium Risk:**
- Discovering old admin had features we need
- Reserved subdomain handling complexity

**Mitigation:**
- Document old admin features before removal
- Test thoroughly before cleanup
- Keep backup of old admin files
- Use git branch for easy rollback

## Estimated Time

- Phase 1 (Discovery): 15 min
- Phase 2 (Prepare): 30 min
- Phase 3 (Remove old): 30 min
- Phase 4 (Deploy new): 15 min
- Phase 5 (Cleanup): 10 min
- Phase 6 (Testing): 20 min

**Total: ~2 hours**

## Open Questions

1. **Where exactly is the old embedded admin stored?** (Discovery phase will answer)
2. **What features did old admin have that new one doesn't?** (Need to check)
3. **Are there any external references to `admin-ui.*` that need updating?** (Check docs, scripts)
4. **Should we keep a "legacy admin" flag for gradual migration?** (Suggest: no, clean break)

## Success Criteria

✅ `admin.*` serves new Vue admin-ui directly
✅ No redirects or dual systems
✅ `admin-ui.*` removed completely
✅ All admin features work
✅ Authentication works for all roles
✅ No old admin code remains
✅ Documentation updated
✅ Server logs clean (no errors)

---

**Ready to implement?** Review this plan, then we execute phase by phase.
