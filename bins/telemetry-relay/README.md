# Nuon BYOC Telemetry Relay

This OpenTelemetry Collector distribution accepts OTLP/HTTP from install runners, verifies the runner-scoped telemetry
JWT issued by ctl-api, replaces all caller-supplied `nuon.*` attributes with authoritative resource identity, and
forwards telemetry to one vendor OTLP/HTTP backend.

The relay intentionally has no sending queue or retry loop. A leaf Collector keeps each request in its dedicated
persistent queue until the vendor backend accepts it through the relay.

A downstream partial rejection permanently fails the whole upstream request. The vendor receives its accepted subset
once; the leaf drops the rejected subset and records a permanent export failure rather than retrying the accepted data
or treating the partial rejection as full success.

## Configuration

| Environment variable | Purpose |
|---|---|
| `NUON_TELEMETRY_ISSUER` | Exact ctl-api JWT issuer |
| `NUON_TELEMETRY_JWKS_URL` | Pinned ctl-api JWKS URL |
| `VENDOR_OTLP_ENDPOINT` | Vendor OTLP/HTTP base endpoint |
| `VENDOR_OTLP_AUTHORIZATION` | Complete vendor `Authorization` header value |

Production issuer, JWKS, and vendor endpoints must use HTTPS. Loopback HTTP is supported for local validation of the
issuer and JWKS endpoint.

## Build and validate

```bash
go run go.opentelemetry.io/collector/cmd/builder@v0.150.0 \
  --config build-config.yaml \
  --skip-compilation
go build -C otelcol-build -o nuon-telemetry-relay

NUON_TELEMETRY_ISSUER=https://ctl-api.example.com \
NUON_TELEMETRY_JWKS_URL=https://ctl-api.example.com/.well-known/jwks.json \
VENDOR_OTLP_ENDPOINT=https://otlp.example.com \
VENDOR_OTLP_AUTHORIZATION='Bearer example' \
./otelcol-build/nuon-telemetry-relay validate --config config.yaml
```
