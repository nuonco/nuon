ARG BUILD_IMAGE=buildimage

FROM ${BUILD_IMAGE} AS code

ARG BIN_NAME=stack-cli
ARG SERVICE=stack-cli

WORKDIR /src/bins/${SERVICE}

# Shared monorepo sources
COPY --link sdks /src/sdks/
COPY --link pkg /src/pkg/

# Service-specific
COPY --link main.go .
COPY --link install.sh .

FROM code AS build
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/${BIN_NAME}

FROM code AS test
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    go test -v ./...

FROM code AS lint
COPY .golangci.yml .
RUN --mount=type=cache,target=/root/.cache/golangci-lint,id=golint-stack-cli \
    --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    golangci-lint run -c .golangci.yml -v

FROM code AS build-darwin-arm64
ARG BIN_NAME=stack-cli
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
    go build -ldflags="-X 'main.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_arm64 .

FROM code AS build-darwin-amd64
ARG BIN_NAME=stack-cli
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
    go build -ldflags="-X 'main.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_amd64 .

FROM code AS build-linux-arm64
ARG BIN_NAME=stack-cli
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
    go build -ldflags="-X 'main.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_arm64 .

FROM code AS build-linux-amd64
ARG BIN_NAME=stack-cli
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -ldflags="-X 'main.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_amd64 .

FROM scratch AS artifacts-build
COPY --from=build-darwin-arm64 /out/ /
COPY --from=build-darwin-amd64 /out/ /
COPY --from=build-linux-arm64  /out/ /
COPY --from=build-linux-amd64  /out/ /

FROM alpine:latest AS artifacts
RUN apk add --no-cache pigz

ARG BIN_NAME=stack-cli

COPY --from=artifacts-build / /artifacts/
COPY install.sh /artifacts/install.sh

RUN cd /artifacts && \
    for f in ${BIN_NAME}_*; do \
      echo "Compressing $f..."; \
      pigz -k "$f"; \
    done

FROM alpine:latest AS export-linux-amd64
RUN apk add --no-cache pigz

ARG BIN_NAME=stack-cli

COPY --from=build-linux-amd64 /out/${BIN_NAME}_linux_amd64 /artifacts/${BIN_NAME}_linux_amd64

RUN cd /artifacts && pigz -k "${BIN_NAME}_linux_amd64"
