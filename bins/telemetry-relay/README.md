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

The default configuration still requires a Nuon control plane or compatible token issuer for the install listener;
arbitrary JWTs are not supported. The JWT extension must load signing keys at startup, so a startup failure prevents
both listeners from serving. The paths share one process and destination, not independent failure domains.

Neither relay exporter has a sending queue or retry loop. Senders provide buffering, and delivery is not lossless.
Requests remain open until the synchronous backend export completes; the existing downstream-response guard is
used on both paths. There is no new retention or delivery guarantee for environment telemetry.

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

`nuonidentity` removes only these caller-supplied attributes before stamping verified IDs on each resource:

- `nuon.org.id`
- `nuon.app.id`
- `nuon.install.id`
- `nuon.runner.id`

Matching is case-insensitive and includes underscore-normalized aliases such as `nuon_org_id`. These identity
keys are removed from resource, scope, log, span/event/link, metric metadata/datapoint, and exemplar attributes
so lower-level values cannot shadow verified resource identity. Other `nuon.*` and `nuon_*` attributes survive,
including names, instrumentation dimensions, and unrelated IDs such as `nuon.control_plane.id`.

Runner-supplied `nuon.org.name`, `nuon.app.name`, `nuon.install.name`, and `nuon.runner.name` are preserved where
they appear. They are display/filtering labels, not verified identity: the relay does not look names up, attach
missing names, or check that a name belongs to a JWT ID. Use the IDs for authoritative filtering.

To restrict install telemetry to one or more organizations, set a YAML/JSON list of exact JWT org IDs:

```sh
NUON_TELEMETRY_ALLOWED_ORG_IDS='["orgrok933tcyzji01s7us3aeo3", "org98e2wpzdxwoey393edtqj45"]'
```

Alternatively, configure `processors.nuonidentity.allowed_org_ids` directly as a YAML list. An unset setting or
empty list allows all verified orgs. Blank or whitespace-padded entries are configuration errors; IDs are matched
exactly, without case folding, prefix matching, or wildcards.

The allowlist is checked against the authenticated principal before payload transformation or forwarding, never
against payload IDs or names. Excluded orgs receive a permanent processor error rather than a silent successful
drop or a retryable failure. A sender may discard that request; later allowlist changes do not replay it. This
is an install-ingress policy only: the trusted environment listener and relay self-metrics are unaffected.

## Resource attributes

Each path uses the standard Collector resource processor. By default it upserts `nuon.telemetry.source` to
`install` or `environment`. This attribute identifies the ingestion path, not an application's identity.
Install resource actions run after `nuonidentity`; environment resource actions do not require a verified principal.

The relay overwrites incoming resource-level `nuon.telemetry.source` values with the pipeline's configured value.
It does not copy that value onto records or remove same-named attributes at other levels. The canonical field for
queries is the exact `nuon.telemetry.source` resource attribute, not a record, datapoint, span, or scope attribute.

Customize `processors.resource/install.attributes` and `processors.resource/environment.attributes` in a mounted
Collector configuration or an override file. Use `insert` for a default that preserves a producer's value and
`upsert` to explicitly replace it. Other standard resource-processor actions are also available. Resource actions
do not rewrite datapoint, log-record, span, or instrumentation-scope attributes.

For example, this override adds different datasets to the two paths and an environment default:

```yaml
processors:
  resource/install:
    attributes:
      - key: nuon.telemetry.source
        value: install
        action: upsert
      - key: telemetry.dataset
        value: ${env:INSTALL_TELEMETRY_DATASET}
        action: upsert
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

Collector configuration merging replaces action lists rather than appending them, so include the source-marker
action when overriding a list. Keep at least one action per configured resource processor. Setting
`OTEL_RESOURCE_ATTRIBUTES` on the relay is not a substitute for configuring forwarded resource attributes.

The environment pipeline otherwise preserves source identity, including `service.*`, host/cloud/Kubernetes
attributes, and any `nuon.*` attributes. Sources may set `nuon.control_plane.id` when appropriate; the relay does
not require or assign it. Avoid configuring one shared `service.instance.id` or the relay's service name on all
forwarded telemetry. Likewise, the relay's cloud region need not be the downstream workload's region.

Operator-supplied install resource actions run after verified identity injection, so they can overwrite it.
Leave `nuon.org.id`, `nuon.app.id`, `nuon.install.id`, and `nuon.runner.id` unchanged to retain their JWT-derived
meaning. Configuration is trusted; there is no additional attribute-policy layer.

### Backend filtering

OTLP keeps resource, scope, and record/datapoint attributes separate. Use resource-aware filters in backends that
support them. For a backend that flattens attributes into labels, configure selective resource promotion rather
than duplicating all resource fields on every record or datapoint. OTel does not define a universal precedence
between same-named attributes at different levels: check the backend's collision handling so a sender-supplied
record or datapoint value does not take precedence over the relay's resource marker.

For example, in a Prometheus server accepting OTLP, merge these keys into its existing resource-promotion list:

```yaml
otlp:
  promote_resource_attributes:
    - nuon.telemetry.source
    - nuon.org.id
    - nuon.app.id
    - nuon.install.id
```

This is backend configuration, not relay configuration. Preserve any existing promoted keys such as service
identity. Choose additional IDs and names against the backend's cardinality budget, and use the backend's actual
label spelling when querying (for example, `nuon_telemetry_source` with underscore normalization).
See [Prometheus OTLP resource promotion](https://prometheus.io/docs/guides/opentelemetry/#promoting-resource-attributes).

## Relay self-observability

`service.telemetry.metrics` exports the Collector's own metrics every 60 seconds through the environment listener,
with a five-second export budget. This works without any external environment producers. The resource identifies
`service.name=nuon-telemetry-relay`; the Collector supplies its version and a unique process instance ID. Environment
resource actions also apply to these metrics. Relay logs remain on stderr and internal traces are not enabled.

The `normal` telemetry level includes process uptime, CPU and memory, receiver accepted/refused items, and exporter
sent/failed items where supported by each component. There are no queue occupancy metrics to rely on while exporter
queues are disabled. Component IDs distinguish `otlp/install` from `otlp/environment` receivers and
`otlp_http/install` from `otlp_http/environment` exporters. These replace the previous `otlp` receiver and
`otlp_http/vendor` exporter IDs; update queries that filter those component labels.

Self-metric export contributes to the environment pipeline's traffic counters. Identify relay metrics by their
service resource when separating them from other environment telemetry. Failed self-exports do not add a relay
queue; later periodic collection attempts continue. A failed relay or unavailable backend cannot reliably report
its own outage through this path, so retain independent health checks, the local scrape endpoint, and missing-data
monitoring. Do not route the relay's error logs or internal traces back through itself without addressing feedback.

## Build and acceptance checks

From this directory, regenerate the Collector distribution and build it:

```sh
go run go.opentelemetry.io/collector/cmd/builder@v0.150.0 --config build-config.yaml --skip-compilation
go build -C ./otelcol-build -o nuon-telemetry-relay .
```

With the required environment variables set, validate configuration without starting listeners:

```sh
./otelcol-build/nuon-telemetry-relay validate --config config.yaml
```

Before rollout, exercise the built relay against disposable JWKS and OTLP backend fixtures:

- Valid JWT install requests export all three signals with verified identity and `source=install`; invalid or
  missing JWTs do not reach the backend.
- Caller-supplied identity IDs and their aliases cannot shadow the verified resource IDs, while names and other
  Nuon attributes survive at every signal level.
- A configured org allowlist accepts each listed JWT org and rejects unlisted orgs without forwarding, even when
  the payload claims an allowed ID. Environment telemetry and relay self-metrics remain unaffected.
- Unauthenticated environment requests export all three signals with `source=environment`, preserving source
  resource and signal attributes without requiring any Nuon identity.
- Incoming resource-level `nuon.telemetry.source` values are overwritten on both paths. Same-named attributes at
  other levels, names, and unrelated attributes survive; backend filters select the resource marker.
- Override configuration proves `insert` preserves source values and `upsert` replaces them on both paths.
- With no external producers, relay metrics arrive over multiple intervals, remain bounded, and have distinct
  process instance IDs across restarts. The local Prometheus endpoint remains available.
- A delayed backend does not receive premature upstream acknowledgement. Backend failures retain the existing
  response behavior, and periodic self-metrics resume after backend recovery.
