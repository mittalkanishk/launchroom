# stage: renders explainer scenes written in JavaScript with anime.js (docker/stage/runtime), frame by
# frame in headless Chromium, narrated with Microsoft Edge's free voices. Build context: the repo
# root. Like the Manim renderer it holds no keys: the scenes it runs are code the agents wrote.
FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -o /out/renderer ./cmd/renderer

# The Playwright image carries Chromium and its system libraries for this exact Playwright version.
FROM mcr.microsoft.com/playwright:v1.64.1-noble
RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg python3-venv fonts-inter fonts-jetbrains-mono fonts-noto-color-emoji \
    && rm -rf /var/lib/apt/lists/* \
    && python3 -m venv /opt/voice && /opt/voice/bin/pip install --no-cache-dir edge-tts==7.3.1
WORKDIR /opt/stage
COPY docker/stage/package.json ./
RUN PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 npm install --omit=dev --no-audit --no-fund
COPY docker/stage/runtime ./runtime
COPY docker/stage/examples ./examples
COPY --from=build /out/renderer /usr/local/bin/renderer
ENV RENDERER_ENGINE=stage \
    RENDERER_LISTEN=:8092 \
    STAGE_SCRIPT=/opt/stage/runtime/render.mjs \
    EDGE_TTS=/opt/voice/bin/edge-tts \
    LAUNCHROOM_VOICE_CACHE=/tmp/launchroom-voice \
    HOME=/tmp
USER pwuser
WORKDIR /tmp
EXPOSE 8092
ENTRYPOINT ["/usr/local/bin/renderer"]
