# Issue 05: Test Database Connection Hang

**Status**: Open
**Priority**: High
**Created**: 2026-02-07
**Component**: Testing Infrastructure
**Affects**: Plan 46 Phase 1.4

---

## Summary

Full handler test suite hangs after ~28 seconds on `TestCmdGateway_AcceptsValidAPIKey`. The test times out waiting to acquire a database connection in `getAliasesForApp()`.

## Symptoms

```bash
$ go test ./internal/handlers -count=1 -timeout=30s
# ... tests run ...
panic: test timed out after 30s
	running tests:
		TestCmdGateway_AcceptsValidAPIKey (28s)
```

**Stack trace shows**:
- Goroutine stuck in `database/sql.(*DB).conn()` waiting for connection
- Called from `getAliasesForApp()` at `apps_handler_v2.go:796`
- Multiple database connection opener goroutines waiting

## Observations

**Tests that PASS** (fail auth early, don't query DB):
- ✅ `TestCmdGateway_RejectsUnauthenticated` (0.02s)
- ✅ `TestCmdGateway_RejectsInvalidToken` (0.01s)

**Tests that HANG** (pass auth, execute commands):
- ❌ `TestCmdGateway_AcceptsValidAPIKey` (timeout after 28s)

**Call chain to hang point**:
```
CmdGatewayHandler
  → ValidateAPIKey (SELECT + UPDATE api_keys)
  → executeCommand
    → executeAppCommand
      → cmdAppList
        → getAliasesForApp  ← HANGS HERE trying to db.Query()
```

## Test Setup

Each test calls `setupCmdTestDB(t)`:
```go
func setupCmdTestDB(t *testing.T) {
    db := setupTestDB(t)  // Creates :memory: DB, SetMaxOpenConns(1)
    // ... add schema, insert test data ...
    database.SetDB(db)  // Set global DB
    t.Cleanup(func() {
        db.Close()
        database.SetDB(nil)
    })
}
```

## Hypothesis

**Database connection not returned to pool**:
- `SetMaxOpenConns(1)` means only one connection allowed
- If `ValidateAPIKey()` or any handler code holds a connection open (e.g., unclosed transaction, unclosed rows), the next `db.Query()` call will block forever
- Note: `ValidateAPIKey()` was recently fixed to close rows before UPDATE (commit 4836058), but issue persists

**Possible causes**:
1. Transaction started but not committed/rolled back
2. Rows cursor not closed (despite `defer rows.Close()`)
3. Global `database.SetDB()` race condition between tests
4. In-memory SQLite :memory: DB with `SetMaxOpenConns(1)` has subtle locking behavior

## Environment

- Go 1.25.1
- SQLite driver: `modernc.org/sqlite`
- In-memory databases (`:memory:`)
- Single connection pool (`SetMaxOpenConns(1)`)

## Related

- **Fixed in commit 4836058**: API key validation lock (rows not closed before UPDATE)
- **Plan 46 Phase 1.4**: Critical handler tests (includes CMD gateway tests)

## Impact

- ❌ Full handler test suite cannot complete
- ❌ Blocks Plan 46 Phase 1.4 completion
- ❌ CI/CD cannot verify handler test coverage
- ✅ Individual tests pass when run in isolation
- ✅ Middleware, routing, schema tests all pass

## Next Steps (for Opus)

1. **Investigate connection lifecycle**:
   - Add debug logging to track `db.Query()` / `rows.Close()` calls
   - Check if any code path skips `defer rows.Close()`
   - Look for unclosed transactions

2. **Check ValidateAPIKey flow**:
   - Verify rows are truly closed before UPDATE runs
   - Check if UPDATE holds locks that block SELECT in getAliasesForApp

3. **Test isolation**:
   - Verify `database.SetDB()` / `database.GetDB()` thread safety
   - Check if `t.Cleanup()` order causes race conditions

4. **Potential fixes**:
   - Switch to multiple connections (`SetMaxOpenConns(10)`) for tests
   - Use separate DB per test (don't use global `database.SetDB()`)
   - Add explicit connection/transaction cleanup in test helpers
   - Check for SQLite PRAGMA settings that might cause locking

5. **Workaround**:
   - Skip problematic tests temporarily to unblock Plan 46
   - File separate issue to fix properly

## Reproduction

```bash
# Full suite hangs:
go test ./internal/handlers -count=1 -timeout=30s

# Specific test hangs:
go test ./internal/handlers -run TestCmdGateway_AcceptsValidAPIKey -timeout=10s

# But these pass:
go test ./internal/handlers -run "TestCmdGateway_Rejects" -count=1 -v
```

---

**Assignment**: Opus
**Estimated effort**: 1-2 hours investigation + fix
