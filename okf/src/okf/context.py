from __future__ import annotations

from pathlib import Path
from typing import Any

from okf.bundle.document import OKFDocument
from okf.bundle.paths import path_to_concept_id


class ContextGenerator:
    """Generates prompt-ready context snippets for coding agents from OKF bundles."""

    def __init__(self, bundle_root: Path):
        self.bundle_root = Path(bundle_root).resolve()

    def get_summary_context(self) -> str:
        """Generates a concise progressive disclosure tree index of the bundle."""
        if not self.bundle_root.exists():
            return "No OKF bundle found."

        index_path = self.bundle_root / "index.md"
        if index_path.exists():
            index_content = index_path.read_text(encoding="utf-8")
        else:
            index_content = "Root index.md not found."

        output = [
            "# Project Knowledge Base (OKF)",
            f"**Bundle Location**: `{self.bundle_root.name}`",
            "",
            "## Progressive Disclosure Index",
            "",
            index_content,
            "",
            "---",
            "*(Agent Note: Use `okf get <concept_id>` or open specific `.md` files under `.okf/` for detailed schemas and architecture details.)*"
        ]
        return "\n".join(output)

    def get_full_concept_context(self, concept_id: str) -> str:
        """Fetch full markdown content of a specific concept."""
        file_path = self.bundle_root / f"{concept_id.lstrip('/')}.md"
        if not file_path.exists():
            return f"Concept '{concept_id}' not found in bundle."

        doc = OKFDocument.parse(file_path.read_text(encoding="utf-8"))
        fm = doc.frontmatter

        header = [
            f"# Concept: {fm.get('title', concept_id)}",
            f"- **Type**: {fm.get('type', 'Unknown')}",
            f"- **Description**: {fm.get('description', 'N/A')}",
        ]
        if fm.get("resource"):
            header.append(f"- **Resource**: `{fm.get('resource')}`")
        if fm.get("tags"):
            header.append(f"- **Tags**: {', '.join(fm.get('tags'))}")

        return "\n".join(header) + "\n\n" + doc.body
