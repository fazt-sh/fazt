# Plan 26: Refactor Test Harness to Integration Tests

## Problem

The test harness was incorrectly implemented as a `fazt harness` subcommand, embedding
test code into the production binary. This:
- Bloats binary size
- Ships test code to production
- Is not idiomatic Go

## Solution

Convert to proper Go integration tests using `_test.go` files with build tags.

## Target Structure

```
internal/harness/
├── config.go              # Shared types, Expected, Actual, Results
├── report.go              # Report generation (text/json/markdown)
├── testutil.go            # Test helpers, HTTP client setup
├── gaps/
│   ├── tracker.go         # Gap discovery (keep as library)
│   └── tracker_test.go    # Unit tests for tracker
├── baseline_test.go       # //go:build integration
├── requests_test.go       # //go:build integration
├── resilience_test.go     # //go:build integration
└── security_test.go       # //go:build integration
```

## Files to Remove

- `cmd/server/harness.go` - CLI handler
- `internal/harness/harness.go` - Orchestrator (merge into tests)
- `internal/harness/baseline/*.go` - Convert to `baseline_test.go`
- `internal/harness/requests/*.go` - Convert to `requests_test.go`
- `internal/harness/resilience/*.go` - Convert to `resilience_test.go`
- `internal/harness/security/*.go` - Convert to `security_test.go`

## Files to Keep (as library code)

- `config.go` - Types used by tests and reports
- `report.go` - Report generation
- `gaps/tracker.go` - Gap tracking

## Usage

```bash
# Unit tests only (default)
go test ./...

# Integration tests against running server
FAZT_TARGET=http://localhost:8080 go test -tags=integration ./internal/harness/...

# Specific test category
go test -tags=integration -run TestBaseline ./internal/harness/...
go test -tags=integration -run TestSecurity ./internal/harness/...

# With verbose output
go test -tags=integration -v ./internal/harness/...

# Generate report
go test -tags=integration -json ./internal/harness/... > results.json
```

## Test Structure Pattern

```go
//go:build integration

package harness

import (
    "os"
    "testing"
)

func getTarget(t *testing.T) string {
    target := os.Getenv("FAZT_TARGET")
    if target == "" {
        t.Skip("FAZT_TARGET not set")
    }
    return target
}

func TestBaseline_Throughput(t *testing.T) {
    target := getTarget(t)
    // ... test implementation
}

func TestBaseline_Latency(t *testing.T) {
    target := getTarget(t)
    // ... test implementation
}
```

## Changes to main.go

Remove from switch statement:
```go
case "harness":
    handleHarnessCommand(os.Args[2:])
```

## CI Integration

```yaml
# .github/workflows/integration.yml
- name: Start fazt server
  run: ./fazt server start --port 8080 &

- name: Run integration tests
  env:
    FAZT_TARGET: http://localhost:8080
  run: go test -tags=integration -v ./internal/harness/...
```

## Migration Steps

1. Create `internal/harness/testutil.go` with shared test helpers
2. Convert each subdirectory to single `*_test.go` file:
   - `baseline/*.go` → `baseline_test.go`
   - `requests/*.go` → `requests_test.go`
   - `resilience/*.go` → `resilience_test.go`
   - `security/*.go` → `security_test.go`
3. Remove `cmd/server/harness.go`
4. Remove harness case from `main.go`
5. Delete empty subdirectories
6. Update STATE.md

## Benefits

- **No binary bloat** - `_test.go` never compiled into release
- **Sync with code** - Tests live alongside implementation
- **Standard tooling** - Works with `go test`, CI, coverage
- **Selective execution** - Build tags separate unit/integration
- **Better reporting** - `go test -json` for structured output
