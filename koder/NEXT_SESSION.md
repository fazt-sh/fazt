# Next Session Handoff: API Standardization & Admin Console

**Date**: December 8, 2025
**Status**: 🟡 **API INCONSISTENCY**: Server is stable, but API surface is fragmented.

## 📂 Context Payload (Read These First)
When starting this session, read these files to build your working memory:
1.  **Analysis**: `koder/analysis/07_gemini-api-design-suggestions.md` (The Blueprint).
2.  **Reality**: `koder/docs/admin-api/request-response.md` (The Problem).
3.  **Code**: `internal/handlers/` (The Target).

## 🛑 Immediate Action Required
1.  **Standardize API**: The current API returns mixed response formats.
    *   **Goal**: Unify all endpoints to use the standard envelope defined in the Analysis doc.
    *   **Action**: Create `internal/api/response.go` helpers and refactor `internal/handlers/*.go`.

2.  **Update Spec**:
    *   Update `koder/docs/admin-api/spec.md`.
    *   Regenerate `request-response.md` using `probe_api.sh` (ensure it exists or recreate it) to verify consistency.

3.  **Build Admin Console**:
    *   **Goal**: Build a SPA dashboard in `internal/assets/system/admin/`.
    *   **Tech**: Vanilla JS or Alpine.js (if lightweight), embedded.

## ⚠️ Known Issues
*   **API Fragmentation**: Clients need to handle 3 different response shapes.
*   **Deletion Verbs**: Mixed usage of Query params vs Path params for DELETE.

## ✅ Accomplished (Previous Session)
1.  **Ground Truth Analysis**: Captured actual server responses in `request-response.md`.
2.  **API Design Review**: Proposed standardization in `07_gemini-api-design-suggestions.md`.
3.  **Documentation Updates**: Updated `GEMINI.md` and `start.md`.

## 📋 Next Steps (The Plan)

### 1. API Standardization (The Great Refactor)
*   Implement `internal/api/response.go` helpers.
*   Refactor all handlers to use `api.Success`, `api.SuccessList`, `api.Error`.
*   Standardize DELETE routes to `DELETE /resource/{id}`.

### 2. Admin Console Implementation
*   Develop the SPA in `internal/assets/system/admin/`.
*   Connect to the new standardized API.
