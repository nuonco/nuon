# Control-plane operational telemetry

Nuon exports **operational metrics and selected lifecycle logs over OTLP** so you can monitor control-plane activity,
performance and failures in your observability backend.

| Area | What you can observe |
| --- | --- |
| API traffic | Request volume, latency, response codes, active requests and declared body size. |
| Dependencies and runtime | PostgreSQL/ClickHouse connection pools, readiness checks, Go memory, goroutines and process uptime. |
| Install operations | Config sync outcomes, drift-plan interpretation and health-evaluator activity. |
| Runner jobs | Polling/listeners, reported execution results and recorded lifecycle failures. |
| Queues and notifications | Temporal dispatch, local channel pressure, webhook/Slack delivery and lifecycle hooks. |
| Policies and storage | Policy decisions, install-state operations and blob-storage outcomes. |
| Workflow history | Workflow/step lifecycle events, approvals, retries and related resource IDs in logs. |
| Fleet state | Inventory, recorded health and current workflows through the optional [state telemetry collector](../statetelemetry/README.md). |

Metrics describe activity and recorded outcomes; logs provide context for individual workflows and installs.
Audit events, product log streams and tracing are separate signals.

## Enable export

Configure each API or worker process that performs work with an OTLP/HTTP protobuf endpoint:

```sh
OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
OTEL_EXPORTER_OTLP_TIMEOUT=5000
OTEL_RESOURCE_ATTRIBUTES=nuon.control_plane.id=cp-example,deployment.environment.name=production
```

Without an endpoint, metrics and lifecycle-log export are disabled. State collection is separately enabled with
`STATE_TELEMETRY_ENABLED=true`; operational metrics and logs do not require it.

The endpoint is a base URL: metrics use `/v1/metrics` and logs use `/v1/logs`. Enable both receiving pipelines.
Use HTTPS for remote destinations and keep authentication headers and client keys in deployment secrets.

| Setting | Purpose |
| --- | --- |
| `OTEL_EXPORTER_OTLP_HEADERS` | Backend authentication headers. |
| `OTEL_EXPORTER_OTLP_CERTIFICATE` | Trusted server CA certificate. |
| `OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE`, `OTEL_EXPORTER_OTLP_CLIENT_KEY` | Client TLS authentication. |
| `OTEL_EXPORTER_OTLP_COMPRESSION` | `gzip` or `none`. |
| `OTEL_EXPORTER_OTLP_TIMEOUT` | Positive timeout in milliseconds; default 10,000. |
| `OTEL_METRIC_EXPORT_INTERVAL`, `OTEL_METRIC_EXPORT_TIMEOUT` | Metric reader timing in milliseconds; export interval defaults to 60,000. |

Endpoint and protocol also accept service keys `otel_exporter_otlp_endpoint` and `otel_exporter_otlp_protocol`;
environment values take precedence. Use generic `OTEL_EXPORTER_OTLP_*` transport settings: nonempty signal-specific
`OTEL_EXPORTER_OTLP_METRICS_*` and `OTEL_EXPORTER_OTLP_LOGS_*` transport overrides are rejected when export is enabled.
TLS certificates require HTTPS, and `OTEL_EXPORTER_OTLP_INSECURE`, if set, must match the endpoint scheme.

### Resource identity

Use a stable `nuon.control_plane.id` to distinguish control planes. Each process also exports `service.name`,
`service.version`, `service.instance.id`, `nuon.service.type` and `nuon.service.deployment`.
`service.instance.id` is generated per process lifetime; any override must be unique per live process.
Set resource attributes with `OTEL_RESOURCE_ATTRIBUTES` and the service name with `OTEL_SERVICE_NAME`.
Do not include secrets in resource attributes.

## Metric reference

Names below are OpenTelemetry names, before backend-specific translation. Counters are cumulative: calculate rates
per instance before aggregating replicas. Retries count as separate attempts. Activity metrics appear when observed
and are not idle heartbeats. Durations are in seconds unless noted otherwise.

### API, dependencies and runtime

| Metrics | Meaning |
| --- | --- |
| `http.server.request.duration` | Request-duration histogram; its count provides request volume. Breakdowns include API, method, response code and normalized route. |
| `http.server.active_requests` | Requests currently being handled, by API and method. |
| `nuon.http.server.request.declared_body.size` | Declared request size in bytes, not bytes actually read; unknown sizes are omitted. |
| `pgxpool.*` | PostgreSQL pool connections, capacity, acquisitions, cancellations and waits. |
| `db.sql.connection.*` | ClickHouse pool connections, capacity, waits and connection closures. |
| `nuon.dependency.checks`, `nuon.dependency.check.duration` | Readiness-check outcomes and latency for PostgreSQL, ClickHouse and Temporal. |
| `nuon.dependency.check.status`, `.last_completed`, `.last_success` | Last completed check result (1 = success, 0 = failure) and Unix-second timestamps. |
| `go.memory.*`, `go.goroutine.count`, `go.processor.limit`, `go.config.gogc` | Go runtime memory, concurrency and garbage-collection settings. |
| `process.uptime` | Seconds since process creation. |

- HTTP duration includes the full handler lifetime, including streaming. Routes are templates, not resource URLs;
  unmatched requests have no route label. Health endpoints are included.
- Pool metrics distinguish database system and pool. PostgreSQL pool durations use **nanoseconds**; SQL wait duration
  uses **milliseconds**. A zero SQL connection limit means unlimited.
- Dependency gauges refresh when `/readyz` runs. Read timestamp age alongside status; skipped checks do not refresh it.
- `go.memory.used` is not RSS, and `go.memory.limit` is the Go soft limit, not a container limit.

### Operations

| Metrics | Meaning |
| --- | --- |
| `nuon.policy.evaluation.count`, `nuon.policy.evaluation.duration` | Policy attempts and latency, distinguishing pass/warn/deny decisions from evaluator errors. |
| `nuon.install.drift.plan.evaluation.attempts` | Component/sandbox drift-plan interpretation: drift, no drift, error or cancellation. |
| `nuon.install.component.health.evaluation.attempts`, `.duration` | Install-level health-evaluator attempts, skips, errors and latency. |
| `nuon.app.config.sync.attempts`, `.duration` | App-config sync success, rejection, error or cancellation, with failure stage. |
| `nuon.queue.enqueuer.dispatch.attempts`, `.duration` | Signal dispatch to Temporal, by dispatch source and outcome. |
| `nuon.queue.enqueuer.operations` | Dispatch-related persistence outcomes. |
| `nuon.queue.enqueuer.channel.dropped`, `.local.backlog`, `.local.processing` | Process-local channel overflow, queued signals and signals being processed. |
| `nuon.install.state.operations`, `nuon.install.state.operation.duration` | Install-state reads/saves and returned outcomes. |
| `nuon.blobstore.operations`, `nuon.blobstore.operation.duration` | S3/GCS reads, writes, metadata and streaming outcomes, including early close. |
| `nuon.notification.delivery.attempts`, `.duration` | Outbound webhook/Slack calls by channel, operation and outcome. |
| `nuon.event.hook.invocations` | Lifecycle-hook outcomes by hook, phase and before/after invocation. |

Successful health evaluation does not mean healthy workloads; use state telemetry for recorded verdicts.
Drift interpretation is not plan generation or a complete drift-check outcome. Config-sync success does not imply
downstream build/install success, and successful Temporal dispatch is not workflow completion.

Local queue gauges do not measure the durable queue. Notification retries and fan-out count separately; delivery
success means the outbound client call succeeded. Blob-stream body duration includes consumer time.

### Runner jobs

| Metrics | Meaning |
| --- | --- |
| `nuon.runner.job_tail.sessions`, `.probes`, `.notification.wakes`, `.listener.*` | Long-poll outcomes, probes, wakeups and listener connectivity/failures. |
| `nuon.runner.job.execution.results` | Newly recorded runner execution results by job type, operation and success/failure. |
| `nuon.runner.job.lifecycle.failures` | Recorded lifecycle failures such as unavailable/disabled runners, queue/pickup/execution timeouts and missing results. |

Empty polling timeouts are normal idle results. Execution results include plans, applies and actions; they do not
establish workflow completion or application health. Duplicate result reports do not count again, but new retry
executions do. Lifecycle-failure recordings can repeat and overlap execution results: **do not sum the two counters
as a total job-failure count**.

### Entity correlation

The following counters carry resource IDs when available; their duration histograms remain aggregate:

| Counter | Additional dimensions |
| --- | --- |
| `nuon.install.component.health.evaluation.attempts` | `nuon.install.id` |
| `nuon.install.drift.plan.evaluation.attempts` | `nuon.install.id`, plus `nuon.component.id` for component checks |
| `nuon.runner.job.execution.results` | `nuon.install.id` |
| `nuon.runner.job.lifecycle.failures` | `nuon.install.id` |
| `nuon.app.config.sync.attempts` | `nuon.app.id` |

Install IDs identify the affected install, not the control plane's hosting install. Unavailable IDs are omitted.
Combine them with `nuon.control_plane.id` to correlate downstream telemetry. Workflow, job, deploy and build IDs
remain log fields, not metric dimensions.

Counter/histogram streams allow 2,000 attribute sets per process; observable gauges allow 4,096. Excess combinations
aggregate into `otel.metric.overflow=true` without entity identity, so entity-filtered queries cannot include them.
Entity churn consumes the cumulative counter budget until process restart. State telemetry has its own
[collection-level and series limits](../statetelemetry/README.md#collection-level).

## Operational lifecycle logs

OTLP logs contain selected workflow/step, health, drift, config-update and runner-job lifecycle events. Structured JSON
bodies include event names, resource/workflow IDs, status, plan-only mode and retry context when available.
Use the body's `install_id` for the affected install and the resource's `nuon.control_plane.id` for its control plane.
IDs remain JSON fields rather than indexed Loki stream labels.

Workflow completion, failure, cancellation, approval and retry-wait transitions carry `source=workflow_status`.
They describe recorded status changes, not immutable final outcomes: an errored workflow can later retry.
Logs are diagnostic observations, not an exactly-once transition ledger or a source of exact workflow totals.

Raw errors, status descriptions, arbitrary metadata, account emails, stack traces and general process logs are not
included in this OTLP lifecycle stream. Existing stderr logging, audit and product log-stream destinations are unchanged.

## Delivery and scope

Metrics and logs are buffered in memory and exported best-effort. Export failures, queue saturation and process loss
can lose observations; collector persistence protects only data it has already accepted. Use independent availability
probes rather than interpreting silence as health or recovery.

Export destinations must be authorized for the resources they receive. Labels and dashboard filters are not tenant
access controls. Application-side metrics do not replace database-server, Temporal-server, node or workload monitoring.
