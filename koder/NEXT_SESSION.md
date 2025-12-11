# Current Task

**Project**: Fazt.sh
**Location**: `/home/testman/workspace`
**Phase**: Transition from Phase 2C (UI Mocks) to Phase 3 (Integration)
**Last Session**: Dec 11, 2025

## 🛑 Context (Read First)
We paused the session after successfully completing two major parallel tracks:
1.  **Frontend**: Completed the complex `SiteDetail` page UI with 5 tabs (Overview, Env, Keys, Logs, Settings). It currently runs on `mockData`.
2.  **Backend**: Implemented and tested the `internal/analytics/buffer` to prevent DB write storms.

**Current State**: The repo is stable. Frontend is fully navigable (Mock Mode). Backend infrastructure is solid, but some API endpoints (DELETEs) defined in the spec are still missing.

## ✅ Completed Recently
- **UI**: `SiteDetail.tsx` is done. All modals, visibility toggles, and "Danger Zone" interactions work in mock mode.
- **Backend Tests**: Created `internal/analytics/buffer_test.go` (covered flush, shutdown, concurrent adds).
- **Backend Core**: Analytics buffering and System Observability APIs (`/api/system/*`) are in place.

## 📋 Next Session Goals

### 1. Verification (First 10 mins)
- Run backend tests: `go test -v ./internal/analytics`
- Run frontend: `cd admin && npm run dev`
- verify the `SiteDetail` page looks good and interactions (like adding an env var) work in the UI.

### 2. Decision Point: API vs Integration
We need to choose one path to proceed:

**Path A (Recommended): Finish Backend API**
- The `implementation-review` noted that `DELETE /api/redirects` and `DELETE /api/webhooks` are missing.
- Implement these to reach full `v0.7.2` feature parity.
- Standardize the remaining handlers to use the new `api.Envelope`.

**Path B: Start Frontend Integration**
- Switch frontend from `mockData` to `TanStack Query`.
- This will likely break the UI initially until the API is fully aligned.

### 3. Execution Plan (If Path A)
1.  Implement `DELETE` handlers in Go.
2.  Update `internal/api/response.go` if needed.
3.  Bump version to `v0.7.2` in `main.go`.

## References
- **UI Logic**: `admin/src/pages/sites/SiteDetail.tsx`
- **Backend Review**: `koder/analysis/06_implementation-review.md` (Detailed gap analysis)
- **API Spec**: `koder/docs/admin-api/spec.md`