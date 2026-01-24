# Fazt Capacity Guide

Reference for understanding fazt's performance characteristics and limits.
Benchmarked on a $6 VPS (1 vCPU, 1GB RAM, SSD).

## TL;DR

 | Metric             | Capacity   | Notes                |
 | --------           | ---------- | -------              |
 | Concurrent users   | 2,000+     | Tested, 100% success |
 | Read throughput    | ~20,000/s  | Static files, VFS    |
 | Write throughput   | ~800/s     | SQLite single-writer |
 | Mixed (30% writes) | ~2,300/s   | Typical app workload |
 | RAM under load     | ~60MB      | Stable, no growth    |

## Monthly Capacity Estimates

Assuming 30% write ratio (typical for interactive apps):

 | Tier   | Requests/month   | Use Case                 |
 | ------ | ---------------- | ----------               |
 | Light  | 10M              | Personal blog, portfolio |
 | Medium | 50M              | Small SaaS, dashboard    |
 | Heavy  | 200M             | Active community app     |

**Write budget**: ~2M writes/month sustained (forms, storage, analytics)

## Architecture Constraints

### SQLite Single-Writer

Fazt uses SQLite with a single-writer model via WriteQueue. This means:

- All writes serialize through one goroutine
- No write contention or SQLITE_BUSY errors
- Predictable ~800 writes/sec ceiling
- Reads are unlimited (concurrent, no locking)

**This is intentional.** Predictability > raw throughput for personal infra.

### When You Hit Limits

Signs you're approaching capacity:

1. Write queue depth consistently > 500
2. Response times > 100ms for writes
3. RAM usage climbing (memory leak, not capacity)

**Solutions:**
- Add another fazt instance (horizontal scaling)
- Each instance = separate SQLite = separate write budget
- Route by app or user segment

## Workload Profiles

### Static Site (Blog, Docs, Portfolio)
- 99% reads, 1% writes (analytics only)
- Effective capacity: **50M+ pageviews/month**
- Bottleneck: Network bandwidth, not fazt

### Interactive App (Dashboard, Forms)
- 70% reads, 30% writes
- Effective capacity: **~6M requests/month**
- Bottleneck: Write throughput

### Write-Heavy (Chat, Logging, Analytics)
- 50% reads, 50% writes
- Effective capacity: **~4M requests/month**
- Bottleneck: SQLite serialization

### API Backend (CRUD operations)
- Mixed, depends on endpoints
- Effective capacity: **2-10M requests/month**
- Consider: Batch writes where possible

## Optimization Tips

### Reduce Write Pressure

1. **Batch operations** - Insert multiple docs in one call
2. **Debounce client writes** - Don't save on every keystroke
3. **Use KV for hot data** - Faster than doc store for simple values
4. **Analytics are async** - They don't block requests

### Maximize Read Performance

1. **Static files are fast** - VFS serves from SQLite efficiently
2. **Cache at client** - Set appropriate Cache-Control headers
3. **Minimize JS payload** - Faster initial load

## Benchmark Commands

Run these against local server to verify capacity:

```bash
# Pure reads (static files)
go run /tmp/loadtest.go -users 2000 -duration 20

# Pure writes (document inserts)
go run /tmp/writetest.go -users 500 -duration 20

# Mixed workload (30% writes)
go run /tmp/mixedtest.go -users 1000 -writes 30 -duration 20

# Check write queue health
curl -H "Host: admin.DOMAIN" \
  -H "Authorization: Bearer $TOKEN" \
  http://HOST/api/system/capacity
```

## Real-Time Capabilities (WebSocket)

Fazt v0.17+ includes native WebSocket support with pub/sub channels. This
enables collaborative features, presence, and live updates.

### Connection Limits ($6 VPS)

 | Metric                | Capacity   | Notes              |
 | --------              | ---------- | -------            |
 | WebSocket connections | 5,000      | Safe baseline      |
 | With tuning           | 10,000     | Memory optimized   |
 | Memory per connection | ~50KB      | Buffers + state    |
 | Broadcasts/sec        | 10,000+    | CPU-bound, not I/O |

### Critical Insight: Broadcast vs Persist

Real-time features separate two concerns:

1. **Broadcast** (in-memory): Cursors, typing, presence → unlimited*
2. **Persist** (SQLite): Messages, documents, state → 800/s limit

Most real-time data is **ephemeral** and never hits disk:

 | Data Type          | Broadcast        | Persist               |
 | -----------        | -----------      | ---------             |
 | Cursor position    | ✓ High frequency | ✗ Never               |
 | Typing indicator   | ✓ While typing   | ✗ Never               |
 | Presence status    | ✓ On change      | ✓ Occasional snapshot |
 | Chat message       | ✓ Once           | ✓ Once                |
 | Document operation | ✓ Immediately    | ✓ Batched             |

---

## Real-Time Scenario Models

### Collaborative Document (CRDT/Yjs-like)

Google Docs-style editor with real-time sync.

**10 concurrent editors:**

 | Activity           | Rate        | Type                      |
 | ----------         | ------      | ------                    |
 | Keystrokes         | 50/s total  | Broadcast                 |
 | Cursor moves       | 100/s total | Broadcast only            |
 | Selection changes  | 20/s total  | Broadcast only            |
 | CRDT operations    | 50/s        | Broadcast + batch persist |
 | **Writes to disk** | 5-10/s      | Batched every 1-2 sec     |

**Capacity:** 50+ concurrent documents, 500+ editors total

**100 editors on one doc:**
- Broadcasts: ~1,000/s (CPU-bound, achievable)
- Writes: ~50/s (trivial)
- Connections: 100 (trivial)

✅ **Verdict: Very achievable**

---

### Presence System (Who's Online)

Slack-style online indicators.

**100 users in workspace:**

 | Activity           | Rate         | Type                |
 | ----------         | ------       | ------              |
 | Heartbeats         | 3/s (100÷30) | Broadcast           |
 | Status changes     | 0.1/s        | Broadcast + persist |
 | Initial sync       | On connect   | Read                |
 | **Writes to disk** | ~1/s         | Status snapshots    |

**Capacity:** 2,000+ users across workspaces

✅ **Verdict: Trivial**

---

### Chat with Typing Indicators

Real-time chat room with "user is typing..." display.

**50-user chat room:**

 | Activity           | Rate   | Type                |
 | ----------         | ------ | ------              |
 | Messages sent      | 1-5/s  | Broadcast + persist |
 | Typing indicators  | 20/s   | Broadcast only      |
 | Read receipts      | 10/s   | Broadcast + batch   |
 | **Writes to disk** | 2-5/s  | Messages + receipts |

**Capacity:** 100+ active rooms, 1,000+ users

✅ **Verdict: Easy**

---

### Cursor Sharing (Figma/Penpot-lite)

Real-time collaborative canvas with visible cursors.

**10 collaborators on canvas:**

 | Activity           | Rate      | Type                                  |
 | ----------         | ------    | ------                                |
 | Cursor positions   | 150-300/s | Broadcast only (10 users × 15-30 fps) |
 | Shape operations   | 10-30/s   | Broadcast + persist                   |
 | Selection changes  | 10/s      | Broadcast only                        |
 | **Writes to disk** | 10-30/s   | Shape ops only                        |

**This is the most demanding scenario.**

**Bottleneck analysis:**
- 300 broadcasts/sec = ~3ms CPU per user per frame
- JSON serialization dominates
- Binary protocol would 3x capacity

**Capacity estimates:**

 | Collaborators/doc   | Docs concurrent   | Total users   |
 | ------------------- | ----------------- | ------------- |
 | 5                   | 50                | 250           |
 | 10                  | 25                | 250           |
 | 20                  | 10                | 200           |

⚠️ **Verdict: Achievable with limits (5-20 users per document)**

---

### AI Agent Monitoring (Shopify Plugin)

Real-time behavioral analysis with AI interventions.

**Per user being monitored:**

 | Activity           | Rate            | Type           |
 | ----------         | ------          | ------         |
 | Page views         | 0.1/s           | Event → AI     |
 | Scroll events      | 3/s (throttled) | Event → AI     |
 | Click events       | 0.05/s          | Event → AI     |
 | AI responses       | 0.02/s          | Push to client |
 | **Writes to disk** | 0.2/s           | Batched events |

**100 concurrent monitored users:**
- Events to process: 300/s
- AI invocations: 2-10/s (batched/debounced)
- Writes: 20/s

**Capacity:** 500+ monitored users

✅ **Verdict: Easy**

---

### Penpot-Lite Capacity Model

A simplified Figma competitor with real-time collaboration.

**Target: 10 collaborators per project**

 | Component             | Requirement       | Fazt Capacity   | Status   |
 | -----------           | -------------     | --------------- | -------- |
 | WebSocket connections | 10/project        | 5,000 total     | ✅       |
 | Cursor broadcast      | 300/s per project | 10,000/s total  | ✅       |
 | Shape operations      | 30/s per project  | 800/s total     | ✅       |
 | Undo/redo stack       | In-memory         | Unlimited       | ✅       |
 | Asset storage         | Blob store        | 800 uploads/s   | ✅       |
 | Projects concurrent   | 25-50             | -               | ✅       |

**Realistic limits on $6 VPS:**
- 25-50 active projects simultaneously
- 5-15 collaborators per project
- 200-500 total concurrent users

✅ **Verdict: Achievable for small-team use case**

---

## Real-Time Architecture Tips

### Batch Persistence

Never persist on every operation:

```javascript
// Bad: write every keystroke
onKeystroke(char) {
  fazt.docs.update(docId, { content: newContent })  // 800/s limit!
}

// Good: batch persist
const pendingOps = []
onKeystroke(char) {
  pendingOps.push(char)
  fazt.realtime.broadcast('doc:' + docId, { op: char })  // Unlimited
}
setInterval(() => {
  if (pendingOps.length) {
    fazt.docs.update(docId, { ops: pendingOps })  // 1 write/sec
    pendingOps = []
  }
}, 1000)
```

### Ephemeral vs Durable Channels

```javascript
// Cursors: ephemeral (never persist)
fazt.realtime.broadcast('cursors:' + docId, { userId, x, y })

// Chat: durable (persist then broadcast)
await fazt.docs.insert('messages', { text, userId, docId })
fazt.realtime.broadcast('chat:' + docId, { text, userId })
```

### Throttle High-Frequency Events

```javascript
// Client-side throttle for cursor
let lastSent = 0
onMouseMove(e) {
  if (Date.now() - lastSent > 33) {  // 30fps max
    ws.send({ type: 'cursor', x: e.x, y: e.y })
    lastSent = Date.now()
  }
}
```

---

## Summary

Fazt on a $6 VPS handles more traffic than most personal projects will ever
see. The single-writer model trades maximum throughput for **reliability** -
you'll never see failed writes due to contention.

**Real-time capacity:**
- Presence, chat, typing indicators: **Thousands of users**
- Collaborative docs: **Hundreds of concurrent editors**
- Design tools (Penpot-lite): **200-500 users across 25-50 projects**

For 99% of use cases: just build your app, don't worry about scaling.
