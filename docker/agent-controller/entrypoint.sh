#!/bin/sh
# Enrolls the controller once with the one-time code from Switch's Machines page, then runs it
# in the foreground. The enrollment is kept in the data volume, so later starts skip it.
set -eu

ENV_FILE="$HOME/.local/state/agents.env"

if ! switch-agent-controller status >/dev/null 2>&1; then
  if [ -z "${SWITCH_CONTROLLER_ENROLL_CODE:-}" ]; then
    echo "agent-controller: not enrolled. In Switch, open Machines > Add machine, put the code in" >&2
    echo "SWITCH_CONTROLLER_ENROLL_CODE in .env (it works once, for ten minutes) and start the stack again." >&2
    exit 0
  fi
  switch-agent-controller enroll \
    --server "$SWITCH_SERVER_URL" \
    --code "$SWITCH_CONTROLLER_ENROLL_CODE" \
    --name "${SWITCH_CONTROLLER_NAME:-docker}"
fi

# The agents inherit their provider logins from this file, not from the controller's own environment.
umask 077
: > "$ENV_FILE"
for name in CLAUDE_CODE_OAUTH_TOKEN ANTHROPIC_API_KEY; do
  value="$(printenv "$name" || true)"
  [ -n "$value" ] && printf '%s=%s\n' "$name" "$value" >> "$ENV_FILE"
done

# The controller hands agents only a fixed list of provider settings, so the settings of the
# launch-room commands (site, form, video, transcribe) go in a file those commands read themselves.
SETTINGS=/home/node/.config/launch-room/settings.env
mkdir -p "$(dirname "$SETTINGS")"
rm -f "$(dirname "$SETTINGS")/site.env"
: > "$SETTINGS"
for name in SITE_TEMPLATES SITES_DOMAIN FORMSPREE_PROJECT_ID FORMSPREE_DEPLOY_KEY OPENAI_API_KEY TRANSCRIBE_MODEL STUDIO_URL; do
  value="$(printenv "$name" || true)"
  [ -n "$value" ] && printf '%s=%s\n' "$name" "$value" >> "$SETTINGS"
done

# Codex counts as signed in only with a stored login (~/.codex, in its volume), so an API key
# is saved there; without one, sign in once with ChatGPT (codex login --device-auth).
if [ -n "${OPENAI_API_KEY:-}" ]; then
  printenv OPENAI_API_KEY | codex login --with-api-key >/dev/null
fi

# The launch-room roles (cmd/roles): Claude Code builds the sites and Codex reviews them.
roles install >/dev/null

exec switch-agent-controller run --env-file "$ENV_FILE"
