from __future__ import annotations

from pathlib import Path

from okf.bundle.document import OKFDocument
from okf.bundle.validator import validate_bundle


def test_validator_detects_missing_type_and_broken_links(tmp_path: Path):
    bundle = tmp_path / "bundle"
    bundle.mkdir()

    # Valid doc
    doc1 = OKFDocument(
        frontmatter={"type": "Architecture", "title": "Overview", "description": "System overview"},
        body="Links to [Broken Target](/services/nonexistent.md).",
    )
    (bundle / "overview.md").write_text(doc1.serialize(), encoding="utf-8")

    # Invalid doc (missing type)
    doc2 = OKFDocument(
        frontmatter={"title": "Invalid"},
        body="Missing type frontmatter.",
    )
    (bundle / "invalid.md").write_text(doc2.serialize(), encoding="utf-8")

    report = validate_bundle(bundle)
    assert not report.is_valid
    assert any("type" in e.message for e in report.errors)
    assert any("Broken link target" in w.message for w in report.warnings)
