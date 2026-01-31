# Rename internal "sites" to "apps"

**Status**: Backlog
**Priority**: Low (cosmetic consistency)

## Context

The external nomenclature is "apps" but internal code still uses "sites":

- `internal/handlers/sites_handler.go` → should be `apps_handler.go`
- `/api/sites` endpoint → should be `/api/apps`
- Database tables reference "sites"
- Various variable names

## Why

- Consistency between CLI, API, UI, and internal code
- Reduces cognitive load when navigating codebase
- SDK uses "apps" - internal should match

## Scope

```bash
# Files to rename
grep -rn "sites" internal/ --include="*.go" | grep -v "test" | head -20

# Estimated changes
# - Handler files
# - Route definitions
# - Database queries (careful with migrations)
# - Type/struct names
```

## Approach

1. Add new `/api/apps` endpoints alongside existing
2. Mark `/api/sites` as deprecated (LEGACY_CODE)
3. Update internal naming incrementally
4. Remove old endpoints after SDK/admin migrate

## Notes

- Don't break existing deployments
- Coordinate with SDK/admin development
