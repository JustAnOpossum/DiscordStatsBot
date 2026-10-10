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

USER vscode
