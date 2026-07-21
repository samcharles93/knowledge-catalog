from __future__ import annotations

import re
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

from okf.bundle.document import OKFDocument, OKFDocumentError
from okf.bundle.paths import path_to_concept_id

_LINK_RE = re.compile(r"\[([^\]]+)\]\(([^)]+)\)")
_RESERVED_NAMES = {"index.md", "log.md", "config.yaml"}


@dataclass
class ValidationError:
    concept_id: str
    file_path: Path
    message: str
    is_warning: bool = False


@dataclass
class ValidationReport:
    total_concepts: int = 0
    errors: list[ValidationError] = field(default_factory=list)
    warnings: list[ValidationError] = field(default_factory=list)

    @property
    def is_valid(self) -> bool:
        return len(self.errors) == 0


def validate_bundle(bundle_root: Path) -> ValidationReport:
    bundle_root = Path(bundle_root)
    report = ValidationReport()

    if not bundle_root.exists() or not bundle_root.is_dir():
        report.errors.append(
            ValidationError(
                concept_id="",
                file_path=bundle_root,
                message=f"Bundle directory does not exist: {bundle_root}",
            )
        )
        return report

    concept_files: dict[str, Path] = {}
    documents: dict[str, OKFDocument] = {}
    outgoing_links: dict[str, list[str]] = {}
    incoming_links: dict[str, list[str]] = defaultdict(list)

    # 1. Discover all concept files
    for md_path in bundle_root.rglob("*.md"):
        if md_path.name in _RESERVED_NAMES:
            continue
        rel_parts = path_to_concept_id(bundle_root, md_path)
        concept_id = "/".join(rel_parts)
        concept_files[concept_id] = md_path

    report.total_concepts = len(concept_files)

    # 2. Parse & Validate individual documents
    for concept_id, path in concept_files.items():
        try:
            content = path.read_text(encoding="utf-8")
            doc = OKFDocument.parse(content)
            doc.validate()  # Required frontmatter check (e.g. type)

            for warn_msg in doc.get_warnings():
                report.warnings.append(
                    ValidationError(
                        concept_id=concept_id,
                        file_path=path,
                        message=warn_msg,
                        is_warning=True,
                    )
                )

            documents[concept_id] = doc
        except OKFDocumentError as e:
            report.errors.append(
                ValidationError(
                    concept_id=concept_id, file_path=path, message=str(e)
                )
            )
        except Exception as e:
            report.errors.append(
                ValidationError(
                    concept_id=concept_id,
                    file_path=path,
                    message=f"Failed to read/parse document: {e}",
                )
            )

    # 3. Check link integrity
    for concept_id, doc in documents.items():
        links: list[str] = []
        for match in _LINK_RE.finditer(doc.body):
            target = match.group(2).strip()
            if target.startswith(("http://", "https://", "mailto:", "#")):
                continue
            
            # Resolve target path relative to bundle root or document
            if target.startswith("/"):
                # Bundle-absolute link
                target_clean = target.lstrip("/").removesuffix(".md")
            else:
                # Relative link
                doc_dir = concept_files[concept_id].parent
                resolved = (doc_dir / target).resolve()
                try:
                    rel = resolved.relative_to(bundle_root).with_suffix("")
                    target_clean = str(rel)
                except ValueError:
                    target_clean = target

            links.append(target_clean)
            incoming_links[target_clean].append(concept_id)

            # Check if target concept exists
            if target_clean not in concept_files and target_clean != "index":
                report.warnings.append(
                    ValidationError(
                        concept_id=concept_id,
                        file_path=concept_files[concept_id],
                        message=f"Broken link target: '{target}' (resolved as '{target_clean}')",
                        is_warning=True,
                    )
                )
        outgoing_links[concept_id] = links

    # 4. Check for orphan concepts (no incoming and no outgoing links, excluding root concept)
    if report.total_concepts > 1:
        for concept_id in concept_files:
            if not outgoing_links.get(concept_id) and not incoming_links.get(concept_id):
                report.warnings.append(
                    ValidationError(
                        concept_id=concept_id,
                        file_path=concept_files[concept_id],
                        message="Orphan concept: has no incoming or outgoing links in bundle graph",
                        is_warning=True,
                    )
                )

    return report


from collections import defaultdict
