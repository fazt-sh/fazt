# v0.12 - Agentic

**Theme**: AI-native platform capabilities.

## Summary

v0.12 makes Fazt a first-class platform for AI agents. Apps can call LLMs,
the kernel exposes itself via MCP (Model Context Protocol), and "harness"
apps can evolve their own code.

## Goals

1. **AI Namespace**: `fazt ai` CLI for all AI-related features
2. **AI Shim**: Unified `fazt.ai` API for LLM providers
3. **MCP Server**: Expose Fazt to external AI agents
4. **Skill Management**: Install/update Claude skills from fazt
5. **Harness Apps**: Self-modifying agentic applications
6. **Git Integration**: VFS versioning for safe evolution

## Documents

- `ai-shim.md` - The `fazt.ai` namespace
- `mcp.md` - Model Context Protocol server
- `harness.md` - Self-evolving apps
- `skill.md` - Claude skill management (`fazt ai skill`)

## Dependencies

- v0.10 (Runtime): Serverless execution
- v0.11 (Distribution): App manifest for permissions
