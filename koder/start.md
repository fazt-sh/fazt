# Fazt.sh - AI Bootloader

This document provides context to bootstrap a new coding session.

## 1. Core Philosophy 🧠
- **One Binary**: Single executable (`fazt`) + Single DB (`data.db`).
- **Zero Dependencies**: Pure Go + ModernC SQLite. No CGO.
- **Cartridge Architecture**: The DB is the filesystem. Sites live in SQL.
- **Safety**: `CGO_ENABLED=0` always.

## 2. Architecture Pillars 🏛️
- **VFS**: Sites stored in `files` table. In-memory LRU cache.
- **System Sites**: `root` and `404` seeded from binary (`internal/assets`).
- **Runtime**: `goja` JS runtime for serverless (`main.js`).
- **Routing**: Host-based (`admin.`, `root.`, `*.domain`).

## 3. Initialization Protocol 🚀
Perform these actions in order to load the correct context:

1. **Understand the Mission**:
   - `read_file koder/NEXT_SESSION.md` (Current Status, Plan, and **Required Context**).

2. **Understand the System**:
   - `read_file koder/analysis/04_comprehensive_technical_overview.md` (Architecture & Data Flow).
   - *Do not read raw source code yet unless directed by NEXT_SESSION.md.*

3. **Verify Environment**:
   - `read_file ~/.info.json` (Ports & Path).
   - Check for active tool scripts: `ls *.sh` (e.g., `probe_api.sh`).

## 4. Session Goal 🎯
After reading the above:
1. Summarize the **Current Phase** defined in `NEXT_SESSION.md`.
2. List the **Specific Files** you have loaded into context based on that plan.
3. State your readiness to execute the first step.