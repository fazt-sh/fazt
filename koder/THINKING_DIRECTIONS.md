# Fazt Thinking Directions

Living document of ideas, evaluations, and strategic directions to explore.
Reference specific sections from STATE.md when actively pursuing them.

---

## Product

### P1. Google Sign-in Redirect
- Currently always redirects to root after OAuth
- Either fix the redirect to original page, or make the landing page pretty

### P2. Nexus App (Stress Test All Capabilities)
- Find/build an app idea that uses ALL fazt capabilities
- Consider making Nexus do this
- Goal: comprehensive stress test of the platform

### P3. App Audit
- Verify all docs are synced
- List all deployed apps
- Enhance apps with Google sign-in
- Games should have high score tracking

### P4. Target Apps (Small Startup IT Suite)
- Meet (scheduling)
- Docs (collaborative editing)
- Chat (real-time messaging)
- Notes (Notion-like)
- Sign (document signing)
- Files (file sharing)

---

## Engineering

### E1. `fazt @peer` Pattern Audit
- Review all commands for `@peer` support
- `fazt app list` seems incomplete - should show better results
- Ensure CLI ↔ API 1:1 parity (all CLI commands have API equivalents)

### E2. Analytics Deep Dive
- Audit: are all analytics properly collected & stored?
- Can analytics track users for comprehensive data flow view?
- Evaluate: is every state change captured? Should it be? (perf tradeoffs)
- Consider config options to disable some analytics for efficiency
- Need visualization/dashboard

### E3. Role-Based Access Control (RBAC)
- Can owner also Google sign-in and system recognizes them by email?
- Granular permissions per app/resource
- Current: owner vs user (OAuth) - needs refinement

### E4. Plan 24: Mock OAuth Provider
- Dev login form at `/auth/dev/login` (local only)
- Creates real sessions (same as production OAuth)
- Role selection for testing admin/owner flows
- Why: Can't test auth flows locally without HTTPS

### E5. Plan 25: SQL Command
```bash
fazt sql "SELECT * FROM apps"              # Local
fazt @zyt sql "SELECT * FROM auth_users"   # Remote
```
- Read-only by default, `--write` flag for mutations
- Why: Currently requires SSH + sqlite3 for remote debugging

### E6. Qor Extraction Evaluation
- Review `~/Projects/qor` for reusable components
- Identify patterns, utilities, or features worth extracting into fazt
- Use `/lite-extract` skill for evaluation
- Questions:
  - What does qor do well that fazt lacks?
  - Any battle-tested code worth porting?
  - Architectural patterns to consider?

---

## Documentation

### D1. Documentation Overhaul
- Build comprehensive markdown-based, multi-file fazt documentation
- Organize as a Claude skill (usable directly in Claude Code)
- Structure should:
  - Sync with API/CLI changes
  - Generate documentation site
  - Drive development vision
- README.md is outdated - needs refresh

---

## Strategy

### S1. Capability Comparison
- Re-evaluate vs Supabase & Vercel
- What do they have that we don't?
- What do we have that they don't?
- Where's the gap?

### S2. Concept: "Break Hyperscaler Stack"
- Can fazt be the unit of compute that replaces cloud lock-in?
- Each fazt = sovereign, portable, interconnectable

### S3. Vertical Scaling Evaluation
- Current: $6 VPS handles 2,300 req/s
- What about $50 VPS? $500 VPS?
- At what point does horizontal > vertical?

### S4. External Integration Value Matrix
- Rank by value-add: S3, Litestream, Turso, Cloudflare, others
- Which integration unlocks most capability?
- Which has best effort/reward ratio?

---

## Business

### B1. License Discussion
- Reconsider MIT license
- Proposed model: **Fair Code License**
  - MIT/free for everyone doing <$1M revenue
  - $1000 per $1M revenue above threshold
- Questions:
  - Is $1000/$1M too high? Too low?
  - How to enforce/verify?
  - Precedents (Elastic, MongoDB, BSL, FSL)?

### B2. Cloud Provider Partnerships
- DigitalOcean, MS, Google may want one-click installer
- Don't scare them off - perhaps volume/partnership deals?

### B3. Private Repo Feasibility
- Evaluate making repo private until license figured out
- Will any dependencies break?
- Impact on current users?

---

## Vision

### V1. Philosophy Rewrite
- Update `koder/philosophy/` docs
- Fazt has evolved significantly (see `koder/scratch/03_whats-fazt.md`)
- Rewrite for new vision & direction
- Should reflect: sovereign compute, single binary, hyperscaler alternative
