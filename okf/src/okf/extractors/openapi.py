from __future__ import annotations

import datetime
import json
from pathlib import Path
from typing import Any

import yaml

from okf.bundle.document import OKFDocument
from okf.extractors.base import BaseExtractor


class OpenAPIExtractor(BaseExtractor):
    """Parses OpenAPI 3.0 / Swagger specifications into OKF API Endpoint concepts."""

    def __init__(self, spec_path: Path):
        self.spec_path = Path(spec_path).resolve()

    def extract_concepts(self) -> dict[str, OKFDocument]:
        concepts: dict[str, OKFDocument] = {}
        now = datetime.datetime.now(datetime.timezone.utc).isoformat()

        content = self.spec_path.read_text(encoding="utf-8")
        if self.spec_path.suffix in (".yaml", ".yml"):
            data = yaml.safe_load(content)
        else:
            data = json.loads(content)

        info = data.get("info", {})
        title = info.get("title", "API Service")
        version = info.get("version", "1.0.0")

        paths = data.get("paths", {})
        for path_str, path_item in paths.items():
            if not isinstance(path_item, dict):
                continue
            for method in ("get", "post", "put", "delete", "patch"):
                if method not in path_item:
                    continue
                op = path_item[method]
                op_id = op.get("operationId") or f"{method}_{path_str.replace('/', '_')}"
                clean_op_id = "".join(c if c.isalnum() or c in "_-" else "_" for c in op_id)
                concept_id = f"api/{clean_op_id}"

                summary = op.get("summary") or op.get("description") or f"{method.upper()} {path_str}"

                doc = OKFDocument(
                    frontmatter={
                        "type": "API",
                        "title": f"{method.upper()} {path_str}",
                        "description": summary,
                        "resource": f"{path_str}#{method}",
                        "tags": ["api", method.lower()] + op.get("tags", []),
                        "version": version,
                        "timestamp": now,
                    },
                    body=(
                        f"# {method.upper()} {path_str}\n\n"
                        f"{op.get('description', summary)}\n\n"
                        f"## Details\n"
                        f"- **Operation ID**: `{op_id}`\n"
                        f"- **Method**: `{method.upper()}`\n"
                        f"- **Path**: `{path_str}`\n"
                    )
                )
                concepts[concept_id] = doc

        return concepts
