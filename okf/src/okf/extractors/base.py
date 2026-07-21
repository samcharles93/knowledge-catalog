from __future__ import annotations

from abc import ABC, abstractmethod
from pathlib import Path

from okf.bundle.document import OKFDocument


class BaseExtractor(ABC):
    """Abstract base class for vendor-neutral knowledge extractors."""

    @abstractmethod
    def extract_concepts(self) -> dict[str, OKFDocument]:
        """Extract concepts mapped by concept_id -> OKFDocument."""
        pass

    def export_bundle(self, bundle_root: Path) -> int:
        """Write extracted concept documents to bundle_root and regenerate indexes."""
        bundle_root = Path(bundle_root)
        bundle_root.mkdir(parents=True, exist_ok=True)

        concepts = self.extract_concepts()
        written = 0

        for concept_id, doc in concepts.items():
            file_path = bundle_root / f"{concept_id}.md"
            file_path.parent.mkdir(parents=True, exist_ok=True)
            file_path.write_text(doc.serialize(), encoding="utf-8")
            written += 1

        from okf.bundle.index import regenerate_indexes
        regenerate_indexes(bundle_root)

        return written
