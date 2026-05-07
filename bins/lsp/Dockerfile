ARG BUILD_IMAGE=buildimage

FROM ${BUILD_IMAGE} AS code

ARG BIN_NAME=nuon-lsp
ARG SERVICE=lsp
ARG PKG_ROOT=github.com/nuonco/nuon/bins/${SERVICE}

# copy each file that is specific to the lsp into the correct directory
WORKDIR /src/bins/${SERVICE}

# Copy monorepo shared sources (SDK go.mod/go.sum are in the base image)
COPY --link sdks /src/sdks/
COPY --link pkg /src/pkg/

# Copy service-specific files (most frequently changed, last for cache efficiency)
COPY --link handlers ./handlers/
COPY --link mappers ./mappers/
COPY --link models ./models/
COPY --link main.go .

FROM code AS build

RUN --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/${BIN_NAME}

FROM code AS test
RUN --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    go test ./...

FROM code AS lint
COPY .golangci.yml .
RUN --mount=type=cache,target=/root/.cache/golangci-lint,id=golint-lsp \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    golangci-lint run -c .golangci.yml -v

FROM golang:1.25-alpine AS final

ARG BIN_NAME=nuon-lsp
ARG SERVICE=lsp

WORKDIR /src/bins/${BIN_NAME}

COPY --from=build /bin/${BIN_NAME} /bin/${BIN_NAME}

ENTRYPOINT ["/bin/nuon-lsp"]

FROM code AS build-darwin-arm64
ARG BIN_NAME=nuon-lsp
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=darwin \
    GOARCH=arm64 \
    go build \
    -o /out/${BIN_NAME}_darwin_arm64 .

FROM code AS build-darwin-amd64
ARG BIN_NAME=nuon-lsp
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=darwin \
    GOARCH=amd64 \
    go build \
    -o /out/${BIN_NAME}_darwin_amd64 .

FROM code AS build-linux-arm
ARG BIN_NAME=nuon-lsp
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=linux \
    GOARCH=arm \
    go build \
    -o /out/${BIN_NAME}_linux_arm .

FROM code AS build-linux-arm64
ARG BIN_NAME=nuon-lsp
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=linux \
    GOARCH=arm64 \
    go build \
    -o /out/${BIN_NAME}_linux_arm64 .

FROM code AS build-linux-386
ARG BIN_NAME=nuon-lsp
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=linux \
    GOARCH=386 \
    go build \
    -o /out/${BIN_NAME}_linux_386 .

FROM code AS build-linux-amd64
ARG BIN_NAME=nuon-lsp
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
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
RUN apk add --no-cache pigz

ARG BIN_NAME=nuon-lsp

COPY --from=artifacts-build / /artifacts/

# Create gzip compressed versions of all binaries
RUN cd /artifacts && \
    for f in ${BIN_NAME}_*; do \
      echo "Compressing $f with pigz..." && \
      pigz -k "$f"; \
    done

# Export single binary for quick builds (no multi-platform)
FROM alpine:latest AS export-linux-amd64
RUN apk add --no-cache pigz

ARG BIN_NAME=nuon-lsp

COPY --from=build-linux-amd64 /out/${BIN_NAME}_linux_amd64 /artifacts/${BIN_NAME}_linux_amd64

# Create gzip compressed version
RUN cd /artifacts && \
    echo "Compressing ${BIN_NAME}_linux_amd64 with pigz..." && \
    pigz -k "${BIN_NAME}_linux_amd64"
