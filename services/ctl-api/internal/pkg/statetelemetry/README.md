# State telemetry

State telemetry provides periodic snapshots of your Nuon fleet: **inventory, recorded health, reporting coverage and
current workflow activity**. Metrics are exported over OTLP, including when no operations are running.

| What you can see | Coverage |
| --- | --- |
| Fleet size and lifecycle | Organizations, apps, component definitions, installs, installed components and install-owned runners. |
| Recorded workload health | Install health summaries and component verdicts, including unknown and unevaluated components. |
| Health reporting | Recent, older and missing health-report arrivals. |
| Sandbox state | Provisioning lifecycle and recorded runtime health, shown separately. |
| Work in progress | Install workflows by operation, status and plan-only mode, including approval/retry waits and oldest creation time. |
| Collection status | Collection freshness, leadership, refresh failures and series limits. |

State metrics complement [operational metrics and logs](../telemetry/README.md); they are not live probes or
completed-operation counters.

## Enable collection

Configure [OTLP export](../telemetry/README.md#enable-export), then set:

```sh
STATE_TELEMETRY_ENABLED=true
STATE_TELEMETRY_LEVEL=summary
```

Collection is **off by default** and runs in internal or all-in-one API processes. It requires an OTLP endpoint and a
stable `nuon.control_plane.id`. The switch controls state collection only, not operational metrics, logs or health checks.

### Collection level

`STATE_TELEMETRY_LEVEL` selects the detail and series cost for the whole catalog. The default is `summary`.

| Level | Detail | Series cost |
| --- | --- | --- |
| `fleet` | Fleet totals without entity IDs. | Bounded independently of fleet size. |
| `summary` | Org counts by org; other counts by org/app. | Grows with orgs, apps and occupied categories. |
| `detailed` | Adds install IDs to install-owned measurements, plus install information and report timestamps. | Grows with installs and occupied categories. |

Metric names stay the same at every level; only their grouping changes. Org, app and component-definition counts keep
their natural ownership at detailed level. Individual installed-component IDs are not exported.

Each metric stream is limited to **4,095 series**. Exceeding a limit does not produce partial totals or automatically
reduce detail. Database query cost still grows with the retained population at every level.

## Metric reference

Names below are OpenTelemetry names, before backend-specific translation. All `.count` metrics are **gauges**:
they describe the recorded population at a point in time, not cumulative events.

### Inventory

Each inventory count is broken down by `resource.state`.

| Metric | Counts |
| --- | --- |
| `nuon.org.count` | Organizations. |
| `nuon.app.count` | Apps. |
| `nuon.component.count` | App-defined components. |
| `nuon.install.count` | Installs. |
| `nuon.install.component.count` | Installed component targets. |
| `nuon.runner.count` | Install-owned runner records, not live processes. |

Counts include retained resources and exclude soft-deleted records and invalid ownership links. Retained deprovisioned
resources still count; component definitions and installed targets are distinct populations.

### Health, reporting and sandboxes

| Metric | Breakdown | Meaning |
| --- | --- | --- |
| `nuon.install.health.count` | `nuon.health.status` | Installs by recorded component-health summary. |
| `nuon.install.component.health.count` | `nuon.health.status` | Installed components by recorded health verdict. |
| `nuon.install.health.report_age.count` | `nuon.health.report_age` | Installs by age of their last health-report arrival. |
| `nuon.install.sandbox.lifecycle.count` | `nuon.sandbox.status` | Installs by sandbox lifecycle, including missing sandboxes. |
| `nuon.install.sandbox.health.count` | `nuon.health.status` | Installs by recorded sandbox runtime health. |

Health has eight values: `healthy`, `progressing`, `degraded`, `unhealthy`, `unknown`, `not-applicable`, `unset`, `other`.
An install's summary can be healthy while some components remain unevaluated, so read it alongside component coverage.
`unset` does not mean healthy, and an active sandbox does not establish runtime health.

| Report age | Meaning |
| --- | --- |
| `within_5m` | A valid report arrived within five minutes of collection. |
| `over_5m` | The last valid arrival was older than five minutes. |
| `never_reported` | No recorded report arrival. |
| `invalid_time` | The recorded arrival is at/before Unix epoch or in the future. |

Arrival age is not health-evaluation freshness. Five minutes is a reporting bucket, not a promised reporting cadence.

### Workflow activity and waits

| Metric | Meaning |
| --- | --- |
| `nuon.workflow.current.count` | Current install-owned workflows, including recorded approval and retry waits. |
| `nuon.workflow.oldest_created_at` | Earliest creation time among those workflows, in Unix seconds. |

Both use `nuon.workflow.type`, `nuon.workflow.status` and `nuon.workflow.plan_only` as breakdowns. They cover install
operations such as provisioning, deployment, configuration changes, actions and teardown; app/branch builds are excluded.
Retries do not create additional workflow counts. Creation age includes retries and is **not time in the current wait**.
Empty groups or groups with invalid creation times have no oldest timestamp.

### Install details

The `detailed` level also provides:

| Metric | Meaning |
| --- | --- |
| `nuon.install.info` | One per install, with its resolved boolean `nuon.install.sandbox_mode`. |
| `nuon.install.health.last_report_at` | Last valid health-report arrival, in Unix seconds; absent when unavailable or invalid. |

Collection diagnostics use the `nuon.fleet.snapshot.*` prefix and describe collection status, not workload health.

## Reading the metrics

- Sum count contributions at one instant; do not apply counter rates to state gauges. Use MIN for oldest workflow
  timestamps, never SUM.
- Collection intervals are five minutes for org/app/component definitions and one minute for install-level state and
  workflows. Measurements refresh independently. Short workflows can finish between reads; use logs for their history.
- Missing data is unknown, not zero or recovery. Sparse categories and old samples require care when writing custom
  queries; successful export alone does not guarantee complete, coherent backend ingestion.
- These snapshots describe recorded state, not workload availability, uptime or complete operation history.
  They are intended for observational dashboards, not correctness-dependent aggregate health or recovery alerts.

Collection covers the configured database. Send it only to destinations authorized for that scope; dashboard filters
and entity labels are not tenant access controls.
