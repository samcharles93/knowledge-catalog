# Open Knowledge Standard (OKF)

> **An open, neutral, and universal knowledge standard for software projects and AI coding agents.**

The **Open Knowledge Format (OKF)** is a tool-agnostic standard for representing codebase architecture, API contracts, data schemas, domain models, and operational guidelines as plain Markdown files with YAML frontmatter.

It requires **zero cloud vendor accounts, zero SaaS lock-in, and zero proprietary SDKs**.

---

## 💡 Why OKF?

As AI coding agents (such as Antigravity, Claude Code, Cursor, Copilot, and Windsurf) become integral to modern software development, maintaining clean, accessible, and structured context is critical.

OKF grounds knowledge management on universal developer primitives:
- 📄 **Plain Text Markdown**: Human-readable and natively understood by LLMs.
- 🏷️ **YAML Frontmatter**: Machine-parseable metadata for querying, indexing, and validation.
- 🌿 **Git Native**: Diffs, pull requests, blame, and code reviews work automatically.
- 🕸️ **Graph-Shaped Hyperlinks**: Interlink concepts to build an explicit relationship graph.
- ⚡ **Progressive Disclosure**: `index.md` directory manifests prevent overflowing LLM context windows.

---

## 🚀 Quickstart

### 1. Installation

```bash
cd okf
pip install -e .
```

### 2. Initialize Knowledge Bundle in Any Project

```bash
okf init --path .okf
```

This creates a standard `.okf/` knowledge directory hierarchy:

```
my-project/
├── .okf/
│   ├── config.yaml            # Bundle config
│   ├── index.md               # Progressive disclosure entry manifest
│   ├── architecture/          # Architecture decisions & topology
│   ├── services/              # Service & module descriptions
│   ├── api/                   # API specifications & contracts
│   ├── database/              # Schema tables & data models
│   └── rules/                 # Coding guidelines & security directives
```

### 3. Harvest Knowledge from Source Code, APIs & Databases

Harvest concepts automatically from local code, OpenAPI specs, SQL DDL files, or web docs:

```bash
# Harvest from codebase files
okf harvest --type codebase --src ./src --out .okf

# Harvest from OpenAPI spec
okf harvest --type openapi --src openapi.yaml --out .okf

# Harvest from SQL DDL schema
okf harvest --type sql --src schema.sql --out .okf
```

### 4. Remember a Concept Directly

Not everything worth keeping comes from a mechanical harvest. Capture a coding rule, session insight, or runbook step as a validated concept document:

```bash
okf remember --type Rule --title "Always run golangci-lint before committing Go changes" \
  --body "Run golangci-lint before every Go commit." --out .okf

# Or pipe a longer body in via stdin:
echo "Harvest pruning is scoped per-namespace; see extract.go." \
  | okf remember --type Concept --title "Harvest pruning scoping" --out .okf
```

`--type` must be `Rule`, `Runbook`, `Concept`, `Service`, `Architecture`, or a custom type of your choosing — `Codebase`, `API`, `Table`, and `Reference` are reserved for harvested concepts and are rejected, since a manually-written file there would be silently deleted by the next matching harvest.

### 5. Validate Knowledge Base Integrity

```bash
okf validate --bundle .okf
```

Checks YAML frontmatter schema compliance, required fields, broken links, and orphan concepts.

### 6. Export Context for AI Coding Agents

```bash
# Output concise summary context for system prompts
okf context --bundle .okf

# Output detailed concept context
okf context --bundle .okf --concept services/user-service
```

### 7. Generate Offline Interactive Visualization

```bash
okf visualize --bundle .okf --out .okf/viz.html
```

Renders a standalone, interactive graph viewer in HTML (`viz.html`).

---

## 🤖 Agent Adapters & Integration

OKF is designed to power every AI coding agent seamlessly:

| Agent | Integration Guide | Protocol |
|---|---|---|
| **Antigravity AI** | [agent-adapters/antigravity.md](agent-adapters/antigravity.md) | MCP Server / System Prompt |
| **Claude Code** | [agent-adapters/claude_code.md](agent-adapters/claude_code.md) | `SessionStart` hook (auto context) + `okf-librarian` Skill |
| **Cursor** | [agent-adapters/cursor_rules.md](agent-adapters/cursor_rules.md) | `.cursorrules` / OKF context |
| **GitHub Copilot** | [agent-adapters/copilot_instructions.md](agent-adapters/copilot_instructions.md) | `.github/copilot-instructions.md` |

### Running the OKF MCP Server

Launch the built-in Model Context Protocol (MCP) server for any coding agent:

```bash
okf mcp --bundle .okf
```

Available MCP Tools:
- `okf_list_concepts`: Returns all concepts and metadata.
- `okf_get_concept`: Returns detailed concept body and links.
- `okf_get_context`: Returns progressive disclosure summary context.
- `okf_search_concepts`: Case-insensitive search across tags, titles, and bodies.
- `okf_validate`: Returns bundle health and validation report.
- `okf_remember`: Writes a new free-form concept (rule, insight, runbook step, etc.) directly, without a harvest.

---

## 📖 Specification

Read the full **[Open Knowledge Format v1.0 Specification (SPEC.md)](SPEC.md)** for complete details on frontmatter schemas, standard concept types, link semantics, and graph rules.

---

## 📂 Sample Bundles

Browse ready-to-use sample bundles under `bundles/`:
- [`bundles/microservices_app/`](bundles/microservices_app/): Complete e-commerce microservices system knowledge bundle.

---

## 📜 License

Distributed under the [Apache 2.0 License](LICENSE.md).
