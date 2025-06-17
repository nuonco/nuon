FROM golang:1.24-alpine

ENV CGO_ENABLED=0

RUN --mount=type=cache,target=/var/cache/apk \
    apk add --no-cache \
    libc6-compat \
    build-base \
    make \
    ca-certificates-bundle \
    coreutils \
    curl \
    protoc \
    bash \
    jq \
    git \
    python3 \
    py3-pip \
    zip && \
    pip3 install awscli --upgrade --user --break-system-packages

ENV PATH=/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/root/.local/bin

