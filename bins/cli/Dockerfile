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
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go generate ./...

FROM code as build

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/${BIN_NAME}

FROM code as test
RUN --mount=type=cache,target=/root/.cache/go-build \
        go test -v ./...

FROM code as lint
COPY .golangci.yml .
RUN --mount=type=cache,target=/root/.cache/golangci-lint \
    golangci-lint run -c .golangci.yml --issues-exit-code 0 -v

FROM golang:1.24-alpine as final

WORKDIR /src/bins/${SERVICE}

COPY --from=build /bin/${SERVICE} /bin/${BIN_NAME}

ENTRYPOINT ["/bin/nuon"]

FROM code as build-darwin-arm64
ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=darwin \
    GOARCH=arm64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_arm64 .

FROM code as build-darwin-amd64
ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=darwin \
    GOARCH=amd64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_darwin_amd64 .

FROM code as build-linux-arm
ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux \
    GOARCH=arm \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_arm .

FROM code as build-linux-arm64
ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux \
    GOARCH=arm64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_arm64 .

FROM code as build-linux-386
ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux \
    GOARCH=386 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_386 .

FROM code as build-linux-amd64
ARG BIN_NAME=nuon
ARG tag=
RUN mkdir -p /out
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags="-X '${PKG_ROOT}/internal/version.Version=${tag}'" \
    -o /out/${BIN_NAME}_linux_amd64 .

FROM scratch AS artifacts 
COPY --from=build-darwin-arm64 /out/ /
COPY --from=build-darwin-amd64 /out/ /
COPY --from=build-linux-arm /out/ /
COPY --from=build-linux-arm64 /out/ /
COPY --from=build-linux-386 /out/ /
COPY --from=build-linux-amd64 /out/ /