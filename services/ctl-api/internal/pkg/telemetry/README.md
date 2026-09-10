# Control-plane operational telemetry

This package configures OTLP metric export for the control plane. It provides an
injected process resource and meter provider, including shutdown handling.
HTTP metric instrumentation lives in `internal/pkg/metrics`.

`internal/pkg/otel` remains separate: it owns product-facing OTLP payload types
and ingestion conversion helpers. This package exports telemetry about Nuon's
control plane; it does not ingest or process runner/application telemetry.

Metrics export automatically when an OTLP endpoint is configured. With no endpoint,
the meter provider is a no-op. This package does not configure log or trace
providers.

## Enable export

Configure each API process to send OTLP/HTTP protobuf to a control-plane-local
Collector or an authenticated OTLP backend:

```sh
OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
OTEL_EXPORTER_OTLP_TIMEOUT=5000
OTEL_RESOURCE_ATTRIBUTES=nuon.control_plane.id=cp-example,deployment.environment.name=production
```

The loopback URL assumes a sidecar Collector. Use a private Collector service
address for a separate deployment; use HTTPS across untrusted networks. Restrict
Collector ingress to control-plane workloads. The existing application relay
requires runner credentials and is not this metrics endpoint.

`OTEL_EXPORTER_OTLP_ENDPOINT` is the shared base URL; metrics append `/v1/metrics`
while retaining any base path. `OTEL_EXPORTER_OTLP_PROTOCOL` defaults to
`http/protobuf`. The same fields can be set in service configuration as
`otel_exporter_otlp_endpoint` and `otel_exporter_otlp_protocol`; environment values
take precedence. A configured invalid endpoint or unsupported protocol fails
configuration rather than silently selecting a destination. No endpoint means no
export, not an implicit localhost destination.

Use the generic base endpoint rather than `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT`.
With export enabled, nonempty metrics-specific transport overrides (endpoint, protocol, headers,
certificate, client certificate/key, timeout, compression and insecure) are
configuration errors. With no shared endpoint, transport settings are ignored and
no exporter is constructed. This prevents the SDK's signal-specific precedence
from silently diverging from the common transport policy.

Standard SDK settings supply headers and TLS certificates, including
`OTEL_EXPORTER_OTLP_HEADERS`, `OTEL_EXPORTER_OTLP_CERTIFICATE`,
`OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE`, and `OTEL_EXPORTER_OTLP_CLIENT_KEY`.
Configure these generic settings consistently for all signals and keep credentials
in deployment secrets. Invalid headers, unreadable/invalid certificates and
incomplete client certificate/key pairs fail configuration instead of allowing the
SDK to log and ignore them. Header values use percent encoding; `+` stays literal.
Certificates require HTTPS. `OTEL_EXPORTER_OTLP_INSECURE`, if supplied, must be
`true` for HTTP or `false` for HTTPS; it cannot override the endpoint scheme.
`OTEL_EXPORTER_OTLP_COMPRESSION` accepts `gzip` or `none`.

`OTEL_EXPORTER_OTLP_TIMEOUT` is a positive integer in milliseconds (SDK default:
10,000). Invalid values fail configuration. The metrics reader uses the SDK's
periodic export behavior, with a 60-second interval by default;
`OTEL_METRIC_EXPORT_INTERVAL` and `OTEL_METRIC_EXPORT_TIMEOUT` remain signal-specific
reader tuning, not destination/credential settings. Existing trace/log settings
are not rejected or migrated by this package.

Generic OTLP transport settings can also apply to existing log exporters in the
same process. Audit export is a no-op unless `AUDIT_OTLP_ENDPOINT` (service config:
`audit_otlp_endpoint`) is explicitly configured; the generic endpoint alone does
not enable it. If enabled without `AUDIT_OTLP_TOKEN`, it can inherit generic OTLP
headers, including credentials. Review signal-specific `OTEL_EXPORTER_OTLP_LOGS_*`
settings and explicit exporter options before enabling audit delivery alongside
operational metrics.

Workflow log delivery is configured separately in `internal/pkg/log` and does not
inherit OTEL environment settings for transport, resources, or record processing.

The default resource includes `service.name`, `service.version`, a random
process-lifetime `service.instance.id`, `nuon.service.type`, and
`nuon.service.deployment`. `OTEL_SERVICE_NAME` and `OTEL_RESOURCE_ATTRIBUTES` can
override resource values. Set a stable, opaque `nuon.control_plane.id` for the
BYOC deployment. If overriding `service.instance.id`, keep it unique per live
process; do not use one shared replica identifier. `nuon.service.deployment` is
the existing service deployment configuration, not a customer/install identity.

`DISABLE_METRICS` continues to control the existing Datadog path. It does not
disable this independent export. The OTel configuration/resource and meter provider
are injected, not installed globally; existing providers are not replaced.
Construct the shared configuration once per process,
not once per signal, to reuse the same generated instance ID.

The SDK exports operator-supplied `OTEL_RESOURCE_ATTRIBUTES` without an attribute
allowlist. Do not put credentials or sensitive identifiers in resource
configuration. Keep instance identity through export so replica counters remain
separate.

## Metrics

| Metric | Type | Unit | Dimensions |
| --- | --- | --- | --- |
| `http.server.request.duration` | Explicit-bucket histogram | seconds | `nuon.api`, `http.request.method`, `url.scheme`, `http.response.status_code`, matched `http.route`, `error.type` for 5xx |
| `http.server.active_requests` | Up/down counter | requests | `nuon.api`, `http.request.method`, `url.scheme` |

`nuon.api` is one of `public`, `runner`, `auth`, `internal`, `admin-dashboard`,
`slack`, or `mcp`. Metrics measure HTTP handling, not Temporal signals, individual
MCP tool outcomes, database operations, or downstream runner execution.

Histogram count supplies request volume; 5xx counts divided by total counts
supply an HTTP error ratio. The explicit boundaries in seconds are:
`0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10`.
Aggregation and temporality are fixed to explicit histograms and cumulative
values. Apply rates per instance before aggregating across replicas.

Gin routes use templates such as `/v1/apps/:app_id`. Unmatched requests and
router-generated redirects have no route label. MCP's catch-all route is `/`.
Raw paths, query strings, headers, user/org identifiers, and request bodies are
not metric dimensions. Unknown methods become `_OTHER`; the standard
`OTEL_INSTRUMENTATION_HTTP_KNOWN_METHODS` setting can override the known method
list. Scheme reflects the connection to the API, not untrusted forwarded headers.
Health requests are included and can be excluded by route in alert queries.
Streaming request duration measures the full handler lifetime.

## Failure behavior

Requests update in-memory aggregations; network export runs periodically outside
the request path. Each instrument is limited to 2,000 attribute sets with SDK
overflow aggregation. Exemplars are disabled so internal trace IDs are not
exported with customer metrics. FX orders API shutdown before provider shutdown.
The final export is best-effort: if API drain exhausts the application stop budget,
FX can skip the provider hook. When invoked, provider shutdown has a five-second
deadline bounded by the remaining application shutdown context.

Exporter errors pass through unchanged to the SDK error handler or flush/shutdown
caller, preserving endpoint, HTTP status and receiver response details supplied by
the SDK for debugging. Exporter initialization errors retain their underlying
cause. Diagnostics may contain sensitive receiver data. This package does not
replace the global OTel error handler.

The SDK has no persistent queue. Cumulative counts can survive a temporary export
failure while the process lives, but intermediate timing resolution is lost;
process loss can lose unexported data. Collector buffering and destination routing
are configured separately. Use missing-data alerts and independent availability
probes; an API cannot report its own total outage through this export path.

## Testing

Run the tests and request-recording benchmark from the repository root:

```sh
go test -race ./services/ctl-api/internal/pkg/telemetry ./services/ctl-api/internal/pkg/metrics ./services/ctl-api/internal/pkg/api ./services/ctl-api/internal/app/mcp/server
go test -run '^$' -bench '^BenchmarkHTTPMetrics$' -benchmem ./services/ctl-api/internal/pkg/telemetry
```

The benchmark measures `Start`/finish recording with no endpoint, a healthy
receiver and a blocked receiver. It does not measure the full HTTP stack or
production process memory usage.
