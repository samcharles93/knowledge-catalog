from __future__ import annotations

import datetime
import re
from pathlib import Path

from okf.bundle.document import OKFDocument
from okf.extractors.base import BaseExtractor

_CREATE_TABLE_RE = re.compile(r"CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?([`\"\w\.]+)\s*\((.*?)\);", re.IGNORECASE | re.DOTALL)


class SQLExtractor(BaseExtractor):
    """Parses SQL DDL files to extract Table schema concepts."""

    def __init__(self, sql_path: Path):
        self.sql_path = Path(sql_path).resolve()

    def extract_concepts(self) -> dict[str, OKFDocument]:
        concepts: dict[str, OKFDocument] = {}
        now = datetime.datetime.now(datetime.timezone.utc).isoformat()

        content = self.sql_path.read_text(encoding="utf-8")
        matches = _CREATE_TABLE_RE.findall(content)

        for table_name, body in matches:
            clean_name = table_name.strip("`\"").split(".")[-1]
            concept_id = f"database/{clean_name}"

            lines = [line.strip() for line in body.splitlines() if line.strip()]
            col_lines = []
            for line in lines:
                if not line.startswith(("--", "PRIMARY", "CONSTRAINT", "FOREIGN", "KEY", "INDEX", ")")):
                    col_lines.append(f"| `{line.split()[0] if line.split() else ''}` | `{line}` |")

            table_md = (
                f"# Table `{clean_name}`\n\n"
                f"Extracted from SQL DDL `{self.sql_path.name}`.\n\n"
                f"## Columns\n\n"
                f"| Column | Full Definition |\n"
                f"|---|---|\n" + "\n".join(col_lines) + "\n"
            )

            doc = OKFDocument(
                frontmatter={
                    "type": "Table",
                    "title": clean_name,
                    "description": f"Database table {clean_name}.",
                    "resource": str(self.sql_path),
                    "tags": ["database", "table", "sql"],
                    "timestamp": now,
                },
                body=table_md,
            )
            concepts[concept_id] = doc

        return concepts
