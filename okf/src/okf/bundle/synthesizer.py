from __future__ import annotations

import logging

log = logging.getLogger(__name__)


def _fallback(children: list[tuple[str, str]]) -> str:
    titles = ", ".join(t for t, _ in children[:5] if t) or "concepts"
    more = f" and {len(children) - 5} more" if len(children) > 5 else ""
    return f"Contains {len(children)} concepts ({titles}{more})."


def synthesize_description(
    rel_path: str,
    children: list[tuple[str, str]],
    *,
    model: str = "default",
) -> str:
    if not children:
        return ""
    # Deterministic, lightweight, vendor-neutral directory description synthesis
    descs = [d for _, d in children if d]
    if descs:
        # Use first child's summary pattern or combined count
        return f"Directory containing {len(children)} items, including: {children[0][0]}."
    return _fallback(children)
