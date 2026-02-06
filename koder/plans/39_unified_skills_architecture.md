# Plan 39: Unified Skills Architecture (Claude + Codex)

**Created**: 2026-02-06
**Updated**: 2026-02-06 (Simplified based on user feedback)
**Status**: APPROVED - IN PROGRESS
**Goal**: Consolidate all skills into knowledge-base/ with multi-agent support via symlinks

---

## Overview

### Why Unified Architecture?

**Current Problem:**
- Skills scattered: `.claude/commands/`, `knowledge-base/skills/app/`, `~/.claude/skills/`
- Inconsistent structure: flat files vs directories
- Single-agent: Only supports Claude Code CLI
- Not git-tracked: Project skills in `.claude/` hidden from version control
- CLAUDE.md and AGENTS.md would drift (dual maintenance)

**Solution:**
Single source of truth in `knowledge-base/` for both skills AND instructions, with symlinks for agent discovery.

**Benefits:**
1. **Skills as documentation** - Part of knowledge-base, versioned with project
2. **Multi-agent ready** - Claude Code + OpenAI Codex (+ future agents)
3. **Git-tracked** - Skills are code, should be versioned
4. **Consistent structure** - All skills follow same pattern
5. **Flexible deployment** - Local or global via symlinks
6. **Room for growth** - Each skill can have `references/`, `examples/`, `scripts/`
7. **Single AGENTS.md** - No drift! Both CLAUDE.md and AGENTS.md symlink to knowledge-base/AGENTS.md

### What Changes

| Aspect | Current | Target |
|--------|---------|--------|
| **Project skills location** | `.claude/commands/*.md` | `knowledge-base/skills/*/SKILL.md` |
| **Structure** | Flat markdown files | Directories with SKILL.md + resources |
| **Git tracking** | Not tracked | Tracked in knowledge-base/ |
| **Agent support** | Claude only | Claude + Codex |
| **Discovery** | Direct (`.claude/commands/`) | Symlinked (`.claude/skills/` → `knowledge-base/skills/`) |
| **Frontmatter** | Claude-specific | Unified (both compatible) |

### What Stays the Same

✅ **Skill invocation** - `/fazt-release`, `/fazt-open` work exactly the same (with fazt- prefix)
✅ **Skill content** - Markdown bodies unchanged
✅ **Claude features** - `model: opus`, `allowed-tools` preserved
✅ **Backward compatibility** - Existing workflows unaffected (no shims needed - skills work fine)

### User Decisions (from koder/scratch/08_skills-restructure.md)

1. ✅ **No shims for .claude/commands/** - Skills work fine, remove completely
2. ✅ **Single AGENTS.md in knowledge-base/** - Symlink to root for both CLAUDE.md and AGENTS.md
3. ✅ **Clean .agents/skills/** - Symlink skills into it, remove agent-browser (manage globally)
4. ✅ **tooling.md in knowledge-base/** - Move to knowledge-base/agent-context/
5. ✅ **fazt-app is global** - Rest are local symlinks

### Known Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Symlink issue #8943 (Codex) | Low-Medium | Medium | Test early; fallback to copies |
| Git history loss on move | Low | Low | Use `git mv` with careful commits |
| Per-skill model selection lost in Codex | Certain | Low | Document workaround (Codex profiles) |
| AGENTS.md/CLAUDE.md drift | High | Low | Accept during experiment; sync tooling later |
| Skill discovery breaks | Low | High | Test thoroughly; rollback plan ready |

---

## Detailed Inventory

### Skills to Migrate

| Skill | Current Path | Lines | Frontmatter | Notes |
|-------|-------------|-------|-------------|-------|
| **release** | `.claude/commands/release.md` | 213 | `description`, `model: opus` | Complex; references scripts, env vars |
| **open** | `.claude/commands/open.md` | 95 | `description` | Simple; quick session start |
| **close** | `.claude/commands/close.md` | 160 | `description` | Medium; updates STATE.md |
| **ideate** | `.claude/commands/ideate.md` | 91 | `description`, `model: opus` | Uses Opus for brainstorming |
| **lite-extract** | `.claude/commands/lite-extract.md` | 295 | `description`, `allowed-tools` | Complex; web search, analysis |

### Skills Already Migrated

| Skill | Location | Status |
|-------|----------|--------|
| **fazt-app** | `knowledge-base/skills/app/` | ✅ Already in target location |

### External Global Skills (Not Migrating)

| Skill | Location | Action |
|-------|----------|--------|
| **qwen-research** | `~/.claude/skills/qwen-research/` | Keep in global; optionally add Codex symlink |
| **agent-browser** | Symlink to external repo | Keep as-is |
| **keybindings-help** | System skill | No action |
| **frontend-design** | Unknown location | Investigate if needed |

---

## Target Architecture (Simplified)

```
knowledge-base/
├── AGENTS.md                       # SOURCE for instructions (single source!)
├── agent-context/
│   └── tooling.md                  # Moved from root, updated
└── skills/                         # SOURCE for skills (git-tracked, versioned)
    ├── app/                        # fazt-app (already here)
    │   ├── SKILL.md                # name: fazt-app
    │   ├── examples/
    │   ├── patterns/
    │   ├── references/
    │   └── fazt/
    ├── release/                    # NEW (migrated)
    │   └── SKILL.md                # name: fazt-release
    ├── open/                       # NEW (migrated)
    │   └── SKILL.md                # name: fazt-open
    ├── close/                      # NEW (migrated)
    │   └── SKILL.md                # name: fazt-close
    ├── ideate/                     # NEW (migrated)
    │   └── SKILL.md                # name: fazt-ideate
    └── lite-extract/               # NEW (migrated)
        └── SKILL.md                # name: fazt-lite-extract

# Root symlinks (both agents read same source!)
./CLAUDE.md → knowledge-base/AGENTS.md
./AGENTS.md → knowledge-base/AGENTS.md

# Local skill symlinks (Claude discovery)
.claude/skills/
├── open → ../../knowledge-base/skills/open
├── release → ../../knowledge-base/skills/release
├── close → ../../knowledge-base/skills/close
├── ideate → ../../knowledge-base/skills/ideate
└── lite-extract → ../../knowledge-base/skills/lite-extract

# Local skill symlinks (Codex/OpenCode discovery)
.agents/skills/
├── open → ../../knowledge-base/skills/open
├── release → ../../knowledge-base/skills/release
├── close → ../../knowledge-base/skills/close
├── ideate → ../../knowledge-base/skills/ideate
└── lite-extract → ../../knowledge-base/skills/lite-extract

# Global symlinks (fazt-app ONLY - rest are local)
~/.claude/skills/fazt-app → ~/Projects/fazt/knowledge-base/skills/app
~/.agents/skills/fazt-app → ~/Projects/fazt/knowledge-base/skills/app
```

### Key Simplifications

1. **Single AGENTS.md** - Both CLAUDE.md and AGENTS.md symlink to `knowledge-base/AGENTS.md` (no drift!)
2. **No shims** - `.claude/commands/` removed completely (skills work fine)
3. **Clean .agents/skills/** - Only symlinks to knowledge-base/, agent-browser removed
4. **Prefixed names** - `fazt-open`, `fazt-release` etc. to avoid accidental triggers
5. **Global only for fazt-app** - Project-specific skills stay local

---

## Unified Frontmatter Format

Both Claude and Codex ignore unknown frontmatter keys, allowing a merged format.
**IMPORTANT**: Use `fazt-` prefix to avoid accidental triggers on generic words like "open", "close".

```yaml
---
name: fazt-release                 # Required by Codex (with fazt- prefix!)
description: Release a new...      # Required by both
model: opus                        # Claude reads; Codex ignores
allowed-tools: Read, Write, Bash   # Claude reads; Codex ignores
---
```

**Example: release skill**

```yaml
---
name: fazt-release
description: Release a new version of fazt with proper versioning, documentation, and deployment. Use when asked to release, publish, or deploy a new version.
model: opus
---
# Fazt Release Skill
...
```

---

## Migration Steps

### Phase 0: Validation ✅ (Already Done)

**Goal:** Verify `.claude/skills/` pattern works

- [x] Created test skill in `.claude/skills/test-skill/`
- [x] Verified Claude discovers it
- [x] Confirmed directory-level symlinks work

**Result:** `.claude/skills/` fully supported for project-level skills.

---

### Phase 1: Knowledge Base Restructure

**Goal:** Move skills from `.claude/commands/` to `knowledge-base/skills/`

#### Step 1.1: Create Skill Directories

```bash
cd knowledge-base/skills
mkdir -p release open close ideate lite-extract
```

#### Step 1.2: Migrate Files with Git History

Use `git mv` to preserve history:

```bash
# From repo root
git mv .claude/commands/release.md knowledge-base/skills/release/SKILL.md
git mv .claude/commands/open.md knowledge-base/skills/open/SKILL.md
git mv .claude/commands/close.md knowledge-base/skills/close/SKILL.md
git mv .claude/commands/ideate.md knowledge-base/skills/ideate/SKILL.md
git mv .claude/commands/lite-extract.md knowledge-base/skills/lite-extract/SKILL.md

# Commit the moves
git commit -m "refactor: migrate skills to knowledge-base/

- Move .claude/commands/*.md to knowledge-base/skills/*/SKILL.md
- Preserves git history for all skills
- Part of unified skills architecture (Plan 39)"
```

#### Step 1.3: Update Frontmatter

Add `name` field to each SKILL.md (required by Codex):

**release:**
```yaml
---
name: release
description: Release a new version of fazt with proper versioning, documentation, and deployment. Use when asked to release, publish, or deploy a new version.
model: opus
---
```

**open:**
```yaml
---
name: open
description: Fazt Session Open - Get up to speed quickly at the beginning of a work session.
---
```

**close:**
```yaml
---
name: close
description: Fazt Session Close - Document session work and update state.
---
```

**ideate:**
```yaml
---
name: ideate
description: Brainstorm ideas for Fazt's evolution (project) - uses reasoning model for strategic planning.
model: opus
---
```

**lite-extract:**
```yaml
---
name: lite-extract
description: Evaluate if a library/project can be "lite-extracted" into Fazt - analyze size, dependencies, and integration feasibility.
allowed-tools: Read, Write, Edit, Glob, Grep, Bash, WebFetch, WebSearch
---
```

Commit:
```bash
git add knowledge-base/skills/*/SKILL.md
git commit -m "feat: add Codex-compatible frontmatter to skills

- Add 'name' field to all SKILL.md files
- Preserve Claude-specific fields (model, allowed-tools)
- Unified format works for both Claude and Codex"
```

#### Step 1.4: Remove Old Commands Directory

```bash
rmdir .claude/commands
git add .claude/commands
git commit -m "cleanup: remove deprecated .claude/commands directory

Skills now live in knowledge-base/skills/"
```

**Testing:**
- [ ] All SKILL.md files have valid frontmatter
- [ ] Git history intact (`git log --follow knowledge-base/skills/release/SKILL.md`)
- [ ] No files left in `.claude/commands/`

---

### Phase 2: Local Symlinks (Claude)

**Goal:** Point Claude to skills in knowledge-base/

#### Step 2.1: Create Symlink Directory

```bash
mkdir -p .claude/skills
```

#### Step 2.2: Create Symlinks

```bash
cd .claude/skills

# Symlink all skills
ln -s ../../knowledge-base/skills/app app
ln -s ../../knowledge-base/skills/release release
ln -s ../../knowledge-base/skills/open open
ln -s ../../knowledge-base/skills/close close
ln -s ../../knowledge-base/skills/ideate ideate
ln -s ../../knowledge-base/skills/lite-extract lite-extract

# Verify symlinks
ls -la
```

**Expected output:**
```
app -> ../../knowledge-base/skills/app
release -> ../../knowledge-base/skills/release
open -> ../../knowledge-base/skills/open
close -> ../../knowledge-base/skills/close
ideate -> ../../knowledge-base/skills/ideate
lite-extract -> ../../knowledge-base/skills/lite-extract
```

#### Step 2.3: Update .gitignore

Add symlink directories to gitignore:

```bash
# Add to .gitignore
echo "" >> .gitignore
echo "# Agent skill symlinks (sources tracked in knowledge-base/)" >> .gitignore
echo ".claude/skills/" >> .gitignore
echo ".agents/skills/" >> .gitignore
```

Remove old exception (if exists):
```bash
# Remove line: !.claude/commands/fazt-*.md
sed -i '/!\.claude\/commands\/fazt-\*\.md/d' .gitignore
```

Commit:
```bash
git add .gitignore
git commit -m "chore: update .gitignore for unified skills architecture

- Ignore .claude/skills/ and .agents/skills/ (symlinks)
- Remove deprecated .claude/commands/ exception"
```

**Testing:**
- [ ] Claude discovers all skills: `claude "List available skills"`
- [ ] Skill invocation works: Try `/open`, `/release`
- [ ] Git status clean (symlinks ignored)

---

### Phase 3: Codex Setup

**Goal:** Configure Codex to use the same skills

#### Step 3.1: Install Codex CLI

```bash
npm install -g @openai/codex
codex --version
```

#### Step 3.2: Create Global AGENTS.md

Copy from Claude's CLAUDE.md:

```bash
mkdir -p ~/.codex
cp ~/.claude/CLAUDE.md ~/.codex/AGENTS.md
```

**Manual edit:** Remove Claude-specific references if any.

#### Step 3.3: Create config.toml

Create `~/.codex/config.toml`:

```toml
#:schema https://developers.openai.com/codex/config-schema.json

model = "gpt-5.3-codex"
approval_policy = "on-request"
sandbox_mode = "workspace-write"

[features]
web_search = "cached"
```

#### Step 3.4: Create Project AGENTS.md

Copy from CLAUDE.md:

```bash
cp CLAUDE.md AGENTS.md
```

**Manual edits:**
1. Update skill paths: `.claude/commands/` → `knowledge-base/skills/`
2. Add note at top: "This file is synced from CLAUDE.md for Codex CLI compatibility"

Commit:
```bash
git add AGENTS.md
git commit -m "feat: add AGENTS.md for Codex CLI support

- Ported from CLAUDE.md
- Updated skill references to knowledge-base/skills/
- Coexists with CLAUDE.md during evaluation period"
```

#### Step 3.5: Create Local Symlinks (Codex)

```bash
mkdir -p .agents/skills
cd .agents/skills

# Same symlinks as Claude
ln -s ../../knowledge-base/skills/app app
ln -s ../../knowledge-base/skills/release release
ln -s ../../knowledge-base/skills/open open
ln -s ../../knowledge-base/skills/close close
ln -s ../../knowledge-base/skills/ideate ideate
ln -s ../../knowledge-base/skills/lite-extract lite-extract
```

**Testing:**
- [ ] Codex discovers skills: `codex "List available skills"`
- [ ] Skill invocation works: Try invoking skills
- [ ] Symlinks resolve correctly: `ls -la .agents/skills/`
- [ ] No symlink issue #8943 (if fails, use copies instead)

---

### Phase 4: Global Symlinks (Optional)

**Goal:** Make fazt-app available globally for both agents

#### Step 4.1: Claude Global Symlink

```bash
mkdir -p ~/.claude/skills
ln -s ~/Projects/fazt/knowledge-base/skills/app ~/.claude/skills/fazt-app
```

#### Step 4.2: Codex Global Symlink

```bash
mkdir -p ~/.agents/skills
ln -s ~/Projects/fazt/knowledge-base/skills/app ~/.agents/skills/fazt-app
```

**Testing:**
- [ ] `/fazt-app` works from any directory (Claude)
- [ ] Codex discovers fazt-app globally
- [ ] Source remains single (knowledge-base/skills/app/)

---

### Phase 5: Documentation & Cleanup

**Goal:** Update all references and finalize migration

#### Step 5.1: Update CLAUDE.md

**Section: Key Paths**

Replace:
```markdown
.claude/commands/     # Project commands
```

With:
```markdown
knowledge-base/skills/  # Project & global skills (source)
.claude/skills/         # Symlinks to knowledge-base/skills/
.agents/skills/         # Symlinks for Codex
```

**Section: Deep Context**

Update skill references to point to `knowledge-base/skills/`.

**Section: Essential Commands**

Add note about skill architecture:
```markdown
## Skills Architecture

Skills live in `knowledge-base/skills/` (git-tracked).
Discovery via symlinks:
- Claude: `.claude/skills/` → `knowledge-base/skills/`
- Codex: `.agents/skills/` → `knowledge-base/skills/`
```

Commit:
```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for unified skills architecture

- Updated Key Paths section
- Documented symlink pattern
- Clarified skills location"
```

#### Step 5.2: Update Memory Files

Update `~/.claude/projects/-home-kodeman-Projects-fazt/memory/MEMORY.md` if it references `.claude/commands/`.

#### Step 5.3: Update Skill Documentation

If any skills reference other skills by path, update to `knowledge-base/skills/`.

**Testing:**
- [ ] All documentation accurate
- [ ] No broken links
- [ ] Skill references correct

---

## Testing Checklist

### Discovery

- [ ] Claude discovers all 6 skills locally (app, release, open, close, ideate, lite-extract)
- [ ] Codex discovers all 6 skills locally
- [ ] Global fazt-app works in Claude
- [ ] Global fazt-app works in Codex

### Invocation

- [ ] `/release` works in Claude
- [ ] `/open` works in Claude
- [ ] `/close` works in Claude
- [ ] `/ideate` works in Claude (uses opus)
- [ ] `/lite-extract` works in Claude
- [ ] `/fazt-app` works in Claude
- [ ] Skills invoke correctly in Codex

### Git Tracking

- [ ] `knowledge-base/skills/` tracked
- [ ] `.claude/skills/` ignored
- [ ] `.agents/skills/` ignored
- [ ] AGENTS.md tracked
- [ ] Git history preserved for migrated skills

### Backward Compatibility

- [ ] Existing workflows unaffected
- [ ] Claude-specific features work (model selection, tool restrictions)
- [ ] Skill content unchanged
- [ ] No regression in functionality

### Cross-Agent

- [ ] Same source for both agents (no drift)
- [ ] Edit once, both agents see changes
- [ ] Unified frontmatter works for both

---

## Rollback Plan

If anything breaks, rollback steps:

### 1. Restore Commands Directory

```bash
git checkout HEAD~N -- .claude/commands/
git commit -m "rollback: restore .claude/commands/"
```

Replace `N` with number of commits to undo.

### 2. Remove Symlinks

```bash
rm -rf .claude/skills/ .agents/skills/
```

### 3. Remove Codex Files

```bash
rm -f AGENTS.md
rm -rf ~/.codex/ ~/.agents/
```

### 4. Revert Knowledge Base Changes

```bash
git checkout HEAD~N -- knowledge-base/skills/
git commit -m "rollback: revert knowledge-base skills"
```

### 5. Verify Rollback

- [ ] `/release` works in Claude (old location)
- [ ] All skills functional
- [ ] Git status clean

---

## Benefits Analysis

### Before (Current State)

| Metric | Value |
|--------|-------|
| **Skill locations** | 3 (`.claude/commands/`, `knowledge-base/skills/app/`, `~/.claude/skills/`) |
| **Git-tracked** | Partial (fazt-app yes, project commands no) |
| **Agent support** | 1 (Claude only) |
| **Structure consistency** | Low (flat files + directories) |
| **Discoverability** | Medium (hidden in `.claude/`) |
| **Maintenance** | Scattered, inconsistent |

### After (Target State)

| Metric | Value |
|--------|-------|
| **Skill locations** | 1 (`knowledge-base/skills/`) |
| **Git-tracked** | Full (all skills versioned) |
| **Agent support** | 2+ (Claude, Codex, future-ready) |
| **Structure consistency** | High (all follow same pattern) |
| **Discoverability** | High (part of knowledge-base docs) |
| **Maintenance** | Single source, easy to manage |

### Quantified Improvements

1. **Single source** - 3 locations → 1 location (67% reduction)
2. **Git tracking** - ~20% → 100% of skills tracked
3. **Multi-agent** - 0 → 2 agents supported (Claude + Codex)
4. **Consistency** - All skills follow `SKILL.md` + optional subdirs pattern
5. **Documentation** - Skills discoverable in knowledge-base/ (part of docs)

---

## Known Limitations

### 1. Per-Skill Model Selection (Codex)

**Issue:** Claude's `model: opus` directive doesn't work in Codex.

**Workaround:** Use Codex profiles:

```toml
# ~/.codex/config.toml
[profiles.ideate]
model = "gpt-5.3-codex"  # or reasoning model

[profiles.release]
model = "gpt-5.2-codex"
```

Invoke: `codex --profile ideate "brainstorm features"`

**Impact:** Less ergonomic than Claude's per-skill model, but functional.

### 2. Codex Symlink Issue #8943

**Issue:** Known Codex bug with symlink handling in some edge cases.

**Status:** Low likelihood, but test thoroughly.

**Fallback:** If symlinks fail, copy skills to `.agents/skills/` instead of symlinking.

### 3. Dual Instruction Files

**Issue:** CLAUDE.md and AGENTS.md must stay in sync.

**Current Strategy:** Manual sync during experimental phase.

**Future:** Script to generate AGENTS.md from CLAUDE.md, or commit to one system.

### 4. Tool Restrictions Lost (Codex)

**Issue:** Claude's `allowed-tools` directive ignored by Codex.

**Impact:** Minimal - tool restrictions are convenience, not security.

**Workaround:** None needed; Codex has its own tool authorization model.

---

## Future Considerations

### 1. Other Agents

This architecture is extensible to other coding agents:

- **Cursor**: Likely uses similar patterns
- **Windsurf**: May have different discovery
- **GitHub Copilot Workspace**: Unknown at this time

Pattern: Add `.cursor/skills/`, `.windsurf/skills/` symlinks as needed.

### 2. Skill Marketplace

`knowledge-base/skills/` could become a shareable collection:

```bash
# Another project
ln -s ~/Projects/fazt/knowledge-base/skills/app .agents/skills/fazt-app
```

### 3. Cross-Project Skill Imports

```bash
# In another project's .agents/skills/
ln -s ~/Projects/fazt/knowledge-base/skills/release release
```

Fazt's skills become reusable across projects.

### 4. Skill Subdirectories

As skills mature, add:

- `references/` - Detailed docs (progressive disclosure)
- `examples/` - Real-world usage examples
- `scripts/` - Executable helpers
- `assets/` - Templates, configs

**Example: release skill**

```
knowledge-base/skills/release/
├── SKILL.md
├── references/
│   ├── version-strategy.md
│   └── changelog-format.md
├── scripts/
│   └── verify-release.sh
└── templates/
    └── release-notes.md
```

### 5. Skill Testing Framework

Future: Automated tests for skills

```bash
# Test skill invocation
./scripts/test-skills.sh

# Verify discovery in all agents
./scripts/verify-discovery.sh
```

---

## Success Criteria

This plan is successful when:

1. ✅ All 6 skills work in both Claude and Codex
2. ✅ `knowledge-base/skills/` is single source of truth
3. ✅ Git tracking complete and history preserved
4. ✅ Symlinks work reliably for both agents
5. ✅ Documentation updated and accurate
6. ✅ No regression in existing workflows
7. ✅ Rollback plan tested and viable
8. ✅ Architecture extensible to future agents

---

## Timeline Estimate

| Phase | Estimated Time |
|-------|----------------|
| Phase 0: Validation | ✅ Complete |
| Phase 1: Restructure | 30 minutes |
| Phase 2: Claude symlinks | 15 minutes |
| Phase 3: Codex setup | 30 minutes |
| Phase 4: Global symlinks | 10 minutes |
| Phase 5: Documentation | 20 minutes |
| **Testing & verification** | 30 minutes |
| **Buffer for issues** | 30 minutes |
| **Total** | ~3 hours |

---

## Approval Required

This plan requires user approval before execution due to:

1. **Git history changes** - Moving files with `git mv`
2. **Directory structure changes** - New locations for skills
3. **Multi-agent commitment** - Adding Codex support
4. **Backward compatibility** - Removing `.claude/commands/`

**Questions for reviewer:**

1. Approve full migration or start with single skill POC?
2. Keep global symlinks or local only?
3. Commit to Codex or experimental only?
4. Prefer AGENTS.md sync script or manual sync?

**Next Steps:**

1. User/reviewer approves plan
2. Execute Phase 1 (git mv with careful commits)
3. Test each phase before proceeding
4. Document any issues encountered
5. Update this plan with actual results

---

**Status:** PROPOSED - Awaiting review and approval
