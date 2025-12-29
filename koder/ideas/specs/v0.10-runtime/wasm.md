# WASM Runtime (Internal Primitive)

## Summary

WebAssembly execution primitive for kernel services. Powered by wazero (pure Go,
no CGO). Not exposed to JS apps—services use this internally for performance-
critical operations.

## Why Internal

WASM as a user-facing feature would:
- Shift complexity to app developers (toolchains, memory management)
- Violate "batteries included" philosophy
- Create WASI compatibility burden

Instead: Kernel provides WASM primitive. Services use it invisibly. Apps get
simple APIs backed by fast WASM.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Apps (JS)                                                   │
│  fazt.services.media.resize()   ← simple API                │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  Services Layer (Go)                                         │
│  services/media/resize.go       ← calls kernel WASM         │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  Kernel WASM Primitive                                       │
│  kernel/runtime/wasm.go         ← this spec                 │
│  - Module cache                                              │
│  - Resource limits                                           │
│  - Host function bridge                                      │
└─────────────────────────────────────────────────────────────┘
```

## Go API (Internal)

### Module Loading

```go
// Load WASM module from embedded binary or filesystem
module, err := wasm.Load(ctx, wasmBytes, wasm.Config{
    Name:        "libimage",
    MemoryLimit: 64 * 1024 * 1024,  // 64MB
    FuelLimit:   1_000_000_000,      // ~1 second of compute
})
defer module.Close()
```

### Function Invocation

```go
// Call exported function
result, err := module.Call(ctx, "resize", wasm.Args{
    "width":   800,
    "height":  600,
    "quality": 85,
    "input":   imageBytes,
})

// result.Bytes() - output data
// result.Int() - integer return
// result.String() - string return
```

### Memory Transfer

```go
// Write data to WASM memory
ptr, err := module.WriteBytes(imageBytes)
defer module.Free(ptr)

// Call function with pointer
resultPtr, err := module.Call(ctx, "process", wasm.Args{
    "input_ptr": ptr,
    "input_len": len(imageBytes),
})

// Read result from WASM memory
output := module.ReadBytes(resultPtr.Int(), resultLen)
```

### Host Functions

```go
// Expose Go functions to WASM module
module.Export("log", func(msg string) {
    log.Info("wasm", "msg", msg)
})

module.Export("read_file", func(path string) []byte {
    // Kernel-mediated file access
    return vfs.Read(appID, path)
})
```

## Resource Limits

### Fuel Metering

```go
// Fuel = instruction count approximation
// ~1 billion fuel ≈ 1 second on typical hardware

module.SetFuelLimit(1_000_000_000)  // 1 second
module.SetFuelLimit(100_000_000)    // 100ms

// Check remaining
remaining := module.FuelRemaining()

// Refuel for long operations
module.AddFuel(500_000_000)
```

### Memory Limits

```go
// Hard cap on WASM linear memory
wasm.Config{
    MemoryLimit: 64 * 1024 * 1024,   // 64MB default
    MemoryLimit: 256 * 1024 * 1024,  // 256MB max
}

// Check usage
used := module.MemoryUsed()
```

### Timeouts

```go
// Context-based timeout
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
defer cancel()

result, err := module.Call(ctx, "process", args)
if errors.Is(err, context.DeadlineExceeded) {
    // Timed out
}
```

## Module Cache

```go
// Modules are compiled once, cached
cache := wasm.NewCache(wasm.CacheConfig{
    MaxModules:  10,
    MaxMemory:   512 * 1024 * 1024,  // 512MB total
    TTL:         1 * time.Hour,
})

// Load uses cache automatically
module, err := cache.Load(ctx, "libimage", wasmBytes)
```

## Embedded Modules

Shipped with Fazt binary:

```go
//go:embed wasm/libimage.wasm
var libImageWasm []byte

//go:embed wasm/libpdf.wasm
var libPdfWasm []byte

//go:embed wasm/libxlsx.wasm
var libXlsxWasm []byte
```

## Safety

### No WASI by Default

WASM modules have no:
- Filesystem access
- Network access
- Environment variables
- System calls

All external access goes through explicit host functions controlled by kernel.

### Sandboxing

```go
// Each invocation is isolated
// - Fresh memory
// - No shared state between calls
// - Deterministic execution

// Multiple concurrent calls are safe
go module.Call(ctx, "resize", args1)
go module.Call(ctx, "resize", args2)
```

### Panic Recovery

```go
// WASM panics don't crash kernel
result, err := module.Call(ctx, "process", args)
if err != nil {
    if wasm.IsPanic(err) {
        log.Error("wasm panic", "module", "libimage", "err", err)
        // Module is still usable
    }
}
```

## Use Cases (Services)

### Image Processing (services/media)

```go
func (s *MediaService) Resize(path string, opts ResizeOpts) ([]byte, error) {
    input, err := s.vfs.Read(path)
    if err != nil {
        return nil, err
    }

    module, err := s.wasm.Load(ctx, "libimage")
    if err != nil {
        return nil, err
    }

    return module.Call(ctx, "resize", wasm.Args{
        "input":   input,
        "width":   opts.Width,
        "height":  opts.Height,
        "quality": opts.Quality,
        "format":  opts.Format,
    })
}
```

### PDF Generation (services/pdf)

```go
func (s *PdfService) FromHTML(html string, opts PdfOpts) ([]byte, error) {
    module, err := s.wasm.Load(ctx, "libpdf")
    if err != nil {
        return nil, err
    }

    return module.Call(ctx, "render", wasm.Args{
        "html":      html,
        "page_size": opts.PageSize,
        "margin":    opts.Margin,
    })
}
```

### Excel Parsing (services/parse)

```go
func (s *ParseService) XLSX(data []byte) ([][]string, error) {
    module, err := s.wasm.Load(ctx, "libxlsx")
    if err != nil {
        return nil, err
    }

    result, err := module.Call(ctx, "parse", wasm.Args{
        "input": data,
    })

    return result.JSON().([][]string), err
}
```

## CLI (Admin Only)

```bash
# List loaded modules
fazt wasm list

# Module stats
fazt wasm stats libimage

# Clear cache
fazt wasm cache clear

# Test module (debug)
fazt wasm test libimage resize --input test.png --width 100
```

## Metrics

```go
// Exposed via kernel metrics
wasm_module_loads_total{module="libimage"}
wasm_call_duration_seconds{module="libimage", function="resize"}
wasm_fuel_consumed_total{module="libimage"}
wasm_memory_bytes{module="libimage"}
wasm_errors_total{module="libimage", type="timeout|panic|oom"}
```

## Implementation Notes

- Uses wazero v2+ (pure Go, zero CGO)
- Modules compiled to native code on first load
- Cache stores compiled modules, not WASM bytes
- Host functions use wazero's `api.GoModuleFunc`
- Memory transfer uses wazero's linear memory API

## Not Exposed

This primitive is NOT available to JS apps:
- No `fazt.wasm.*` namespace
- No way to load custom WASM from JS
- Services are the only consumers

Future kernel extensions may allow owner-installed WASM modules via CLI.
