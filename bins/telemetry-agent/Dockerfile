# syntax=docker/dockerfile:1.7
# nuonctl assembles a service context. For a repository-root context, set AGENT_SOURCE=bins/telemetry-agent.
ARG BUILD_IMAGE=821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.27-alpine
ARG RUNTIME_IMAGE=821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/alpine:3.21
ARG ARTIFACT_PLATFORMS=full

FROM ${BUILD_IMAGE} AS code
RUN apk add --no-cache bash git ca-certificates perl-utils
WORKDIR /src
ARG AGENT_SOURCE=.
COPY --link go.mod go.sum ./
COPY --link sdks ./sdks
COPY --link pkg ./pkg
COPY --link services/ctl-api ./services/ctl-api
COPY --link ${AGENT_SOURCE}/*.go ${AGENT_SOURCE}/build-config.yaml ./bins/telemetry-agent/
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-telemetry-agent \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-telemetry-agent \
    NUON_GEN_TARGETS=runner go generate ./services/ctl-api && \
    go -C sdks/nuon-runner-go generate .

FROM code AS build
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-telemetry-agent \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-telemetry-agent \
    CGO_ENABLED=0 go build -o /bin/nuon-telemetry-agent ./bins/telemetry-agent && \
    go -C bins/telemetry-agent run go.opentelemetry.io/collector/cmd/builder@v0.150.0 --config build-config.yaml --skip-compilation && \
    CGO_ENABLED=0 go build -C bins/telemetry-agent/otelcol-build -o /bin/nuon-otelcol

FROM code AS lint
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-telemetry-agent \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-telemetry-agent \
    go vet ./pkg/telemetryexport ./bins/telemetry-agent && \
    go -C sdks/nuon-runner-go vet ./...

FROM build AS test
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-telemetry-agent \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-telemetry-agent \
    NUON_TEST_OTELCOL=/bin/nuon-otelcol NUON_TEST_TELEMETRY_AGENT=/bin/nuon-telemetry-agent \
    go test -count=1 ./pkg/telemetryexport ./bins/telemetry-agent && \
    go -C sdks/nuon-runner-go test -count=1 ./...

FROM ${RUNTIME_IMAGE} AS final
ARG VERSION=dev
ARG GIT_REF=local
LABEL org.opencontainers.image.title="Nuon telemetry agent" \
    org.opencontainers.image.version=${VERSION} \
    org.opencontainers.image.revision=${GIT_REF} \
    org.opencontainers.image.source="https://github.com/nuonco/nuon"
RUN apk add --no-cache ca-certificates && \
    addgroup -S -g 10001 agent && \
    adduser -S -D -H -u 10001 -G agent agent && \
    mkdir -p /var/lib/nuon/telemetry-agent && chown agent:agent /var/lib/nuon/telemetry-agent
COPY --from=build /bin/nuon-telemetry-agent /bin/nuon-telemetry-agent
COPY --from=build /bin/nuon-otelcol /bin/nuon-otelcol
USER agent
EXPOSE 4317 4318 13133
ENTRYPOINT ["/bin/nuon-telemetry-agent"]

FROM build AS build-linux-amd64
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-telemetry-agent \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-telemetry-agent \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/nuon-telemetry-agent_linux_amd64 ./bins/telemetry-agent && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -C bins/telemetry-agent/otelcol-build -o /out/nuon-otelcol_linux_amd64

FROM build AS build-linux-arm64
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-telemetry-agent \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-telemetry-agent \
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /out/nuon-telemetry-agent_linux_arm64 ./bins/telemetry-agent && \
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -C bins/telemetry-agent/otelcol-build -o /out/nuon-otelcol_linux_arm64

FROM scratch AS artifacts-build-pr
COPY --from=build-linux-amd64 /out/ /

FROM scratch AS artifacts-build-full
COPY --from=build-linux-amd64 /out/ /
COPY --from=build-linux-arm64 /out/ /

FROM artifacts-build-${ARTIFACT_PLATFORMS} AS artifacts-build

FROM ${RUNTIME_IMAGE} AS artifacts
RUN apk add --no-cache pigz
COPY --from=artifacts-build / /artifacts/
RUN cd /artifacts && pigz -k nuon-telemetry-agent_* nuon-otelcol_*
