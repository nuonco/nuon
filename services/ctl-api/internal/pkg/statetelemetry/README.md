# State telemetry

This package provides background collection and OTLP export for database-backed
state metrics and collector diagnostics. Its inventory gauges describe recorded
lifecycle state, not workload health or successful deployment.

## Configuration

Configure [OTLP export](../telemetry/README.md#enable-export), then enable collection:

```sh
STATE_TELEMETRY_ENABLED=true
```

The equivalent service configuration is `state_telemetry_enabled: true`.
This global toggle controls all database-backed state collection and metric export
in this package, not existing request/operation metrics, logs or traces. Collection
is disabled by default and requires an OTLP endpoint. When disabled, it creates
neither a metric callback nor a database connection.

The collector runs in internal and all-in-one API processes, not in other API or
worker-only processes. Set a stable `nuon.control_plane.id` resource attribute.
The exporter generates `service.instance.id` per process; any override must remain
unique per live process. Enabled fleet collection rejects a missing or empty
control-plane or instance identity on the resolved telemetry resource.

## Collection behavior

- One process per primary database holds a session advisory lock and reads
  snapshots. Other processes remain on standby. The dedicated connection is
  separate from request pools, checked every 30 seconds and renewed every ten
  minutes. Leadership checks expire after 180 seconds without verification.
- A **family** groups metrics that share a refresh schedule and validity period
  (TTL). Each family refreshes independently in a bounded, read-only transaction.
  Metric export reads cached values, never the database.
- Successful refreshes replace the complete family together. Failed refreshes
  retain the previous values only until their TTL expires, measured from read
  start. Expiry or leadership loss stops value export; diagnostic timestamps remain.
- A family refresh exceeding 4,095 distinct attribute sets in any metric stream
  is rejected, rather than exporting partial totals. This leaves headroom below
  the observable-gauge SDK limit of 4,096. Counter/histogram limits remain 2,000.
  Fleet, app and org inventory refresh independently; scoped overflow does not
  suppress the fleet snapshot. They are not one database transaction together.

## Diagnostic metrics

All metric names below have the `nuon.fleet.snapshot.` prefix. Family metrics use
the `nuon.fleet.snapshot.family` dimension and appear only for registered families.
Diagnostics have no entity-ID dimensions. Timestamp values are Unix seconds.

| Suffix | Type | Meaning/dimensions |
| --- | --- | --- |
| `leader` | Gauge | 1 for a recently verified leader session, 0 otherwise |
| `leader.last_verified` | Gauge, Unix seconds | Last successful session check; absent before first leadership |
| `leadership.attempts` | Counter | Acquisition attempts; `outcome=acquired\|standby\|error` |
| `active` | Gauge | Per-family cache validity, including leadership and TTL; 0 before first success |
| `started_at` | Gauge, Unix seconds | Read start of last successfully published snapshot; absent before first success |
| `collected_at` | Gauge, Unix seconds | Publication of last successfully published snapshot; absent before first success |
| `observed_at` | Gauge, Unix seconds | When this family was observed for export; not database freshness |
| `refresh.attempts` | Counter | Completed attempts per family; `outcome=success\|error\|cancelled\|series_limit` |
| `refresh.duration` | Histogram, seconds | Refresh time including transaction cleanup; same dimensions as attempts |

Cancellation includes context deadlines. Database statement and lock timeouts
count as `error`. Session ping failures revoke leadership but are not acquisition
attempts. Collector leadership does not indicate workload health.

## Querying snapshot values

Snapshots describe database state, not independent contributions from replicas.
For each control plane and family, select one reporter before aggregating values:

1. Require recent samples, `leader=1`, leadership verification age at most 180
   seconds, `active=1`, and read-start age within the family's TTL.
2. Select the newest valid `collected_at`, breaking ties by instance ID. Use only
   that reporter's values; do not sum replicas or deduplicate counts with `max`.
3. Exclude value samples whose original timestamps precede that reporter's latest
   `observed_at`. This prevents backend lookback from retaining deleted entities
   or retired states.

Preserve original OTLP point timestamps and make control-plane and instance
resource attributes available for joins. Use the provider's single periodic
reader; do not merge multiple readers under one instance identity. Use
`observed_at`, not publication time, for the lookback filter because the SDK
timestamps gauge points after callbacks finish. Backend ingestion can temporarily
omit values or mix export frames while an export arrives. Timestamp precision
loss can also invalidate this filter. These steps alone do not establish atomic
backend totals: validate your backend's timestamp preservation, frame completeness
and alert hold/no-data behavior before using sums for alerts. Database query
validation does not prove backend ingestion or alert correctness.

Missing values are not zero. A successful empty snapshot, an expired snapshot,
and a family that has never collected data are different states.

## Inventory

All inventory metrics are integer gauges with unit `{resource}`. Families refresh
immediately after acquiring leadership, then every five minutes; successful data
expires fifteen minutes after read start. Export interval and ingestion delay add
to the time before a change is visible.

App/org grouping is enabled with the same global toggle as fleet collection:

| Metric | Family | Dimensions | Series per group |
| --- | --- | --- | --- |
| `nuon.inventory.current` | `inventory` | `resource.kind`, `resource.state`; runners add `nuon.runner.group.type` | 71 per control plane |
| `nuon.inventory.install.current` | `inventory.apps` | `nuon.org.id`, `nuon.app.id`, `resource.state` | 7 per app |
| `nuon.inventory.component.current` | `inventory.apps` | Same | 6 per app |
| `nuon.inventory.install_component.current` | `inventory.apps` | Same | 15 per app |
| `nuon.inventory.install_runner.current` | `inventory.apps` | Same, plus `nuon.runner.group.type=install` | 14 per app |
| `nuon.inventory.org_runner.current` | `inventory.orgs` | `nuon.org.id`, `resource.state`, `nuon.runner.group.type=org` | 14 per org |

Only retained apps/orgs with valid parents become groups. Apps with no installs,
components or runners and orgs with no org runners still emit zero categories.
Org/build runners have no unique owning app and are not duplicated across apps.
Names, customer labels and install/component/runner IDs are not dimensions.

Business-series cost is 71 + 42 per retained app + 14 per retained org, before
diagnostics and historical instance identities. At 200 apps and 20 orgs this is
8,751 series. Current bounds accept at most 273 apps and 292 orgs per snapshot;
one additional group rejects its entire scoped family. Query results and cache
allocation are bounded before publication. Backend aggregation does not remove
producer-side limits. Disable the global toggle to stop all snapshot collection.

| Kind | Counting unit | Recorded states (plus `unset` and `other`) |
| --- | --- | --- |
| `org` | Organization | `active`, `error`, `provisioning`, `deleting`, `deprovisioning`, `deprovisioned` |
| `app` | Application definition | `active`, `error`, `provisioning`, `deprovisioning`, `delete_queued` |
| `install` | Install | `provisioning`, `provisioned`, `deprovisioning`, `deprovisioned`, `reprovisioning` |
| `component` | App-defined component | `active`, `error`, `deprovisioning`, `delete_queued` |
| `install_component` | Component target within an install | `deleted`, `delete_failed`, `queued`, `active`, `inactive`, `error`, `noop`, `planning`, `syncing`, `executing`, `unknown`, `pending`, `disabled` |
| `runner` | Runner record, not a runner process or group | `error`, `active`, `pending`, `provisioning`, `deprovisioning`, `deprovisioned`, `reprovisioning`, `offline`, `awaiting-install-stack-run`, `awaiting-heartbeat`, `disabled`, `unknown` |

Installs use `lifecycle_phase.phase`. Other kinds prefer nonempty
`status_v2.status`, falling back to legacy `status`; this explicit precedence also
applies when runner status columns disagree. Empty/null state becomes `unset`;
unrecognized values become `other`. Every category emits zero after a successful
empty read, including categories that became empty since the previous refresh.

Membership excludes soft-deleted resources and parents. Apps must belong to a
retained org; installs and components must belong to a retained app in the same
org. Installed targets must link an install and component from the same app/org.
Runners must link a retained group and its matching org/install owner, with
consistent org IDs. All retained runners count, not only the newest runner in a
group. Invalid links, unknown group types and mismatched owner types are excluded.

Deprovisioned records and lifecycle-deleted component targets remain counted
until soft deletion. Sandbox, integration/test and ordinary populations all count;
inventory has no production-only or sandbox-mode dimension. It covers the entire
configured control-plane database, not one request's org. Route these totals only
to destinations authorized for that scope. Org/app dimensions support filtering;
they are not tenant access controls or permission to export another org's data.

After applying the reporter/freshness selection above:

- **Fleet size:** filter `resource.kind=install`, then sum across states. For only
  provisioned installs, also filter `resource.state=provisioned`.
- **Component footprint:** compare totals for `component` (definitions) with
  `install_component` (installed targets); they are not interchangeable.
- **App overview:** filter scoped metrics by `nuon.org.id` and `nuon.app.id`,
  then group by `resource.state`. Use `inventory.apps` freshness rather than
  the independent fleet family's freshness. Do not sum scoped and fleet totals.
- **Runner state:** filter `resource.kind=runner` and group by state and group
  type. `offline`/`error` are recorded states, not live process checks; disabled,
  startup and deprovisioned states must not automatically page as outages.
- **Drill-down:** use the app/org IDs to navigate authenticated resource listings
  and lifecycle logs. Inventory does not identify individual failed operations;
  logs provide that context, while traces investigate active execution latency.
- **Data quality:** chart `unset` and `other` separately rather than treating
  them as active or healthy.

These are snapshot totals: do not apply `rate` or sum across refreshes. For
inventory alerts, require `inventory` freshness; a missing or expired fleet must
produce a no-data condition, not a zero-fleet or recovered-health result.

## Troubleshooting

- **No leadership metrics:** check the enable flag, OTLP endpoint, process type
  and Collector metrics pipeline.
- **No recently verified leader:** check primary-database connectivity and
  credentials. Standby processes reporting `leader=0` are normal.
- **Refresh failures:** inspect process logs and snapshot age. A fresh cached
  value can remain valid after a failed refresh; failure does not mean an empty fleet.
- **RLS or permission failures:** inventory requires an authorized database-wide
  read. Its transaction sets `row_security=off` so PostgreSQL errors if row-level
  policies would filter the result, rather than publishing partial totals. This
  does not bypass policies or grant privileges. Check the configured role and
  deployment's access policy; do not disable tenant security to repair telemetry.
- **`series_limit`:** reduce the reader's detail cardinality or use lower-cardinality
  aggregates. Fleet totals remain independent of scoped limits. A rejected
  refresh does not contain usable partial results, and old scoped data expires.

## Adding a snapshot reader

Register `family` values with `newCollector` before startup. Create their
instruments using the collector's `statetelemetry` meter. Readers return immutable
cached observations and accurate distinct attribute-set counts per instrument.
Keep aggregate and detail families independent, with no shared instruments.

Reads use a six-second context deadline, five-second statement timeout,
100-millisecond lock timeout, 4 MB work memory and no parallel query workers.
Transactions are read-only and repeatable-read; cleanup is bounded to one second.
Bound result allocation as well as query time, exclude soft-deleted entities and
parents, and respect tenant ownership. The connection uses configured database
credentials without changing role privileges or setting tenant context. Inventory
uses one projected aggregate query over base tables, not GORM models or views.
