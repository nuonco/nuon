# syntax=docker/dockerfile:1.7

ARG GOCACHE_S3_URI
ARG GOCACHE_S3_ETAG
ARG GOCACHE_COMPACT=false
ARG GOBUILD_CACHE_ID=gobuild-pkg
ARG GOMOD_CACHE_ID=gomod-pkg

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.26 AS tools
ARG GOBUILD_CACHE_ID
ARG GOMOD_CACHE_ID
ENV GOCACHE=/go/pkg/build-cache \
    GOMODCACHE=/go/pkg/mod
RUN apt-get update && apt-get install --no-install-recommends -y zstd && rm -rf /var/lib/apt/lists/*
ARG S5CMD_VERSION=2.2.2
RUN curl -sSL -o /tmp/s5cmd.tar.gz https://github.com/peak/s5cmd/releases/download/v${S5CMD_VERSION}/s5cmd_${S5CMD_VERSION}_Linux-64bit.tar.gz && \
    tar -xz -C /usr/local/bin -f /tmp/s5cmd.tar.gz s5cmd && \
    rm /tmp/s5cmd.tar.gz

FROM tools AS gocache-restore
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ARG AWS_REGION
ARG GOCACHE_S3_URI
ARG GOCACHE_S3_ETAG
ARG GOCACHE_RESTORE_NONCE
ARG GOCACHE_COMPACT
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=locked \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=locked \
    --mount=type=secret,id=aws_access_key_id \
    --mount=type=secret,id=aws_secret_access_key \
    --mount=type=secret,id=aws_session_token \
    echo "restore check ${GOCACHE_RESTORE_NONCE}" && \
    if [ -n "${GOCACHE_S3_URI}" ] && \
       { [ "$(cat /go/pkg/build-cache/.nuon-s3-etag 2>/dev/null || true)" != "${GOCACHE_S3_ETAG}" ] || \
         [ "$(du -sk /go/pkg/build-cache 2>/dev/null | cut -f1)" -lt 51200 ]; }; then \
      export AWS_ACCESS_KEY_ID="$(cat /run/secrets/aws_access_key_id)" && \
      export AWS_SECRET_ACCESS_KEY="$(cat /run/secrets/aws_secret_access_key)" && \
      export AWS_SESSION_TOKEN="$(cat /run/secrets/aws_session_token)" && \
      echo "restoring Go cache ${GOCACHE_S3_ETAG}" && \
      stash() { \
        dest="$1/.stale-$$$$"; \
        mkdir -p "$dest"; \
        find "$1" -mindepth 1 -maxdepth 1 ! -name ".stale-$$$$" -exec mv -t "$dest" {} +; \
        rm -rf "$dest" & \
      } && \
      RENAME=$(date +%s) && \
      stash /go/pkg/build-cache && \
      stash /go/pkg/mod && \
      echo "rename: $(($(date +%s) - RENAME))s" && \
      STREAM=$(date +%s) && \
      s5cmd cat "${GOCACHE_S3_URI}" | zstd -T0 -d | tar -xf - -C /go/pkg; \
      stream_status=$?; \
      echo "download+extract: $(($(date +%s) - STREAM))s"; \
      WAIT=$(date +%s); \
      wait; \
      echo "wipe-wait: $(($(date +%s) - WAIT))s"; \
      if [ "$stream_status" -ne 0 ]; then \
        echo "WARNING: Go cache restore failed, this build compiles cold"; \
        stash /go/pkg/build-cache; stash /go/pkg/mod; wait; \
      else \
        if [ "${GOCACHE_COMPACT}" = "true" ]; then \
          find /go/pkg/build-cache -mindepth 2 -maxdepth 2 -exec touch -d "@0" {} +; \
          touch /go/pkg/build-cache/.nuon-compact-since; \
          date +%s > /go/pkg/build-cache/trim.txt; \
        fi && \
        echo "${GOCACHE_S3_ETAG}" > /go/pkg/build-cache/.nuon-s3-etag; \
      fi; \
    elif [ -z "${GOCACHE_S3_URI}" ]; then \
      echo "WARNING: no Go cache seed available, this build compiles cold"; \
    else \
      echo "Go cache already current"; \
    fi

FROM tools AS base

WORKDIR /src

# Copy module files and local replace targets, then download dependencies
COPY --from=nuon go.mod go.sum ./
COPY --from=nuon sdks ./sdks/
RUN --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    go mod download

# Copy pkg source
COPY --from=nuon pkg ./pkg/

# Generate code
RUN --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    go generate ./pkg/... && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM base AS test
# Only packages with tests (test-less ones serialize go test); -p 16 overlaps idle test runs with links.
RUN --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    PKGS="$(go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./pkg/...)" && \
    go test -v -p 16 ${PKGS} && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM tools AS gocache-status
ARG GOCACHE_RESTORE_NONCE
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    echo "cache status ${GOCACHE_RESTORE_NONCE}" && \
    mkdir -p /gocache-status && \
    if [ -f /go/pkg/build-cache/.nuon-dirty ] && \
       [ "$(du -sk /go/pkg/build-cache | cut -f1)" -ge 51200 ]; then \
      echo true > /gocache-status/changed; \
    else \
      echo false > /gocache-status/changed; \
    fi

FROM scratch AS gocache-status-output
COPY --from=gocache-status /gocache-status /

FROM tools AS gocache-s3
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ARG AWS_REGION
ARG GOCACHE_S3_UPLOAD_URI
ARG GOCACHE_COMPACT
ARG tag
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    --mount=type=secret,id=aws_access_key_id \
    --mount=type=secret,id=aws_secret_access_key \
    --mount=type=secret,id=aws_session_token \
    export AWS_ACCESS_KEY_ID="$(cat /run/secrets/aws_access_key_id)" && \
    export AWS_SECRET_ACCESS_KEY="$(cat /run/secrets/aws_secret_access_key)" && \
    export AWS_SESSION_TOKEN="$(cat /run/secrets/aws_session_token)" && \
    if [ "${GOCACHE_COMPACT}" = "true" ] && [ -f /go/pkg/build-cache/.nuon-compact-since ]; then \
      KEPT_KB="$(find /go/pkg/build-cache -mindepth 2 -maxdepth 2 -type f -newer /go/pkg/build-cache/.nuon-compact-since -printf '%s\n' | awk '{s+=$1} END {print int(s/1024)}')"; \
      if [ "${KEPT_KB}" -ge 51200 ]; then \
        BEFORE="$(du -sh /go/pkg/build-cache | cut -f1)"; \
        find /go/pkg/build-cache -mindepth 2 -maxdepth 2 ! -newer /go/pkg/build-cache/.nuon-compact-since -exec rm -rf {} +; \
        rm /go/pkg/build-cache/.nuon-compact-since; \
        echo "compacted Go build cache from ${BEFORE} to $(du -sh /go/pkg/build-cache | cut -f1)"; \
      else \
        echo "skipping compaction, this build used only ${KEPT_KB}KB of the cache"; \
      fi; \
    fi && \
    if [ "$(du -sk /go/pkg/build-cache | cut -f1)" -lt 51200 ]; then \
      echo "cache mount held <50MB, refusing to upload an empty cache"; \
      exit 1; \
    fi && \
    echo "uploading cache for ${tag}" && \
    echo "mounts hold $(du -sh /go/pkg/build-cache /go/pkg/mod | paste -sd ' ' -)" && \
    START=$(date +%s) && \
    tar -C /go/pkg --exclude=build-cache/.nuon-dirty --exclude=build-cache/.nuon-s3-etag --exclude=build-cache/.nuon-compact-since --exclude=build-cache/.stale-* --exclude=mod/.stale-* -cf - build-cache mod \
      | zstd -T0 -1 \
      | s5cmd pipe --metadata "nuon-tag=${tag}" "${GOCACHE_S3_UPLOAD_URI}" && \
    echo "${tag}" > /go/pkg/build-cache/.nuon-s3-etag && \
    rm -f /go/pkg/build-cache/.nuon-dirty && \
    echo "S3 stream total: $(($(date +%s) - START))s"
