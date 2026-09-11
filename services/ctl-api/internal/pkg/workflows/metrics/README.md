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
| `nuon.workflow.executions.completed` | Completed execute-workflow signal executions by `workflow.type` and persisted `workflow.outcome`: success, error, cancelled, or unknown. |
| `nuon.workflow.step.retries` | Persisted step retry decisions by `workflow.type` and `retry.source`: auto or manual; not attempts started or Temporal retries. |
| `nuon.workflow.start_delay` | Seconds from creation to first persisted execution start, by `workflow.type`. |
| `nuon.workflow.elapsed_time` | Seconds from creation to observed execution completion, by `workflow.type` and `workflow.outcome`; includes queueing, approvals and retry waits. |

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

Counters run in status activities across workers, independently of snapshot
leadership. They start at zero per process; apply reset-aware rates before summing
replicas. Retry parks do not complete executions. A terminal signal without a
terminal workflow status counts as unknown. Sequential duplicate writes are
suppressed, but concurrent writers can duplicate observations; process crashes,
lookup failures and export failures can lose them. These are best-effort execution
counts, not lifetime-unique workflow totals.

Duration histograms observe transitions, not periodic snapshots. The start
activity preserves the first timestamp; elapsed time uses the same completion
boundary as the counter, not `FinishedAt`. Missing or reversed timestamps are
omitted. No observations means absent data, not a zero-duration sample. Explicit
buckets extend to one hour for start delay and one day for elapsed time; larger
values remain in the overflow bucket. Apply rates before aggregating replicas.

Migration 132 adds the partial index used by snapshot queries. PostgreSQL tests
use the migrated `INTEGRATION` harness; see `conventions/testing.md`.
