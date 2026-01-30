# Plan 30: User Isolation & Analytics (Split)

**Status**: Split into sub-plans
**Created**: 2026-01-30

## Overview

This plan was split into three focused sub-plans for easier implementation:

| Plan | Focus | Target | Status |
|------|-------|--------|--------|
| [30a](30a_config_consolidation.md) | Config Consolidation | v0.14.0 | Ready |
| [30b](30b_user_data_foundation.md) | User Data Foundation | v0.15.0 | Ready |
| [30c](30c_access_control.md) | Access Control (RBAC + Domain Gating) | v0.16.0 | Ready |

## Dependency Order

```
30a: Config in DB (foundation)
 |
 v
30b: User isolation, IDs, analytics
 |
 v
30c: RBAC, domain gating
```

## Implementation Order

1. **30a first** - Restores single-DB philosophy, required for 30b/30c
2. **30b second** - User data structures needed for 30c
3. **30c last** - Builds on user structures from 30b

See individual plan files for details.
