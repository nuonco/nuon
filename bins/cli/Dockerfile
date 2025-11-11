ARG BUILD_IMAGE=buildimage

FROM ${BUILD_IMAGE} AS code

ARG BIN_NAME=nuon
ARG SERVICE=cli
ARG GOCACHE=/go-cache
ARG GOMODCACHE=/go-mod-cache
ARG PKG_ROOT=github.com/powertoolsdev/mono/bins/${SERVICE}


# copy each file that is specific to the api into the correct directory
WORKDIR /src/bins/${SERVICE}

COPY cmd  ./cmd/
COPY internal ./internal/
COPY main.go .
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    go generate ./...

FROM code AS build

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/${BIN_NAME}

FROM code AS test
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    go test -v ./...

FROM code AS lint
COPY .golangci.yml .
RUN --mount=type=cache,target=/root/.cache/golangci-lint,id=golint-cli \
    golangci-lint run -c .golangci.yml --issues-exit-code 0 -v

FROM golang:1.24-alpine AS final

ARG BIN_NAME=nuon
ARG SERVICE=cli

WORKDIR /src/bins/${BIN_NAME}

COPY --from=build /bin/${BIN_NAME} /bin/${BIN_NAME}

ENTRYPOINT ["/bin/nuon"]

FROM code AS build-darwin-arm64
ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    GOOS=darwin \
    GOARCH=arm64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_arm64 .

FROM code AS build-darwin-amd64
ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    GOOS=darwin \
    GOARCH=amd64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_amd64 .

FROM code AS build-linux-arm
ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    GOOS=linux \
    GOARCH=arm \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_arm .

FROM code AS build-linux-arm64
ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    GOOS=linux \
    GOARCH=arm64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_arm64 .

FROM code AS build-linux-386
ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    GOOS=linux \
    GOARCH=386 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_386 .

FROM code AS build-linux-amd64
ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-cli \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-cli \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
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

ARG BIN_NAME=nuon

# Install p7zip and parallel for compression
RUN apk add --no-cache p7zip parallel

COPY --from=artifacts-build / /artifacts/
COPY install.sh /artifacts/install.sh

# Create 7z compressed versions of all binaries in parallel
RUN cd /artifacts && \
    ls -1 ${BIN_NAME}_* | \
    parallel -j0 --halt soon,fail=1 --ungroup \
      'echo "Compressing {} with 7z..." && \
       7z a -t7z -mx=9 "{}.7z" {}'

# Generate checksums for both original and compressed files
RUN cd /artifacts && sha256sum * > checksums.txt

# Export single binary for quick builds (no multi-platform)
FROM alpine:latest AS export-linux-amd64
ARG BIN_NAME=nuon

# Install p7zip for compression
RUN apk add --no-cache p7zip

COPY --from=build-linux-amd64 /out/${BIN_NAME}_linux_amd64 /artifacts/${BIN_NAME}_linux_amd64

# Create 7z compressed version
RUN cd /artifacts && \
    echo "Compressing ${BIN_NAME}_linux_amd64 with 7z..." && \
    7z a -t7z -mx=9 "${BIN_NAME}_linux_amd64.7z" "${BIN_NAME}_linux_amd64"

# Generate checksums for both original and compressed files
RUN sha256sum /artifacts/* > /artifacts/checksums.txt
