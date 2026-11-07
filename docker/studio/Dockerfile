# studio: the video crew's back end (cmd/studio), with ffmpeg for joining scenes, burning in
# subtitles, colour, loudness and measurement. Build context: the repo root (see docker-compose.yml).
FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -o /out/studio ./cmd/studio

FROM alpine:3.22
# The subtitles filter (libass) burns in the narration; DejaVu Sans is the subtitle font.
RUN apk add --no-cache ffmpeg font-dejavu ca-certificates \
    && ffmpeg -hide_banner -filters 2>/dev/null | grep -q ' subtitles '
COPY --from=build /out/studio /usr/local/bin/studio
ENV STUDIO_DATA=/data
VOLUME /data
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/studio"]
