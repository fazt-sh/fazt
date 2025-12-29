# Timekeeper (Local Time Consensus)

## Summary

Local time consensus without NTP. When internet is unavailable, Fazt nodes
agree on approximate time by polling each other. Provides "good enough"
time for cron, TTLs, and signatures when external time sources are gone.

## Why Kernel-Level

Time is invisible infrastructure until it breaks:
- Cron relies on accurate time
- TTLs and expiry depend on timestamps
- Signatures include timestamps for validity
- When NTP is gone, clocks drift minutes/hours per day

Timekeeper provides a fallback: consensus time from local peers.

## The Resilience Contract

```
NTP available:
  Kernel uses system time (NTP-synced) → Timekeeper tracks but defers

NTP unavailable, peers available:
  Kernel uses Timekeeper consensus → cron/TTL/signatures work

NTP unavailable, no peers:
  Kernel uses system time (drifting) → best effort
```

**Apps don't change.** `fazt.schedule()` just works in more conditions.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    KERNEL TIME                              │
│                         │                                   │
│                    time.Now()                               │
│                         │                                   │
│            ┌────────────┴────────────┐                      │
│            ▼                         ▼                      │
│     System Time (NTP)          Timekeeper                   │
│            │                    (consensus)                 │
│            │                         │                      │
│            └────────────┬────────────┘                      │
│                         ▼                                   │
│                  Best available time                        │
└─────────────────────────────────────────────────────────────┘
```

## Consensus Algorithm

Simple averaging with outlier rejection:

```
1. Poll all Beacon-discovered peers for their time
2. Collect responses: [T1, T2, T3, T4, T5]
3. Remove outliers (> 2 std deviations from median)
4. Average remaining: consensus_time
5. Track drift: local_clock - consensus_time
```

**Accuracy target**: ±60 seconds (enough for cron, TTLs, coordination)

Not trying to replace NTP precision. Just preventing multi-hour drift.

## Time Response

```javascript
{
  localTime: 1705312200000,      // Node's system clock
  consensusTime: 1705312215000,  // Node's current consensus (if any)
  confidence: 0.95,              // How confident in consensus
  sources: 3,                    // How many peers contributed
  ntpSynced: false               // Is NTP working?
}
```

## Kernel Integration

Timekeeper wraps all internal time calls:

```go
// In kernel/time.go (pseudo-code)
func Now() time.Time {
    if timekeeper.HasConsensus() && !ntp.IsAvailable() {
        return timekeeper.ConsensusTime()
    }
    return time.Now()  // System time
}
```

All kernel subsystems use `kernel.Now()` instead of `time.Now()`:
- Cron scheduler
- TTL expiry checks
- Event timestamps
- Signature timestamps

**Single integration point.** Existing code paths unchanged.

## Sync Process

```
Every 60 seconds (when peers available):

1. Query Beacon for peers
2. For each peer, request time (parallel, 2s timeout)
3. Collect responses
4. Calculate consensus
5. Update local drift estimate

If no peers for 10 minutes:
  - Confidence degrades
  - Eventually falls back to system time only
```

## Explicit Usage (Optional)

Apps can query time status if needed:

```javascript
// Usually not needed - kernel handles this
const status = await fazt.time.status();
// {
//   local: 1705312200000,
//   consensus: 1705312215000,
//   drift: -15000,          // Local is 15s behind consensus
//   confidence: 0.95,
//   sources: 3,
//   ntpAvailable: false
// }

// Get consensus time directly
const now = fazt.time.now();  // Returns consensus if available

// Check drift
const drift = fazt.time.drift();  // milliseconds
if (Math.abs(drift) > 60000) {
  console.warn('Clock drift exceeds 1 minute');
}
```

## CLI

```bash
# Check time status
fazt time status
# Local:     2024-01-15 14:03:22
# Consensus: 2024-01-15 14:03:07
# Drift:     +15s (local ahead)
# Sources:   3 peers
# NTP:       unavailable

# Force sync now
fazt time sync
# Polled 3 peers, consensus updated

# See contributing peers
fazt time peers
# neighbor-a: 14:03:05 (included)
# neighbor-b: 14:03:09 (included)
# neighbor-c: 14:15:00 (outlier, excluded)
```

## JS API (Explicit, Optional)

```javascript
fazt.time.now()
// Returns: consensus time (ms since epoch) or system time if no consensus

fazt.time.status()
// Returns: { local, consensus, drift, confidence, sources, ntpAvailable }

fazt.time.drift()
// Returns: milliseconds (positive = local ahead, negative = local behind)

fazt.time.peers()
// Returns: [{ id, name, time, included }]

fazt.time.sync()
// Force immediate sync, returns updated status
```

## Storage

Timekeeper persists drift estimates for faster convergence on restart:

```sql
CREATE TABLE timekeeper_state (
    key TEXT PRIMARY KEY,
    value TEXT,              -- JSON
    updated_at INTEGER
);

-- Stores:
-- 'drift_estimate': last known drift
-- 'peer_offsets': per-peer time offsets
-- 'confidence': last confidence level
```

## Configuration

```bash
# Timekeeper is enabled by default
fazt config set timekeeper.enabled true

# Sync interval (default 60s)
fazt config set timekeeper.interval 60

# Outlier threshold (default 2.0 std deviations)
fazt config set timekeeper.outlierThreshold 2.0

# Minimum peers for consensus (default 1)
fazt config set timekeeper.minPeers 1
```

## Implementation Notes

- ~150 lines of Go
- Uses Beacon for peer discovery
- Parallel peer polling with timeouts
- Exponential backoff if peers unavailable
- Graceful degradation to system time

## Limits

| Limit | Default |
|-------|---------|
| `syncInterval` | 60s |
| `peerTimeout` | 2s |
| `maxPeers` | 10 (for consensus) |
| `outlierThreshold` | 2.0 std dev |
| `minConfidence` | 0.5 (below this, use system time) |
| `maxDrift` | 1 hour (beyond this, warn user) |

## Accuracy Expectations

| Condition | Expected Accuracy |
|-----------|-------------------|
| NTP available | < 100ms (system handles) |
| 3+ peers, no NTP | ± 30 seconds |
| 1-2 peers, no NTP | ± 60 seconds |
| No peers, no NTP | Degrading (system clock drift) |

This is intentionally coarse. Goal is preventing hours of drift, not millisecond precision.

## Security Considerations

- Malicious peer could report wrong time
- Outlier detection mitigates single bad actor
- If >50% peers are malicious, consensus is compromised
- For high-security: disable Timekeeper, require NTP
- Signatures should tolerate ±5 minute timestamp variance
