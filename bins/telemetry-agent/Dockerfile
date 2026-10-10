# syntax=docker/dockerfile:1.7
# Build from the repository root: docker build -f bins/telemetry-agent/Dockerfile .
FROM golang:1.25.8-alpine AS code
RUN apk add --no-cache bash git ca-certificates perl-utils
WORKDIR /src
COPY go.mod go.sum ./
COPY sdks ./sdks
COPY pkg ./pkg
COPY services/ctl-api ./services/ctl-api
COPY bins/telemetry-agent ./bins/telemetry-agent
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    NUON_GEN_TARGETS=runner go generate ./services/ctl-api && \
    go -C sdks/nuon-runner-go generate .

FROM code AS build
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o /bin/nuon-telemetry-agent ./bins/telemetry-agent && \
    go -C bins/telemetry-agent run go.opentelemetry.io/collector/cmd/builder@v0.150.0 --config build-config.yaml --skip-compilation && \
    CGO_ENABLED=0 go build -C bins/telemetry-agent/otelcol-build -o /bin/nuon-otelcol

FROM build AS test
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    NUON_TEST_OTELCOL=/bin/nuon-otelcol NUON_TEST_TELEMETRY_AGENT=/bin/nuon-telemetry-agent \
    go test -count=1 ./pkg/telemetryexport ./bins/telemetry-agent && \
    go -C sdks/nuon-runner-go test -count=1 ./...

FROM alpine:3.21 AS final
RUN apk add --no-cache ca-certificates && \
    addgroup -S -g 10001 agent && \
    adduser -S -D -H -u 10001 -G agent agent && \
    mkdir -p /var/lib/nuon/telemetry-agent && chown agent:agent /var/lib/nuon/telemetry-agent
COPY --from=build /bin/nuon-telemetry-agent /bin/nuon-telemetry-agent
COPY --from=build /bin/nuon-otelcol /bin/nuon-otelcol
USER agent
EXPOSE 4317 4318 13133
ENTRYPOINT ["/bin/nuon-telemetry-agent"]
