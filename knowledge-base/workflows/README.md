---
title: Fazt Workflows
description: Task-oriented guides for developing Fazt components
updated: 2026-01-31
category: index
tags: [workflows, guides, development]
---

# Fazt Workflows

Task-oriented guides for developing and extending Fazt. These workflows follow the **backend-first** principle: always verify API support before building UI.

## Quick Navigation

| Task | Guide | Level |
|------|-------|-------|
| Add Admin UI feature | [admin-ui/adding-features.md](admin-ui/adding-features.md) | ⭐⭐ |
| Understand UI architecture | [admin-ui/architecture.md](admin-ui/architecture.md) | ⭐ |
| Extend fazt-sdk | [fazt-sdk/extending.md](fazt-sdk/extending.md) | ⭐⭐ |
| Add backend API | [fazt-binary/adding-apis.md](fazt-binary/adding-apis.md) | ⭐⭐⭐ |
| Test with mock data | [admin-ui/testing.md](admin-ui/testing.md) | ⭐ |

## Workflow Categories

### 🎨 Admin UI
Location: `workflows/admin-ui/`

- **[architecture.md](admin-ui/architecture.md)** - State management, data flow, component model
- **[adding-features.md](admin-ui/adding-features.md)** - Step-by-step feature implementation
- **[testing.md](admin-ui/testing.md)** - Mock vs real mode, fixtures, validation
- **[checklist.md](admin-ui/checklist.md)** - Pre-implementation validation checklist

### 🔌 Fazt SDK
Location: `workflows/fazt-sdk/`

- **[api-reference.md](fazt-sdk/api-reference.md)** - Complete API endpoint reference
- **[extending.md](fazt-sdk/extending.md)** - Adding new SDK methods
- **[adapters.md](fazt-sdk/adapters.md)** - Mock adapter and custom adapters

### ⚙️ Fazt Binary
Location: `workflows/fazt-binary/`

- **[handlers.md](fazt-binary/handlers.md)** - API handler structure and patterns
- **[adding-apis.md](fazt-binary/adding-apis.md)** - Backend development workflow
- **[storage.md](fazt-binary/storage.md)** - Database and VFS patterns

## Core Principles

### 1. Backend-First Development

```mermaid
graph LR
    A[Backend API] --> B[SDK Method]
    B --> C[Data Store]
    C --> D[UI Component]
```

Always implement in this order. Never build UI for non-existent APIs.

### 2. Single Source of Truth

- **Backend**: PostgreSQL/SQLite database
- **API**: Go handlers in `internal/handlers/`
- **SDK**: JavaScript client in `admin/packages/fazt-sdk/`
- **State**: Reactive stores in `admin/src/stores/`
- **UI**: Components read from stores, never call API directly

### 3. Mock Mode Parity

Mock data must match real API structure exactly. Use mock mode for:
- ✅ Development without backend
- ✅ Testing UI logic
- ✅ Demonstrations

But always validate against real mode before deploying.

## When to Read What

| Situation | Read This |
|-----------|-----------|
| "I want to add a new page to Admin UI" | [admin-ui/adding-features.md](admin-ui/adding-features.md) |
| "How does state management work?" | [admin-ui/architecture.md](admin-ui/architecture.md) |
| "I need a new API endpoint" | [fazt-binary/adding-apis.md](fazt-binary/adding-apis.md) |
| "How do I extend the SDK?" | [fazt-sdk/extending.md](fazt-sdk/extending.md) |
| "What endpoints exist?" | [fazt-sdk/api-reference.md](fazt-sdk/api-reference.md) |
| "Mock data doesn't match real data" | [admin-ui/testing.md](admin-ui/testing.md) |

## Documentation Format

All workflow docs use frontmatter for metadata:

```yaml
---
title: Document Title
description: Brief description
version: 0.17.0
updated: 2026-01-31
category: workflows
tags: [relevant, tags]
---
```

This enables:
- Website generation (static site, docs portal)
- Search and categorization
- Last-updated timestamps
- Content freshness tracking

## Contributing

When adding workflows:
1. Use frontmatter template
2. Include practical examples
3. Update this README
4. Update the `updated` date
5. Test instructions before committing
