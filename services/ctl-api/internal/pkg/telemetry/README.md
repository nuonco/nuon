# Control-plane operational telemetry

This package provides the control plane's resource identity, meter provider, and
OTLP metric exporter. Export is enabled when an endpoint is configured; otherwise
the meter provider is a no-op. Log and trace providers are configured separately.

## Enable export

Configure each API or worker process with an OTLP/HTTP protobuf endpoint:

```sh
OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
OTEL_EXPORTER_OTLP_TIMEOUT=5000
OTEL_RESOURCE_ATTRIBUTES=nuon.control_plane.id=cp-example,deployment.environment.name=production
```

The example uses a local Collector. Use HTTPS for remote backends and store
authentication headers and client keys in deployment secrets.

- Metrics append `/v1/metrics` to the base endpoint, preserving its path.
- Endpoint and protocol can also be set as `otel_exporter_otlp_endpoint` and
  `otel_exporter_otlp_protocol` in service configuration; environment values take precedence.
- Use generic `OTEL_EXPORTER_OTLP_*` transport settings. Nonempty
  `OTEL_EXPORTER_OTLP_METRICS_*` transport overrides are rejected when export is enabled.
- Authentication and TLS use `OTEL_EXPORTER_OTLP_HEADERS`,
  `OTEL_EXPORTER_OTLP_CERTIFICATE`, `OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE`, and
  `OTEL_EXPORTER_OTLP_CLIENT_KEY`. Certificates require HTTPS.
- `OTEL_EXPORTER_OTLP_INSECURE`, if set, must match the endpoint scheme.
  `OTEL_EXPORTER_OTLP_COMPRESSION` accepts `gzip` or `none`.
- `OTEL_EXPORTER_OTLP_TIMEOUT` is a positive integer in milliseconds (default: 10,000).
  Export interval defaults to 60 seconds; tune the reader with
  `OTEL_METRIC_EXPORT_INTERVAL` and `OTEL_METRIC_EXPORT_TIMEOUT`.

Invalid transport configuration fails startup when export is enabled.

### Resource identity

Default attributes are `service.name`, `service.version`, `service.instance.id`,
`nuon.service.type`, and `nuon.service.deployment`. Override them with
`OTEL_SERVICE_NAME` and `OTEL_RESOURCE_ATTRIBUTES`. `service.instance.id` defaults
to a random process-lifetime ID; overrides must be unique per live process.
Use a stable `nuon.control_plane.id` to identify a deployment. Resource attributes
are exported as configured; do not include secrets.

## Metrics

Counters are cumulative; apply rates per instance before aggregating replicas.
Retries count as separate attempts. Request and operation metrics appear when
observed; they do not provide an idle heartbeat.

### HTTP

| Metric | Type | Unit | Dimensions |
| --- | --- | --- | --- |
| `http.server.request.duration` | Explicit-bucket histogram | seconds | `nuon.api`, `http.request.method`, `url.scheme`, `http.response.status_code`, matched `http.route`, `error.type` for 5xx |
| `http.server.active_requests` | Up/down counter | requests | `nuon.api`, `http.request.method`, `url.scheme` |
| `nuon.http.server.request.declared_body.size` | Explicit-bucket histogram | bytes | Same as request duration |

`nuon.api` is one of `public`, `runner`, `auth`, `internal`, `admin-dashboard`,
`slack`, or `mcp`. Routes are templates such as `/v1/apps/:app_id`; unmatched
requests and router-generated redirects have no route label. Health requests are included.

Duration covers the full handler lifetime, including streaming; its histogram
count provides request volume. Declared body size records `ContentLength`, not
bytes read. Unknown lengths are omitted; zero lengths are included.

### Database pools

Pool metrics are registered in `internal/pkg/db/poolmetrics`.

| Pool | Library | Metrics |
| --- | --- | --- |
| Native PostgreSQL `pgxpool` | [`otelpgx`](https://github.com/exaring/otelpgx/tree/v0.11.1) | `pgxpool.*`: connections, capacity, acquisitions, cancellations, waits and connection creation/expiry |
| ClickHouse `database/sql` | [`otelsql`](https://github.com/XSAM/otelsql/tree/v0.41.0) | `db.sql.connection.*`: connections, capacity, waits and connection closure by limit |

Dimensions are `db.system.name` (`postgresql`, `clickhouse`) and
`db.client.connection.pool.name` (`primary`, `replica`, `admin_replica`).
ClickHouse uses `primary`; `db.sql.connection.open` adds `status=inuse|idle`.

Connection counts are current values; acquisitions, waits, and closures are cumulative.
PostgreSQL durations use **nanoseconds**; SQL wait duration uses **milliseconds**.
PostgreSQL acquisition time covers successful acquisitions; SQL waits include
canceled waits. A zero SQL connection limit means unlimited.

### Dependency health

API `/readyz` checks emit process-level metrics with `dependency.name` equal to
`postgresql`, `clickhouse`, or `temporal`.

| Metric | Type | Meaning |
| --- | --- | --- |
| `nuon.dependency.checks` | Counter | Outcomes: `success`, `failure`, `skipped`; bounded `error.type` on failure/skip |
| `nuon.dependency.check.duration` | Histogram, seconds | Attempted checks by outcome |
| `nuon.dependency.check.status` | Gauge | Last completed result: 1 = success, 0 = failure |
| `nuon.dependency.check.last_completed` | Gauge, Unix seconds | Last completed check timestamp |
| `nuon.dependency.check.last_success` | Gauge, Unix seconds | Last successful check timestamp |

Failure types are `connection`, `ping`, `query`, `scan`, `iteration`,
`readonly_replicas`, or `incomplete`. Skipped checks use `previous_dependency_failed`
and do not refresh state. ClickHouse success requires ping and replica checks to pass.
Gauges appear after the first corresponding result. Check timestamp age alongside
status: metrics only refresh when `/readyz` runs.

### Runtime/process

[`instrumentation/runtime`](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/v1.43.0/instrumentation/runtime)
emits `go.memory.used`, `go.memory.limit`,
`go.memory.allocated`, `go.memory.allocations`, `go.memory.gc.goal`,
`go.goroutine.count`, `go.processor.limit` and `go.config.gogc`.
`go.memory.used` splits `go.memory.type=stack|other`; it is not RSS.
`go.memory.limit` is Go's soft runtime limit, not a container limit; unlimited is omitted.

`process.uptime` is a gauge in seconds since OS process creation.

### Policy evaluations

Workers export `nuon.policy.evaluation.count` (counter) and
`nuon.policy.evaluation.duration` (histogram, seconds) per policy/input activity
attempt. Successful evaluations have `outcome=success` and `decision=pass|warn|deny`
(deny takes precedence); evaluator failures have `outcome=error` and bounded
`error.type=policy_validation|input_validation|deny_evaluation|warn_evaluation`.

### Drift plan evaluation

| Metric | Type | Unit | Dimensions |
| --- | --- | --- | --- |
| `nuon.install.drift.plan.evaluation.attempts` | Counter | attempts | target, outcome, decision or error.type |

`CheckNoopPlan` records returned interpretation attempts for component and sandbox
drift workflows (`target=component|sandbox`). Successful interpretation uses
`outcome=success` and `decision=drift|no_drift`; failures use `outcome=error|cancelled`
and `error.type=load_plan|evaluate_plan`. Retries count again.

This measures interpretation of existing plans, not plan generation, persisted drift
state, or complete drift-check outcomes. Normal deployment previews and older activity
requests without workflow type are excluded. Skipped checks and failures before plan
interpretation produce no observation; successful interpretation does not imply that
later status writes or notifications succeeded.

### App config sync

Standalone and branch app-config sync emit metrics when the shared `syncer.Run` returns:

| Metric | Type | Unit | Dimensions |
| --- | --- | --- | --- |
| `nuon.app.config.sync.attempts` | Counter | attempts | outcome, stage |
| `nuon.app.config.sync.duration` | Histogram | seconds | outcome |

Outcomes are `success`, `rejected`, `error`, or `cancelled` (including deadlines).
Stages are `load`, `intermediate`, `decode`, `sync_transaction`, or `deferred_queues`;
success uses `none`. A `sync.SyncErr` from any app-config resource counts as
`rejected` (invalid configuration, missing references, or unavailable features).
Operational failures count as `error`; cancellation takes precedence.

Duration includes deferred queue provisioning. Success does not imply downstream
build or install success. Failure logs include `config_committed` to distinguish
queue-setup failures after commit.

### Queue dispatch

The enqueuer emits metrics for dispatching signals to Temporal:

| Metric | Type | Unit | Dimensions |
| --- | --- | --- | --- |
| `nuon.queue.enqueuer.dispatch.attempts` | Counter | attempts | source, outcome |
| `nuon.queue.enqueuer.dispatch.duration` | Histogram | seconds | source, outcome |
| `nuon.queue.enqueuer.operations` | Counter | operations | source, operation, outcome |
| `nuon.queue.enqueuer.channel.dropped` | Counter | signals | None |
| `nuon.queue.enqueuer.local.backlog` | Gauge | signals | None |
| `nuon.queue.enqueuer.local.processing` | Gauge | signals | None |

Dimension keys use the `nuon.queue.enqueuer.` prefix. Sources are `channel`,
`await`, `sweep`, or `other`; outcomes are `success` or `failure`. Persistence
operations are `mark_enqueued` and `update_metadata`, with `other` as a fallback.

Dispatch metrics measure Temporal RPC attempts, not workflow completion.
Persistence failures are counted separately. Backlog and processing gauges cover
only the local channel, not the durable queue or inline/sweep calls. Channel drops
indicate local overflow, not deletion of persisted signals; sweep recovery still applies.

### Install state

State reads through `GetInstallState` and saves through `SaveState` emit:

| Metric | Type | Unit | Dimensions |
| --- | --- | --- | --- |
| `nuon.install.state.operations` | Counter | operations | `operation=get\|save`, `outcome=success\|error` |
| `nuon.install.state.operation.duration` | Explicit-bucket histogram | seconds | Same as operations |

Outcomes describe the returned result, including successful database fallback after
a blob-read failure.

### Blob storage

The shared blob service emits metrics for both S3 and GCS:

| Metric | Type | Unit | Dimensions |
| --- | --- | --- | --- |
| `nuon.blobstore.operations` | Counter | operations | `operation`, `outcome=success\|error\|closed_early` |
| `nuon.blobstore.operation.duration` | Explicit-bucket histogram | seconds | Same as operations |

Operations are `read`, `write`, `write_stream`, `metadata`, `read_stream_open`, and
`read_stream_body`. Stream bodies record once at EOF (`success`), read error, or
close before EOF (`closed_early`, or `error` if close fails). Body duration includes
consumer time. Other operations record `success` or `error` on return.

### Notification delivery

Lifecycle notifications emit metrics around outbound webhook and Slack calls:

| Metric | Type | Unit | Dimensions |
| --- | --- | --- | --- |
| `nuon.notification.delivery.attempts` | Counter | attempts | channel, operation, outcome |
| `nuon.notification.delivery.duration` | Histogram | seconds | Same as attempts |

Dimension keys use the `nuon.notification.` prefix: channel is `webhook` or `slack`,
operation is `post` or `update`, and outcome is `success` or `failure`.
Each outbound call counts separately, including retries and fan-out. Filtered
notifications are excluded. Success means the client call succeeded.

### Lifecycle hooks

| Metric | Type | Unit | Dimensions |
| --- | --- | --- | --- |
| `nuon.event.hook.invocations` | Counter | invocations | name, phase, invocation, outcome |

Dimension keys use the `nuon.event.hook.` prefix. Names are `flow_lifecycle_telemetry`,
`workflow_lifecycle_webhook`, `workflow_lifecycle_slack`, or `other`. Phases are
`validate`, `execute`, `cancel`, or `other`; invocation is `before` or `after`;
outcome is `success`, `error`, or `blocked` (before-phase only). Only supported hooks
that return are counted; delivery is measured separately.

Lifecycle logs use `flow_event` to identify retry, drift, config-update, and
component/install health events. Fields include resource IDs, retry counts, and
health status when available. Install config-update failures are logged at error level.

### Runner-api polling

Runner-api polling exports `nuon.runner.job_tail.sessions` and `.probes` by bounded
`outcome`, `.notification.wakes`, and `.listener.connected`, `.listener.failures`,
and `.listener.notifications`. Sessions begin after validation; empty timeouts
are healthy idle results. Probe retries count separately. Listener state is
observed continuously, including idle periods; routine rotation and shutdown do
not count as failures. These metrics describe attempts, not unique jobs or claims.

### Runner execution results

| Metric | Type | Unit | Dimensions |
| --- | --- | --- | --- |
| `nuon.runner.job.execution.results` | Counter | results | `nuon.runner.job.type`, `nuon.runner.job.operation`, `outcome` |

Runner-api records newly persisted results from compressed and uncompressed reports.
`outcome=success|failure` reflects the reported result, not workflow completion or
application health. Job type and operation are bounded; unrecognized values use `other`.
Duplicate reports do not count again; retries with new execution IDs count separately.

Includes planning, applying and action executions, but does not distinguish drift plans
from deployment previews or health-check actions from other actions. Missing reports,
control-plane-generated results and rejected/failed writes are excluded. Process loss
after persistence can lose the observation; this is not durable completion accounting.

## Export reliability

Metrics are aggregated in memory and exported periodically. There is no persistent
queue; process loss can lose unexported data, and the final shutdown export is
best-effort. Configure Collector buffering separately and use missing-data alerts
alongside independent availability probes.

## Testing

Run the tests and request-recording benchmark from the repository root:

```sh
go test -race ./services/ctl-api/internal/pkg/telemetry ./services/ctl-api/internal/pkg/metrics ./services/ctl-api/internal/pkg/api ./services/ctl-api/internal/pkg/db/poolmetrics ./services/ctl-api/internal/health ./services/ctl-api/internal/app/mcp/server
go test -run '^$' -bench '^BenchmarkHTTPMetrics$' -benchmem ./services/ctl-api/internal/pkg/telemetry
```
