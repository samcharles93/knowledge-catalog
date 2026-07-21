---
description: Root architecture and project structure for knowledge-catalog.
resource: /work/knowledge-catalog
tags:
    - overview
    - architecture
    - codebase
timestamp: "2026-07-21T17:53:05Z"
title: knowledge-catalog Overview
type: Architecture
---

# Overview

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

### 4. Validate Knowledge Base Integrity

```bash
okf validate --bundle .okf
```

Checks YAML frontmatter schema compliance, required fields, broken links, and orphan concepts.

### 5. Export Context for AI Coding Agents

```bash
# Output concise summary context for system prompts
okf context --bundle .okf

# Output detailed concept context
okf context --bundle .okf --concept services/user-service
```

### 6. Generate Offline Interactive Visualization

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
- `okf_validate_bundle`: Returns bundle health and validation report.

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


# Codebase Navigation

* [main.go](/codebase/cmd/okf/main.md) - `cmd/okf/main.go`
* [main_test.go](/codebase/cmd/okf/main_test.md) - `cmd/okf/main_test.go`
* [bundle.go](/codebase/okf/bundle.md) - `okf/bundle.go`
* [bundle_test.go](/codebase/okf/bundle_test.md) - `okf/bundle_test.go`
* [context.go](/codebase/okf/context.md) - `okf/context.go`
* [context_test.go](/codebase/okf/context_test.md) - `okf/context_test.go`
* [document.go](/codebase/okf/document.md) - `okf/document.go`
* [document_test.go](/codebase/okf/document_test.md) - `okf/document_test.go`
* [extract.go](/codebase/okf/extract.md) - `okf/extract.go`
* [extract_codebase.go](/codebase/okf/extract_codebase.md) - `okf/extract_codebase.go`
* [extract_codebase_test.go](/codebase/okf/extract_codebase_test.md) - `okf/extract_codebase_test.go`
* [extract_openapi.go](/codebase/okf/extract_openapi.md) - `okf/extract_openapi.go`
* [extract_openapi_test.go](/codebase/okf/extract_openapi_test.md) - `okf/extract_openapi_test.go`
* [extract_sql.go](/codebase/okf/extract_sql.md) - `okf/extract_sql.go`
* [extract_sql_test.go](/codebase/okf/extract_sql_test.md) - `okf/extract_sql_test.go`
* [extract_test.go](/codebase/okf/extract_test.md) - `okf/extract_test.go`
* [extract_web.go](/codebase/okf/extract_web.md) - `okf/extract_web.go`
* [extract_web_test.go](/codebase/okf/extract_web_test.md) - `okf/extract_web_test.go`
* [fetch.go](/codebase/okf/fetch.md) - `okf/fetch.go`
* [fetch_test.go](/codebase/okf/fetch_test.md) - `okf/fetch_test.go`
* [mcp.go](/codebase/okf/mcp.md) - `okf/mcp.go`
* [mcp_test.go](/codebase/okf/mcp_test.md) - `okf/mcp_test.go`
* [paths.go](/codebase/okf/paths.md) - `okf/paths.go`
* [paths_test.go](/codebase/okf/paths_test.md) - `okf/paths_test.go`
* [service.go](/codebase/okf/service.md) - `okf/service.go`
* [service_test.go](/codebase/okf/service_test.md) - `okf/service_test.go`
* [synthesizer.go](/codebase/okf/synthesizer.md) - `okf/synthesizer.go`
* [synthesizer_test.go](/codebase/okf/synthesizer_test.md) - `okf/synthesizer_test.go`
* [viz.js](/codebase/okf/viewer/viz.md) - `okf/viewer/viz.js`
* [viz.go](/codebase/okf/viz.md) - `okf/viz.go`
* [viz_test.go](/codebase/okf/viz_test.md) - `okf/viz_test.go`
