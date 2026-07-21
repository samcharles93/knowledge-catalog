from __future__ import annotations

import json
from pathlib import Path

from okf.extractors.codebase import CodebaseExtractor
from okf.extractors.openapi import OpenAPIExtractor
from okf.extractors.sql import SQLExtractor


def test_codebase_extractor(tmp_path: Path):
    proj = tmp_path / "my_project"
    proj.mkdir()
    (proj / "README.md").write_text("# My App\nSample app readme.", encoding="utf-8")
    (proj / "main.py").write_text("def hello(): pass\n", encoding="utf-8")

    ext = CodebaseExtractor(proj)
    concepts = ext.extract_concepts()
    assert "architecture/overview" in concepts
    assert "codebase/main" in concepts


def test_openapi_extractor(tmp_path: Path):
    spec = tmp_path / "openapi.json"
    spec_data = {
        "openapi": "3.0.0",
        "info": {"title": "Test API", "version": "1.0"},
        "paths": {
            "/users": {
                "get": {
                    "summary": "List users",
                    "operationId": "listUsers"
                }
            }
        }
    }
    spec.write_text(json.dumps(spec_data), encoding="utf-8")

    ext = OpenAPIExtractor(spec)
    concepts = ext.extract_concepts()
    assert "api/listUsers" in concepts
    assert concepts["api/listUsers"].frontmatter["type"] == "API"


def test_sql_extractor(tmp_path: Path):
    sql_file = tmp_path / "schema.sql"
    sql_file.write_text(
        "CREATE TABLE users (\n"
        "  id INT PRIMARY KEY,\n"
        "  email VARCHAR(255)\n"
        ");",
        encoding="utf-8"
    )

    ext = SQLExtractor(sql_file)
    concepts = ext.extract_concepts()
    assert "database/users" in concepts
    assert concepts["database/users"].frontmatter["type"] == "Table"
