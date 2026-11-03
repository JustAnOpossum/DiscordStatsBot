# syntax=docker/dockerfile:1
#
# Unified build for DiscordStatsBot. This one Dockerfile replaces the old
# Alpine production Dockerfile and the VS Code devcontainer Dockerfile so dev,
# CI and production share a single build recipe (same base, same ImageMagick).
#
#   runtime  (default)  small prod image with the two binaries + ImageMagick
#                       runtime libs. This is what CI builds and pushes to GHCR,
#                       and what the production compose file runs.
#   dev                 full Go + ImageMagick-from-source toolchain for running
#                       `go test`/iteration against a bind-mounted repo
#                       (see docker-compose.dev.yml).
#
# Build the production image (default target, same contract as before):
#     podman build -t discordstatsbot .
# Build the dev image:
#     podman build --target dev -t discordstatsbot:dev .
#
# genImage/imageGen cgo-compiles against ImageMagick 7 (MagickWand), so that
# is built from source exactly once in `base` and shared by every stage.

# ---------------------------------------------------------------------------
# base: Go toolchain + ImageMagick 7 built from source (the cgo dependency)
# ---------------------------------------------------------------------------
ARG GO_VERSION=1.27
ARG DEBIAN_VERSION=bookworm
ARG IMAGEMAGICK_VERSION=7.1.0-57

FROM golang:${GO_VERSION}-${DEBIAN_VERSION} AS base
ARG IMAGEMAGICK_VERSION
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential pkg-config wget ca-certificates \
        libpng-dev libjpeg-dev libfreetype-dev \
    && rm -rf /var/lib/apt/lists/*

# Build once at the version the imagick.v3 binding expects. /usr/local/lib is
# registered with the dynamic linker for both the cgo link step and runtime.
RUN wget -q "https://github.com/ImageMagick/ImageMagick/archive/refs/tags/${IMAGEMAGICK_VERSION}.tar.gz" \
    && tar xf "${IMAGEMAGICK_VERSION}.tar.gz" \
    && cd "ImageMagick-${IMAGEMAGICK_VERSION}" \
    && ./configure >/dev/null \
    && make -j"$(nproc)" >/dev/null \
    && make install \
    && echo /usr/local/lib > /etc/ld.so.conf.d/imagemagick.conf \
    && ldconfig \
    && cd .. \
    && rm -rf "ImageMagick-${IMAGEMAGICK_VERSION}" "${IMAGEMAGICK_VERSION}.tar.gz"

# ---------------------------------------------------------------------------
# build: compile both Go modules against the ImageMagick above
# ---------------------------------------------------------------------------
FROM base AS build
WORKDIR /src

# Main bot module (pure Go, but built here so it lands in /out alongside imageGen)
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /out/bot .

# genImage/imageGen module (CGO, links MagickWand-7)
COPY genImage/go.mod genImage/go.sum ./genImage/
WORKDIR /src/genImage
RUN go mod download
COPY genImage/*.go ./
RUN CGO_ENABLED=1 go build -o /out/imageGen .

# ---------------------------------------------------------------------------
# dev: toolchain only (source is bind-mounted at /app by docker-compose.dev.yml)
# ---------------------------------------------------------------------------
FROM base AS dev
RUN apt-get update \
    && apt-get install -y --no-install-recommends git \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
CMD ["sleep", "infinity"]

# ---------------------------------------------------------------------------
# runtime: default target — what CI pushes and production runs
# ---------------------------------------------------------------------------
FROM debian:${DEBIAN_VERSION}-slim AS runtime
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates libpng16-16 libjpeg62-turbo libgomp1 libfreetype6 fontconfig \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /out/bot /app/bot
COPY --from=build /out/imageGen /app/genImage/imageGen
# ImageMagick shared libraries + configs built in `base` (MagickWand looks up
# /usr/local/etc and /usr/local/share at runtime for its module/format config)
COPY --from=base /usr/local/lib /usr/local/lib
COPY --from=base /usr/local/etc /usr/local/etc
COPY --from=base /usr/local/share/ImageMagick-7 /usr/local/share/ImageMagick-7
RUN echo /usr/local/lib > /etc/ld.so.conf.d/imagemagick.conf && ldconfig
# imageGen reads genImage/Static relative to the bot's working dir (/app)
COPY genImage/Static /app/genImage/Static

CMD ["/app/bot"]
