# Switch's headless agent controller (switch-agent-controller) with the Claude Code and Codex CLIs,
# so the agents Switch places on this machine run in the stack instead of in Switch Console.
# Build context: the repo root (see docker-compose.yml), for the Go commands in cmd/.

# The launch-room commands: `site`, `form`, `video` and `transcribe`, which the agents run; `gate`, the
# approval hook Claude Code runs before every shell command; and `roles`, which the entrypoint runs
# to give the agents their prompts and settings (embedded from internal/prompts and internal/roles).
FROM golang:1.27-bookworm AS commands
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -o /out/ ./cmd/site ./cmd/form ./cmd/video ./cmd/transcribe ./cmd/gate ./cmd/roles

FROM node:22-bookworm-slim

ARG CONTROLLER_VERSION=0.31.0

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl git procps \
    && rm -rf /var/lib/apt/lists/*

# Pandoc turns a deck site's Markdown slides into an editable PowerPoint (internal/sites/deck.go).
ARG PANDOC_VERSION=3.11
RUN arch="$(dpkg --print-architecture)" \
    && curl -fsSL -o /tmp/pandoc.deb \
        "https://github.com/jgm/pandoc/releases/download/${PANDOC_VERSION}/pandoc-${PANDOC_VERSION}-1-${arch}.deb" \
    && dpkg -i /tmp/pandoc.deb \
    && rm /tmp/pandoc.deb \
    && pandoc --version >/dev/null

RUN npm install --global --ignore-scripts \
        "https://github.com/sandbox-quantum/switch/releases/download/switch-agent-controller-v${CONTROLLER_VERSION}/switch-agent-controller-${CONTROLLER_VERSION}.tgz" \
    && npm install --global @anthropic-ai/claude-code @openai/codex @formspree/cli \
    && npm cache clean --force

COPY docker/agent-controller/entrypoint.sh /usr/local/bin/entrypoint.sh
COPY --from=commands /out/site /out/form /out/video /out/transcribe /out/gate /out/roles /usr/local/bin/

# The controller and every agent it starts run as this unprivileged user.
USER node
WORKDIR /home/node
RUN mkdir -p .local/state .switch .claude .codex sites videos public/preview public/live

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
