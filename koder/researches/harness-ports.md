# Specification: Porting Claude Code CLI Artifacts to OpenAI Codex

**Version:** 1.0
**Date:** 2026-02-06
**Author:** Manus AI

## 1. Introduction

This document provides a comprehensive, self-contained specification for porting existing `claude code cli` artifacts—specifically `CLAUDE.md` configuration and custom skills—to the equivalent formats used by `openai codex`. The instructions herein are designed to be executed by an AI agent, such as Claude Code CLI itself, to automate the migration process.

The migration involves two primary tasks:

1.  **Configuration Porting**: Translating the guidance and settings from a `CLAUDE.md` file into Codex's `AGENTS.md` for instructions and `config.toml` for settings.
2.  **Skill Porting**: Adapting the structure and metadata of Claude skills to conform to the OpenAI Codex skill specification.

---

## 2. Part 1: Porting Configuration (`CLAUDE.md`)

The `CLAUDE.md` file serves a dual purpose: providing instructions to the agent and defining certain configurations. In Codex, this responsibility is split between two distinct files: `AGENTS.md` for instructions and `config.toml` for configuration.

### 2.1. Conceptual Mapping

| Source (Claude) | Destination (Codex) | Purpose |
| :--- | :--- | :--- |
| Instructional Content in `CLAUDE.md` | `AGENTS.md` | Provides natural language guidance, context, and rules for the agent's behavior within a project. |
| Configuration-like settings in `CLAUDE.md` | `config.toml` | Defines structured settings, model selection, tool configurations, and other operational parameters. |

### 2.2. `AGENTS.md` Specification

`AGENTS.md` is the direct equivalent of the instructional portion of `CLAUDE.md`. It is a plain Markdown file used to provide context and guidance to Codex.

#### File Locations and Precedence

Codex builds an instruction chain by concatenating `AGENTS.md` files from different locations. This allows for a layered guidance system, from global defaults to project-specific rules.

1.  **Global Scope**: `~/.codex/AGENTS.md` (or `AGENTS.override.md`)
    - This file provides universal instructions for all projects.
2.  **Project Scope**: `.`, `./<dir>`, `./<dir>/<subdir>`, etc.
    - Codex walks from the project root to the current working directory.
    - In each directory, it looks for `AGENTS.override.md`, then `AGENTS.md`.
    - At most one file per directory is included.
3.  **Merge Order**: Files are concatenated from the root down. Content in files closer to the current directory appears later in the prompt and thus takes higher precedence.

#### Porting Action

- **Extract** all natural language instructions, rules, and contextual information from `CLAUDE.md`.
- **Create** a new `AGENTS.md` file in the corresponding location (e.g., `~/.claude/CLAUDE.md` -> `~/.codex/AGENTS.md`, or `.claude/CLAUDE.md` -> `AGENTS.md` in the project root).
- **Write** the extracted instructional content into the new `AGENTS.md` file.

### 2.3. `config.toml` Specification

`config.toml` is a TOML file for defining structured configuration settings. It has no direct equivalent in the Claude system, so its creation requires mapping conceptual settings to explicit configuration keys.

#### File Locations

- **User-level**: `~/.codex/config.toml` (global settings)
- **Project-level**: `.codex/config.toml` (project-specific overrides)

#### Key Configuration Mappings

While `CLAUDE.md` does not have explicit key-value settings, any instructions that imply a configuration should be mapped to `config.toml`.

| Implied Claude Setting | Codex `config.toml` Key | Example Value |
| :--- | :--- | :--- |
| "Use the gpt-4 model" | `model` | `"gpt-4.1-mini"` |
| "Always ask before running commands" | `approval_policy` | `"on-request"` |
| "You can access my local filesystem" | `sandbox_mode` | `"workspace-write"` |
| "Use the `my-api` tool" | `[mcp_servers.my-api]` | (See section 2.3.1) |

#### 2.3.1. Example `config.toml` Structure

```toml
#:schema https://developers.openai.com/codex/config-schema.json

# 1. Core Model Selection
# The primary model to be used by Codex.
model = "gpt-5.2-codex"

# 2. Approval & Sandbox Policy
# Defines when to ask for approval and the filesystem access level.
approval_policy = "on-request"  # Other values: "untrusted", "on-failure", "never"
sandbox_mode = "read-only"      # Other values: "workspace-write", "danger-full-access"

# 3. AGENTS.md Configuration
# Controls how AGENTS.md files are processed.
project_doc_max_bytes = 32768
project_doc_fallback_filenames = ["TEAM_GUIDE.md"]

# 4. Model Context Protocol (MCP) Servers
# Equivalent to defining tools or services the agent can use.
[mcp_servers.my-api]
enabled = true
command = "my-api-server-executable"

# 5. Skills Configuration
# Enable or disable specific skills.
[[skills.config]]
path = "~/.codex/skills/my-ported-skill"
enabled = true
```

#### Porting Action

1.  **Analyze** `CLAUDE.md` for any statements that imply a configuration setting.
2.  **Create** a `config.toml` file at `~/.codex/config.toml`.
3.  **Translate** the implied settings into the appropriate TOML key-value pairs as demonstrated above.

---

## 3. Part 2: Porting Skills

Codex skills are structurally similar to Claude skills but have a more formalized directory structure and metadata format.

### 3.1. Codex Skill Anatomy

A Codex skill is a directory containing a `SKILL.md` file and optional subdirectories for resources.

```
my-ported-skill/                    # Skill root directory
├── SKILL.md                        # Required: Main skill definition
├── agents/
│   └── openai.yaml                 # Recommended: UI metadata
├── scripts/
│   └── helper_script.py            # Optional: Executable scripts
├── references/
│   └── api_docs.md                 # Optional: Reference documentation
└── assets/
    └── template.html               # Optional: Static assets for output
```

### 3.2. `SKILL.md` Specification

This is the core file of a skill. It contains YAML front matter for metadata and a Markdown body for instructions.

#### YAML Front Matter (Required)

The front matter is critical, as its `name` and `description` fields are what Codex reads to decide when to use the skill.

```yaml
---
name: my-ported-skill
description: A detailed, comprehensive description of what this skill does, its primary purpose, and the specific situations in which it should be activated.
---
```

#### Markdown Body

The body contains the instructions for the agent, loaded *after* the skill is triggered. It should follow the **progressive disclosure** principle: keep the main `SKILL.md` concise and link to more detailed information in the `references/` directory.

### 3.3. Porting Action: Step-by-Step

For each skill located in `~/.claude/skills/`:

1.  **Create Skill Directory**: Create a new directory at `~/.codex/skills/<skill-name>/`.

2.  **Migrate `SKILL.md`**:
    a.  Copy the existing `SKILL.md` from the Claude skill directory to the new Codex skill directory.
    b.  **Add YAML Front Matter**: Prepend the `SKILL.md` file with the required `name` and `description` fields, enclosed in `---`.
    c.  **Refine Description**: Ensure the `description` is clear and comprehensive for accurate triggering.

3.  **Organize Resources**: Move any auxiliary files from the Claude skill directory into the corresponding Codex subdirectories:
    -   **Executable scripts** -> `scripts/`
    -   **Reference documents** (e.g., API guides, long text files) -> `references/`
    -   **Static files** to be used in the output (e.g., templates, images) -> `assets/`

4.  **Update `SKILL.md` Body**: Review the Markdown body and update any file paths to reflect the new `scripts/`, `references/`, or `assets/` structure.

### 3.4. Side-by-Side Example

#### Claude Skill (`~/.claude/skills/run-linter/`)

```
run-linter/
├── SKILL.md
└── run_linter.py
```

**`SKILL.md` (Claude)**
```markdown
This skill runs the project linter.

To run the linter, execute the `run_linter.py` script.
```

#### Ported Codex Skill (`~/.codex/skills/run-linter/`)

```
run-linter/
├── SKILL.md
└── scripts/
    └── run_linter.py
```

**`SKILL.md` (Codex)**
```markdown
---
name: run-linter
description: Runs the configured linter on the current project to check for code quality and style issues. Use this skill when asked to lint the code, check for style errors, or run a quality check.
---

## Linter Execution Workflow

To run the project linter, execute the `scripts/run_linter.py` script using the `shell` tool.

Example:

`$ python3 scripts/run_linter.py`
```

---

## 4. Automated Porting Workflow

This section outlines a high-level algorithm for an agent to perform the migration.

**BEGIN WORKFLOW**

1.  **Initialize**: Ensure destination directories `~/.codex/` and `~/.codex/skills/` exist.

2.  **Process Global `CLAUDE.md`**:
    a.  Check for the existence of `~/.claude/CLAUDE.md`.
    b.  If it exists, read its content.
    c.  **Extract instructional text** (heuristics: identify headings, paragraphs, lists).
    d.  Write extracted instructions to `~/.codex/AGENTS.md`.
    e.  **Analyze for configuration hints** (e.g., model names, tool references).
    f.  Generate a `~/.codex/config.toml` file with corresponding settings.

3.  **Process Project-level `CLAUDE.md`**:
    a.  Recursively search the current project for `.claude/CLAUDE.md` files.
    b.  For each one found, extract its instructional content.
    c.  Write the content to a new `AGENTS.md` file in the same directory as the `.claude` folder (i.e., at the same level).

4.  **Process Global Skills**:
    a.  List all subdirectories in `~/.claude/skills/`.
    b.  For each skill directory:
        i.    Perform the skill porting actions described in **Section 3.3**.
        ii.   Create the new skill directory under `~/.codex/skills/`.
        iii.  Migrate and update `SKILL.md` with YAML front matter.
        iv.   Relocate resource files to `scripts/`, `references/`, and `assets/`.
        v.    Update paths within the `SKILL.md` body.

5.  **Process Project-level Skills**:
    a.  Recursively search the current project for `.claude/skills/` directories.
    b.  For each skill found, perform the same porting actions as for global skills, creating the new skill under a `.codex/skills/` directory at the same project level.

6.  **Finalization**: Report completion and list all created/modified files.

**END WORKFLOW**

---

## 5. References

[1] OpenAI Developers. (2026). *Custom instructions with AGENTS.md*. [https://developers.openai.com/codex/guides/agents-md/](https://developers.openai.com/codex/guides/agents-md/)
[2] OpenAI Developers. (2026). *Configuration Reference*. [https://developers.openai.com/codex/config-reference/](https://developers.openai.com/codex/config-reference/)
[3] OpenAI Developers. (2026). *Sample Configuration*. [https://developers.openai.com/codex/config-sample/](https://developers.openai.com/codex/config-sample/)
[4] GitHub. (2026). *openai/skills repository*. [https://github.com/openai/skills](https://github.com/openai/skills)
