from __future__ import annotations

from pathlib import Path

from okf.bundle.document import OKFDocument
from okf.viewer.generator import generate_visualization


def test_generate_visualization(tmp_path: Path):
    bundle = tmp_path / "bundle"
    bundle.mkdir()

    doc = OKFDocument(
        frontmatter={"type": "Architecture", "title": "Overview", "description": "Overview concept"},
        body="# Overview\nContent.\n",
    )
    (bundle / "overview.md").write_text(doc.serialize(), encoding="utf-8")

    out_file = tmp_path / "viz.html"
    stats = generate_visualization(bundle, out_file)
    assert stats["concepts"] == 1
    assert out_file.exists()
    assert "Cytoscape" in out_file.read_text(encoding="utf-8") or "graph" in out_file.read_text(encoding="utf-8")
