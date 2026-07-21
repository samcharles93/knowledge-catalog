from __future__ import annotations

import argparse
import asyncio
import logging
import sys
from pathlib import Path

from okf.bundle.document import OKFDocument
from okf.bundle.index import regenerate_indexes
from okf.bundle.validator import validate_bundle
from okf.context import ContextGenerator
from okf.viewer.generator import generate_visualization


def _init_bundle(path: Path) -> None:
    path = Path(path).resolve()
    path.mkdir(parents=True, exist_ok=True)

    # Create directory structure
    (path / "architecture").mkdir(exist_ok=True)
    (path / "services").mkdir(exist_ok=True)
    (path / "rules").mkdir(exist_ok=True)

    # Create root config
    config_path = path / "config.yaml"
    if not config_path.exists():
        config_path.write_text("name: project-knowledge\nversion: '1.0'\n", encoding="utf-8")

    # Create sample concept doc
    overview_path = path / "architecture" / "system-overview.md"
    if not overview_path.exists():
        doc = OKFDocument(
            frontmatter={
                "type": "Architecture",
                "title": "System Overview",
                "description": "High level system components and architecture principles.",
                "tags": ["overview", "architecture"],
            },
            body="# System Overview\n\nWelcome to the project OKF Knowledge Base.\n"
        )
        overview_path.write_text(doc.serialize(), encoding="utf-8")

    # Regenerate index.md
    regenerate_indexes(path)
    print(f"Initialized OKF Knowledge Bundle at: {path}", file=sys.stderr)


def _parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(prog="okf", description="Open Knowledge Format (OKF) Toolkit")
    sub = p.add_subparsers(dest="command", required=True)

    # init
    init_cmd = sub.add_parser("init", help="Initialize an OKF knowledge bundle in a project.")
    init_cmd.add_argument("--path", type=Path, default=Path(".okf"), help="Target bundle directory (default: .okf)")

    # validate
    val_cmd = sub.add_parser("validate", help="Validate an OKF bundle structure, schemas, and links.")
    val_cmd.add_argument("--bundle", type=Path, default=Path(".okf"), help="Path to bundle root.")

    # harvest
    harv_cmd = sub.add_parser("harvest", help="Extract concepts from neutral sources (codebase, openapi, sql, web).")
    harv_cmd.add_argument("--type", choices=["codebase", "openapi", "sql", "web"], required=True)
    harv_cmd.add_argument("--src", type=Path, help="Source file or directory path.")
    harv_cmd.add_argument("--url", action="append", help="URL for web harvester (repeatable).")
    harv_cmd.add_argument("--out", type=Path, default=Path(".okf"), help="Output bundle root directory.")

    # context
    ctx_cmd = sub.add_parser("context", help="Export agent prompt context snippets.")
    ctx_cmd.add_argument("--bundle", type=Path, default=Path(".okf"), help="Path to bundle root.")
    ctx_cmd.add_argument("--concept", type=str, help="Specific concept ID to fetch.")

    # visualize
    viz_cmd = sub.add_parser("visualize", help="Generate interactive standalone graph HTML.")
    viz_cmd.add_argument("--bundle", type=Path, default=Path(".okf"), help="Path to bundle root.")
    viz_cmd.add_argument("--out", type=Path, default=None, help="Output HTML path (default: <bundle>/viz.html)")
    viz_cmd.add_argument("--name", type=str, default=None, help="Bundle display name.")

    # mcp
    mcp_cmd = sub.add_parser("mcp", help="Run Model Context Protocol (MCP) server for coding agents.")
    mcp_cmd.add_argument("--bundle", type=Path, default=Path(".okf"), help="Path to bundle root.")

    return p


def main(argv: list[str] | None = None) -> int:
    args = _parser().parse_args(argv)

    if args.command == "init":
        _init_bundle(args.path)
        return 0

    elif args.command == "validate":
        report = validate_bundle(args.bundle)
        print(f"Validation Report for {args.bundle}:")
        print(f"- Total Concepts: {report.total_concepts}")
        print(f"- Valid: {report.is_valid}")
        if report.errors:
            print("\nErrors:")
            for e in report.errors:
                print(f"  ❌ [{e.concept_id}] {e.message}")
        if report.warnings:
            print("\nWarnings:")
            for w in report.warnings:
                print(f"  ⚠️  [{w.concept_id}] {w.message}")
        return 0 if report.is_valid else 1

    elif args.command == "harvest":
        if args.type == "codebase":
            from okf.extractors.codebase import CodebaseExtractor
            ext = CodebaseExtractor(args.src or Path("."))
        elif args.type == "openapi":
            from okf.extractors.openapi import OpenAPIExtractor
            ext = OpenAPIExtractor(args.src)
        elif args.type == "sql":
            from okf.extractors.sql import SQLExtractor
            ext = SQLExtractor(args.src)
        elif args.type == "web":
            from okf.extractors.web import WebExtractor
            ext = WebExtractor(args.url or [])

        n = ext.export_bundle(args.out)
        print(f"Harvested {n} concepts into {args.out}", file=sys.stderr)
        return 0

    elif args.command == "context":
        cg = ContextGenerator(args.bundle)
        if args.concept:
            print(cg.get_full_concept_context(args.concept))
        else:
            print(cg.get_summary_context())
        return 0

    elif args.command == "visualize":
        out = args.out or (args.bundle / "viz.html")
        stats = generate_visualization(args.bundle, out, bundle_name=args.name)
        print(
            f"Wrote {stats['concepts']} concept(s), "
            f"{stats['edges']} edge(s) → {out}",
            file=sys.stderr,
        )
        return 0

    elif args.command == "mcp":
        from okf.mcp.server import run_mcp_server
        asyncio.run(run_mcp_server(args.bundle))
        return 0

    return 1


if __name__ == "__main__":
    sys.exit(main())
