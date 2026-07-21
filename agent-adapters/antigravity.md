# OKF Integration for Antigravity AI

This guide explains how to integrate the **Open Knowledge Format (OKF)** into **Antigravity AI**.

## 1. Directory Structure

Add `.okf/` to your project repository root:

```
my-project/
├── .okf/
│   ├── index.md               # Progressive disclosure manifest
│   ├── architecture/          # Architecture decisions & topology
│   ├── services/              # Microservice & component docs
│   ├── api/                   # OpenAPI contracts & endpoints
│   ├── database/              # Schema tables & models
│   └── rules/                 # Coding guidelines & constraints
```

## 2. Antigravity Agent Directives

Include the following prompt instructions in your system prompt or project `.gemini/` rules:

```markdown
### Project Knowledge (OKF Standard)
This project uses the Open Knowledge Format (OKF) standard located in `.okf/`.

1. **Initialization**: At start of a task, inspect `.okf/index.md` to learn architecture context, services, and coding rules.
2. **Rule Adherence**: Strictly follow guidelines in `.okf/rules/*.md`.
3. **Knowledge Evolution**: When adding new services, tables, or major APIs, create or update the corresponding OKF concept markdown document under `.okf/`.
```

## 3. MCP Integration

Configure the OKF MCP server in your Antigravity MCP settings:

```json
{
  "mcpServers": {
    "okf": {
      "command": "okf",
      "args": ["mcp", "--bundle", ".okf"]
    }
  }
}
```

The agent can now invoke tools like `okf_get_context`, `okf_list_concepts`, `okf_get_concept`, and `okf_validate_bundle`.
