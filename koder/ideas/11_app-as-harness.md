# Vision Document: Fazt v0.12 - The Agentic Harness Architecture

**Status**: Draft / Request for Comment (RFC)
**Target**: v0.12+ Architecture
**Scope**: Self-Evolving Infrastructure & Agentic Execution

---

## 1. Executive Summary: The Intelligence Layer

Fazt v0.8 established the "Kernel" metaphor; Fazt v0.12 introduces the **"Harness."** Instead of treating AI agents as external tools that occasionally touch the server, we move the agentic "Brain" inside the "Body" (the Cartridge).

A **Harness** is a specialized Fazt App designed to host, manage, and evolve LLM-based autonomous agents. By leveraging the **Goja (JS) Runtime**, the **VFS**, and a new **Git-Native SDK**, Fazt becomes a self-improving "Biological Operating System".

---

## 2. Core Pillars

### A. The "Harness-as-App" Model

We reject the complexity of containerizing Node.js-based CLI agents (like Claude Code). Instead, we build the harness as a native Fazt App.

* **Low-Density Execution**: Harnesses are lightweight JS apps that leverage the **Hibernate Architecture**—consuming zero RAM until a trigger occurs.
* **Portable Intelligence**: Because it is an app, the entire harness (logic, tools, and memory) lives in `data.db`. Moving your "Architect" to a new server is as simple as moving the SQLite file.

### B. The Native Agentic SDK

To empower these harnesses, the `fazt` JS library is extended with native "Syscalls" that provide a read-only baseline for observation and high-speed intervention:

* **`fazt.git.*`**: Atomic commits, branching, and diffing directly against the SQLite-backed VFS.
* **`fazt.ai.*`**: Standardized, credential-managed gateway to LLM providers (Gemini, OpenAI, Anthropic).
* **`fazt.kernel.*`**: Privileged tools to deploy new apps, manage subdomains, and monitor system health.

### C. Self-Evolution (The Meta-Harness)

A harness is "Resident Intelligence" that can modify its own source code.

* **Recursive Tooling**: If a harness identifies a missing capability, it uses its VFS/Git tools to write new JS logic into its own `/api` folder.
* **The Growth Loop**: Harnesses can "spawn" specialized child agents for sub-tasks (e.g., a "Security Auditor" harness or a "UI Designer" harness) within the same VFS environment.

---

## 3. Implementation Details

### A. Permission-Based Autonomy

Harnesses follow a "Graceful Degradation" permission model:

* **Superuser Requests**: A harness requests elevated access in `app.json` (e.g., `kernel:deploy`, `vfs:global_read`).
* **Safety Intercepts**: If permissions are denied, the harness operates in a restricted mode—capable of internal tasks but unable to affect other apps or system-wide routing.
* **Localized Blast Radius**: Failures or "agentic hallucinations" are contained within the app's UUID or the specific `data.db` instance, preventing global system compromise.

### B. Dual Interface: Headless & Dashboard

* **Headless Mode**: The harness operates as a background worker, processing webhooks, email sinks, or scheduled cron tasks (`js-cron`).
* **Visual Cockpit**: The app's frontend provides a UI within the Admin Dashboard to visualize token usage, tool-call logs, "thought" streams, and active deployments.

---

## 4. Strategic Maxima: The Living Database

The ultimate goal of the Harness architecture is the **Sovereign Entity**. The `data.db` file no longer just holds static content; it holds a **Living Infrastructure** that:

1. **Observes** its own state via native baseline tools.
2. **Architects** new solutions by deploying sub-apps.
3. **Evolves** its own logic to become more efficient over time.

**Verdict**: By building the harness as an app, we achieve the rapid iteration of JS engineering with the stability and portability of the Fazt Kernel.