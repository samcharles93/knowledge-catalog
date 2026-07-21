from __future__ import annotations

import datetime
from pathlib import Path
from typing import Any

from okf.bundle.document import OKFDocument
from okf.extractors.base import BaseExtractor

_IGNORE_DIRS = {
    ".git", ".venv", "node_modules", "__pycache__", ".pytest_cache",
    "dist", "build", ".okf", "target", "vendor"
}


class CodebaseExtractor(BaseExtractor):
    """Scans local source code repositories to generate codebase knowledge concepts."""

    def __init__(self, project_root: Path):
        self.project_root = Path(project_root).resolve()

    def extract_concepts(self) -> dict[str, OKFDocument]:
        concepts: dict[str, OKFDocument] = {}
        now = datetime.datetime.now(datetime.timezone.utc).isoformat()

        # 1. Project Overview Concept
        readme_path = self.project_root / "README.md"
        readme_content = readme_path.read_text(encoding="utf-8") if readme_path.exists() else ""
        
        overview_doc = OKFDocument(
            frontmatter={
                "type": "Architecture",
                "title": f"{self.project_root.name} Overview",
                "description": f"Root architecture and project structure for {self.project_root.name}.",
                "resource": str(self.project_root),
                "tags": ["overview", "architecture", "codebase"],
                "timestamp": now,
            },
            body=f"# Overview\n\n{readme_content}\n\n# Codebase Navigation\n\n"
        )

        # 2. Walk codebase directories & source files
        module_links = []
        for path in sorted(self.project_root.rglob("*")):
            if any(part in _IGNORE_DIRS for part in path.parts):
                continue
            if path.is_file() and path.suffix in (".py", ".ts", ".js", ".go", ".rs", ".java", ".cpp", ".h"):
                rel_path = path.relative_to(self.project_root)
                concept_id = f"codebase/{rel_path.with_suffix('')}"
                
                content = path.read_text(encoding="utf-8", errors="replace")
                lines = content.splitlines()
                first_lines = "\n".join(lines[:30])

                doc = OKFDocument(
                    frontmatter={
                        "type": "Module",
                        "title": path.name,
                        "description": f"Source module {rel_path} ({len(lines)} lines).",
                        "resource": str(rel_path),
                        "tags": [path.suffix.lstrip("."), "source"],
                        "timestamp": now,
                    },
                    body=(
                        f"# Module {path.name}\n\n"
                        f"**Path**: `{rel_path}`  \n"
                        f"**Lines**: {len(lines)}\n\n"
                        f"## Snippet Preview\n\n```\n{first_lines}\n```\n"
                    )
                )
                concepts[concept_id] = doc
                module_links.append(f"* [{path.name}](/{concept_id}.md) - `{rel_path}`")

        overview_doc.body += "\n".join(module_links[:50])
        concepts["architecture/overview"] = overview_doc

        return concepts
