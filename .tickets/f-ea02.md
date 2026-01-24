---
id: f-ea02
status: closed
deps: []
links: []
created: 2026-01-24T17:03:03Z
type: task
priority: 2
assignee: Jikku Jose
---
# Fix analytics SQLITE_BUSY via WriteQueue

Route analytics batch writes through WriteQueue to eliminate SQLITE_BUSY errors at high concurrency. Expected: 2K user success rate 80% → 95%+

