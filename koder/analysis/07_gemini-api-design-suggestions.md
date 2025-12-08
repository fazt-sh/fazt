# Gemini API Design Review & Standardization Proposals

**Date**: December 8, 2025
**Status**: Proposal / Draft
**Based on**: `koder/docs/admin-api/request-response.md` (Ground Truth Analysis)

## 1. Executive Summary

A recent "Ground Truth" probe of the Fazt Admin API revealed significant inconsistencies in response formatting and URL patterns. Currently, the API uses at least three different response structures and two different deletion patterns.

This fragmentation increases the cognitive load for client developers (CLI, Dashboard) and makes the codebase harder to maintain and test. This document proposes a **Strict Standardization Strategy** to unify the API surface.

## 2. Current State Analysis

### 2.1 The "Envelope" Chaos
The API currently returns data in three conflicting formats:

| Format Type | Structure | Used By | Issues |
|:---|:---|:---|:---|
| **Raw JSON** | `[...]` or `{...}` | `GET /api/events`<br>`GET /api/stats`<br>`GET /api/redirects` | No metadata support (pagination), no success/error distinction in body. |
| **Data Wrapper** | `{"data": ...}` | `GET /api/system/*`<br>`GET /api/sites` | Missing explicit `success` bool, inconsistent with other formats. |
| **Custom Keys** | `{"success": true, "key": ...}` | `GET /api/deployments` (`deployments`)<br>`GET /api/keys` (`keys`)<br>`GET /api/logs` (`logs`) | Client must know the specific key for every endpoint. |

### 2.2 The Deletion Inconsistency
Resources are deleted using two different patterns, often arbitrarily:

| Pattern | Example | Used By | Issues |
|:---|:---|:---|:---|
| **Query Param** | `DELETE /api/sites?site_id=...` | Sites, EnvVars, Keys | Not RESTful. Harder to protect with standard WAF/RBAC rules. |
| **Path Param** | `DELETE /api/redirects/:id` | Redirects, Webhooks | Standard REST pattern. Preferred. |

### 2.3 Error Handling Variability
*   Some endpoints return `http.Error` (text/plain).
*   Some return `{"error": "message"}`.
*   Some return `{"data": null, "error": {"code": "...", "message": "..."}}`.

## 3. Standardization Proposal

I propose adopting a **Single Standard Envelope** for all API interactions. This mimics modern best practices (e.g., Stripe, Slack, JSend).

### 3.1 The Standard Response Object
Every API response (200-299) MUST return:

```json
{
  "success": true,
  "data": <Object|Array>,
  "meta": <Object|null>   // Optional: Pagination, counts, etc.
}
```

*   **`success`**: Always `true` for 2xx responses.
*   **`data`**: The actual payload. If it's a list, it's an array. If it's a single resource, it's an object.
*   **`meta`**: Side-channel data (e.g., `{"total": 100, "limit": 50, "offset": 0}`).

**Example: List Sites (Before vs After)**
*   *Before*: `{"data": [...]}`
*   *After*:
    ```json
    {
      "success": true,
      "data": [{"name": "blog", ...}],
      "meta": {"count": 1}
    }
    ```

**Example: Get Stats (Before vs After)**
*   *Before*: `{"total_events": 100, ...}` (Raw)
*   *After*:
    ```json
    {
      "success": true,
      "data": {"total_events": 100, ...}
    }
    ```

### 3.2 The Standard Error Object
Every API Error (4xx-5xx) MUST return:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",     // Stable, machine-readable (e.g., VALIDATION_FAILED)
    "message": "Human text",  // Human-readable details
    "details": {...}          // Optional: Field-level validation errors
  }
}
```

### 3.3 RESTful Resource URLs
All CRUD routes MUST follow standard resource patterns:

| Action | Pattern | Example |
|:---|:---|:---|
| List | `GET /api/:resource` | `GET /api/sites` |
| Create | `POST /api/:resource` | `POST /api/sites` |
| Get | `GET /api/:resource/:id` | `GET /api/sites/blog` |
| Update | `PUT/PATCH /api/:resource/:id` | `PUT /api/sites/blog` |
| Delete | `DELETE /api/:resource/:id` | `DELETE /api/sites/blog` |

**Required Changes:**
*   Refactor `DELETE /api/sites?site_id=X` -> `DELETE /api/sites/X`
*   Refactor `DELETE /api/envvars?id=X` -> `DELETE /api/envvars/X`
*   Refactor `DELETE /api/keys?id=X` -> `DELETE /api/keys/X`

## 4. Implementation Strategy

To avoid breaking the CLI immediately, we can implement this in phases, or if we accept a breaking change (since this is v0.x), we do it in one go.

**Recommendation: "The Great Refactor" (Breaking Change)**
Since we are in `v0.x` and the user base is small (internal), we should rip off the bandage.

1.  **Create `internal/api/response.go`**: A strict helper package.
    *   `api.Success(w, data)`
    *   `api.SuccessList(w, data, meta)`
    *   `api.Error(w, code, err)`
2.  **Refactor Handlers**: Go through `internal/handlers/*.go` one by one and replace ad-hoc JSON encoding with these helpers.
3.  **Update Router**: Change `http.NewServeMux` paths to support the new DELETE patterns (Go 1.22+ supports `DELETE /api/sites/{id}`).
4.  **Update Spec**: Update `spec.md` to reflect the new reality.
5.  **Verify**: Re-run the `probe_api.sh` logic (updated) to verify consistency.

## 5. Benefits

1.  **Frontend Simplicity**: The Dashboard JS can use a single `fetchWrapper` that always expects `.data`.
2.  **CLI Consistency**: The Go CLI client can use a single struct for unmarshalling responses.
3.  **Observability**: Standardized error codes make logs easier to parse.
