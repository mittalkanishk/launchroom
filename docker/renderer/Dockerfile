# renderer: renders explainer scenes (cmd/renderer) with Manim Community and the launchroom scene
# library, narrated with Microsoft Edge's free voices. Build context: the repo root. It holds no keys:
# the scenes it runs are code the agents wrote.
FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -o /out/renderer ./cmd/renderer

FROM manimcommunity/manim:v0.22.1
USER root
# Manim's Python is the virtual environment in /opt/venv.
RUN /opt/venv/bin/python3 -m pip install --no-cache-dir edge-tts==7.3.1
COPY docker/renderer/launchroom /opt/launchroom/launchroom
COPY docker/renderer/examples /opt/launchroom/examples
COPY --from=build /out/renderer /usr/local/bin/renderer
ENV PYTHONPATH=/opt/launchroom \
    LAUNCHROOM_VOICE_CACHE=/tmp/launchroom-voice \
    XDG_CACHE_HOME=/tmp/cache \
    HOME=/tmp
USER manimuser
WORKDIR /tmp
EXPOSE 8091
ENTRYPOINT ["/usr/local/bin/renderer"]
