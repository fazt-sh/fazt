# Fazt Session Stop

Intelligent session close that adapts to what was accomplished.

## Step 1: Gather State

Run these commands to understand current state:

```bash
# Versions
grep "var Version" internal/config/config.go
fazt --version
fazt remote status zyt | grep -E "Version|Status"

# Git status
git status --short

# Recent commits this session (last 5)
git log --oneline -5
```

## Step 2: Detect Session Type

Analyze the conversation to determine what happened:

| Session Type | Indicators | STATE.md Approach |
|--------------|------------|-------------------|
| **Planning** | New plan created in `koder/plans/` | Reference plan, set as "Next Up" |
| **Implementation** | Code written, tests added | Summarize changes, note completion % |
| **Bug Fix** | Issue investigated/fixed | Note fix, any remaining issues |
| **Exploration** | Research, reading, discussion | Brief note of what was explored |
| **Release** | Version bumped, deployed | Update version refs, note release |
| **Minimal** | Just Q&A, no artifacts | Sparse update or skip if truly empty |

## Step 3: Update STATE.md

Based on session type, update `koder/STATE.md` appropriately.

### For Planning Sessions

```markdown
## Status
State: CLEAN
[Previous work description]. Plan N drafted, ready for implementation.

## Next Up: Plan N - [Title]
**Plan**: `koder/plans/N_name.md`

[2-4 bullet summary of what the plan covers]

**Key decision**: [Any important constraint or choice made]
```

### For Implementation Sessions

```markdown
## Status
State: IN_PROGRESS (or CLEAN if done)
Implementing Plan N - [progress description]

## Active Work
- [x] Completed item
- [x] Another completed
- [ ] Still pending
- [ ] Also pending

## Recent Changes
- [Concrete list of what was built/changed]
```

### For Exploration/Research Sessions

```markdown
## Status
State: CLEAN
Explored [topic]. No implementation changes.

## Notes from Exploration
- [Key findings or decisions]
- [Links to relevant docs/specs if any]

## Next Steps (if determined)
- [What to do next, or "TBD - discuss with user"]
```

### For Minimal Sessions

If truly nothing substantive happened (just greetings, quick Q&A):
- Don't update STATE.md unnecessarily
- Or add a single line: "Session: brief Q&A, no changes"

## Step 4: Clarify If Needed

If the session outcome is unclear, ASK the user:

```
Before closing, I want to capture the right context for next session.

What should the next session focus on?
- [ ] Implement Plan N
- [ ] Continue exploring [topic]
- [ ] Fix/investigate [issue]
- [ ] Other: ___

Any specific notes to capture?
```

## Step 5: Update Other Docs (Only If Needed)

**CLAUDE.md** - Only if:
- New capabilities added (commands, APIs)
- Version number changed
- Workflow significantly changed

**CHANGELOG.md** - Only if:
- New version was released
- Need to document unreleased changes

## Step 6: Commit and Push

```bash
git add -A
git status --short

# Commit message based on session type:
# Planning:      "docs: Add Plan N (Title)"
# Implementation: "docs: Update STATE.md for [feature] progress"
# Exploration:   "docs: Update STATE.md for session close"
# Release:       "docs: Update docs for vX.Y.Z release"

git commit -m "docs: [appropriate message]

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

git push origin master
```

## Step 7: Session Summary

Output concise summary:

```
## Session Summary

**Version**: vX.Y.Z (local = zyt = source: ✓ aligned)
**zyt.app**: healthy

### This Session
- [What was accomplished - be specific]

### Next Session
- [Clear entry point - usually "Implement Plan N" or specific task]

### Files Changed
- [List only if relevant]
```

## Principles

1. **Adapt to context** - Don't force a template; match the session
2. **Be specific** - Future Claude needs concrete details
3. **Don't over-document** - Empty sessions don't need verbose updates
4. **Ask if unclear** - Better to clarify than guess wrong
5. **Version alignment** - Always verify source = binary = remote
