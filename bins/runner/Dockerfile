# syntax=docker/dockerfile:1.7

ARG BUILD_IMAGE=buildimage
ARG NUON_CLI_VERSION=0.19.1257
ARG AWS_IAM_AUTHENTICATOR_VERSION=0.7.23
ARG KUBELOGIN_VERSION=0.2.21
ARG GCLOUD_VERSION=590.0.0
ARG AZURE_CLI_VERSION=2.91.0
ARG PULUMI_VERSION=3.271.0
ARG KANIKO_VERSION=1.25.19

ARG GOCACHE_S3_URI
ARG GOCACHE_S3_ETAG
ARG GOCACHE_RELEASE_S3_URI
ARG GOCACHE_RELEASE_S3_ETAG
ARG GOCACHE_COMPACT=false
ARG ARTIFACT_PLATFORMS=full

FROM ${BUILD_IMAGE} AS gocache-restore
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ARG AWS_REGION
ARG GOCACHE_S3_URI
ARG GOCACHE_S3_ETAG
ARG GOCACHE_RESTORE_NONCE
ARG GOCACHE_COMPACT
ARG GOBUILD_CACHE_ID=gobuild-runner
ARG GOMOD_CACHE_ID=gomod-runner
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

FROM ${BUILD_IMAGE} AS gocache-restore-release
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
ARG AWS_REGION
ARG GOCACHE_RELEASE_S3_URI
ARG GOCACHE_RELEASE_S3_ETAG
ARG GOCACHE_RESTORE_NONCE
ARG GOCACHE_COMPACT
ARG GOBUILD_RELEASE_CACHE_ID=gobuild-runner-release
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_RELEASE_CACHE_ID},sharing=locked \
    --mount=type=secret,id=aws_access_key_id \
    --mount=type=secret,id=aws_secret_access_key \
    --mount=type=secret,id=aws_session_token \
    echo "restore check ${GOCACHE_RESTORE_NONCE}" && \
    if [ -n "${GOCACHE_RELEASE_S3_URI}" ] && \
       { [ "$(cat /go/pkg/build-cache/.nuon-s3-etag 2>/dev/null || true)" != "${GOCACHE_RELEASE_S3_ETAG}" ] || \
         [ "$(du -sk /go/pkg/build-cache 2>/dev/null | cut -f1)" -lt 51200 ]; }; then \
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
      if [ "$stream_status" -ne 0 ]; then \
        echo "WARNING: Go cache restore failed, this build compiles cold"; \
        stash /go/pkg/build-cache; wait; \
      else \
        if [ "${GOCACHE_COMPACT}" = "true" ]; then \
          find /go/pkg/build-cache -mindepth 2 -maxdepth 2 -exec touch -d "@0" {} +; \
          touch /go/pkg/build-cache/.nuon-compact-since; \
          date +%s > /go/pkg/build-cache/trim.txt; \
        fi && \
        echo "${GOCACHE_RELEASE_S3_ETAG}" > /go/pkg/build-cache/.nuon-s3-etag; \
      fi; \
    elif [ -z "${GOCACHE_RELEASE_S3_URI}" ]; then \
      echo "WARNING: no release Go cache seed available, this build compiles cold"; \
    else \
      echo "release Go cache already current"; \
    fi

FROM ${BUILD_IMAGE} AS code

WORKDIR /src
ARG SERVICE=runner
ARG PKG_ROOT=github.com/nuonco/nuon/bins/${SERVICE}
ARG GOBUILD_CACHE_ID=gobuild-runner
ARG GOBUILD_RELEASE_CACHE_ID=gobuild-runner-release
ARG GOMOD_CACHE_ID=gomod-runner


# copy each file that is specific to the api into the correct directory
WORKDIR /src/bins/${SERVICE}

# Copy monorepo shared sources (SDK go.mod/go.sum are in the base image)
COPY --link sdks /src/sdks/
COPY --link pkg /src/pkg/

# Copy service-specific files (most frequently changed, last for cache efficiency)
COPY --link cmd  ./cmd/
COPY --link internal ./internal/
COPY --link main.go .
COPY --link build-config.yaml .

FROM code AS build-otelcol

RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    go run go.opentelemetry.io/collector/cmd/builder@v0.163.0 --config build-config.yaml --skip-compilation && \
    CGO_ENABLED=0 GOOS=linux go build -C ./otelcol-build -o /bin/nuon-runner-otelcol && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS build

RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=linux \
    go build -o /bin/runner && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS build-versioned

ARG tag=

RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -ldflags="-X 'github.com/nuonco/nuon/pkg/runner/version.Version=${tag}'" \
    -o /bin/runner && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS test
# Only packages with tests (test-less ones serialize go test); -p 16 overlaps idle test runs with links.
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    export CGO_ENABLED=0 && \
    PKGS="$(go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)" && \
    go test -vet=off -p 16 ${PKGS} && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS lint
COPY .golangci.yml .
RUN --mount=type=cache,target=/root/.cache/golangci-lint,id=golint-runner \
    --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 golangci-lint run -c .golangci.yml -v && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.27-alpine AS nuon-cli

ARG NUON_CLI_VERSION

RUN apk add --update --no-cache curl ca-certificates
RUN curl "https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli/${NUON_CLI_VERSION}/nuon_linux_amd64" -o /bin/nuon && chmod +x /bin/nuon

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.27-alpine AS kaniko

ARG KANIKO_VERSION

RUN apk add --update --no-cache git ca-certificates
WORKDIR /src
RUN git clone --depth 1 --branch "v${KANIKO_VERSION}" https://github.com/chainguard-dev/kaniko.git .
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go get golang.org/x/net@v0.61.0 golang.org/x/crypto@v0.58.0 golang.org/x/text@v0.43.0 \
    google.golang.org/grpc@v1.84.0 github.com/containerd/containerd/v2@v2.4.2 \
    github.com/klauspost/compress@v1.20.3 \
    && go mod tidy && go mod vendor
RUN go build -o /usr/local/bin/docker-credential-gcr github.com/GoogleCloudPlatform/docker-credential-gcr/v2 \
    && go build -o /usr/local/bin/docker-credential-ecr-login github.com/awslabs/amazon-ecr-credential-helper/ecr-login/cli/docker-credential-ecr-login \
    && go build -o /usr/local/bin/docker-credential-acr-env github.com/chrismellard/docker-credential-acr-env \
    && go build -ldflags "-X github.com/chainguard-dev/kaniko/pkg/version.version=v${KANIKO_VERSION} -w -s" \
    -o out/executor github.com/chainguard-dev/kaniko/cmd/executor
RUN mkdir -p /kaniko/.docker /kaniko/ssl/certs \
    && cp out/executor /usr/local/bin/docker-credential-gcr /usr/local/bin/docker-credential-ecr-login /usr/local/bin/docker-credential-acr-env /kaniko/ \
    && cp /etc/ssl/certs/ca-certificates.crt /kaniko/ssl/certs/ \
    && cp files/nsswitch.conf /etc/nsswitch.conf

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.27-alpine AS aws-iam-authenticator

ARG AWS_IAM_AUTHENTICATOR_VERSION

RUN apk add --update --no-cache curl ca-certificates
WORKDIR /bin
RUN curl \
    -Lo aws-iam-authenticator \
    "https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v${AWS_IAM_AUTHENTICATOR_VERSION}/aws-iam-authenticator_${AWS_IAM_AUTHENTICATOR_VERSION}_linux_amd64" \
    && chmod +x aws-iam-authenticator

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.27-alpine AS kubelogin

ARG KUBELOGIN_VERSION

RUN apk add --update --no-cache curl ca-certificates
WORKDIR /bin
# kubelogin download url
RUN curl -LO \
    "https://github.com/Azure/kubelogin/releases/download/v${KUBELOGIN_VERSION}/kubelogin-linux-amd64.zip" \
    && unzip kubelogin-linux-amd64.zip \
    && mv bin/linux_amd64/kubelogin . \
    && rm kubelogin-linux-amd64.zip \
    && rm -r bin \
    && chmod +x kubelogin

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.27-alpine AS gcloud

ARG GCLOUD_VERSION

RUN apk add --update --no-cache curl ca-certificates python3
WORKDIR /opt
RUN curl -sSL "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-${GCLOUD_VERSION}-linux-x86_64.tar.gz" | tar xz \
    && /opt/google-cloud-sdk/install.sh --quiet --usage-reporting false --path-update false \
    && /opt/google-cloud-sdk/bin/gcloud components install gke-gcloud-auth-plugin --quiet \
    && find /opt/google-cloud-sdk/platform/gsutil \( -name Pipfile.lock -o -name "*requirements*.txt" \) -delete

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/python:3.12-alpine AS azure-cli

ARG AZURE_CLI_VERSION

RUN apk add --update --no-cache gcc musl-dev libffi-dev
RUN pip install --no-cache-dir --prefix=/azure-cli azure-cli==${AZURE_CLI_VERSION}

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.27-alpine AS pulumi

ARG PULUMI_VERSION

RUN apk add --update --no-cache curl ca-certificates
WORKDIR /opt
RUN curl -sSL "https://get.pulumi.com/releases/sdk/pulumi-v${PULUMI_VERSION}-linux-x64.tar.gz" | tar xz

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.27-alpine AS certs

# NOTE(jdt): ideally we would wget -a but busybox doesn't support it
# or use `update-ca-certificate` but it can't handle a single file w/ multiple certs
RUN wget \
    https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem \
    && cat global-bundle.pem >> /global-bundle.pem

# START

FROM code AS build-darwin-arm64

ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out

RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=darwin \
    GOARCH=arm64 \
    go build \
    -ldflags="-X 'github.com/nuonco/nuon/pkg/runner/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_arm64 . && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS build-darwin-amd64
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_RELEASE_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=darwin \
    GOARCH=amd64 \
    go build \
    -ldflags="-X 'github.com/nuonco/nuon/pkg/runner/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_amd64 . && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS build-linux-arm
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_RELEASE_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=arm \
    go build \
    -ldflags="-X 'github.com/nuonco/nuon/pkg/runner/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_arm . && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS build-linux-arm64
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_RELEASE_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=arm64 \
    go build \
    -ldflags="-X 'github.com/nuonco/nuon/pkg/runner/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_arm64 . && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS build-linux-386
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_RELEASE_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=386 \
    go build \
    -ldflags="-X 'github.com/nuonco/nuon/pkg/runner/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_386 . && \
    touch /go/pkg/build-cache/.nuon-dirty

FROM code AS build-linux-amd64
ARG BIN_NAME=runner
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_CACHE_ID},sharing=shared \
    --mount=type=cache,target=/go/pkg/mod,id=${GOMOD_CACHE_ID},sharing=shared \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags="-X 'github.com/nuonco/nuon/pkg/runner/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_amd64 . && \
    touch /go/pkg/build-cache/.nuon-dirty

# END

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.27-alpine AS runtime-base

RUN apk add --no-cache \
    bash \
    libc6-compat \
    build-base \
    make \
    ca-certificates-bundle \
    coreutils \
    curl \
    protoc \
    jq \
    git \
    aws-cli \
    kubectl \
    zip \
    openssl \
    helm \
    python3 \
    oras-cli

# add dependencies
COPY --from=kaniko /kaniko/* /kaniko/
COPY --from=kaniko /kaniko/.docker /kaniko/.docker
COPY --from=kaniko /etc/nsswitch.conf /etc/nsswitch.conf
COPY --from=nuon-cli /bin/nuon /usr/bin/nuon
COPY --from=aws-iam-authenticator /bin/aws-iam-authenticator /usr/local/bin/aws-iam-authenticator
COPY --from=kubelogin /bin/kubelogin /usr/local/bin/kubelogin
COPY --from=gcloud /opt/google-cloud-sdk /opt/google-cloud-sdk
COPY --from=azure-cli /azure-cli /usr/local
COPY --from=pulumi /opt/pulumi /opt/pulumi
COPY --from=certs /global-bundle.pem /etc/ssl/cert.pem

ENV PATH="/opt/pulumi:/opt/google-cloud-sdk/bin:${PATH}"

# udpate the ca certs so python's urllib cert verification passes
RUN update-ca-certificates

FROM runtime-base AS final

WORKDIR /src/bins/runner

COPY ./bundle /bundle
COPY --from=build-versioned /bin/runner /bin/runner
COPY --from=build-otelcol /bin/nuon-runner-otelcol /bin/nuon-runner-otelcol

ARG tag
ENV GIT_REF=$tag

ENTRYPOINT ["/bin/runner"]

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

# Prepare artifacts for upload
# Go code will handle the S3 upload part
FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/alpine:latest AS artifacts

ARG BIN_NAME=runner

COPY --from=artifacts-build / /artifacts/
COPY install.sh /artifacts/install.sh

FROM code AS gocache-status
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

FROM code AS gocache-release-status
ARG GOCACHE_RESTORE_NONCE
RUN --mount=type=cache,target=/go/pkg/build-cache,id=${GOBUILD_RELEASE_CACHE_ID},sharing=shared \
    echo "cache status ${GOCACHE_RESTORE_NONCE}" && \
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

FROM code AS gocache-s3-release
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
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
      KEPT_KB="$(find /go/pkg/build-cache -mindepth 2 -maxdepth 2 -type f -newer /go/pkg/build-cache/.nuon-compact-since -printf '%s\n' | awk '{s+=$1} END {print int(s/1024)}')"; \
      if [ "${KEPT_KB}" -ge 51200 ]; then \
        BEFORE="$(du -sh /go/pkg/build-cache | cut -f1)"; \
        find /go/pkg/build-cache -mindepth 2 -maxdepth 2 ! -newer /go/pkg/build-cache/.nuon-compact-since -exec rm -rf {} +; \
        rm /go/pkg/build-cache/.nuon-compact-since; \
        echo "compacted release Go build cache from ${BEFORE} to $(du -sh /go/pkg/build-cache | cut -f1)"; \
      else \
        echo "skipping compaction, this build used only ${KEPT_KB}KB of the cache"; \
      fi; \
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
