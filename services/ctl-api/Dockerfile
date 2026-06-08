ARG BUILD_IMAGE=buildimage
ARG RUNTIME_IMAGE=runtimeimage

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

# Generate all code artifacts required for compile/test (swagger + temporal + mocks).
RUN --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    go generate ./... && \
    cd /src/pkg/temporal/client && go generate ./...

FROM code AS build

COPY --from=generated /src/services/${SERVICE}/internal ./internal/
COPY --from=generated /src/services/${SERVICE}/docs ./docs/
COPY --from=generated /src/pkg/temporal/client/mock_client.go /src/pkg/temporal/client/mock_client.go

RUN --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/service

FROM code AS unit-test

COPY --from=generated /src/services/${SERVICE}/internal ./internal/
COPY --from=generated /src/services/${SERVICE}/docs ./docs/
COPY --from=generated /src/pkg/temporal/client/mock_client.go /src/pkg/temporal/client/mock_client.go

RUN --mount=type=cache,target=/go/pkg/build-cache,id=gobuild \
    go test ./...

FROM code AS lint

COPY --from=generated /src/services/${SERVICE}/internal ./internal/
COPY --from=generated /src/services/${SERVICE}/docs ./docs/
COPY --from=generated /src/pkg/temporal/client/mock_client.go /src/pkg/temporal/client/mock_client.go

COPY .golangci.yml .

RUN --mount=type=cache,target=/root/.cache/golangci-lint,id=golint-ctl-api \
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
COPY --from=admin-dashboard-build /src/dist ./internal/app/admin-dashboard/dist

ENTRYPOINT ["/bin/service"]
