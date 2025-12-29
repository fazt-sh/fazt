# v0.10 - Runtime

**Theme**: Intelligent serverless execution with code splitting and scheduling.

## Summary

v0.10 transforms the JavaScript runtime from a simple request handler into a
capable application server. Apps can split code across files, use standard
libraries without npm, and schedule background tasks.

## Goals

1. **Code Splitting**: `require()` for local imports
2. **Standard Library**: Bundled utilities (lodash, cheerio, etc.)
3. **Scheduling**: JS-Cron for background tasks
4. **Zero-Build**: No npm, no bundlers, no build step

## Key Changes

| Capability | Description |
|------------|-------------|
| `api/` folder | Dedicated serverless code location |
| `require()` | Local file imports |
| Stdlib | Pre-bundled libraries in binary |
| JS-Cron | Scheduled function execution |
| Hibernate | Zero RAM when idle |
| WASM primitive | Internal wazero runtime for services |

## Documents

- `serverless.md` - The `api/` folder convention and request handling
- `stdlib.md` - Embedded standard library
- `cron.md` - Scheduled execution and hibernate pattern
- `sandbox.md` - Safe code execution for agents and dynamic evaluation
- `wasm.md` - Internal WASM primitive for services (not exposed to JS)

## Dependencies

- v0.9 (Storage): `fazt.storage` for persistent state

## Risks

- **Security**: `require()` must be sandboxed to app files only
- **Memory**: Multiple concurrent VMs can exhaust RAM
- **Stdlib Maintenance**: Bundled libraries need security updates
