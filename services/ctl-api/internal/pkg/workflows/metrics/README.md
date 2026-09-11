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
