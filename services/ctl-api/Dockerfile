ARG BUILD_IMAGE=buildimage
ARG RUNTIME_IMAGE=runtimeimage

FROM ${BUILD_IMAGE} AS code

# Service specific arguments
ARG SERVICE=ctl-api
ARG PKG_ROOT=github.com/powertoolsdev/mono/services/ctl-api
ARG ETCSSL=/etc/ssl/cert.pem

# Copy service-specific files
WORKDIR /src/services/${SERVICE}

COPY cmd  ./cmd/
COPY internal ./internal/
COPY docs ./docs/
COPY main.go admin.go public.go runner.go ./
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-ctl-api \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-ctl-api \
    go generate ./...

FROM code AS build

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-ctl-api \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-ctl-api \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/service

FROM code AS test
RUN --mount=type=cache,target=/root/.cache/go-build,id=gobuild-ctl-api \
    go test -v ./...

FROM code AS lint

COPY .golangci.yml .

RUN --mount=type=cache,target=/root/.cache/golangci-lint,id=golint-ctl-api \
    golangci-lint run -c .golangci.yml --issues-exit-code 0 -v

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

ENTRYPOINT ["/bin/service"]
