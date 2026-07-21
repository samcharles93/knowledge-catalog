from __future__ import annotations

import datetime
from pathlib import Path
import urllib.request
import urllib.parse
from html.parser import HTMLParser

from markdownify import markdownify as md

from okf.bundle.document import OKFDocument
from okf.extractors.base import BaseExtractor


class TitleExtractor(HTMLParser):
    def __init__(self):
        super().__init__()
        self.title = ""
        self._in_title = False

    def handle_starttag(self, tag, attrs):
        if tag.lower() == "title":
            self._in_title = True

    def handle_endtag(self, tag):
        if tag.lower() == "title":
            self._in_title = False

    def handle_data(self, data):
        if self._in_title:
            self.title += data


class WebExtractor(BaseExtractor):
    """Fetches web page URLs and transforms them into OKF reference concepts."""

    def __init__(self, urls: list[str]):
        self.urls = urls

    def extract_concepts(self) -> dict[str, OKFDocument]:
        concepts: dict[str, OKFDocument] = {}
        now = datetime.datetime.now(datetime.timezone.utc).isoformat()

        for url in self.urls:
            try:
                req = urllib.request.Request(url, headers={"User-Agent": "OKF-Extractor/1.0"})
                with urllib.request.urlopen(req, timeout=10) as resp:
                    html = resp.read().decode("utf-8", errors="replace")

                title_parser = TitleExtractor()
                title_parser.feed(html)
                title = title_parser.title.strip() or url

                markdown_body = md(html, heading_style="ATX").strip()

                parsed_url = urllib.parse.urlparse(url)
                slug = parsed_url.netloc + parsed_url.path.rstrip("/")
                clean_slug = "".join(c if c.isalnum() else "_" for c in slug)
                concept_id = f"references/{clean_slug}"

                doc = OKFDocument(
                    frontmatter={
                        "type": "Reference",
                        "title": title,
                        "description": f"Reference documentation from {url}",
                        "resource": url,
                        "tags": ["web", "reference"],
                        "timestamp": now,
                    },
                    body=f"# {title}\n\n**Source**: [{url}]({url})\n\n{markdown_body}\n",
                )
                concepts[concept_id] = doc
            except Exception as e:
                print(f"Warning: Failed to fetch {url}: {e}")

        return concepts
