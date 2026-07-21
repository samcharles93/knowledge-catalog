from __future__ import annotations

import pytest

from okf.bundle.document import OKFDocument, OKFDocumentError


def test_roundtrip_preserves_frontmatter_and_body():
    src = (
        "---\n"
        "type: Service\n"
        "title: User Service\n"
        "description: User authentication and management.\n"
        "tags: [auth, users]\n"
        "timestamp: 2026-07-22T00:00:00+00:00\n"
        "---\n"
        "\n"
        "# User Service\n"
        "\n"
        "Service body details.\n"
    )
    doc = OKFDocument.parse(src)
    assert doc.frontmatter["type"] == "Service"
    assert doc.frontmatter["tags"] == ["auth", "users"]
    assert doc.body.startswith("# User Service")

    serialized = doc.serialize()
    reparsed = OKFDocument.parse(serialized)
    assert reparsed.frontmatter == doc.frontmatter
    assert reparsed.body.strip() == doc.body.strip()


def test_parse_no_frontmatter_treats_all_as_body():
    src = "# Heading\n\nNo frontmatter here.\n"
    doc = OKFDocument.parse(src)
    assert doc.frontmatter == {}
    assert "Heading" in doc.body


def test_unterminated_frontmatter_raises():
    src = "---\ntype: Service\nstill in frontmatter\n"
    with pytest.raises(OKFDocumentError):
        OKFDocument.parse(src)


def test_validate_rejects_missing_required_type():
    doc = OKFDocument(frontmatter={"title": "Missing Type"})
    with pytest.raises(OKFDocumentError) as exc:
        doc.validate()
    assert "type" in str(exc.value)


def test_validate_accepts_valid_document():
    doc = OKFDocument(
        frontmatter={
            "type": "Architecture",
            "title": "System Topology",
            "description": "High level topology.",
        }
    )
    doc.validate()
    assert len(doc.get_warnings()) == 0
