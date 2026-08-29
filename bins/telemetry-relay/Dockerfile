FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/golang:1.25-alpine AS build

WORKDIR /src/bins/telemetry-relay

COPY bins/telemetry-relay ./

RUN --mount=type=cache,target=/go/pkg/mod,id=gomod-telemetry-relay \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild-telemetry-relay \
    go run go.opentelemetry.io/collector/cmd/builder@v0.150.0 \
    --config build-config.yaml \
    --skip-compilation && \
    CGO_ENABLED=0 GOOS=linux go build -C ./otelcol-build -o /bin/nuon-telemetry-relay

FROM 821160992543.dkr.ecr.us-west-2.amazonaws.com/docker.io/library/alpine:3.21 AS final

RUN apk add --no-cache ca-certificates && \
    addgroup -S -g 10001 relay && \
    adduser -S -D -H -u 10001 -G relay relay

COPY --from=build /bin/nuon-telemetry-relay /bin/nuon-telemetry-relay
COPY bins/telemetry-relay/config.yaml /etc/nuon-telemetry-relay/config.yaml

USER relay
EXPOSE 4318 13133

ENTRYPOINT ["/bin/nuon-telemetry-relay"]
CMD ["--config", "/etc/nuon-telemetry-relay/config.yaml"]
