# Stash: Living Entity (MCP + Agent)

**Status**: Stashed (not ready for spec)
**Source**: `tmp-mcp-handoff.md`
**Date**: 2024-12-30

---

## Vision

Transform Fazt from passive toolbox to living digital entity:
```
User ↔ Claude ↔ Fazt ↔ World
       ↑            ↓
       └── pulse ←──┘
```

Characteristics:
- Controllable via natural language (MCP)
- Self-aware (Pulse)
- Autonomous (Agent)
- Prioritizes longevity (auto-renew, cleanup, alerts)

---

## Key Components

### 1. MCP.kernel (extends v0.12)
- Full tool surface for owner
- Capability tokens for auth
- Destructive ops trigger halt

### 2. MCP.app (new)
- Per-app tool registration
- Scoped permissions
- Apps expose their own MCP endpoints

### 3. Policy Engine (proposed primitive)
```javascript
fazt.policy.add({
    trigger: "pulse.ssl_expiring",
    condition: "cert.daysUntilExpiry < 7",
    action: "net.ssl.renew(cert.domain)",
    autonomy: "medium",
    escalate: { reason: "SSL cert expiring", halt: true }
})
```

### 4. Agent (pattern, not primitive)
- Observe (pulse subscription)
- Decide (ai shim + policy evaluation)
- Act (kernel operations)
- Learn (events + storage)

---

## Analysis Notes

**Is Agent a primitive?** Mostly no - it's a pattern composed of:
- Pulse (v0.8)
- AI shim (v0.12)
- Events (v0.8)
- Halt (v0.14)
- Storage (v0.9)

**What's actually missing?** Policy engine - declarative trigger→action rules with autonomy gates.

**Autonomy levels**:
- low: notify only
- medium: safe ops auto, destructive escalates
- high: everything except payments/deletion

---

## Open Questions

1. Where does policy engine fit? (v0.12 extension or new version?)
2. Per-instance vs per-app policies?
3. How does mesh + MCP work? (which node responds?)
4. Notification channels for escalations (use existing dev.email/sms?)

---

## When to Revisit

Consider spec'ing when:
- v0.12 MCP basics are implemented
- v0.14 halt primitive exists
- There's a concrete use case driving the need

---

## Reference

- Full discussion: `tmp-mcp-handoff.md`
- Existing MCP spec: `koder/ideas/specs/v0.12-agentic/mcp.md`
- Pulse spec: `koder/ideas/specs/v0.8-kernel/pulse.md`
