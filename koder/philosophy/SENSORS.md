# Sensors Philosophy

This document extends `CORE.md` with principles specific to the `fazt-sh/sensors` library.

Read `CORE.md` first. This document assumes familiarity with Fazt's foundational philosophy.

---

## Relationship to Core

The sensors library implements Fazt's peripheral layer—the boundary between Fazt and the physical world. It inherits all core principles:

- **Single Binary:** Sensors must not break the single-binary promise
- **Pure Go:** No CGO, ever (extended in detail below)
- **JSON Everywhere:** All readings are JSON-serializable
- **Events as Spine:** Sensors emit events, don't call Fazt directly
- **Schema Validation:** All reading types have schemas

This document adds sensor-specific principles that elaborate on how these apply to hardware abstraction.

---

## The Core Thesis

**Trust is the product. Purity is how we build it.**

Users of Fazt sensors should be able to assume:
- If a sensor type is supported, it works perfectly
- If it works on one platform, it works on all platforms
- If it worked yesterday, it works today
- There are no gotchas, no footnotes, no "but on macOS you need to..."

This trust takes years to build and moments to destroy.

---

## Principle 1: Pure Go, Absolute

### The Rule

No CGO. Ever. For any reason.

This principle exists in CORE.md, but sensors requires the most detailed application because hardware abstraction is where CGO temptation is strongest.

### Acceptable

```go
// Pure Go, standard library only
http.Get("http://camera.local/stream")

// Pure Go with syscalls - this is NOT CGO
unix.IoctlSetInt(fd, I2C_SLAVE, addr)

// Pure Go with platform build tags
// +build linux

// Pure Go third-party libraries (that themselves use no CGO)
import "periph.io/x/conn/v3/i2c"
```

### Not Acceptable

```go
// CGO for "just this one thing"
// #include <libusb.h>
import "C"

// CGO behind a build flag
// +build cgo

// CGO as "optional enhancement"
// "Works without CGO but better with it"
```

### Why the Line is Absolute

CGO is not a technical decision. It is architectural cancer.

| Without CGO | With CGO |
|-------------|----------|
| `GOOS=linux go build` | Cross-compiler toolchain |
| Copy binary, run | Install shared libraries |
| Reproducible builds | "Works on my machine" |
| Single static binary | Dependency hell |

One CGO dependency justifies the next. The line must be absolute because blurry lines move.

### Syscalls Are Not CGO

Go can call the operating system directly through syscalls. This is fundamentally different from CGO:

- **Syscall:** Go → Kernel (direct, no intermediary)
- **CGO:** Go → C runtime → C library → Kernel (dependency chain)

Syscalls are the operating system boundary. Using them maintains purity while enabling hardware access on platforms where Go can make those calls.

### When Something Requires CGO

Then we don't support it in sensors.

Document it clearly. Provide the external daemon protocol. The ugliness lives outside Fazt's boundary. The core remains pure.

---

## Principle 2: Perfection or Absence

### The Rule

Each supported sensor type is flawless, or it doesn't exist.

### The Checklist

Before a sensor type ships:

**State Machine:**
- Every lifecycle state is handled
- Every state transition is defined
- No undefined behavior, ever
- Hot-plug works perfectly
- Hot-unplug works perfectly
- Reconnection works perfectly

**Error Handling:**
- Every error is recoverable or clearly fatal
- No silent failures
- No data loss on transient errors
- Graceful degradation is explicit

**Data Quality:**
- Invalid readings are detected
- Spurious data is filtered
- Timestamps are accurate (Unix milliseconds)
- Gaps are detectable

**Cross-Platform:**
- Works identically on supported platforms
- No platform-specific gotchas
- Same binary, same behavior

### Why

Users don't read documentation for things that "just work." If sensors require documentation to use correctly, we've failed.

The goal: A user who has used one Fazt sensor can use any Fazt sensor without reading anything.

### "Good Enough" is Unacceptable

```
Acceptable:
  "Temperature sensor: Works on all platforms, handles disconnect/reconnect,
   validates readings, reports clearly when hardware fails."

Unacceptable:
  "Temperature sensor: Works on Linux. macOS support is experimental.
   May report stale readings after reconnect. Known issue #47."
```

If we can't do it perfectly, we don't do it yet.

---

## Principle 3: Learn One, Know All

### The Rule

All sensors follow identical patterns. No exceptions.

### What This Means

```go
// Temperature sensor
temp := sensors.New(sensors.Temperature, config)
temp.Start(ctx)
for reading := range temp.Readings() {
    // reading.Value, reading.Quality, reading.Timestamp
}

// Camera - identical pattern
cam := sensors.New(sensors.Camera, config)
cam.Start(ctx)
for reading := range cam.Readings() {
    // reading.Value, reading.Quality, reading.Timestamp
}

// GPS - identical pattern
gps := sensors.New(sensors.GPS, config)
gps.Start(ctx)
for reading := range gps.Readings() {
    // reading.Value, reading.Quality, reading.Timestamp
}
```

Same lifecycle. Same state machine. Same error model. Same JSON schema structure.

### Why

Consistency compounds:
- Learning cost is O(1), not O(n)
- One state machine implementation, thoroughly tested
- User code portable across sensor types
- Adding new sensors adds zero cognitive load

### No Special Cases

```
Unacceptable:
  "Cameras work differently because video is special"
  "GPS needs a different state machine because satellites"
  "Audio has unique buffering requirements"

Acceptable:
  "All sensors work identically. Type-specific data lives in the
   reading's Value field. Lifecycle, state machine, and error
   handling are universal."
```

If a sensor type can't fit the universal model:
1. Evolve the model (carefully, with full justification), or
2. Don't support that sensor type yet

---

## Principle 4: Patience is a Feature

### The Rule

Missing capabilities are better than compromised capabilities.

### The Timeline Comparison

**Compromised Feature:**
```
Day 1:    Ship with CGO / gotchas / platform issues
Day 30:   Users complain about build issues
Day 90:   Workarounds accumulate
Day 180:  "Known issues" section grows
Day 365:  Feature is legacy burden
Forever:  Technical debt compounds
```

**Patient Feature:**
```
Day 1:    Document as "not yet supported"
Day 30:   Some users use external daemon
Day 180:  Pure implementation exists (maybe)
Day 181:  Ship perfect implementation
Forever:  Trust maintained
```

The second timeline is longer. It is also better.

### Demand Creates Signal

"Not yet supported" makes demand visible. If many users need USB webcams, we see it. We can prioritize.

Shipping a compromised implementation hides demand. Users suffer quietly. We never know how much they wanted a good version.

---

## Principle 5: The External Daemon Boundary

### The Rule

For things that cannot be pure Go, define a clean protocol. Ugliness lives outside Fazt.

### The Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    OUTSIDE FAZT                             │
│                                                             │
│   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│   │ usb-camera  │  │ bluetooth   │  │ proprietary │        │
│   │   daemon    │  │   daemon    │  │   daemon    │        │
│   └──────┬──────┘  └──────┬──────┘  └──────┬──────┘        │
│          │                │                │                │
└──────────┼────────────────┼────────────────┼────────────────┘
           │                │                │
           │    Clean Protocol (JSON/HTTP)   │
           │                │                │
┌──────────┼────────────────┼────────────────┼────────────────┐
│          ▼                ▼                ▼                │
│   ┌─────────────────────────────────────────────────┐      │
│   │              sensors (Pure Go)                   │      │
│   │                                                  │      │
│   │  Receives clean JSON. No CGO. No compromise.    │      │
│   └─────────────────────────────────────────────────┘      │
│                                                             │
│                      INSIDE FAZT                            │
└─────────────────────────────────────────────────────────────┘
```

### Why This Is Not Compromise

The boundary is explicit and intentional:
- Fazt's purity is preserved
- The daemon is optional
- The protocol is well-defined (any language can implement it)
- Users who don't need those sensors never see complexity

We own the protocol specification. We provide reference implementations. We don't own the CGO code that talks to hardware.

---

## Sensor Support Status

### Supported (Pure Go)

| Sensor Type | Interface | Platform |
|-------------|-----------|----------|
| IP Camera | RTSP, HTTP | All |
| Network Audio | RTSP | All |
| Temperature | I2C | Linux (syscalls) |
| Humidity | I2C | Linux (syscalls) |
| Pressure | I2C | Linux (syscalls) |
| IMU | I2C | Linux (syscalls) |
| GPS | Serial | All |
| MQTT Sensor | MQTT | All |
| HTTP Endpoint | HTTP | All |

### Not Yet Supported

| Sensor Type | Blocker | Workaround |
|-------------|---------|------------|
| USB Webcam | Platform APIs need pure impl | External daemon |
| USB Microphone | Platform APIs need pure impl | External daemon |
| Bluetooth | BLE stack needs pure impl | External daemon |

### External Daemon Only

| Sensor Type | Reason |
|-------------|--------|
| Proprietary SDK hardware | Vendor lock-in |
| Platform-specific with no syscall path | Architecturally impossible |

---

## Go Libraries

Pure Go libraries for sensor protocols:

### IP Cameras (RTSP)

Uses [gortsplib](https://github.com/bluenviron/gortsplib) - THE standard for
RTSP in Go (~23k lines, MIT license):

```go
import "github.com/bluenviron/gortsplib/v5"

client := gortsplib.Client{}
client.Start("rtsp://camera.local:554/stream")

client.OnPacketRTP = func(medi *media.Media, pkt *rtp.Packet) {
    // Decode frame, emit as sensor reading
    fazt.events.Emit("sensor.camera.frame", frame)
}
```

Features: UDP/TCP/multicast, RTSPS/TLS, auto transport switching.
Binary impact: ~800KB (with mediacommon codec lib).

### MQTT Sensors

Uses [mochi-mqtt](https://github.com/mochi-mqtt/server) - embeddable MQTT v5
broker (~11k lines, MIT license):

```go
import mqtt "github.com/mochi-mqtt/server/v2"

server := mqtt.New(nil)
server.AddHook(new(FaztSQLiteHook), nil)  // Custom persistence
server.AddListener(listeners.NewTCP("tcp", ":1883", nil))

// Subscribe to sensor topics
server.Subscribe("sensors/+/+", func(cl *mqtt.Client, sub mqtt.Subscription, pk packets.Packet) {
    // Parse reading, emit event
    fazt.events.Emit("sensor.mqtt.reading", pk.Payload)
})
```

Heavy deps (badger, redis) are OPTIONAL hooks - core is stdlib + websocket.
Binary impact: ~500KB (core only).

### I2C Sensors (Temperature, Humidity, Pressure)

Uses [periph.io](https://periph.io) - referenced in Principle 1 above.
Syscalls for I2C, no CGO. See periph.io/x/devices for BME280, SHT4x, etc.

---

## Mock Generators

Every sensor type MUST have a mock implementation.

```go
// Mock enables testing without hardware
temp := sensors.NewMock(sensors.Temperature, scenarios.HeatingCycle)
temp.Start(ctx)

for reading := range temp.Readings() {
    // Scripted readings arrive at appropriate intervals
    // Test your code without physical sensors
}
```

Scenarios include:
- Normal operation
- Sensor failure
- Reconnection
- Spurious data
- Edge cases

This enables:
- Unit testing without hardware
- Integration testing in CI
- Development without physical sensors
- Demo mode

---

## Decision Process

### Adding a Sensor Type

1. **Can it be pure Go?** (syscalls acceptable)
   - No → Document in "Not Yet Supported," provide daemon protocol
   - Yes → Continue

2. **Can we make it perfect?**
   - No → Wait until we can
   - Yes → Continue

3. **Does it fit the universal model?**
   - No → Evolve model (with justification) or don't add
   - Yes → Implement

### Reviewing Changes

1. **CGO introduced?** → Reject
2. **Platform gotchas?** → Reject or fix
3. **Universal patterns followed?** → Required
4. **Perfect or "good enough"?** → Perfect required

---

## Amendments

This document inherits the amendment process from CORE.md.

Changes require:
1. Clear reasoning
2. Explicit acknowledgment of modified principle
3. Analysis of trust impact
4. Documentation as amendment, not silent edit

---

## The Vision

We're building the peripheral nervous system for personal AI infrastructure.

Sensors connect Fazt to the physical world. They must be:
- Perfectly reliable
- Universally consistent
- Transparently simple
- Trustworthy for decades

A compromised sensors library cannot serve as the foundation for autonomous personal infrastructure. Only a pure implementation can.

We choose purity.
