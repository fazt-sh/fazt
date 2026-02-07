# Issue 05: Test Database Connection Hang

**Status**: RESOLVED
**Priority**: High
**Created**: 2026-02-07
**Resolved**: 2026-02-07
**Component**: cmd_gateway.go (production bug, exposed by tests)

---

## Summary

Full handler test suite hung on `TestCmdGateway_AcceptsValidAPIKey` due to a nested query deadlock in `cmdAppList`.

## Root Cause

`cmdAppList` in `cmd_gateway.go` iterated over an open `rows` cursor (from `sqlDB.Query`) and called `getAliasesForApp()` inside the loop, which ran another `db.Query()`. With `SetMaxOpenConns(1)` in tests, this deadlocked — the single connection was held by the outer cursor, and the inner query blocked forever waiting for it.

```
cmdAppList:
  rows := sqlDB.Query(apps)     ← holds connection
  for rows.Next() {
    getAliasesForApp(sqlDB, id)  ← tries to get another connection → DEADLOCK
  }
```

Production used `SetMaxOpenConns(10)` so it didn't deadlock, but the pattern was still a correctness issue (wasted connections, potential deadlock under load).

## Fix

Collect all rows into a slice first, close the cursor, then query aliases:

```go
// Collect rows first to release the DB connection
var appRows []appRow
for rows.Next() { ... }
rows.Close()

// Now query aliases (connection is free)
for _, r := range appRows {
    aliases := getAliasesForApp(sqlDB, r.id)
    ...
}
```

## Results

- Full handler suite: 64 tests, 2.6s (was: infinite hang)
- Full project suite: all packages pass
- Handler coverage: 14.2% (up from 7.3% before Plan 46)
