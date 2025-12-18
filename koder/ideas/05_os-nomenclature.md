# Vision Document: Fazt v0.8 - Shift to OS Nomenclature

**Status**: Draft / Request for Comment (RFC)
**Target**: v0.8 Architecture
**Scope**: High-Level Strategic Pivot

---
# Cartridge OS Nomenclature

## Alignment with 50 Years of Systems Engineering

### 1. Vision

Adopt 1:1 nomenclature with Operating System (OS) Kernels.
Eliminate mental translation between "App logic" and "System logic."
Leverage existing world-class mental models for long-term design.

### 2. Component Mapping (Domain Language)

* **`proc/` (Process)**: Lifecycle, systemd, and binary reboots.
* **`fs/` (File System)**: DB-backed Virtual File System (VFS).
* **`net/` (Network)**: SSL, routing, and domain management.
* **`storage/` (Block Storage)**: SQLite engine and EDD migrations.
* **`security/` (Access)**: JWT, auth, and root permissions.
* **`driver/` (External)**: Litestream S3 backup driver.
* **`syscall/` (Interface)**: The API bridge between App and Kernel.

### 3. CLI Restructure (OS Interface)

* **`fazt proc start/upgrade`**: Manages the binary lifecycle.
* **`fazt net route add`**: Provisions domains and SSL.
* **`fazt fs mount/ls`**: Manages the Virtual File System.
* **`fazt storage migrate/backup`**: Manages the data layer.
* **`fazt security root-pass`**: Manages system-level auth.

### 4. Internal API (The Syscall Pattern)

Applications (Verticals) never touch kernel packages directly.
They issue "Syscalls" through a unified Kernel object.
**Example**: `kernel.FS.Read()` instead of `vfs.Get()`.

### 5. Significance for AI Coding Agents

Agents respond better to rigid, standard system domains.
**Command**: "Add a field to storage following EDD rules."
**Command**: "Update the net module to support IPv6."
The agent knows exactly where to look and what logic to follow.

### 6. Expected Outcomes

* **Zero Ambiguity**: File paths match CLI commands match mental models.
* **Resilience**: The Kernel protects the system from App-level crashes.
* **Portability**: The binary acts as a self-contained, portable OS.

---

### Solo Architect Advice

Consistency is the ultimate safeguard for solo developers.
When the CLI feels like an OS, you treat your server like an appliance.
The DB holds the world; the Binary defines the rules.
