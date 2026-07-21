#!/bin/bash
# SessionStart hook: auto-loads OKF knowledge bundle context into every
# Claude Code session, so agents get architecture/API/rules context without
# anyone having to remember to run `okf context` by hand.
set -euo pipefail

cd "${CLAUDE_PROJECT_DIR:-.}" || exit 0

# Find the OKF bundle: the documented default is .okf/, but repos that keep
# multiple named bundles under bundles/ (as this repo does) are supported too.
bundle=""
if [ -f ".okf/index.md" ]; then
  bundle=".okf"
else
  for candidate in bundles/*/index.md; do
    [ -f "$candidate" ] || continue
    bundle="$(dirname "$candidate")"
    break
  done
fi

if [ -z "$bundle" ]; then
  exit 0
fi

echo "## Open Knowledge Format (OKF) bundle detected: \`$bundle\`"
echo

if command -v okf >/dev/null 2>&1; then
  okf context --bundle "$bundle"
else
  echo "(okf CLI not on PATH — showing raw index.md; run 'pip install -e ./okf' for full context/harvest/validate support)"
  echo
  cat "$bundle/index.md"
fi

echo
echo "*(Before architectural changes: check $bundle/rules/. After adding a service/API/table: update the matching $bundle/ concept doc and run 'okf validate --bundle $bundle'.)*"
exit 0
