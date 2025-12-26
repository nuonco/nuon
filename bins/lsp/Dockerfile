ARG BUILD_IMAGE=buildimage

FROM ${BUILD_IMAGE} AS code

ARG BIN_NAME=nuon-lsp
ARG SERVICE=lsp
ARG GOCACHE=/go-cache
ARG GOMODCACHE=/go-mod-cache
ARG PKG_ROOT=github.com/powertoolsdev/mono/bins/${SERVICE}

# copy each file that is specific to the lsp into the correct directory
WORKDIR /src/bins/${SERVICE}

COPY handlers ./handlers/
COPY mappers ./mappers/
COPY models ./models/
# Copy in-tree SDKs (required for go.mod replace directives)
COPY sdks ./sdks/
COPY main.go .
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-lsp \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-lsp \
    go generate ./...

FROM code AS build

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-lsp \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-lsp \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/${BIN_NAME}

FROM code AS test
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild-lsp \
    go test -v ./...

FROM code AS lint
COPY .golangci.yml .
RUN --mount=type=cache,target=/root/.cache/golangci-lint,id=golint-lsp \
    golangci-lint run -c .golangci.yml -v

FROM golang:1.25-alpine AS final

ARG BIN_NAME=nuon-lsp
ARG SERVICE=lsp

WORKDIR /src/bins/${BIN_NAME}

COPY --from=build /bin/${BIN_NAME} /bin/${BIN_NAME}

ENTRYPOINT ["/bin/nuon-lsp"]

FROM code AS build-darwin-arm64
ARG BIN_NAME=nuon-lsp
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-lsp \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-lsp \
    GOOS=darwin \
    GOARCH=arm64 \
    go build \
    -o /out/${BIN_NAME}_darwin_arm64 .

FROM code AS build-darwin-amd64
ARG BIN_NAME=nuon-lsp
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-lsp \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-lsp \
    GOOS=darwin \
    GOARCH=amd64 \
    go build \
    -o /out/${BIN_NAME}_darwin_amd64 .

FROM code AS build-linux-arm
ARG BIN_NAME=nuon-lsp
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-lsp \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-lsp \
    GOOS=linux \
    GOARCH=arm \
    go build \
    -o /out/${BIN_NAME}_linux_arm .

FROM code AS build-linux-arm64
ARG BIN_NAME=nuon-lsp
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-lsp \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-lsp \
    GOOS=linux \
    GOARCH=arm64 \
    go build \
    -o /out/${BIN_NAME}_linux_arm64 .

FROM code AS build-linux-386
ARG BIN_NAME=nuon-lsp
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-lsp \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-lsp \
    GOOS=linux \
    GOARCH=386 \
    go build \
    -o /out/${BIN_NAME}_linux_386 .

FROM code AS build-linux-amd64
ARG BIN_NAME=nuon-lsp
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-lsp \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-lsp \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -o /out/${BIN_NAME}_linux_amd64 .

FROM scratch AS artifacts-build
COPY --from=build-darwin-arm64 /out/ /
COPY --from=build-darwin-amd64 /out/ /
COPY --from=build-linux-arm /out/ /
COPY --from=build-linux-arm64 /out/ /
COPY --from=build-linux-386 /out/ /
COPY --from=build-linux-amd64 /out/ /

# Compress, checksum, and prepare artifacts for upload
# Go code will handle the S3 upload part
FROM alpine:latest AS artifacts

ARG BIN_NAME=nuon-lsp

# Install p7zip and parallel for compression
RUN apk add --no-cache p7zip parallel

COPY --from=artifacts-build / /artifacts/

# Create 7z compressed versions of all binaries in parallel
RUN cd /artifacts && \
    ls -1 ${BIN_NAME}_* | \
    parallel -j0 --halt soon,fail=1 --ungroup \
      'echo "Compressing {} with 7z..." && \
       7z a -t7z -mx=9 "{}.7z" {}'

# Export single binary for quick builds (no multi-platform)
FROM alpine:latest AS export-linux-amd64
ARG BIN_NAME=nuon-lsp

# Install p7zip for compression
RUN apk add --no-cache p7zip

COPY --from=build-linux-amd64 /out/${BIN_NAME}_linux_amd64 /artifacts/${BIN_NAME}_linux_amd64

# Create 7z compressed version
RUN cd /artifacts && \
    echo "Compressing ${BIN_NAME}_linux_amd64 with 7z..." && \
    7z a -t7z -mx=9 "${BIN_NAME}_linux_amd64.7z" "${BIN_NAME}_linux_amd64"
