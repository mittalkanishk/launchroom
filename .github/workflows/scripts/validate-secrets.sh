#!/usr/bin/env bash
# Variadic check: every named env var must be set and non-empty. Used to fail jobs early
# when GHA secrets injected as env vars are missing. Names are passed as arguments and
# resolved against the current process environment via bash indirect expansion (${!name}).
#
# Usage: bash validate-secrets.sh STAGING_DATABASE_URL R2_ACCESS_KEY_ID ...

set -euo pipefail

missing=""
for name in "$@"; do
  if [ -z "${!name:-}" ]; then
    missing="${missing} ${name}"
  fi
done

if [ -n "${missing}" ]; then
  echo "Missing required secrets:${missing}" >&2
  echo "Add them under Settings → Secrets and variables → Actions." >&2
  exit 1
fi
