from __future__ import annotations

from pathlib import Path

from okf.bundle.document import OKFDocument
from okf.bundle.index import regenerate_indexes


def test_regenerate_indexes_creates_index(tmp_path: Path):
    bundle = tmp_path / "bundle"
    services = bundle / "services"
    services.mkdir(parents=True)

    doc = OKFDocument(
        frontmatter={
            "type": "Service",
            "title": "Auth Service",
            "description": "Handles authentication.",
        },
        body="# Auth Service\n",
    )
    (services / "auth.md").write_text(doc.serialize(), encoding="utf-8")

    written = regenerate_indexes(bundle)
    assert len(written) >= 1
    root_index = (bundle / "index.md").read_text(encoding="utf-8")
    assert "Subdirectories" in root_index or "Auth Service" in root_index
