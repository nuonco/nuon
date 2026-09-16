# syntax=docker/dockerfile:1.7

ARG BUILD_IMAGE=buildimage
ARG RUNTIME_IMAGE=runtimeimage

ARG GOCACHE_S3_URI
ARG GOCACHE_S3_ETAG
ARG GOCACHE_COMPACT=false

FROM ${BUILD_IMAGE} AS gocache-restore
ARG AWS_REGION
ARG GOCACHE_S3_URI
ARG GOCACHE_S3_ETAG
ARG GOCACHE_RESTORE_NONCE
ARG GOCACHE_COMPACT
ARG GOBUILD_CACHE_ID=gobuild-ctl-api
ARG GOMOD_CACHE_ID=gomod-ctl-api
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

# --- Admin dashboard React SPA build ---

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/node:20-alpine AS admin-dashboard-build

RUN npm i -g bun

WORKDIR /src

COPY --link internal/app/admin-dashboard/package.json \
     internal/app/admin-dashboard/bun.lock \
     ./

RUN --mount=type=cache,target=/root/.bun/install/cache,id=bun-admin-dashboard bun install --frozen-lockfile

COPY --link internal/app/admin-dashboard/client ./client
COPY --link internal/app/admin-dashboard/scripts ./scripts
COPY --link internal/app/admin-dashboard/postcss.config.js ./

RUN bun run build

# --- Go service build ---

FROM ${BUILD_IMAGE} AS code

# Service specific arguments
ARG SERVICE=ctl-api
ARG PKG_ROOT=github.com/nuonco/nuon/services/ctl-api
ARG ETCSSL=/etc/ssl/cert.pem
ARG GOBUILD_CACHE_ID=gobuild-ctl-api
ARG GOMOD_CACHE_ID=gomod-ctl-api

# Copy service-specific files
WORKDIR /src/services/${SERVICE}

# Copy monorepo shared sources (SDK go.mod/go.sum are in the base image)
COPY --link sdks /src/sdks/
COPY --link pkg /src/pkg/

# Copy service-specific files (most frequently changed, last for cache efficiency)
COPY --link cmd  ./cmd/
COPY --link internal ./internal/
RUN mkdir -p ./docs/admin/descriptions ./docs/public/descriptions ./docs/runner/descriptions
COPY --link docs/admin/descriptions ./docs/admin/descriptions/
COPY --link docs/public/descriptions ./docs/public/descriptions/
COPY --link docs/runner/descriptions ./docs/runner/descriptions/
COPY --link tests ./tests/
COPY --link main.go admin.go public.go runner.go ./

FROM code AS generated

RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 go generate ./... && \
    cd /src/pkg/temporal/client && \
    CGO_ENABLED=0 go generate ./... && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS build

COPY --from=generated /src/services/${SERVICE}/internal ./internal/
COPY --from=generated /src/services/${SERVICE}/docs ./docs/
COPY --from=generated /src/pkg/temporal/client/mock_client.go /src/pkg/temporal/client/mock_client.go

RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/service && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS unit-test

COPY --from=generated /src/services/${SERVICE}/internal ./internal/
COPY --from=generated /src/services/${SERVICE}/docs ./docs/
COPY --from=generated /src/pkg/temporal/client/mock_client.go /src/pkg/temporal/client/mock_client.go

RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 go test -vet=off ./... && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS lint

COPY --from=generated /src/services/${SERVICE}/internal ./internal/
COPY --from=generated /src/services/${SERVICE}/docs ./docs/
COPY --from=generated /src/pkg/temporal/client/mock_client.go /src/pkg/temporal/client/mock_client.go

COPY .golangci.yml .

RUN --mount=type=cache,target=/root/.cache/golangci-lint,id=golint-ctl-api \
    --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 golangci-lint run -c .golangci.yml -v && \
    touch /go/pkg/build-cache/.nuon-dirty

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

# Use the provided runtime image
FROM ${RUNTIME_IMAGE} AS final

WORKDIR /src/services/ctl-api

ENV PATH=/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/root/.local/bin

# expected as --build-args flags
ARG VERSION
ARG GIT_REF
ENV VERSION=$VERSION
ENV DD_VERSION=$VERSION
ENV GIT_REF=$GIT_REF

COPY bundle /bundle
COPY --from=build /bin/service /bin/service
COPY --from=admin-dashboard-build /src/dist ./internal/app/admin-dashboard/dist

ENTRYPOINT ["/bin/service"]
