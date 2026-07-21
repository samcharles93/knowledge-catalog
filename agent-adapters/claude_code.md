# OKF Integration for Claude Code

This directory ships two concrete, copy-paste integrations — not just prose to
retype into `CLAUDE.md`. Together they mean the bundle gets read on every
session automatically, and gets kept in sync by an agent instead of by memory.

| Artifact | What it does |
|---|---|
| [`claude-code/hooks/okf-session-context.sh`](claude-code/hooks/okf-session-context.sh) | `SessionStart` hook — auto-injects the bundle's `okf context` output into every session, no one has to remember to run it |
| [`claude-code/skills/okf-librarian/SKILL.md`](claude-code/skills/okf-librarian/SKILL.md) | A Skill that validates, harvests gaps from recent commits, and regenerates the bundle on request |
| [`claude-code/settings.snippet.json`](claude-code/settings.snippet.json) | The hook wiring to drop into `.claude/settings.json` |

## Install (in the target repo, not this one)

1. Copy the `claude-code/` directory (or just `hooks/` and `skills/`) into the
   target repo — e.g. as `.claude/okf/`.
2. Merge the `SessionStart` entry from `settings.snippet.json` into that
   repo's `.claude/settings.json`, adjusting the `command` path to wherever
   you copied the hook script.
3. Copy or symlink `skills/okf-librarian/` into that repo's `.claude/skills/`
   so Claude Code auto-discovers it.
4. Make sure the `okf` CLI is installed (`pip install -e /path/to/okf`) —
   without it on `PATH`, the hook falls back to printing raw `index.md`.
5. Restart Claude Code (hooks load at session start; edits mid-session don't
   take effect).

## What this gets you

- **Every session**: the bundle's progressive-disclosure index is injected as
  context automatically — the "on task start" step from the OKF spec happens
  without the agent needing to remember it.
- **On demand**: ask Claude to "sync the OKF bundle" / "update the knowledge
  base" / "check okf validate" and the `okf-librarian` skill triggers,
  diffing recent commits against the bundle and filling gaps via the
  harvesters (`okf harvest --type codebase|openapi|sql`) or hand-written
  concept docs where a harvester can't infer intent.
- **CLAUDE.md** (optional, still useful as a written policy even with the hook
  installed):

```markdown
## Open Knowledge Format (OKF)

This project uses OKF (`.okf/`) for repository architecture, API contracts,
and coding guidelines. A SessionStart hook already loads `.okf/index.md`
context automatically; see `.claude/skills/okf-librarian/` to keep it in
sync.

- Before architectural changes, check `.okf/rules/`.
- When adding a service/API/table, update the matching OKF concept doc (or
  ask the librarian skill to do it).
- Validate with `okf validate --bundle .okf` before committing.
```

## Reference example

This repo (`knowledge-catalog`) runs its own hook against
`bundles/microservices_app/` — see `.claude/settings.json` and
`agent-adapters/claude-code/hooks/okf-session-context.sh` for a working
example rather than a hypothetical one.

## Manual commands

```bash
okf context --bundle <bundle>    # what the hook injects, on demand
okf validate --bundle <bundle>   # what the librarian skill checks first
okf visualize --bundle <bundle> --out <bundle>/viz.html
```
