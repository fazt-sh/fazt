# Project Fazt: Stability & Safeguards Architecture
**Target:** Transform Fazt from a "fragile" single-binary PaaS into an **"Unsinkable"** platform that survives extreme resource constraints (e.g., 200MB RAM host with a 20TB Database).

**Core Philosophy:**
1.  **Single Binary + Single Database:** No external files, no sidecar processes.
2.  **Stability > Performance:** It is better to be slow than to crash (OOM).
3.  **The "Cockpit" Rule:** The Admin Dashboard and Authentication must *always* be responsive, even if the user data is thrashing the disk.

---

## 1. The "Cockpit" Architecture (Admin Persistence)
**Goal:** Decouple Admin UI performance from User Data I/O. Solves the "Frozen Login" problem on heavy loads.

* **Boot-Time Hydration (Pinned Cache):**
    * **Mechanism:** Utilize the "Quiet Window" at startup (before opening port 443).
    * **Action:** Query the SQLite DB for all `admin/*` assets and the `config` table.
    * **Storage:** Load these specific assets into a special `fs.pinned` map in RAM.
    * **Constraint:** These items are **never evicted**, regardless of memory pressure.
    * **Result:** Admin UI loads from RAM (0 disk I/O), bypassing any database locks or thrashing.

* **RAM-Resident Authentication:**
    * **Action:** Load the `PasswordHash` and `SessionKey` from the DB into Global RAM variables during the Boot/Hydration phase.
    * **Logic:** `POST /api/login` verifies against RAM variables, not the DB.
    * **Benefit:** Login is instant even if the database is 100% saturated by a large query.

* **Config Write Strategy:**
    * **Read:** From RAM (Instant).
    * **Write:** To DB via **WAL Mode** (Write-Ahead Logging). This allows Admin config updates to succeed even if the main DB file is locked by a massive read operation (like streaming video).

## 2. The Stability Engine (Dynamic Thresholds)
**Goal:** Auto-tune the application to fit the hardware "Container" it is installed in.

* **Install-Time Probe:**
    * During `fazt server init` or `install.sh`, probe `syscall.Sysinfo` (RAM) and `runtime.NumCPU()`.
    * **Formula:** `Usable_Resources = (Total_Hardware - OS_Overhead) * 0.7 Safety_Factor`.
    * **Output:** Calculate specific limits (Cache Size, Concurrency) and write them to the `config` table.

* **Recalibration:**
    * **UI:** Add a "Recalibrate System" button in the Admin Dashboard.
    * **Logic:** Re-runs the hardware probe, updates the `config` table, and triggers a "Soft Restart" of the VFS and Runtime engines (flushing caches and resizing semaphores) without dropping network connections.

* **Maintenance Mode (The Failsafe):**
    * **Boot Check:** On startup, calculate `DB_Size / Physical_RAM`.
    * **Logic:** If the ratio is critical (e.g., >100x), boot into **Maintenance Mode**.
    * **Behavior:** Only the Pinned Admin UI is served. All User Sites (`*.domain`) return `503 Service Unavailable`.
    * **Purpose:** Allows the admin to fix/export data without the server crashing OOM immediately.

## 3. Resource Safeguards (The "Bouncers")
**Goal:** Shift from "Implicit Limits" (Crash) to "Explicit Limits" (Rejection).

* **RAM (The Byte-Cap VFS):**
    * **Refactor:** Replace `count > 1000` eviction logic in `vfs.go`.
    * **New Logic:** Implement `TotalBytes > Config.VFSCacheSizeMB`.
    * **Behavior:** Track total bytes used by `fs.cache`. If limit exceeded, evict LRU items.
    * **Large Files:** If a requested file is > 5MB (configurable), **never** load it into the RAM cache. Use a zero-copy stream.

* **CPU (Runtime Semaphore):**
    * **Mechanism:** Buffered Channel `make(chan struct{}, Config.MaxConcurrentJS)`.
    * **Logic:** JS Runtimes (`goja`) must acquire a token before execution. If no tokens available, return `429 Too Many Requests`.
    * **Purpose:** Prevents 100 concurrent scripts from freezing the single vCPU.

* **Disk (SQLite Constraints):**
    * **PRAGMA:** Apply strict `PRAGMA cache_size` and `PRAGMA mmap_size=0` based on the Hardware Probe.
    * **Benefit:** Forces SQLite to thrash the disk rather than filling RAM, ensuring the OS doesn't kill the process.

## 4. Educational & UI Modules
**Goal:** Transparently communicate limits to the user.

* **Interactive Simulator (Website):**
    * A JS-based module for the landing page.
    * **Inputs:** Sliders for RAM, CPU, Storage.
    * **Outputs:** Real-time calculation of "Safe Capabilities" (e.g., "Max Upload: 50MB", "Concurrent Users: 10").
    * **Purpose:** Sets user expectations before they even install.

* **Dashboard Warnings:**
    * If a user tries to manually override thresholds (e.g., set 500MB upload on a 1GB server), show a "Crash Probability" warning or require a "Break Glass" confirmation.

---

**Implementation Priority:**
1.  **Refactor VFS:** Implement Byte-Limited LRU & Pinned Caching (Hydration).
2.  **Refactor Auth:** Implement RAM-based check.
3.  **Implement Probe:** Build the hardware detection and `ThresholdConfig` logic.
4.  **Implement Safeguards:** Semaphore for JS and Streaming for large blobs.
