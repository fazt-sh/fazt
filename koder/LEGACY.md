# Legacy Code Tracking

Items marked for removal. Search: `grep -rn "LEGACY_CODE" internal/`

## How to Remove

1. Pick an item from the list below
2. Remove the code (search for the marker)
3. Run tests: `go test ./...`
4. Tests prefixed `TestLegacy_` should fail - delete or update them
5. Update this file

## Active Legacy Items

### 1. Old App ID Format (`app_*`)

**Location**: `internal/appid/appid.go`
**Marker**: `LEGACY_CODE: Old format constants`
**Tests**: `TestLegacy_OldAppIDFormat`
**Remove when**: All apps use `fazt_app_*` format

Code to remove:
- `legacyAppPrefix`, `legacyAlphabet` constants
- `isValidLegacy()` function
- `GenerateLegacyApp()` function
- Legacy format acceptance in `IsValid()`

### 2. `fazt.storage.*` Namespace

**Location**: `internal/storage/bindings.go` (entire file)
**Marker**: `LEGACY_CODE: fazt.storage.* namespace`
**Tests**: `TestLegacy_StorageNamespace`
**Remove when**: All apps use `fazt.app.*`

Code to remove:
- Entire `bindings.go` file
- `storageInjector` in `internal/runtime/handler.go`

### 3. `generateUUID()` Function

**Location**: `internal/auth/service.go:73`
**Marker**: `LEGACY_CODE: generateUUID`
**Tests**: None (unused)
**Remove when**: Confirmed no usage

Code to remove:
- `generateUUID()` function

---

## Removed Items

_(Move items here when removed)_

## Commands

```bash
# Find all legacy markers
grep -rn "LEGACY_CODE" internal/

# Run only legacy tests
go test ./... -run "TestLegacy_"

# Count legacy items
grep -c "LEGACY_CODE" internal/**/*.go
```
