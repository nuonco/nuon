# State telemetry

State telemetry provides periodic snapshots of your Nuon fleet: **inventory, recorded health, reporting coverage and
deployment/workflow state**. Metrics are exported over OTLP, including when no operations are running.

| What you can see | Coverage |
| --- | --- |
| Fleet size and lifecycle | Organizations, apps, component definitions, installs and installed components. |
| Install identity | Install and tracked app-branch names, sandbox mode and opt-in vendor labels at detailed level. |
| Install runner | Recorded current-runner status, with missing and unknown states shown explicitly. |
| Recorded workload health | Install health summaries and component verdicts, including unknown and unevaluated components. |
| Health reporting | Recent, older and missing health-report arrivals. |
| Sandbox state | Provisioning lifecycle and recorded runtime health, shown separately. |
| Work in progress | Install workflows by operation, status and plan-only mode, including approval/retry waits and oldest creation time. |
| Latest deployment | Each install's most recent non-plan deployment attempt and recorded status, including installs with no eligible history. |
| Configuration position | Selected install versions and newest-created branch versions/times—not applied versions or deployability. |
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
Branch configuration versions/times are available at summary and detailed levels; selected install versions require
detailed collection. Configuration positions are absent at fleet level: version numbers cannot meaningfully be summed.

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

Counts include retained resources and exclude soft-deleted records and invalid ownership links. Retained deprovisioned
resources still count; component definitions and installed targets are distinct populations.

### Install runner

`nuon.install.runner.count` counts **installs**, broken down by `nuon.runner.status`. Each retained install contributes
once: its runner's recorded status when exactly one retained install-owned group and runner identify it, or `missing`
when no runner exists. `unknown` covers both a recorded unknown status and multiple groups or runners preventing a
unique selection.

Deleted and invalid ownership links are excluded. Neither runner name, creation time nor an active status resolves
ambiguity. Disabled and deprovisioned runners keep their recorded status; a missing runner is not deprovisioned.
This is not a live-process count or proof that jobs can execute. Detailed collection uses one series per install,
regardless of how many runner records are retained.

### Health, reporting and sandboxes

| Metric | Breakdown | Meaning |
| --- | --- | --- |
| `nuon.install.health.count` | `nuon.health.status`, `nuon.install.lifecycle` | Installs by recorded component-health summary and lifecycle phase. |
| `nuon.install.component.health.count` | `nuon.health.status` | Installed components by recorded health verdict. |
| `nuon.install.health.report_age.count` | `nuon.health.report_age` | Installs by age of their last health-report arrival. |
| `nuon.install.sandbox.lifecycle.count` | `nuon.sandbox.status` | Installs by sandbox lifecycle, including missing sandboxes. |
| `nuon.install.sandbox.health.count` | `nuon.health.status` | Installs by recorded sandbox runtime health. |

Health has eight values: `healthy`, `progressing`, `degraded`, `unhealthy`, `unknown`, `not-applicable`, `unset`, `other`.
An install's summary can be healthy while some components remain unevaluated, so read it alongside component coverage.
The install summary matches the API's install `composite_health_status`: it rolls up every non-deleted installed-component record of
the install, while `nuon.install.component.health.count` counts only targets with retained, tenant-consistent component
definitions, so the two populations can differ. `nuon.install.lifecycle` uses the `nuon.install.count` lifecycle values; fleet
collection reports all 56 health/lifecycle pairs.
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

### Latest deployment

`nuon.install.deploy.latest.count` counts installs by their latest eligible deployment's `nuon.workflow.type` and
`nuon.workflow.status`. Each install contributes once; `type=none`, `status=never_deployed` means no eligible retained
history, not a failed read. Results remain visible when workflows finish. This is not a completion counter or success rate.

Deployment eligibility follows the install deployment feed: provisioning/reprovisioning, stack/sandbox changes,
component deployments and teardown, component enablement/recovery, app-branch/install sync, input/secret changes
and deprovisioning. Configuration-only changes can qualify without rolling out workloads. Plan-only workflows,
actions, runbooks and app-owned workflows are excluded. Actual deployments to preview installs still qualify.

The newest creation time wins, with ID as a tie-breaker—not the newest success or finish time. Recorded approval/retry
statuses are preserved; a finish timestamp alone proves neither success nor terminality. Missing or unrecognized
statuses appear as `unset` or `other` rather than success.

### Configuration position

| Metric | Meaning | Level |
| --- | --- | --- |
| `nuon.install.config.selected.ordinal` | App-wide version ordinal of the install's selected config—not necessarily applied. | Detailed |
| `nuon.app_branch.config.latest_created.ordinal` | App-wide ordinal of the newest retained config created on each branch, including branches without installs. | Summary/detailed |
| `nuon.app_branch.config.latest_created.created_at` | Valid creation time of that same branch config, in Unix seconds. | Summary/detailed |

Ordinals include deleted configuration history and are app-wide, not branch-local run counts. Equal creation times
can make ordinals ambiguous; unresolved ordinals are omitted, never reported as zero. A non-unique newest branch config
has neither measurement. Missing/deleted configs and invalid ownership are excluded; invalid creation clocks are omitted.

`nuon.app_branch.id` identifies the **configuration's branch**, which can differ from the install's tracked branch.
Unbranched selections have no branch attribute. `nuon.config.status` describes config sync (`active`, `pending`,
`syncing`, `error`, `outdated`, `unset`, `other`), not build readiness; `nuon.config.source` reports a bounded recorded
source (`git-run`, `manual-run`, `git-preview-run`, `unset`, `other`). Newest-created can be a preview, failed or
incomplete configuration. It is not the latest deployable release, a drift verdict or proof of a completed rollout.

These streams cost at most one selected point per install and two points per branch. They refresh independently of
install metadata; read positions as observations, not an atomic convergence result.

### Install details

The `detailed` level also provides:

| Metric | Meaning |
| --- | --- |
| `nuon.install.info` | Value 1 per install, with its name, tracked app-branch ID/name, resolved sandbox mode and optional vendor labels. |
| `nuon.install.health.last_report_at` | Last valid health-report arrival, in Unix seconds; absent when unavailable or invalid. |
| `nuon.install.deploy.latest.created_at` | Creation time of the selected latest deployment, in Unix seconds. |
| `nuon.install.deploy.latest.finished_at` | Recorded finish time of that same deployment, in Unix seconds; not proof of success. |

Deployment timestamps use the count's workflow type/status dimensions. Missing, nonpositive and future timestamps
are omitted; an older deployment's timestamps are never substituted. Detailed deployment telemetry costs at most
three points per install across its count and two timestamp streams.

Install info uses `nuon.install.name`, `nuon.app_branch.id`, `nuon.app_branch.name` and boolean
`nuon.install.sandbox_mode`, alongside org/app/install IDs. Branch metadata describes the newest active connection,
not the configuration successfully deployed. Unavailable branch metadata is omitted without removing the install.

To correlate installs with your own customer/account identifiers, explicitly allow selected stored install labels:

```sh
STATE_TELEMETRY_LEVEL=detailed
TELEMETRY_INSTALL_LABELS=customer_id,tier
```

No install labels are exported by default. Selected values appear only on `nuon.install.info` as
`nuon.install.label.customer_id`, for example; other state metrics do not gain these dimensions. Names and labels can
contain sensitive data: allow only nonsecret identifiers approved for your telemetry destination.

The allowlist accepts at most eight distinct keys, each 1–64 ASCII characters, starting with a letter and containing
only letters, digits, `_`, `.` or `-`. Reserved `nuon`, `otel` and `service` prefixes and keys colliding after Prometheus
normalization are rejected. Metadata values must fit 256 UTF-8 bytes without control characters. Invalid/oversized
values are omitted, never truncated or logged; `nuon.install.metadata.status=omitted` signals this. `complete` means
no invalid values were withheld, not that every optional label or branch exists.

There is still one current info point per install. Renames, branch moves and label changes create historical backend
series. Use fresh info with one unambiguous entry per org/app/install before joining by those IDs; do not join raw
retained info series or attach metadata to every measurement.

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
