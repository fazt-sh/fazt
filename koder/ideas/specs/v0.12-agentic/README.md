# v0.12 - Agentic

**Theme**: AI-native platform capabilities.

## Summary

v0.12 makes Fazt a first-class platform for AI agents. Apps can call LLMs,
the kernel exposes itself via MCP (Model Context Protocol), and "harness"
apps can evolve their own code.

## Goals

1. **AI Shim**: Unified `fazt.ai` API for LLM providers
2. **MCP Server**: Expose Fazt to external AI agents
3. **Harness Apps**: Self-modifying agentic applications
4. **Git Integration**: VFS versioning for safe evolution

## Documents

- `ai-shim.md` - The `fazt.ai` namespace
- `mcp.md` - Model Context Protocol server
- `harness.md` - Self-evolving apps

## Dependencies

- v0.10 (Runtime): Serverless execution
- v0.11 (Distribution): App manifest for permissions
