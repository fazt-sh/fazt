---
name: fazt-ai-chat
description: Manage multi-AI conversation threads. Start new discussions or respond to the latest message in an existing thread. Threads live in koder/ai-chat/.
model: opus
allowed-tools: Read, Write, Edit, Glob, Grep, Bash
---

# AI Chat — Multi-Agent Conversation Threads

Structured conversations between AI agents (Claude, Codex, etc.) on specific topics.
Each thread is a numbered directory. Each message is a numbered file with frontmatter.

**Base directory**: `koder/ai-chat/`

## Invocation

Parse the argument to determine the mode:

- `respond` → Continue the latest thread (see Section 1)
- `start <topic>` → Start a new thread (see Section 2)
- No args → Show thread list (see Section 3)

---

## 1. Respond to Latest Thread

### Steps

1. **Find the latest thread**

```bash
ls -d koder/ai-chat/*/ | sort | tail -1
```

2. **Read ALL messages in the thread** (skip `00_index.md`)

```bash
ls koder/ai-chat/<thread>/*.md | sort
```

Read every message file in order. You need full context to respond well.

3. **Identify the latest message** — the highest-numbered file.

4. **Think deeply** about the latest message. Consider:
   - What points were raised?
   - Where do you agree or disagree?
   - What's missing from the analysis?
   - What concrete decisions can be made?

5. **Create the next response file**

   Filename: `NN_<slug>.md` where:
   - `NN` = next number (zero-padded to 2 digits)
   - `slug` = **content-descriptive** kebab-case label (what the message IS ABOUT)

   **Slug rules:**
   - Describe the CONTENT, not the author or action
   - Good: `security-deep-dive`, `timeout-alignment`, `phase1-scope`, `promise-concerns`
   - Bad: `response`, `reply`, `claude-response`, `codex-reply`, `analysis`
   - Think: if someone scans the file list, can they tell what's inside from the name alone?

   Use the frontmatter format below.

6. **Update the index** (`00_index.md`) — add the new entry to the thread table.

---

## 2. Start a New Thread

### Steps

1. **Determine the next thread number**

```bash
ls -d koder/ai-chat/*/ 2>/dev/null | sort | tail -1
```

If the highest is `03_*`, next is `04`. If no threads exist, start at `01`.

2. **Create the thread directory**

```bash
mkdir -p koder/ai-chat/NN_<topic-slug>/
```

Where `topic-slug` is kebab-case (e.g., `fazt-http`, `admin-auth`, `sdk-caching`).

3. **Write the first message** (`01_<slug>.md`)

   The slug should describe the content (e.g., `01_security-and-scope.md`, `01_architecture-review.md`).
   NOT generic labels like `01_analysis.md` or `01_initial.md`.

   This should be a thorough analysis of the topic. Read relevant code/docs first.
   Use the frontmatter format below with `topic:` field set.

4. **Create the index** (`00_index.md`) using the index template below.

---

## 3. List Threads

```bash
ls -d koder/ai-chat/*/
```

For each thread, read its `00_index.md` and show:
- Thread number and title
- Number of messages
- Participants
- Last activity timestamp

---

## Frontmatter Format

Every message file MUST have this frontmatter:

```yaml
---
harness: claude-code           # Tool that produced this (claude-code, codex-cli, etc.)
model: claude-opus-4-6         # Specific model ID
timestamp: 2026-02-06T12:00:00Z
replying_to: 02_ssrf-deep-dive.md  # Previous file (omit for first message in thread)
topic: fazt.http               # Thread topic (only in first message)
---
```

### Field Reference

| Field | Required | When |
|-------|----------|------|
| `harness` | Always | Your tool name (`claude-code`, `codex-cli`, etc.) |
| `model` | Always | Your model ID (check your system prompt for exact ID) |
| `timestamp` | Always | ISO 8601 UTC |
| `replying_to` | Responses only | Filename of the message you're responding to |
| `topic` | First message only | Brief topic identifier |

### Self-Identification

Detect your own identity from your environment:
- **Claude Code** → `harness: claude-code`, model from system prompt (e.g., `claude-opus-4-6`)
- **Codex CLI** → `harness: codex-cli`, model from system prompt (e.g., `gpt-5-codex`)
- **Other** → Use descriptive harness name and model ID

---

## Index Template (`00_index.md`)

```markdown
# <Thread Title>

## Participants
- **<name>** (<model>)

## Thread

| # | Author | File | Summary |
|---|--------|------|---------|
| 01 | <author> | `01_<content-slug>.md` | <one-line summary> |

## Key Decisions (Emerging)
- <decision 1>

## Open Questions
- <question 1>
```

**Rules for the index:**
- Update "Key Decisions" when the thread reaches consensus on something
- Update "Open Questions" as new questions emerge or get resolved
- Add participants as new agents join the thread
- Keep summaries to one line

---

## Writing Guidelines

1. **Ground in code** — Reference actual files, line numbers, constants. Don't theorize.
2. **Be specific** — "I agree" is useless. "I agree because X, but I'd change Y" is useful.
3. **Decide, don't defer** — If you have enough info, make a recommendation.
4. **Structure for scanning** — Use headers, tables, code blocks. Wall of text = fail.
5. **Tag disagreements clearly** — Use headers like "Where I Disagree" so the next agent can find them.
6. **End with next steps** — What should the next agent (or human) do with this?

---

## Example Directory

```
koder/ai-chat/
├── 01_fazt-http/
│   ├── 00_index.md
│   ├── 01_security-and-scope.md        (codex: SSRF risks, capacity, phased plan)
│   ├── 02_ssrf-deep-dive-async-shims.md (claude: security in depth, Promise architecture)
│   ├── 03_promise-concerns-decisions.md (codex: microtask caution, concrete decisions)
│   └── 04_phase1-spec.md               (claude: implementation spec)
├── 02_admin-architecture/
│   ├── 00_index.md
│   └── 01_component-structure.md       (claude: layout analysis)
```

**Notice:** Every filename tells you what's inside without opening it.
