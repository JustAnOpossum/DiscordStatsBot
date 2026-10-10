FROM mcr.microsoft.com/devcontainers/go:2-1.27-trixie

ARG MONGO_VERSION=8.0.32

USER root

# MongoDB has no server package for Debian 13 (trixie) yet, so install the
# official debian12 tarball binary (links against libssl.so.3, provided by
# trixie's libssl3t64).
RUN curl -fsSLO "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-debian12-${MONGO_VERSION}.tgz" \
    && tar -xzf "mongodb-linux-x86_64-debian12-${MONGO_VERSION}.tgz" \
    && cp "mongodb-linux-x86_64-debian12-${MONGO_VERSION}/bin/mongod" /usr/local/bin/ \
    && rm -rf "mongodb-linux-x86_64-debian12-${MONGO_VERSION}" "mongodb-linux-x86_64-debian12-${MONGO_VERSION}.tgz"

# mongosh is published for trixie in the mongodb apt repo
RUN curl -fsSL https://pgp.mongodb.com/server-8.0.asc | gpg -o /usr/share/keyrings/mongodb-server-8.0.gpg --dearmor --yes \
    && echo "deb [ arch=amd64 signed-by=/usr/share/keyrings/mongodb-server-8.0.gpg ] https://repo.mongodb.org/apt/debian trixie/mongodb-org/8.0 main" > /etc/apt/sources.list.d/mongodb-org-8.0.list \
    && apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends mongodb-mongosh \
    && rm -rf /var/lib/apt/lists/*

# ImageMagick 7 built from source, mirroring dockerScripts/buildImageGen.sh
# (which targets Alpine/apk). The genImage module links against MagickWand-7
# through cgo, so the dev container needs the same headers/libs as the
# production image. Same version as prod: 7.1.0-57.
RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        build-essential pkg-config wget ca-certificates libpng-dev libjpeg-dev \
    && wget -q https://github.com/ImageMagick/ImageMagick/archive/refs/tags/7.1.0-57.tar.gz \
    && tar xf 7.1.0-57.tar.gz \
    && cd ImageMagick-7.1.0-57 \
    && ./configure \
    && make -j"$(nproc)" \
    && make install \
    && echo /usr/local/lib > /etc/ld.so.conf.d/imagemagick.conf \
    && ldconfig \
    && cd .. \
    && rm -rf ImageMagick-7.1.0-57 7.1.0-57.tar.gz \
    && rm -rf /var/lib/apt/lists/*

USER vscode
