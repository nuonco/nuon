# Nuon BYOC Telemetry Relay

An OpenTelemetry Collector distribution that forwards logs, metrics, and traces from Nuon install runners and
the environment hosting the relay to an OTLP/HTTP backend. Both ingestion paths are enabled by default.

| Path | Default listener | Processing |
|---|---|---|
| Install | `0.0.0.0:4318` | Verify runner JWT, check optional org allowlist, replace the four identity IDs with verified values, apply install resource attributes |
| Environment | `0.0.0.0:5318` | No authentication or identity validation; preserve incoming attributes except configured resource changes |

Both paths accept `/v1/logs`, `/v1/metrics`, and `/v1/traces`. The environment path supports application,
infrastructure, and control-plane telemetry without requiring a Nuon control-plane or downstream install ID.
The relay also sends its own metrics through this path.

The install listener requires a Nuon control plane or compatible token issuer; arbitrary JWTs are not supported.
Signing keys must load successfully before either listener starts.

Forwarding is synchronous, with no relay queue or retry loop; senders own buffering and retries.

## Configuration

The supplied [Collector configuration](config.yaml) uses these environment variables:

| Variable | Purpose |
|---|---|
| `NUON_TELEMETRY_ISSUER` | Expected JWT issuer |
| `NUON_TELEMETRY_JWKS_URL` | Issuer's public signing-key URL |
| `NUON_TELEMETRY_JWKS_ALLOW_INSECURE` | Permit HTTP issuer and JWKS URLs; defaults to `false` |
| `NUON_TELEMETRY_ALLOWED_ORG_IDS` | YAML/JSON list of allowed JWT org IDs for install telemetry; unset or `[]` allows all verified orgs |
| `NUON_TELEMETRY_ENVIRONMENT_ENDPOINT` | Environment receiver bind address; defaults to `0.0.0.0:5318` |
| `NUON_TELEMETRY_SELF_OTLP_ENDPOINT` | Relay self-metrics OTLP/HTTP endpoint; defaults to `http://127.0.0.1:5318` |
| `VENDOR_OTLP_ENDPOINT` | Backend OTLP/HTTP base URL |
| `VENDOR_OTLP_AUTHORIZATION` | Backend `Authorization` header value |

Issuer and JWKS URLs require HTTPS by default. Enable HTTP only over a trusted network; exact JWT issuer matching
and HTTPS certificate verification remain enforced. Keep private signing keys on the issuer.

Terminate TLS in front of the install listener. The environment listener is intended for operator-managed internal
access: do not route public ingress to it. The relay does not enforce this networking policy. For local-only
producers, set `NUON_TELEMETRY_ENVIRONMENT_ENDPOINT=127.0.0.1:5318`. If changing the port or binding an address that
does not accept loopback traffic, update `NUON_TELEMETRY_SELF_OTLP_ENDPOINT` to reach that listener too.

Health checks listen on `0.0.0.0:13133`; keep them private. Prometheus self-metrics remain available on
`127.0.0.1:8888/metrics`.

## Install identity and organization filtering

`nuonidentity` stamps JWT-verified `nuon.org.id`, `nuon.app.id`, `nuon.install.id`, and `nuon.runner.id` on each
resource. It first removes these keys from all resource, scope, record, span/event/link, metric metadata/datapoint,
and exemplar attribute maps, matching case-insensitively and including fully underscored aliases such as
`nuon_org_id`. Other attributes, including unrelated Nuon IDs and `*.name` labels, are preserved.

Names are sender-supplied display/filtering labels: the relay neither looks them up nor verifies them against IDs.
Use verified IDs for authoritative filtering.

Restrict install telemetry by setting a YAML/JSON list of allowed JWT org IDs:

```sh
NUON_TELEMETRY_ALLOWED_ORG_IDS='["orgrok933tcyzji01s7us3aeo3", "org98e2wpzdxwoey393edtqj45"]'
```

The equivalent Collector setting is `processors.nuonidentity.allowed_org_ids`. Unset or empty allows all verified
orgs; entries must be nonblank, without surrounding whitespace. Matching is exact and uses the JWT org, not payload
IDs or names. Excluded requests are rejected before forwarding with a permanent error and may be discarded by the
sender; allowlist changes do not replay them. Environment telemetry and relay self-metrics are unaffected.

## Resource attributes

The standard resource processor overwrites resource-level `nuon.telemetry.source` with `install` or `environment`
to identify the ingestion path. Same-named attributes at other levels are untouched; query the resource attribute.

Customize `processors.resource/install.attributes` and `processors.resource/environment.attributes` in a mounted
Collector configuration or override file. Standard resource actions affect only resources: `insert` supplies a
missing value; `upsert` sets or replaces a value. For example:

```yaml
processors:
  resource/environment:
    attributes:
      - key: nuon.telemetry.source
        value: environment
        action: upsert
      - key: telemetry.dataset
        value: ${env:ENVIRONMENT_TELEMETRY_DATASET}
        action: upsert
      - key: deployment.environment.name
        value: production
        action: insert
```

Supply the override after the base configuration:

```sh
nuon-telemetry-relay \
  --config /etc/nuon-telemetry-relay/config.yaml \
  --config /etc/nuon-telemetry-relay/attributes.yaml
```

Override lists replace rather than append: retain the source-marker action and at least one action per processor.
`OTEL_RESOURCE_ATTRIBUTES` on the relay does not configure forwarded resource attributes.

Preserve producer identity (`service.*`, host/cloud/Kubernetes attributes) rather than assigning the relay's values
to every source. `nuon.control_plane.id` is optional and source-owned. Install resource actions run after
`nuonidentity`, so avoid overwriting the four verified IDs. Operator configuration is trusted.

### Backend filtering

OTLP separates resource, scope, and record/datapoint attributes without defining cross-level precedence. Use
resource-aware filters or selective resource-to-label promotion, and check collision handling in flattening backends.

For example, in a Prometheus server accepting OTLP, merge these keys into its existing resource-promotion list:

```yaml
otlp:
  promote_resource_attributes:
    - nuon.telemetry.source
    - nuon.org.id
    - nuon.app.id
    - nuon.install.id
```

This belongs in the backend, not the relay. Retain existing promoted keys, budget for ID/name cardinality, and query
the backend's label spelling (for example, `nuon_telemetry_source` with underscore normalization). See
[Prometheus OTLP resource promotion](https://prometheus.io/docs/guides/opentelemetry/#promoting-resource-attributes).

## Relay self-observability

The relay exports process and pipeline metrics every 60 seconds through the environment listener, with a five-second
export budget and `service.name=nuon-telemetry-relay`. The Collector adds its version and a unique process instance
ID; environment resource actions also apply. Self-metrics contribute to environment traffic counters. Logs go to
stderr; internal traces are disabled. Retain independent health or scrape monitoring to detect relay/backend outages.
