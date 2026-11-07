#!/bin/sh
# Enrolls the controller once with the one-time code from Switch's Machines page, then runs it
# in the foreground. The enrollment is kept in the data volume, so later starts skip it.
set -eu

# The image starts as root only for this part, then runs the rest of this script as node.
if [ "$(id -u)" = 0 ]; then
  # One volume holds everything (a Railway service gets only one, mounted root-owned): the
  # enrollment, the logins, the sites and videos and the published pages. Each folder is kept in it
  # and linked from its usual place, so the controller and the commands find them unchanged.
  if [ -n "${CONTROLLER_DATA:-}" ]; then
    if ! mountpoint -q "$CONTROLLER_DATA"; then
      echo "agent-controller: no volume is mounted at $CONTROLLER_DATA; its data is lost when the container is replaced" >&2
    fi
    for pair in .local/state:state .switch:switch .claude:claude sites:sites public:public videos:videos; do
      home_path="/home/node/${pair%%:*}"
      data_path="$CONTROLLER_DATA/${pair#*:}"
      mkdir -p "$data_path"
      chown node:node "$data_path"
      if mountpoint -q "$home_path"; then
        echo "agent-controller: $home_path is a volume of its own; leaving it, not linking it to $data_path" >&2
      elif [ ! -L "$home_path" ]; then
        rm -rf "$home_path"
        ln -s "$data_path" "$home_path"
      fi
    done
    mkdir -p "$CONTROLLER_DATA/public/preview" "$CONTROLLER_DATA/public/live"
    chown node:node "$CONTROLLER_DATA" "$CONTROLLER_DATA/public/preview" "$CONTROLLER_DATA/public/live"
  fi
  exec setpriv --reuid=node --regid=node --init-groups env HOME=/home/node USER=node "$0" "$@"
fi

# Serves the published sites, preview-<name>.<domain> and <name>.<domain>, on :8080.
caddy run --config /etc/caddy/Caddyfile --adapter caddyfile &

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

# The launch-room roles (cmd/roles): every agent is Claude Code, its role picked by its name.
roles install >/dev/null

exec switch-agent-controller run --env-file "$ENV_FILE"
