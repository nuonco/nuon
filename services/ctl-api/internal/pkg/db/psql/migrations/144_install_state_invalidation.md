# Install-state publication rollout

This is a coordinated upgrade, not a binary-only rolling upgrade or rollback.
Old binaries write snapshot metadata directly; new binaries use `install_state_heads`.
Do not run old and new state writers together.

## Deploy

1. Pause ingress and drain ctl-api API processes, workers, and other source writers.
   Preserve Temporal histories: legacy `SaveState` activity requests remain supported
   and regenerate from current sources rather than saving their cached payload.
2. Run migrations from one new-binary process before starting the remaining processes.
   Source writers must use PostgreSQL's default `READ COMMITTED` isolation; the
   source barrier relies on fresh snapshots for the trigger's target lookup.
   Check `SHOW default_transaction_isolation`. The invalidation function rejects
   writes at any other isolation level with SQLSTATE `25000`, rolling back both
   the source mutation and its invalidation. This also covers explicit transaction
   overrides and writes that currently map to no install. Read-only transactions
   are unaffected; do not enable repeatable-read source transactions.
   Migration 144 installs the coordination functions. Each 145 migration installs
   one table's triggers in its own transaction with `lock_timeout = '2s'`.
   Successful stages remain committed; a blocked stage rolls back and fails startup.
   Remove the blocker, then retry migrations. Do not raise the timeout or resume on
   partial migration success.
3. Migration 146 creates missing heads for all existing live installs with no
   snapshot pointer and all partials stale. The first read rebuilds from current
   sources and publishes through the new protocol. Legacy inline JSON and blobs
   can disagree even on rows marked clean, so neither is adopted. Archived legacy
   rows also cannot block rebuilding when blob reads are disabled. Existing heads
   are preserved on retries. Installs created after migration get their heads from
   the install trigger. Confirm that every 144/145/146 stage is `applied` before
   starting the new processes and resuming traffic. Expect increased regeneration
   and blob uploads while existing installs are first accessed.
4. Compare write latency, lock waits, PostgreSQL deadlocks, state regeneration rate,
   publication conflicts and Temporal activity errors with the pre-deploy baseline.
   Stop the rollout if write latency or lock waits rise materially. Investigate any
   `40P01` or `55P03` error before proceeding. Production load measurements are still
   required; the integration suite is not a write-throughput benchmark.

Runner-job status is not an invalidation input. Execution status only invalidates
when finished/non-finished priority changes. Actions expose their run status, so
those status changes must invalidate. Secrets targets match only secrets-sync jobs.
UPDATE barriers lock only orgs whose rows changed a listed source column, including
both old and new row identities. This changed-row filter uses only the transition
tables. Target lookup must remain a separate statement **after** the barrier lock,
even when no install currently matches the source row.
Transition tables still contain PostgreSQL's row images, but our materialized
projection carries IDs/FKs and eligibility fields, not output or template payloads.

With OTLP metrics enabled, monitor `nuon.install.state.source_wait.timeouts`.
Alert on any increase over five minutes during rollout: each increment is a state
refresh that exhausted its five-second source barrier wait. Caller cancellation
and shorter caller deadlines do not increment this counter. Inspect long-running
source transactions in the affected database before retrying; do not bypass the
barrier. `nuon.install.state.operations` and
`nuon.install.state.operation.duration` with `operation=source_wait` also expose
wait counts, errors and duration. Alert configuration is a separate rollout step.

The barrier deliberately waits for all same-org source transactions, including
long config syncs. Keep the five-second bound: state API reads return HTTP 503 with
`Retry-After: 1` on this timeout or publication retry exhaustion. Publication retries
back off and remain retryable Temporal failures after five attempts. Both state
read and regeneration activities allow five minutes for source reads and uploads;
this budget does not extend the barrier wait or eliminate org-level coupling.

Every successful publication runs best-effort retention outside its transaction,
keeping inline JSON for the latest 50 snapshots and the current head. Failures are
logged and retried on the next publication (and by legacy ArchiveState activities).
Archived snapshots retain their immutable blobs; history rows are not deleted.

Head invalidation uses one sorted upsert pass for both missing and existing rows.
It does not guarantee deadlock freedom for arbitrary multi-statement transactions
that visit installs in opposite orders. A `40P01` abort must be retried at the whole
source transaction boundary, including its source mutation; never retry just the
trigger or publish after a failed write. This change does not add blanket retries
to API mutations or transactions with external side effects.

## Roll back

1. Pause ingress and stop **all** new/old readers, workers and source writers.
2. Back up the database. Verify the intended database connection, then run:

   ```sh
   psql "$DATABASE_URL" -v install_state_writers_stopped=on \
     -f services/ctl-api/internal/pkg/db/psql/migrations/rollback_install_state_publication.sql
   ```

   The script refuses current blob-only snapshots: restore their inline JSON from
   the blobs first. It drops only the state-invalidation triggers, one autocommit
   statement at a time under a two-second lock timeout. It marks nondeleted inline
   snapshots stale in the old format, including restored archived rows, and detaches
   their blob pointers (so old read-repair cannot re-read older bytes). It also marks
   unarchived blob-only rows stale, clears heads and removes only this feature's
   migration records/functions. It keeps snapshots, stored blobs, source records,
   the head table and migration 112's job-notification trigger.
3. If any step fails, keep traffic paused and fix the blocker before rerunning the
   script. Its statements are repeatable; never resume a partially rolled-back DB.
4. Start the previous binary and verify repeated reads after an input update.
   A later upgrade reruns 144/145/146 and creates dirty heads for a fresh rebuild.

The script does not give the old binary the new publication guarantees; it only
converts coordination metadata back to the format that binary understands.
