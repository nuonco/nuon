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

## Runner integration (pilot)

Configure `TELEMETRY_RELAY_ENDPOINT` in ctl-api with the relay's HTTPS OTLP base URL and configure its telemetry signing
key. Eligible install runners receive the endpoint through their authenticated settings. Existing runners need a build
that includes vendor telemetry support. No per-install secret is needed; the existing
`nuon/<install-id>/telemetry-export-config` secret controls only audit export.

Telemetry defaults to disabled for existing installs when the setting is first introduced and for new installs.
Configuring the relay endpoint alone does not activate collection. Enable each pilot install with an authenticated
`PATCH /v1/installs/{install_id}/telemetry` request containing `{"enabled": true}`. Use `{"enabled": false}` to disable
it, and `GET` on the same endpoint to read its setting.

Changing the schema default does not reset stored flags in an environment that already ran the default-enabled pilot.
Disable those installs explicitly before configuring the relay endpoint there.

The runner polls settings every 15 seconds and starts, stops, or replaces its separate vendor Collector without
restarting the runner or audit Collector. It obtains and renews short-lived relay JWTs independently, stores them in a
protected file, and removes credentials on disable/shutdown. Existing tokens can remain valid at the relay until expiry.
The vendor Collector accepts OTLP/gRPC on port 4317 and OTLP/HTTP on port 4318. These listeners require deliberately
scoped cloud firewall/network access; they do not authenticate local senders.

Logs, metrics, and traces each have a persistent byte-sized queue (1 GiB logical capacity per signal) under
`/var/lib/nuon/telemetry-export/vendor`. Disabling stops collection/export but preserves queued data for re-enablement.
The queue directories are separate from audit, not separate physical disks or quotas.

This remains a pilot: expired or temporarily unverifiable tokens can cause permanent export failures and data loss.
Production rollout requires the token-expiry/JWKS failure matrix, physical storage and process resource limits, and KMS
key management. `nuonctl`/Grafana local-development integration is a separate Mono change.

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
