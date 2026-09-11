# Workflow metrics

The general worker exports snapshots when an OTLP endpoint is configured. A
dedicated primary PostgreSQL session holds an advisory lock so only one replica
queries at a time. Collection runs every 30 seconds with a five-second timeout;
export callbacks only read memory. Failures do not affect workflow execution.

Scope: non-deleted, install-owned `provision`, `manual_deploy`, and
`deploy_components` workflows on non-deleted installs, excluding plan-only runs.

| Metric | Meaning |
| --- | --- |
| `nuon.workflow.current` | Current counts by `workflow.type` and `workflow.state`; empty categories emit zero. |
| `nuon.workflow.oldest_created_at` | Oldest workflow creation timestamp in each nonempty category, in Unix seconds; not phase entry time. |
| `nuon.workflow.snapshot.collected_at` | Last successful snapshot query start, in Unix seconds; unchanged on failure. |

States are queued, executing, awaiting approval, and awaiting retry. These are
current-state gauges, not historical outcome counters. Counts expire after 180
seconds or loss of leadership. Select the newest fresh reporter snapshot before
aggregating replicas; filter oldest timestamps against positive current counts.
The default freshness policy assumes OTLP export at least every 60 seconds.

Awaiting retry requires a linked, nonterminal `execute-workflow` queue signal.
Older workflows without that ownership link are excluded. External Temporal
termination remains visible until the queue signal records its terminal state;
the default dispatcher timeout is 30 days. Executing and approval states use
the workflow's persisted status, not a live Temporal check.

Migration 132 adds the partial index used by snapshot queries. PostgreSQL tests
use the migrated `INTEGRATION` harness; see `conventions/testing.md`.

## Deployment snapshots

The same leader collects deployment snapshots every five minutes, independently
of workflow snapshot freshness. Migration 133 adds three partial indexes. Reads
use a five-second statement timeout, 100 ms lock timeout, and no parallel workers.

| Metric | Meaning |
| --- | --- |
| `nuon.deployment.attempts.recent` | Apply records created in the preceding 24 hours, by current recorded state; retries are separate attempts. |
| `nuon.deployment.applies.recent` | Apply records with `applied_at` in that window; best-effort, not health-verified successes. |
| `nuon.deployment.latest` | Install targets by latest eligible apply state, not live workload health. |
| `nuon.deployment.snapshot.collected_at` | Last successful collection completion, in Unix seconds. |

Counts carry org/app/component IDs, plus `deployment.state` where applicable.
Deleted entities, plan-only workflows, unlinked deploys, teardown and recovery
operations are excluded. Windows end at query start. Use these as gauges, not
with `rate()` or `increase()`. Select the newest fresh reporter and reject samples
older than its collection completion; preserve original OTLP sample timestamps.
Counts expire after 15 minutes or leadership loss. Failures retain the previous
snapshot; exceeding 1,999 series per metric rejects the whole refresh rather
than exporting partial totals. Observed states emit zero when emptied; components
without eligible deployment history are absent.

## Inventory and queue snapshots

The same leader collects these every five minutes with independent freshness and
the deployment query limits. Failures are also throttled to five minutes.

| Metric | Meaning |
| --- | --- |
| `nuon.inventory.current` | Non-deleted orgs, apps and installs by bounded `resource.kind` and `resource.state`; deleted parents are excluded. |
| `nuon.queue.current` | Non-deleted signals on live queues/orgs in `awaiting_dispatch`, `queued`, or `executing`, by `queue.state`. |
| `nuon.queue.oldest_created_at` | Oldest creation timestamp in each nonempty queue state, in Unix seconds; not phase entry time. |
| `nuon.inventory.snapshot.collected_at` | Last successful inventory/queue snapshot completion, in Unix seconds. |

These are recorded states, not live health or Temporal liveness. Terminal signals
are excluded. Unknown resource states use `unknown`. Empty categories emit zero;
counts expire after 15 minutes or leadership loss. Select the newest fresh reporter
before aggregation, using the same timestamp rules as deployment snapshots.
