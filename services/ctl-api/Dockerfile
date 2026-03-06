ARG BUILD_IMAGE=buildimage
ARG RUNTIME_IMAGE=runtimeimage

FROM ${BUILD_IMAGE} AS code

# Service specific arguments
ARG SERVICE=ctl-api
ARG PKG_ROOT=github.com/nuonco/nuon/services/ctl-api
ARG ETCSSL=/etc/ssl/cert.pem

# Copy service-specific files
WORKDIR /src/services/${SERVICE}

COPY cmd  ./cmd/
COPY internal ./internal/
# Copy only static docs inputs so generated swagger artifacts do not churn the cache.
RUN mkdir -p ./docs/admin/descriptions ./docs/public/descriptions ./docs/runner/descriptions
COPY docs/admin/descriptions ./docs/admin/descriptions/
COPY docs/public/descriptions ./docs/public/descriptions/
COPY docs/runner/descriptions ./docs/runner/descriptions/
COPY tests ./tests/
COPY main.go admin.go public.go runner.go ./

FROM code AS generated

# Generate all code artifacts required for compile/test (swagger + temporal + mocks).
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    go generate ./...

FROM code AS build

COPY --from=generated /src/services/${SERVICE}/internal ./internal/
COPY --from=generated /src/services/${SERVICE}/docs ./docs/

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/service

FROM code AS unit-test

COPY --from=generated /src/services/${SERVICE}/internal ./internal/
COPY --from=generated /src/services/${SERVICE}/docs ./docs/

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    go test -v ./...

FROM code AS lint

COPY --from=generated /src/services/${SERVICE}/internal ./internal/
COPY --from=generated /src/services/${SERVICE}/docs ./docs/

COPY .golangci.yml .

RUN --mount=type=cache,target=/root/.cache/golangci-lint,id=golint-ctl-api \
    --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    golangci-lint run -c .golangci.yml  -v

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
COPY --from=code /src/services/ctl-api/internal/app/admin-dashboard/assets ./internal/app/admin-dashboard/assets

ENTRYPOINT ["/bin/service"]
