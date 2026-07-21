---
name: okf-librarian
description: Keep an Open Knowledge Format (OKF) bundle in sync with the codebase — validate structure/links, harvest new services/APIs/tables/rules from recent changes, and regenerate index.md and viz.html. Use when the user asks to update, sync, validate, or clean up the OKF/knowledge bundle, after adding a new service/endpoint/table, or when starting work in a repo that has a .okf/ or bundles/*/ directory and recent commits haven't touched it.
---

# OKF Librarian

You are maintaining an OKF knowledge bundle (a tree of Markdown-with-frontmatter
concept docs under `.okf/` or `bundles/<name>/`) so it stays a trustworthy map of
the codebase for both humans and agents. Work in this order:

## 1. Locate the bundle

Check `.okf/` first, then `bundles/*/`. If neither exists, ask the user whether to
run `okf init --path .okf` before continuing — do not silently create one.

## 2. Check current health

```bash
okf validate --bundle <bundle>
```

Read every error and warning before touching anything. Broken links and orphan
concepts are cheap to introduce and easy to miss.

## 3. Find what's undocumented

Compare recent code changes against the bundle instead of guessing:

```bash
git log --oneline -20
git diff HEAD~10 --stat   # or since the bundle's last update
```

Look specifically for: new services/modules, new API routes, new DB
tables/migrations, new security- or architecture-relevant decisions. For each,
check whether a matching concept doc already exists under the bundle.

## 4. Fill gaps

Prefer the harvesters over hand-written docs when the source is structured:

```bash
okf harvest --type codebase --src <dir> --out <bundle>
okf harvest --type openapi --src openapi.json --out <bundle>
okf harvest --type sql --src schema.sql --out <bundle>
```

For anything a harvester can't infer (an architectural decision, a rule, a
"why"), write the concept doc by hand: YAML frontmatter (`type`, `title`,
`description`, `tags`) + a short Markdown body. Match the existing schema style
in sibling files in that bundle — don't invent a new frontmatter shape.

## 5. Regenerate derived artifacts

```bash
okf validate --bundle <bundle>   # confirm the gap-filling didn't break anything
okf visualize --bundle <bundle> --out <bundle>/viz.html
```

`index.md` files are regenerated automatically by the bundle/harvest/init
commands — don't hand-edit them.

## 6. Report, don't auto-commit

Summarize what changed (new/updated concept docs, validation status) and let
the user decide whether to commit. Do not commit or push on their behalf unless
they've asked you to in this conversation.

## Guardrails

- Never delete a concept doc just because it looks stale — flag it as a
  validate warning or ask the user; deletion loses history that isn't always
  recoverable from git alone (e.g. if the doc predates the repo).
- Don't invent concepts for things that aren't real yet (planned-but-unbuilt
  services) — the bundle should describe what exists, not the roadmap.
- If `okf` isn't installed, tell the user to run `pip install -e ./okf` (or
  wherever the toolkit lives in this repo) rather than working around it with
  raw file edits.
