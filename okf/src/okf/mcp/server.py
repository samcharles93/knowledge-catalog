from __future__ import annotations

import asyncio
from pathlib import Path
from typing import Any

from mcp.server import Server
from mcp.server.stdio import stdio_server
import mcp.types as types

from okf.bundle.document import OKFDocument
from okf.bundle.validator import validate_bundle
from okf.context import ContextGenerator

def create_mcp_server(bundle_root: Path) -> Server:
    bundle_root = Path(bundle_root).resolve()
    server = Server("okf-knowledge-server")

    @server.list_tools()
    async def handle_list_tools() -> list[types.Tool]:
        return [
            types.Tool(
                name="okf_list_concepts",
                description="List all knowledge concepts in the OKF bundle with their type, title, and metadata.",
                inputSchema={"type": "object", "properties": {}},
            ),
            types.Tool(
                name="okf_get_concept",
                description="Get full content and metadata for a specific concept ID (e.g., 'architecture/system-overview').",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "concept_id": {"type": "string", "description": "Concept ID relative to bundle root"}
                    },
                    "required": ["concept_id"],
                },
            ),
            types.Tool(
                name="okf_get_context",
                description="Get summary prompt context for progressive disclosure in coding agents.",
                inputSchema={"type": "object", "properties": {}},
            ),
            types.Tool(
                name="okf_validate_bundle",
                description="Validate OKF bundle for structural integrity, frontmatter schema compliance, and link health.",
                inputSchema={"type": "object", "properties": {}},
            ),
        ]

    @server.call_tool()
    async def handle_call_tool(name: str, arguments: dict[str, Any] | None) -> list[types.TextContent]:
        args = arguments or {}

        if name == "okf_list_concepts":
            concepts = []
            for md_path in sorted(bundle_root.rglob("*.md")):
                if md_path.name in ("index.md", "log.md"):
                    continue
                rel_id = "/".join(md_path.relative_to(bundle_root).with_suffix("").parts)
                try:
                    doc = OKFDocument.parse(md_path.read_text(encoding="utf-8"))
                    fm = doc.frontmatter
                    concepts.append(
                        f"- `{rel_id}` [{fm.get('type', 'Concept')}]: {fm.get('title', rel_id)} - {fm.get('description', '')}"
                    )
                except Exception:
                    concepts.append(f"- `{rel_id}` (Parse Error)")

            return [types.TextContent(type="text", text="\n".join(concepts) or "No concepts found.")]

        elif name == "okf_get_concept":
            concept_id = args.get("concept_id", "")
            cg = ContextGenerator(bundle_root)
            content = cg.get_full_concept_context(concept_id)
            return [types.TextContent(type="text", text=content)]

        elif name == "okf_get_context":
            cg = ContextGenerator(bundle_root)
            content = cg.get_summary_context()
            return [types.TextContent(type="text", text=content)]

        elif name == "okf_validate_bundle":
            report = validate_bundle(bundle_root)
            lines = [f"Bundle Valid: {report.is_valid}", f"Total Concepts: {report.total_concepts}"]
            if report.errors:
                lines.append("\nErrors:")
                for e in report.errors:
                    lines.append(f"- [{e.concept_id}] {e.message}")
            if report.warnings:
                lines.append("\nWarnings:")
                for w in report.warnings:
                    lines.append(f"- [{w.concept_id}] {w.message}")
            return [types.TextContent(type="text", text="\n".join(lines))]

        raise ValueError(f"Unknown tool: {name}")

    return server


async def run_mcp_server(bundle_root: Path):
    server = create_mcp_server(bundle_root)
    async with stdio_server() as (read_stream, write_stream):
        await server.run(read_stream, write_stream, server.create_initialization_options())
