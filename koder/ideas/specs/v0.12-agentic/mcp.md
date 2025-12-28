# MCP Server

## Summary

Fazt exposes itself as a Model Context Protocol (MCP) server. External AI
agents (like Claude Code) can deploy apps, read logs, query data, and manage
the server through a standardized tool interface.

## What is MCP?

MCP is Anthropic's protocol for AI-tool communication. It provides:
- Standardized tool discovery
- Typed inputs/outputs
- Resource management

## Starting the Server

```bash
fazt mcp start --port 3001

# Or configure in system config
fazt config set mcp.enabled true
fazt config set mcp.port 3001
```

## Exposed Tools

### App Management

```json
{
  "name": "fazt_app_deploy",
  "description": "Deploy files to create or update an app",
  "parameters": {
    "slug": "string",
    "files": [{ "path": "string", "content": "string" }]
  }
}
```

```json
{
  "name": "fazt_app_list",
  "description": "List all deployed apps",
  "parameters": {}
}
```

```json
{
  "name": "fazt_app_delete",
  "description": "Delete an app",
  "parameters": { "uuid": "string" }
}
```

### Storage

```json
{
  "name": "fazt_storage_query",
  "description": "Query document storage",
  "parameters": {
    "app_uuid": "string",
    "collection": "string",
    "query": "object"
  }
}
```

### System

```json
{
  "name": "fazt_system_status",
  "description": "Get system health and metrics",
  "parameters": {}
}
```

```json
{
  "name": "fazt_logs_read",
  "description": "Read app or system logs",
  "parameters": {
    "app_uuid": "string (optional)",
    "lines": "number"
  }
}
```

## Authentication

MCP connections require an API token:

```bash
# Generate token
fazt mcp token create --name "claude-code"

# Token is passed in MCP handshake
```

## Use Case: Claude Code Integration

```bash
# In Claude Code, add Fazt as MCP server
claude config add-mcp fazt --url http://vps:3001 --token xxx
```

Now Claude Code can:
- "Deploy this React app to fazt"
- "Show me the logs for the blog app"
- "What's the server status?"

## Security

- Tokens are scoped: read-only, deploy-only, full-access
- All actions are audit-logged
- Rate limiting per token
