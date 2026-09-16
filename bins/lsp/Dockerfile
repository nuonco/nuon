# syntax=docker/dockerfile:1.7

ARG BUILD_IMAGE=buildimage

ARG GOCACHE_S3_URI
ARG GOCACHE_S3_ETAG
ARG GOCACHE_RELEASE_S3_URI
ARG GOCACHE_RELEASE_S3_ETAG
ARG GOCACHE_COMPACT=false

# Which artifact platform set to assemble: "full" for the release matrix, "pr"
# for the two platforms we actually run. Defaults to full so a caller that does
# not set it can never silently ship fewer binaries.
ARG ARTIFACT_PLATFORMS=full

FROM ${BUILD_IMAGE} AS gocache-restore
ARG AWS_REGION
ARG GOCACHE_S3_URI
ARG GOCACHE_S3_ETAG
ARG GOCACHE_RESTORE_NONCE
ARG GOCACHE_COMPACT
ARG GOBUILD_CACHE_ID=gobuild-lsp
ARG GOMOD_CACHE_ID=gomod-lsp
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=locked \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=locked \
    --mount=type=secret,id=aws_access_key_id \
    --mount=type=secret,id=aws_secret_access_key \
    --mount=type=secret,id=aws_session_token \
    echo "restore check ${GOCACHE_RESTORE_NONCE}" && \
    if [ -n "${GOCACHE_S3_URI}" ] && \
       [ "$(cat /go/pkg/build-cache/.nuon-s3-etag 2>/dev/null || true)" != "${GOCACHE_S3_ETAG}" ]; then \
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
      [ "$stream_status" -eq 0 ] && \
      echo "${GOCACHE_S3_ETAG}" > /go/pkg/build-cache/.nuon-s3-etag; \
    elif [ -z "${GOCACHE_S3_URI}" ]; then \
      echo "WARNING: no Go cache seed available, this build compiles cold"; \
    else \
      echo "Go cache already current"; \
    fi && \
    if [ "${GOCACHE_COMPACT}" = "true" ]; then \
      find /go/pkg/build-cache -mindepth 2 -maxdepth 2 -exec touch -d "@0" {} +; \
      touch /go/pkg/build-cache/.nuon-compact-since; \
      date +%s > /go/pkg/build-cache/trim.txt; \
    fi

FROM ${BUILD_IMAGE} AS gocache-restore-release
ARG AWS_REGION
ARG GOCACHE_RELEASE_S3_URI
ARG GOCACHE_RELEASE_S3_ETAG
ARG GOCACHE_RESTORE_NONCE
ARG GOCACHE_COMPACT
ARG GOBUILD_RELEASE_CACHE_ID=gobuild-lsp-release
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_RELEASE_CACHE_ID},sharing=locked \
    --mount=type=secret,id=aws_access_key_id \
    --mount=type=secret,id=aws_secret_access_key \
    --mount=type=secret,id=aws_session_token \
    echo "restore check ${GOCACHE_RESTORE_NONCE}" && \
    if [ -n "${GOCACHE_RELEASE_S3_URI}" ] && \
       [ "$(cat /go/pkg/build-cache/.nuon-s3-etag 2>/dev/null || true)" != "${GOCACHE_RELEASE_S3_ETAG}" ]; then \
      export AWS_ACCESS_KEY_ID="$(cat /run/secrets/aws_access_key_id)" && \
      export AWS_SECRET_ACCESS_KEY="$(cat /run/secrets/aws_secret_access_key)" && \
      export AWS_SESSION_TOKEN="$(cat /run/secrets/aws_session_token)" && \
      echo "restoring release Go cache ${GOCACHE_RELEASE_S3_ETAG}" && \
      stash() { \
        dest="$1/.stale-$$$$"; \
        mkdir -p "$dest"; \
        find "$1" -mindepth 1 -maxdepth 1 ! -name ".stale-$$$$" -exec mv -t "$dest" {} +; \
        rm -rf "$dest" & \
      } && \
      RENAME=$(date +%s) && \
      stash /go/pkg/build-cache && \
      echo "rename: $(($(date +%s) - RENAME))s" && \
      STREAM=$(date +%s) && \
      s5cmd cat "${GOCACHE_RELEASE_S3_URI}" | zstd -T0 -d | tar -xf - -C /go/pkg/build-cache; \
      stream_status=$?; \
      echo "download+extract: $(($(date +%s) - STREAM))s"; \
      WAIT=$(date +%s); \
      wait; \
      echo "wipe-wait: $(($(date +%s) - WAIT))s"; \
      [ "$stream_status" -eq 0 ] && \
      echo "${GOCACHE_RELEASE_S3_ETAG}" > /go/pkg/build-cache/.nuon-s3-etag; \
    elif [ -z "${GOCACHE_RELEASE_S3_URI}" ]; then \
      echo "WARNING: no release Go cache seed available, this build compiles cold"; \
    else \
      echo "release Go cache already current"; \
    fi && \
    if [ "${GOCACHE_COMPACT}" = "true" ]; then \
      find /go/pkg/build-cache -mindepth 2 -maxdepth 2 -exec touch -d "@0" {} +; \
      touch /go/pkg/build-cache/.nuon-compact-since; \
      date +%s > /go/pkg/build-cache/trim.txt; \
    fi

FROM ${BUILD_IMAGE} AS code

ARG BIN_NAME=nuon-lsp
ARG SERVICE=lsp
ARG PKG_ROOT=github.com/nuonco/nuon/bins/${SERVICE}
ARG GOBUILD_CACHE_ID=gobuild-lsp
ARG GOBUILD_RELEASE_CACHE_ID=gobuild-lsp-release
ARG GOMOD_CACHE_ID=gomod-lsp

WORKDIR /src/bins/${SERVICE}

COPY --link sdks /src/sdks/
COPY --link pkg /src/pkg/

COPY --link handlers ./handlers/
COPY --link mappers ./mappers/
COPY --link models ./models/
COPY --link main.go .

FROM code AS build

RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/${BIN_NAME} && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS test
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 go test -vet=off ./... && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS lint
COPY .golangci.yml .
RUN --mount=type=cache,target=/root/.cache/golangci-lint,id=golint-lsp \
    --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 golangci-lint run -c .golangci.yml -v && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.25-alpine AS final

ARG BIN_NAME=nuon-lsp
ARG SERVICE=lsp

WORKDIR /src/bins/${BIN_NAME}

COPY --from=build /bin/${BIN_NAME} /bin/${BIN_NAME}

ENTRYPOINT ["/bin/nuon-lsp"]

FROM code AS build-darwin-arm64
ARG BIN_NAME=nuon-lsp
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=darwin \
    GOARCH=arm64 \
    go build \
    -o /out/${BIN_NAME}_darwin_arm64 . && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS build-darwin-amd64
ARG BIN_NAME=nuon-lsp
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_RELEASE_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=darwin \
    GOARCH=amd64 \
    go build \
    -o /out/${BIN_NAME}_darwin_amd64 . && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS build-linux-arm
ARG BIN_NAME=nuon-lsp
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_RELEASE_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=arm \
    go build \
    -o /out/${BIN_NAME}_linux_arm . && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS build-linux-arm64
ARG BIN_NAME=nuon-lsp
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_RELEASE_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=arm64 \
    go build \
    -o /out/${BIN_NAME}_linux_arm64 . && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS build-linux-386
ARG BIN_NAME=nuon-lsp
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_RELEASE_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=386 \
    go build \
    -o /out/${BIN_NAME}_linux_386 . && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS build-linux-amd64
ARG BIN_NAME=nuon-lsp
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -o /out/${BIN_NAME}_linux_amd64 . && \
    touch /go/pkg/build-cache/.nuon-dirty

# Each extra GOOS/GOARCH is a full cross-compile. Releases come off main, so main
# builds the whole matrix while PRs build only what we actually run: linux/amd64
# for CI and containers, darwin/arm64 for dev machines. BuildKit never executes a
# stage the requested target does not reach, so the unselected cross-compiles
# cost nothing on a PR.
FROM scratch AS artifacts-build-pr
COPY --from=build-darwin-arm64 /out/ /
COPY --from=build-linux-amd64 /out/ /

FROM scratch AS artifacts-build-full
COPY --from=build-darwin-arm64 /out/ /
COPY --from=build-darwin-amd64 /out/ /
COPY --from=build-linux-arm /out/ /
COPY --from=build-linux-arm64 /out/ /
COPY --from=build-linux-386 /out/ /
COPY --from=build-linux-amd64 /out/ /

FROM artifacts-build-${ARTIFACT_PLATFORMS} AS artifacts-build

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/alpine:latest AS artifacts
RUN apk add --no-cache pigz

ARG BIN_NAME=nuon-lsp

COPY --from=artifacts-build / /artifacts/

RUN cd /artifacts && \
    for f in ${BIN_NAME}_*; do \
      echo "Compressing $f with pigz..." && \
      pigz -k "$f"; \
    done

FROM code AS gocache-status
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    mkdir -p /gocache-status && \
    if [ -f /go/pkg/build-cache/.nuon-dirty ] && \
       [ "$(du -sk /go/pkg/build-cache | cut -f1)" -ge 51200 ]; then \
      echo true > /gocache-status/changed; \
    else \
      echo false > /gocache-status/changed; \
    fi

FROM scratch AS gocache-status-output
COPY --from=gocache-status /gocache-status /

FROM code AS gocache-release-status
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_RELEASE_CACHE_ID},sharing=shared \
    mkdir -p /gocache-status && \
    if [ -f /go/pkg/build-cache/.nuon-dirty ] && \
       [ "$(du -sk /go/pkg/build-cache | cut -f1)" -ge 51200 ]; then \
      echo true > /gocache-status/changed; \
    else \
      echo false > /gocache-status/changed; \
    fi

FROM scratch AS gocache-release-status-output
COPY --from=gocache-release-status /gocache-status /

FROM code AS gocache-s3
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
      BEFORE="$(du -sh /go/pkg/build-cache | cut -f1)"; \
      find /go/pkg/build-cache -mindepth 2 -maxdepth 2 ! -newer /go/pkg/build-cache/.nuon-compact-since -exec rm -rf {} +; \
      rm /go/pkg/build-cache/.nuon-compact-since; \
      echo "compacted Go build cache from ${BEFORE} to $(du -sh /go/pkg/build-cache | cut -f1)"; \
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

FROM code AS gocache-s3-release
ARG AWS_REGION
ARG GOCACHE_RELEASE_S3_UPLOAD_URI
ARG GOCACHE_COMPACT
ARG tag
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_RELEASE_CACHE_ID},sharing=shared \
    --mount=type=secret,id=aws_access_key_id \
    --mount=type=secret,id=aws_secret_access_key \
    --mount=type=secret,id=aws_session_token \
    export AWS_ACCESS_KEY_ID="$(cat /run/secrets/aws_access_key_id)" && \
    export AWS_SECRET_ACCESS_KEY="$(cat /run/secrets/aws_secret_access_key)" && \
    export AWS_SESSION_TOKEN="$(cat /run/secrets/aws_session_token)" && \
    if [ "${GOCACHE_COMPACT}" = "true" ] && [ -f /go/pkg/build-cache/.nuon-compact-since ]; then \
      BEFORE="$(du -sh /go/pkg/build-cache | cut -f1)"; \
      find /go/pkg/build-cache -mindepth 2 -maxdepth 2 ! -newer /go/pkg/build-cache/.nuon-compact-since -exec rm -rf {} +; \
      rm /go/pkg/build-cache/.nuon-compact-since; \
      echo "compacted release Go build cache from ${BEFORE} to $(du -sh /go/pkg/build-cache | cut -f1)"; \
    fi && \
    if [ "$(du -sk /go/pkg/build-cache | cut -f1)" -lt 51200 ]; then \
      echo "release cache mount held <50MB, refusing to upload an empty cache"; \
      exit 1; \
    fi && \
    echo "uploading release cache for ${tag}" && \
    echo "mount holds $(du -sh /go/pkg/build-cache | cut -f1)" && \
    START=$(date +%s) && \
    tar -C /go/pkg/build-cache --exclude=.nuon-dirty --exclude=.nuon-s3-etag --exclude=.nuon-compact-since --exclude=.stale-* -cf - . \
      | zstd -T0 -1 \
      | s5cmd pipe --metadata "nuon-tag=${tag}" "${GOCACHE_RELEASE_S3_UPLOAD_URI}" && \
    echo "${tag}" > /go/pkg/build-cache/.nuon-s3-etag && \
    rm -f /go/pkg/build-cache/.nuon-dirty && \
    echo "release S3 stream total: $(($(date +%s) - START))s"

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/alpine:latest AS export-linux-amd64
RUN apk add --no-cache pigz

ARG BIN_NAME=nuon-lsp

COPY --from=build-linux-amd64 /out/${BIN_NAME}_linux_amd64 /artifacts/${BIN_NAME}_linux_amd64

RUN cd /artifacts && \
    echo "Compressing ${BIN_NAME}_linux_amd64 with pigz..." && \
    pigz -k "${BIN_NAME}_linux_amd64"
