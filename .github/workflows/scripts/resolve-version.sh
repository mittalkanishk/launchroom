#!/usr/bin/env bash
# Resolves the deploy VERSION (e.g. v1.2.0) and writes `version=<v>` to $GITHUB_OUTPUT.
# Source order:
#   1. workflow_dispatch input  -> $INPUT_VERSION
#   2. tag from triggering Release workflow_run -> $WORKFLOW_RUN_HEAD_BRANCH
#   3. exact tag on HEAD        -> `git describe --tags --exact-match HEAD`
# Fails if none resolve. Requires a full git checkout (fetch-depth: 0) for the fallback.

set -euo pipefail

if [ "${EVENT_NAME:-}" = "workflow_dispatch" ]; then
  echo "version=${INPUT_VERSION}" >> "$GITHUB_OUTPUT"
  exit 0
fi

v="${WORKFLOW_RUN_HEAD_BRANCH:-}"
if [[ "${v}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "version=${v}" >> "$GITHUB_OUTPUT"
  exit 0
fi

v="$(git describe --tags --exact-match HEAD 2>/dev/null || true)"
if [ -n "${v}" ]; then
  echo "version=${v}" >> "$GITHUB_OUTPUT"
  exit 0
fi

echo "Could not resolve version (EVENT_NAME='${EVENT_NAME:-}', WORKFLOW_RUN_HEAD_BRANCH='${WORKFLOW_RUN_HEAD_BRANCH:-}', git describe empty)." >&2
exit 1
